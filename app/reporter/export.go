package reporter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	tbapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"

	"github.com/radio-t/gitter-rt-bot/app/storage"
)

// Exporter performs conversion from log file to html
type Exporter struct {
	ExporterParams
	location *time.Location
	botAPI   *tbapi.BotAPI
	s3       *storage.S3

	fileIDToURL map[string]string
}

// ExporterParams for locations
type ExporterParams struct {
	OutputRoot   string
	InputRoot    string
	TemplateFile string
	SuperUsers   SuperUser
}

type SuperUser interface {
	IsSuper(user string) bool
}

// NewExporter from params, initializes time.Location
func NewExporter(botAPI *tbapi.BotAPI, s3 *storage.S3, params ExporterParams) *Exporter {
	log.Printf("[INFO] exporter with %v", params)
	result := Exporter{
		ExporterParams: params,
		botAPI:         botAPI,
		s3:             s3,
		fileIDToURL:    map[string]string{},
	}

	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatalf("[ERROR] can't load location, error %v", err)
	}
	result.location = location
	return &result
}

// Export to html with showNum
func (e Exporter) Export(showNum int, yyyymmdd int) {
	from := fmt.Sprintf("%s/%s.log", e.InputRoot, time.Now().Format("20060102")) // current day by default
	if yyyymmdd != 0 {
		from = fmt.Sprintf("%s/%d.log", e.InputRoot, yyyymmdd)
	}
	to := fmt.Sprintf("%s/radio-t-%d.html", e.OutputRoot, showNum)

	messages, err := readMessages(from)
	if err != nil {
		log.Fatalf("[ERROR] failed to read messages from %s, %v", from, err)
	}
	fh, err := os.OpenFile(to, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("[ERROR] failed to export, %v", err)
	}
	defer fh.Close()

	fh.WriteString(e.toHTML(messages, showNum))
	log.Printf("[INFO] exported %d lines to %s", len(messages), to)

}

func (e Exporter) toHTML(messages []tbapi.Message, num int) string {

	type Record struct {
		Time   string
		IsHost bool
		Msg    tbapi.Message
	}

	type Data struct {
		Num     int
		Records []Record
	}

	data := Data{Num: num}
	for _, msg := range messages {
		switch {
		case msg.Photo != nil:
			for _, photo := range *msg.Photo {
				fileURL, err := e.maybeUploadFile(photo.FileID, false)
				if err != nil {
					log.Printf("[ERROR] failed to get file URL for %s", photo.FileID)
					continue
				}
				e.fileIDToURL[photo.FileID] = fileURL
				// hacky: need to pass fileURL to template
				// using this map in template FuncMap later
			}
		case msg.Sticker != nil:
			fileURL, err := e.maybeUploadFile(msg.Sticker.Thumbnail.FileID, true)
			if err != nil {
				log.Printf("[ERROR] failed to get file URL for %s", msg.Sticker.Thumbnail.FileID)
				continue
			}
			e.fileIDToURL[msg.Sticker.Thumbnail.FileID] = fileURL
		}

		username := ""
		if msg.From != nil {
			username = msg.From.UserName
		}

		rec := Record{
			Time:   time.Unix(int64(msg.Date), 0).In(e.location).Format("15:04:05"),
			IsHost: e.SuperUsers.IsSuper(username),
			Msg:    msg,
		}
		data.Records = append(data.Records, rec)
	}

	funcMap := template.FuncMap{
		"fileURL": func(fileID string) string {
			if url, found := e.fileIDToURL[fileID]; found {
				return url
			}
			return ""
		},
		"counter": func() func() int {
			i := -1
			return func() int {
				i++
				return i
			}
		},
		"jpg": jpg,
	}
	name := e.TemplateFile[strings.LastIndex(e.TemplateFile, "/")+1:]
	t, err := template.New(name).Funcs(funcMap).ParseFiles(e.TemplateFile)
	if err != nil {
		log.Fatalf("failed to parse, %v", err)
	}

	var html bytes.Buffer
	if err := t.ExecuteTemplate(&html, name, data); err != nil {
		log.Fatalf("[ERROR] failed, error %v", err)
	}
	return html.String()
}

func (e Exporter) maybeUploadFile(fileID string, doConvertWebPToJpg bool) (string, error) {
	url, err := e.botAPI.GetFileDirectURL(fileID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get file direct URL: %s", fileID)
	}

	if !strings.Contains(url, ".") {
		return "", errors.New("FileDirectURL has no extension")
	}

	ext := url[strings.LastIndex(url, "."):]
	fileName := fileID + ext

	fileExists, err := e.s3.FileExists(fileName)
	if err != nil {
		return "", errors.Wrapf(err, "failed to check if file exists alredy: %s", fileID)
	}

	if fileExists {
		return e.s3.BuildLink(fileName), nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get file by direct URL (fileID: %s)", fileID)
		// don't expose url – it contains Bot API Token
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.Wrapf(err, "non-200 response from file direct URL (fileID: %s)", fileID)
	}

	if strings.HasSuffix(fileName, ".webp") && doConvertWebPToJpg {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", errors.Wrapf(err, "failed to read response body for file direct URL (fileID: %s)", fileID)
		}
		resp.Body.Close()

		jpgBody, err := convertWebPToJpg(bytes.NewBuffer(bodyBytes))
		if err != nil {
			return "", errors.Wrapf(err, "failed to convert WebP to JPG (fileID: %s)", fileID)
		}

		_, err = e.s3.UploadFile(jpg(fileName), jpgBody)
		if err != nil {
			return "", err
		}

		resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	return e.s3.UploadFile(fileName, resp.Body)
}

func convertWebPToJpg(reader io.Reader) (io.Reader, error) {
	var b bytes.Buffer

	cmd := exec.Command("dwebp", "-o", "-", "--", "-")
	cmd.Stdin = reader
	cmd.Stdout = &b
	err := cmd.Start()

	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func jpg(fileURL string) string {
	if !strings.Contains(fileURL, ".") {
		log.Printf("[ERROR] fileURL has no extension: %s", fileURL)
		return ""
	}

	return fileURL[0:strings.LastIndex(fileURL, ".")] + ".jpg"
}

func readMessages(path string) ([]tbapi.Message, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	messages := []tbapi.Message{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		msg := tbapi.Message{}
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			log.Printf("[ERROR] failed to unmarshal %s, error=%v", line, err)
			continue
		}
		if !filter(msg) {
			messages = append(messages, msg)
		}
	}
	return messages, scanner.Err()
}

func filter(msg tbapi.Message) bool {
	contains := func(s []string, e string) bool {
		e = strings.TrimSpace(strings.ToLower(e))
		for _, a := range s {
			if strings.ToLower(a) == e {
				return true
			}
		}
		return false
	}
	return contains([]string{"+1", "-1", ":+1:", ":-1:"}, msg.Text)
}
