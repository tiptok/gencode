package dm_api

import (
	"bytes"
	"fmt"
	"github.com/tiptok/gencode/cmd/dddgen/api"
	"github.com/tiptok/gencode/cmd/dddgen/dm"
	"github.com/tiptok/gencode/common"
	"github.com/tiptok/gencode/constant"
	"github.com/tiptok/gencode/model"
	"github.com/urfave/cli"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func DmApiRun(ctx *cli.Context) {
	var (
		o       dm.DMOptions = dm.DMOptions{}
		results              = make(chan *api.GenResult, 100)
	)
	o.ProjectPath = ctx.String("p")
	o.SaveTo = ctx.String("st")
	o.Lib = ctx.String("lib")
	o.Language = ctx.String("lang")
	o.ModulePath = common.GoModuleName(o.SaveTo)
	readPath := o.ProjectPath
	if !strings.Contains(readPath, "domain-model") {
		readPath = filepath.Join(o.ProjectPath, "domain-model")
	}
	dms := dm.ReadDomainModels(readPath)
	if len(dms) == 0 {
		log.Println("domain-model not found")
		return
	}
	controllers, err := dmsToApiControllers(dms)
	if err != nil {
		log.Println(err)
	}
	var genServer api.ServeGen = GoBeeDomainApiServeGen{}

	api.GenApiServer(genServer, o.SvrOptions, controllers, results)
}

func dmsToApiControllers(dms []dm.DomainModel) (controllers []api.Controller, err error) {
	for _, dm := range dms {
		if !dm.NeedRestful() {
			continue
		}
		newCtr := api.Controller{
			Controller: dm.Name,
			Paths:      make([]api.ApiPath, 0),
		}
		newCtr.Paths = append(newCtr.Paths, []api.ApiPath{
			getApi(dm, "Create", http.MethodPost),
			getApi(dm, "Update", http.MethodPut),
			getApi(dm, "Get", http.MethodGet),
			getApi(dm, "Delete", http.MethodDelete),
			getApi(dm, "List", http.MethodGet),
		}...)
		controllers = append(controllers, newCtr)
	}
	return
}

func getApi(dm dm.DomainModel, prefix, method string) api.ApiPath {
	function := prefix + dm.Name
	lowercase := common.LowCasePaddingUnderline(dm.Name)
	lowerFirst := common.LowFirstCase(dm.Name)
	apiPath := api.ApiPath{
		Path:    fmt.Sprintf("/%v/:%vId", lowercase, lowerFirst),
		Method:  method,
		Content: "json",
	}
	apiPath.ServiceName = function
	apiPath.Operator = string(constant.COMMAND)
	switch strings.ToUpper(method) {
	case http.MethodPost:
		apiPath.Path = fmt.Sprintf("/%v/", lowercase)
	case http.MethodPut:
		apiPath.Path = fmt.Sprintf("/%v/:%vId", lowercase, lowerFirst)
	case http.MethodDelete:
		apiPath.Path = fmt.Sprintf("/%v/:%vId", lowercase, lowerFirst)
	case http.MethodGet:
		apiPath.Operator = string(constant.QUERY)
		apiPath.Path = fmt.Sprintf("/%v/:%vId", lowercase, lowerFirst)
	}
	if prefix == "List" {
		apiPath.Path = fmt.Sprintf("/%v/", lowercase)
	}
	apiPath.Summary = fmt.Sprintf("%v execute %v  %v  %v", function, apiPath.Operator, strings.ToLower(prefix), dm.Name)
	apiPath.Request = api.RefObject{RefPath: function + "Request"}
	apiPath.Response = api.RefObject{RefPath: function + "Response"}
	return apiPath
}

type GoBeeDomainApiServeGen struct {
	api.GoBeeApiServeGen
}

func (g GoBeeDomainApiServeGen) GenApplication(c api.Controller, options model.SvrOptions, result chan<- *api.GenResult) error {
	buf := bytes.NewBuffer(nil)
	bufMethods := bytes.NewBuffer(nil)
	for i := 0; i < len(c.Paths); i++ {
		bufMethods.WriteString("\n\n")
		p := c.Paths[i]
		pName, _, _ := p.ParsePath()
		//log.Println(pName,req,rsp)
		if err := common.ExecuteTmpl(bufMethods, applicationMethod, map[string]interface{}{
			"Method":  common.UpperFirstCase(pName),
			"Service": c.Controller,
			"Logic":   g.getApplicationLogic(c, p, options),
		}); err != nil {
			return err
		}
	}

	if err := common.ExecuteTmpl(buf, application, map[string]interface{}{
		"Package": common.LowCasePaddingUnderline(c.Controller),
		"Module":  options.ModulePath,
		"Service": c.Controller,
		"Methods": bufMethods.String(),
	}); err != nil {
		return err
	}

	result <- &api.GenResult{
		Root:        options.SaveTo,
		SaveTo:      constant.WithApplication(common.LowCasePaddingUnderline(c.Controller)),
		FileName:    common.LowCasePaddingUnderline(c.Controller) + ".go",
		FileData:    buf.Bytes(),
		JumpExisted: true,
	}
	return nil
}

func (g GoBeeDomainApiServeGen) GenProtocol(c api.Controller, options model.SvrOptions, result chan<- *api.GenResult) error {
	domainPath := filepath.Join(options.ProjectPath, "domain-model", common.LowCasePaddingUnderline(c.Controller)+".json")
	models := dm.ReadDomainModels(domainPath)
	if len(models) == 0 {
		return nil
	}
	domainModel := models[0]

	for i := 0; i < len(c.Paths); i++ {

		p := c.Paths[i]

		parseModel := func(refPath string) error {
			buf := bytes.NewBuffer(nil)
			bufFields := bytes.NewBuffer(nil)
			ref := refPath
			arrays := strings.Split(ref, "/")
			modelName := arrays[len(arrays)-1]
			fields := g.getProtocolField(c, p, options, domainModel, ref)
			for i, field := range fields {
				if strings.HasPrefix(p.ServiceName, "List") {
					continue
				}
				if err := common.ExecuteTmpl(bufFields, api.ProtocolField, map[string]interface{}{
					"Desc":   field.Desc,
					"Column": field.Name,
					"Type":   field.TypeValue,
					"Tags":   fmt.Sprintf("`json:\"%v,omitempty\"`", common.LowFirstCase(field.Name)),
				}); err != nil {
					return err
				}
				if i != (len(fields) - 1) {
					bufFields.WriteString("\n")
				}
			}

			if err := common.ExecuteTmpl(buf, api.ProtocolModel, map[string]interface{}{
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
			result <- &api.GenResult{
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
	return nil
}

func (g GoBeeDomainApiServeGen) getApplicationLogic(c api.Controller, path api.ApiPath, options model.SvrOptions) string {
	buf := bytes.NewBuffer(nil)
	// TODO 单个文件
	domainPath := filepath.Join(options.ProjectPath, "domain-model", common.LowCasePaddingUnderline(c.Controller)+".json")
	models := dm.ReadDomainModels(domainPath)
	if len(models) == 0 {
		return ""
	}
	domainModel := models[0]
	if strings.HasPrefix(path.ServiceName, "Create") {
		buf.WriteString(fmt.Sprintf("	new%v:=&domain.%v{\n", c.Controller, c.Controller))
		for _, field := range domainModel.Fields {
			if equalAnyInArray(field.Name, "Id", c.Controller+"Id") {
				continue
			}
			if containAnyInArray(field.Name, "At", "Time") && field.TypeValue == "time.Time" {
				buf.WriteString(fmt.Sprintf("		%v: time.Now(),\n", field.Name))
				continue
			}
			buf.WriteString(fmt.Sprintf("		%v: request.%v,\n", field.Name, field.Name))
		}
		buf.WriteString("	}\n")
		buf.WriteString(strings.ReplaceAll(`	
    var DomainRepository,_ = factory.CreateDomainRepository(transactionContext)
	if m,err:=DomainRepository.Save(newDomain);err!=nil{
		return nil,err
	}else{
		rsp = m
	}`, "Domain", c.Controller))
		return buf.String()
	}
	if strings.HasPrefix(path.ServiceName, "Update") {
		common.ExecuteTmpl(buf, `	
    var {{.Domain}}Repository,_ = factory.Create{{.Domain}}Repository(transactionContext)
	var {{.domain}} *domain.{{.Domain}}
	if {{.domain}},err={{.Domain}}Repository.FindOne(map[string]interface{}{"id":request.Id});err!=nil{
		return
	}
	if err ={{.domain}}.Update(common.ObjectToMap(request));err!=nil{
		return
	}
	if {{.domain}},err = {{.Domain}}Repository.Save({{.domain}});err!=nil{
		return
	}`, map[string]interface{}{"Domain": c.Controller, "domain": common.LowFirstCase(c.Controller)})
		return buf.String()
	}
	if strings.HasPrefix(path.ServiceName, "Get") {
		common.ExecuteTmpl(buf, `	
    var {{.Domain}}Repository,_ = factory.Create{{.Domain}}Repository(transactionContext)
	var {{.domain}} *domain.{{.Domain}}
	if {{.domain}},err={{.Domain}}Repository.FindOne(common.ObjectToMap(request));err!=nil{
		return
	}
	rsp = {{.domain}}`, map[string]interface{}{"Domain": c.Controller, "domain": common.LowFirstCase(c.Controller)})
		return buf.String()
	}
	if strings.HasPrefix(path.ServiceName, "Delete") {
		common.ExecuteTmpl(buf, `	
    var {{.Domain}}Repository,_ = factory.Create{{.Domain}}Repository(transactionContext)
	var {{.domain}} *domain.{{.Domain}}
	if {{.domain}},err={{.Domain}}Repository.FindOne(common.ObjectToMap(request));err!=nil{
		return
	}
	if {{.domain}},err = {{.Domain}}Repository.Remove({{.domain}});err!=nil{
		return 
	}
	rsp = {{.domain}}`, map[string]interface{}{"Domain": c.Controller, "domain": common.LowFirstCase(c.Controller)})
		return buf.String()
	}
	if strings.HasPrefix(path.ServiceName, "List") {
		common.ExecuteTmpl(buf, `
	var {{.Domain}}Repository,_ = factory.Create{{.Domain}}Repository(transactionContext)
	var {{.domain}} []*domain.{{.Domain}}
	var total int64
	if total,{{.domain}},err={{.Domain}}Repository.Find(common.ObjectToMap(request));err!=nil{
		return
	}
	rsp =map[string]interface{}{
		"total":total,
		"list":{{.domain}},
	}`, map[string]interface{}{"Domain": c.Controller, "domain": common.LowFirstCase(c.Controller)})
		return buf.String()
	}
	return buf.String()
}

func (g GoBeeDomainApiServeGen) getProtocolField(c api.Controller, path api.ApiPath, options model.SvrOptions, domainModel dm.DomainModel, ref string) []*model.Field {
	rsp := make([]*model.Field, 0)
	if strings.HasSuffix(ref, "Response") {
		return rsp
	}
	newField := func(name, t, desc string) *model.Field {
		return &model.Field{
			Name:      name,
			TypeValue: t,
			Desc:      desc,
		}
	}
	if strings.HasPrefix(path.ServiceName, "Create") {
		for _, f := range domainModel.Fields {
			if containAnyInArray(f.Name, "Id", domainModel.Name+"Id", "At", "Time") {
				continue
			}
			rsp = append(rsp, newField(f.Name, f.TypeValue, f.Desc))
		}
		return rsp
	}
	if strings.HasPrefix(path.ServiceName, "Update") {
		for _, f := range domainModel.Fields {
			//if equalAnyInArray(f.Name, "Id", domainModel.Name+"Id") {
			//	continue
			//}
			if containAnyInArray(f.Name, "At", "Time") {
				continue
			}
			rsp = append(rsp, newField(f.Name, f.TypeValue, f.Desc))
		}
		return rsp
	}
	if strings.HasPrefix(path.ServiceName, "Get") {
		for _, f := range domainModel.Fields {
			if equalAnyInArray(f.Name, "Id", domainModel.Name+"Id") {
				rsp = append(rsp, newField(f.Name, f.TypeValue, f.Desc))
				continue
			}
		}
		return rsp
	}
	if strings.HasPrefix(path.ServiceName, "Delete") {
		for _, f := range domainModel.Fields {
			if equalAnyInArray(f.Name, "Id", domainModel.Name+"Id") {
				rsp = append(rsp, newField(f.Name, f.TypeValue, f.Desc))
				continue
			}
		}
		return rsp
	}
	if strings.HasPrefix(path.ServiceName, "List") {
		rsp = append(rsp, newField("offset", "int", "偏移位置（分页查询）"))
		rsp = append(rsp, newField("limit", "int", "限制数量（分页查询）"))
	}
	return rsp
}

func containAnyInArray(c string, array ...string) bool {
	for i := range array {
		if strings.Contains(c, array[i]) {
			return true
		}
	}
	return false
}

func equalAnyInArray(c string, array ...string) bool {
	for i := range array {
		if strings.EqualFold(c, array[i]) {
			return true
		}
	}
	return false
}
