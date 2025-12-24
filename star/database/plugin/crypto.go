package plugin

import (
	"log/slog"
	"reflect"

	"github.com/tubblelive/star/star/crypto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

const (
	encryptionTag = "encrypt"
)

var cryptoPlugin CryptoPlugin

type CryptoPlugin struct {
	encryptionKey [32]byte
}

func CreateEncryptionPlugin(key [32]byte) CryptoPlugin {
	return CryptoPlugin{encryptionKey: key}
}

func (p CryptoPlugin) Name() string {
	return "CryptoPlugin"
}

func (p CryptoPlugin) Initialize(db *gorm.DB) error {
	cryptoPlugin = p

	if err := db.Callback().Create().Before("gorm:create").Register("crypto:create_encrypt", encryptModel); err != nil {
		return err
	}

	if err := db.Callback().Create().After("gorm:create").Register("crypto:create_decrypt", decryptModel); err != nil {
		return err
	}

	if err := db.Callback().Update().Before("gorm:update").Register("crypto:update_encrypt", encryptUpdate); err != nil {
		return err
	}

	if err := db.Callback().Update().After("gorm:update").Register("crypto:update_decrypt", decryptModel); err != nil {
		return err
	}

	if err := db.Callback().Query().Before("gorm:query").Register("crypto:query_encrypt", encryptQuery); err != nil {
		return err
	}

	if err := db.Callback().Query().After("gorm:query").Register("crypto:query_decrypt", decryptModel); err != nil {
		return err
	}

	return nil
}

func filterFilters(fields []*schema.Field) []*schema.Field {
	var list []*schema.Field

	for _, field := range fields {
		if field.Tag.Get("encrypt") != "true" {
			continue
		}

		list = append(list, field)
	}

	return list
}

func getFields(schem *schema.Schema) []*schema.Field {
	if schem == nil {
		return nil
	}

	var fields []*schema.Field

	for _, field := range schem.FieldsByName {
		fields = append(fields, field)
	}

	return filterFilters(fields)
}

func getReflectElem(iface interface{}) (reflect.Value, bool) {
	destValue := reflect.ValueOf(iface)

	if destValue.Kind() == reflect.Struct {
		ptr := reflect.New(destValue.Type())
		ptr.Elem().Set(destValue)
		destValue = ptr.Elem()
	}

	for destValue.Kind() == reflect.Ptr {
		destValue = destValue.Elem()
	}

	if destValue.Kind() == reflect.Struct {
		return destValue, true
	}

	return reflect.ValueOf(nil), false
}

func findClauseField(column interface{}, fields []*schema.Field) *schema.Field {
	name, ok := column.(string)
	if !ok {
		column, ok := column.(clause.Column)
		if !ok {
			return nil
		}

		name = column.Name
	}

	if name == "" {
		return nil
	}

	for _, f := range fields {
		if f.DBName == name {
			return f
		}
	}

	return nil
}

func modifyValues(values []interface{}, callback func(string) string) {
	for i, v := range values {
		if v, ok := v.(string); ok {
			values[i] = callback(v)
		}
	}
}

func modifyFields(reflectValue reflect.Value, fields []*schema.Field, db *gorm.DB, callback func(string) string) {
	ctx := db.Statement.Context
	for _, field := range fields {
		value := field.ReflectValueOf(ctx, reflectValue)

		if value.IsZero() {
			continue
		}

		if value.Kind() == reflect.Ptr {
			value = reflect.Indirect(value)
		}

		if value.Kind() != reflect.String {
			slog.Warn("failed to modify field, it was not a string", "field", field.Name, "kind", value.Kind())
			continue
		}

		newValue := callback(value.String())
		if err := field.Set(ctx, reflectValue, newValue); err != nil {
			panic(err)
		}
	}
}

func encryptModel(db *gorm.DB) {
	schema := db.Statement.Schema
	fields := getFields(schema)
	if fields == nil {
		return
	}

	switch db.Statement.ReflectValue.Kind() {
	case reflect.Slice:
		for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
			sliceV := db.Statement.ReflectValue.Index(i)
			modifyFields(sliceV, fields, db, func(value string) string {
				encrypted, err := crypto.EncryptAES(value, cryptoPlugin.encryptionKey)
				if err != nil {
					slog.Error("failed to encrypt value", "err", err)
					return value
				}

				return encrypted
			})
		}
	case reflect.Struct:
		modifyFields(db.Statement.ReflectValue, fields, db, func(value string) string {
			encrypted, err := crypto.EncryptAES(value, cryptoPlugin.encryptionKey)
			if err != nil {
				slog.Error("failed to encrypt value", "err", err)
				return value
			}

			return encrypted
		})
	default:
		return
	}
}

func decryptModel(db *gorm.DB) {
	schema := db.Statement.Schema
	fields := getFields(schema)
	if fields == nil {
		return
	}

	ref := db.Statement.ReflectValue

	switch ref.Kind() {
	case reflect.Struct:
		modifyFields(ref, fields, db, func(encrypted string) string {
			decrypted, err := crypto.DecryptAES(encrypted, cryptoPlugin.encryptionKey)
			if err != nil {
				slog.Error("failed to encrypt value", "err", err)
				return encrypted
			}

			return decrypted
		})
	case reflect.Slice:
		for i := 0; i < ref.Len(); i++ {
			curVal := ref.Index(i)

			if curVal.Kind() == reflect.Ptr {
				curVal = curVal.Elem()
			}

			if curVal.Kind() == reflect.Struct {
				modifyFields(curVal, fields, db, func(encrypted string) string {
					decrypted, err := crypto.DecryptAES(encrypted, cryptoPlugin.encryptionKey)
					if err != nil {
						slog.Error("failed to encrypt value", "err", err)
						return encrypted
					}

					return decrypted
				})
			}
		}
	}
}

func encryptUpdate(db *gorm.DB) {
	dbSchema := db.Statement.Schema
	if dbSchema == nil {
		return
	}

	if updateInfo, ok := db.Statement.Dest.(map[string]interface{}); ok {
		for updateColumn := range updateInfo {
			updateV := updateInfo[updateColumn]
			updateField := dbSchema.LookUpField(updateColumn)

			stringVal, ok := updateV.(string)
			if !ok {
				continue
			}

			if updateField.Tag.Get(encryptionTag) != "true" {
				continue
			}

			encrypted, err := crypto.EncryptAES(stringVal, cryptoPlugin.encryptionKey)
			if err != nil {
				slog.Error("failed to encrypt field", "field", updateColumn, "err", err)
			}

			updateInfo[updateColumn] = encrypted
		}

		return
	}

	destValue, ok := getReflectElem(db.Statement.Dest)
	if !ok {
		return
	}

	t := destValue.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := destValue.Field(i)

		if fieldValue.IsZero() {
			continue
		}

		if fieldValue.Kind() == reflect.Ptr {
			fieldValue = fieldValue.Elem()
		}

		if field.Tag.Get(encryptionTag) != "true" {
			continue
		}

		val := fieldValue.String()
		if len(val) == 0 {
			continue
		}

		encrypted, err := crypto.EncryptAES(val, cryptoPlugin.encryptionKey)
		if err != nil {
			slog.Error("failed to encrypt value", "value", val, "err", err)
			continue
		}

		fieldValue.SetString(encrypted)
	}
}

func encryptQuery(db *gorm.DB) {
	fields := getFields(db.Statement.Schema)
	if len(fields) == 0 {
		return
	}

	where := db.Statement.Clauses["WHERE"].Expression
	if where == nil {
		return
	}

	exprs := where.(clause.Where).Exprs

	for i, expr := range exprs {
		switch expr := expr.(type) {
		case clause.Eq:
			value, ok := expr.Value.(string)
			if !ok {
				continue
			}

			if field := findClauseField(expr.Column, fields); field != nil {
				encrypted, err := crypto.EncryptAES(value, cryptoPlugin.encryptionKey)
				if err != nil {
					slog.Error("failed to encrypt value", "vaue", value, "err", err)
					continue
				}

				exprs[i] = clause.Eq{
					Column: expr.Column,
					Value:  encrypted,
				}
			}

		case clause.IN:
			if field := findClauseField(expr.Column, fields); field != nil {
				modifyValues(expr.Values, func(value string) string {
					encrypted, err := crypto.EncryptAES(value, cryptoPlugin.encryptionKey)
					if err != nil {
						slog.Error("failed to encrypt value", "vaue", value, "err", err)
						return value
					}

					return encrypted
				})

				exprs[i] = clause.IN{
					Column: expr.Column,
					Values: expr.Values,
				}
			}
		}
	}
}
