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

	//тест работы с бд
	QCDB.DBExec("INSERT INTO QC.QUERY (QUERY, REM, ID_CONNECTION) VALUES ('SELECT * FROM KL.PW','TEST QUERY', 1)")
	testResult, testRowCount := QCDB.DBQuery("SELECT * FROM QC.QUERY AS Q INNER JOIN QC.CONNECTION AS C ON C.ID=Q.ID_CONNECTION WHERE Q.ID=1", true)
	fmt.Println(testResult, testRowCount)

	fmt.Println(ExecQuery(1))

}

func mainDefer() {
	QCDB.DBClose()
	log.Println(x_func.FuncName(), "Завершение работы")
}

func ExecQuery(id int) (SliceMap []map[string]string, RowCount int) {
	var tmpConn x_func.TDatabase //соединение для выполнения запроса из базы
	qcText, qcRowCount := QCDB.DBQuery("SELECT * FROM QC.QUERY AS Q INNER JOIN QC.CONNECTION AS C ON C.ID=Q.ID_CONNECTION WHERE Q.ID="+strconv.Itoa(id), true)
	if qcRowCount != 0 {
		tmpConn.Driver = qcText[0]["DRIVER"]
		tmpConn.DSN = "HOSTNAME=" + qcText[0]["HOSTNAME"] + "; DATABASE=" + qcText[0]["DATABASE"] + "; PORT=" + qcText[0]["PORT"] + "; UID=" + qcText[0]["UID"] + "; PWD=" + qcText[0]["PWD"]
		tmpConn.DBOpen()
		defer tmpConn.DBClose()
		queryResult, queryResultRowCount := tmpConn.DBQuery(qcText[0]["QUERY"], true)
		return queryResult, queryResultRowCount
	} else {
		log.Println(x_func.FuncName(), "Нет строки запроса с ID -", id)
		return nil, 0
	}

}
