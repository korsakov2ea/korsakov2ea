package main

import (
	"fmt"
	"korsakov2ea/x_func"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"text/template"
)

var QCDB x_func.TDatabase //Основная база программы
var QCon x_func.TDatabase //
var QBConnection QBEntity
var QBQuery QBEntity

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

	//сопоставление QBEntity с таблицами
	QBConnection.name = "CONNECTION"
	QBQuery.name = "QUERY"

	/*
		//создание пустой базы
		createEmptyQC()

		//создание тестового запроса
		connectionNew := make(map[string]string)
		connectionNew["DRIVER"] = "sqlite3"
		connectionNew["DSN"] = "SQLite\\base\\testbase.db"
		connectionNew["NAME"] = "QB"
		QBConnection.Create(connectionNew)
		QBConnection.Read(4)
		QBConnection.Update(4, connectionNew)
		QBConnection.Delete(5)

		//выполнение тестового запроса
		fmt.Println(execQuery(2))
	*/

	startServer()

}

// mainDefer - вызывается после закрытия main программы (для закрытия основного соединия, логгирования, обработки ошибок)
func mainDefer() {
	QCDB.DBClose()
	log.Println(x_func.FuncName(), "Завершение работы")
}

// createEmptyQC - удаляет таблицы CONNECTION и QUERY основной базы QC, если были ранее, создает типовые с одним тестовым запросом
func createEmptyQC() {
	log.Printf("%v Создание пустой базы", x_func.FuncName())

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
}

// execQuery - выполняет запрос из базы QC под номером id и возвращает массив карт с результатами выборки, а также кол-во срок в результате
func execQuery(id int) (SliceMap []map[string]string, RowCount int) {
	log.Printf("%v Выполнение запроса", x_func.FuncName())
	var tmpConn x_func.TDatabase //соединение для выполнения запроса из базы
	qcStringMap, qcRowCount := QCDB.DBQuery("SELECT * FROM QUERY AS Q INNER JOIN CONNECTION AS C ON Q.ID_CONNECTION=C.ID AND Q.ID="+strconv.Itoa(id), false)
	if qcRowCount != 0 {
		tmpConn.Driver = qcStringMap[0]["DRIVER"]
		tmpConn.DSN = qcStringMap[0]["DSN"]
		tmpConn.Name = qcStringMap[0]["NAME"]
		decodeParam := false
		if tmpConn.Driver == "go_ibm_db" {
			decodeParam = true
		}
		tmpConn.DBOpen()
		defer tmpConn.DBClose()
		queryResult, queryResultRowCount := tmpConn.DBQuery(qcStringMap[0]["QUERY"], decodeParam)
		return queryResult, queryResultRowCount
	} else {
		log.Println(x_func.FuncName(), "Нет строки запроса с ID -", id)
		return nil, -1
	}

}

func startServer() {
	fmt.Println("Старт сервера")
	FileServer := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", FileServer))
	http.HandleFunc("/queries", queries)
	http.HandleFunc("/query", query)
	http.HandleFunc("/connections", connections)
	http.HandleFunc("/connection", connection)
	http.ListenAndServe(":4444", nil)
}

// renderPage - Собирает страницу из макета страницы (templatePage), подстановочной страницы (commonPage), объекта данных (date) и выводит в поток (w)
func renderPage(w http.ResponseWriter, templatePage string, commonPage string, date interface{}) {
	templatePath := filepath.Join("public", "html", templatePage)
	commonPath := filepath.Join("public", "html", commonPage)

	tmpl, err := template.ParseFiles(templatePath, commonPath)
	if err != nil {
		log.Println(x_func.FuncName(), "Ошибка парсинга шаблона страницы", err)
		http.Error(w, err.Error(), 400)
	} else {
		log.Println(x_func.FuncName(), "Парсинг шаблона страницы", templatePath)
	}

	err = tmpl.Execute(w, date)
	if err != nil {
		log.Println(x_func.FuncName(), "Ошибка выполнения шаблона страницы", err)
		http.Error(w, err.Error(), 400)
	} else {
		log.Println(x_func.FuncName(), "Построение шаблона страницы", templatePath)
	}
}
