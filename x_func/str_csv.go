package x_func

import (
	"encoding/csv"
	"log"
	"mime/multipart"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

// GetStrMapFromCSVWebFile - возвращает двумерный массив string из CSV файла fileName, считанного с формы
func GetStrMapFromCSVWebFile(file multipart.File) [][]string {
	csvReader := csv.NewReader(file)
	csvReader.Comma = ';'
	csvRecords, err := csvReader.ReadAll()
	if err != nil {
		log.Printf("%v Ошибка получения записей из CSV файла", FuncName())
	}
	return csvRecords
}

// DecodeStrMap1251toUTF8 - возвращает двумерный массив string перекодированый из 1251 в UTF-8
func DecodeStrMap1251toUTF8(records1251 [][]string) (recordsUTF [][]string) {
	recordsUTF = records1251
	maxRow := len(records1251)
	maxCol := len(records1251[0])
	decoder := charmap.Windows1251.NewDecoder()
	for i := 0; i < maxRow; i++ {
		for j := 0; j < maxCol; j++ {
			RecordUTF, err := decoder.String(records1251[i][j])
			if err != nil {
				log.Printf("%v Ошибка перекодирования значения 1251 > UTF-8 %v", FuncName(), err)
			} else {
				recordsUTF[i][j] = RecordUTF
			}
		}
	}
	return
}

// DecodeStr1251toUTF8 - возвращает строку перекодированую из 1251 в UTF-8
func DecodeStr1251toUTF8(W1251 string) string {
	decoder := charmap.Windows1251.NewDecoder()
	UTF8, err := decoder.String(W1251)
	if err != nil {
		log.Printf("%v Ошибка декодирования строки %v", FuncName(), err)
	}
	return UTF8
}

// DecodeStr1251toUTF8 - возвращает строку перекодированую из UTF-8 в 1251
func DecodeStrUTF8to1251(UTF8 string) string {
	encoder := charmap.Windows1251.NewEncoder()
	W1251, err := encoder.String(UTF8)
	if err != nil {
		log.Printf("%v Ошибка декодирования строки %v", FuncName(), err)
	}
	return W1251
}

// GetSizeStrMap - возвращает размер карты строки
func GetSizeStrMap(strMap [][]string) (rows int, cols int) {
	rows = len(strMap)
	cols = len(strMap[0])
	return rows, cols
}

// GetSizeStrMap - строки вида (COL1, СOL2... COLn) и (COL1 SQLType, СOL2 SQLType ... COLn SQLType)
func GetCOLStrMap(strMap [][]string, SQLType string) (COLs, COLTypes string) {
	COLs = ""
	for i := range strMap[0] {
		COLs = COLs + ", COL" + strconv.Itoa(i+1)
	}
	COLs = COLs[2:]
	COLTypes = strings.Replace(COLs, ",", " VARCHAR(255),", -1) + " VARCHAR(255)"
	return COLs, COLTypes
}
