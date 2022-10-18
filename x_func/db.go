package x_func

import (
	"database/sql"
	"log"

	_ "github.com/ibmdb/go_ibm_db"
)

type TDatabase struct {
	IniSection string
	Driver     string
	HOSTNAME   string
	DATABASE   string
	PORT       string
	UID        string
	PWD        string
	DSN        string
	DB         *sql.DB
}

// Считывает из конфигурационного INI файла настройки БД и устанавливает структуре db
func DBGetIniCfg(configFile string, iniSection string, db *TDatabase) {
	db.IniSection = iniSection
	db.Driver = GetIniValue(configFile, db.IniSection, "Driver")
	db.HOSTNAME = GetIniValue(configFile, db.IniSection, "HOSTNAME")
	db.DATABASE = GetIniValue(configFile, db.IniSection, "DATABASE")
	db.PORT = GetIniValue(configFile, db.IniSection, "PORT")
	db.UID = GetIniValue(configFile, db.IniSection, "UID")
	db.PWD = GetIniValue(configFile, db.IniSection, "PWD")
	db.DSN = GetIniValue(configFile, db.IniSection, "DSN")
	if db.DSN == "" {
		db.DSN = "HOSTNAME=" + db.HOSTNAME + "; DATABASE=" + db.DATABASE + "; PORT=" + db.PORT + "; UID=" + db.UID + "; PWD=" + db.PWD
	}
}

func (database *TDatabase) DBOpen() {
	var err error
	database.DB, err = sql.Open(database.Driver, database.DSN)
	if err != nil {
		log.Println(FuncName(), "Ошибка открытия соединения с базой", database.DATABASE, "на", database.HOSTNAME, err)
	}

	err = database.DB.Ping()
	if err != nil {
		log.Println(FuncName(), "Отсутствует пинг с базой", database.DATABASE, "на", database.HOSTNAME, err)
	} else {
		log.Println(FuncName(), "Уставлено соединение с базой", database.DATABASE, "на", database.HOSTNAME)
	}
}

func (database *TDatabase) DBExec(sqlCode string) {
	result, err := database.DB.Exec(sqlCode)
	if err != nil {
		log.Println(FuncName(), "Ошибка выполнения команды", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		log.Println(FuncName(), "Выполнения команды", sqlCode, ". Изменено строк -", rowsAffected)
	}
}

func (database *TDatabase) DBClose() {
	err := database.DB.Close()
	if err != nil {
		log.Println(FuncName(), "Ошибка закрытия соединения с базой", database.DATABASE, "на", database.HOSTNAME, err)
	} else {
		log.Println(FuncName(), "Cоединение с базой", database.DATABASE, "на", database.HOSTNAME, "закрыто")
	}

}
