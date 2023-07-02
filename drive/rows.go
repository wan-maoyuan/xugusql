package drive

import (
	"C"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"time"
	"unsafe"
)

type Row struct {
	// Data information storage carrier of each column in the result set
	columns []xugusqlField
	// The name of each column in the result set of the current query
	names []string
	done  bool
}

type xugusqlRows struct {
	// A context handle pointer, which can be used to obtain
	// information about the result set
	result unsafe.Pointer
	// The return value of the function cgo_xgc_read_next()
	lastRowRelt int

	// The return value of the function cgo_xgc_next_result()
	lastRelt int
	// Boolean value, used to identify whether the executed
	// SQL statement has been prepared
	prepared bool

	// Context connection handle pointer
	rows_conn unsafe.Pointer
	rowset    Row
}

func (self *xugusqlRows) get_error() error {

	conn := self.rows_conn
	message := cgo_c_calloc(ERROR_BUFF_SIZE)
	defer func() {
		cgo_c_free(unsafe.Pointer(message))
	}()

	var length C.int
	cgo_xgc_error(&conn, message, &length)
	return errors.New(C.GoString(message))
}

/*
 * Columns returns the column names.
 * Columns returns an error if the rows are closed.
 */
func (self *xugusqlRows) Columns() []string {

	var FieldCount C.int

	result := self.result
	if self.rowset.names != nil {
		return self.rowset.names
	}

	re := cgo_xgc_get_fields_count(&result, &FieldCount)
	if re < 0 {
		return self.rowset.names
	}

	column_name := cgo_c_calloc(COLUMN_NAME_BUFF_SIZE)
	defer func() {
		cgo_c_free(unsafe.Pointer(column_name))
	}()

	columns := make([]string, int(FieldCount))
	fields := make([]xugusqlField, int(FieldCount))

	for j := range columns {
		cgo_c_memset(column_name, COLUMN_NAME_BUFF_SIZE)
		re = cgo_xgc_get_column_name(&result, j+1, column_name)
		if re < 0 {
			return columns
		}
		columns[j] = C.GoString(column_name)
		fields[j].name = C.GoString(column_name)

		var dtype C.int
		re = cgo_xgc_get_column_type(&result, j+1, &dtype)
		if re < 0 {
			return columns
		}
		fields[j].fieldType = fieldType(dtype)
	}

	self.rowset.columns = fields
	self.rowset.names = columns

	return columns
}

func (self *xugusqlRows) Close() error {

	result := self.result

	self.rowset.columns = nil
	self.rowset.names = nil

	if result != nil {
		re := cgo_xgc_free_rowset(&result)
		if re < 0 {
			return self.get_error()
		}
		self.result = nil
	}

	return nil
}

// TODO(bradfitz): for now we need to defensively clone all
// []byte that the driver returned (not permitting
// *RawBytes in Rows.Scan), since we're about to close
// the Rows in our defer, when we return from this function.
// the contract with the driver.Next(...) interface is that it
// can return slices into read-only temporary memory that's
// only valid until the next Scan/Close. But the TODO is that
// for a lot of drivers, this copy will be unnecessary. We
// should provide an optional interface for drivers to
// implement to say, "don't worry, the []bytes that I return
// from Next will not be modified again." (for instance, if
// they were obtained from the network anyway) But for now we
// don't care.
func (self *xugusqlRows) Next(dest []driver.Value) error {

	if self.result == nil {
		return errors.New("The result set has been released")
	}

	result := self.result
	self.lastRowRelt = cgo_xgc_read_next(&result)
	if self.lastRowRelt < 0 {
		return self.get_error()
	}

	if self.lastRowRelt == RET_NO_DATA {
		return io.EOF
	}

	pVal := cgo_c_calloc(FIELD_BUFF_SIZE)
	defer func() {
		cgo_c_free(unsafe.Pointer(pVal))
	}()

	var FieldCount = len(self.rowset.names)
	var length C.int

	for j := 0; j < FieldCount; j++ {

		coluType := self.rowset.columns[j].fieldType
		switch coluType {

		case fieldTypeBinary, fieldTypeLob,
			fieldTypeClob, fieldTypeBlob:

			var pLob unsafe.Pointer
			cgo_xgc_new_lob(&pLob)

			re := cgo_xgc_get_lob(&result, j+1, int(coluType), &pLob, LOB_BUFF_SIZE, &length)
			if re < 0 && re != SQL_XG_C_NULL {
				return self.get_error()
			}

			dest[j] = make([]byte, int(length)+1)

			if re == SQL_XG_C_NULL {
				dest[j] = nil
			} else {
				data := make([]byte, int(length))
				cgo_xgc_get_lob_data(&pLob, unsafe.Pointer(&data[0]), length)
				dest[j] = data
			}

			cgo_xgc_lob_distroy(&pLob)

		case fieldTypeDate:
			cgo_c_memset(pVal, FIELD_BUFF_SIZE)
			re := cgo_xgc_get_data(&result, j+1, int(fieldTypeChar), pVal, FIELD_BUFF_SIZE, &length)
			if re < 0 && re != SQL_XG_C_NULL {
				return self.get_error()
			}

			if re == SQL_XG_C_NULL {
				dest[j] = nil
			} else {
				//tzone, _ := time.LoadLocation("Asia/Shanghai")
				//tv, _ := time.ParseInLocation("2006-01-02", C.GoString(pVal), tzone)
				tv, _ := time.Parse("2006-01-02", C.GoString(pVal))
				dest[j] = tv
			}

		case fieldTypeTime,
			fieldTypeTimeTZ:
			cgo_c_memset(pVal, FIELD_BUFF_SIZE)
			re := cgo_xgc_get_data(&result, j+1, int(fieldTypeChar), pVal, FIELD_BUFF_SIZE, &length)
			if re < 0 && re != SQL_XG_C_NULL {
				return self.get_error()
			}

			if re == SQL_XG_C_NULL {
				dest[j] = nil
			} else {
				//tzone, _ := time.LoadLocation("Asia/Shanghai")
				//tv, _ := time.ParseInLocation("15:04:05", C.GoString(pVal), tzone)
				tv, _ := time.Parse("15:04:05", C.GoString(pVal))
				dest[j] = tv
			}

		case fieldTypeDatetime,
			fieldTypeDatetimeTZ:

			cgo_c_memset(pVal, FIELD_BUFF_SIZE)
			re := cgo_xgc_get_data(&result, j+1, int(fieldTypeChar), pVal, FIELD_BUFF_SIZE, &length)
			if re < 0 && re != SQL_XG_C_NULL {
				return self.get_error()
			}

			if re == SQL_XG_C_NULL {
				dest[j] = nil
			} else {
				//tzone, _ := time.LoadLocation("Asia/Shanghai")
				//tv, _ := time.ParseInLocation("2006-01-02 15:04:05", C.GoString(pVal), tzone)
				tv, _ := time.Parse("2006-01-02 15:04:05", C.GoString(pVal))
				dest[j] = tv
			}

		default:
			cgo_c_memset(pVal, FIELD_BUFF_SIZE)
			re := cgo_xgc_get_data(&result, j+1, int(fieldTypeChar), pVal, FIELD_BUFF_SIZE, &length)
			if re < 0 && re != SQL_XG_C_NULL {
				return self.get_error()
			}

			dest[j] = make([]byte, int(length)+1)
			if re == SQL_XG_C_NULL {
				dest[j] = nil
			} else {
				dest[j] = []byte(C.GoString(pVal))
			}
		}
	}

	return nil
}

// The driver is at the end of the current result set.
// Test to see if there is another result set after the current one.
// Only close Rows if there is no further result sets to read.
func (self *xugusqlRows) HasNextResultSet() bool {

	result := self.result
	if self.prepared {
		return false
	}

	if self.lastRowRelt == RET_NO_DATA {
		self.lastRelt = cgo_xgc_next_result(&result)
		if self.lastRelt == RET_NO_DATA {
			return false
		}

		self.rowset.columns = nil
		self.rowset.names = nil
		self.result = result
		return true
	}

	return false
}

// NextResultSet prepares the next result set for reading. It reports whether
// there is further result sets, or false if there is no further result set
// or if there is an error advancing to it. The Err method should be consulted
// to distinguish between the two cases.
//
// After calling NextResultSet, the Next method should always be called before
// scanning. If there are further result sets they may not have rows in the result
// set.
func (self *xugusqlRows) NextResultSet() error {

	if self.result == nil {
		return errors.New("The result set has been released")
	}

	result := self.result
	if self.prepared {
		return io.EOF
	}

	if self.lastRelt == RET_NO_DATA {
		return io.EOF
	}

	self.result = result
	return nil
}

/* {{ */
/* {{ type ColumnTypeScanType interface }} */
func (self *xugusqlRows) ColumnTypeScanType(index int) reflect.Type {
	return self.rowset.columns[index].scanType()
}

/* {{ */
/* {{ RowsColumnTypeDatabaseTypeName }} */
func (self *xugusqlRows) ColumnTypeDatabaseTypeName(index int) string {
	return self.rowset.columns[index].typeDatabaseName()
}

/* {{ */
/* {{ RowsColumnTypeLength }} */
func (self *xugusqlRows) ColumnTypeLength(index int) (int64, bool) {
	return 0, false
}

/* {{ */
/* {{ RowsColumnTypeNullable */
func (self *xugusqlRows) ColumnTypeNullable(index int) (nullable, ok bool) {
	/* not support */
	return false, false
}

/* {{ */
/* {{ RowsColumnTypePrecisionScale */
func (self *xugusqlRows) ColumnTypePrecisionScale(index int) (int64, int64, bool) {
	/* not support */
	return 0, 0, false
}
