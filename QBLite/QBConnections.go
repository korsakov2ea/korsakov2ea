package main

import (
	"korsakov2ea/xfunc"
	"log"
	"net/http"
	"strconv"
)

// connections - обработчик HTTP (список соединений)
func connections(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v Переход к списку соединений ────────────────────────────────────────┐", xfunc.FuncName())
	defer log.Printf("%v Переход к списку соединений ────────────────────────────────────────┘", xfunc.FuncName())

	sqlErr = QBConnection.ReadAll(0)
	if sqlErr != nil {
		RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
	}
	RenderData.Data = QBConnection.Data
	RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Всего соединений - " + strconv.Itoa(len(RenderData.Data)), Class: "info"})
	xfunc.RenderPage(w, "connections.html", "common.html", RenderData)
	RenderData.Alerts = nil
}

// connection - обработчик HTTP (одиночное соединение)
func connection(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("ID"))
	if err != nil && r.Method == "POST" && (r.FormValue("submitBtn") == "Update" || r.FormValue("submitBtn") == "Delete") {
		log.Printf("%v Ошибка преобразования ID = %v из GET запроса в число", xfunc.FuncName(), r.FormValue("ID"))
	} else {
		log.Printf("%v HTTP запрос с параметрами %v", xfunc.FuncName(), r.URL.RawQuery)

		switch {

		// Отмена изменений соединения
		case r.Method == "POST" && r.FormValue("submitBtn") == "Cancel":
			log.Printf("%v Отмена изменений соединения ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil
			RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Нажата кнопка [Отмена]", Class: "info"})
			http.Redirect(w, r, "connections", http.StatusFound)
			log.Printf("%v Отмена изменений соединения ────────────────────────────────────────┘", xfunc.FuncName())

		// Создание соединения
		case r.Method == "POST" && r.FormValue("submitBtn") == "Create":
			log.Printf("%v Создание соединения ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil
			newConnection := make(map[string]string)
			newConnection["NAME"] = r.FormValue("Name")
			newConnection["DRIVER"] = r.FormValue("Driver")
			newConnection["DSN"] = r.FormValue("DSN")

			sqlErr = QBConnection.Create(newConnection)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}

			sqlErr = QBConnection.ReadAll(0)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}

			if sqlErr == nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Соединение создано", Class: "success"})
			}

			RenderData.Data = QBConnection.Data
			http.Redirect(w, r, "connections", http.StatusFound)
			log.Printf("%v Создание соединения ────────────────────────────────────────┘", xfunc.FuncName())

		// Изменение соединения
		case r.Method == "POST" && r.FormValue("submitBtn") == "Update":
			log.Printf("%v Изменение соединения ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil
			newConnection := make(map[string]string)
			newConnection["NAME"] = r.FormValue("Name")
			newConnection["DRIVER"] = r.FormValue("Driver")
			newConnection["DSN"] = r.FormValue("DSN")

			sqlErr = QBConnection.Update(id, newConnection)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}

			sqlErr = QBConnection.ReadAll(0)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}
			RenderData.Data = QBConnection.Data

			if sqlErr == nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Соединение изменено", Class: "success"})
			}

			http.Redirect(w, r, "connections", http.StatusFound)
			log.Printf("%v Изменение соединения ────────────────────────────────────────┘", xfunc.FuncName())

		// Удаление соединения
		case r.Method == "POST" && r.FormValue("submitBtn") == "Delete":
			log.Printf("%v Удаление соединения ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil

			sqlErr = QBConnection.Delete(id)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}

			sqlErr = QBConnection.ReadAll(0)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}
			RenderData.Data = QBConnection.Data

			if sqlErr == nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Соединение удалено", Class: "success"})
			}

			http.Redirect(w, r, "connections", http.StatusFound)
			log.Printf("%v Удаление соединения ────────────────────────────────────────┘", xfunc.FuncName())

		// Переход к добавления соединения
		case r.Method == "GET" && r.FormValue("mode") == "add":
			log.Printf("%v Переход к добавления соединения ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil
			RenderData.Data = nil
			xfunc.RenderPage(w, "connection.html", "common.html", RenderData)
			log.Printf("%v Переход к добавления соединения ────────────────────────────────────────┘", xfunc.FuncName())

		// Переход к изменению соединения
		case r.Method == "GET" && r.FormValue("mode") == "edit":
			log.Printf("%v Переход к изменению соединения ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil

			sqlErr = QBConnection.Read(id)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}
			RenderData.Data = QBConnection.Data

			xfunc.RenderPage(w, "connection.html", "common.html", RenderData)
			log.Printf("%v Переход к изменению соединения ────────────────────────────────────────┘", xfunc.FuncName())

		default:
			// сделать заглушку если неизветный путь
		}
	}
}

// Получение id соединения по id запроса
func GetIdConnFromQuery(id int) (int, error) {
	log.Printf("%v Получение ID соединения для запроса с ID %v", xfunc.FuncName(), id)
	connectionResult, err := QB.DBQuery("SELECT ID_CONNECTION FROM QUERY WHERE ID=" + strconv.Itoa(id))
	connectionResultCount := len(connectionResult)
	idConn := -1
	if connectionResultCount != 0 {
		id, err := strconv.Atoi(connectionResult[0].ByName["ID_CONNECTION"])
		if err != nil {
			log.Println(xfunc.FuncName(), "Ошибка преобразования qcStringMap[0][\"ID_CONNECTION\"] = %v в число", connectionResult[0].ByName["ID_CONNECTION"])
		} else {
			idConn = id
		}
	} else {
		log.Println(xfunc.FuncName(), "Нет строки запроса с ID -", id)
	}
	return idConn, err
}

// Выполняет SQL команду в базе, по соединению c указанным id
func executeSQLConn(sqlCode string, id int) error {
	shortSqlCode := sqlCode
	if len(shortSqlCode) > 1000 {
		shortSqlCode = shortSqlCode[:1000] + "... "
	}
	log.Printf("%v \n\tВыполнение SQL команды \n%v", xfunc.FuncName(), shortSqlCode)
	connectionRows, err := QB.DBQuery("SELECT DRIVER, DSN, NAME FROM CONNECTION WHERE ID=" + strconv.Itoa(id))
	connectionRowsCount := len(connectionRows)

	var targetConn xfunc.TDatabase //соединение для выполнения запроса из базы

	if connectionRowsCount != 0 {
		targetConn.Driver = connectionRows[0].ByName["DRIVER"]
		targetConn.DSN = connectionRows[0].ByName["DSN"]
		targetConn.SetDecodeParam()

		targetConn.DBOpen()
		defer targetConn.DBClose()

		targetConn.DBExec(sqlCode)

	} else {
		log.Printf("%v \n\tНет соединения с ID = %v", xfunc.FuncName(), id)
	}
	return err
}
