package mvcgen

import (
	"bytes"
	"github.com/tiptok/gencode/common"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

func Run(ctx *cli.Context) {
	var (
		controller string = ctx.String("c")
		method     string = ctx.String("m")
	)
	tC, err := template.New("controller").Parse(tmplControllerMethod)
	if err != nil {
		log.Fatal(err)
	}
	//param  -c Auth -m Login
	//Controller Auth
	//ControllerLowcase auth
	//Method Login
	//MethodRequest LoginRequest
	m := make(map[string]string)
	m["Controller"] = controller
	m["ControllerLowcase"] = common.LowFirstCase(controller)
	m["Method"] = method
	m["MethodLowcase"] = common.LowFirstCase(method)
	buf := bytes.NewBuffer(nil)
	tC.Execute(buf, m)

	tP, err := template.New("protocol").Parse(tmplProtocolModel)
	tP.Execute(buf, m)

	tH, err := template.New("protocol").Parse(tmplHandler)
	tH.Execute(buf, m)

	tR, err := template.New("route").Parse(tmplRouter)
	tR.Execute(buf, m)
	//log.Println(buf.String())
	ioutil.WriteFile("gencode.out", buf.Bytes(), os.ModePerm)
}
