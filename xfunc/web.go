package xfunc

import (
	"html/template"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"
)

// Структура для передачи данных в предствление (рендер)
type TRenderData struct {
	Alerts  []TAlert                 // срез уведомлений
	Data    TResultRows              // основные данные для рендера (например: результаты выборки) в виде структуры
	DataMap []map[string]interface{} // основные данные для рендера (например: результаты выборки) в виде карты
	Dict    map[string]interface{}   // карта вспомогательных данных - словарь (например: название, пояснение, результаты подзапросов и т.д.)
	Param   map[string]TParamOptions // карта параметров для выполнения запроса
	User    *TUser                   // пользователь вызвавший метод (для рендера в зависимоти от прав)
}

// Информация о пользователе
type TUser struct {
	Name, Access string
}

// Опции параметров запроса
type TParamOptions struct {
	Rem          string // описание параметра
	DefaultValue string // значение по умолчанию из базы
	CurrentValue string // текущее значение с формы
}

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
// 2 - есть роль Администратора Отделения
// 3 - нет роли Администратора Отделения, но есть одна из: 'контроль', 'архив', 'назначение и перерасчёт', 'приём'
func AuthBIUD(login, pass string, hashed bool, biud *TDatabase) int {
	result := 0

	passHash := pass
	if !hashed {
		passHash = BiudPassHash(pass)
	}

	// искать оператора в БИУД
	sqlCode := "SELECT * FROM CS.OPERATOR WHERE LOGIN='" + login + "'"
	operator, _ := biud.DBQuery(sqlCode)
	// если есть такой логин
	if len(operator) == 1 {
		buidHash := operator[0].ByName["PASS"]

		// если совпали хэши паролей
		if buidHash == passHash {
			result = 1

			sqlCode := "SELECT * FROM CS.OPERATORROLE WHERE LOGIN='" + login + "' AND ROLE='Администратор отделения'"
			role, _ := biud.DBQuery(sqlCode)
			if len(role) > 0 {
				result = 2
			}

			sqlCode = `SELECT * FROM CS.OPERATORROLE WHERE LOGIN='` + login + `'
				AND NOT EXISTS (SELECT 1 FROM CS.OPERATORROLE WHERE LOGIN='` + login + `' AND ROLE = 'Администратор отделения')
				AND EXISTS (SELECT 1 FROM CS.OPERATORROLE WHERE LOGIN='` + login + `' AND ROLE IN ('контроль', 'архив', 'назначение и перерасчёт', 'приём'))`
			role, _ = biud.DBQuery(sqlCode)
			if len(role) > 0 {
				result = 3
			}

		}
	}
	return result
}

// Собирает страницу из макета страницы (templatePage), подстановочной страницы (commonPage) и выводит в поток (w)
func (RenderData *TRenderData) RenderMap(w http.ResponseWriter, templatePage string, commonPage string) {
	templatePath := filepath.Join("public", "html", templatePage)
	commonPath := filepath.Join("public", "html", commonPage)

	tmpl, err := template.ParseFiles(templatePath, commonPath)
	if err != nil {
		log.Printf("%v Ошибка парсинга шаблона страницы %v", FuncName(), err)
		http.Error(w, err.Error(), 400)
	} else {
		log.Printf("%v Парсинг шаблона страницы %v", FuncName(), templatePath)
	}

	err = tmpl.Execute(w, RenderData)
	if err != nil {
		log.Printf("%v Ошибка постооения шаблона страницы %v", FuncName(), err)
		http.Error(w, err.Error(), 400)
	} else {
		log.Printf("%v Построение шаблона страницы %v", FuncName(), templatePath)
	}
}

// Удаляет все уведомления из среза
func (RenderData *TRenderData) Clear() {
	RenderData.Alerts = nil
	RenderData.Data = nil
	RenderData.DataMap = nil
	RenderData.Dict = nil
}

// Добавляет уведомление в срез
func (RenderData *TRenderData) AddAlert(text, class string) {
	RenderData.Alerts = append(RenderData.Alerts, TAlert{Text: text, Class: class})
}

// Добавляет уведомление типа "danger" в срез, если есть err не nil и возвращает ту же ошибку для пост обработки
func (RenderData *TRenderData) AddAlertIfErr(err error) error {
	if err != nil {
		RenderData.AddAlert(err.Error(), "dander")
	}
	return err
}

// Добавляет уведомление типа "success" в срез, если срез пустой
func (RenderData *TRenderData) AddAlertIfOk(text string) {
	if len(RenderData.Alerts) == 0 {
		RenderData.AddAlert(text, "success")
	}
}

// Получение всех данных (см. fetch) из таблицы в карту данных для рендера
func (RenderData *TRenderData) GetAllFromTable(table *TTable, fetch int) {
	err := table.DFReadAll(fetch)
	if err != nil {
		RenderData.DataMap = nil
		RenderData.AddAlert(err.Error(), "danger")
	} else {
		RenderData.DataMap = table.DataFrame.Maps()
	}
}

// Получение одной строки по ID из таблицы в карту данных для рендера
func (RenderData *TRenderData) GetOneFromTable(table *TTable, id int) {
	err := table.DFRead(id)
	if err != nil {
		RenderData.DataMap = nil
		RenderData.AddAlert(err.Error(), "danger")
	} else {
		RenderData.DataMap = table.DataFrame.Maps()
	}
}

// Заполняет словарь dictName значениями таблицы table
func (RenderData *TRenderData) AddDictFromTable(dictName string, table *TTable) {
	if RenderData.Dict == nil {
		RenderData.Dict = make(map[string]interface{})
	}
	err := table.DFReadAll(0)
	if err != nil {
		RenderData.Dict[dictName] = nil
		RenderData.AddAlert(err.Error(), "danger")
	} else {
		RenderData.Dict[dictName] = table.DataFrame.Maps()
	}
}

// Заполняет словарь dictName значениями типа table.DataFrame.Maps
func (RenderData *TRenderData) AddDictFromMaps(dictName string, maps []map[string]interface{}) {
	if RenderData.Dict == nil {
		RenderData.Dict = make(map[string]interface{})
	}
	RenderData.Dict[dictName] = maps

}

// Заполняет словарь dictName значениями типа table.DataFrame.Maps
func (RenderData *TRenderData) AddDictFromArr(dictName string, strArr []string) {
	if RenderData.Dict == nil {
		RenderData.Dict = make(map[string]interface{})
	}
	RenderData.Dict[dictName] = strArr

}
