package xfunc

import (
	"html/template"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
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

// Чтение куки с заданным именем
func GetCookie(r *http.Request, name string) string {
	result := ""
	for _, cookie := range r.Cookies() {
		if cookie.Name == name {
			result = cookie.Value
			break
		}
	}
	return result
}

// Установка куки с заданным именем, значением и продолжительностью
func SetCookie(w http.ResponseWriter, name string, value string, duration time.Duration) {
	expiration := time.Now().Add(duration)
	cookie := http.Cookie{Name: name, Value: value, Expires: expiration}
	http.SetCookie(w, &cookie)
}

// Аутентификация / авторизация по БИУД (проверка введеного логи, пароля и наличия ролей). Hashed = true - если передаем уже хэшированный пароль, если просто пароль то fasle
// Может возвращать значения (в порядке проверки):
// 0 - аутентификация не пройдена (нет пользователя или не один),
// 1 - авторизация не пройдена (нет нужной роли для пользователя),
// 2 - обе проверки пройдены
func AuthBIUD(login, pass string, hashed bool, biud *TDatabase) int {
	result := 0

	passHash := pass
	if !hashed {
		passHash = BiudPassHash(pass)
	}

	// искать оперетора в БИУД
	sqlCode := "SELECT * FROM CS.OPERATOR WHERE LOGIN='" + login + "'"
	operator := biud.DBQuery(sqlCode)
	// если есть такой логин
	if len(operator) == 1 {
		buidHash := operator[0].ByName["PASS"]

		// если совпали хэши паролей
		if buidHash == passHash {
			result++
			//искать роль пользователя
			sqlCode := "SELECT * FROM CS.OPERATORROLE WHERE LOGIN='" + login + "' AND ROLE='Администратор отделения'"
			// если есть роль
			if len(biud.DBQuery(sqlCode)) > 0 {
				result++
			}
		}
	}
	return result
}

// Собирает страницу из макета страницы (templatePage), подстановочной страницы (commonPage), объекта данных (date) и выводит в поток (w)
func RenderPage(w http.ResponseWriter, templatePage string, commonPage string, date interface{}) {
	templatePath := filepath.Join("public", "html", templatePage)
	commonPath := filepath.Join("public", "html", commonPage)

	tmpl, err := template.ParseFiles(templatePath, commonPath)
	if err != nil {
		log.Printf("%v Ошибка парсинга шаблона страницы %v", FuncName(), err)
		http.Error(w, err.Error(), 400)
	} else {
		log.Printf("%v Парсинг шаблона страницы %v", FuncName(), templatePath)
	}

	err = tmpl.Execute(w, date)
	if err != nil {
		log.Printf("%v Ошибка постооения шаблона страницы %v", FuncName(), err)
		http.Error(w, err.Error(), 400)
	} else {
		log.Printf("%v Построение шаблона страницы %v", FuncName(), templatePath)
	}
}
