package main

import (
	"korsakov2ea/x_func"
	"log"
	"net/http"
	"strconv"
)

// connections - обработчик HTTP (список соединений)
func connections(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v Обработка НTTP запроса", x_func.FuncName())

	QBConnection.ReadAll()
	renderPage(w, "connections.html", "common.html", QBConnection.data)
}

// connection - обработчик HTTP (одиночное соединение)
func connection(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v Обработка НTTP запроса", x_func.FuncName())

	id, err := strconv.Atoi(r.FormValue("ID"))
	if err != nil && r.Method == "POST" && (r.FormValue("submitBtn") == "Update" || r.FormValue("submitBtn") == "Delete") {
		log.Printf("%v Ошибка преобразования ID = %v из GET запроса в число", x_func.FuncName(), r.FormValue("ID"))
	} else {

		switch {

		case r.Method == "POST" && r.FormValue("submitBtn") == "Cancel":
			QBConnection.ReadAll()
			renderPage(w, "connections.html", "common.html", QBConnection.data)

		case r.Method == "POST" && r.FormValue("submitBtn") == "Create":
			name := r.FormValue("Name")
			driver := r.FormValue("Driver")
			dsn := r.FormValue("DSN")
			newConnection := make(map[string]string)
			newConnection["NAME"] = name
			newConnection["DRIVER"] = driver
			newConnection["DSN"] = dsn
			QBConnection.Create(newConnection)
			QBConnection.ReadAll()
			renderPage(w, "connections.html", "common.html", QBConnection.data)

		case r.Method == "POST" && r.FormValue("submitBtn") == "Update":
			name := r.FormValue("Name")
			driver := r.FormValue("Driver")
			dsn := r.FormValue("DSN")
			newConnection := make(map[string]string)
			newConnection["NAME"] = name
			newConnection["DRIVER"] = driver
			newConnection["DSN"] = dsn
			QBConnection.Update(id, newConnection)
			QBConnection.ReadAll()
			renderPage(w, "connections.html", "common.html", QBConnection.data)

		case r.Method == "POST" && r.FormValue("submitBtn") == "Delete":
			QBConnection.Delete(id)
			QBConnection.ReadAll()
			renderPage(w, "connections.html", "common.html", QBConnection.data)

		case r.Method == "GET" && r.FormValue("mode") == "add":
			QBConnection.data = nil
			renderPage(w, "connection.html", "common.html", QBConnection.data)

		case r.Method == "GET" && r.FormValue("mode") == "view":
			QBConnection.Read(id)
			renderPage(w, "connection.html", "common.html", QBConnection.data)

		default:

		}
	}
}
