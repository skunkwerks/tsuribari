package models

import (
	"testing"
	"time"
)

func TestTransformWebhookToWorkflow_Success(t *testing.T) {
	// Create a valid webhook document
	doc := &WebhookDoc{
		ID:  "test-doc-id",
		UTC: time.Now().UTC(),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"repository": map[string]interface{}{
				"ssh_url": "git@github.com:testorg/testrepo.git",
				"owner": map[string]interface{}{
					"login": "testorg",
				},
			},
			"head_commit": map[string]interface{}{
				"id": "abc123def456789",
			},
		},
	}

	workflow := TransformWebhookToWorkflow(doc)

	if workflow == nil {
		t.Fatal("Expected workflow to be created, got nil")
	}

	if workflow.ID != "test-doc-id" {
		t.Errorf("Expected ID 'test-doc-id', got '%s'", workflow.ID)
	}

	if workflow.Ref != "abc123def456789" {
		t.Errorf("Expected Ref 'abc123def456789', got '%s'", workflow.Ref)
	}

	if workflow.URL != "git@github.com:testorg/testrepo.git" {
		t.Errorf("Expected URL 'git@github.com:testorg/testrepo.git', got '%s'", workflow.URL)
	}

	if workflow.Org != "testorg" {
		t.Errorf("Expected Org 'testorg', got '%s'", workflow.Org)
	}

	if workflow.Cache == "" {
		t.Error("Expected Cache to be set")
	}

	// Verify cache is SHA256 hash of URL
	expectedCacheLength := 64 // SHA256 hex string length
	if len(workflow.Cache) != expectedCacheLength {
		t.Errorf("Expected cache length %d, got %d", expectedCacheLength, len(workflow.Cache))
	}

	if workflow.UTC != doc.UTC {
		t.Error("Expected UTC to match document UTC")
	}
}

func TestTransformWebhookToWorkflow_MissingRepository(t *testing.T) {
	doc := &WebhookDoc{
		ID:  "test-doc-id",
		UTC: time.Now().UTC(),
		Body: map[string]interface{}{
			"head_commit": map[string]interface{}{
				"id": "abc123def456789",
			},
		},
	}

	workflow := TransformWebhookToWorkflow(doc)

	if workflow != nil {
		t.Error("Expected nil workflow when repository is missing")
	}
}

func TestTransformWebhookToWorkflow_MissingOwner(t *testing.T) {
	doc := &WebhookDoc{
		ID:  "test-doc-id",
		UTC: time.Now().UTC(),
		Body: map[string]interface{}{
			"repository": map[string]interface{}{
				"ssh_url": "git@github.com:testorg/testrepo.git",
			},
			"head_commit": map[string]interface{}{
				"id": "abc123def456789",
			},
		},
	}

	workflow := TransformWebhookToWorkflow(doc)

	if workflow != nil {
		t.Error("Expected nil workflow when owner is missing")
	}
}

func TestTransformWebhookToWorkflow_MissingHeadCommit(t *testing.T) {
	doc := &WebhookDoc{
		ID:  "test-doc-id",
		UTC: time.Now().UTC(),
		Body: map[string]interface{}{
			"repository": map[string]interface{}{
				"ssh_url": "git@github.com:testorg/testrepo.git",
				"owner": map[string]interface{}{
					"login": "testorg",
				},
			},
		},
	}

	workflow := TransformWebhookToWorkflow(doc)

	if workflow != nil {
		t.Error("Expected nil workflow when head_commit is missing")
	}
}

func TestTransformWebhookToWorkflow_EmptyFields(t *testing.T) {
	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "Empty SSH URL",
			body: map[string]interface{}{
				"repository": map[string]interface{}{
					"ssh_url": "",
					"owner": map[string]interface{}{
						"login": "testorg",
					},
				},
				"head_commit": map[string]interface{}{
					"id": "abc123def456789",
				},
			},
		},
		{
			name: "Empty org name",
			body: map[string]interface{}{
				"repository": map[string]interface{}{
					"ssh_url": "git@github.com:testorg/testrepo.git",
					"owner": map[string]interface{}{
						"login": "",
					},
				},
				"head_commit": map[string]interface{}{
					"id": "abc123def456789",
				},
			},
		},
		{
			name: "Empty commit ID",
			body: map[string]interface{}{
				"repository": map[string]interface{}{
					"ssh_url": "git@github.com:testorg/testrepo.git",
					"owner": map[string]interface{}{
						"login": "testorg",
					},
				},
				"head_commit": map[string]interface{}{
					"id": "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &WebhookDoc{
				ID:   "test-doc-id",
				UTC:  time.Now().UTC(),
				Body: tt.body,
			}

			workflow := TransformWebhookToWorkflow(doc)

			if workflow != nil {
				t.Error("Expected nil workflow when required fields are empty")
			}
		})
	}
}

func TestTransformWebhookToWorkflow_WrongTypes(t *testing.T) {
	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "Repository not map",
			body: map[string]interface{}{
				"repository": "not a map",
				"head_commit": map[string]interface{}{
					"id": "abc123def456789",
				},
			},
		},
		{
			name: "Owner not map",
			body: map[string]interface{}{
				"repository": map[string]interface{}{
					"ssh_url": "git@github.com:testorg/testrepo.git",
					"owner":   "not a map",
				},
				"head_commit": map[string]interface{}{
					"id": "abc123def456789",
				},
			},
		},
		{
			name: "Head commit not map",
			body: map[string]interface{}{
				"repository": map[string]interface{}{
					"ssh_url": "git@github.com:testorg/testrepo.git",
					"owner": map[string]interface{}{
						"login": "testorg",
					},
				},
				"head_commit": "not a map",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &WebhookDoc{
				ID:   "test-doc-id",
				UTC:  time.Now().UTC(),
				Body: tt.body,
			}

			workflow := TransformWebhookToWorkflow(doc)

			if workflow != nil {
				t.Error("Expected nil workflow when fields have wrong types")
			}
		})
	}
}

func TestGetKeys(t *testing.T) {
	testMap := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	keys := getKeys(testMap)

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Check that all expected keys are present
	expectedKeys := map[string]bool{
		"key1": false,
		"key2": false,
		"key3": false,
	}

	for _, key := range keys {
		if _, exists := expectedKeys[key]; exists {
			expectedKeys[key] = true
		} else {
			t.Errorf("Unexpected key: %s", key)
		}
	}

	// Check that all expected keys were found
	for key, found := range expectedKeys {
		if !found {
			t.Errorf("Expected key not found: %s", key)
		}
	}
}

func TestGetKeys_EmptyMap(t *testing.T) {
	testMap := map[string]interface{}{}

	keys := getKeys(testMap)

	if len(keys) != 0 {
		t.Errorf("Expected 0 keys for empty map, got %d", len(keys))
	}
}
