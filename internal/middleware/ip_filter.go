package middleware

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

func IPFilter(trustedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := getClientIP(c)

		if !isTrustedIP(clientIP, trustedIPs) {
			c.Header("X-Capnhook", "invalid source ip")
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		c.Set("trusted_ip", true)
		c.Next()
	}
}

func getClientIP(c *gin.Context) string {
	// Check X-Real-IP header first (from proxy)
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Check X-Forwarded-For header
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		return forwardedFor
	}

	// Fall back to remote address
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return ip
}

func isTrustedIP(clientIP string, trustedIPs []string) bool {
	for _, trustedIP := range trustedIPs {
		if clientIP == trustedIP {
			return true
		}

		// Check if it's a CIDR range
		if _, cidr, err := net.ParseCIDR(trustedIP); err == nil {
			if ip := net.ParseIP(clientIP); ip != nil && cidr.Contains(ip) {
				return true
			}
		}
	}
	return false
}
