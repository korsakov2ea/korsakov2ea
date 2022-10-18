package x_func

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/text/encoding/charmap"
)

// FuncName - получает имя вызвавшей функции (используется для вывода в лог)
func FuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

// getExecFilePath - получает путь с исполняемому файлу без последнего слэша
func GetExecFilePath() (execFilePath string) {
	exFile, err := os.Executable()
	if err != nil {
		log.Println(FuncName(), "Ошибка определения пути исполняемому файлу", err)
	} else {
		execFilePath = filepath.Dir(exFile)
	}
	return
}

// Decode1251toUTF8 - декодирует строку из 1251 в UTF8
func Decode1251toUTF8(W1251 string) string {
	decoder := charmap.Windows1251.NewDecoder()
	UTF8, err := decoder.String(W1251)
	if err != nil {
		log.Println(FuncName(), "Ошибка декодирования строки", err)
	}
	return UTF8
}
