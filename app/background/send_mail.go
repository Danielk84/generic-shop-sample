package background

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"net/smtp"
)

const (
	mailChannel = "mailmessage"
)

type MailMessage struct {
	To  string `json:"to"`
	Msg []byte `json:"msg"`
}

type emailBroker struct {
	cache    cache.CacheClient
	from     string
	host     string
	port     int
	password string
	log      logger.Logger
}

func (b *emailBroker) start(ctx context.Context) {
	sub := b.cache.Subscribe(ctx, mailChannel)
	defer sub.Close()

	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				b.log.Warn("mail subscription closed")
				return
			}
			var input MailMessage
			if err := json.NewDecoder(bytes.NewBufferString(msg.Payload)).Decode(&input); err != nil {
				b.log.Error("failed decode EmailMessage", "error", err)
				continue
			}
			b.send(input.To, input.Msg)
		}
	}
}

func (b *emailBroker) send(to string, msg []byte) {
	auth := smtp.PlainAuth("", b.from, b.password, b.host)
	if err := smtp.SendMail(fmt.Sprintf("%s:%d", b.host, b.port), auth, b.from, []string{to}, msg); err != nil {
		b.log.Error("failed to send mail", "error", err)
	}
}

func SendMail(ctx context.Context, cache cache.CacheClient, mail *MailMessage) error {
	msg, err := json.Marshal(&mail)
	if err != nil {
		return fmt.Errorf("failed to encode MailMessage, %s", err)
	}
	if err := cache.Publish(ctx, mailChannel, msg).Err(); err != nil {
		return fmt.Errorf(`failed to publish msg to channel "%s", %s`, mailChannel, err)
	}
	return nil
}
