package main

import (
	"korsakov2ea/x_func"
	"log"
	"strconv"
)

// QBEntityMode - тип описания режима работы с QBEntity (CRUD)
type QBEntityMode int

const (
	open QBEntityMode = iota + 1 // EnumIndex = 1
	add                          // EnumIndex = 2
)

// QBEntityMode.String - возвращает текстовое значения режима QBEntityMode
func (m QBEntityMode) String() string {
	return [...]string{"open", "add"}[m-1]
}

// QBEntityMode.String - возвращает числовое значения режима QBEntityMode
func (m QBEntityMode) EnumIndex() int {
	return int(m)
}

// QBEntity - тип для описания сущностей QB вида "соединение", "запрос", "пользователь" и т.д.. Где
// name - наименование сущности (таблица в БД)
// mode - режим CRUD,
// id - идентификатор сущности в БД,
// data - набор (карта) данных полученный из БД
type QBEntity struct {
	name string
	mode QBEntityMode
	id   int
	data []map[string]string
}

// Create - добавляет в БД новую запись из карты entityMap
func (qbe *QBEntity) Create(entityMap map[string]string) {
	log.Printf("%v Создание %v", x_func.FuncName(), qbe.name)
	qbe.data = nil
	var cols string = ""
	var vals string = ""
	for name, val := range entityMap {
		cols = cols + name + ", "
		vals = vals + "'" + val + "', "
	}
	sqlCode := "INSERT INTO " + qbe.name + " (" + cols[:len(cols)-2] + ") VALUES (" + vals[:len(vals)-2] + ")"
	QCDB.DBExec(sqlCode)
}

// Read - считывает из БД запись по id
func (qbe *QBEntity) Read(id int) {
	log.Printf("%v Чтение %v для id = %v", x_func.FuncName(), qbe.name, id)
	sqlCode := "SELECT * FROM " + qbe.name + " WHERE ID=" + strconv.Itoa(id)
	qbe.data, _ = QCDB.DBQuery(sqlCode, false)
}

// Update - заменяет запись в БД по id значениями из карты entityMap
func (qbe *QBEntity) Update(id int, entityMap map[string]string) {
	log.Printf("%v Изменение %v для id = %v", x_func.FuncName(), qbe.name, id)
	qbe.data = nil
	var cols_vals string = ""
	for col, val := range entityMap {
		cols_vals = cols_vals + col + " = '" + val + "', "
	}
	sqlCode := "UPDATE " + qbe.name + " SET " + cols_vals[:len(cols_vals)-2] + " WHERE ID = " + strconv.Itoa(id)
	QCDB.DBExec(sqlCode)
}

// Delete - удаляет запись из БД по id
func (qbe *QBEntity) Delete(id int) {
	log.Printf("%v Удаление %v для id = %v", x_func.FuncName(), qbe.name, id)
	qbe.data = nil
	sqlCode := "DELETE FROM " + qbe.name + " WHERE ID = " + strconv.Itoa(id)
	QCDB.DBExec(sqlCode)
}

// ReadAll - считывает из БД все записи
func (qbe *QBEntity) ReadAll() {
	log.Printf("%v Чтение всех %v", x_func.FuncName(), qbe.name)
	sqlCode := "SELECT * FROM " + qbe.name
	qbe.data, _ = QCDB.DBQuery(sqlCode, false)
}

// ReadSQL - считывает из БД записи по SQL
func (qbe *QBEntity) ReadSQL(sqlCode string) {
	log.Printf("%v Чтение SQL для %v", x_func.FuncName(), qbe.name)
	qbe.data, _ = QCDB.DBQuery(sqlCode, false)
}
