package utils

import (
	"encoding/json"

	"github.com/sqlc-dev/pqtype"
)

func UnmarshalNullableJSON[T any](nr pqtype.NullRawMessage) (T, error) {
	var zero T
	if !nr.Valid {
		return zero, nil
	}
	var out T
	if err := json.Unmarshal(nr.RawMessage, &out); err != nil {
		return zero, err
	}
	return out, nil
}
