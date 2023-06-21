package main

import (
	"korsakov2ea/xfunc"
	"log"
	"net/http"
	"strconv"
)

// Возвращает все группы запросов +
func groups(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v --------- HTTP METHOD: %v, PARAM: %v", xfunc.FuncName(), r.Method, r.URL.RawQuery)
	RenderData.GetAllFromTable(&QBGroup, 0)
	RenderData.RenderMap(w, "groups.html", "common.html")
	RenderData.Clear()
}

// Обработка событий при работе с группой +
func group(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v --------- HTTP METHOD: %v, PARAM: %v", xfunc.FuncName(), r.Method, r.URL.RawQuery)
	id, err := strconv.Atoi(r.FormValue("ID"))
	if err != nil && r.Method == "POST" && (r.FormValue("submitBtn") == "Update" || r.FormValue("submitBtn") == "Delete") {
		log.Printf("%v Ошибка преобразования ID = %v из GET запроса в число", xfunc.FuncName(), r.FormValue("ID"))
	} else {
		log.Printf("%v HTTP запрос с параметрами %v", xfunc.FuncName(), r.URL.RawQuery)
		RenderData.Clear()
		switch {

		// Отмена изменений группы +
		case r.Method == "POST" && r.FormValue("submitBtn") == "Cancel":
			RenderData.AddAlert("Нажата кнопка [Отмена]", "info")
			http.Redirect(w, r, "groups", http.StatusFound)

		// Создание / изменение группы +
		case r.Method == "POST" && (r.FormValue("submitBtn") == "Create" || r.FormValue("submitBtn") == "Update"):
			newGroup := make(map[string]string)
			newGroup["NAME"] = r.FormValue("Name")
			action := ""
			switch r.FormValue("submitBtn") {
			case "Create":
				RenderData.AddAlertIfErr(QBGroup.Create(newGroup))
				action = "создана"
			case "Update":
				RenderData.AddAlertIfErr(QBGroup.Update(id, newGroup))
				action = "изменена"
			}
			RenderData.AddAlertIfOk("Группа " + action)
			http.Redirect(w, r, "groups", http.StatusFound)

		// Удаление группы +
		case r.Method == "POST" && r.FormValue("submitBtn") == "Delete":
			RenderData.AddAlertIfErr(QBGroup.Delete(id))
			RenderData.AddAlertIfOk("Группа удалена")
			http.Redirect(w, r, "groups", http.StatusFound)

		// Переход к добавлению / изменению группы +
		case r.Method == "GET" && (r.FormValue("mode") == "add" || r.FormValue("mode") == "edit"):
			if r.FormValue("mode") == "edit" {
				RenderData.GetOneFromTable(&QBGroup, id)
			}
			RenderData.RenderMap(w, "group.html", "common.html")

		default:
			w.Write([]byte("Неожиданный запрос!"))
		}
	}
}
