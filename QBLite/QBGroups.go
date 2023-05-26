package main

import (
	"korsakov2ea/xfunc"
	"log"
	"net/http"
	"strconv"
)

// список групп запросов
func groups(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v Переход к списку групп ────────────────────────────────────────┐", xfunc.FuncName())
	defer log.Printf("%v Переход к списку групп ────────────────────────────────────────┘", xfunc.FuncName())

	sqlErr = QBGroup.ReadAll(0)
	if sqlErr != nil {
		RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
	}
	RenderData.Data = QBGroup.Data
	RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Всего групп - " + strconv.Itoa(len(RenderData.Data)), Class: "info"})
	xfunc.RenderPage(w, "groups.html", "common.html", RenderData)
	RenderData.Alerts = nil
}

// группа запроса
func group(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("ID"))
	if err != nil && r.Method == "POST" && (r.FormValue("submitBtn") == "Update" || r.FormValue("submitBtn") == "Delete") {
		log.Printf("%v Ошибка преобразования ID = %v из GET запроса в число", xfunc.FuncName(), r.FormValue("ID"))
	} else {
		log.Printf("%v HTTP запрос с параметрами %v", xfunc.FuncName(), r.URL.RawQuery)

		switch {

		// Отмена изменений группы
		case r.Method == "POST" && r.FormValue("submitBtn") == "Cancel":
			log.Printf("%v Отмена изменений группы ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil
			RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Нажата кнопка [Отмена]", Class: "info"})
			http.Redirect(w, r, "groups", http.StatusFound)
			log.Printf("%v Отмена изменений группы ────────────────────────────────────────┘", xfunc.FuncName())

		// Создание группы
		case r.Method == "POST" && r.FormValue("submitBtn") == "Create":
			log.Printf("%v Создание группы ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil
			newGroup := make(map[string]string)
			newGroup["NAME"] = r.FormValue("Name")

			sqlErr = QBGroup.Create(newGroup)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			} else {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Группа создана", Class: "success"})
			}

			http.Redirect(w, r, "groups", http.StatusFound)
			log.Printf("%v Создание группы ────────────────────────────────────────┘", xfunc.FuncName())

		// Изменение группы
		case r.Method == "POST" && r.FormValue("submitBtn") == "Update":
			log.Printf("%v Изменение группы ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil
			newGroup := make(map[string]string)
			newGroup["NAME"] = r.FormValue("Name")

			sqlErr = QBGroup.Update(id, newGroup)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			} else {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Группа изменена", Class: "success"})
			}

			http.Redirect(w, r, "groups", http.StatusFound)
			log.Printf("%v Изменение группы ────────────────────────────────────────┘", xfunc.FuncName())

		// Удаление группы
		case r.Method == "POST" && r.FormValue("submitBtn") == "Delete":
			log.Printf("%v Удаление группы ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil

			sqlErr = QBGroup.Delete(id)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			} else {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: "Группа удалена", Class: "success"})
			}

			http.Redirect(w, r, "groups", http.StatusFound)
			log.Printf("%v Удаление группы ────────────────────────────────────────┘", xfunc.FuncName())

		// Переход к добавления группы
		case r.Method == "GET" && r.FormValue("mode") == "add":
			log.Printf("%v Переход к добавления группы ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil
			RenderData.Data = nil
			xfunc.RenderPage(w, "group.html", "common.html", RenderData)
			log.Printf("%v Переход к добавления группы ────────────────────────────────────────┘", xfunc.FuncName())

		// Переход к изменению группы
		case r.Method == "GET" && r.FormValue("mode") == "edit":
			log.Printf("%v Переход к изменению группы ────────────────────────────────────────┐", xfunc.FuncName())
			RenderData.Alerts = nil

			sqlErr = QBGroup.Read(id)
			if sqlErr != nil {
				RenderData.Alerts = append(RenderData.Alerts, xfunc.TAlert{Text: sqlErr.Error(), Class: "danger"})
			}
			RenderData.Data = QBGroup.Data

			xfunc.RenderPage(w, "group.html", "common.html", RenderData)
			log.Printf("%v Переход к изменению группы ────────────────────────────────────────┘", xfunc.FuncName())

		default:
			// сделать заглушку если неизветный путь
		}
	}
}
