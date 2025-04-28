package openapi

import (
	"context"
	"encoding/json"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"net/http"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

type MultiErrorResponse struct {
	Timestamp string   `json:"timestamp"`
	Path      string   `json:"path"`
	Errors    []string `json:"errors"`
}

func NiceMultiErrorHandler(me openapi3.MultiError, r *http.Request) (int, []byte) {
	resp := MultiErrorResponse{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      r.URL.Path,
		Errors:    make([]string, 0, len(me)),
	}
	for _, e := range me {
		fullMsg := e.Error()
		firstLine := fullMsg
		if idx := strings.IndexRune(fullMsg, '\n'); idx != -1 {
			firstLine = fullMsg[:idx]
		}
		resp.Errors = append(resp.Errors, firstLine)
	}

	data, _ := json.Marshal(resp)
	return http.StatusBadRequest, data
}

func ErrorHandlerWithMultiError(
	ctx context.Context,
	err error,
	w http.ResponseWriter,
	r *http.Request,
	opts middleware.ErrorHandlerOpts,
) {
	if me, ok := err.(openapi3.MultiError); ok {
		status, body := NiceMultiErrorHandler(me, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(body)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"path":      r.URL.Path,
		"error":     err.Error(),
	})
}
