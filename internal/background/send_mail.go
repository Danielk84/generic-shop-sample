package background

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"generic-shop-sample/db"
	"log/slog"
	"net/smtp"
)

const (
	mailChannel = "mailMessage"
)

type MailMessage struct {
	To  string `json:"to"`
	Msg []byte `json:"msg"`
}

type emailBroker struct {
	cache    db.CacheClient
	from     string
	host     string
	port     int
	password string
}

func (eb *emailBroker) start(ctx context.Context) {
	sub := eb.cache.Subscribe(ctx, mailChannel).Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-sub:
			if !ok {
				slog.Warn("mail subscription closed")
				return
			}
			var input MailMessage
			if err := json.NewDecoder(bytes.NewBufferString(msg.Payload)).Decode(&input); err != nil {
				slog.Error("failed decode EmailMessage", "error", err)
				continue
			}
			eb.send(input.To, input.Msg)
		}
	}
}

func (eb *emailBroker) send(to string, msg []byte) {
	auth := smtp.PlainAuth("", eb.from, eb.password, eb.host)
	if err := smtp.SendMail(fmt.Sprintf("%s:%d", eb.host, eb.port), auth, eb.from, []string{to}, msg); err != nil {
		slog.Error("failed to send mail", "error", err)
	}
}

func SendMail(ctx context.Context, cache db.CacheClient, mail *MailMessage) error {
	msg, err := json.Marshal(&mail)
	if err != nil {
		return fmt.Errorf("failed to encode MailMessage, %s", err)
	}
	if _, err := cache.Publish(ctx, mailChannel, msg).Result(); err != nil {
		return fmt.Errorf(`failed to publish msg to channel "%s", %s`, mailChannel, err)
	}
	return nil
}
