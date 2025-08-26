package smartgzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func decompressGzip(data []byte) (string, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func TestGzipOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		paths          []string
		requestPath    string
		expectedEncode bool
		expectedBody   string
	}{
		{
			name:           "匹配路径时启用gzip",
			paths:          []string{"/gzip"},
			requestPath:    "/gzip",
			expectedEncode: true,
			expectedBody:   "hello gzip",
		},
		{
			name:           "不匹配路径时不启用gzip",
			paths:          []string{"/gzip"},
			requestPath:    "/plain",
			expectedEncode: false,
			expectedBody:   "hello plain",
		},
		{
			name:           "路径未加斜杠也能匹配",
			paths:          []string{"gzip"},
			requestPath:    "/gzip",
			expectedEncode: true,
			expectedBody:   "hello gzip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(GzipOnly(tt.paths...))

			r.GET("/gzip", func(c *gin.Context) {
				c.String(http.StatusOK, "hello gzip")
			})
			r.GET("/plain", func(c *gin.Context) {
				c.String(http.StatusOK, "hello plain")
			})

			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			req.Header.Set("Accept-Encoding", "gzip")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			res := w.Result()
			body, _ := io.ReadAll(res.Body)

			if tt.expectedEncode {
				if res.Header.Get("Content-Encoding") != "gzip" {
					t.Errorf("expected Content-Encoding=gzip, got %q", res.Header.Get("Content-Encoding"))
				}
				decoded, err := decompressGzip(body)
				if err != nil {
					t.Errorf("failed to decompress gzip: %v", err)
				}
				if decoded != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, decoded)
				}
			} else {
				if res.Header.Get("Content-Encoding") != "" {
					t.Errorf("expected no Content-Encoding, got %q", res.Header.Get("Content-Encoding"))
				}
				if string(body) != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, string(body))
				}
			}
		})
	}
}
