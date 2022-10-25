package x_func

import (
	"database/sql"
	"log"

	_ "github.com/ibmdb/go_ibm_db"
	_ "github.com/mattn/go-sqlite3"
)

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
func (database *TDatabase) DBQuery(sqlCode string, decode1251toUTF8 bool) (SliceMap []map[string]string, RowCount int) {
	rows, err := database.DB.Query(sqlCode)
	defer rows.Close()
	if err != nil {
		log.Printf("%v Ошибка выполнения SQL запроса %v %v", FuncName(), sqlCode, err)
	} else {
		log.Printf("%v Выполнение SQL запроса %v", FuncName(), sqlCode)
	}
	return rowsToMap(rows, decode1251toUTF8)
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
		log.Println(FuncName(), "Ошибка преобразования sql.Row в Map", err)
	}

	columns := make([]string, len(cols))
	columnPointers := make([]interface{}, len(cols))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}

	RowCount = 0
	for rows.Next() {
		err = rows.Scan(columnPointers...)
		if err != nil {
			log.Println(err)
		}

		currentMap := make(map[string]string)
		for i, columnName := range cols {
			val := columnPointers[i].(*string)
			if decode1251toUTF8 {
				currentMap[columnName] = Decode1251toUTF8(*val)
			} else {
				currentMap[columnName] = *val
			}
		}

		SliceMap = append(SliceMap, currentMap)
		RowCount++
	}
	return SliceMap, RowCount
}
