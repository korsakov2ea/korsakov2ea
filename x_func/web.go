package x_func

import (
	"log"
	"mime/multipart"
	"net/http"
	"time"
)

// Уведомление для рендера страницы
type TAlert struct {
	Text  string // текст уведомления
	Class string // класс уведомления (alert-primary, -secondary, -success, -danger, -warning, -info, -light, -dark)
}

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

// Проверка наличия куки с заданным именем и значением
func CheckCookies(r *http.Request, name string, value string) bool {
	result := false
	for _, cookie := range r.Cookies() {
		if cookie.Name == name && cookie.Value == value {
			result = true
		}
	}
	return result
}

// Установка куки с заданным именем, значением и продолжительностью
func SetCookies(w http.ResponseWriter, name string, value string, duration time.Duration) {
	expiration := time.Now().Add(duration)
	cookie := http.Cookie{Name: name, Value: value, Expires: expiration}
	http.SetCookie(w, &cookie)
}
