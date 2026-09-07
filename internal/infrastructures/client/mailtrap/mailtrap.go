package mailtrap

import (
	"context"
	"fmt"

	"github.com/mailtrap/mailtrap-go"

	"email-job/internal/config"
)

type Client struct {
	client *mailtrap.Client
	cfg    config.AppConfig
}

func NewMailTrapClient(cfg config.AppConfig) (*Client, error) {
	client, err := mailtrap.NewClient(cfg.MailTrapApiKey)
	if err != nil {
		return nil, fmt.Errorf("mailtrap: failed to create client %w", err)
	}

	return &Client{
		client: client,
		cfg:    cfg,
	}, nil
}

func (c *Client) SendEmail(ctx context.Context, to, subject, htmlBody string) ([]string, error) {
	resp, _, err := c.client.Send(ctx, &mailtrap.SendRequest{
		From:     mailtrap.Address{Email: c.cfg.MailTrapRole + "@tyomaaans.cloud", Name: c.cfg.MailTrapName},
		To:       []mailtrap.Address{{Email: to}},
		Subject:  subject,
		HTML:     htmlBody,
		Category: "Transactional",
	})
	if err != nil {
		return nil, fmt.Errorf("mailtrap: failed to send email %w", err)
	}

	return resp.MessageIDs, nil
}