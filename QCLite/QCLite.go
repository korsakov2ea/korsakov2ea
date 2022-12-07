package main

import (
	"korsakov2ea/x_func"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
)

var QB x_func.TDatabase        // QB - Query base - база с запросами
var QBQuery x_func.TTable      // Представление таблицы запросов
var QBConnection x_func.TTable // Представление таблицы соединений
var configFile string = "config.ini"

func main() {

	defer mainDefer()

	// Общие настройки
	logFile := x_func.GetExecFilePath() + x_func.GetIniValue(configFile, "Common", "LogFile")
	serverPort := x_func.GetIniValue(configFile, "Common", "Port")

	// Включение логгирования
	x_func.StartLogging(logFile)

	// Соединение с базой запросов
	x_func.DBGetIniCfg(x_func.GetExecFilePath()+"\\"+configFile, "QB", &QB)
	QB.DBOpen()

	// Связываение объектов таблиц с реальными таблицами БД
	QBQuery.BindTable("QUERY", &QB)
	QBConnection.BindTable("CONNECTION", &QB)

	// Запуск сервера
	startServer(serverPort)
}

// Вызывается после закрытия main программы (для закрытия основного соединия, логгирования, обработки ошибок)
func mainDefer() {
	QB.DBClose()
	log.Println(x_func.FuncName(), "Завершение работы")
}

// Создание пустой базы. Удаляет таблицы CONNECTION и QUERY базы запросов, если были ранее, создает типовые с одним тестовым запросом
func createEmptyQC() {
	log.Printf("%v Создание пустой базы", x_func.FuncName())

	sqlCommand := ` 
	DROP TABLE IF EXISTS CONNECTION;`
	QB.DBExec(sqlCommand)

	sqlCommand = `
	CREATE TABLE IF NOT EXISTS CONNECTION (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		NAME VARCHAR(255) NOT NULL,
		DRIVER VARCHAR(255) NOT NULL,
		DSN  VARCHAR(255) NOT NULL);`
	QB.DBExec(sqlCommand)

	sqlCommand = `
	INSERT INTO CONNECTION (NAME, DRIVER, DSN) VALUES ('QC', 'sqlite3', '` + QB.DSN + `');`
	QB.DBExec(sqlCommand)

	sqlCommand = `
	DROP TABLE IF EXISTS QUERY;`
	QB.DBExec(sqlCommand)

	sqlCommand = `
	CREATE TABLE IF NOT EXISTS QUERY (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		ID_CONNECTION INTEGER,
		NAME VARCHAR(255) NOT NULL UNIQUE,
		QUERY TEXT,
		REM TEXT);`
	QB.DBExec(sqlCommand)

	sqlCommand = `
	INSERT INTO QUERY (ID_CONNECTION, NAME, QUERY, REM) VALUES (1, 'TEST', 'SELECT * FROM QUERY', 'Тестовый запрос');`
	QB.DBExec(sqlCommand)
}

// Определяет парамерты сервера, роутинг и обработчики, запускает сервер на указанном порту
func startServer(startOnPort string) {
	log.Printf("%v Запуск сервера на порту %v", x_func.FuncName(), startOnPort)
	FileServer := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", FileServer))
	http.HandleFunc("/queries", queries)
	http.HandleFunc("/query", query)
	http.HandleFunc("/connections", connections)
	http.HandleFunc("/connection", connection)
	http.ListenAndServe(":"+startOnPort, nil)
}

// Собирает страницу из макета страницы (templatePage), подстановочной страницы (commonPage), объекта данных (date) и выводит в поток (w)
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

// Проверка введеного пароля по БИУД
func checkBIUD(login, pass string) bool {
	var BIUD x_func.TDatabase
	x_func.DBGetIniCfg(x_func.GetExecFilePath()+"\\"+configFile, "BIUD", &BIUD)
	BIUD.DBOpen()
	defer BIUD.DBClose()

	var BIUDOperator x_func.TTable
	BIUDOperator.BindTable("OPERATOR", &BIUD)

	sqlCode := "SELECT * FROM CS.OPERATOR WHERE LOGIN='" + login + "'"
	BIUDOperator.ReadSQL(sqlCode)

	buidHash := BIUDOperator.Data[0].ByName["PASS"]
	passHash := x_func.BiudPassHash(pass)

	result := false
	if buidHash == passHash {
		result = true
	}
	return result
}
