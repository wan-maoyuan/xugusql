/*PACK NAME*/
package drive

import (
	"database/sql"
	"reflect"
	"time"
)

type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

type xugusqlField struct {
	tableName string

	/*
	 * Store the name of the column
	 * name of the current field
	 * */
	name   string
	length int

	/*
	 * Store the data type information
	 * of the current field column
	 * */
	fieldType fieldType
}

type fieldType byte

const (
	fieldTypeBool fieldType = iota + 0x01
	fieldTypeChar
	fieldTypeTinyint
	fieldTypeShort
	fieldTypeInteger
	fieldTypeBigint
	fieldTypeFloat
	fieldTypeDouble
	fieldTypeNumeric
	fieldTypeDate
	fieldTypeTime
	fieldTypeTimeTZ
	fieldTypeDatetime   fieldType = 23
	fieldTypeDatetimeTZ fieldType = 14
	fieldTypeBinary     fieldType = 15

	fieldTypeInterval    fieldType = 21
	fieldTypeIntervalY2M fieldType = 28
	fieldTypeIntervalD2S fieldType = 31
	fieldTypeLob         fieldType = 40
	fieldTypeClob        fieldType = 41
	fieldTypeBlob        fieldType = 42
)

/* {{ */
func (self *xugusqlField) typeDatabaseName() string {
	switch self.fieldType {
	case fieldTypeBool:
		return "BOOLEAN"
	case fieldTypeChar:
		return "CHAR"
	case fieldTypeTinyint:
		return "TINYINT"
	case fieldTypeShort:
		return "SHORT"
	case fieldTypeInteger:
		return "INTEGER"
	case fieldTypeBigint:
		return "BIGINT"
	case fieldTypeFloat:
		return "FLOAT"
	case fieldTypeDouble:
		return "DOUBLE"
	case fieldTypeNumeric:
		return "NUMERIC"
	case fieldTypeDate:
		return "DATE"
	case fieldTypeTime:
		return "TIME"
	case fieldTypeTimeTZ:
		return "TIMEZONE"
	case fieldTypeDatetime:
		return "DATETIME"
	case fieldTypeDatetimeTZ:
		return "DATETIME TIMEZONE"
	case fieldTypeBinary:
		return "BINARY"
	case fieldTypeInterval:
		return "INTERVAL"
	case fieldTypeIntervalY2M:
		return "INTERVAL YEAR TO MONTH"
	case fieldTypeIntervalD2S:
		return "INTERVAL DAY TO SECOND"
	case fieldTypeClob:
		return "CLOB"
	case fieldTypeBlob:
		return "BLOB"
	default:
		return ""
	}
}

/* {{ */
func (self *xugusqlField) scanType() reflect.Type {
	switch self.fieldType {
	case fieldTypeBool:
		return scanTypeBool
	case fieldTypeTinyint:
		return scanTypeInt8
	case fieldTypeShort:
		return scanTypeInt16
	case fieldTypeInteger:
		return scanTypeInt32
	case fieldTypeBigint:
		return scanTypeInt64
	case fieldTypeFloat:
		return scanTypeFloat32
	case fieldTypeDouble:
		return scanTypeFloat64
	case fieldTypeDate,
		fieldTypeTime,
		fieldTypeDatetime:
		return scanTypeNullTime
	case fieldTypeTimeTZ,
		fieldTypeDatetimeTZ,
		fieldTypeChar,
		fieldTypeBinary,
		fieldTypeInterval,
		fieldTypeNumeric,
		fieldTypeIntervalY2M,
		fieldTypeIntervalD2S,
		fieldTypeLob,
		fieldTypeClob,
		fieldTypeBlob:
		return scanTypeRawBytes
	default:
		return scanTypeUnknown

	}
}

var (
	scanTypeFloat32   = reflect.TypeOf(float32(0))
	scanTypeFloat64   = reflect.TypeOf(float64(0))
	scanTypeNullFloat = reflect.TypeOf(sql.NullFloat64{})
	scanTypeNullInt   = reflect.TypeOf(sql.NullInt64{})
	scanTypeNullTime  = reflect.TypeOf(time.Time{})
	scanTypeInt8      = reflect.TypeOf(int8(0))
	scanTypeInt16     = reflect.TypeOf(int16(0))
	scanTypeInt32     = reflect.TypeOf(int32(0))
	scanTypeInt64     = reflect.TypeOf(int64(0))
	scanTypeUnknown   = reflect.TypeOf(new(interface{}))
	scanTypeRawBytes  = reflect.TypeOf(sql.RawBytes{})
	scanTypeUint8     = reflect.TypeOf(uint8(0))
	scanTypeUint16    = reflect.TypeOf(uint16(0))
	scanTypeUint32    = reflect.TypeOf(uint32(0))
	scanTypeUint64    = reflect.TypeOf(uint64(0))
	scanTypeBool      = reflect.TypeOf(bool(false))
)
