package storage

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/go-kivik/couchdb/v3"
	"github.com/go-kivik/kivik/v3"

	"tsuribari/internal/models"
)

type CouchDB struct {
	client *kivik.Client
	db     *kivik.DB
}

func NewCouchDB(url, database string) (*CouchDB, error) {
	client, err := kivik.New("couch", url)
	if err != nil {
		return nil, err
	}

	db := client.DB(context.Background(), database)

	return &CouchDB{
		client: client,
		db:     db,
	}, nil
}

func (c *CouchDB) StoreWebhook(headers map[string]string, body []byte) (*models.WebhookDoc, error) {
	// Generate document ID from body hash
	hash := sha1.Sum(body)
	docID := hex.EncodeToString(hash[:])

	// Parse JSON body
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(body, &bodyMap); err != nil {
		return nil, err
	}

	doc := &models.WebhookDoc{
		ID:      docID,
		UTC:     time.Now().UTC(),
		Headers: headers,
		Body:    bodyMap,
	}

	_, err := c.db.Put(context.Background(), docID, doc)
	if err != nil {
		// for 409 conflicts, return the document anyway since it
		// already exists with the correct checksum
		if kivik.StatusCode(err) == http.StatusConflict {
			log.Printf("409 conflict from webhook with doc.id: %s", docID)
			return doc, nil
		}

		// Log other errors for debugging
		log.Printf("ERROR: couchdb: %v (status: %d)", err, kivik.StatusCode(err))
		return nil, err
	}

	return doc, nil
}
