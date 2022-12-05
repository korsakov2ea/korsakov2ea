package x_func

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/ibmdb/go_ibm_db"
	_ "github.com/mattn/go-sqlite3"
)

// Представление базы данных
type TDatabase struct {
	Driver string  // тип драйвера (go_ibm_db или go_ibm_db)
	DSN    string  // источник данных (data source name)
	Name   string  // имя БД для вывода в лог
	DB     *sql.DB // ссылка на БД
}

// Представление одной строки результатов SQL выборки. Реализует доступ к данным (по имени, по индексу) и ключам (по индексу).
type TResultRow struct {
	ByName map[string]string // карта вида "КЛЮЧ: ЗНАЧЕНИЕ", хранит данные (в алфавитном порядке ключей)
	ByInd  []*string         // массив (в порядке как вернула СУБД) со ссылками на ЗНАЧЕНИЯ
	Ind    []*string         // массив (в порядке как вернула СУБД) со ссылками на КЛЮЧИ
}

// Представление множества строй с результами SQL выборки. Реализует доступ к данным (по имени, по индексу) и ключам (по индексу).
// ByName - карта вида "КЛЮЧ: ЗНАЧЕНИЕ", хранит непосредственно данные (не сортируется, в алфавитном порядке ключей).
// ByInd - массив (в порядке как вернула СУБД) со ссылками на ЗНАЧЕНИЯ.
// Ind - массив (в порядке как вернула СУБД) со ссылками на КЛЮЧИ (заголовки столбцов).
type TResultRows []TResultRow

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

// УСТАРЕЛО 2022-12-03
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
func (database *TDatabase) DBQuery(sqlCode string, decode1251toUTF8 bool) (result TResultRows) {
	rows, err := database.DB.Query(sqlCode)
	if err != nil {
		log.Printf("%v Ошибка выполнения SQL запроса %v %v", FuncName(), sqlCode, err)
	} else {
		log.Printf("%v Выполнение SQL запроса %v", FuncName(), sqlCode)
	}
	defer rows.Close()
	return rowsToResult(rows, decode1251toUTF8)
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

// УСТАРЕЛО 2022-12-03
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

// УСТАРЕЛО 2022-12-03
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

// Возвращает ссылку на значение переменной типа string. Необходимо для обхода ограничений GO
func strAdr(str string) *string {
	return &str
}

// Преобразует *sql.Rows в TResultRows (массив карт со значениями и индексированными ссылками на значения и ключи).
// В случае decode1251toUTF8 = true изменяет кодировку
func rowsToResult(rows *sql.Rows, decode1251toUTF8 bool) (resultRows TResultRows) {
	cols, err := rows.Columns()
	if err != nil {
		log.Println(FuncName(), "Ошибка получения списка столбцов из *sql.Rows.Columns", err)
	}

	columns := make([]sql.NullString, len(cols))
	columnPointers := make([]interface{}, len(cols))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}

	var resultRow TResultRow

	for rows.Next() {
		err = rows.Scan(columnPointers...)
		if err != nil {
			log.Println(FuncName(), "Ошибка сканирования значений *sql.Rows", err)
		}

		rowMap := make(map[string]string)
		rowValuePointers := make([]*string, len(cols))
		rowKeyPointers := make([]*string, len(cols))

		for i, columnName := range cols {
			val := columnPointers[i].(*sql.NullString)
			if decode1251toUTF8 {
				rowMap[columnName] = DecodeStr1251toUTF8(val.String)
			} else {
				rowMap[columnName] = val.String
			}
			if !val.Valid {
				rowMap[columnName] = "NULL"
			}
			rowValuePointers[i] = strAdr(rowMap[columnName])
			rowKeyPointers[i] = strAdr(columnName)
		}

		resultRow.ByName = rowMap
		resultRow.ByInd = rowValuePointers
		resultRow.Ind = rowKeyPointers
		resultRows = append(resultRows, resultRow)
	}

	return resultRows
}
