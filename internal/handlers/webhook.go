package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tsuribari/internal/models"
)

type WebhookHandler struct {
	storage Storage
	queue   Queue
}

func NewWebhookHandler(storage Storage, queue Queue) *WebhookHandler {
	return &WebhookHandler{
		storage: storage,
		queue:   queue,
	}
}

func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	// Get raw body from middleware
	rawBody, exists := c.Get("raw_body")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no body found"})
		return
	}

	body := rawBody.([]byte)

	// Extract headers
	headers := make(map[string]string)
	for name, values := range c.Request.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}

	// Store webhook in CouchDB
	doc, err := h.storage.StoreWebhook(headers, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store webhook"})
		return
	}

	// Transform to workflow
	workflow := models.TransformWebhookToWorkflow(doc)
	if workflow == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "webhook stored but cannot transform to workflow",
			"id":      doc.ID,
		})
		return
	}

	// Publish to RabbitMQ
	if err := h.queue.PublishWorkflow(workflow); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to publish workflow",
			"id":    doc.ID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "you have achieved enlightenment",
		"id":      doc.ID,
	})
}
