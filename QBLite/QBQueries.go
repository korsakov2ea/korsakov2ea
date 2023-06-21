package main

import (
	"encoding/csv"
	"korsakov2ea/xfunc"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

// Возвращает все запросы +
func queries(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v --------- HTTP METHOD: %v, PARAM: %v", xfunc.FuncName(), r.Method, r.URL.RawQuery)
	sqlErr = QBQuery.DFReadSQL("SELECT Q.ID, Q.REM, Q.QUERY, Q.NAME, Q.ID_GROUP, C.NAME AS C_NAME, G.NAME AS G_NAME FROM QUERY AS Q LEFT JOIN CONNECTION AS C ON Q.ID_CONNECTION=C.ID LEFT JOIN QUERY_GROUP AS G ON G.ID=Q.ID_GROUP ORDER BY G_NAME, Q.ID")
	if sqlErr != nil {
		RenderData.AddAlert(sqlErr.Error(), "danger")
	}
	RenderData.DataMap = QBQuery.DataFrame.Maps()
	RenderData.RenderMap(w, "queries.html", "common.html")
	RenderData.Clear()
}

// Обработчик событий при работе с запросом
func query(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("ID"))
	if err != nil && r.Method == "POST" && (r.FormValue("submitBtn") == "Update" || r.FormValue("submitBtn") == "Delete") {
		log.Printf("%v Ошибка преобразования ID = %v из GET запроса в число", xfunc.FuncName(), r.FormValue("ID"))
	} else {
		log.Printf("%v HTTP запрос с параметрами %v", xfunc.FuncName(), r.URL.RawQuery)
		RenderData.Clear()
		switch {

		// Отмена изменения запроса +
		case r.Method == "POST" && r.FormValue("submitBtn") == "Cancel":
			RenderData.AddAlert("Нажата кнопка [Отмена]", "info")
			http.Redirect(w, r, "queries", http.StatusFound)

		// Создание / изменение запроса +
		case r.Method == "POST" && (r.FormValue("submitBtn") == "Create" || r.FormValue("submitBtn") == "Update"):
			newQuery := make(map[string]string)
			newQuery["NAME"] = r.FormValue("Name")
			newQuery["QUERY"] = strings.Replace(r.FormValue("Query"), "'", "''", -1)
			newQuery["REM"] = strings.Replace(r.FormValue("Rem"), "'", "''", -1)
			newQuery["ID_CONNECTION"] = r.FormValue("Id_connection")
			newQuery["ID_GROUP"] = r.FormValue("Id_group")
			action := ""
			switch r.FormValue("submitBtn") {
			case "Create":
				RenderData.AddAlertIfErr(QBQuery.Create(newQuery))
				action = "создан"
			case "Update":
				RenderData.AddAlertIfErr(QBQuery.Update(id, newQuery))
				// удаление параметров...
				RenderData.AddAlertIfErr(QB.DBExec("DELETE FROM PARAM WHERE ID_QUERY=" + strconv.Itoa(id)))
				action = "изменен"
			}

			// добавление параметров в базу
			RenderData.AddAlertIfErr(QBQuery.DFReadSQL("SELECT ID FROM QUERY WHERE NAME='" + newQuery["NAME"] + "'"))
			newQueryID := QBQuery.DataFrame.Elem(0, 0).String()
			newParam := make(map[string]string)
			for formKey := range r.Form {
				if strings.Contains(formKey, "val@") {
					newParam["ID_QUERY"] = newQueryID
					newParam["NAME"] = strings.Replace(formKey, "val@", "", 1)
					newParam["REM"] = r.Form[strings.Replace(formKey, "val@", "rem", 1)][0]
					newParam["VALUE"] = r.Form[formKey][0]
					RenderData.AddAlertIfErr(QBParam.Create(newParam))
				}
			}

			RenderData.AddAlertIfOk("Запрос " + action)
			http.Redirect(w, r, "queries", http.StatusFound)

		// Переход к добавлению / изменению запроса +
		case (r.Method == "GET" && (r.FormValue("mode") == "add" || r.FormValue("mode") == "edit")) || (r.Method == "POST" && r.FormValue("submitBtn") == "InitParam"):
			if r.FormValue("mode") == "edit" {
				RenderData.GetOneFromTable(&QBQuery, id)
				RenderData.AddAlertIfErr(QBParam.DFReadSQL("SELECT * FROM PARAM WHERE ID_QUERY=" + strconv.Itoa(id)))
				RenderData.AddDictFromMaps("PARAMS", QBParam.DataFrame.Maps())
			}
			if r.FormValue("submitBtn") == "InitParam" {
				query := strings.Replace(r.FormValue("Query"), "'", "''", -1)
				params := xfunc.GetParam(query)
				if len(params) > 0 {
					RenderData.AddDictFromMaps("PARAMS", dataframe.New(series.New(params, series.String, "NAME")).Maps())
				}
				RenderData.DataMap[0]["QUERY"] = r.FormValue("Query")
			}
			RenderData.AddDictFromTable("CONNECTIONS", &QBConnection)
			RenderData.AddDictFromTable("GROUPS", &QBGroup)
			RenderData.RenderMap(w, "query.html", "common.html")

		// Переход к выполнению запроса
		case r.Method == "GET" && r.FormValue("mode") == "open":
			// получение параметров из базы для верстки страницы
			RenderData.AddAlertIfErr(QBParam.DFReadSQL("SELECT * FROM PARAM WHERE ID_QUERY=" + strconv.Itoa(id)))
			RenderData.AddDictFromMaps("PARAMS", QBParam.DataFrame.Maps())
			// получение запроса из базы для верстки страницы (название и описание)
			RenderData.AddAlertIfErr(QBQuery.DFRead(id))
			RenderData.AddDictFromMaps("QUERY", QBQuery.DataFrame.Maps())
			// проверка необходимости загрузки файла
			if needUploadData(id) {
				RenderData.Dict["NEED_CSV"] = true
			}
			RenderData.RenderMap(w, "query_open.html", "common.html")

		// Выполнение запроса
		case r.Method == "POST" && (r.FormValue("submitBtn") == "viewResult" || r.FormValue("submitBtn") == "downloadResult"):
			// получение параметров со страницы
			params := make(map[string]string)
			for formKey := range r.Form {
				if strings.Contains(formKey, "val@") {
					NAME := strings.Replace(formKey, "val@", "", 1)
					VALUE := r.Form[formKey][0]
					params[NAME] = VALUE
				}
			}

			// получение запроса из базы для верстки страницы (название и описание)
			RenderData.AddAlertIfErr(QBQuery.DFRead(id))
			RenderData.AddDictFromMaps("QUERY", QBQuery.DataFrame.Maps())

			idConn, _ := GetIdConnFromQuery(id)
			needUpload := needUploadData(id)
			dataUploaded := false
			tmpTableName := ""

			if needUpload {
				log.Printf("%v Загрузка файла с web формы", xfunc.FuncName())
				if len(r.MultipartForm.File) > 0 {
					multiPartFile := xfunc.UploadFile(r, "uploadFile")
					csvDataFrame := dataframe.ReadCSV(multiPartFile, dataframe.WithDelimiter(';'), dataframe.WithLazyQuotes(false), dataframe.HasHeader(true))
					log.Printf("%v Получен массив размерностью [%v : %v]", xfunc.FuncName(), csvDataFrame.Ncol(), csvDataFrame.Nrow())

					recordsWithoutHeaders := csvDataFrame.Records()[1:]
					strMapForImport := xfunc.DecodeStrMap1251toUTF8(recordsWithoutHeaders)
					tmpTableName = "QB.QB" + xfunc.GenerateTimeStamp()
					if RenderData.AddAlertIfErr(importTmpTable(strMapForImport, tmpTableName, idConn)) == nil {
						dataUploaded = true
					}
				} else {
					log.Printf("%v Не выбран файл для загрузки. Запрос не выполнен", xfunc.FuncName())
					RenderData.AddAlert("Не выбран файл для загрузки. Запрос не выполнен", "danger")
				}
			}

			// если не требуется загрузка или файл был успешно загружен в базу
			if !needUpload || dataUploaded {
				resultDataFrame, sqlErr := executeQuery(id, params, tmpTableName)

				newLogRecord := make(map[string]string)
				newLogRecord["ID_QUERY"] = strconv.Itoa(id)
				newLogRecord["ID_USER"] = user.Name
				newLogRecord["TIME"] = xfunc.GenerateTimeStamp()
				newLogRecord["RESULT"] = "Получено строк: " + strconv.Itoa(resultDataFrame.Nrow())

				if sqlErr != nil {
					RenderData.AddAlert(sqlErr.Error(), "danger")
					newLogRecord["RESULT"] = sqlErr.Error()
				}

				QBLog.Create(newLogRecord)

				switch {
				case r.FormValue("submitBtn") == "viewResult":
					RenderData.DataMap = resultDataFrame.Maps()
					RenderData.RenderMap(w, "query_open.html", "common.html")
				case r.FormValue("submitBtn") == "downloadResult":
					fileName := "QUERY_" + strconv.Itoa(id) + "_RESULT" + xfunc.GenerateTimeStamp() + ".csv"
					w.Header().Set("Content-Type", "application/octet-stream")
					w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
					w.WriteHeader(http.StatusOK)
					csv := csv.NewWriter(w)
					csv.Comma = ';'
					csv.UseCRLF = true
					csv.WriteAll(xfunc.DecodeStrMapUTF8to1251(resultDataFrame.Records()))
				}
			} else if !dataUploaded {
				RenderData.RenderMap(w, "query_open.html", "common.html")
			}

			// если файл был загружен в базу, то удалить
			if dataUploaded {
				dropTmpTable(tmpTableName, idConn)
			}

		default:
			w.Write([]byte("Неожиданный запрос! " + r.URL.RawPath + " " + r.URL.RawQuery))
		}
	}
}

// Выполняет запрос из базы запросов под номером id и возвращает возвращает результат
func executeQuery(id int, paramMap map[string]string, tmpTableName string) (result dataframe.DataFrame, err error) {
	log.Printf("%v Выполнение запроса из QB c id = %v", xfunc.FuncName(), id)
	queryRows, _ := QB.DBDFQuery("SELECT C.DRIVER, C.DSN, C.NAME, Q.QUERY FROM QUERY AS Q INNER JOIN CONNECTION AS C ON Q.ID_CONNECTION=C.ID AND Q.ID=" + strconv.Itoa(id))

	var targetDB xfunc.TDatabase //соединение для выполнения запроса из базы

	if queryRows.Nrow() != 0 {
		targetDB.Driver = queryRows.Maps()[0]["DRIVER"].(string)
		targetDB.DSN = queryRows.Maps()[0]["DSN"].(string)
		targetDB.Name = queryRows.Maps()[0]["NAME"].(string)
		targetQuery := queryRows.Maps()[0]["QUERY"].(string)

		targetDB.SetDecodeParam()
		targetDB.DBOpen()
		defer targetDB.DBClose()

		for param_name, param_val := range paramMap {
			targetQuery = strings.Replace(targetQuery, "{@"+param_name+"}", param_val, -1)
		}
		if tmpTableName != "" {
			targetQuery = strings.Replace(targetQuery, "#TABLE#", tmpTableName, -1)
		}
		result, err = targetDB.DBDFQuery(targetQuery)
	} else {
		log.Println(xfunc.FuncName(), "Нет строки запроса с ID -", id)
	}
	return result, err
}

// Возвращает true, если в запросе с id есть метка #TABLE#, т.е. предполагается загрузка внешних данных
func needUploadData(id int) bool {
	log.Printf("%v Проверка необходимости загрузки данных перед выполнением запроса (см. ниже)", xfunc.FuncName())
	queryRows, err := QB.DBDFQuery("SELECT QUERY FROM QUERY WHERE ID=" + strconv.Itoa(id))
	if err != nil {
		log.Printf("%v Ошибка получения запроса по id", xfunc.FuncName())
		return false
	} else {
		queryRowsCount := queryRows.Nrow()
		result := false
		if queryRowsCount != 0 {
			if strings.Contains(queryRows.Elem(0, 0).String(), "#TABLE#") {
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
			colValues = colValues + ", '" + strings.ReplaceAll(cell, "'", "''") + "'"
		}
		colValues = colValues[2:]

		sqlCode = sqlCode + "INSERT INTO " + tableName + " (" + colNames + ") VALUES (" + colValues + ");\n"

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
