package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHMACValidator_ValidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secrets := map[string]string{
		"testorg": "testsecret123",
		"demo":    "demosecret456",
	}

	testBody := `{"test": "data"}`

	tests := []struct {
		name           string
		org            string
		body           string
		header         string
		expectedStatus int
	}{
		{
			name:           "Valid X-Hub-Signature",
			org:            "testorg",
			body:           testBody,
			header:         "X-Hub-Signature",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid X-Koan-Signature",
			org:            "demo",
			body:           testBody,
			header:         "X-Koan-Signature",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid signature",
			org:            "testorg",
			body:           testBody,
			header:         "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Unknown organization",
			org:            "unknown",
			body:           testBody,
			header:         "X-Hub-Signature",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set up request
			req := httptest.NewRequest("POST", "/webhooks/"+tt.org, bytes.NewBufferString(tt.body))

			// Generate valid signature if header is specified
			if tt.header != "" && secrets[tt.org] != "" {
				mac := hmac.New(sha1.New, []byte(secrets[tt.org]))
				mac.Write([]byte(tt.body))
				signature := "sha1=" + hex.EncodeToString(mac.Sum(nil))
				req.Header.Set(tt.header, signature)
			} else if tt.header != "" {
				// Invalid signature
				req.Header.Set(tt.header, "sha1=invalidsignature")
			}

			c.Request = req
			c.Params = gin.Params{{Key: "organisation", Value: tt.org}}

			// Run middleware
			handler := HMACValidator(secrets)
			handler(c)

			if !c.IsAborted() {
				c.Status(http.StatusOK)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				// Check that body is stored
				rawBody, exists := c.Get("raw_body")
				if !exists {
					t.Error("Expected raw_body to be set")
				}
				if string(rawBody.([]byte)) != tt.body {
					t.Error("Raw body doesn't match original")
				}

				// Check that hmac_valid is set
				hmacValid, exists := c.Get("hmac_valid")
				if !exists || hmacValid != true {
					t.Error("Expected hmac_valid to be set to true")
				}
			}

			if tt.expectedStatus == http.StatusForbidden {
				capnhookHeader := w.Header().Get("X-Capnhook")
				if tt.org == "unknown" {
					if capnhookHeader != "no secret found" {
						t.Errorf("Expected 'no secret found' header, got '%s'", capnhookHeader)
					}
				} else {
					if capnhookHeader != "invalid hmac" {
						t.Errorf("Expected 'invalid hmac' header, got '%s'", capnhookHeader)
					}
				}
			}
		})
	}
}

func TestValidateHMAC(t *testing.T) {
	secret := "testsecret123"
	body := []byte(`{"test": "data"}`)

	// Generate valid signature
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(body)
	validSignature := "sha1=" + hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name      string
		signature string
		secret    string
		body      []byte
		expected  bool
	}{
		{
			name:      "Valid signature",
			signature: validSignature,
			secret:    secret,
			body:      body,
			expected:  true,
		},
		{
			name:      "Invalid signature",
			signature: "sha1=invalidsignature",
			secret:    secret,
			body:      body,
			expected:  false,
		},
		{
			name:      "Wrong secret",
			signature: validSignature,
			secret:    "wrongsecret",
			body:      body,
			expected:  false,
		},
		{
			name:      "Empty signature",
			signature: "",
			secret:    secret,
			body:      body,
			expected:  false,
		},
		{
			name:      "Wrong algorithm",
			signature: "sha256=somehash",
			secret:    secret,
			body:      body,
			expected:  false,
		},
		{
			name:      "Malformed signature",
			signature: "invalidsignature",
			secret:    secret,
			body:      body,
			expected:  false,
		},
		{
			name:      "Different body",
			signature: validSignature,
			secret:    secret,
			body:      []byte(`{"different": "data"}`),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateHMAC(tt.signature, tt.secret, tt.body)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
