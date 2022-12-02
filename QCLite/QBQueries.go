package main

import (
	"korsakov2ea/x_func"
	"log"
	"net/http"
	"sort"
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
			QBQuery.Directory = nil
			QBQuery.Directory = append(QBQuery.Directory, QBConnection.Data)
			renderPage(w, "query.html", "common.html", QBQuery)

		case r.Method == "GET" && r.FormValue("mode") == "view":
			QBQuery.Read(id)
			QBConnection.ReadAll()
			QBQuery.Directory = nil
			QBQuery.Directory = append(QBQuery.Directory, QBConnection.Data)
			renderPage(w, "query.html", "common.html", QBQuery)

		case r.Method == "GET" && r.FormValue("mode") == "execute":
			if needUploadData(id) {
				QBQuery.tmpTableName = "QB.QB" + x_func.GenerateTimeStamp()
				renderPage(w, "uploadfile.html", "common.html", QBQuery)
			} else {
				QBQuery.ExecQuery(id)
				renderPage(w, "result.html", "common.html", QBQuery)
			}

		case r.Method == "POST" && r.FormValue("submitBtn") == "uploadExecute":
			log.Printf("Загрузка файла с web формы \n%v", x_func.FuncName())
			CSVData := x_func.GetStrMapFromCSVWebFile(x_func.UploadFile(r, "uploadFile"))
			CSVData = x_func.DecodeStrMap1251toUTF8(CSVData)

			idConn := GetIdConnFromQuery(id)
			importTmpTable(CSVData, QBQuery.tmpTableName, idConn)
			QBQuery.ExecQuery(id)
			dropTmpTable(QBQuery.tmpTableName, idConn)
			renderPage(w, "result.html", "common.html", QBQuery)

		case r.Method == "POST" && r.FormValue("submitBtn") == "uploadDownload":
			log.Printf("Загрузка файла с web формы \n%v", x_func.FuncName())
			CSVData := x_func.GetStrMapFromCSVWebFile(x_func.UploadFile(r, "uploadFile"))
			CSVData = x_func.DecodeStrMap1251toUTF8(CSVData)

			idConn := GetIdConnFromQuery(id)
			importTmpTable(CSVData, QBQuery.tmpTableName, idConn)
			QBQuery.ExecQuery(id)
			dropTmpTable(QBQuery.tmpTableName, idConn)

			CSV := ""
			for _, row := range QBQuery.Data {

				keys := make([]string, 0, len(row))
				for k := range row {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				CSVRow := ""
				for _, k := range keys {
					CSVRow = CSVRow + row[k] + ";"
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

// Создание временной таблицы tableName (схема.имя) в соединении idСonn с колонками типа COL1 VARCHAR(255) ... COLn VARCHAR(255) и загрузка в неё данных из strMap
func importTmpTable(strMap [][]string, tableName string, idConn int) {

	log.Printf("%v \n\tСоздание временной таблицы %v", x_func.FuncName(), tableName)
	colNames, colNameTypes := x_func.GetCOLStrMap(strMap, "VARCHAR (255)")
	sqlCode := "CREATE TABLE " + tableName + " (" + colNameTypes + ")"
	ExecSQL(sqlCode, idConn)

	log.Printf("%v \n\tЗаполнение временной таблицы %v", x_func.FuncName(), tableName)
	sqlCode = ""
	for _, row := range strMap {
		colValues := ""
		for _, cell := range row {
			colValues = colValues + ", '" + cell + "'"
		}
		colValues = colValues[2:]
		sqlCode = sqlCode + "INSERT INTO " + QBQuery.tmpTableName + " (" + colNames + ") VALUES (" + colValues + ");\n"
	}
	ExecSQL(sqlCode, idConn)
}

// Удаление временной таблицы tableName (схема.имя) в соединении idСonn
func dropTmpTable(tableName string, idConn int) {
	log.Printf("%v \n\tУдаление временной таблицы %v", x_func.FuncName(), tableName)
	ExecSQL("DROP TABLE "+tableName, idConn)
	QBQuery.tmpTableName = ""
}

/*
func postFile(filename string, targetUrl string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(targetUrl, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
	return nil
}
*/
