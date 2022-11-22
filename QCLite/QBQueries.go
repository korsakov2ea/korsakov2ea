package main

import (
	"korsakov2ea/x_func"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// queries - обработчик HTTP (список запросов)
func queries(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v Обработка НTTP запроса", x_func.FuncName())
	QBQuery.ReadSQL("SELECT Q.ID, Q.REM, Q.QUERY, C.NAME FROM QUERY AS Q INNER JOIN CONNECTION AS C ON Q.ID_CONNECTION=C.ID")
	renderPage(w, "queries.html", "common.html", QBQuery)
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
			renderPage(w, "queries.html", "common.html", QBQuery)

		case r.Method == "POST" && r.FormValue("submitBtn") == "Create":
			newQuery := make(map[string]string)
			newQuery["NAME"] = r.FormValue("Name")
			newQuery["QUERY"] = strings.Replace(r.FormValue("Query"), "'", "''", -1)
			newQuery["REM"] = strings.Replace(r.FormValue("Rem"), "'", "''", -1)
			newQuery["ID_CONNECTION"] = r.FormValue("Id_connection")
			QBQuery.Create(newQuery)
			QBQuery.ReadAll()
			renderPage(w, "queries.html", "common.html", QBQuery)

		case r.Method == "POST" && r.FormValue("submitBtn") == "Update":
			newQuery := make(map[string]string)
			newQuery["NAME"] = r.FormValue("Name")
			newQuery["QUERY"] = strings.Replace(r.FormValue("Query"), "'", "''", -1)
			newQuery["REM"] = strings.Replace(r.FormValue("Rem"), "'", "''", -1)
			newQuery["ID_CONNECTION"] = r.FormValue("Id_connection")
			QBQuery.Update(id, newQuery)
			QBQuery.ReadAll()
			renderPage(w, "queries.html", "common.html", QBQuery)

		case r.Method == "POST" && r.FormValue("submitBtn") == "Delete":
			QBQuery.Delete(id)
			QBQuery.ReadAll()
			renderPage(w, "queries.html", "common.html", QBQuery)

		case r.Method == "GET" && r.FormValue("mode") == "add":
			QBQuery.Data = nil
			QBConnection.ReadAll()
			QBQuery.Directory = nil
			QBQuery.Directory = append(QBQuery.Directory, QBConnection.Data)
			renderPage(w, "query.html", "common.html", QBQuery)

		case r.Method == "GET" && r.FormValue("mode") == "view":
			QBQuery.Read(id)
			QBConnection.ReadAll()
			QBQuery.Directory = nil
			QBQuery.Directory = append(QBQuery.Directory, QBConnection.Data)
			renderPage(w, "query.html", "common.html", QBQuery)

		case r.Method == "GET" && r.FormValue("mode") == "prepare":
			if needUpload(id) && QBQuery.NeedUpload == false {
				QBQuery.NeedUpload = true
				renderPage(w, "uploadfile.html", "common.html", QBQuery)
			} else {
				QBQuery.Data, _ = execQuery(id)
				renderPage(w, "result.html", "common.html", QBQuery)
			}

		case r.Method == "POST" && r.FormValue("submitBtn") == "Upload":
			CSVData := x_func.GetStrMapFromCSVWebFile(x_func.UploadFile(r, "uploadFile"))
			CSVData = x_func.DecodeStrMap1251toUTF8(CSVData)
			_, _, colNames := x_func.GetSizeStrMap(CSVData)
			colNameTypes := strings.Replace(colNames, ",", " VARCHAR(255),", -1) + " VARCHAR(255)"
			idConn := GetIdConnFromIdQuery(id)
			tableName := "QB" + x_func.GenerateTimeStamp()
			sqlCode := "CREATE TABLE QB." + tableName + " (" + colNameTypes + ")"
			ExecSQL(sqlCode, idConn)

			for _, row := range CSVData {
				sqlCode = ""
				for _, cell := range row {
					sqlCode = sqlCode + ", '" + cell + "'"
				}
				sqlCode := sqlCode[2:]
				sqlCode = "INSERT INTO QB." + tableName + " (" + colNames + ") VALUES (" + x_func.DecodeStrUTF8to1251(sqlCode) + ")"
				ExecSQL(sqlCode, idConn)
			}

			/*
				QBQuery.Data, _ = execQuery(id)
				renderPage(w, "result.html", "common.html", QBQuery)
			*/
		default:
		}
	}
}

func needUpload(id int) bool {
	_, rowsCount := execQuery(id)
	if rowsCount == -2 {
		return true
	} else {
		return false
	}
}
