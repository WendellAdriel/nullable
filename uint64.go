package nullable

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// Uint64 SQL type that can retrieve NULL value
type Uint64 struct {
	realValue uint64
	isValid   bool
}

// NewUint64 creates a new nullable 64-bit integer
func NewUint64(value *uint64) Uint64 {
	if value == nil {
		return Uint64{
			realValue: 0,
			isValid:   false,
		}
	}
	return Uint64{
		realValue: *value,
		isValid:   true,
	}
}

// Get either nil or 64-bit integer
func (n Uint64) Get() *uint64 {
	if !n.isValid {
		return nil
	}
	return &n.realValue
}

// Set either nil or 64-bit integer
func (n *Uint64) Set(value *uint64) {
	n.isValid = (value != nil)
	if n.isValid {
		n.realValue = *value
	} else {
		n.realValue = 0
	}
}

// MarshalJSON converts current value to JSON
func (n Uint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.Get())
}

// UnmarshalJSON writes JSON to this type
func (n *Uint64) UnmarshalJSON(data []byte) error {
	dataString := string(data)
	if len(dataString) == 0 || dataString == "null" {
		n.isValid = false
		n.realValue = 0
		return nil
	}

	var parsed uint64
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}

	n.isValid = true
	n.realValue = parsed
	return nil
}

// Scan implements scanner interface
func (n *Uint64) Scan(value interface{}) error {
	if value == nil {
		n.realValue, n.isValid = 0, false
		return nil
	}

	var scanned string
	if err := convertAssign(&scanned, value); err != nil {
		return err
	}

	radix := 10
	if len(scanned) == 64 {
		radix = 2
	}

	parsed, err := strconv.ParseUint(scanned, radix, 64)
	if err != nil {
		return err
	}
	n.realValue = parsed

	n.isValid = true
	return nil
}

// Value implements the driver Valuer interface.
func (n Uint64) Value() (driver.Value, error) {
	if !n.isValid {
		return nil, nil
	}
	return strconv.FormatUint(n.realValue, 10), nil
}

// GormValue implements the driver Valuer interface via GORM.
func (n Uint64) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	switch db.Dialector.Name() {
	case "sqlite", "mysql":
		// MySQL and SQLite are using Value() instead of GormValue()
		value, err := n.Value()
		if err != nil {
			db.AddError(err)
			return clause.Expr{}
		}
		return clause.Expr{SQL: "?", Vars: []interface{}{value}}
	case "postgres":
		if !n.isValid {
			return clause.Expr{SQL: "?", Vars: []interface{}{nil}}
		}

		return clause.Expr{SQL: "?", Vars: []interface{}{n.realValue}}
	}
	return clause.Expr{}
}

// GormDataType gorm common data type
func (Uint64) GormDataType() string {
	return "uint64_null"
}

// GormDBDataType gorm db data type
func (Uint64) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite", "mysql":
		return "BIGINT UNSIGNED"
	case "postgres":
		return "numeric"
	}
	return ""
}
