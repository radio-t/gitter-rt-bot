package events

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	log "github.com/go-pkgz/lgr"
)

var pinned = []string{
	" --> Официальный кат!",
}

// Rtjc is a listener for incoming rtjc commands. Publishes whatever it got from the socket
// compatible with the legacy rtjc bot. Primarily use case is to push news events from news.radio-t.com
type Rtjc struct {
	Port      int
	Submitter Submitter
}

// Submitter defines interface to submit (usually asynchronously) to the chat
type Submitter interface {
	Submit(ctx context.Context, text string, pin bool) error
}

// Listen on Port accept and forward to telegram
func (l Rtjc) Listen(ctx context.Context) {
	log.Printf("[INFO] rtjc listener on port %d", l.Port)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", l.Port))
	if err != nil {
		log.Fatalf("[ERROR] can't listen on %d, %v", l.Port, err)
	}

	for {
		conn, e := ln.Accept()
		if e != nil {
			log.Printf("[WARN] can't accept, %v", e)
			time.Sleep(time.Second * 1)
			continue
		}
		if message, rerr := bufio.NewReader(conn).ReadString('\n'); rerr == nil {
			if serr := l.Submitter.Submit(ctx, message, l.isPinned(message)); serr != nil {
				log.Printf("[WARN] can't send message, %v", serr)
			}
		} else {
			log.Printf("[WARN] can't read message, %v", rerr)
		}
		_ = conn.Close()
	}
}

func (l Rtjc) isPinned(msg string) bool {
	for _, p := range pinned {
		if strings.EqualFold(msg, p) {
			return true
		}
	}
	return false
}
