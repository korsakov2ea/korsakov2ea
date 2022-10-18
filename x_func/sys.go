package x_func

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// FuncName - Получение имени вызвавшей функции (используется для вывода в лог)
func FuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

// getExecFilePath - Возвращает путь с исполняемому файлу без последнего слэша
func GetExecFilePath() (execFilePath string) {
	exFile, err := os.Executable()
	if err != nil {
		log.Println(FuncName(), "Ошибка определения пути исполняемому файлу", err)
	} else {
		execFilePath = filepath.Dir(exFile)
	}
	return
}
