package drive

import (
	"unsafe"
)

/*
#cgo CFLAGS : -I/usr/include
#cgo LDFLAGS : -L/usr/lib64 -lxugusql

#include <stdlib.h>
#include <string.h>
#include "xugusql.h"
*/
import "C"

var IPS_COUNTER int = 0
var IPS_BODY unsafe.Pointer

/* Collect error information from the database server */
func cgo_xgc_error(__pConn *unsafe.Pointer, pLog *C.char, act *C.int) int {
	return int(C.XGC_GetError(__pConn, pLog, act))
}

/*
 * 'C.XGC_OpenConn' is used to establish a new connection session with XGDB,
 * return value:
 *      (int) 2 : Success               (int)-1 : Failure
 *      (int)-8 : TCP/IP socket error.  (int)-9 : Login xgdb failure.
 */
func cgo_xgc_connect(pdsn *C.char, __pConn *unsafe.Pointer) int {
	return int(C.XGC_OpenConn(pdsn, __pConn))
}

/* 'C.XGC_OpenConn_Ips' is used to establish a new connection session with XGDB,
 * it is different from'C.XGC_OpenConn' in that'C.XGC_OpenConn_Ips' can achieve
 * connection load balancing between distributed database nodes.
 */
func cgo_xgc_connect_ips(pdsn *C.char, __pConn *unsafe.Pointer) int {
	return int(C.XGC_OpenConn_Ips(pdsn, C.int(IPS_COUNTER), &IPS_BODY, __pConn))
}

/*
 * The cgo-level call,
 * to realize the user's memory allocation application.
 */
func cgo_c_calloc(Size uint) *C.char {
	return (*C.char)(C.calloc(C.ulong(1), C.ulong(Size)))
}

/*
 * The cgo-level call,
 * to cleans up the data in the memory requested by cgo_c_calloc.
 */
func cgo_c_memset(pointer *C.char, length uint) {
	C.memset(unsafe.Pointer(pointer), 0x0, C.ulong(length))
}

/*
 * The cgo-level call,
 * releases the memory requested by cgo_c_calloc.
 */
func cgo_c_free(__Pr unsafe.Pointer) {
	C.free(__Pr)
}

// Execute SQL statements without result set return, including DDL and DML
func cgo_xgc_execnoquery(__pConn *unsafe.Pointer, query *C.char) int {
	return int(C.XGC_Execute_no_query(__pConn, query))
}

/*
 * Return the type of the SQL statement,
 * confirm it is DDL, DML and DQL.
 */
func cgo_xgc_sql_type(sql *C.char) int {
	return int(C.fun_sql_type(sql))
}

/*
 * Binding parameters,
 * the binding method uses the form of placeholders.
 */
func cgo_xgc_bindparambypos(__pConn *unsafe.Pointer, seq int, ArgType int,
	Type int, Valu unsafe.Pointer, Buff C.int, act *C.int) int {
	return int(C.XGC_BindParamByPos(__pConn, C.int(seq), C.int(ArgType),
		C.int(Type), Valu, Buff, act))
}

/*
 * Binding parameters,
 * the binding method uses the form of the parameter name.
 */
func cgo_xgc_bindparambyname(__pConn *unsafe.Pointer, Name *C.char, ArgType int,
	Type int, Valu unsafe.Pointer, Buff C.int, Rcode *C.int, act *C.int) int {
	return int(C.XGC_BindParamByName(__pConn, Name, C.int(ArgType), C.int(Type),
		Valu, Buff, Rcode, act))
}

/*
 * Disconnect the database session connection established
 * by'C.XGC_OpenConn_Ips'.
 */
func cgo_xgc_disconnect(__pConn *unsafe.Pointer) int {
	return int(C.XGC_CloseConn(__pConn))
}

// Prepare the executed SQL statement.
func cgo_xgc_prepare(__pConn *unsafe.Pointer, query *C.char, prename *C.char) int {
	return int(C.XGC_Prepare2(__pConn, query, prename))
}

// Execute the SQL statement prepared by'C.XGC_Prepare2'.
func cgo_xgc_execute(__pConn *unsafe.Pointer, prename *C.char,
	curname *C.char, res *unsafe.Pointer) int {
	return int(C.XGC_Execute2(__pConn, prename, curname, res))
}

// Cancel the SQL statement prepared by'C.XGC_Prepare2'.
func cgo_xgc_unprepare(__pConn *unsafe.Pointer, prename *C.char) int {
	return int(C.XGC_UnPrepare(__pConn, prename))
}

// Close server cursor.
func cgo_xgc_close_cursor(__pConn *unsafe.Pointer, curname *C.char) int {
	return int(C.XGC_CloseCursor(__pConn, curname))
}

// Receive the result set from the database server.
func cgo_xgc_get_result_set(__pConn *unsafe.Pointer, pCT *C.int, pCC *C.int,
	pRC *C.int, pEC *C.int, pID *C.char) int {
	return int(C.XGC_getResultRet(__pConn, pCT, pCC, pRC, pEC, pID))
}

// Release result set.
func cgo_xgc_free_rowset(__pRes *unsafe.Pointer) int {
	return int(C.XGC_FreeRowset(__pRes))
}

// Get data in the form of a cursor.
func cgo_xgc_fetch_with_cursor(__pConn *unsafe.Pointer,
	curname *C.char, __pRes *unsafe.Pointer) int {
	return int(C.XGC_FetchServerCursorRowset(__pConn, curname, __pRes))
}

// Get the column name of the specified column.
func cgo_xgc_get_column_name(__pRes *unsafe.Pointer, Seq int, cname *C.char) int {
	return int(C.XGC_getResultcolname(__pRes, C.int(Seq), cname))
}

// Get the number of fields in the current query.
func cgo_xgc_get_fields_count(__pRes *unsafe.Pointer, CCnt *C.int) int {
	return int(C.XGC_getResultColumnsnum(__pRes, CCnt))
}

// Get the next row of result set data.
func cgo_xgc_read_next(__pRes *unsafe.Pointer) int {
	return int(C.XGC_ReadNext(__pRes))
}

// Get the number of rows in the result set.
func cgo_xgc_get_rows_count(__pRes *unsafe.Pointer, Rows *C.int) int {
	return int(C.XGC_getResultRecordnum(__pRes, Rows))
}

// Get the next result set.
func cgo_xgc_next_result(__pRes *unsafe.Pointer) int {
	return int(C.XGC_NextResult(__pRes))
}

func cgo_xgc_exec_with_cursor(__pConn *unsafe.Pointer, query *C.char,
	curname *C.char, __pRes *unsafe.Pointer, fields *C.int, rows *C.longlong, effects *C.int) int {
	return int(C.XGC_ExecwithServerCursorReader(__pConn, query, curname, __pRes, fields, rows, effects))
}

// Get the column data type of the specified column.
func cgo_xgc_get_column_type(__pRes *unsafe.Pointer, Seq int, ColuType *C.int) int {
	return int(C.XGC_getResultcolType(__pRes, C.int(Seq), ColuType))
}

// Get the data of the specified column.
func cgo_xgc_get_data(__pRes *unsafe.Pointer, Seq int, tartype int,
	pVal *C.char, Buff uint, act *C.int) int {
	return int(C.XGC_GetData(__pRes, C.int(Seq), C.int(tartype), unsafe.Pointer(pVal), C.int(Buff), act))
}

// Obtain large object data.
func cgo_xgc_get_lob(__pRes *unsafe.Pointer, Seq int, tartype int,
	__pLob *unsafe.Pointer, Buff uint, act *C.int) int {
	return int(C.XGC_GetData(__pRes, C.int(Seq), C.int(tartype), unsafe.Pointer(__pLob), C.int(Buff), act))
}

// Create a large object data box.
func cgo_xgc_new_lob(__pLob *unsafe.Pointer) int {
	return int(C.XGC_Create_Lob(__pLob))
}

// Obtain large object data.
func cgo_xgc_get_lob_data(__pLob *unsafe.Pointer, pVal unsafe.Pointer, act C.int) int {
	return int(C.XGC_Get_Lob_data(__pLob, pVal, act))
}

// Obtain large object data.
func cgo_xgc_put_lob_data(__pLob *unsafe.Pointer, pVal unsafe.Pointer, act int) int {
	return int(C.XGC_Put_Lob_data(__pLob, pVal, C.int(act)))
}

// Release large object data resources.
func cgo_xgc_lob_distroy(__pLob *unsafe.Pointer) int {
	return int(C.XGC_Distroy_Lob(__pLob))
}

// Receive data from the database server.
func cgo_xgc_exec_with_reader(__pConn *unsafe.Pointer, Sql *C.char,
	__pRes *unsafe.Pointer, fieldCount *C.int, rowCount *C.longlong, effectCount *C.int) int {
	return int(C.XGC_ExecwithDataReader(__pConn, Sql, __pRes, fieldCount, rowCount, effectCount))
}

/* }}*/
