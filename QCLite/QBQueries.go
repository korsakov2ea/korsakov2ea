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
	log.Printf("%v Вызов обработчика HTTP запроса", x_func.FuncName())
	QBQuery.ReadSQL("SELECT Q.ID, Q.REM, Q.QUERY, C.NAME FROM QUERY AS Q INNER JOIN CONNECTION AS C ON Q.ID_CONNECTION=C.ID")
	renderPage(w, "queries.html", "common.html", QBQuery)
}

// query - обработчик HTTP (одиночный запрос)
func query(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v  Вызов обработчика HTTP запроса", x_func.FuncName())

	id, err := strconv.Atoi(r.FormValue("ID"))
	if err != nil && r.Method == "POST" && (r.FormValue("submitBtn") == "Update" || r.FormValue("submitBtn") == "Delete") {
		log.Printf("%v Ошибка преобразования ID = %v из GET запроса в число", x_func.FuncName(), r.FormValue("ID"))
	} else {

		switch {

		case r.Method == "POST" && r.FormValue("submitBtn") == "cancel":
			QBQuery.ReadAll()
			renderPage(w, "queries.html", "common.html", QBQuery)

		case r.Method == "POST" && r.FormValue("submitBtn") == "create":
			newQuery := make(map[string]string)
			newQuery["NAME"] = r.FormValue("Name")
			newQuery["QUERY"] = strings.Replace(r.FormValue("Query"), "'", "''", -1)
			newQuery["REM"] = strings.Replace(r.FormValue("Rem"), "'", "''", -1)
			newQuery["ID_CONNECTION"] = r.FormValue("Id_connection")
			QBQuery.Create(newQuery)
			QBQuery.ReadAll()
			renderPage(w, "queries.html", "common.html", QBQuery)

		case r.Method == "POST" && r.FormValue("submitBtn") == "update":
			newQuery := make(map[string]string)
			newQuery["NAME"] = r.FormValue("Name")
			newQuery["QUERY"] = strings.Replace(r.FormValue("Query"), "'", "''", -1)
			newQuery["REM"] = strings.Replace(r.FormValue("Rem"), "'", "''", -1)
			newQuery["ID_CONNECTION"] = r.FormValue("Id_connection")
			QBQuery.Update(id, newQuery)
			QBQuery.ReadAll()
			renderPage(w, "queries.html", "common.html", QBQuery)

		case r.Method == "POST" && r.FormValue("submitBtn") == "delete":
			QBQuery.Delete(id)
			QBQuery.ReadAll()
			renderPage(w, "queries.html", "common.html", QBQuery)

		case r.Method == "GET" && r.FormValue("mode") == "add":
			QBQuery.Data = nil
			QBConnection.ReadAll()
			QBQuery.Dict = nil
			QBQuery.Dict = append(QBQuery.Dict, QBConnection.Data)
			renderPage(w, "query.html", "common.html", QBQuery)

		case r.Method == "GET" && r.FormValue("mode") == "view":
			QBQuery.Read(id)
			QBConnection.ReadAll()
			QBQuery.Dict = nil
			QBQuery.Dict = append(QBQuery.Dict, QBConnection.Data)
			renderPage(w, "query.html", "common.html", QBQuery)

		case r.Method == "GET" && r.FormValue("mode") == "execute":
			if needUploadData(id) {
				QBQuery.TmpTable = "QB.QB" + x_func.GenerateTimeStamp()
				renderPage(w, "uploadfile.html", "common.html", QBQuery)
			} else {
				executeQuery(id)
				renderPage(w, "result.html", "common.html", QBQuery)
			}

		case r.Method == "POST" && r.FormValue("submitBtn") == "uploadExecute":
			log.Printf("Загрузка файла с web формы \n%v", x_func.FuncName())
			CSVData := x_func.GetStrMapFromCSVWebFile(x_func.UploadFile(r, "uploadFile"))
			CSVData = x_func.DecodeStrMap1251toUTF8(CSVData)

			idConn := GetIdConnFromQuery(id)
			importTmpTable(CSVData, QBQuery.TmpTable, idConn)
			executeQuery(id)
			dropTmpTable(QBQuery.TmpTable, idConn)
			renderPage(w, "result.html", "common.html", QBQuery)

		case r.Method == "POST" && r.FormValue("submitBtn") == "uploadDownload":
			log.Printf("Загрузка файла с web формы \n%v", x_func.FuncName())
			CSVData := x_func.GetStrMapFromCSVWebFile(x_func.UploadFile(r, "uploadFile"))
			CSVData = x_func.DecodeStrMap1251toUTF8(CSVData)

			idConn := GetIdConnFromQuery(id)
			importTmpTable(CSVData, QBQuery.TmpTable, idConn)
			executeQuery(id)
			dropTmpTable(QBQuery.TmpTable, idConn)

			CSV := ""
			for _, row := range QBQuery.Data {

				/*
					keys := make([]string, 0, len(row))
					for k := range row {
						keys = append(keys, k)
					}
					sort.Strings(keys)
				*/
				CSVRow := ""
				for _, cell := range row.ByName {
					CSVRow = CSVRow + cell + ";"
				}
				CSV = CSV + CSVRow + "\r\n"
			}
			CSV = x_func.DecodeStrUTF8to1251(CSV)

			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", "attachment; filename=result.csv")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(CSV))

		default:
		}
	}
}

// Выполняет запрос из базы запросов под номером id и возвращает возвращает результат в QBQuery.Data
func executeQuery(id int) {
	log.Printf("%v Выполнение запроса из QB c id = %v", x_func.FuncName(), id)
	queryRows := QB.DBQuery("SELECT C.DRIVER, C.DSN, C.NAME, Q.QUERY FROM QUERY AS Q INNER JOIN CONNECTION AS C ON Q.ID_CONNECTION=C.ID AND Q.ID=" + strconv.Itoa(id))
	queryRowsCount := len(queryRows)

	var targetDB x_func.TDatabase //соединение для выполнения запроса из базы

	if queryRowsCount != 0 {
		targetDB.Driver = queryRows[0].ByName["DRIVER"]
		targetDB.DSN = queryRows[0].ByName["DSN"]
		targetQuery := queryRows[0].ByName["QUERY"]
		targetDB.SetDecodeParam()

		targetDB.DBOpen()
		defer targetDB.DBClose()

		if len(QBQuery.TmpTable) > 0 {
			targetQuery = strings.Replace(targetQuery, "@TABLE", QBQuery.TmpTable, -1)
		}
		QBQuery.Data = targetDB.DBQuery(targetQuery)

	} else {
		log.Println(x_func.FuncName(), "Нет строки запроса с ID -", id)
	}
}

// Возвращает true, если в запросе с id есть метка @TABLE, т.е. предполагается загрузка внешних данных
func needUploadData(id int) bool {
	log.Printf("%v Проверка необходимости загрузки данных перед выполнением запроса", x_func.FuncName())
	queryRows := QB.DBQuery("SELECT QUERY FROM QUERY WHERE ID=" + strconv.Itoa(id))
	queryRowsCount := len(queryRows)
	result := false
	if queryRowsCount != 0 {
		if strings.Contains(queryRows[0].ByName["QUERY"], "@TABLE") {
			log.Printf("%v Необходима загрузка данных из файла", x_func.FuncName())
			result = true
			return result
		}
	} else {
		log.Println(x_func.FuncName(), "Нет строки запроса с ID -", id)
	}
	return result
}

// Создание временной таблицы tableName (схема.имя) в соединении idСonn с колонками типа COL1 VARCHAR(255) ... COLn VARCHAR(255) и загрузка в неё данных из strMap
func importTmpTable(strMap [][]string, tableName string, idConn int) {

	log.Printf("%v \n\tСоздание временной таблицы %v", x_func.FuncName(), tableName)
	colNames, colNameTypes := x_func.GetCOLStrMap(strMap, "VARCHAR (255)")
	sqlCode := "CREATE TABLE " + tableName + " (" + colNameTypes + ")"
	executeSQLConn(sqlCode, idConn)

	log.Printf("%v \n\tЗаполнение временной таблицы %v", x_func.FuncName(), tableName)
	sqlCode = ""
	for _, row := range strMap {
		colValues := ""
		for _, cell := range row {
			colValues = colValues + ", '" + cell + "'"
		}
		colValues = colValues[2:]
		sqlCode = sqlCode + "INSERT INTO " + QBQuery.TmpTable + " (" + colNames + ") VALUES (" + colValues + ");\n"
	}
	executeSQLConn(sqlCode, idConn)
}

// Удаление временной таблицы tableName (схема.имя) в соединении idСonn
func dropTmpTable(tableName string, idConn int) {
	log.Printf("%v \n\tУдаление временной таблицы %v", x_func.FuncName(), tableName)
	executeSQLConn("DROP TABLE "+tableName, idConn)
	QBQuery.TmpTable = ""
}
