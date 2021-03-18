package common

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//单词首字母小写
func LowFirstCase(input string) string {
	array := []byte(input)
	if len(array) == 0 {
		return ""
	}
	rspArray := make([]byte, len(array))
	copy(rspArray[:1], strings.ToLower(string(array[:1])))
	copy(rspArray[1:], array[1:])
	return string(rspArray)
}

//首字母小写 多个字母用下划线拼接
// LowCase  low_case
func LowCasePaddingUnderline(input string) string {
	data := []byte(input)
	var rspData []byte
	for i := range data {
		c := data[i]
		if c >= byte('A') && c <= byte('Z') {
			if i != 0 { //首字母除外
				rspData = append(rspData, byte('_'))
			}
			rspData = append(rspData, c+32)
			continue
		}
		rspData = append(rspData, c)
	}
	return string(rspData)
}

func UpperFirstCase(input string) string {
	if len(input) == 0 {
		return input
	}
	data := []byte(input)
	var rspData []byte
	for i := range data {
		c := data[i]
		if i == 0 && c >= byte('a') && c <= byte('z') {
			if i != 0 { //首字母除外
				rspData = append(rspData, byte('_'))
			}
			rspData = append(rspData, c-32)
			continue
		}
		rspData = append(rspData, data[i:]...)
		break
	}
	return string(rspData)
}

func GoModuleName(work string) string {
	path := filepath.Join(work, "go.mod")
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("mod file not exists:", path, err)
		return ""
	}
	r := bufio.NewReader(file)
	line, err := r.ReadBytes('\n')
	if err != nil {
		fmt.Println("mod file read error:", err)
		return ""
	}
	module := strings.Split(string(line), " ")[1]
	return strings.Trim(module, "\n")
}
