package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIPFilter_TrustedIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	trustedIPs := []string{"192.168.1.100", "10.0.0.0/8"}

	tests := []struct {
		name           string
		clientIP       string
		remoteAddr     string
		xRealIP        string
		xForwardedFor  string
		expectedStatus int
	}{
		{
			name:           "Direct trusted IP",
			remoteAddr:     "192.168.1.100:12345",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "X-Real-IP trusted",
			remoteAddr:     "1.2.3.4:12345",
			xRealIP:        "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "X-Forwarded-For trusted",
			remoteAddr:     "1.2.3.4:12345",
			xForwardedFor:  "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "CIDR range match",
			remoteAddr:     "10.1.2.3:12345",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Untrusted IP",
			remoteAddr:     "1.2.3.4:12345",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "X-Real-IP untrusted",
			remoteAddr:     "192.168.1.100:12345",
			xRealIP:        "1.2.3.4",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}

			c.Request = req

			// Add a handler that sets status OK if middleware passes
			handler := IPFilter(trustedIPs)
			handler(c)

			if !c.IsAborted() {
				c.Status(http.StatusOK)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusForbidden {
				if w.Header().Get("X-Capnhook") != "invalid source ip" {
					t.Error("Expected X-Capnhook header for forbidden request")
				}
			}

			if tt.expectedStatus == http.StatusOK {
				trusted, exists := c.Get("trusted_ip")
				if !exists || trusted != true {
					t.Error("Expected trusted_ip to be set to true")
				}
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		remoteAddr    string
		xRealIP       string
		xForwardedFor string
		expectedIP    string
	}{
		{
			name:          "X-Real-IP priority",
			remoteAddr:    "1.2.3.4:12345",
			xRealIP:       "192.168.1.100",
			xForwardedFor: "10.0.0.1",
			expectedIP:    "192.168.1.100",
		},
		{
			name:          "X-Forwarded-For fallback",
			remoteAddr:    "1.2.3.4:12345",
			xForwardedFor: "10.0.0.1",
			expectedIP:    "10.0.0.1",
		},
		{
			name:       "RemoteAddr fallback",
			remoteAddr: "1.2.3.4:12345",
			expectedIP: "1.2.3.4",
		},
		{
			name:       "IPv6 RemoteAddr",
			remoteAddr: "[::1]:12345",
			expectedIP: "::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}

			c.Request = req

			ip := getClientIP(c)
			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

func TestIsTrustedIP(t *testing.T) {
	trustedIPs := []string{
		"192.168.1.100",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"::1",
		"2001:db8::/32",
	}

	tests := []struct {
		name     string
		clientIP string
		expected bool
	}{
		{"Exact match IPv4", "192.168.1.100", true},
		{"CIDR match 10.x", "10.1.2.3", true},
		{"CIDR match 172.16.x", "172.16.1.1", true},
		{"CIDR no match", "172.15.1.1", false},
		{"No match", "1.2.3.4", false},
		{"IPv6 exact", "::1", true},
		{"IPv6 CIDR match", "2001:db8::1", true},
		{"IPv6 no match", "2001:db9::1", false},
		{"Invalid IP", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTrustedIP(tt.clientIP, trustedIPs)
			if result != tt.expected {
				t.Errorf("Expected %v for IP %s, got %v", tt.expected, tt.clientIP, result)
			}
		})
	}
}
