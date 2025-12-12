package plugin

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/tubblelive/star/server/crypto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

var encryptionKey [32]byte

func EncryptionPlugin(db *gorm.DB, key [32]byte) error {
	encryptionKey = key

	_ = db.Callback().Create().Before("gorm:create").Register("star:before_create", encryptParamBeforeCreate)
	_ = db.Callback().Create().After("gorm:create").Register("star:after_create", decryptModel)

	_ = db.Callback().Update().Before("gorm:update").Register("star:before_update", encryptParamBeforeUpdate)
	
	_ = db.Callback().Query().Before("gorm:query").Register("star:before_query", encryptParamBeforeQuery)
	_ = db.Callback().Query().After("gorm:query").Register("star:after_query", decryptModel)
	
	return nil
}

type CryptoValue struct {
	Column string
	Value  string
}

func NewCryptoValue(column string, val string) *CryptoValue {
	return &CryptoValue{
		Column: column,
		Value:  val,
	}
}

func encryptParamBeforeCreate(db *gorm.DB) {
	fields := getSchemaCryptoFields(db.Statement.Schema)
	names := make([]string, len(fields))
	
	for i, f := range fields {
		names[i] = f.Name
	}

	if len(fields) == 0 {
		return
	}

	switch db.Statement.ReflectValue.Kind() {
	case reflect.Slice:
		for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
			sliceV := db.Statement.ReflectValue.Index(i)
			encryptFields(sliceV, fields, db)
		}
	case reflect.Struct:
		encryptFields(db.Statement.ReflectValue, fields, db)
	default:
		return
	}
}

func encryptParamBeforeQuery(db *gorm.DB) {
	fields := getSchemaCryptoFields(db.Statement.Schema)
	if len(fields) == 0 {
		return
	}

	if we := db.Statement.Clauses["WHERE"].Expression; we != nil {
		exprs := we.(clause.Where).Exprs
		for i, expr := range exprs {
			switch expr := expr.(type) {
			case clause.Eq:
				if cf := findCryptFieldWithClauseColumn(expr.Column, fields); cf != nil {
					encrypted, err := crypto.EncryptAES(expr.Value.(string), encryptionKey)
					if err != nil {
						panic(err)
					}

					exprs[i] = clause.Eq{
						Column: expr.Column,
						Value:  encrypted,
					}
				}
			case clause.IN:
				if cf := findCryptFieldWithClauseColumn(expr.Column, fields); cf != nil {
					values, err := encryptValues(expr.Values, cf, db)
					if err != nil {
						panic(err)
					}

					exprs[i] = clause.IN{
						Column: expr.Column,
						Values: values,
					}
				}
			case clause.Expr:
				if expr.Vars != nil {
					for i, v := range expr.Vars {
						if v, ok := v.(*CryptoValue); ok {
							if cf := findCryptFieldWithClauseColumn(v.Column, fields); cf != nil {
								encrypted, err := crypto.EncryptAES(v.Value, encryptionKey)
								if err != nil {
									panic(err)
								}

								expr.Vars[i] = encrypted
							}
						}
					}
				}
			}

		}
	}
}

func encryptParamBeforeUpdate(db *gorm.DB) {
	slog.Info("wefewfwe schema", "table", db.Statement.Table)

	// dbSchema := db.Statement.Schema
	// if dbSchema == nil {
	// 	return
	// }

	// // TODO: make all the lookups to check if the tag is true
	// if updateInfo, ok := db.Statement.Dest.(map[string]interface{}); ok {
	// 	for updateColumn := range updateInfo {
	// 		updateV := updateInfo[updateColumn]
	// 		updateField := dbSchema.LookUpField(updateColumn)
	// 		if _, ok := updateField.Tag.Lookup("encrypt"); ok {
	// 			stringVal, ok := updateV.(string)
	// 			if !ok {
	// 				continue
	// 			}

	// 			encrypted, err := crypto.EncryptAES(stringVal, encryptionKey)
	// 			if err != nil {
	// 				panic(err)
	// 			}

	// 			encryptionValue := encrypted
	// 			updateInfo[updateColumn] = encryptionValue
	// 		}
	// 	}

	// 	return
	// }

	// destType, destValue := getReflectElem(db.Statement.Dest)
	// if destType != nil {
	// 	for i := 0; i < destType.NumField(); i++ {
	// 		field := destType.Field(i)
	// 		fieldValue := destValue.Field(i)
	// 		if _, ok := field.Tag.Lookup("encrypt"); ok {
	// 			val := fieldValue.String()
	// 			if len(val) == 0 {
	// 				continue
	// 			}

	// 			encrypted, err := crypto.EncryptAES(val, encryptionKey)
	// 			if err != nil {
	// 				panic(err)
	// 			}

	// 			encryptionValue := encrypted
	// 			fieldValue.SetString(encryptionValue)
	// 		}
	// 	}
	// }
}

func getReflectElem(i interface{}) (reflect.Type, reflect.Value) {
	destType := reflect.TypeOf(i)
	destValue := reflect.ValueOf(i)
	
	if destType.Kind() == reflect.Struct {
		ptr := reflect.New(destType)
		ptr.Elem().Set(destValue)
		destType = ptr.Elem().Type()
		destValue = ptr.Elem()
	}
	
	for destType.Kind() == reflect.Ptr {
		destType = destType.Elem()
		destValue = destValue.Elem()
	}
	
	if destType.Kind() == reflect.Struct {
		return destType, destValue
	}
	
	return nil, reflect.ValueOf(nil)
}

func decryptModel(db *gorm.DB) {
	fields := getSchemaCryptoFields(db.Statement.Schema)
	if len(fields) == 0 {
		return
	}

	refVal := reflect.ValueOf(db.Statement.Dest).Elem()
	
	switch refVal.Kind() {
	case reflect.Struct:
		err := decryptFields(refVal, fields, db)
		if err != nil {
			slog.Error("failed to decrypt model", "table", db.Statement.Table, "error", err)
		}
	case reflect.Slice:
		if refVal.Len() >= 1 {
			for i := 0; i < refVal.Len(); i++ {
				curVal := refVal.Index(i)
				
				if curVal.Kind() == reflect.Pointer {
					curVal = curVal.Elem()
				}

				if curVal.Kind() == reflect.Struct {
					err := decryptFields(curVal, fields, db)
					if err != nil {
						slog.Error("failed to decrypt model", "table", db.Statement.Table, "error", err)
					}
				}
			}
		}
	}
}

func getSchemaCryptoFields(schem *schema.Schema) []*schema.Field {
	if schem == nil {
		return nil
	}

	var fields []*schema.Field = []*schema.Field{}
	fieldNames := schem.FieldsByName

	for k := range fieldNames {
		v := fieldNames[k]
		if _, ok := v.Tag.Lookup("encrypt"); ok {
			if v.FieldType.Kind() == reflect.String {
				fields = append(fields, v)
				continue
			}

			if v.FieldType.Kind() == reflect.Pointer && v.FieldType.Elem().Kind() == reflect.String {
				fields = append(fields, v)
				continue
			}

			slog.Warn("field is not a string, it will not be encrypted", "name", v.Name)
		}
	}

	return fields
}

func encryptValues(values []interface{}, cf *schema.Field, db *gorm.DB) ([]interface{}, error) {
	for i, v := range values {
		if v, ok := v.(string); ok {
			encrypted, err := crypto.EncryptAES(v, encryptionKey)
			if err != nil {
				return []interface{}{}, err
			}

			values[i] = encrypted
		}
	}

	return values, nil
}

func findCryptFieldWithClauseColumn(column interface{}, fields []*schema.Field) *schema.Field {
	if c, ok := (column).(string); ok {
		return findCryptField(c, fields)
	}
	
	if c, ok := (column).(clause.Column); ok {
		return findCryptField(c.Name, fields)
	}

	return nil
}

func findCryptField(fieldName string, fields []*schema.Field) *schema.Field {
	if fieldName == "" {
		return nil
	}

	for _, f := range fields {
		if f.DBName == fieldName {
			return f
		}
	}
	
	return nil
}

func decryptFields(src reflect.Value, fields []*schema.Field, db *gorm.DB) error {
	for _, field := range fields {
		refValue := src.FieldByName(field.Name)
		
		if refValue.Kind() == reflect.Ptr {
			refValue = refValue.Elem()
		}

		if !refValue.IsValid() {
			continue
		}

		encryptValue := refValue.String()

		if len(encryptValue) > 0 {
			decrypted, err := crypto.DecryptAES(encryptValue, encryptionKey)
			if err != nil {
				return err
			}

			slog.Info("decrypted", "table", db.Statement.Table, "name", field.Name, "before", encryptValue, "after", decrypted)
			field.Set(db.Statement.Context, src, decrypted)
		}
	}

	return nil
}

func getFieldValue(field *schema.Field, ctx context.Context, value reflect.Value) (string, bool) {
	reflected := value.FieldByName(field.Name)
	if !reflected.IsValid() {
		slog.Info("got value", "strat", "nil", "key", field.Name)
		return "", false
	}

	slog.Info("got value", "strat", "struct", "struct", value.Type().Name(), "key", field.Name, "kind", reflected.Kind(), "value", reflected.Interface(), "fullStruct", value.Interface())

	switch reflected.Kind() {
		case reflect.Ptr:
			elem := reflected.Elem()
			if !elem.IsValid() || elem.IsNil() {
				return "", true
			}

			slog.Info("got value", "strat", "ptr", "value", elem.String())
			return elem.String(), true
		case reflect.String:
			slog.Info("got value", "strat", "reg", "key", field.Name, "kind", reflected.Kind(), "value", reflected.String())
			return reflected.String(), true
		default:
			return "", false
	}
}

func encryptFields(reflectValue reflect.Value, fields []*schema.Field, db *gorm.DB) {
	dbCtx := db.Statement.Context
	for _, field := range fields {
		value, ok := getFieldValue(field, dbCtx, reflectValue)
		if !ok {
			slog.Warn("field is not a string, it will not be encrypted", "name", field.Name)
			continue
		}

		encrypted, err := crypto.EncryptAES(value, encryptionKey)
		if err != nil {
			panic(err)
		}

		slog.Info("encrypted", "table", db.Statement.Table, "name", field.Name, "before", value, "after", encrypted)
		if err := field.Set(dbCtx, reflectValue, encrypted); err != nil {
			panic(err)
		}
	}
}