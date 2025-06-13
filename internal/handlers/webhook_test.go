package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"tsuribari/internal/models"
)

// Mock storage
type MockStorage struct {
	storeWebhookFunc func(headers map[string]string, body []byte) (*models.WebhookDoc, error)
}

func (m *MockStorage) StoreWebhook(headers map[string]string, body []byte) (*models.WebhookDoc, error) {
	if m.storeWebhookFunc != nil {
		return m.storeWebhookFunc(headers, body)
	}
	return nil, errors.New("not implemented")
}

// Mock queue
type MockQueue struct {
	publishWorkflowFunc func(workflow *models.Workflow) error
}

func (m *MockQueue) PublishWorkflow(workflow *models.Workflow) error {
	if m.publishWorkflowFunc != nil {
		return m.publishWorkflowFunc(workflow)
	}
	return errors.New("not implemented")
}

func (m *MockQueue) Close() error {
	return nil
}

func TestHandleWebhook_ResponseContainsDocID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMocks     func(*MockStorage, *MockQueue)
		expectedStatus int
		expectedDocID  string
	}{
		{
			name: "Success response includes doc_id",
			setupMocks: func(storage *MockStorage, queue *MockQueue) {
				doc := &models.WebhookDoc{
					ID: "test-doc-123",
					Body: map[string]interface{}{
						"repository": map[string]interface{}{
							"ssh_url": "git@github.com:test/repo.git",
							"owner":   map[string]interface{}{"login": "test"},
						},
						"head_commit": map[string]interface{}{"id": "abc123"},
					},
				}
				storage.storeWebhookFunc = func(headers map[string]string, body []byte) (*models.WebhookDoc, error) {
					return doc, nil
				}
				queue.publishWorkflowFunc = func(workflow *models.Workflow) error {
					return nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedDocID:  "test-doc-123",
		},
		{
			name: "Transform failure response includes doc_id",
			setupMocks: func(storage *MockStorage, queue *MockQueue) {
				doc := &models.WebhookDoc{
					ID:   "test-doc-456",
					Body: map[string]interface{}{"invalid": "structure"},
				}
				storage.storeWebhookFunc = func(headers map[string]string, body []byte) (*models.WebhookDoc, error) {
					return doc, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedDocID:  "test-doc-456",
		},
		{
			name: "Queue failure response includes doc_id",
			setupMocks: func(storage *MockStorage, queue *MockQueue) {
				doc := &models.WebhookDoc{
					ID: "test-doc-789",
					Body: map[string]interface{}{
						"repository": map[string]interface{}{
							"ssh_url": "git@github.com:test/repo.git",
							"owner":   map[string]interface{}{"login": "test"},
						},
						"head_commit": map[string]interface{}{"id": "abc123"},
					},
				}
				storage.storeWebhookFunc = func(headers map[string]string, body []byte) (*models.WebhookDoc, error) {
					return doc, nil
				}
				queue.publishWorkflowFunc = func(workflow *models.Workflow) error {
					return errors.New("queue error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedDocID:  "test-doc-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockStorage{}
			mockQueue := &MockQueue{}
			handler := NewWebhookHandler(mockStorage, mockQueue)

			tt.setupMocks(mockStorage, mockQueue)

			body := `{"test": "data"}`
			req := httptest.NewRequest("POST", "/webhooks/test", bytes.NewBufferString(body))
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("raw_body", []byte(body))

			handler.HandleWebhook(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response["id"] != tt.expectedDocID {
				t.Errorf("Expected doc_id %s, got %v", tt.expectedDocID, response["id"])
			}
		})
	}
}
