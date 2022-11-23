package x_func

import (
	"log"
	"mime/multipart"
	"net/http"
)

// UploadFile - получение multipart.File файла с формы, где inputName - name соответствующего input-а с формы
func UploadFile(r *http.Request, inputName string) multipart.File {
	log.Printf("%v Загрузка файла с формы", FuncName())
	source, _, err := r.FormFile(inputName)
	if err != nil {
		log.Printf("%v Ошибка получения файла c формы", FuncName())
	} else {
		log.Printf("%v Считан файл %v", FuncName(), r.MultipartForm.File["uploadFile"][0].Filename)
		return source
	}
	defer source.Close()
	return nil
}
