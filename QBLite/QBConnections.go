package main

import (
	"korsakov2ea/xfunc"
	"log"
	"net/http"
	"strconv"
)

// connections - обработчик HTTP (список соединений)
func connections(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v --------- HTTP METHOD: %v, PARAM: %v", xfunc.FuncName(), r.Method, r.URL.RawQuery)
	RenderData.GetAllFromTable(&QBConnection, 0)
	xfunc.RenderPage(w, "connections.html", "common.html", RenderData)
	RenderData.Clear()
}

// connection - обработчик HTTP (одиночное соединение)
func connection(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v --------- HTTP METHOD: %v, PARAM: %v", xfunc.FuncName(), r.Method, r.URL.RawQuery)
	id, err := strconv.Atoi(r.FormValue("ID"))
	if err != nil && r.Method == "POST" && (r.FormValue("submitBtn") == "Update" || r.FormValue("submitBtn") == "Delete") {
		log.Printf("%v Ошибка преобразования ID = %v из GET запроса в число", xfunc.FuncName(), r.FormValue("ID"))
	} else {
		RenderData.Clear()
		switch {

		// Отмена изменений соединения
		case r.Method == "POST" && r.FormValue("submitBtn") == "Cancel":
			RenderData.AddAlert("Нажата кнопка [Отмена]", "info")
			http.Redirect(w, r, "connections", http.StatusFound)

		// Создание / изменение соединения
		case r.Method == "POST" && (r.FormValue("submitBtn") == "Create" || r.FormValue("submitBtn") == "Update"):
			newConnection := make(map[string]string)
			newConnection["NAME"] = r.FormValue("Name")
			newConnection["DRIVER"] = r.FormValue("Driver")
			newConnection["DSN"] = r.FormValue("DSN")
			action := ""
			switch r.FormValue("submitBtn") {
			case "Create":
				sqlErr = QBConnection.Create(newConnection)
				action = "создано"
			case "Update":
				sqlErr = QBConnection.Update(id, newConnection)
				action = "изменено"
			}
			if sqlErr != nil {
				RenderData.AddAlert(sqlErr.Error(), "danger")
			} else {
				RenderData.AddAlert("Соединение "+action, "success")
			}
			http.Redirect(w, r, "connections", http.StatusFound)

		// Удаление соединения
		case r.Method == "POST" && r.FormValue("submitBtn") == "Delete":
			sqlErr = QBConnection.Delete(id)
			if sqlErr != nil {
				RenderData.AddAlert(sqlErr.Error(), "danger")
			} else {
				RenderData.AddAlert("Соединение удалено", "success")
			}
			http.Redirect(w, r, "connections", http.StatusFound)

		// Переход к добавления соединения
		case r.Method == "GET" && r.FormValue("mode") == "add":
			xfunc.RenderPage(w, "connection.html", "common.html", RenderData)

		// Переход к изменению соединения
		case r.Method == "GET" && r.FormValue("mode") == "edit":
			RenderData.GetIdFromIdTable(&QBConnection, id)
			xfunc.RenderPage(w, "connection.html", "common.html", RenderData)

		default:
			w.Write([]byte("Неожиданный запрос!"))
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
