package main

import (
	"korsakov2ea/x_func"
	"log"
)

var QCDB x_func.TDatabase

func main() {

	defer mainDefer()
	configFile := "config.ini"

	logFile := x_func.GetExecFilePath() + x_func.GetIniValue(configFile, "Common", "LogFile")
	x_func.StartLogging(logFile)

	x_func.DBGetIniCfg(configFile, "QCDB", &QCDB)
	QCDB.DB = x_func.DBOpen(QCDB.Driver, QCDB.DSN)

}

func mainDefer() {
	QCDB.DB.Close()
	log.Println(x_func.FuncName(), "<<<<<< Завершение работы <<<<<<")
}
