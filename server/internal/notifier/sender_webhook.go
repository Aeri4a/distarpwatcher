package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"server/internal/analyzer"
	"server/internal/database"
	"strings"
)

type WebhookSender struct{}

type WebhookPayload struct {
	Agent   string `json:"agent_id"`
	Message string `json:"message"`
}

func (sen *WebhookSender) SendAlert(ctx context.Context, channel database.NotificationChannel, report *analyzer.AnalysisReport) error {
	url := channel.Target

	payload := WebhookPayload{
		Agent:   report.Event.AgentId,
		Message: strings.Join(report.Findings, "\n"),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("Alert sent to [%s] channel.", channel.Name)

	return nil
}
