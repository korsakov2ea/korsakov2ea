package x_func

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/ibmdb/go_ibm_db"
	_ "github.com/mattn/go-sqlite3"
)

type TNamedIndStr struct {
	ByName map[string]string
	ByInd  []*string
}

type TDatabase struct {
	Driver string
	Name   string
	DSN    string
	DB     *sql.DB
}

// DBGetIniCfg - считывает из конфигурационного INI файла настройки БД и устанавливает структуре db
func DBGetIniCfg(configFile string, iniSection string, db *TDatabase) {
	db.Driver = GetIniValue(configFile, iniSection, "Driver")
	db.Name = GetIniValue(configFile, iniSection, "Name")
	db.DSN = GetIniValue(configFile, iniSection, "DSN")
	log.Printf("%v Считана конфигурация из секции %v файла %v", FuncName(), iniSection, configFile)
}

// DBOpen - метод TDatabase для открытия и проверки соединения с БД
func (database *TDatabase) DBOpen() {
	var err error
	database.DB, err = sql.Open(database.Driver, database.DSN)
	if err != nil {
		log.Printf("%v Ошибка открытия соединения с базой %v %v", FuncName(), database.Name, err)
	}

	err = database.DB.Ping()
	if err != nil {
		log.Printf("%v Отсутствует пинг с базой %v %v", FuncName(), database.Name, err)
	} else {
		log.Printf("%v Подтверждено соединение (ping) с базой %v", FuncName(), database.Name)
	}
}

// DBExec - метод TDatabase для выполения SQL инструкций, которые не возвращают результат (например INSERT)
func (database *TDatabase) DBExec(sqlCode string) {
	result, err := database.DB.Exec(sqlCode)
	if err != nil {
		log.Printf("%v Ошибка выполнения SQL команды %v %v", FuncName(), sqlCode, err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		log.Printf("%v Выполнена SQL команда. (Строк изменено - %d) %v", FuncName(), rowsAffected, sqlCode)
	}
}

// DBQuery - метод TDatabase для выполения SQL инструкций, которые возвращают результат (например SELECT). Возвращает карту значение и кол-во строк
func (database *TDatabase) DBQueryMap(sqlCode string, decode1251toUTF8 bool) (SliceMap [][]string, RowCount int) {
	rows, err := database.DB.Query(sqlCode)
	if err != nil {
		log.Printf("%v Ошибка выполнения SQL запроса %v %v", FuncName(), sqlCode, err)
	} else {
		log.Printf("%v Выполнение SQL запроса %v", FuncName(), sqlCode)
	}
	defer rows.Close()
	return rowsToSlice(rows, decode1251toUTF8)
}

// DBQuery - метод TDatabase для выполения SQL инструкций, которые возвращают результат (например SELECT). Возвращает карту значение и кол-во строк
func (database *TDatabase) DBQuery(sqlCode string, decode1251toUTF8 bool) (SliceMap []TNamedIndStr, RowCount int) {
	rows, err := database.DB.Query(sqlCode)
	if err != nil {
		log.Printf("%v Ошибка выполнения SQL запроса %v %v", FuncName(), sqlCode, err)
	} else {
		log.Printf("%v Выполнение SQL запроса %v", FuncName(), sqlCode)
	}
	defer rows.Close()
	return rowsToData(rows, decode1251toUTF8)
}

// DBClose - метод TDatabase для закрытия соединения с БД
func (database *TDatabase) DBClose() {
	err := database.DB.Close()
	if err != nil {
		log.Printf("%v Ошибка закрытия соединения с базой %v %v", FuncName(), database.Name, err)
	} else {
		log.Printf("%v Соединение с базой %v успешно закрыто", FuncName(), database.Name)
	}

}

// rowsToMap - преобразует sql.Rows в массив карт. В случае decode1251toUTF8 = true изменяет кодировку
func rowsToMap(rows *sql.Rows, decode1251toUTF8 bool) (SliceMap []map[string]string, RowCount int) {
	cols, err := rows.Columns()
	if err != nil {
		log.Println(FuncName(), "Ошибка получения списка столбцов из *sql.Rows.Columns", err)
	}

	columns := make([]sql.NullString, len(cols))
	columnPointers := make([]interface{}, len(cols))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}

	RowCount = 0
	for rows.Next() {
		err = rows.Scan(columnPointers...)
		if err != nil {
			log.Println(FuncName(), "Ошибка сканирования значений *sql.Rows", err)
		}

		currentMap := make(map[string]string)
		for i, columnName := range cols {
			val := columnPointers[i].(*sql.NullString)
			if decode1251toUTF8 {
				currentMap[columnName] = DecodeStr1251toUTF8(val.String)
			} else {
				currentMap[columnName] = val.String
			}
			if !val.Valid {
				currentMap[columnName] = "NULL"
			}
		}

		SliceMap = append(SliceMap, currentMap)
		RowCount++
	}

	return SliceMap, RowCount
}

// rowsToSlice - преобразует sql.Rows в двумерныый срез (массив). В случае decode1251toUTF8 = true изменяет кодировку
func rowsToSlice(rows *sql.Rows, decode1251toUTF8 bool) (Slice [][]string, RowCount int) {
	cols, err := rows.Columns()
	if err != nil {
		log.Println(FuncName(), "Ошибка получения списка столбцов из *sql.Rows.Columns", err)
	} else {
		Slice = append(Slice, cols)
	}

	columns := make([]sql.NullString, len(cols))
	columnPointers := make([]interface{}, len(cols))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}

	RowCount = 0
	for rows.Next() {
		err = rows.Scan(columnPointers...)
		if err != nil {
			log.Println(FuncName(), "Ошибка сканирования значений *sql.Rows", err)
		}

		currentSlice := make([]string, len(cols))
		for i := range cols {
			val := columnPointers[i].(*sql.NullString)
			if decode1251toUTF8 {
				currentSlice[i] = DecodeStr1251toUTF8(val.String)
			} else {
				currentSlice[i] = val.String
			}
			if !val.Valid {
				currentSlice[i] = "NULL"
			}
		}

		Slice = append(Slice, currentSlice)
		RowCount++
	}
	fmt.Print(Slice)
	return Slice, RowCount
}

func strAdr(str string) *string {
	return &str
}

// rowsToMap - преобразует sql.Rows в массив карт. В случае decode1251toUTF8 = true изменяет кодировку
func rowsToData(rows *sql.Rows, decode1251toUTF8 bool) (SliceData []TNamedIndStr, RowCount int) {
	cols, err := rows.Columns()
	if err != nil {
		log.Println(FuncName(), "Ошибка получения списка столбцов из *sql.Rows.Columns", err)
	}

	columns := make([]sql.NullString, len(cols))
	columnPointers := make([]interface{}, len(cols))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}

	RowCount = 0
	var currentData TNamedIndStr
	for rows.Next() {
		err = rows.Scan(columnPointers...)
		if err != nil {
			log.Println(FuncName(), "Ошибка сканирования значений *sql.Rows", err)
		}

		currentMap := make(map[string]string)
		currentSlice := make([]*string, len(cols))
		for i, columnName := range cols {
			val := columnPointers[i].(*sql.NullString)
			if decode1251toUTF8 {
				currentMap[columnName] = DecodeStr1251toUTF8(val.String)
			} else {
				currentMap[columnName] = val.String
			}
			if !val.Valid {
				currentMap[columnName] = "NULL"
			}
			currentSlice[i] = strAdr(currentMap[columnName])
		}

		currentData.ByName = currentMap
		currentData.ByInd = currentSlice
		SliceData = append(SliceData, currentData)
		RowCount++
	}

	return SliceData, RowCount
}
