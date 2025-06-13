package middleware

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func HMACValidator(secrets map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		org := c.Param("organisation")
		secret, exists := secrets[org]
		if !exists {
			c.Header("X-Capnhook", "no secret found")
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		// Read body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
			c.Abort()
			return
		}

		// Store body for later use
		c.Set("raw_body", body)

		// Validate HMAC
		signature := c.GetHeader("X-Hub-Signature")
		if signature == "" {
			signature = c.GetHeader("X-Koan-Signature")
		}

		if !validateHMAC(signature, secret, body) {
			c.Header("X-Capnhook", "invalid hmac")
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid hmac"})
			c.Abort()
			return
		}

		c.Set("hmac_valid", true)
		c.Next()
	}
}

func validateHMAC(signature, secret string, body []byte) bool {
	if signature == "" {
		return false
	}

	parts := strings.SplitN(signature, "=", 2)
	if len(parts) != 2 || parts[0] != "sha1" {
		return false
	}

	expectedMAC := hmac.New(sha1.New, []byte(secret))
	expectedMAC.Write(body)
	expectedSignature := hex.EncodeToString(expectedMAC.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(parts[1]), []byte(expectedSignature)) == 1
}
