package reporter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/radio-t/gitter-rt-bot/app/bot"
	"github.com/radio-t/gitter-rt-bot/app/storage"
)

// magic number used to trim text in reply
// 43 is the number that Telegram clients are using
const maxQuoteLength = 43

// Exporter performs conversion from log file to html
type Exporter struct {
	ExporterParams
	location      *time.Location
	fileRecipient FileRecipient
	converters    map[string]Converter
	storage       storage.Storage

	fileIDToURL map[string]string
}

// ExporterParams for locations
type ExporterParams struct {
	OutputRoot   string
	InputRoot    string
	TemplateFile string
	SuperUsers   SuperUser
}

// SuperUser knows which user is a superuser
type SuperUser interface {
	IsSuper(user string) bool
}

// NewExporter from params, initializes time.Location
func NewExporter(
	fileRecipient FileRecipient,
	converters map[string]Converter,
	storage storage.Storage,
	params ExporterParams,
) *Exporter {
	log.Printf("[INFO] exporter with %v", params)
	result := Exporter{
		ExporterParams: params,
		fileRecipient:  fileRecipient,
		converters:     converters,
		storage:        storage,
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
func (e *Exporter) Export(showNum int, yyyymmdd int) {
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

func (e *Exporter) toHTML(messages []bot.Message, num int) string {

	type Record struct {
		Time   string
		IsHost bool
		Msg    bot.Message
	}

	type Data struct {
		Num     int
		Records []Record
	}

	data := Data{Num: num}
	for _, msg := range messages {
		e.maybeDownloadFiles(msg)

		data.Records = append(
			data.Records,
			Record{
				Time:   msg.Sent.In(e.location).Format("15:04:05"),
				IsHost: e.SuperUsers.IsSuper(msg.From.Username),
				Msg:    msg,
			},
		)
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
		"png": func(fileURL string) string {
			return fileURL + ".png"
		},
		"trim": func(text string) string {
			runes := []rune(text)

			if len(runes) < maxQuoteLength {
				return text
			}

			return string(runes[0:maxQuoteLength]) + "..."
		},
		"sizeHuman":      sizeHuman,
		"timestampHuman": e.timestampHuman,
	}
	name := e.TemplateFile[strings.LastIndex(e.TemplateFile, "/")+1:]
	t, err := template.New(name).Funcs(funcMap).ParseFiles(e.TemplateFile)
	if err != nil {
		log.Fatalf("[ERROR] failed to parse, %v", err)
	}

	var html bytes.Buffer
	if err := t.ExecuteTemplate(&html, name, data); err != nil {
		log.Fatalf("[ERROR] failed, error %v", err)
	}
	return html.String()
}

func (e *Exporter) timestampHuman(t time.Time) string {
	return t.In(e.location).Format("15:04:05")
}

func (e *Exporter) maybeDownloadFiles(msg bot.Message) {
	switch {
	case msg.Picture != nil:
		imageType := "png"
		if msg.Picture.Class == "sticker" || msg.Picture.Class == "animated-sticker" {
			imageType = "webp"
		}

		e.maybeDownloadFile(msg.Picture.Image.FileID, imageType)

		if msg.Picture.Thumbnail != nil {

			e.maybeDownloadFile(msg.Picture.Thumbnail.FileID, imageType)
		}

		for _, source := range (*msg.Picture).Image.Sources {
			e.maybeDownloadFile(source.FileID, source.Type)
		}

		for _, source := range (*msg.Picture).Sources {
			e.maybeDownloadFile(source.FileID, source.Type)
		}

	case msg.Document != nil:
		e.maybeDownloadFile(msg.Document.FileID, "")

		if msg.Document.Thumbnail != nil {
			e.maybeDownloadFile(msg.Document.Thumbnail.FileID, "")
		}

	case msg.Animation != nil:
		e.maybeDownloadFile(msg.Animation.FileID, "")

		if msg.Animation.Thumbnail != nil {
			e.maybeDownloadFile(msg.Animation.Thumbnail.FileID, "")
		}

	case msg.Voice != nil:
		for _, source := range (*msg.Voice).Sources {
			e.maybeDownloadFile(source.FileID, source.Type)
		}

	case msg.Video != nil:
		e.maybeDownloadFile(msg.Video.FileID, "")

		if msg.Video.Thumbnail != nil {
			e.maybeDownloadFile(msg.Video.Thumbnail.FileID, "")
		}
	}
}

func (e *Exporter) maybeDownloadFile(fileID string, fileType string) {
	if fileID == "" {
		return
	}

	if _, found := e.fileIDToURL[fileID]; found {
		// already downloaded
		return
	}

	if strings.Contains(fileID, ".") {
		// hacky way to handle WebP & TGS stickers
		// convertion to happens after image "FileID" download
		e.fileIDToURL[fileID] = e.storage.BuildLink(fileID)
		return
	}

	fileExists, err := e.storage.FileExists(fileID)
	if err != nil {
		log.Printf("[ERROR] failed to check if file exists alredy: %s: %v", fileID, err)
		return
	}

	if fileExists {
		e.fileIDToURL[fileID] = e.storage.BuildLink(fileID)
		return
	}

	log.Printf("[DEBUG] downloading file %s", fileID)
	body, err := e.fileRecipient.GetFile(fileID)
	if err != nil {
		log.Printf("[ERROR] failed to get file body for %s: %v", fileID, err)
		return
	}
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		log.Printf("[ERROR] failed to read file body for %s: %v", fileID, err)
		return
	}
	log.Printf("[DEBUG] downloaded file %s", fileID)

	fileURL, err := e.storage.CreateFile(fileID, bodyBytes)
	if err != nil {
		log.Printf("[ERROR] failed to create file %s: %v", fileID, err)
		return
	}

	e.fileIDToURL[fileID] = fileURL
	// hacky: need to pass fileURL to template
	// using this map in template FuncMap later

	if fileType == "" {
		return
	}

	converter, found := e.converters[fileType]
	if !found {
		log.Printf("[DEBUG] no convertion will happen (converter for type \"%s\" not found)", fileType)
		return
	}

	log.Printf("[DEBUG] converting file %s (%s)", fileID, fileType)

	err = converter.Convert(fileID)
	if err != nil {
		log.Printf("[ERROR] failed to convert file %s: %v", fileID, err)
	}

	return
}

func readMessages(path string) ([]bot.Message, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	messages := []bot.Message{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		msg := bot.Message{}
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

func filter(msg bot.Message) bool {
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

func sizeHuman(b int) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf(
		"%.1f %cB",
		float64(b)/float64(div),
		"kMGTPE"[exp],
	)
}
