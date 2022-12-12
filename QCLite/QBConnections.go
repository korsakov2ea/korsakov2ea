package main

import (
	"korsakov2ea/xfunc"
	"log"
	"net/http"
	"strconv"
)

// connections - обработчик HTTP (список соединений)
func connections(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v >>>>> Вызов обработчика HTTP запроса", xfunc.FuncName())
	QBConnection.ReadAll()
	xfunc.RenderPage(w, "connections.html", "common.html", QBConnection)
}

// connection - обработчик HTTP (одиночное соединение)
func connection(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v >>>>> Вызов обработчика HTTP запроса", xfunc.FuncName())

	id, err := strconv.Atoi(r.FormValue("ID"))
	if err != nil && r.Method == "POST" && (r.FormValue("submitBtn") == "Update" || r.FormValue("submitBtn") == "Delete") {
		log.Printf("%v Ошибка преобразования ID = %v из GET запроса в число", xfunc.FuncName(), r.FormValue("ID"))
	} else {

		switch {

		case r.Method == "POST" && r.FormValue("submitBtn") == "Cancel":
			QBConnection.ReadAll()
			xfunc.RenderPage(w, "connections.html", "common.html", QBConnection)

		case r.Method == "POST" && r.FormValue("submitBtn") == "Create":
			newConnection := make(map[string]string)
			newConnection["NAME"] = r.FormValue("Name")
			newConnection["DRIVER"] = r.FormValue("Driver")
			newConnection["DSN"] = r.FormValue("DSN")
			QBConnection.Create(newConnection)
			QBConnection.ReadAll()
			xfunc.RenderPage(w, "connections.html", "common.html", QBConnection)

		case r.Method == "POST" && r.FormValue("submitBtn") == "Update":
			newConnection := make(map[string]string)
			newConnection["NAME"] = r.FormValue("Name")
			newConnection["DRIVER"] = r.FormValue("Driver")
			newConnection["DSN"] = r.FormValue("DSN")
			QBConnection.Update(id, newConnection)
			QBConnection.ReadAll()
			xfunc.RenderPage(w, "connections.html", "common.html", QBConnection)

		case r.Method == "POST" && r.FormValue("submitBtn") == "Delete":
			QBConnection.Delete(id)
			QBConnection.ReadAll()
			xfunc.RenderPage(w, "connections.html", "common.html", QBConnection)

		case r.Method == "GET" && r.FormValue("mode") == "add":
			QBConnection.Data = nil
			xfunc.RenderPage(w, "connection.html", "common.html", QBConnection)

		case r.Method == "GET" && r.FormValue("mode") == "view":
			QBConnection.Read(id)
			xfunc.RenderPage(w, "connection.html", "common.html", QBConnection)

		default:

		}
	}
}

// Получение id соединения по id запроса
func GetIdConnFromQuery(id int) int {
	log.Printf("%v Получение ID соединения для запроса с ID %v", xfunc.FuncName(), id)
	connectionResult := QB.DBQuery("SELECT ID_CONNECTION FROM QUERY WHERE ID=" + strconv.Itoa(id))
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
	return idConn
}

// Выполняет SQL команду в базе, по соединению c указанным id
func executeSQLConn(sqlCode string, id int) {
	log.Printf("%v \n\tВыполнение SQL команды \n%v", xfunc.FuncName(), sqlCode)
	connectionRows := QB.DBQuery("SELECT DRIVER, DSN, NAME FROM CONNECTION WHERE ID=" + strconv.Itoa(id))
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
}
