package xfunc

import (
	"log"

	"gopkg.in/ini.v1"
)

// Возвращает строчное значения ключа key из секции section файла filename и код ошибки
func GetIniValue(filename, section, key string) (value string) {

	cfg, err := ini.Load(filename)
	if err != nil {
		log.Println(FuncName(), "Ошибка чтения значения из INI файла", err)
	}

	value = cfg.Section(section).Key(key).String()
	return value
}
