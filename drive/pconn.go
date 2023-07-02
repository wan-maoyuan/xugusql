package drive

import (
	"C"
	"context"
	"database/sql/driver"
	"errors"
	"unsafe"
)

type xugusqlConn struct {
	conn         unsafe.Pointer
	affectedRows int
	insertId     int
}

func (self *xugusqlConn) get_error() error {
	message := cgo_c_calloc(ERROR_BUFF_SIZE)
	defer func() {
		cgo_c_free(unsafe.Pointer(message))
	}()

	var length C.int
	cgo_xgc_error(&self.conn, message, &length)
	return errors.New(C.GoString(message))
}

func (self *xugusqlConn) Begin() (driver.Tx, error) {
	err := self.exec("set auto_commit off;")
	if err != nil {
		return nil, self.get_error()
	}
	return &xugusqlTx{tconn: self}, nil
}

func (self *xugusqlConn) Close() error {
	re := cgo_xgc_disconnect(&self.conn)
	if re < 0 {
		return self.get_error()
	}
	return nil
}

func (self *xugusqlConn) Prepare(query string) (driver.Stmt, error) {
	sql := C.CString(query)
	defer func() {
		cgo_c_free(unsafe.Pointer(sql))
	}()

	switch cgo_xgc_sql_type(sql) {
	case SQL_PROCEDURE:
		return nil, errors.New("prepare does not support stored procedures")
	case SQL_UNKNOWN:
		return nil, errors.New("unknown SQL statement type")
	case SQL_CREATE:
		return nil, errors.New("prepare does not support DDL")
	}

	stmt := &xugusqlStmt{
		stmt_conn:   self.conn,
		prepared:    false,
		prename:     nil,
		curopend:    false,
		curname:     nil,
		param_count: 0,
		result:      nil,
		mysql:       query,
	}

	if stmt.prename == nil {
		stmt.prename = cgo_c_calloc(PREPARE_NAME_BUFF_SIZE)
	}

	re := cgo_xgc_prepare(&self.conn, sql, stmt.prename)
	if re < 0 {
		return nil, self.get_error()
	}

	stmt.prepared = true

	return stmt, nil
}

func (self *xugusqlConn) Query(query string,
	args []driver.Value) (driver.Rows, error) {
	sql := C.CString(query)
	defer func() {
		cgo_c_free(unsafe.Pointer(sql))
	}()

	parser := &parse{
		bind_type:   0,
		param_count: 0,
		position:    0,
	}

	if len(args) != 0 {
		for pos, param := range args {
			err := parser.assertParamType(param, pos)
			if err != nil {
				return nil, err
			}
		}

		if len(parser.Val) != parser.assertParamCount(query) {
			return nil, errors.New("the number of parameters does not match")
		}

		switch parser.assertBindType(query) {

		case BIND_PARAM_BY_POS:
			for pos, param := range parser.Val {
				if !param.islob {
					re := cgo_xgc_bindparambypos(&self.conn, pos+1, SQL_PARAM_INPUT,
						param.types, unsafe.Pointer(param.value), param.buff, &param.length)
					if re < 0 {
						return nil, self.get_error()
					}
				} else {

					re := cgo_xgc_bindparambypos(&self.conn, pos+1, SQL_PARAM_INPUT,
						param.types, unsafe.Pointer(&param.plob), param.buff, &param.length)
					if re < 0 {
						return nil, self.get_error()
					}
				}

			}

		case BIND_PARAM_BY_NAME:
			_ = parser.assertParamName(query)

			for pos, param := range parser.Val {

				if !param.islob {
					re := cgo_xgc_bindparambyname(&self.conn, parser.param_names[pos], SQL_PARAM_INPUT,
						param.types, unsafe.Pointer(param.value), param.buff, &param.rcode, &param.length)
					if re < 0 {
						return nil, self.get_error()
					}
				} else {

					re := cgo_xgc_bindparambyname(&self.conn, parser.param_names[pos], SQL_PARAM_INPUT,
						param.types, unsafe.Pointer(&param.plob), param.buff, &param.rcode, &param.length)
					if re < 0 {
						return nil, self.get_error()
					}
				}
			}

		}
	}

	defer func() {
		for pos, param := range parser.Val {
			if parser.bind_type == BIND_PARAM_BY_NAME {
				cgo_c_free(unsafe.Pointer(parser.param_names[pos]))
			}

			if !param.islob {
				cgo_c_free(unsafe.Pointer(param.value))
			} else {
				cgo_xgc_lob_distroy(&param.plob)
			}
		}
	}()

	rows := &xugusqlRows{
		rows_conn:   self.conn,
		result:      nil,
		lastRowRelt: 0,
		lastRelt:    0,
		prepared:    false,
	}

	var fieldCount, effectCount C.int
	var rowCount C.longlong

	re := cgo_xgc_exec_with_reader(&self.conn, sql, &rows.result,
		&fieldCount, &rowCount, &effectCount)
	if re < 0 {
		return nil, self.get_error()
	}

	return rows, nil
}

func (self *xugusqlConn) Exec(query string,
	args []driver.Value) (driver.Result, error) {
	sql := C.CString(query)
	switch cgo_xgc_sql_type(sql) {
	case SQL_SELECT:
		return nil, errors.New("exec does not support queries")
	case SQL_UNKNOWN:
		return nil, errors.New("unknown SQL statement type")
	}

	parser := &parse{
		bind_type:   0,
		param_count: 0,
		position:    0,
	}

	if len(args) != 0 {
		for pos, param := range args {
			err := parser.assertParamType(param, pos)
			if err != nil {
				return nil, err
			}
		}

		if len(parser.Val) != parser.assertParamCount(query) {
			return nil, errors.New("the number of parameters does not match")
		}

		switch parser.assertBindType(query) {
		case BIND_PARAM_BY_POS:
			for pos, param := range parser.Val {
				if !param.islob {
					re := cgo_xgc_bindparambypos(&self.conn, pos+1, SQL_PARAM_INPUT,
						param.types, unsafe.Pointer(param.value), param.buff, &param.length)
					if re < 0 {
						return nil, self.get_error()
					}
				} else {
					re := cgo_xgc_bindparambypos(&self.conn, pos+1, SQL_PARAM_INPUT,
						param.types, unsafe.Pointer(&param.plob), param.buff, &param.length)
					if re < 0 {
						return nil, self.get_error()
					}
				}

			}

		case BIND_PARAM_BY_NAME:
			_ = parser.assertParamName(query)
			for pos, param := range parser.Val {
				if !param.islob {
					re := cgo_xgc_bindparambyname(&self.conn, parser.param_names[pos], SQL_PARAM_INPUT,
						param.types, unsafe.Pointer(param.value), param.buff, &param.rcode, &param.length)
					if re < 0 {
						return nil, self.get_error()
					}
				} else {

					re := cgo_xgc_bindparambyname(&self.conn, parser.param_names[pos], SQL_PARAM_INPUT,
						param.types, unsafe.Pointer(&param.plob), param.buff, &param.rcode, &param.length)
					if re < 0 {
						return nil, self.get_error()
					}
				}
			}
		}
	}

	defer func() {
		cgo_c_free(unsafe.Pointer(sql))
		for pos, param := range parser.Val {
			if parser.bind_type == BIND_PARAM_BY_NAME {
				cgo_c_free(unsafe.Pointer(parser.param_names[pos]))
			}

			if !param.islob {
				cgo_c_free(unsafe.Pointer(param.value))
			} else {
				cgo_xgc_lob_distroy(&param.plob)
			}
		}
	}()

	self.affectedRows = 0
	self.insertId = 0

	err := self.exec(query)
	if err == nil {
		return &xugusqlResult{
			affectedRows: int64(self.affectedRows),
			insertId:     int64(self.insertId),
		}, nil
	}

	return nil, err
}

func (self *xugusqlConn) exec(query string) error {
	sql := C.CString(query)
	defer func() {
		cgo_c_free(unsafe.Pointer(sql))
	}()

	self.affectedRows = cgo_xgc_execnoquery(&self.conn, sql)
	if self.affectedRows < 0 {
		return self.get_error()
	}

	return nil
}

func (self *xugusqlConn) ExecContext(ctx context.Context,
	query string, args []driver.NamedValue) (driver.Result, error) {

	Value, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}

	return self.Exec(query, Value)
}

func (self *xugusqlConn) Ping(ctx context.Context) error {

	sql := C.CString("SELECT COUNT(*) FROM dual;")
	defer func() {
		cgo_c_free(unsafe.Pointer(sql))
	}()

	var fieldCount, effectCount C.int
	var rowCount C.longlong
	var result unsafe.Pointer

	re := cgo_xgc_exec_with_reader(&self.conn, sql, &result,
		&fieldCount, &rowCount, &effectCount)
	if re < 0 {
		return self.get_error()
	}

	return nil
}
