package service

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"pluto-backend/internal/manifest/api/gen"
	"sort"
)

// canonicalInterface из предыдущего шага
func canonicalInterface(raw json.RawMessage) (interface{}, error) {
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// buildCanonicalPayload формирует компактный отсортированный JSON
func buildCanonicalPayload(
	id uuid.UUID,
	version string,
	req gen.ManifestCreate,
	scriptCode string,
) ([]byte, error) {
	// сортируем простые массивы
	sort.Strings(req.Tags)
	sort.Strings(req.Permissions)

	// actions
	var actionsIface []interface{}
	if req.Actions != nil {
		for _, raw := range *req.Actions {
			var m map[string]interface{}
			if err := json.Unmarshal(raw, &m); err != nil {
				return nil, fmt.Errorf("unmarshal action: %w", err)
			}
			actionsIface = append(actionsIface, m)
		}
		sort.Slice(actionsIface, func(i, j int) bool {
			ai, _ := actionsIface[i].(map[string]interface{})["id"].(string)
			aj, _ := actionsIface[j].(map[string]interface{})["id"].(string)
			return ai < aj
		})
	}

	// формируем итоговый map с meta-объектом
	canon := map[string]interface{}{
		"meta": map[string]interface{}{
			"id":      id.String(),
			"version": version,
			"author": map[string]string{
				"email": req.Author.Email,
				"name":  req.Author.Name,
			},
			"category": req.Category,
			"icon":     req.Icon,
			"tags":     req.Tags,
		},
		"permissions": req.Permissions,
		"script":      map[string]string{"code": scriptCode},
		"ui": map[string]interface{}{
			"components": req.Ui,
		},
	}

	if len(actionsIface) > 0 {
		canon["actions"] = actionsIface
	}

	return json.Marshal(canon)
}

// 2) Быстро получить []byte «канонического» JSON
func canonicalBytes(raw json.RawMessage) ([]byte, error) {
	v, err := canonicalInterface(raw)
	if err != nil {
		return nil, err
	}
	// Compact: без отступов, ровно то, что нам нужно для подписи
	return json.Marshal(v)
}

func canonicalizeUI(raw json.RawMessage) (interface{}, error) {
	var ui map[string]interface{}
	if err := json.Unmarshal(raw, &ui); err != nil {
		return nil, err
	}

	// сортировка компонентов по id
	if components, ok := ui["components"].([]interface{}); ok {
		sort.Slice(components, func(i, j int) bool {
			ci, _ := components[i].(map[string]interface{})["id"].(string)
			cj, _ := components[j].(map[string]interface{})["id"].(string)
			return ci < cj
		})

		// actions внутри каждого компонента тоже сортируем по onTap
		for _, comp := range components {
			compMap, _ := comp.(map[string]interface{})
			if actions, exists := compMap["actions"].([]interface{}); exists {
				sort.Slice(actions, func(i, j int) bool {
					ai, _ := actions[i].(map[string]interface{})["onTap"].(string)
					aj, _ := actions[j].(map[string]interface{})["onTap"].(string)
					return ai < aj
				})
			}
		}
		ui["components"] = components
	}

	// рекурсивная сортировка layout
	var sortLayout func(layout map[string]interface{})
	sortLayout = func(layout map[string]interface{}) {
		if children, ok := layout["children"].([]interface{}); ok {
			for _, child := range children {
				if childMap, ok := child.(map[string]interface{}); ok {
					sortLayout(childMap)
				}
			}
		}
	}

	if layout, ok := ui["layout"].(map[string]interface{}); ok {
		sortLayout(layout)
	}

	return ui, nil
}
