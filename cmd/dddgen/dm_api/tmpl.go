package dm_api

const applicationMethod = `func(svr *{{.Service}}Service){{.Method}}(header *protocol.RequestHeader, request *protocolx.{{.Method}}Request) (rsp interface{}, err error) {
	var (
		transactionContext, _          = factory.CreateTransactionContext(nil)
	)
	rsp = &protocolx.{{.Method}}Response{}
	if err=request.ValidateCommand();err!=nil{
		err = protocol.NewCustomMessage(2,err.Error())
		return
	}
	if err = transactionContext.StartTransaction(); err != nil {
		log.Error(err)
		return nil, err
	}
	defer func() {
		transactionContext.RollbackTransaction()
	}()
{{.Logic}}	
	err = transactionContext.CommitTransaction()
	return
}`

const application = `package {{.Package}}

import (
	"github.com/tiptok/gocomm/common"
	"github.com/tiptok/gocomm/pkg/log"
	"{{.Module}}/pkg/domain"
	"{{.Module}}/pkg/protocol"
	"{{.Module}}/pkg/application/factory"
	protocolx "{{.Module}}/pkg/protocol/{{.Package}}"
)

type {{.Service}}Service struct {
	
}

{{.Methods}}


func New{{.Service}}Service(options map[string]interface{}) *{{.Service}}Service {
	svr := &{{.Service}}Service{}
	return svr
}
`
