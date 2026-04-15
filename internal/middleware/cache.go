package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

type cacheResponseWriter struct {
	http.ResponseWriter
	buf        bytes.Buffer
	statusCode int
}

func (w *cacheResponseWriter) Write(data []byte) (int, error) {
	return w.buf.Write(data)
}

func (w *cacheResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

// Cache adds ETag and Cache-Control headers to responses.
// It generates an ETag from the response body hash and returns 304 if unchanged.
func Cache(maxAge time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only cache GET requests
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// Capture the response
			crw := &cacheResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(crw, r)

			// Generate ETag from response body
			hash := sha256.Sum256(crw.buf.Bytes())
			etag := `"` + hex.EncodeToString(hash[:16]) + `"`

			// Check If-None-Match header
			if r.Header.Get("If-None-Match") == etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			// Copy headers from captured response
			for k, v := range crw.Header() {
				w.Header()[k] = v
			}

			// Set cache headers
			w.Header().Set("ETag", etag)
			if maxAge > 0 {
				w.Header().Set("Cache-Control", "public, max-age="+itoa(int(maxAge.Seconds())))
			}

			// Write the response
			w.WriteHeader(crw.statusCode)
			_, _ = w.Write(crw.buf.Bytes())
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
