package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		headers  map[string]string
		remoteAddr string
		expected string
	}{
		{
			name: "Forwarded header (IPv4)",
			headers: map[string]string{
				"Forwarded": "for=192.0.2.60;proto=http;by=203.0.113.43",
			},
			expected: "192.0.2.60",
		},
		{
			name: "Forwarded header (IPv6 with port)",
			headers: map[string]string{
				"Forwarded": `for="[2001:db8:cafe::17]:4711"`,
			},
			expected: "2001:db8:cafe::17",
		},
		{
			name: "Forwarded header (multiple)",
			headers: map[string]string{
				"Forwarded": "for=192.0.2.60, for=198.51.100.17",
			},
			expected: "192.0.2.60",
		},
		{
			name: "X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195, 70.41.3.18, 150.172.238.178",
			},
			expected: "203.0.113.195",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.195",
			},
			expected: "203.0.113.195",
		},
		{
			name: "Fallback to RemoteAddr",
			remoteAddr: "1.2.3.4:1234",
			expected: "1.2.3.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, r := gin.CreateTestContext(httptest.NewRecorder())
			r.SetTrustedProxies(nil) // Trust all for tests to work with headers
			c.Request, _ = http.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				c.Request.Header.Set(k, v)
			}
			if tt.remoteAddr != "" {
				c.Request.RemoteAddr = tt.remoteAddr
			}

			assert.Equal(t, tt.expected, GetClientIP(c))
		})
	}
}
