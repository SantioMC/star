package plugin

import (
	"log/slog"
	"reflect"

	"github.com/tubblelive/star/server/crypto"
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
	
	if err := db.Callback().Create().Before("gorm:update").Register("crypto:update_encrypt", encryptUpdate); err != nil {
		return err
	}
	
	if err := db.Callback().Create().After("gorm:update").Register("crypto:update_decrypt", decryptModel); err != nil {
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

func getFields(schem *schema.Schema) []*schema.Field {
	if schem == nil {
		return nil
	}

	var list []*schema.Field

	fields := schem.FieldsByName
	for k := range fields {
		field := fields[k]
		if field.Tag.Get("encrypt") != "true" {
			continue
		}

		list = append(list, field)
	}

	return list
}

func getReflectElem(iface interface{}) (reflect.Value, bool) {
	destValue := reflect.ValueOf(iface)

	if destValue.Kind() == reflect.Struct {
		ptr := reflect.New(destValue.Type())
		ptr.Elem().Set(destValue)
		destValue = ptr.Elem()
	}

	slog.Info("destType kind", "kind", destValue.Kind())
	for destValue.Kind() == reflect.Ptr {
		destValue = destValue.Elem()
		slog.Info("new destType kind", "kind", destValue.Kind())
	}

	if destValue.Kind() == reflect.Struct {
		return destValue, true
	}

	return reflect.ValueOf(nil), false
}

func modifyValues(values []interface{}, fields []*schema.Field, callback func(string) string) {
	for i, v := range values {
		if v, ok := v.(string); ok {
			values[i] = callback(v)
		}
	}
}

func modifyFields(reflectValue reflect.Value, fields []*schema.Field, db *gorm.DB, callback func(string) string) {
	ctx := db.Statement.Context
	for _, field := range fields {
		if actualValue, isZero := field.ValueOf(ctx, reflectValue); !isZero {
			value, ok := actualValue.(string); 
			if !ok {
				slog.Warn("Failed to modify field, it was not a string", "field", field.Name)
				continue
			}

			newValue := callback(value)
			if err := field.Set(ctx, reflectValue, newValue); err != nil {
				panic(err)
			}
		} else {
			continue
		}
	}
}

func encryptModel(db *gorm.DB) {
	slog.Info("encrypting model")
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
	slog.Info("decrypting model", "table", db.Statement.Table)

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

			if curVal.Kind() == reflect.Pointer {
				curVal = curVal.Elem()
			}

			if curVal.Kind() == reflect.Struct {
				modifyFields(ref, fields, db, func(encrypted string) string {
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

		slog.Info("fieldValue", "kind", fieldValue.Kind(), "field", field.Name, "value", fieldValue)
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
			if cf := findCryptFieldWithClauseColumn(expr.Column, fields); cf != nil {
				exprs[i] = clause.Eq{
					Column: expr.Column,
					Value:  cf.Strategy.Encrypt(expr.Value.(string), cf.Field, db),
				}
			}
		case clause.IN:
			if cf := findCryptFieldWithClauseColumn(expr.Column, fields); cf != nil {
				exprs[i] = clause.IN{
					Column: expr.Column,
					Values: modifyValues(expr.Values, cf, func(value) string {
						return ""
					}),
				}
			}
		case clause.Expr:
			if expr.Vars != nil {
				for i, v := range expr.Vars {
					if v, ok := v.(*CryptoValue); ok {
						if cf := findCryptFieldWithClauseColumn(v.Column, fields); cf != nil {
							expr.Vars[i] = cf.Strategy.Encrypt(v.Value, cf.Field, db)
						}
					}
				}
			}
		}
	}
}