package xfunc

import (
	"database/sql"
	"log"
	"strconv"

	_ "github.com/ibmdb/go_ibm_db"
	_ "github.com/mattn/go-sqlite3"
)

// ---------------------------------------------------------------------------------------------------------------------------- ОПИСАНИЕ ТИПОВ ДАННЫХ
// Представление базы данных
type TDatabase struct {
	Driver      string  // тип драйвера (go_ibm_db или go_ibm_db)
	DSN         string  // источник данных (data source name)
	Name        string  // имя БД для вывода в лог
	DB          *sql.DB // ссылка на БД
	DecodeParam bool    // необходимость перекодирования из UTF8 в Windows1251 (нужно для баз DB2)
}

// Представление таблицы базы данных
type TTable struct {
	db   *TDatabase  // ссылка на родительскую БД
	name string      // наименоение таблицы, при необходимости Схема.Таблица
	Data TResultRows // + срез с данными
}

// Представление ОДНОЙ строки результатов SQL выборки. Реализует доступ к данным (по имени, по индексу) и ключам (по индексу).
type TResultRow struct {
	ByName map[string]string // карта вида "КЛЮЧ: ЗНАЧЕНИЕ", хранит данные (в алфавитном порядке ключей)
	ByInd  []*string         // массив (в порядке как вернула СУБД) со ссылками на ЗНАЧЕНИЯ
	Ind    []*string         // массив (в порядке как вернула СУБД) со ссылками на КЛЮЧИ
}

// Представление ВСЕХ строк с результами SQL выборки. Реализует доступ к данным (по имени, по индексу) и ключам (по индексу).
// ByName - карта вида "КЛЮЧ: ЗНАЧЕНИЕ", хранит непосредственно данные (не сортируется, в алфавитном порядке ключей).
// ByInd - массив (в порядке как вернула СУБД) со ссылками на ЗНАЧЕНИЯ.
// Ind - массив (в порядке как вернула СУБД) со ссылками на КЛЮЧИ (заголовки столбцов).
type TResultRows []TResultRow

// ---------------------------------------------------------------------------------------------------------------------------- РАБОТА С БАЗОЙ ДАННЫХ (НИЗКИЙ УРОВЕНЬ)

// Читает настройки БД из конфигурационного INI файла и устанавливает структуре db
func DBGetIniCfg(configFile string, iniSection string, db *TDatabase) {
	db.Driver = GetIniValue(configFile, iniSection, "Driver")
	db.Name = GetIniValue(configFile, iniSection, "Name")
	db.DSN = GetIniValue(configFile, iniSection, "DSN")

	DATABASE := GetIniValue(configFile, iniSection, "DATABASE")
	HOSTNAME := GetIniValue(configFile, iniSection, "HOSTNAME")
	PORT := GetIniValue(configFile, iniSection, "PORT")
	PROTOCOL := GetIniValue(configFile, iniSection, "PROTOCOL")
	UID := GetIniValue(configFile, iniSection, "UID")
	PWD := GetIniValue(configFile, iniSection, "PWD")

	DSN := "DATABASE=" + DATABASE + "; HOSTNAME=" + HOSTNAME + "; PORT=" + PORT + "; PROTOCOL=" + PROTOCOL + "; UID=" + UID + "; PWD=" + PWD
	if db.DSN == "" && DSN != "" {
		db.DSN = DSN
	}

	log.Printf("%v Считана конфигурация БД из секции %v файла %v", FuncName(), iniSection, configFile)
	db.SetDecodeParam()
}

// Устанавливает значение database.DecodeParam в зависимости от database.Driver
func (database *TDatabase) SetDecodeParam() {
	if database.Driver == "go_ibm_db" {
		database.DecodeParam = true
	} else {
		database.DecodeParam = false
	}
	log.Printf("%v Установлено значение DecodeParam = %v для БД %v", FuncName(), database.DecodeParam, database.Name)
}

// Открывает соедиенние с БД и пингует его
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
		log.Printf("%v Соединение с базой %v успешно открыто. Есть PING", FuncName(), database.Name)
	}
}

// Закрывает соединение с БД
func (database *TDatabase) DBClose() {
	err := database.DB.Close()
	if err != nil {
		log.Printf("%v Ошибка закрытия соединения с базой %v %v", FuncName(), database.Name, err)
	} else {
		log.Printf("%v Соединение с базой %v успешно закрыто", FuncName(), database.Name)
	}
}

// Выполняет SQL инструкцию, которые не возвращают результат (например INSERT)
func (database *TDatabase) DBExec(sqlCode string) error {
	result, err := database.DB.Exec(sqlCode)
	shortSqlCode := sqlCode
	if len(shortSqlCode) > 1000 {
		shortSqlCode = shortSqlCode[:1000] + "... "
	}
	if err != nil {
		log.Printf("%v Ошибка выполнения SQL команды %v %v", FuncName(), shortSqlCode, err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		log.Printf("%v Выполнена SQL команда. (Строк изменено - %d)\n\t%v", FuncName(), rowsAffected, shortSqlCode)
	}
	return err
}

// Выполняет SQL инструкцию, которые возвращают результат (например SELECT)
func (database *TDatabase) DBQuery(sqlCode string) (TResultRows, error) {
	var result TResultRows
	rows, err := database.DB.Query(sqlCode)
	shortSqlCode := sqlCode
	if len(shortSqlCode) > 1000 {
		shortSqlCode = shortSqlCode[:1000] + "... "
	}
	if err != nil {
		log.Printf("%v Ошибка выполнения SQL запроса:\n\t%v\n\t%v", FuncName(), shortSqlCode, err)
	} else {
		log.Printf("%v Выполнение SQL запроса:\n\t%v", FuncName(), shortSqlCode)
		result = rowsToResult(rows, database.DecodeParam)
		rows.Close()
	}
	return result, err
}

// ---------------------------------------------------------------------------------------------------------------------------- РАБОТА С ДАННЫМИ (СРЕДНИЙ УРОВЕНЬ)

// Ассоциирует объект TTable по наименованию реальной таблицы (СХЕМА.ТАБЛИЦА) и БД
func (tab *TTable) Bind(tableName string, database *TDatabase) {
	log.Printf("%v Ассоциация таблицы %v в базе %v", FuncName(), tableName, database.Name)
	tab.db = database
	tab.name = tableName
	tab.Data = nil
}

// Cчитывает запись из таблицы по id и сохраняет результат в Data
func (tab *TTable) Read(id int) (err error) {
	log.Printf("%v Чтение записи из %v с id = %v", FuncName(), tab.name, id)
	sqlCode := "SELECT * FROM " + tab.name + " WHERE ID=" + strconv.Itoa(id)
	tab.Data, err = tab.db.DBQuery(sqlCode)
	return err
}

// Считывает все записи из таблицы и сохраняет результат в Data, если fetchRowsCount <= 0. Иначе считывает первые fetchRowsCount строк
func (tab *TTable) ReadAll(fetchRowsCount int) (err error) {
	fetchStatement := ""
	if fetchRowsCount > 0 {
		fetchStatement = " FETCH FIRST " + strconv.Itoa(fetchRowsCount) + " ROWS ONLY"
	}
	log.Printf("%v Чтение всех записей из %v", FuncName(), tab.name)
	sqlCode := "SELECT * FROM " + tab.name + fetchStatement
	tab.Data, err = tab.db.DBQuery(sqlCode)
	return err
}

// Считывает записи из таблицы SQL запросом и сохраняет результат в Data
func (tab *TTable) ReadSQL(sqlCode string) (err error) {
	log.Printf("%v Чтение записей из таблицы %v запросом SQL (см. ниже)", FuncName(), tab.name)
	tab.Data, err = tab.db.DBQuery(sqlCode)
	return err
}

// Добавляет в таблицу новую запись из карты insertingData
func (tab *TTable) Create(insertingData map[string]string) error {
	log.Printf("%v Создание записи в %v", FuncName(), tab.name)
	tab.Data = nil
	var cols string = ""
	var vals string = ""
	for name, val := range insertingData {
		if name[:1] == "_" {
			name = name[1:]
		} else {
			val = "'" + val + "'"
		}
		cols = cols + name + ", "
		vals = vals + val + ", "
	}
	sqlCode := "INSERT INTO " + tab.name + " (" + cols[:len(cols)-2] + ") VALUES (" + vals[:len(vals)-2] + ")"
	return tab.db.DBExec(sqlCode)
}

// Заменяет запись в таблице по id значениями из карты updatingData
func (tab *TTable) Update(id int, updatingData map[string]string) error {
	log.Printf("%v Изменение записи из %v с id = %v", FuncName(), tab.name, id)
	tab.Data = nil
	var cols_vals string = ""
	for col, val := range updatingData {
		if col[:1] == "_" {
			col = col[1:]
		} else {
			val = "'" + val + "'"
		}
		cols_vals = cols_vals + col + " = " + val + ", "
	}
	sqlCode := "UPDATE " + tab.name + " SET " + cols_vals[:len(cols_vals)-2] + " WHERE ID = " + strconv.Itoa(id)
	return tab.db.DBExec(sqlCode)
}

// Удаляет запись из таблицы по id
func (tab *TTable) Delete(id int) error {
	log.Printf("%v Удаление записи из %v с id = %v", FuncName(), tab.name, id)
	tab.Data = nil
	sqlCode := "DELETE FROM " + tab.name + " WHERE ID = " + strconv.Itoa(id)
	return tab.db.DBExec(sqlCode)
}

// ---------------------------------------------------------------------------------------------------------------------------- ОБРАБОТЧИКИ

// Возвращает ссылку на значение переменной типа string. Необходимо для релизации rowsToResult (обхода ограничений GO)
func strAdr(str string) *string {
	return &str
}

// Преобразует *sql.Rows в TResultRows (массив карт со значениями и индексированными ссылками на значения и ключи)
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
