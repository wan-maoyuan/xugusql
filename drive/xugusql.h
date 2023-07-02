
#ifndef _XG_DRIVERAPI_H_ 
#define _XG_DRIVERAPI_H_ 

#ifdef WIN32 
#define XG_API __cdecl
#else
//linuxs about
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <fcntl.h>
#include <netdb.h>
#include <grp.h>
#include <arpa/inet.h>
#include <sys/file.h>
#include <sys/types.h>
#include <sys/stat.h>
#define XG_API
#endif
typedef long long int64;

#ifdef __cplusplus
extern "C"{
#endif 

#define XGC_ATTR_SERVER_VERSION   1 
#define XGC_ATTR_DBNAME           2 
#define XGC_ATTR_ISO_LEVEL        3
#define XGC_ATTR_SERVER_CHARSET   4
#define XGC_ATTR_CLIENT_CHARSET   5

#define XGC_ATTR_USESSL           6
#define XGC_ATTR_SRV_TURN_IPS      7
#define XGC_ATTR_TIMEZONE          8
#define XGC_ATTR_LOB_DESCRIBER     9
#define XGC_ATTR_AUTOCOMMIT       11
#define XGC_ATTR_STMT_SERVER_CURSOR   12

#define XGC_ATTR_USE_CURSOR        0
#define XGC_ATTR_NOTUSE_CURSOR     1
#define XGC_ATTR_USE_CURSORDEFAULT 0

typedef enum tagPARAMINOUT_TYPE {
		PARAM_INPUT = 1,
		PARAM_OUTPUT = 2,
		PARAM_INPUTOUTPUT = 3,
		PARAM_RETURNVALUE = 6,
}PARAMINOUT_TYPE;
#define XGC_ATTR_COL_COUNT  61
#define XGC_ATTR_ROW_COUNT  62
#define XGC_ATTR_EFFECT_NUM 63
#define XGC_ATTR_RESULT_TYPE 64
#define XGC_ATTR_SQL_TYPE 65
#define XGC_ATTR_IS_MUTIRESULT  66

#define   XGC_ISO_READONLY    1
#define   XGC_ISO_READCOMMIT  2
#define   XGC_ISO_READREPEAT  3
#define   XGC_ISO_SERIAL      4
#define   XGC_CHARSET_GBK     1
#define   XGC_CHARSET_GB2312  2
#define   XGC_CHARSET_UTF8    3


#define  XG_C_NULL                    0
#define  XG_C_BOOL                    1
#define  XG_C_CHAR                    2
#define  XG_C_TINYINT                 3
#define  XG_C_SHORT                   4
#define  XG_C_INTEGER                 5
#define  XG_C_BIGINT                  6
#define  XG_C_FLOAT                   7
#define  XG_C_DOUBLE                  8
#define  XG_C_NUMERIC                 9
#define  XG_C_DATE	                  10
#define  XG_C_TIME			          11
#define  XG_C_TIME_TZ                 12
#define  XG_C_DATETIME                13
#define  XG_C_DATETIME_TZ             14
#define  XG_C_BINARY                  15


#define DATETIME_ASLONG 23 
#define  XG_C_NVARBINARY        18 
#define  XG_C_REFCUR       58
#define  XG_C_CHARN1  63

#define  XG_C_NCHAR                   62   /* only for c# wchar use */ 

#define  XG_C_INTERVAL                21
#define  XG_C_INTERVAL_YEAR_TO_MONTH  28
#define  XG_C_INTERVAL_DAY_TO_SECOND  31

#define  XG_C_TIMESTAMP     XG_C_DATETIME
#define  XG_C_LOB                     40
#define  XG_C_CLOB                    41
#define  XG_C_BLOB                    42

#define XG_SUCCESS              0
#define XG_NO_DATA              100  

#define XG_ERROR               -1
#define XG_NET_ERROR           -4
#define XG_INVALID_ARG         -3
#define XG_SOCKET_ERROR        -8 
#define XG_LOGIN_ERROR         -9 

#define XG_NULL_DATA           -11
#define XG_TRUNCATED_DATA      -12
#define XG_DATATYPE_ERROR      -13      /* Data type cannot be converted */  
#define XG_FLOW_DATA           -14      /* Data type out of bounds */
#define XG_COL_SEQ_ERR         -15      /* Data serial number out of bounds */
#define XG_COL_EXCEPT_DATAOFF  -18      /* Data offset out of bounds */ 

#define XG_COL_DATA_OVERFLOW    98   





int XG_API  SetConnStr(char* str, void** p_conn);
int XG_API  GetConnStr(char* str, void** p_conn);
 
/* return :
 *   2 : Successful connection
 *  -1 : Connection string incoming error
 *  -8 : Failed to create sock
 *  -9 : Login database failed
 *  */
int XG_API XGC_OpenConn(char* Conn_str,void** p_conn);

/* return :
 *   0 : Success
 *  -1 : Fail
 * */
int XG_API XGC_CloseConn(void** p_conn);

/* return :
 *   2 : Successful connection
 *  -3 : Parameter error 
 *  -8 : Failed to create sock
 *  -9 : Login database failed
 *  */
int XG_API XGC_OpenConn_Ips(char* Conn_str,int ntimes,void** turnIP_attrs,void** p_conn);


/* Explicitly create parameters 
 * return :
 *   0 : Success
 *  -3 : Parameter error
 * */
int XG_API XGC_CreateParams(void** p_params);


/* Reset parameters in connection 
 * Note: explicitly created parameters will not be processed
 * */
int XG_API XGC_ResetParams(void** p_conn);


/* Bind explicitly created parameters to the connection */
/* X//将显式创建的参数绑定到连接上
* 配合 XGC_CreateParams 使用
* p_conn 连接指针
* p_params 显式创建的参数结构指针
*返回值  成功返回0 参数错误 传入类型不匹配 返回 -3 ;
*/
int XG_API XGC_BindParams2Conn(void** p_conn,void** p_params);


int XG_API XGC_BindParamByName(void** p_conn, char* param_name, int param_type, 
           int datatype, void*  value, int param_size,  int* rt_code,  int* rlen_val);

/* 参数 按名进行批量绑定
* p_conn 连接句柄 或显式申明的参数句柄
* param_name 参数在sql中的名
* param_num sql中 按名绑定参数的个数
* param_type 参数输入输出型  1 输入 2 输出 3输入输出  6 返回值 ，
* datatype 参数C类型
* array_size 参数数组长度-参数的批量的个数
* array_value 参数数组 首地址
* param_size  参数固定长度， 变长的填入总体长度（ 内部 长度值 长度值 这样)
* rlen_val  int型 数组  存放参数数组中 数组内每个元素的实际长度 按组元序号对应
返回值：正确返回 0 错误返回 -1 ，参数传入错误返回-3  ，参数名错误 -53 ；
*/
int XG_API XGC_BindParamArrayByName(void** p_conn, char* param_name,int param_num, 
     int param_type,int datatype, int array_size, void* array_value, 
           int param_size, int * rlen_val);
//按序号绑定 2 种用法
/*=====================================
* p_conn     连接句柄 （隐式参数句柄）   2 p_conn  参数句柄（显式创建参数句柄）
* param_no   参数号： 从1开始
* param_type 参数输入输出型 1236
* datatype   参数数据类型
* value      参数值
* param_size 单个参数的空间大小 buffer
* rlen_val   具体的每个参数 的对应实际大小
返回值 正确返回 0  ；传入指针参数错误  返回 -3 ;参数序号超界 -51； 参数输入输出型错：-52 ；参数号小于1 -54 ；参数跳跃未按序 -55； 尚未实现功能 -8；
======================================*/
int XG_API XGC_BindParamByPos(void** p_conn, int param_no,int param_type, 
                int datatype, void* value, int param_size, int * rlen_val);
/*批量按序号绑定
* p_conn 连接句柄 或显式申明的参数句柄
* param_no 参数号 从1 开始
* param_num sql中 按名绑定参数的个数
* param_type 参数输入输出型  1 输入 2 输出 3输入输出  6 返回值 ，
* datatype 参数C类型
* array_size 参数数组长度-参数的批量的个数
* array_value 参数数组 首地址
* param_size  参数固定长度， 变长的填入总体长度（ 内部 长度值 长度值 这样)
* rlen_val  int型 数组  存放参数数组中 数组内每个元素的实际长度 按组元序号对应
* 返回值：正确返回 0；传入指针参数错误  返回 -3 ;参数序号超界 -51； 参数输入输出型错：-52 ；参数号小于1 -54 ；  尚未实现功能 -8；
*/
int XG_API XGC_BindParamArrayByPos(void** p_conn, int param_no, int param_num, 
     int param_type,int datatype, int array_size, void* array_value, int param_size, int * rlen_val); 
 

/* SQL execution without result set return */

/* 无结果集返回的sql执行 --支持DDL ,insert update ，delete等执行
* p_conn 连接指针 ，cmd_sql sql语句 ，如sql里面有参数 请提前在 p_conn连接句柄里面绑定
*返回值：  update 和delete时 返回影响的行数 ，insert 返回插入行数， 其他成功返回0 ，一般错误返回-1 ； 网络错 -4；
无结果集返回的执行，最多支持影响的行数 rowid 这些
*/
int XG_API XGC_Execute_no_query(void** p_conn,char* cmd_sql); 

/* 查询 返回首行首列
*查询简便化封装，返回结果集的首行首列 --
根据type 来解析re_val 数据为数值的是定长 数据是变长的 re_val 为长度（4字节int)+指向数据的指针 （或者是数组）
* p_conn 连接指针 ，cmd_sql sql语句 常为 select count（*） 等
* re_val  存放值的buffer缓存区， 一般为字符串返回。
* type 空值 时返回 0 ，，如果值为101 说明buff空间不足返回的是指向值的指针
返回值 ：成功 返回 0  网络错 -4；一般错误 -1； ，insert返回1 ，update 返回2 ；delete  返回3 ；
*/
int XG_API XGC_Execute_query_with_one(void** p_conn ,char* cmd_sql,void* re_val,int* type);

/* usage: prepare name can be given a specific name or NULL
 *  (1) If the SQL statement is a query, the prepare_name parameter can 
 *      be given a specific value.
 *  (2) If the SQL statement is not a query, the prepare_name parameter 
 *      must be NULL 
 * */
int XG_API XGC_Prepare2(void** p_conn,char* cmd_sql,char* prepare_name); 

/* usage: 
 *  (1) If the SQL statement is a query, when both prepare_name and servercursor_name 
 *      are given as NULL, it means that the server cursor is not used.
 *  (2) If the SQL statement is a query, when prepare_name and servercursor_name are 
 *      given specific values, it means that the server cursor is used.
 *  
 *  notice:
 *     If the 'prepare_name' parameter in the 'XGC_Prepare2' phase is NULL, 
 *     then the 'prepare_name' in 'XGC_Execute2' must also be NULL. 
 * */
int XG_API XGC_Execute2(void** p_conn, char* prepare_name, char* servercursor_name,void** pres);
int XG_API XGC_ExecBatch(void**  p_conn,char* cmd_sql, int ArrayCount);

int XG_API XGC_UnPrepare(void** p_conn,char* prepare_name);

/* 关闭服务器端游标
*p_conn 连接指针， * cursor_name 游标名 ，游标释放应 在unprepare之前调用
* 返回值 成功返回0 失败返回-1 ；网络错 返回-4 ；
*/
int XG_API XGC_CloseCursor(void** p_conn,char* cursor_name);
/*带返回结果集的 查询语句执行 生成reader
* *p_conn 连接指针，* cmd_sql 查询sql语句 ，
* * p_res 返回的结果集指针 ，
* 输出型参数 field_num 结果集的列数 ，   rowcount 结果集的行数   effected_num：  update delete 影响的行数 ，无则不填，
*返回值 成功返回0   ；网络错返回-4 ；失败返回-1 ；
*/
int XG_API XGC_ExecwithDataReader(void** p_conn ,char* cmd_sql,void** p_res,
                int* field_num,int64* rowcount,int* effected_num);

/* 
 * Get result set from server cursor
 * int XG_API XGC_FetchServerCursorRowset(void** p_conn ,char* cmd_sql,void** p_res);
 *
 * */
int XG_API XGC_FetchServerCursorRowset(void** p_conn ,char* servercursor_name,void** p_res);

int XG_API XGC_FetchServerCursorRowset_V2(void** p_conn, char* sql_cmd, void** p_res);

/* Fetching data from the server cursor header (extra)*/
int XG_API XGC_FetchRefCursorHead(void** p_conn ,char* Cursor_name ,void** p_res,
                int* field_num,int64* rowcount,int* cached);

/* 服务器游标获取数据   XGC_Prepare2+XGC_Execute2 (冗余项)
*  *p_conn 连接指针，* cmd_sql  需要游标执行的 select 查询sql语句
* Cursor_name 服务器端游标名 由用户自行命名后传入
* * p_res 返回的结果集指针 ，输出型参数 field_num 结果集的列数 ，   rowcount 结果集的行数   effected_num：  update delete 影响的行数
* 正常返回0 失败返回-1
*/
int XG_API XGC_ExecwithServerCursorReader(void** p_conn ,char* cmd_sql, 
    char* Cursor_name ,void** p_res,int* field_num,int64* rowcount,int* effected_num);

/*
 * Execution of stored procedures and functions, 
 * involving input and output of parameters
 * */
int XG_API XGC_Execute_procesure(void** p_conn , char*  cmd_sql,void* para); 

/*
 *  RESULT
 *
 * */
int XG_API  XGC_GetData(void** pTr_Result,int col_no,int TarCtype, 
           void* TarValuePtr,int BuffLen,int* lenPtr);

int XG_API  XGC_getResultcolType(void**  pTr_Result,int col_no,int* col_type) ;
int XG_API  XGC_getResultcolname(void**  pTr_Result,int col_no,char* col_name) ;
int XG_API  XGC_getResultcolseq(void**  pTr_Result,char* col_name);
/* 返回结果集的列个数
  输出参数 field_num 返回结果集列个数
  正常返回 0 ，输入结果集类型异常返回 -3
*/
int XG_API  XGC_getResultColumnsnum(void**  pTr_Result,int* field_num);
/* 返回结果集 行数
**  pTr_Result 结果集指针
*  输出参数 record_num 返回结果集 行数
*  正常返回 0 ，输入结果集类型异常返回 -3
*/
int XG_API  XGC_getResultRecordnum(void**  pTr_Result,int* record_num);
int XG_API  XGC_getResultcolmodi(void**  pTr_Result, int col_no, int* modi);//add 202-02-19

int XG_API  XGC_getResultColInfo(void**  pTr_Result,int col_no, 
    char* col_Tabname, char* col_name, char* col_alias, int* datatype,
         int* col_modi,int* col_flag);

/* Result set cursor moves to the next result set */
int XG_API XGC_ReadNext(void** p_res);

/* Release result set */
int XG_API XGC_FreeRowset(void** p_res);

/*
 * Get the next result set, suitable for multiple result sets
 * */
int XG_API XGC_NextResult(void** p_res);

/* Attribute */
int XG_API XGC_GetAttr(void** hd_ptr, int attrtype, void * ValuePtr, 
                int  BuffLen, int* ret_attr_type, int* re_len);
int XG_API XGC_SetAttr(void** hd_ptr, int attrtype, const void * ValuePtr, int  BuffLen);

/*
 * BLOB\CLOB
 * */
int XG_API XGC_Create_Lob(void** Lob_ptr);
int XG_API XGC_Put_Lob_data(void** Lob_ptr, void* data, int len );
int XG_API XGC_Get_Lob_data(void** Lob_ptr, void* data, int len);
int XG_API XGC_Distroy_Lob(void** Lob_ptr);
int XG_API XGC_LobWrite_SetPos(void** Lob_ptr,int setpos);
int XG_API XGC_LobRead_SetPos(void** Lob_ptr,int setpos);
int XG_API XGC_Reset_Lob(void** Lob_ptr);

/* 
 * ERROR INFO
 * */
int  XG_API XGC_GetError(void** hd_ptr, char* err_text,int* rlen);
int  XG_API XGC_GetErrorInfo(void** p_handptr, char* ccode, char* errmessage, int* rlen); 
int  XG_API XGC_GetErrorInfoOption(void** p_handptr, char* ccode, int * ret_code, 
                char* errmessage, int max_message_len, int* rlen);

/*
 *  OTHER
 *
 * */
void XG_API XGC_FreePtr(void**Ptr);
/* 释放对象资源  --可用对象有 连接， 结果集 ，显式参数结构指针 ，大对象
* *Ptr_obj 对象指针地址传入
*/
void XG_API XGC_Drop(void**Ptr_obj);
int  XG_API dt2dtm_Api(long long  t,char * p_dt);
int  XG_API Release_IpsAttrs(void** pconn_IpsAttr);//ips= 
int  XG_API fun_sql_type(char* sql);
/* 重置对象资源 -包括连接中显式参数结构和大对象
* 不包括结果集
* Ptr_obj 对象指针地址
*/
int  XG_API XGC_Reset(void**Ptr_obj); 
/*/
获取结果集类型，并根据结果集的类型 type 不同: 返回 结果集的行，列数，  update delete 影响的行数 ，insert 返回的 rowid值
// insert_rowid 为 char(24)的字符串
* 返回值  成功返回0 ，参数错误 返回-3；
*/
int  XG_API  XGC_getResultRet(void**  pTr_Result,int * type, 
       int* field_num,int * rowcount, int *effected_num ,char* insert_rowid);

/* Get the rowid of the last insert operation */
int  XG_API XGC_GetLastInsertId(void** p_conn, char* insert_rowid);
int  XG_API XGC_GetFunReturnType(void** p_conn, int * type);
#ifdef __cplusplus

}
#endif 

#endif
