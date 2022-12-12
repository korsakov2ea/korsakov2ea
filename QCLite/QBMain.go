package main

import (
	"korsakov2ea/x_func"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
	"time"
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
	http.HandleFunc("/queries", auth(queries))
	http.HandleFunc("/query", auth(query))
	http.HandleFunc("/connections", auth(connections))
	http.HandleFunc("/connection", auth(connection))
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

// Аутентификация / авторизация по БИУД (проверка введеного пароля и ролей). Может возвращать значения (в порядке проверки):
// 0 - аутентификация не пройдена (нет пользователя или не один),
// 1 - авторизация не пройдена (нет нужной роли для пользователя),
// 2 - обе проверки пройдены
func authBIUD(login, pass string) int {
	result := 0

	var BIUD x_func.TDatabase
	x_func.DBGetIniCfg(x_func.GetExecFilePath()+"\\"+configFile, "BIUD", &BIUD)
	BIUD.DBOpen()
	defer BIUD.DBClose()

	// искать оперетора в БИУД
	sqlCode := "SELECT * FROM CS.OPERATOR WHERE LOGIN='" + login + "'"
	operator := BIUD.DBQuery(sqlCode)
	// если есть такой логин
	if len(operator) == 1 {
		buidHash := operator[0].ByName["PASS"]
		passHash := x_func.BiudPassHash(pass)
		// если совпали хэши паролей
		if buidHash == passHash {
			result++
			//искать роль пользователя
			sqlCode := "SELECT * FROM CS.OPERATORROLE WHERE LOGIN='" + login + "' AND ROLE='Администратор отделения'"
			// если есть роль
			if len(BIUD.DBQuery(sqlCode)) > 0 {
				result++
			}
		}
	}

	return result
}

// Аутентификация / авторизация по куки или паролю. Если успешно, то передает управление вложенной функции
func auth(nextFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookieLogin := x_func.GetCookie(r, "QBLogin")
		cookiePass := x_func.GetCookie(r, "QBPass")

		// если есть логин и пароль в куки
		if len(cookieLogin+cookiePass) > 1 {
			if authBIUD(cookieLogin, cookiePass) == 2 {
				// аутентификация и авторизация проходит
				log.Printf("%v Аутентификация/авторизация пользователя %v по cookie", x_func.FuncName(), cookieLogin)
				nextFunc(w, r)
			} else {
				// если аутентификация не прошли, то сбросить куки и отправить на ввод пароля
				x_func.SetCookie(w, "QBLogin", "", 1*time.Millisecond)
				x_func.SetCookie(w, "QBPass", "", 1*time.Millisecond)
				renderPage(w, "login.html", "common.html", "")
			}
		} else {
			// иначе (если не через куки)
			r.ParseMultipartForm(0)
			if r.Method == "POST" && r.FormValue("formType") == "login" {
				// если получаем данные с формы то пытаемся войти с этими данными
				formLogin := r.FormValue("login")
				formPass := r.FormValue("password")
				log.Printf("%v Аутентификация пользователя %v по паролю", x_func.FuncName(), formLogin)

				authResult := authBIUD(formLogin, formPass)
				switch {
				case authResult == 0:
					log.Printf("%v Аутентификация пользователя %v не выполнена", x_func.FuncName(), formLogin)
					renderPage(w, "login.html", "common.html", "")
				case authResult == 1:
					log.Printf("%v Авторизация пользователя %v не пройдена (нет ролей)", x_func.FuncName(), formLogin)
					renderPage(w, "login.html", "common.html", "")
				case authResult == 2:
					log.Printf("%v Авторизация пользователя %v пройдена", x_func.FuncName(), formLogin)
					x_func.SetCookie(w, "QBLogin", formLogin, 15*time.Minute)
					x_func.SetCookie(w, "QBPass", x_func.BiudPassHash(formPass), 15*time.Minute)
					nextFunc(w, r)
				}
			} else {
				// если данные были не с формы (а просто обратились к странице)
				renderPage(w, "login.html", "common.html", "")
			}
		}
	}
}
