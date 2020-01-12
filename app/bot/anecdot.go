package bot

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/go-pkgz/lgr"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Anecdote bot, returns from http://rzhunemogu.ru/RandJSON.aspx?CType=1
type Anecdote struct{}

// NewAnecdote makes a bot for http://rzhunemogu.ru
func NewAnecdote() *Anecdote {
	log.Printf("[INFO] anecdote bot with http://rzhunemogu.ru/RandJSON.aspx?CType=1 and http://api.icndb.com/jokes/random")
	return &Anecdote{}
}

// OnMessage returns one entry
func (a Anecdote) OnMessage(msg Message) (response string, answer bool) {

	if !contains(a.ReactOn(), msg.Text) {
		return "", false
	}

	if contains([]string{"chuck!"}, msg.Text) {
		return a.chuck()
	}

	return a.rzhunemogu()
}

func (a Anecdote) rzhunemogu() (response string, answer bool) {
	reqURL := "http://rzhunemogu.ru/RandJSON.aspx?CType=1"

	client := http.Client{Timeout: time.Second * 5}
	req, err := makeHTTPRequest(reqURL)
	if err != nil {
		log.Printf("[WARN] failed to make request %s, error=%v", reqURL, err)
		return "", false
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[WARN] failed to send request %s, error=%v", reqURL, err)
		return "", false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[WARN] failed to read body, error=%v", err)
		return "", false
	}

	text := string(body)
	// this json is not really json? body with \r
	text = strings.TrimLeft(text, `{"content":"`)
	text = strings.TrimRight(text, `"}`)

	tr := transform.NewReader(strings.NewReader(text), charmap.Windows1251.NewDecoder())
	buf, err := ioutil.ReadAll(tr)
	if err != nil {
		log.Printf("[WARN] failed to convert string to utf, error=%v", err)
		return "", false
	}

	return string(buf), true
}

func (a Anecdote) chuck() (response string, answer bool) {

	chuckResp := struct {
		Type  string
		Value struct {
			Categories []string
			Joke       string
		}
	}{}

	client := http.Client{Timeout: time.Second * 5}
	reqURL := "http://api.icndb.com/jokes/random"
	req, err := makeHTTPRequest(reqURL)
	if err != nil {
		log.Printf("[WARN] failed to make request %s, error=%v", reqURL, err)
		return "", false
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[WARN] failed to send request %s, error=%v", reqURL, err)
		return "", false
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&chuckResp); err != nil {
		log.Printf("[WARN] failed to convert from json, error=%v", err)
		return "", false
	}
	return "- " + strings.Replace(chuckResp.Value.Joke, "&quot;", "\"", -1), true
}

// ReactOn keys
func (a Anecdote) ReactOn() []string {
	return []string{"анекдот!", "анкедот!", "joke!", "chuck!"}
}
