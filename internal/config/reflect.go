package config

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func getEnv(key string, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	return value
}

func loadConfig(config any) error {
	configValue := reflect.ValueOf(config).Elem()
	typeOfConfig := reflect.TypeOf(config).Elem()

	for i := 0; i < configValue.NumField(); i++ {
		field := configValue.Field(i)
		fieldType := typeOfConfig.Field(i)

		if fieldType.Type.Kind() == reflect.Struct {
			instance := reflect.New(fieldType.Type)

			if err := loadConfig(instance.Interface()); err != nil {
				return err
			}

			field.Set(instance.Elem())
			continue
		}

		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			continue
		}

		data := strings.Split(envTag, ",")
		defaultValue := ""

		if len(data) > 1 {
			envTag = data[0]
			defaultValue = data[1]
		}

		value := getEnv(envTag, defaultValue)
		isEmptySlice := field.Kind() == reflect.Slice && len(data) > 1 && value == ""

		if value == "" && fieldType.Type.Kind() == reflect.Ptr {
			field.Set(reflect.Zero(fieldType.Type))
			continue
		} else if isEmptySlice {
			field.Set(reflect.Zero(fieldType.Type))
			continue
		} else if value == "" {
			return errors.New("missing required environment variable: " + envTag)
		}

		elem := field
		if fieldType.Type.Kind() == reflect.Ptr {
			elem = reflect.New(fieldType.Type.Elem()).Elem()
		}

		parsed, err := getParsed(elem, value)
		if err != nil {
			return err
		}

		if fieldType.Type.Kind() == reflect.Ptr {
			ptr := reflect.New(fieldType.Type.Elem())
			ptr.Elem().Set(reflect.ValueOf(parsed))
			field.Set(ptr)
		} else {
			field.Set(reflect.ValueOf(parsed))
		}
	}

	return nil
}

func getParsed(field reflect.Value, value string) (interface{}, error) {
	switch field.Kind() {

	case reflect.String:
		return value, nil

	case reflect.Int:
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}

		return intValue, nil

	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}

		return boolValue, nil

	case reflect.Slice:
		values := strings.Split(value, ",")
		slice := reflect.MakeSlice(field.Type(), len(values), len(values))

		for i, item := range values {
			slice.Index(i).Set(reflect.ValueOf(item))
		}

		return slice.Interface(), nil

	default:
		return nil, errors.New("unable to handle field: " + field.Kind().String())
	}
}
