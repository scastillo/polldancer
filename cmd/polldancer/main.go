package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	PollingURL       = "https://jsonplaceholder.typicode.com/todos/1"
	WebhookURL       = "http://localhost:8080"
	PollingInterval  = 5 * time.Second
	ExpectedMimeType = "application/json"
	SlackToken       = ""
	SlackChannel     = "#channel-name"
)

func main() {
	// Initialize logger with debug level, console and file outputs
	logger, _ := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: true,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.FullCallerEncoder,
		},
		OutputPaths:      []string{"stdout", "polldancer.log"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	// Initialize Slack client
	slackClient := slack.New(SlackToken)

	// Create a ticker for polling interval
	ticker := time.NewTicker(PollingInterval)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				sugar.Debugw("Polling cancelled",
					// Structured context as loosely typed key-value pairs.
					"module", "polling",
				)
				return
			case <-ticker.C:
				err := pollAndForward(sugar, slackClient)
				if err != nil {
					sugar.Errorw("Error in poll and forward",
						"module", "main",
						"error", err,
					)
					sendSlackError(slackClient, fmt.Sprintf("Error in poll and forward: %v\n", err))
				}
			}
		}
	}()

	// Block main goroutine until it's cancelled
	<-ctx.Done()
}

func pollAndForward(sugar *zap.SugaredLogger, slackClient *slack.Client) error {
	resp, err := http.Get(PollingURL)
	if err != nil {
		return fmt.Errorf("error polling %s: %v", PollingURL, err)
	}
	defer resp.Body.Close()

	mimeType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(mimeType, ExpectedMimeType) {
		return fmt.Errorf("unexpected Content-Type, expected %s but got %s", ExpectedMimeType, mimeType)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	err = sendToWebhook(sugar, mimeType, body)
	if err != nil {
		return fmt.Errorf("error sending to webhook: %v", err)
	}

	return nil
}

func sendToWebhook(sugar *zap.SugaredLogger, mimeType string, body []byte) error {
	webhookResp, err := http.Post(WebhookURL, mimeType, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error sending to webhook %s: %v", WebhookURL, err)
	}
	defer webhookResp.Body.Close()

	if webhookResp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(webhookResp.Body)
		return fmt.Errorf("non-OK HTTP status from webhook %s: %d\n%s", WebhookURL, webhookResp.StatusCode, string(body))
	}

	return nil
}

func sendSlackError(slackClient *slack.Client, message string) {
	if SlackToken != "" {
		slackClient.PostMessage(
			SlackChannel,
			slack.MsgOptionText(message, false),
		)
	}
}
