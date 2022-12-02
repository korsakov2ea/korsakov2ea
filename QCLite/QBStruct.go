package main

import (
	"korsakov2ea/x_func"
	"log"
	"strconv"
	"strings"
)

// Структура для описание сущностей БД (соединение, запрос и т.д.)
type QBEntity struct {
	name         string       // название как в БД
	Data         [][]string   // двумерный срез (массив) строк для хранения результатов выборки
	DataRows     int          // кол-во строк результатов выборки
	Directory    [][][]string // массив карт строк для хранения справочников
	tmpTableName string       // имя временной таблицы для загрузки внешних данных
}

// Create - добавляет в БД новую запись из карты entityMap
func (qbe *QBEntity) Create(entityMap map[string]string) {
	log.Printf("%v Создание %v", x_func.FuncName(), qbe.name)
	qbe.Data = nil
	var cols string = ""
	var vals string = ""
	for name, val := range entityMap {
		cols = cols + name + ", "
		vals = vals + "'" + val + "', "
	}
	sqlCode := "INSERT INTO " + qbe.name + " (" + cols[:len(cols)-2] + ") VALUES (" + vals[:len(vals)-2] + ")"
	QCDB.DBExec(sqlCode)
}

// Read - считывает из БД запись по id
func (qbe *QBEntity) Read(id int) {
	log.Printf("%v Чтение %v для id = %v", x_func.FuncName(), qbe.name, id)
	sqlCode := "SELECT * FROM " + qbe.name + " WHERE ID=" + strconv.Itoa(id)
	qbe.Data, _ = QCDB.DBQuery(sqlCode, false)
}

// Update - заменяет запись в БД по id значениями из карты entityMap
func (qbe *QBEntity) Update(id int, entityMap map[string]string) {
	log.Printf("%v Изменение %v для id = %v", x_func.FuncName(), qbe.name, id)
	qbe.Data = nil
	var cols_vals string = ""
	for col, val := range entityMap {
		cols_vals = cols_vals + col + " = '" + val + "', "
	}
	sqlCode := "UPDATE " + qbe.name + " SET " + cols_vals[:len(cols_vals)-2] + " WHERE ID = " + strconv.Itoa(id)
	QCDB.DBExec(sqlCode)
}

// Delete - удаляет запись из БД по id
func (qbe *QBEntity) Delete(id int) {
	log.Printf("%v Удаление %v для id = %v", x_func.FuncName(), qbe.name, id)
	qbe.Data = nil
	sqlCode := "DELETE FROM " + qbe.name + " WHERE ID = " + strconv.Itoa(id)
	QCDB.DBExec(sqlCode)
}

// ReadAll - считывает из БД все записи
func (qbe *QBEntity) ReadAll() {
	log.Printf("%v Чтение всех %v", x_func.FuncName(), qbe.name)

	sqlCode := "SELECT * FROM " + qbe.name
	qbe.Data, _ = QCDB.DBQuery(sqlCode, false)
}

// ReadSQL - считывает из БД записи по SQL
func (qbe *QBEntity) ReadSQL(sqlCode string) {
	log.Printf("%v Чтение данные чистым SQL для %v", x_func.FuncName(), qbe.name)
	qbe.Data, _ = QCDB.DBQuery(sqlCode, false)
}

// execQuery - выполняет запрос из базы QC под номером id и возвращает массив карт с результатами выборки, а также кол-во срок в результате.
// Кол-во строк -1 означает, что нет строки запроса с указанным ID.
func (qbe *QBEntity) ExecQuery(id int) {
	log.Printf("%v Выполнение запроса", x_func.FuncName())
	var tmpConn x_func.TDatabase //соединение для выполнения запроса из базы
	qcStringMap, qcRowCount := QCDB.DBQuery("SELECT DRIVER, DSN, NAME, QUERY FROM QUERY AS Q INNER JOIN CONNECTION AS C ON Q.ID_CONNECTION=C.ID AND Q.ID="+strconv.Itoa(id), false)
	if qcRowCount != 0 {
		tmpConn.Driver = qcStringMap[0][0]
		tmpConn.DSN = qcStringMap[0][1]
		tmpConn.Name = qcStringMap[0][2]
		decodeParam := false
		if tmpConn.Driver == "go_ibm_db" {
			decodeParam = true
		}
		tmpConn.DBOpen()
		defer tmpConn.DBClose()

		query := qcStringMap[0][3]
		if len(qbe.tmpTableName) > 0 {
			query = strings.Replace(query, "@TABLE", qbe.tmpTableName, -1)
		}

		qbe.Data, qbe.DataRows = tmpConn.DBQuery(query, decodeParam)

	} else {
		log.Println(x_func.FuncName(), "Нет строки запроса с ID -", id)
	}
}

// needUploadData - возвращает true, если в запросе с idQuery есть метка @TABLE, т.е. предполагается загрузка внешних данных
func needUploadData(idQuery int) bool {
	log.Printf("%v Проверка необходимости загрузки данных перед выполнением запроса", x_func.FuncName())
	qcStringMap, qcRowCount := QCDB.DBQuery("SELECT QUERY FROM QUERY WHERE ID="+strconv.Itoa(idQuery), false)
	result := false
	if qcRowCount != 0 {
		if strings.Contains(qcStringMap[0][0], "@TABLE") {
			log.Printf("%v Необходима загрузка данных из файла", x_func.FuncName())
			result = true
			return result
		}
	} else {
		log.Println(x_func.FuncName(), "Нет строки запроса с ID -", idQuery)
	}
	return result
}

// GetIdConnFromQuery - получение id соединения по id запроса
func GetIdConnFromQuery(idQuery int) int {
	log.Printf("%v Получение ID соединения для запроса с ID %v", x_func.FuncName(), idQuery)
	qcStringMap, qcRowCount := QCDB.DBQuery("SELECT ID_CONNECTION FROM QUERY WHERE ID="+strconv.Itoa(idQuery), false)
	idConn := -1
	if qcRowCount != 0 {
		id, err := strconv.Atoi(qcStringMap[0][0])
		if err != nil {
			log.Println(x_func.FuncName(), "Ошибка преобразования qcStringMap[0][\"ID_CONNECTION\"] = %v в число", qcStringMap[0][0])
		} else {
			idConn = id
		}
	} else {
		log.Println(x_func.FuncName(), "Нет строки запроса с ID -", idQuery)
	}
	return idConn
}

// ExecSQL - выполняет SQL команду в базе, по соединению c указанным id
func ExecSQL(sqlCode string, idConn int) {
	log.Printf("%v \n\tВыполнение SQL команды \n%v", x_func.FuncName(), sqlCode)
	var tmpConn x_func.TDatabase //соединение для выполнения запроса из базы
	qcStringMap, qcRowCount := QCDB.DBQuery("SELECT DRIVER, DSN, NAME FROM CONNECTION WHERE ID="+strconv.Itoa(idConn), false)
	if qcRowCount != 0 {
		tmpConn.Driver = qcStringMap[0][0]
		tmpConn.DSN = qcStringMap[0][1]
		tmpConn.Name = qcStringMap[0][2]
		tmpConn.DBOpen()
		defer tmpConn.DBClose()
		tmpConn.DBExec(sqlCode)
	} else {
		log.Printf("%v \n\tНет соединения с ID = %v", x_func.FuncName(), idConn)
	}
}
