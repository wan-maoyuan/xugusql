package drive

import (
	"C"
	"context"
	"database/sql/driver"
	"strings"
	"unsafe"
)

const (
	ERROR_BUFF_SIZE        uint = 1024
	PREPARE_NAME_BUFF_SIZE uint = 128
	CURSOR_NAME_BUFF_SIZE  uint = 128
	ROWID_BUFF_SIZE        uint = 256
	COLUMN_NAME_BUFF_SIZE  uint = 256

	FIELD_BUFF_SIZE uint = 4096
	LOB_BUFF_SIZE   uint = 8
	RET_NO_DATA     int  = 100

	SQL_UNKNOWN   int = 0
	SQL_SELECT    int = 4
	SQL_CREATE    int = 5
	SQL_PROCEDURE int = 10

	SQL_PARAM_INPUT       int = 1
	SQL_PARAM_OUTPUT      int = 2
	SQL_PARAM_INPUTOUTPUT int = 3
	SQL_PARAM_RETURNVALUE int = 6

	SQL_XG_C_CHAR int = 2
	SQL_XG_C_CLOB int = 41
	SQL_XG_C_BLOB int = 42
	SQL_XG_C_NULL int = -11

	BIND_PARAM_BY_NAME int = 62
	BIND_PARAM_BY_POS  int = 63
)

type connector struct {
	dsn string
}

// Connect implements driver.Connector interface.
// Connect returns a connection to the database.
func (self *connector) Connect(ctx context.Context) (driver.Conn, error) {

	obj := &xugusqlConn{conn: nil}
	connKeyValue := C.CString(self.dsn)

	defer func() {
		cgo_c_free(unsafe.Pointer(connKeyValue))
	}()

	pos := strings.Index(strings.ToUpper(self.dsn), "IPS=")
	if pos != -1 {
		IPS_COUNTER++
		re := cgo_xgc_connect_ips(connKeyValue, &obj.conn)
		if re < 0 {
			return nil, obj.get_error()
		}
	} else {
		re := cgo_xgc_connect(connKeyValue, &obj.conn)
		if re < 0 {
			return nil, obj.get_error()
		}
	}

	return obj, nil
}

// Driver implements driver.Connector interface.
// Driver returns &XuguDriver{}
func (self *connector) Driver() driver.Driver {
	return &XuguDriver{}
}
