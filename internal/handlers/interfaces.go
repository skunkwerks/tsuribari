package handlers

import "tsuribari/internal/models"

type Storage interface {
	StoreWebhook(headers map[string]string, body []byte) (*models.WebhookDoc, error)
}

type Queue interface {
	PublishWorkflow(workflow *models.Workflow) error
	Close() error
}
