package main

import (
	"fmt"
	"korsakov2ea/xfunc"
	"log"
	"net/http"
	"time"
)

var configFile string = "config.ini"
var QB xfunc.TDatabase           // QB - Query base - база с запросами
var QBQuery xfunc.TTable         // Представление таблицы запросов
var QBConnection xfunc.TTable    // Представление таблицы соединений
var QBParam xfunc.TTable         // Представление таблицы параметров запросов
var RenderData xfunc.TRenderData // Данные для передачи в предсталение (рендера)
var sqlErr error = nil           // Ошибка при работе с базой

func main() {

	defer mainDefer()

	// Общие настройки
	logFile := xfunc.GetExecFilePath() + xfunc.GetIniValue(configFile, "Common", "LogFile")
	serverPort := xfunc.GetIniValue(configFile, "Common", "Port")

	// Включение логгирования
	xfunc.StartLogging(logFile)

	// Соединение с базой запросов
	xfunc.DBGetIniCfg(xfunc.GetExecFilePath()+"\\"+configFile, "QB", &QB)
	QB.DBOpen()

	// Связываение объектов таблиц с реальными таблицами БД
	QBQuery.Bind("QUERY", &QB)
	QBConnection.Bind("CONNECTION", &QB)
	QBParam.Bind("PARAM", &QB)

	// Запуск сервера
	startServer(serverPort)
}

// Вызывается после закрытия main программы (для закрытия основного соединия, логгирования, обработки ошибок)
func mainDefer() {
	QB.DBClose()
	log.Println(xfunc.FuncName(), "Завершение работы")
}

// Создание пустой базы. Удаляет таблицы CONNECTION и QUERY базы запросов, если были ранее, создает типовые с одним тестовым запросом
func createEmptyQC() {
	log.Printf("%v Создание пустой базы", xfunc.FuncName())

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
	log.Printf("%v Запуск сервера на порту %v \r\n", xfunc.FuncName(), startOnPort)
	FileServer := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", FileServer))
	http.HandleFunc("/queries", auth(queries))
	http.HandleFunc("/query", auth(query))
	http.HandleFunc("/connections", auth(connections))
	http.HandleFunc("/connection", auth(connection))
	fmt.Printf("Запуск сервера на порту %v", startOnPort)
	http.ListenAndServe(":"+startOnPort, nil)
}

// Аутентификация / авторизация по куки или паролю. Если успешно, то передает управление вложенной функции
func auth(nextFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var BIUD xfunc.TDatabase
		xfunc.DBGetIniCfg(xfunc.GetExecFilePath()+"\\"+configFile, "BIUD", &BIUD)
		BIUD.DBOpen()
		defer BIUD.DBClose()

		cookieLogin := xfunc.GetCookie(r, "QBLogin")
		cookiePass := xfunc.GetCookie(r, "QBPass")

		// если есть логин и пароль в куки
		if len(cookieLogin+cookiePass) > 1 {
			if xfunc.AuthBIUD(cookieLogin, cookiePass, true, &BIUD) == 2 {
				// аутентификация и авторизация проходит
				log.Printf("%v Аутентификация/авторизация пользователя %v по cookie", xfunc.FuncName(), cookieLogin)
				xfunc.SetCookie(w, "QBLogin", cookieLogin, 15*time.Minute)
				xfunc.SetCookie(w, "QBPass", cookiePass, 15*time.Minute)
				nextFunc(w, r)
			} else {
				// если аутентификацию не прошли, то сбросить куки и отправить на ввод пароля
				xfunc.SetCookie(w, "QBLogin", "", 0*time.Minute)
				xfunc.SetCookie(w, "QBPass", "", 0*time.Minute)
				xfunc.RenderPage(w, "login.html", "common.html", RenderData)
			}
		} else {
			// иначе (если не через куки)
			r.ParseMultipartForm(0)
			if r.Method == "POST" && r.FormValue("formType") == "login" {
				// если получаем данные с формы то пытаемся войти с этими данными
				formLogin := r.FormValue("login")
				formPass := r.FormValue("password")
				log.Printf("%v Аутентификация пользователя %v по паролю", xfunc.FuncName(), formLogin)

				authResult := xfunc.AuthBIUD(formLogin, formPass, false, &BIUD)
				switch {
				case authResult == 0:
					log.Printf("%v Аутентификация пользователя %v не выполнена", xfunc.FuncName(), formLogin)
					xfunc.RenderPage(w, "login.html", "common.html", RenderData)
				case authResult == 1:
					log.Printf("%v Авторизация пользователя %v не пройдена (нет ролей)", xfunc.FuncName(), formLogin)
					xfunc.RenderPage(w, "login.html", "common.html", RenderData)
				case authResult == 2:
					log.Printf("%v Авторизация пользователя %v пройдена", xfunc.FuncName(), formLogin)
					xfunc.SetCookie(w, "QBLogin", formLogin, 15*time.Minute)
					xfunc.SetCookie(w, "QBPass", xfunc.BiudPassHash(formPass), 15*time.Minute)
					nextFunc(w, r)
				}
			} else {
				// если данные были не с формы (а просто обратились к странице)
				xfunc.RenderPage(w, "login.html", "common.html", RenderData)
			}
		}
	}
}
