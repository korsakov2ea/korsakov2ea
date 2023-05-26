package xfunc

import (
	"log"
	"os"
)

// Запуск логирования в fileName
func StartLogging(fileName string) {

	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			log.Println(FuncName(), "Файл лога не существует, будет создан новый")
		} else {
			log.Println(FuncName(), "Ошибка настройки логгирования в файл. Логгируется в консоль", err)
			log.SetOutput(os.Stdout)
			return
		}
	}

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println(FuncName(), "Ошибка открытия файла лог файла", err)
	}

	log.SetOutput(file)
	log.Println()
	log.Println()
	log.Println()
	log.Println(FuncName(), "Включено логгирование в файл", fileName)

}
