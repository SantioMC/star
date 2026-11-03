package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSON map[string]any

func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}

	encoded, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}

	return string(encoded), nil
}

func (j *JSON) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*j = nil
		return nil

	case []byte:
		return json.Unmarshal(v, j)

	case string:
		return json.Unmarshal([]byte(v), j)

	default:
		return fmt.Errorf("cannot scan %T into JSON", value)
	}
}

type JSONData[T any] struct {
	Data T
}

func (j JSONData[T]) Value() (driver.Value, error) {
	encoded, err := json.Marshal(j.Data)
	if err != nil {
		return nil, err
	}

	return string(encoded), nil
}

func (j *JSONData[T]) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, &j.Data)

	case string:
		return json.Unmarshal([]byte(v), &j.Data)

	default:
		return fmt.Errorf("cannot scan %T into JSONData", value)
	}
}
