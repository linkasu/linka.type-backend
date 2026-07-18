package httpmiddleware

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"

	"github.com/linkasu/linka.type-backend/internal/httpapi"
	"github.com/linkasu/linka.type-backend/internal/requestid"
)

// Recovery converts panics before response commit into a sanitized error response.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writer := &recoveryWriter{ResponseWriter: w}
			defer func() {
				if recover() == nil {
					return
				}
				rid := requestid.FromContext(r.Context())
				if rid == "" {
					rid = requestid.Normalize(w.Header().Get(requestid.Header))
				}
				logger.Error("request panic recovered", "request_id", rid, "error_code", "internal_error")
				if !writer.committed {
					httpapi.WriteError(writer, http.StatusInternalServerError, "internal_error")
				}
			}()
			next.ServeHTTP(writer, r)
		})
	}
}

type recoveryWriter struct {
	http.ResponseWriter
	committed bool
}

func (w *recoveryWriter) WriteHeader(status int) {
	w.committed = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *recoveryWriter) Write(data []byte) (int, error) {
	w.committed = true
	return w.ResponseWriter.Write(data)
}

func (w *recoveryWriter) Flush() {
	w.committed = true
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *recoveryWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("response writer does not support hijacking")
	}
	w.committed = true
	return hijacker.Hijack()
}

func (w *recoveryWriter) Push(target string, options *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, options)
	}
	return http.ErrNotSupported
}

func (w *recoveryWriter) ReadFrom(reader io.Reader) (int64, error) {
	w.committed = true
	if readerFrom, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		return readerFrom.ReadFrom(reader)
	}
	return io.Copy(w.ResponseWriter, reader)
}
