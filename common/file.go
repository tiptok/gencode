package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

func CreateIfNotExist(file string) (*os.File, error) {
	_, err := os.Stat(file)
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("%s already exist", file)
	}

	return os.Create(file)
}

func FileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

//保存文件
func SaveTo(root string, st string, filename string, data []byte) (err error) {
	filePath := filepath.Join(root, st)
	if _, e := os.Stat(filePath); e != nil {
		//log.Println(filePath,e)
		if err = os.MkdirAll(filePath, 0777); err != nil {
			return
		}
	}
	log.Println("【gen code】", "path:", filePath, "file:", filename)
	return ioutil.WriteFile(filepath.Join(filePath, filename), data, os.ModePerm)
}

func SaveJsonTo(root string, st string, filename string, obj interface{}, recover bool) (err error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	if FileExists(filepath.Join(root, st, filename)) && !recover {
		return
	}
	return SaveTo(root, st, filename, data)
}

// 从描述文件里面读取模型
func ReadModelFromJsonFile(path string, value interface{}) error {
	if !FileExists(path) {
		fmt.Println("not exits:", path)
		return nil
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(path, err)
		return err
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	if err := json.Unmarshal(data, value); err != nil {
		log.Println(path, err)
		return err
	}
	return nil
}

// 解析模板
func ExecuteTmpl(writer io.Writer, tmpl string, paramMap map[string]interface{}) error {
	tP, err := template.New("controller").Parse(tmpl)
	if err != nil {
		return err
	}
	return tP.Execute(writer, paramMap)
}
