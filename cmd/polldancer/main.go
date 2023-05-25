package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/slack-go/slack"
	"github.com/sony/gobreaker"
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
	RetryMaxAttempts = 3
)

type PollingService interface {
	Poll() ([]byte, error)
}

type WebhookService interface {
	Send(body []byte, mimeType string) error
}

type SlackService interface {
	SendMessage(message string)
}

type pollAndForwardHandler struct {
	pollingService PollingService
	webhookService WebhookService
	slackService   SlackService
	policyFunc     func([]byte) bool
	logger         *zap.SugaredLogger
	circuitBreaker *gobreaker.CircuitBreaker
}

func (h *pollAndForwardHandler) Run(ctx context.Context) {
	ticker := time.NewTicker(PollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.Debugw("Polling cancelled",
				"module", "polling",
			)
			return
		case <-ticker.C:
			h.execute()
		}
	}
}

func (h *pollAndForwardHandler) execute() {
	_, err := h.circuitBreaker.Execute(func() (interface{}, error) {
		data, err := h.pollingService.Poll()
		if err != nil {
			h.logger.Errorw("Error in poll and forward",
				"module", "main",
				"error", err,
			)
			h.slackService.SendMessage(fmt.Sprintf("Error in poll and forward: %v\n", err))
			return nil, err
		}

		if h.policyFunc(data) {
			err := h.webhookService.Send(data, ExpectedMimeType)
			if err != nil {
				h.logger.Errorw("Error sending to webhook",
					"module", "main",
					"error", err,
				)
				h.slackService.SendMessage(fmt.Sprintf("Error sending to webhook: %v", err))
				return nil, err
			}
		}

		return nil, nil
	})

	if err != nil {
		h.logger.Errorw("Circuit breaker tripped",
			"module", "main",
			"error", err,
		)
		h.slackService.SendMessage("Circuit breaker tripped. Service temporarily unavailable.")
	}
}

type HttpPollingService struct {
	logger *zap.SugaredLogger
}

func (s *HttpPollingService) Poll() ([]byte, error) {
	resp, err := http.Get(PollingURL)
	if err != nil {
		s.logger.Errorw("Error polling",
			"URL", PollingURL,
			"error", err,
		)
		return nil, fmt.Errorf("error polling %s: %v", PollingURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		s.logger.Warnw("Non-OK HTTP status",
			"URL", PollingURL,
			"statusCode", resp.StatusCode,
			"responseBody", string(body),
		)
		return nil, fmt.Errorf("non-OK HTTP status from %s: %d", PollingURL, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.logger.Errorw("Error reading response body",
			"error", err,
		)
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	return body, nil
}

type HttpWebhookService struct {
	logger *zap.SugaredLogger
}

func (s *HttpWebhookService) Send(body []byte, mimeType string) error {
	resp, err := http.Post(WebhookURL, mimeType, bytes.NewReader(body))
	if err != nil {
		s.logger.Errorw("Error sending to webhook",
			"URL", WebhookURL,
			"error", err,
		)
		return fmt.Errorf("error sending to webhook %s: %v", WebhookURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		s.logger.Warnw("Non-OK HTTP status from webhook",
			"URL", WebhookURL,
			"statusCode", resp.StatusCode,
			"responseBody", string(body),
		)
		return fmt.Errorf("non-OK HTTP status from webhook %s: %d\n%s", WebhookURL, resp.StatusCode, string(body))
	}

	return nil
}

type SlackNotificationService struct {
	slackClient *slack.Client
	logger      *zap.SugaredLogger
}

func (s *SlackNotificationService) SendMessage(message string) {
	if SlackToken != "" {
		_, _, err := s.slackClient.PostMessage(
			SlackChannel,
			slack.MsgOptionText(message, false),
		)
		if err != nil {
			s.logger.Errorw("Error sending Slack message",
				"error", err,
			)
		}
	}
}

func shouldForward(data []byte) bool {
	// Add your policy logic here
	// For example, you can check specific conditions in the data
	// and return true if it meets the forwarding criteria, or false otherwise
	return true
}

func main() {
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
	defer logger.Sync()
	sugar := logger.Sugar()

	slackClient := slack.New(SlackToken)

	pollingService := &HttpPollingService{
		logger: sugar,
	}
	webhookService := &HttpWebhookService{
		logger: sugar,
	}
	slackService := &SlackNotificationService{
		slackClient: slackClient,
		logger:      sugar,
	}

	breaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "polling",
		MaxRequests: RetryMaxAttempts,
		Timeout:     5 * time.Second,
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			sugar.Infow("Circuit breaker state changed",
				"name", name,
				"from", from,
				"to", to,
			)
		},
	})

	handler := &pollAndForwardHandler{
		pollingService: pollingService,
		webhookService: webhookService,
		slackService:   slackService,
		policyFunc:     shouldForward,
		logger:         sugar,
		circuitBreaker: breaker,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go handler.Run(ctx)

	<-ctx.Done()
}
