package api

import (
	"bytes"
	"fmt"
	"github.com/tiptok/gencode/common"
	"github.com/tiptok/gencode/constant"
	"github.com/tiptok/gencode/model"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// 通过api dsl描述语言 生成对应的api服务
func RunApiSever(ctx *cli.Context) {
	var (
		o        = ApiSvrOptions{}
		results  = make(chan *GenResult, 100)
		serveGen = serveGenFactory()
	)
	o.ProjectPath = ctx.String("p") //项目文件根目录
	o.SaveTo = ctx.String("st")
	o.Language = ctx.String("lang")
	o.Lib = ctx.String("lib")
	o.ModulePath = common.GoModuleName(o.SaveTo)

	if _, ok := o.Valid(); !ok {
		return
	}
	controllers, err := ReadApiModels(o.ProjectPath)
	if err != nil {
		fmt.Println("read api models err:", err)
		return
	}
	GenApiServer(serveGen, o.SvrOptions, controllers, results)
}

func GenApiServer(serveGen ServeGen, o model.SvrOptions, controllers []Controller, results chan *GenResult) {
	genFactoryTransaction(o, results)
	genDefaultByLib(o, results)
	go func() {
		for i := 0; i < len(controllers); i++ {
			c := controllers[i]
			if err := serveGen.GenController(c, o, results); err != nil {
				fmt.Println("gen controller error:", err)
				return
			}
			if err := serveGen.GenRouter(c, o, results); err != nil {
				fmt.Println("gen router error:", err)
				return
			}
			if err := serveGen.GenApplication(c, o, results); err != nil {
				fmt.Println("gen application error:", err)
				return
			}
			if err := serveGen.GenProtocol(c, o, results); err != nil {
				fmt.Println("gen protocol error:", err)
				return
			}
		}
		close(results)
	}()
	var done sync.WaitGroup
	done.Add(1)
	go func() {
		for result := range results {
			filePath := filepath.Join(result.Root, result.SaveTo, result.FileName)
			if common.FileExists(filePath) && result.JumpExisted {
				log.Println("【gen code】 jump:", filePath)
				continue
			}

			err := common.SaveTo(result.Root, result.SaveTo, result.FileName, result.FileData)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		done.Done()
	}()
	done.Wait()
}

func genDefaultByLib(options model.SvrOptions, result chan<- *GenResult) {
	switch options.Lib {
	case "beego":
		if err := genBeegoRouterInit(options, result); err != nil {
			fmt.Println(err)
		}
		if err := genBeegoMain(options, result); err != nil {
			fmt.Println(err)
		}
		break
	}
}
func genDefaultByPersistence(options model.SvrOptions, result chan<- *GenResult) {

}

// gen /application/transaction  by persistence
func genFactoryTransaction(options model.SvrOptions, result chan<- *GenResult) (err error) {
	buf := bytes.NewBuffer(nil)

	if err := common.ExecuteTmpl(buf, pgFactoryTransaction, map[string]interface{}{
		"Package": "transaction",
		"Module":  options.ModulePath,
	}); err != nil {
		return err
	}
	result <- &GenResult{
		Root:        options.SaveTo,
		SaveTo:      constant.WithApplication("factory"),
		FileName:    common.LowCasePaddingUnderline("transaction") + ".go",
		FileData:    buf.Bytes(),
		JumpExisted: true,
	}
	return nil
}

// gen /pkg/port/beego/router  by lib
func genBeegoRouterInit(options model.SvrOptions, result chan<- *GenResult) (err error) {
	buf := bytes.NewBuffer(nil)

	if err := common.ExecuteTmpl(buf, beegoRouterInit, map[string]interface{}{
		"Module": options.ModulePath,
	}); err != nil {
		return err
	}
	result <- &GenResult{
		Root:        options.SaveTo,
		SaveTo:      constant.WithPort(options.Lib),
		FileName:    common.LowCasePaddingUnderline(options.Lib) + ".go",
		FileData:    buf.Bytes(),
		JumpExisted: true,
	}
	return nil
}

// gen main.go by lib
func genBeegoMain(options model.SvrOptions, result chan<- *GenResult) (err error) {
	buf := bytes.NewBuffer(nil)

	if err := common.ExecuteTmpl(buf, beegoMain, map[string]interface{}{
		"Module": options.ModulePath,
	}); err != nil {
		return err
	}
	result <- &GenResult{
		Root:        options.SaveTo,
		SaveTo:      "",
		FileName:    "main.go",
		FileData:    buf.Bytes(),
		JumpExisted: true,
	}
	return nil
}

func ReadApiModels(p string) (controllers []Controller, err error) {
	var f os.FileInfo
	if f, err = os.Stat(p); err != nil {
		return
	}
	readOne := func(file string) error {
		c := Controller{}
		if err := common.ReadModelFromJsonFile(file, &c); err != nil {
			return err
		}
		controllers = append(controllers, c)
		return nil
	}
	if !f.IsDir() {
		err = readOne(p)
		return
	}
	root := filepath.Join(p, constant.Api)
	files, e := ioutil.ReadDir(root)
	if e != nil {
		err = e
		return
	}
	for i := range files {
		fileItem := files[i]
		if fileItem.IsDir() {
			continue
		}
		if err = readOne(filepath.Join(root, fileItem.Name())); err != nil {
			return
		}
	}
	return
}

// serve生成器
func serveGenFactory() ServeGen {
	return GoBeeApiServeGen{}
}

type ServeGen interface {
	GenController(c Controller, options model.SvrOptions, result chan<- *GenResult) error
	GenRouter(c Controller, options model.SvrOptions, result chan<- *GenResult) error
	GenApplication(c Controller, options model.SvrOptions, result chan<- *GenResult) error
	GenProtocol(c Controller, options model.SvrOptions, result chan<- *GenResult) error
}

// golang beego 框架 serve生成器
type GoBeeApiServeGen struct{}

func (g GoBeeApiServeGen) GenController(c Controller, options model.SvrOptions, result chan<- *GenResult) error {
	buf := bytes.NewBuffer(nil)
	if err := common.ExecuteTmpl(buf, beegonController, map[string]interface{}{
		"Module":          options.ModulePath,
		"ControllerLower": common.LowCasePaddingUnderline(c.Controller),
		"Controller":      c.Controller,
	}); err != nil {
		return err
	}

	for i := 0; i < len(c.Paths); i++ {
		buf.WriteString("\n")
		p := c.Paths[i]
		pName, req, rsp := p.ParsePath()
		//log.Println(pName,req,rsp)
		if err := common.ExecuteTmpl(buf, beegoControllerMethod, map[string]interface{}{
			"Method":              common.UpperFirstCase(pName),
			"Desc":                p.Summary,
			"ControllerName":      c.Controller,
			"RequestModel":        req,
			"ResponseModel":       rsp,
			"Application":         common.LowCasePaddingUnderline(c.Controller),
			"HttpMethod":          strings.ToUpper(p.Method),
			"LowerControllerName": common.LowFirstCase(c.Controller),
		}); err != nil {
			return err
		}
	}

	result <- &GenResult{
		Root:        options.SaveTo,
		SaveTo:      constant.WithController(options.Lib),
		FileName:    common.LowCasePaddingUnderline(c.Controller) + ".go",
		FileData:    buf.Bytes(),
		JumpExisted: true,
	}

	baseBuf := bytes.NewBuffer(nil)
	if err := common.ExecuteTmpl(baseBuf, beegoBaseController, map[string]interface{}{
		"Module": options.ModulePath,
	}); err != nil {
		return err
	}
	result <- &GenResult{
		Root:        options.SaveTo,
		SaveTo:      constant.WithController(options.Lib),
		FileName:    "base.go",
		FileData:    baseBuf.Bytes(),
		JumpExisted: true,
	}
	return nil
}
func (g GoBeeApiServeGen) GenRouter(c Controller, options model.SvrOptions, result chan<- *GenResult) error {
	buf := bytes.NewBuffer(nil)
	bufRouters := bytes.NewBuffer(nil)
	for i := 0; i < len(c.Paths); i++ {
		p := c.Paths[i]
		pName, _, _ := p.ParsePath()
		//log.Println(pName,req,rsp)
		if err := common.ExecuteTmpl(bufRouters, beegoRouter, map[string]interface{}{
			"Url":        p.Path,
			"Controller": c.Controller + "Controller",
			"HttpMethod": p.Method,
			"Method":     common.UpperFirstCase(pName),
		}); err != nil {
			return err
		}
		if i != (len(c.Paths) - 1) {
			bufRouters.WriteString("\n")
		}
	}

	if err := common.ExecuteTmpl(buf, beegoRouters, map[string]interface{}{
		"Module":  options.ModulePath,
		"Routers": bufRouters.String(),
	}); err != nil {
		return err
	}

	result <- &GenResult{
		Root:        options.SaveTo,
		SaveTo:      constant.WithRouter(options.Lib),
		FileName:    common.LowCasePaddingUnderline(c.Controller) + "_router" + ".go",
		FileData:    buf.Bytes(),
		JumpExisted: true,
	}
	return nil
}
func (g GoBeeApiServeGen) GenApplication(c Controller, options model.SvrOptions, result chan<- *GenResult) error {
	buf := bytes.NewBuffer(nil)
	bufMethods := bytes.NewBuffer(nil)
	for i := 0; i < len(c.Paths); i++ {
		bufMethods.WriteString("\n\n")
		p := c.Paths[i]
		pName, _, _ := p.ParsePath()
		//log.Println(pName,req,rsp)
		if err := common.ExecuteTmpl(bufMethods, ApplicationMethod, map[string]interface{}{
			"Method":  common.UpperFirstCase(pName),
			"Service": c.Controller,
			"Logic":   "",
		}); err != nil {
			return err
		}
	}

	if err := common.ExecuteTmpl(buf, Application, map[string]interface{}{
		"Package": common.LowCasePaddingUnderline(c.Controller),
		"Module":  options.ModulePath,
		"Service": c.Controller,
		"Methods": bufMethods.String(),
	}); err != nil {
		return err
	}

	result <- &GenResult{
		Root:        options.SaveTo,
		SaveTo:      constant.WithApplication(common.LowCasePaddingUnderline(c.Controller)),
		FileName:    common.LowCasePaddingUnderline(c.Controller) + ".go",
		FileData:    buf.Bytes(),
		JumpExisted: true,
	}
	return nil
}
func (g GoBeeApiServeGen) GenProtocol(c Controller, options model.SvrOptions, result chan<- *GenResult) error {

	for i := 0; i < len(c.Paths); i++ {

		p := c.Paths[i]

		parseModel := func(refPath string) error {

			buf := bytes.NewBuffer(nil)

			bufFields := bytes.NewBuffer(nil)

			ref := refPath
			arrays := strings.Split(ref, "/")
			modelName := arrays[len(arrays)-1]
			m := model.CustomerModel{}
			err := common.ReadModelFromJsonFile(filepath.Join(options.ProjectPath, ref+".json"), &m)
			if err != nil {
				fmt.Println(err)
			}
			//fmt.Println(filepath.Join(options.ProjectPath,ref),m.Name,len(m.Fields))
			for i := range m.Fields {
				field := m.Fields[i]
				if err := common.ExecuteTmpl(bufFields, ProtocolField, map[string]interface{}{
					"Desc":   field.Desc,
					"Column": field.Name,
					"Type":   field.TypeValue,
					"Tags":   fmt.Sprintf("`json:\"%v\"`", common.LowFirstCase(field.Name)),
				}); err != nil {
					return err
				}
				if i != (len(m.Fields) - 1) {
					bufFields.WriteString("\n")
				}
			}

			if err := common.ExecuteTmpl(buf, ProtocolModel, map[string]interface{}{
				"Package": common.LowCasePaddingUnderline(c.Controller),
				"Model":   modelName,
				"Fields":  bufFields.String(),
			}); err != nil {
				return err
			}
			fileName := common.LowCasePaddingUnderline(modelName) + ".go"
			if len(p.Operator) > 0 {
				fileName = p.Operator + "_" + fileName
			}
			result <- &GenResult{
				Root:        options.SaveTo,
				SaveTo:      constant.WithProtocol(common.LowCasePaddingUnderline(c.Controller)),
				FileName:    fileName,
				FileData:    buf.Bytes(),
				JumpExisted: true,
			}
			return nil
		}

		if err := parseModel(p.Request.RefPath); err != nil {
			fmt.Println(err)
			return err
		}
		if err := parseModel(p.Response.RefPath); err != nil {
			fmt.Println(err)
			return err
		}
	}

	result <- &GenResult{
		Root:        options.SaveTo,
		SaveTo:      constant.ProtocolX,
		FileName:    "protocol.go",
		FileData:    []byte(protocolx),
		JumpExisted: true,
	}
	return nil
}

func FileGen(results chan *GenResult) {
	var done sync.WaitGroup
	done.Add(1)
	go func() {
		for result := range results {
			filePath := filepath.Join(result.Root, result.SaveTo, result.FileName)
			if common.FileExists(filePath) && result.JumpExisted {
				log.Println("【gen code】 jump:", filePath)
				continue
			}

			err := common.SaveTo(result.Root, result.SaveTo, result.FileName, result.FileData)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		done.Done()
	}()
	done.Wait()
}

type GenResult struct {
	Root        string
	SaveTo      string
	FileName    string
	FileData    []byte
	JumpExisted bool //true:已存在的不在生成 false:重新生成
}
type Operation struct {
	Url      ApiPath
	Request  model.CustomerModel
	Response model.CustomerModel
}

// 服务参数
type ApiSvrOptions struct {
	model.SvrOptions
}

func (o ApiSvrOptions) Valid() (error, bool) {
	return nil, true
}
