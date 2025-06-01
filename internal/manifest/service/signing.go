package service

import (
	"encoding/json"
	"github.com/gibson042/canonicaljson-go"
	"github.com/google/uuid"
	"pluto-backend/internal/manifest/api/gen"
)

// buildCanonicalPayload формирует компактный отсортированный JSON
func buildCanonicalPayload(
	id uuid.UUID,
	version string,
	req gen.ManifestCreate,
	scriptCode string,
) ([]byte, error) {

	uiRaw := json.RawMessage(req.Ui)

	var wrap struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(req.Script, &wrap)
	canon := map[string]any{
		"actions": req.Actions,
		"meta": map[string]any{
			"id":      id.String(),
			"version": version,
			"author": map[string]string{
				"name":  req.Author.Name,
				"email": req.Author.Email,
			},
			"category": req.Category,
			"icon":     req.Icon,
			"tags":     req.Tags, // порядок не меняем
		},
		"permissions": req.Permissions, // порядок не меняем
		"script":      map[string]string{"code": scriptCode},
		"ui":          uiRaw, // raw-embed
	}

	return canonicaljson.Marshal(canon)
}
