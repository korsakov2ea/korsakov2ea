package x_func

import (
	"database/sql"
	"fmt"

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
	if db.DSN == "" {
		db.DSN = "HOSTNAME=" + db.HOSTNAME + "; DATABASE=" + db.DATABASE + "; PORT=" + db.PORT + "; UID=" + db.UID + "; PWD=" + db.PWD
	}
}

func DBOpen(driverName, dataSourceName string) *sql.DB {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		fmt.Println(err)
	}
	return db
}
