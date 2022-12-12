package xfunc

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// FuncName - получает имя вызвавшей функции (используется для вывода в лог)
func FuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

// GetExecFilePath - получает путь с исполняемому файлу без последнего слэша
func GetExecFilePath() (execFilePath string) {
	exFile, err := os.Executable()
	if err != nil {
		log.Println(FuncName(), "Ошибка определения пути исполняемому файлу", err)
	} else {
		execFilePath = filepath.Dir(exFile)
	}
	return
}

func GenerateTimeStamp() string {
	dateTime := time.Now()
	return dateTime.Format("20060102_150405.000000")[:15] + dateTime.Format("20060201T150405.000000")[16:]
}
