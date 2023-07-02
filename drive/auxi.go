package drive

import (
	"C"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// Auxiliary Struct
type __Value struct {
	// Boolean value, if the value is true, it means that the data
	// type of the current field is a large object data type
	islob bool

	// A pointer of * C.char data type, usually
	// pointing to the address of the parameter value to be bound
	value *C.char
	plob  unsafe.Pointer

	// length usually specifies the true length of the parameter
	// data to be bound
	length C.int

	// buff usually specifies the memory buffer
	// size of the parameter data to be bound
	buff C.int

	// When parameter binding, specify the data type of the field in the table
	types int

	// Return code
	rcode C.int
}

type parse struct {
	// bind_type is used to identify the type of parameter binding.
	// Parameter binding types include binding by parameter name
	// and binding by parameter placeholder
	bind_type int

	// param_count is used to specify the number
	// of parameters that need to be bound in the SQL statement
	param_count int

	// When the parameter binding type is binding
	// by parameter name, param_names is a collection of parameter names
	param_names []*C.char

	Val []__Value

	// When the parameter binding type is binding
	// by parameter placeholder, position identifies the parameter position
	position int
}

type ParseParam interface {

	// Conversion parameter data type
	assertParamType(driver.Value, int) error

	// Number of parsing parameters
	assertParamCount(string) int

	// Parse parameter binding type (binding type by parameter name
	// and binding type by parameter position)
	assertBindType(string) int

	// Parse parameter name
	assertParamName(string) error
}

func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	__Par := make([]driver.Value, len(named))

	for pos, Param := range named {
		if len(Param.Name) > 0 {
			return nil, errors.New("driver does not support the use of Named Parameters")
		}
		__Par[pos] = Param.Value
	}

	return __Par, nil
}

func (self *parse) assertParamType(dV driver.Value, pos int) error {

	var dest __Value
	switch dV.(type) {

	case int64:
		srcv, ok := dV.(int64)
		if !ok {
			news := errorNews("int64")
			return errors.New(news)
		}

		S := strconv.FormatInt(srcv, 10)
		dest.value = C.CString(S)
		dest.length = C.int(strings.Count(S, "") - 1)
		dest.buff = dest.length + 1
		dest.islob = false
		dest.types = SQL_XG_C_CHAR

	case float32:
		srcv, ok := dV.(float64)
		if !ok {
			news := errorNews("float32")
			return errors.New(news)
		}

		S := strconv.FormatFloat(srcv, 'f', 6, 64)
		dest.value = C.CString(S)
		dest.length = C.int(strings.Count(S, "") - 1)
		dest.buff = dest.length + 1
		dest.islob = false
		dest.types = SQL_XG_C_CHAR

	case float64:
		srcv, ok := dV.(float64)
		if !ok {
			news := errorNews("float64")
			return errors.New(news)
		}

		S := strconv.FormatFloat(srcv, 'f', 15, 64)
		dest.value = C.CString(S)
		dest.length = C.int(strings.Count(S, "") - 1)
		dest.buff = dest.length + 1
		dest.islob = false
		dest.types = SQL_XG_C_CHAR

	case bool:
		srcv, ok := dV.(bool)
		if !ok {
			news := errorNews("bool")
			return errors.New(news)
		}

		S := strconv.FormatBool(srcv)
		dest.value = C.CString(S)
		dest.length = C.int(strings.Count(S, "") - 1)
		dest.buff = dest.length + 1
		dest.islob = false
		dest.types = SQL_XG_C_CHAR

	case string:
		srcv, ok := dV.(string)
		if !ok {
			news := errorNews("string")
			return errors.New(news)
		}

		dest.value = C.CString(srcv)

		// 在Go语言中 string 底层是通过 byte 数组实现的，一个汉字占3个字节
		dest.length = C.int(len(srcv))
		dest.buff = dest.length + 1
		dest.islob = false
		dest.types = SQL_XG_C_CHAR

		if dest.length == 0 {
			dest.length = 1
		}

	case time.Time:
		srcv, ok := dV.(time.Time)
		if !ok {
			news := errorNews("time.Time")
			return errors.New(news)
		}

		tm := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
			srcv.Year(), int(srcv.Month()), srcv.Day(),
			srcv.Hour(), srcv.Minute(), srcv.Second())

		dest.value = C.CString(tm)
		dest.length = C.int(strings.Count(tm, "") - 1)
		dest.buff = dest.length + 1
		dest.islob = false
		dest.types = SQL_XG_C_CHAR

	case []byte:
		re := cgo_xgc_new_lob(&dest.plob)
		if re < 0 {
			return errors.New("Cannot create new large object")
		}

		srcv, ok := dV.([]byte)
		if !ok {

			news := errorNews("[]byte")
			return errors.New(news)
		}

		cgo_xgc_put_lob_data(
			&dest.plob,
			unsafe.Pointer((*C.char)(unsafe.Pointer(&srcv[0]))),
			len(srcv))
		cgo_xgc_put_lob_data(&dest.plob, nil, -1)

		dest.value = nil
		dest.length = C.int(8)
		dest.buff = C.int(8)
		dest.islob = true
		dest.types = SQL_XG_C_BLOB

	case nil:
		dest.value = C.CString("xugusql")
		dest.length = 0
		dest.buff = C.int(strings.Count("xugusql", ""))
		dest.islob = false
		dest.types = SQL_XG_C_CHAR

	default:
		/* OTHER DATA TYPE */
		return errors.New("unknown data type")
	}

	self.position = pos
	self.Val = append(self.Val, dest)

	return nil
}

func errorNews(str string) string {
	return fmt.Sprintf("[%s] asserting data type failed.", str)
}

func (self *parse) assertParamCount(query string) int {

	if self.bind_type == 0 {
		self.bind_type = self.assertBindType(query)
	}

	switch self.bind_type {
	case BIND_PARAM_BY_POS:
		self.param_count = strings.Count(query, "?")
	case BIND_PARAM_BY_NAME:

		self.param_count = 0
		pos := 0
		phead := -1

		for true {
			pos = strings.IndexByte(query[phead+1:], ':')
			if pos == -1 {
				break
			}

			pos += phead + 1
			tmp := pos
			for tmp > phead {
				tmp--
				if query[tmp] == ' ' {
					continue
				}

				if query[tmp] == ',' || query[tmp] == '(' {
					self.param_count++
				}
				break
			}
			phead = pos
		}
	}

	return self.param_count
}

func (self *parse) assertBindType(query string) int {

	self.bind_type = strings.IndexByte(query, '?')
	if self.bind_type != -1 {
		return BIND_PARAM_BY_POS
	}

	return BIND_PARAM_BY_NAME
}

func (self *parse) assertParamName(query string) error {

	if self.param_count <= 0 {
		self.assertParamCount(query)
	}

	pos := 0
	phead := -1

	for true {
		pos = strings.IndexByte(query[phead+1:], ':')
		if pos == -1 {
			break
		}

		pos += phead + 1
		tmp := pos
		for tmp > phead {
			tmp--
			if query[tmp] == ' ' {
				continue
			}

			// Parse parameter positions bound by parameter name
			if query[tmp] == ',' || query[tmp] == '(' {
				parg := pos
				for true {
					parg++
					if query[parg] == ',' || query[parg] == ')' || query[parg] == ' ' {
						self.param_names = append(self.param_names, C.CString(query[pos+1:parg]))
						break
					}
				}
			}
			break
		}

		phead = pos
	}

	return nil
}
