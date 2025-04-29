package errors

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	middleware "github.com/oapi-codegen/nethttp-middleware"
)

type ErrorResponse struct {
	Timestamp string   `json:"timestamp"`
	Path      string   `json:"path"`
	Errors    []string `json:"errors"`
}

func NiceErrorHandler(err error, r *http.Request) (int, []byte) {
	resp := ErrorResponse{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      r.URL.Path,
		Errors:    make([]string, 0),
	}

	if me, ok := err.(openapi3.MultiError); ok {
		for _, e := range me {
			msg := e.Error()
			if idx := strings.IndexRune(msg, '\n'); idx != -1 {
				msg = msg[:idx]
			}
			resp.Errors = append(resp.Errors, msg)
		}
	} else {
		resp.Errors = append(resp.Errors, err.Error())
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
	status, body := NiceErrorHandler(err, r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}

func ChiErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	status, body := NiceErrorHandler(err, r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}
