package fcm

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var client *messaging.Client

func Init(credentialsBase64 string) error {
	if credentialsBase64 == "" {
		slog.Warn("fcm.disabled: FIREBASE_CREDENTIALS tidak diset")
		return nil
	}

	creds, err := base64.StdEncoding.DecodeString(credentialsBase64)
	if err != nil {
		return fmt.Errorf("decode firebase credentials: %w", err)
	}

	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsJSON(creds))
	if err != nil {
		return fmt.Errorf("init firebase app: %w", err)
	}

	client, err = app.Messaging(context.Background())
	if err != nil {
		return fmt.Errorf("init fcm messaging: %w", err)
	}

	slog.Info("fcm.ready")
	return nil
}

func Enabled() bool {
	return client != nil
}

func SendToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
	if client == nil || len(tokens) == 0 {
		return nil
	}

	msg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	resp, err := client.SendEachForMulticast(ctx, msg)
	if err != nil {
		return fmt.Errorf("fcm.send: %w", err)
	}

	if resp.FailureCount > 0 {
		slog.Warn("fcm.partial_failure",
			slog.Int("success", resp.SuccessCount),
			slog.Int("failure", resp.FailureCount),
		)
	}
	return nil
}
