package pkg_control

import (
	"github.com/tiptok/gencode/cmd/dddgen/api"
	"github.com/tiptok/gencode/common"
	"github.com/tiptok/gencode/model"
	"github.com/urfave/cli"
	"strings"
)

func RunPackageControl(ctx *cli.Context) {
	var (
		o       = model.SvrOptions{}
		results = make(chan *api.GenResult, 100)
	)
	o.ProjectPath = ctx.String("p") //项目文件根目录
	o.SaveTo = ctx.String("st")
	o.Language = ctx.String("lang")
	o.Lib = ctx.String("lib")

	opType := ctx.String("t")
	o.ModulePath = common.GoModuleName(o.SaveTo)

	switch strings.ToLower(opType) {
	case "add":
		go AddPackage(o, results)
	}

	api.FileGen(results)
}

func AddPackage(o model.SvrOptions, result chan *api.GenResult) {
	switch o.Lib {
	case "redis":
		result <- &api.GenResult{
			Root:        o.SaveTo,
			SaveTo:      `/pkg/constant`,
			FileName:    common.LowCasePaddingUnderline("redis") + ".go",
			FileData:    []byte(tmplConstantRedis),
			JumpExisted: true,
		}
		result <- &api.GenResult{
			Root:        o.SaveTo,
			SaveTo:      `/pkg/infrastructure/redis`,
			FileName:    common.LowCasePaddingUnderline("init") + ".go",
			FileData:    []byte(tmplRedisInit),
			JumpExisted: true,
		}
		break
	}
	close(result)
}
