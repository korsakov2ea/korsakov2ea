package main

import (
	"korsakov2ea/x_func"
	"log"
	"net/http"
	"strconv"
)

// queries - обработчик HTTP (список запросов)
func queries(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v Обработка НTTP запроса", x_func.FuncName())
	QBQuery.ReadSQL("SELECT Q.ID, Q.REM, Q.QUERY, C.NAME FROM QUERY AS Q INNER JOIN CONNECTION AS C ON Q.ID_CONNECTION=C.ID")
	renderPage(w, "queries.html", "common.html", QBQuery.Data)
}

// query - обработчик HTTP (одиночный запрос)
func query(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v Обработка НTTP запроса", x_func.FuncName())

	id, err := strconv.Atoi(r.FormValue("ID"))
	if err != nil && r.Method == "POST" && (r.FormValue("submitBtn") == "Update" || r.FormValue("submitBtn") == "Delete") {
		log.Printf("%v Ошибка преобразования ID = %v из GET запроса в число", x_func.FuncName(), r.FormValue("ID"))
	} else {

		switch {

		case r.Method == "POST" && r.FormValue("submitBtn") == "Cancel":
			QBQuery.ReadAll()
			renderPage(w, "queries.html", "common.html", QBQuery.Data)

		case r.Method == "POST" && r.FormValue("submitBtn") == "Create":
			name := r.FormValue("Name")
			driver := r.FormValue("Driver")
			dsn := r.FormValue("DSN")
			newConnection := make(map[string]string)
			newConnection["NAME"] = name
			newConnection["DRIVER"] = driver
			newConnection["DSN"] = dsn
			QBQuery.Create(newConnection)
			QBQuery.ReadAll()
			renderPage(w, "queries.html", "common.html", QBQuery.Data)

		case r.Method == "POST" && r.FormValue("submitBtn") == "Update":
			name := r.FormValue("Name")
			driver := r.FormValue("Driver")
			dsn := r.FormValue("DSN")
			newConnection := make(map[string]string)
			newConnection["NAME"] = name
			newConnection["DRIVER"] = driver
			newConnection["DSN"] = dsn
			QBQuery.Update(id, newConnection)
			QBQuery.ReadAll()
			renderPage(w, "queries.html", "common.html", QBQuery.Data)

		case r.Method == "POST" && r.FormValue("submitBtn") == "Delete":
			QBQuery.Delete(id)
			QBQuery.ReadAll()
			renderPage(w, "queries.html", "common.html", QBQuery.Data)

		case r.Method == "GET" && r.FormValue("mode") == "add":
			QBQuery.Data = nil
			renderPage(w, "query.html", "common.html", QBQuery.Data)

		case r.Method == "GET" && r.FormValue("mode") == "view":
			QBQuery.Read(id)
			QBConnection.ReadAll()
			QBQuery.Extra = nil
			QBQuery.Extra = append(QBQuery.Extra, QBConnection.Data)
			renderPage(w, "query.html", "common.html", QBQuery)

		default:

		}
	}
}
