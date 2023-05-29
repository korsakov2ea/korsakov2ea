package main

import (
	"korsakov2ea/xfunc"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// queries - обработчик HTTP (список запросов)
func queries(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v Переход к списку запросов ────────────────────────────────────────┐", xfunc.FuncName())
	defer log.Printf("%v Переход к списку запросов ────────────────────────────────────────┘", xfunc.FuncName())

	sqlErr = QBQuery.ReadSQL("SELECT Q.ID, Q.REM, Q.QUERY, Q.NAME, Q.ID_GROUP, C.NAME AS C_NAME, G.NAME AS G_NAME FROM QUERY AS Q LEFT JOIN CONNECTION AS C ON Q.ID_CONNECTION=C.ID LEFT JOIN QUERY_GROUP AS G ON G.ID=Q.ID_GROUP ORDER BY G_NAME, Q.ID")
	if sqlErr != nil {
		RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
	}
	RenderData.Data = QBQuery.Data
	RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Всего запросов - " + strconv.Itoa(len(RenderData.Data)), Class: "info"})
	xfunc.RenderPage(w, "queries.html", "common.html", RenderData)
	RenderData.Alerts = nil
}

// query - обработчик HTTP (одиночный запрос)
func query(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("ID"))
	if err != nil && r.Method == "POST" && (r.FormValue("submitBtn") == "Update" || r.FormValue("submitBtn") == "Delete") {
		log.Printf("%v Ошибка преобразования ID = %v из GET запроса в число", xfunc.FuncName(), r.FormValue("ID"))
	} else {
		log.Printf("%v HTTP запрос с параметрами %v", xfunc.FuncName(), r.URL.RawQuery)

		switch {

		// Отмена изменения запроса
		case r.Method == "POST" && r.FormValue("submitBtn") == "cancel":
			log.Printf("%v Отмена изменения запроса ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil
			RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Нажата кнопка [Отмена]", Class: "info"})
			http.Redirect(w, r, "queries", http.StatusFound)
			log.Printf("%v Отмена изменения запроса ────────────────────────────────────────┘", xfunc.FuncName())

		// Создание запроса
		case r.Method == "POST" && r.FormValue("submitBtn") == "create":
			log.Printf("%v Создание запроса ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil
			newQuery := make(map[string]string)
			newQuery["NAME"] = r.FormValue("Name")
			newQuery["QUERY"] = strings.Replace(r.FormValue("Query"), "'", "''", -1)
			newQuery["REM"] = strings.Replace(r.FormValue("Rem"), "'", "''", -1)
			newQuery["ID_CONNECTION"] = r.FormValue("Id_connection")
			newQuery["ID_GROUP"] = r.FormValue("Id_group")

			sqlErr = QBQuery.Create(newQuery)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}

			// добавление параметров из запроса в базу
			params := xfunc.GetParamsFromStr(newQuery["QUERY"])
			if len(params) > 0 {
				sqlErr = QBQuery.ReadSQL("SELECT ID FROM QUERY WHERE NAME='" + newQuery["NAME"] + "'")
				if sqlErr != nil {
					RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
				}
				newParam := make(map[string]string)
				for param_key, param := range params {
					newParam["ID_QUERY"] = QBQuery.Data[0].ByName["ID"]
					newParam["NAME"] = param_key
					newParam["REM"] = param.Rem
					newParam["VALUE"] = param.DefaultValue
					sqlErr = QBParam.Create(newParam)
					if sqlErr != nil {
						RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
					}
				}
			}

			if sqlErr == nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Запрос создан", Class: "success"})
			}
			http.Redirect(w, r, "queries", http.StatusFound)
			log.Printf("%v Создание запроса ────────────────────────────────────────┘", xfunc.FuncName())

		// Изменение запроса
		case r.Method == "POST" && r.FormValue("submitBtn") == "update":
			log.Printf("%v Нажата кнопка СОЗДАТЬ на странице создания запроса ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil
			newQuery := make(map[string]string)
			newQuery["NAME"] = r.FormValue("Name")
			newQuery["QUERY"] = strings.Replace(r.FormValue("Query"), "'", "''", -1)
			newQuery["REM"] = strings.Replace(r.FormValue("Rem"), "'", "''", -1)
			newQuery["ID_CONNECTION"] = r.FormValue("Id_connection")
			newQuery["ID_GROUP"] = r.FormValue("Id_group")
			sqlErr = QBQuery.Update(id, newQuery)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}

			// удаление...
			sqlErr = QBQuery.ReadSQL("SELECT ID FROM PARAM WHERE ID_QUERY=" + strconv.Itoa(id))
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}
			if len(QBQuery.Data) > 0 {
				sqlErr = QB.DBExec("DELETE FROM PARAM WHERE ID_QUERY=" + strconv.Itoa(id))
				if sqlErr != nil {
					RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
				}
			}

			// ...и повторное добавление параметров из запроса в базу
			params := xfunc.GetParamsFromStr(newQuery["QUERY"])
			if len(params) > 0 {
				newParam := make(map[string]string)
				for param_key, param_data := range params {
					newParam["ID_QUERY"] = strconv.Itoa(id)
					newParam["NAME"] = param_key
					newParam["REM"] = param_data.Rem
					newParam["VALUE"] = param_data.DefaultValue
					sqlErr = QBParam.Create(newParam)
					if sqlErr != nil {
						RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
					}
				}
			}
			if sqlErr == nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Запрос изменён", Class: "success"})
			}
			http.Redirect(w, r, "queries", http.StatusFound)
			log.Printf("%v Изменение запроса ────────────────────────────────────────┘", xfunc.FuncName())

		// Удаление запроса
		case r.Method == "POST" && r.FormValue("submitBtn") == "delete":
			log.Printf("%v Нажата кнопка УДАЛИТЬ на странице редактирования запроса ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil
			sqlErr = QBQuery.Delete(id)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}

			andParam := ""
			// удаление параметров
			sqlErr = QBQuery.ReadSQL("SELECT ID FROM PARAM WHERE ID_QUERY=" + strconv.Itoa(id))
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}
			if len(QBQuery.Data) > 0 {
				sqlErr = QB.DBExec("DELETE FROM PARAM WHERE ID_QUERY=" + strconv.Itoa(id))
				if sqlErr != nil {
					RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
				} else {
					andParam = " с параметрами"
				}
			}

			if sqlErr == nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Запрос" + andParam + " удалён", Class: "success"})
			}
			http.Redirect(w, r, "queries", http.StatusFound)
			log.Printf("%v Удаление запроса ────────────────────────────────────────┘", xfunc.FuncName())

		// Переход к добавлению запроса
		case r.Method == "GET" && r.FormValue("mode") == "add":
			if user.access == "admin" {
				log.Printf("%v Переход к добавлению запроса ────────────────────────────────────────┐", xfunc.FuncName())
				RenderData.Alerts = nil
				RenderData.Data = nil
				sqlErr = QBConnection.ReadAll(0)
				if sqlErr != nil {
					RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
				}
				sqlErr = QBGroup.ReadAll(0)
				if sqlErr != nil {
					RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
				}
				RenderData.Dict = make(map[string]interface{})
				RenderData.Dict["CONNECTIONS"] = QBConnection.Data
				RenderData.Dict["GROUPS"] = QBGroup.Data
				xfunc.RenderPage(w, "query.html", "common.html", RenderData)
				log.Printf("%v Переход к добавлению запроса ────────────────────────────────────────┘", xfunc.FuncName())
			} else {
				RenderData.Alerts = nil
				RenderData.Data = nil
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Данный функционал доступен только для администраторов отделения", Class: "danger"})
				http.Redirect(w, r, "queries", http.StatusFound)
			}

		// Переход к изменению запроса
		case r.Method == "GET" && r.FormValue("mode") == "edit":
			if user.access == "admin" {
				log.Printf("%v Переход к изменению запроса ────────────────────────────────────────┐", xfunc.FuncName())
				RenderData.Alerts = nil

				sqlErr = QBQuery.Read(id)
				if sqlErr != nil {
					RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
				}
				RenderData.Data = QBQuery.Data

				sqlErr = QBConnection.ReadAll(0)
				if sqlErr != nil {
					RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
				}
				sqlErr = QBGroup.ReadAll(0)
				if sqlErr != nil {
					RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
				}

				RenderData.Dict = make(map[string]interface{})
				RenderData.Dict["CONNECTIONS"] = QBConnection.Data
				RenderData.Dict["GROUPS"] = QBGroup.Data
				xfunc.RenderPage(w, "query.html", "common.html", RenderData)
				log.Printf("%v ППереход к изменению запроса ────────────────────────────────────────┘", xfunc.FuncName())
			} else {
				RenderData.Alerts = nil
				RenderData.Data = nil
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Данный функционал доступен только для администраторов отделения", Class: "danger"})
				http.Redirect(w, r, "queries", http.StatusFound)
			}

		// Переход к выполнению запроса
		case r.Method == "GET" && r.FormValue("mode") == "open":
			log.Printf("%v Переход к выполнению запроса ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil
			RenderData.Data = nil
			RenderData.Dict = make(map[string]interface{})
			RenderData.Param = make(map[string]xfunc.TParamOptions)

			// получение параметров из базы для верстки страницы
			sqlErr = QBParam.ReadSQL("SELECT * FROM PARAM WHERE ID_QUERY=" + strconv.Itoa(id))
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}
			if len(QBParam.Data) > 0 {
				for _, param_key := range QBParam.Data {
					RenderData.Param[param_key.ByName["NAME"]] = xfunc.TParamOptions{Rem: param_key.ByName["REM"], DefaultValue: param_key.ByName["VALUE"]}
				}
			}

			sqlErr = QBQuery.Read(id)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}
			RenderData.Dict["QUERY_NAME"] = QBQuery.Data[0].ByName["NAME"]
			RenderData.Dict["QUERY_REM"] = QBQuery.Data[0].ByName["REM"]
			RenderData.Dict["QUERY_SQL"] = QBQuery.Data[0].ByName["QUERY"]

			// проверка необходимости загрузки файла
			if needUploadData(id) {
				RenderData.Dict["SubTableName"] = "QB.QB" + xfunc.GenerateTimeStamp()
			} else {
				RenderData.Dict["SubTableName"] = ""
			}

			xfunc.RenderPage(w, "query_open.html", "common.html", RenderData)
			log.Printf("%v Переход к выполнению запроса ────────────────────────────────────────┘", xfunc.FuncName())

		// Выполнение запроса
		case r.Method == "POST" && (r.FormValue("submitBtn") == "viewResult" || r.FormValue("submitBtn") == "downloadResult"):
			log.Printf("%v Выполнение запроса ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil
			RenderData.Data = nil

			// получение параметров со страницы для использования в запросе
			for formKey := range r.Form {
				if strings.Contains(formKey, "param_") {
					paramOption := xfunc.TParamOptions{Rem: RenderData.Param[formKey[6:]].Rem, CurrentValue: r.Form[formKey][0], DefaultValue: RenderData.Param[formKey[6:]].DefaultValue}
					RenderData.Param[formKey[6:]] = paramOption
				}
			}

			needUpload := needUploadData(id)
			dataUploaded := false
			idConn, _ := GetIdConnFromQuery(id)

			if needUpload {
				RenderData.Dict["SubTableName"] = "QB.QB" + xfunc.GenerateTimeStamp()
				log.Printf("Загрузка файла с web формы \n%v", xfunc.FuncName())
				if len(r.MultipartForm.File) > 0 {
					CSVData := xfunc.GetStrMapFromCSVWebFile(xfunc.UploadFile(r, "uploadFile"))
					CSVData = xfunc.DecodeStrMap1251toUTF8(CSVData)
					sqlErr = importTmpTable(CSVData, RenderData.Dict["SubTableName"].(string), idConn)
					if sqlErr != nil {
						RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
					} else {
						dataUploaded = true
					}
				} else {
					log.Printf("Не выбран файл для загрузки. Запрос не выполнен \n%v", xfunc.FuncName())
					RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Не выбран файл для загрузки. Запрос не выполнен", Class: "danger"})
				}
			}

			// если не требудется загрузка или файл был успешно загружен в базу
			if !needUpload || dataUploaded {
				RenderData.Data, sqlErr = executeQuery(id)
				if sqlErr != nil {
					RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
				}
				switch {
				case r.FormValue("submitBtn") == "viewResult":
					xfunc.RenderPage(w, "query_open.html", "common.html", RenderData)
				case r.FormValue("submitBtn") == "downloadResult":
					CSV := getCSVfromData(RenderData.Data)
					CSV = xfunc.DecodeStrUTF8to1251(CSV)
					fileName := "QUERY_" + strconv.Itoa(id) + "_RESULT" + xfunc.GenerateTimeStamp() + ".csv"
					w.Header().Set("Content-Type", "application/octet-stream")
					w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(CSV))
				}
			} else if !dataUploaded {
				xfunc.RenderPage(w, "query_open.html", "common.html", RenderData)
			}

			// если файл был загружен в базу то удалить
			if dataUploaded {
				dropTmpTable(RenderData.Dict["SubTableName"].(string), idConn)
			}

			log.Printf("%v Выполнение запроса ────────────────────────────────────────┘", xfunc.FuncName())

		default:
		}
	}
}

// Выполняет запрос из базы запросов под номером id и возвращает возвращает результат
func executeQuery(id int) (result xfunc.TResultRows, err error) {

	log.Printf("%v Выполнение запроса из QB c id = %v", xfunc.FuncName(), id)
	queryRows, _ := QB.DBQuery("SELECT C.DRIVER, C.DSN, C.NAME, Q.QUERY FROM QUERY AS Q INNER JOIN CONNECTION AS C ON Q.ID_CONNECTION=C.ID AND Q.ID=" + strconv.Itoa(id))
	queryRowsCount := len(queryRows)

	var targetDB xfunc.TDatabase //соединение для выполнения запроса из базы

	if queryRowsCount != 0 {
		targetDB.Driver = queryRows[0].ByName["DRIVER"]
		targetDB.DSN = queryRows[0].ByName["DSN"]
		targetDB.Name = queryRows[0].ByName["NAME"]
		targetQuery := queryRows[0].ByName["QUERY"]

		targetDB.SetDecodeParam()

		targetDB.DBOpen()
		defer targetDB.DBClose()

		for param_key, param_data := range RenderData.Param {
			paramStr := "{" + param_key + ":" + param_data.Rem + ":" + param_data.DefaultValue + "}"
			targetQuery = strings.Replace(targetQuery, paramStr, param_data.CurrentValue, -1)
		}

		if len(RenderData.Dict["SubTableName"].(string)) > 0 {
			targetQuery = strings.Replace(targetQuery, "#TABLE#", RenderData.Dict["SubTableName"].(string), -1)
		}

		result, err = targetDB.DBQuery(targetQuery)

	} else {
		log.Println(xfunc.FuncName(), "Нет строки запроса с ID -", id)
	}
	return result, err

}

// Возвращает true, если в запросе с id есть метка #TABLE#, т.е. предполагается загрузка внешних данных
func needUploadData(id int) bool {
	log.Printf("%v Проверка необходимости загрузки данных перед выполнением запроса (см. ниже)", xfunc.FuncName())
	queryRows, _ := QB.DBQuery("SELECT QUERY FROM QUERY WHERE ID=" + strconv.Itoa(id))
	queryRowsCount := len(queryRows)
	result := false
	if queryRowsCount != 0 {
		if strings.Contains(queryRows[0].ByName["QUERY"], "#TABLE#") {
			log.Printf("%v Необходима загрузка данных из файла", xfunc.FuncName())
			result = true
			return result
		} else {
			log.Printf("%v Запрос выполняется без загрузки файла", xfunc.FuncName())
		}
	} else {
		log.Println(xfunc.FuncName(), "Нет строки запроса с ID -", id)
	}
	return result
}

// Создание временной таблицы tableName (схема.имя) в соединении idСonn с колонками типа COL1 VARCHAR(255) ... COLn VARCHAR(255) и загрузка в неё данных из strMap
func importTmpTable(strMap [][]string, tableName string, idConn int) error {

	log.Printf("%v \n\tСоздание временной таблицы %v", xfunc.FuncName(), tableName)
	colNames, colNameTypes := xfunc.GetCOLStrMap(strMap, "VARCHAR (255)")
	sqlCode := "CREATE TABLE " + tableName + " (" + colNameTypes + ")"
	executeSQLConn(sqlCode, idConn)

	log.Printf("%v \n\tЗаполнение временной таблицы %v", xfunc.FuncName(), tableName)
	sqlCode = ""
	for _, row := range strMap {
		colValues := ""
		for _, cell := range row {
			colValues = colValues + ", '" + cell + "'"
		}
		colValues = colValues[2:]

		sqlCode = sqlCode + "INSERT INTO " + RenderData.Dict["SubTableName"].(string) + " (" + colNames + ") VALUES (" + colValues + ");\n"

	}
	return executeSQLConn(sqlCode, idConn)
}

// Удаление временной таблицы tableName (схема.имя) в соединении idСonn
func dropTmpTable(tableName string, idConn int) {
	log.Printf("%v \n\tУдаление временной таблицы %v", xfunc.FuncName(), tableName)
	executeSQLConn("DROP TABLE "+tableName, idConn)
	//QBResult.SubTableName = ""
}

func getCSVfromData(data xfunc.TResultRows) string {
	CSV := ""
	// получение заголовков из первой строки
	for _, col := range data[0].Ind {
		CSV = CSV + *col + ";"
	}
	CSV = CSV + "\r\n"

	// получение данных в порядке как вернула СУБД (по порядку ключей)
	for _, row := range data {

		CSVRow := ""
		for _, cell := range row.ByInd {
			CSVRow = CSVRow + *cell + ";"
		}
		CSV = CSV + CSVRow + "\r\n"
	}
	return CSV
}
