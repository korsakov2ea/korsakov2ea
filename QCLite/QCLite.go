package main

import (
	"fmt"
	"korsakov2ea/x_func"
	"log"
	"strconv"
)

var QCDB x_func.TDatabase //Основная база программы
var QCon x_func.TDatabase //

func main() {

	//общие настройки
	defer mainDefer()
	configFile := "config.ini"

	//включение логгирования
	logFile := x_func.GetExecFilePath() + x_func.GetIniValue(configFile, "Common", "LogFile")
	x_func.StartLogging(logFile)

	//соединение с основной базой
	x_func.DBGetIniCfg(x_func.GetExecFilePath()+"\\"+configFile, "QCDB", &QCDB)
	QCDB.DBOpen()

	//тест работы с
	sqlCommand := ` 
	DROP TABLE IF EXISTS CONNECTION;`
	QCDB.DBExec(sqlCommand)

	sqlCommand = `
	CREATE TABLE IF NOT EXISTS CONNECTION (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		NAME VARCHAR(255) NOT NULL,
		DRIVER VARCHAR(255) NOT NULL,
		DSN  VARCHAR(255) NOT NULL);`
	QCDB.DBExec(sqlCommand)

	sqlCommand = `
	INSERT INTO CONNECTION (NAME, DRIVER, DSN) VALUES ('QC', 'sqlite3', '` + QCDB.DSN + `');`
	QCDB.DBExec(sqlCommand)

	sqlCommand = `
	DROP TABLE IF EXISTS QUERY;`
	QCDB.DBExec(sqlCommand)

	sqlCommand = `
	CREATE TABLE IF NOT EXISTS QUERY (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		ID_CONNECTION INTEGER,
		NAME VARCHAR(255) NOT NULL UNIQUE,
		QUERY TEXT,
		REM TEXT);`
	QCDB.DBExec(sqlCommand)

	sqlCommand = `
	INSERT INTO QUERY (ID_CONNECTION, NAME, QUERY, REM) VALUES (1, 'TEST', 'SELECT * FROM QUERY', 'Тестовый запрос');`
	QCDB.DBExec(sqlCommand)

	fmt.Println(ExecQuery(1))

}

func mainDefer() {
	QCDB.DBClose()
	log.Println(x_func.FuncName(), "Завершение работы")
}

func ExecQuery(id int) (SliceMap []map[string]string, RowCount int) {
	var tmpConn x_func.TDatabase //соединение для выполнения запроса из базы
	qcText, qcRowCount := QCDB.DBQuery("SELECT * FROM QUERY AS Q INNER JOIN CONNECTION AS C ON Q.ID_CONNECTION=C.ID AND Q.ID="+strconv.Itoa(id), true)
	if qcRowCount != 0 {
		tmpConn.Driver = qcText[0]["DRIVER"]
		tmpConn.DSN = qcText[0]["DSN"]
		tmpConn.DBOpen()
		defer tmpConn.DBClose()
		queryResult, queryResultRowCount := tmpConn.DBQuery(qcText[0]["QUERY"], false)
		return queryResult, queryResultRowCount
	} else {
		log.Println(x_func.FuncName(), "Нет строки запроса с ID -", id)
		return nil, 0
	}

}
