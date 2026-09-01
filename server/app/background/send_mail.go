package background

import (
	"context"
	"encoding/json"
	"fmt"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"net/smtp"
	"strings"
)

const (
	mailChannel = "mail-message-channel"
)

type MailMessage struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Msg     string   `json:"msg"`
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
			if err := json.Unmarshal([]byte(msg.Payload), &input); err != nil {
				b.Log.Error("failed decode EmailMessage", "error", err)
				continue
			}
			b.send(input)
		}
	}
}

func (b *EmailBroker) send(mail MailMessage) {
	msg := b.buildMessage(mail)
	var auth smtp.Auth
	if b.Password == "" {
		auth = nil
	} else {
		auth = smtp.PlainAuth("", b.From, b.Password, b.Host)
	}
	if err := smtp.SendMail(fmt.Sprintf("%s:%d", b.Host, b.Port), auth, b.From, mail.To, []byte(msg)); err != nil {
		b.Log.Error("failed to send mail", "error", err)
	}
}

func (b *EmailBroker) buildMessage(mail MailMessage) string {
	msgs := []string{
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";",
		fmt.Sprintf("From: %s", b.From),
		fmt.Sprintf("To: %s", strings.Join(mail.To, ";")),
		fmt.Sprintf("Subject: %s", mail.Subject),
		fmt.Sprintf("\r\n%s\r\n", mail.Msg),
	}
	return strings.Join(msgs, "\r\n")
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
