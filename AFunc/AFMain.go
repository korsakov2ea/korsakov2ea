package main

import (
	"korsakov2ea/xfunc"
	"log"
	"net/http"
	"time"
)

var configFile string = "config.ini"
var R xfunc.TDatabase      // R - REESTR
var EmpCenter xfunc.TTable //центр занятости

func main() {

	defer mainDefer()

	// Общие настройки
	logFile := xfunc.GetExecFilePath() + xfunc.GetIniValue(configFile, "Common", "LogFile")
	serverPort := xfunc.GetIniValue(configFile, "Common", "Port")

	// Включение логгирования
	xfunc.StartLogging(logFile)

	// Соединение с базой запросов
	xfunc.DBGetIniCfg(xfunc.GetExecFilePath()+"\\"+configFile, "REESTR", &R)
	R.DBOpen()

	// Связываение объектов таблиц с реальными таблицами БД
	EmpCenter.Bind("ADDFUNC.EMPLOYMENT_CENTER", &R)

	// Запуск сервера
	startServer(serverPort)
}

// Вызывается после закрытия main программы (для закрытия основного соединия, логгирования, обработки ошибок)
func mainDefer() {
	R.DBClose()
	log.Println(xfunc.FuncName(), "Завершение работы")
}

// Определяет парамерты сервера, роутинг и обработчики, запускает сервер на указанном порту
func startServer(startOnPort string) {
	log.Printf("%v Запуск сервера на порту %v", xfunc.FuncName(), startOnPort)
	FileServer := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", FileServer))
	http.HandleFunc("/empcenter", auth(empCenter))
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
				nextFunc(w, r)
			} else {
				// если аутентификацию не прошли, то сбросить куки и отправить на ввод пароля
				xfunc.SetCookie(w, "QBLogin", "", 0*time.Minute)
				xfunc.SetCookie(w, "QBPass", "", 0*time.Minute)
				xfunc.RenderPage(w, "login.html", "common.html", "")
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
					xfunc.RenderPage(w, "login.html", "common.html", "")
				case authResult == 1:
					log.Printf("%v Авторизация пользователя %v не пройдена (нет ролей)", xfunc.FuncName(), formLogin)
					xfunc.RenderPage(w, "login.html", "common.html", "")
				case authResult == 2:
					log.Printf("%v Авторизация пользователя %v пройдена", xfunc.FuncName(), formLogin)
					xfunc.SetCookie(w, "QBLogin", formLogin, 15*time.Minute)
					xfunc.SetCookie(w, "QBPass", xfunc.BiudPassHash(formPass), 15*time.Minute)
					nextFunc(w, r)
				}
			} else {
				// если данные были не с формы (а просто обратились к странице)
				xfunc.RenderPage(w, "login.html", "common.html", "")
			}
		}
	}
}
