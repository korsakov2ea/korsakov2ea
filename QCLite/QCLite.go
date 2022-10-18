package main

import (
	"korsakov2ea/x_func"
	"log"
)

var QCDB x_func.TDatabase //Основная база программы

func main() {

	//общие настройки
	defer mainDefer()
	configFile := "config.ini"

	//включение логгирования
	logFile := x_func.GetExecFilePath() + x_func.GetIniValue(configFile, "Common", "LogFile")
	x_func.StartLogging(logFile)

	//соединение с основной базой
	x_func.DBGetIniCfg(configFile, "QCDB", &QCDB)
	QCDB.DBOpen()
	QCDB.DBExec("INSERT INTO QC.QUERY (QUERY, REM) VALUES ('SELECT * FROM KL.PW','TEST QUERY')")

}

func mainDefer() {
	QCDB.DBClose()
	log.Println(x_func.FuncName(), "Завершение работы")
}
