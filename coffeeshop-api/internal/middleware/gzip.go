package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// gzipResponseWriter wraps http.ResponseWriter to compress response bodies.
type gzipResponseWriter struct {
	http.ResponseWriter
	writer      io.Writer
	wroteHeader bool
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		// If Content-Type hasn't been set, let Go detect it from the first write
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", http.DetectContentType(b))
		}
		w.wroteHeader = true
	}
	return w.writer.Write(b)
}

func (w *gzipResponseWriter) Flush() {
	if f, ok := w.writer.(*gzip.Writer); ok {
		f.Flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Pool of gzip writers to reduce allocations.
var gzipPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return w
	},
}

// Paths that should never be gzip-compressed (SSE streams, file uploads).
var gzipSkipPrefixes = []string{
	"/api/v1/orders/stream", // SSE stream — needs real-time flushing without compression overhead
}

// Gzip compresses HTTP responses for clients that accept gzip encoding.
// Skips compression for SSE streams and paths listed in gzipSkipPrefixes.
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip if client doesn't accept gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip for SSE and other streaming endpoints
		for _, prefix := range gzipSkipPrefixes {
			if strings.HasPrefix(r.URL.Path, prefix) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Get a gzip writer from the pool
		gz := gzipPool.Get().(*gzip.Writer)
		gz.Reset(w)
		defer func() {
			gz.Close()
			gzipPool.Put(gz)
		}()

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		// Remove Content-Length since it will change after compression
		w.Header().Del("Content-Length")

		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			writer:         gz,
		}

		next.ServeHTTP(gzw, r)
	})
}
