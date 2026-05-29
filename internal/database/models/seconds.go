package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Seconds time.Duration

func (d Seconds) Duration() time.Duration {
	return time.Duration(d)
}

func (d Seconds) Value() (driver.Value, error) {
	return int64(time.Duration(d).Seconds()), nil
}

func (d *Seconds) Scan(value any) error {
	switch v := value.(type) {
	case int64, int32, int16, int8:
		*d = Seconds(time.Duration(v.(int64)) * time.Second)
		return nil

	case float64, float32:
		*d = Seconds(time.Duration(v.(float64) * float64(time.Second)))
		return nil

	default:
		return fmt.Errorf("cannot scan %T into Seconds", value)
	}
}
