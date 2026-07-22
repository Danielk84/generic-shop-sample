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
	mailChannel = "mail-message-channel"
)

type MailMessage struct {
	To  string `json:"to"`
	Msg []byte `json:"msg"`
}

type EmailBroker struct {
	Cache    cache.CacheClient
	From     string
	Host     string
	Port     int
	Password string
	Log      logger.Logger
}

func (b *EmailBroker) Start(ctx context.Context) {
	sub := b.Cache.Subscribe(ctx, mailChannel)
	defer func() { _ = sub.Close() }()

	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				b.Log.Warn("mail subscription closed")
				return
			}
			var input MailMessage
			if err := json.NewDecoder(bytes.NewBufferString(msg.Payload)).Decode(&input); err != nil {
				b.Log.Error("failed decode EmailMessage", "error", err)
				continue
			}
			b.send(input.To, input.Msg)
		}
	}
}

func (b *EmailBroker) send(to string, msg []byte) {
	auth := smtp.PlainAuth("", b.From, b.Password, b.Host)
	if err := smtp.SendMail(fmt.Sprintf("%s:%d", b.Host, b.Port), auth, b.From, []string{to}, msg); err != nil {
		b.Log.Error("failed to send mail", "error", err)
	}
}

var _ BackgroundTask = (*EmailBroker)(nil)

func SendMail(ctx context.Context, cache cache.CacheClient, mail MailMessage) error {
	msg, err := json.Marshal(mail)
	if err != nil {
		return fmt.Errorf("failed to encode MailMessage, %s", err)
	}
	if err := cache.Publish(ctx, mailChannel, msg).Err(); err != nil {
		return fmt.Errorf(`failed to publish msg to channel "%s", %s`, mailChannel, err)
	}
	return nil
}
