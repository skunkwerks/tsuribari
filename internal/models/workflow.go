package models

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"time"
)

type Workflow struct {
	ID    string    `json:"id"`
	Ref   string    `json:"ref"`
	URL   string    `json:"url"`
	Org   string    `json:"org"`
	Cache string    `json:"cache"`
	UTC   time.Time `json:"utc"`
}

type WebhookDoc struct {
	ID      string                 `json:"_id"`
	UTC     time.Time              `json:"utc"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body,omitempty"`
}

func TransformWebhookToWorkflow(doc *WebhookDoc) *Workflow {
	body := doc.Body

	log.Printf("DEBUG: transforming doc ID: %s", doc.ID)

	// Extract repository info (GitHub format)
	repo, ok := body["repository"].(map[string]interface{})
	if !ok {
		log.Printf("DEBUG: has no repository")
		return nil
	}

	owner, ok := repo["owner"].(map[string]interface{})
	if !ok {
		log.Printf("DEBUG: has no owner")
		return nil
	}

	headCommit, ok := body["head_commit"].(map[string]interface{})
	if !ok {
		log.Printf("DEBUG: has no head_commit")
		return nil
	}

	sshURL, _ := repo["ssh_url"].(string)
	orgName, _ := owner["login"].(string)
	commitID, _ := headCommit["id"].(string)

	log.Printf("DEBUG:  org: '%s', commit: '%s', url: '%s'", orgName, commitID, sshURL)

	if sshURL == "" || orgName == "" || commitID == "" {
		log.Printf("DEBUG: empty fields in webhook body")
		return nil
	}

	// Generate cache hash
	hash := sha256.Sum256([]byte(sshURL))
	cache := hex.EncodeToString(hash[:])

	workflow := &Workflow{
		ID:    doc.ID,
		Ref:   commitID,
		URL:   sshURL,
		Org:   orgName,
		Cache: cache,
		UTC:   doc.UTC,
	}

	log.Printf("INFO: created workflow: %+v", workflow)
	return workflow
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
