package events

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	log "github.com/go-pkgz/lgr"
	tbapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"

	"github.com/radio-t/super-bot/app/bot"
	"github.com/radio-t/super-bot/app/reporter"
)

// TelegramListener listens to tg update, forward to bots and send back responses
type TelegramListener struct {
	Terminator
	Token string
	reporter.Reporter
	Bots  bot.Interface
	Group string // can be int64 or public group username (without "@" prefix)
	Debug bool

	botAPI *tbapi.BotAPI
	chatID int64

	msgs struct {
		once sync.Once
		ch   chan bot.Response
	}
}

// Do process all events, blocked call
func (l *TelegramListener) Do(ctx context.Context) (err error) {
	log.Printf("[INFO] start telegram listener for %q", l.Group)

	if l.botAPI, err = tbapi.NewBotAPI(l.Token); err != nil {
		return errors.Wrap(err, "can't make telegram bot")
	}
	if l.chatID, err = l.getChatID(l.Group); err != nil {
		return errors.Wrapf(err, "failed to get chat ID for group %q", l.Group)
	}
	l.botAPI.Debug = l.Debug
	l.msgs.once.Do(func() { l.msgs.ch = make(chan bot.Response, 100) })

	u := tbapi.NewUpdate(0)
	u.Timeout = 60

	var updates tbapi.UpdatesChannel
	if updates, err = l.botAPI.GetUpdatesChan(u); err != nil {
		return errors.Wrap(err, "can't get updates channel")
	}

	for {
		select {

		case <-ctx.Done():
			return ctx.Err()

		case update, ok := <-updates:
			if !ok {
				return errors.Errorf("telegram update chan closed")
			}

			if update.Message == nil {
				log.Print("[DEBUG] empty message body")
				continue
			}

			msgJSON, err := json.Marshal(update.Message)
			if err != nil {
				log.Printf("[ERROR] failed to marshal update.Message to json: %v", err)
				continue
			}
			log.Printf("[DEBUG] %s", string(msgJSON))

			if update.Message.Chat == nil {
				log.Print("[DEBUG] ignoring message not from chat")
				continue
			}

			fromChat := update.Message.Chat.ID

			msg := l.transform(update.Message)
			if fromChat == l.chatID {
				l.Save(msg) // save to report
			}

			log.Printf("[DEBUG] incoming msg: %+v", msg)

			// check for ban
			if b := l.check(msg.From, msg.Sent); b.active {
				if b.new {
					mention := "@" + msg.From.Username
					if msg.From.Username == "" {
						mention = fmt.Sprintf("[%s](tg://user?id=%d)", msg.From.DisplayName, update.Message.From.ID)
					}
					m := fmt.Sprintf("%s _тебя слишком много, отдохни..._", mention)
					tbMsg := tbapi.NewMessage(fromChat, m)
					tbMsg.ParseMode = tbapi.ModeMarkdown
					tbMsg.DisableWebPagePreview = true
					if res, err := l.botAPI.Send(tbMsg); err != nil {
						log.Printf("[WARN] failed to send, %v", err)
					} else {
						l.saveBotMessage(&res, fromChat)
					}

					if err := l.banUser(fromChat, update.Message.From.ID); err != nil {
						log.Printf("[ERROR] failed to ban user %v: %v", msg.From, err)
					}
				}
				continue
			}

			if resp := l.Bots.OnMessage(*msg); resp.Send {
				log.Printf("[DEBUG] bot response - %+v, pin: %t", resp.Text, resp.Pin)
				tbMsg := tbapi.NewMessage(fromChat, resp.Text)
				tbMsg.ParseMode = tbapi.ModeMarkdown
				tbMsg.DisableWebPagePreview = true
				res, err := l.botAPI.Send(tbMsg)
				if err != nil {
					log.Printf("[WARN] can't send tbMsg to telegram, %v", err)
					continue
				}
				l.saveBotMessage(&res, fromChat)
				if resp.Pin {
					_, err = l.botAPI.PinChatMessage(tbapi.PinChatMessageConfig{
						ChatID:              fromChat,
						MessageID:           res.MessageID,
						DisableNotification: false,
					})
					if err != nil {
						log.Printf("[WARN] can't pin tbMsg to telegram, %v", err)
					}
				}
			}

		case resp := <-l.msgs.ch: // publish messages from outside clients
			tbMsg := tbapi.NewMessage(l.chatID, resp.Text)
			tbMsg.ParseMode = tbapi.ModeMarkdown
			res, err := l.botAPI.Send(tbMsg)
			if err != nil {
				log.Printf("[WARN] can't send msg to telegram, %v", err)
				continue
			}
			l.saveBotMessage(&res, l.chatID)
			if resp.Pin {
				_, err = l.botAPI.PinChatMessage(tbapi.PinChatMessageConfig{
					ChatID:              l.chatID,
					MessageID:           res.MessageID,
					DisableNotification: false,
				})
				if err != nil {
					log.Printf("[WARN] can't pin tbMsg to telegram, %v", err)
				}
			}
		}
	}
}

// Submit message text to telegram's group
func (l *TelegramListener) Submit(ctx context.Context, text string, pin bool) error {
	l.msgs.once.Do(func() { l.msgs.ch = make(chan bot.Response, 100) })

	select {
	case <-ctx.Done():
		return ctx.Err()
	case l.msgs.ch <- bot.Response{Text: text, Pin: pin, Send: true}:
	}
	return nil
}

func (l *TelegramListener) getChatID(group string) (int64, error) {
	chatID, err := strconv.ParseInt(l.Group, 10, 64)
	if err == nil {
		return chatID, nil
	}

	chat, err := l.botAPI.GetChat(tbapi.ChatConfig{SuperGroupUsername: "@" + group})
	if err != nil {
		return 0, errors.Wrapf(err, "can't get chat for %s", group)
	}

	return chat.ID, nil
}

func (l *TelegramListener) saveBotMessage(msg *tbapi.Message, fromChat int64) {
	if fromChat != l.chatID {
		return
	}

	l.Save(l.transform(msg))
}

// The bot must be an administrator in the supergroup for this to work
// and must have the appropriate admin rights.
func (l *TelegramListener) banUser(chatID int64, userID int) error {
	banDuration := l.BanDuration

	// From Telegram Bot API documentation:
	// > If user is restricted for more than 366 days or less than 30 seconds from the current time,
	// > they are considered to be restricted forever
	// Because the API query uses unix timestamp rather than "ban duration",
	// you do not want to accidentally get into this 30-second window of a lifetime ban.
	// In practice BanDuration is equal to ten minutes,
	// so this `if` statement is unlikely to be evaluated to true.
	if banDuration < 30*time.Second {
		banDuration = 1 * time.Minute
	}

	resp, err := l.botAPI.RestrictChatMember(tbapi.RestrictChatMemberConfig{
		ChatMemberConfig: tbapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
		UntilDate:             time.Now().Add(banDuration).Unix(),
		CanSendMessages:       new(bool),
		CanSendMediaMessages:  new(bool),
		CanSendOtherMessages:  new(bool),
		CanAddWebPagePreviews: new(bool),
	})
	if err != nil {
		return err
	}

	if !resp.Ok {
		return fmt.Errorf("response is not Ok: %v", string(resp.Result))
	}

	return nil
}

func (l *TelegramListener) transform(msg *tbapi.Message) *bot.Message {
	message := bot.Message{
		ID:   msg.MessageID,
		Sent: msg.Time(),
		Text: msg.Text,
	}

	if msg.From != nil {
		message.From = bot.User{
			Username:    msg.From.UserName,
			DisplayName: msg.From.FirstName + " " + msg.From.LastName,
		}
	}

	switch {
	case msg.Entities != nil && len(*msg.Entities) > 0:
		message.Entities = l.transformEntities(msg.Entities)

	case msg.Photo != nil && len(*msg.Photo) > 0:
		sizes := *msg.Photo
		lastSize := sizes[len(sizes)-1]
		message.Image = &bot.Image{
			FileID:   lastSize.FileID,
			Width:    lastSize.Width,
			Height:   lastSize.Height,
			Caption:  msg.Caption,
			Entities: l.transformEntities(msg.CaptionEntities),
		}
	}

	return &message
}

func (l *TelegramListener) transformEntities(entities *[]tbapi.MessageEntity) *[]bot.Entity {
	if entities == nil || len(*entities) == 0 {
		return nil
	}

	var result []bot.Entity
	for _, entity := range *entities {
		e := bot.Entity{
			Type:   entity.Type,
			Offset: entity.Offset,
			Length: entity.Length,
			URL:    entity.URL,
		}
		if entity.User != nil {
			e.User = &bot.User{
				Username:    entity.User.UserName,
				DisplayName: entity.User.FirstName + " " + entity.User.LastName,
			}
		}
		result = append(result, e)
	}

	return &result
}
