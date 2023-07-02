package drive

import (
	"C"
	"database/sql/driver"
	"errors"
	"unsafe"
)

type xugusqlStmt struct {

	// Context connection handle pointer
	stmt_conn unsafe.Pointer

	// Boolean value, used to identify whether
	// the executed SQL statement has been prepared
	prepared bool
	// Accept the prepared code for
	// the prepared SQL statement
	prename *C.char
	// Boolean value used to identify
	// whether the cursor is enabled
	curopend bool
	// Cursor name
	curname *C.char
	//  The number of parameters
	// in the executed SQL statement
	param_count int
	mysql       string
	// Context result set handle pointer
	result unsafe.Pointer
}

/* Collect error information from the database server */
func (self *xugusqlStmt) get_error() error {
	message := cgo_c_calloc(ERROR_BUFF_SIZE)
	defer func() {
		cgo_c_free(unsafe.Pointer(message))
	}()

	var length C.int
	cgo_xgc_error(&self.stmt_conn, message, &length)
	return errors.New(C.GoString(message))
}

/* {{ */
func (self *xugusqlStmt) Close() error {

	if self.curopend {
		re := cgo_xgc_close_cursor(&self.stmt_conn, self.curname)
		if re < 0 {
			return self.get_error()
		}

		cgo_c_free(unsafe.Pointer(self.curname))
		self.curname = nil
		self.curopend = false
	}

	if self.prepared {
		re := cgo_xgc_unprepare(&self.stmt_conn, self.prename)
		if re < 0 {
			return self.get_error()
		}

		cgo_c_free(unsafe.Pointer(self.prename))
		self.prename = nil
		self.prepared = false
	}

	return nil
}

/* {{ */
func (self *xugusqlStmt) NumInput() int {

	parser := &parse{
		bind_type:   0,
		param_count: 0,
		position:    0,
	}

	return parser.assertParamCount(self.mysql)
}

// Exec executes a prepared statement with the given arguments and
// returns a Result summarizing the effect of the statement.
func (self *xugusqlStmt) Exec(args []driver.Value) (driver.Result, error) {

	sql := C.CString(self.mysql)
	switch cgo_xgc_sql_type(sql) {
	case SQL_SELECT:
		return nil, errors.New("Exec does not support queries")
	}

	if !self.prepared {
		return nil, errors.New("SQL statement is not Prepared")
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

		if len(parser.Val) != parser.assertParamCount(self.mysql) {
			return nil, errors.New("The number of parameters does not match")
		}

		switch parser.assertBindType(self.mysql) {
		case BIND_PARAM_BY_POS:
			for pos, param := range parser.Val {
				if !param.islob {
					re := cgo_xgc_bindparambypos(&self.stmt_conn, pos+1,
						SQL_PARAM_INPUT, param.types,
						unsafe.Pointer(param.value), param.buff, &param.length)
					if re < 0 {
						return nil, self.get_error()
					}
				} else {
					re := cgo_xgc_bindparambypos(&self.stmt_conn, pos+1,
						SQL_PARAM_INPUT, param.types,
						unsafe.Pointer(&param.plob), param.buff, &param.length)
					if re < 0 {
						return nil, self.get_error()
					}
				}
			}

		case BIND_PARAM_BY_NAME:
			parser.assertParamName(self.mysql)
			for pos, param := range parser.Val {
				if !param.islob {
					re := cgo_xgc_bindparambyname(&self.stmt_conn, parser.param_names[pos],
						SQL_PARAM_INPUT, param.types, unsafe.Pointer(param.value),
						param.buff, &param.rcode, &param.length)
					if re < 0 {
						return nil, self.get_error()
					}
				} else {
					re := cgo_xgc_bindparambyname(&self.stmt_conn, parser.param_names[pos],
						SQL_PARAM_INPUT, param.types, unsafe.Pointer(&param.plob),
						param.buff, &param.rcode, &param.length)
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

	result := &xugusqlResult{
		affectedRows: 0,
		insertId:     0,
	}

	re := cgo_xgc_execute(&self.stmt_conn, self.prename, self.curname, &self.result)
	if re < 0 {
		return nil, self.get_error()
	}

	var pCT, pCC, pRC, pEC C.int
	var pID = cgo_c_calloc(ROWID_BUFF_SIZE)

	re = cgo_xgc_get_result_set(&self.result, &pCT, &pCC, &pRC, &pEC, pID)
	if re < 0 {
		return nil, self.get_error()
	}

	cgo_c_free(unsafe.Pointer(pID))
	result.affectedRows = int64(pEC)

	return result, nil
}

// QueryContext executes a prepared query statement with the given arguments
// and returns the query results as a *Rows.
func (self *xugusqlStmt) Query(args []driver.Value) (driver.Rows, error) {

	sql := C.CString(self.mysql)
	if cgo_xgc_sql_type(sql) != SQL_SELECT {
		return nil, errors.New("The executed SQL statement is not a SELECT")
	}

	if !self.prepared {
		return nil, errors.New("SQL statement is not Prepared")
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

		if len(parser.Val) != parser.assertParamCount(self.mysql) {
			return nil, errors.New("The number of parameters does not match")
		}

		switch parser.assertBindType(self.mysql) {
		case BIND_PARAM_BY_POS:
			for pos, param := range parser.Val {
				if !param.islob {
					re := cgo_xgc_bindparambypos(&self.stmt_conn, pos+1,
						SQL_PARAM_INPUT, param.types,
						unsafe.Pointer(param.value), param.buff, &param.length)
					if re < 0 {
						return nil, self.get_error()
					}
				} else {

					re := cgo_xgc_bindparambypos(&self.stmt_conn, pos+1,
						SQL_PARAM_INPUT, param.types,
						unsafe.Pointer(&param.plob), param.buff, &param.length)
					if re < 0 {
						return nil, self.get_error()
					}
				}
			}

		case BIND_PARAM_BY_NAME:
			parser.assertParamName(self.mysql)
			for pos, param := range parser.Val {
				if !param.islob {
					re := cgo_xgc_bindparambyname(&self.stmt_conn,
						parser.param_names[pos],
						SQL_PARAM_INPUT, param.types, unsafe.Pointer(param.value),
						param.buff, &param.rcode, &param.length)
					if re < 0 {
						return nil, self.get_error()
					}
				} else {

					re := cgo_xgc_bindparambyname(&self.stmt_conn,
						parser.param_names[pos],
						SQL_PARAM_INPUT, param.types, unsafe.Pointer(&param.plob),
						param.buff, &param.rcode, &param.length)
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

	//if self.curname == nil {
	//	self.curname = cgo_c_calloc(CURSOR_NAME_BUFF_SIZE)
	//}

	re := cgo_xgc_execute(&self.stmt_conn, self.prename, self.curname, &self.result)
	if re < 0 {
		return nil, self.get_error()
	}

	//re = cgo_xgc_fetch_with_cursor(&self.stmt_conn, self.curname, &self.result)
	//if re < 0 {
	//	return nil, self.get_error()
	//}

	//self.curopend = true
	return &xugusqlRows{
		result:    self.result,
		prepared:  self.prepared,
		rows_conn: self.stmt_conn,
	}, nil

}
