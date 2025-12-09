package database

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var protocols = map[string]string{
	"sqlite":     "sqlite",
	"psql":       "postgres",
	"postgres":   "postgres",
	"postgresql": "postgres",
}

func CreateDialector(value string) (*gorm.Dialector, string, error) {
	separator := strings.Index(value, "://")
	if separator == -1 {
		return nil, "", fmt.Errorf("no protocol specified, prefix your connection with sqlite:// or psql://")
	}

	protocol := value[:separator]
	if protocol, ok := protocols[protocol]; !ok {
		return nil, "", fmt.Errorf("unknown protocol %s", protocol)
	}

	dsn := value[separator+3:]
	var dialector gorm.Dialector

	switch protocol {
	case "sqlite":
		dialector = sqlite.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	}

	return &dialector, protocol, nil
}
