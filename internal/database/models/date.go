package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Date struct {
	time.Time
}

func ParseDate(value string) (Date, error) {
	parsed, err := time.Parse(time.DateOnly, value)
	return Date{parsed}, err
}

func (d Date) String() string {
	return d.Format(time.DateOnly)
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d Date) Value() (driver.Value, error) {
	return d.String(), nil
}

func (d *Date) Scan(value any) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v
		return nil

	case []byte:
		return d.parse(string(v))

	case string:
		return d.parse(v)

	default:
		return fmt.Errorf("cannot scan %T into Date", value)
	}
}

func (d *Date) parse(value string) error {
	parsed, err := ParseDate(value)
	if err != nil {
		return err
	}

	*d = parsed
	return nil
}
