package datatypes

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type StringList []string

func (l *StringList) Scan(value interface{}) error {
	var bytes string
	switch v := value.(type) {
	case []byte:
		bytes = string(v)
	case string:
		bytes = v
	default:
		return fmt.Errorf("value is not a string or bytes, %v", value)
	}

	if len(bytes) != 0 {
		*l = strings.Split(bytes, ",")
	} else {
		*l = make(StringList, 0)
	}

	return nil
}

func (l StringList) Value() (driver.Value, error) {
	return strings.Join(l, ","), nil
}

// MarshalJSON to output non base64 encoded []byte
func (l StringList) MarshalJSON() ([]byte, error) {
	if l == nil {
		return []byte("null"), nil
	}
	t := ([]string)(l)
	return json.Marshal(t)
}

// UnmarshalJSON to deserialize []byte
func (l *StringList) UnmarshalJSON(b []byte) error {
	var t []string
	err := json.Unmarshal(b, &t)
	*l = StringList(t)
	return err
}

// GormDataType gorm common data type
func (l StringList) GormDataType() string {
	return "stringlist"
}

// GormDBDataType gorm db data type
func (StringList) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	case "sqlserver":
		return "NVARCHAR(MAX)"
	}
	return ""
}

func (l StringList) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	data, _ := l.MarshalJSON()
	switch db.Dialector.Name() {
	case "mysql":
		if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
			return gorm.Expr("CAST(? AS JSON)", string(data))
		}
	}
	return gorm.Expr("?", string(data))
}
