package api

const protocolx = `package protocol

import (
	"encoding/json"
)

//CustomErrParse 解析自定义错误结构体
type CustomErrParse interface {
	ParseToMessage() *ResponseMessage
}

//ErrorMap 统一消息错误编码
type ErrorMap map[int]string

//Search 搜索错误描述
func (m ErrorMap) Search(code int) ErrorCode {
	if v, ok := m[code]; ok {
		return ErrorCode{
			Errno:  code,
			Errmsg: v,
		}
	}
	return ErrorCode{Errno: code, Errmsg: "错误码未定义"}
}

func NewMesage(code int) *ResponseMessage {
	return &ResponseMessage{
		ErrorCode: SearchErr(code),
		Data: struct {
		}{},
	}
}

var (
	_ CustomErrParse = new(ErrWithMessage)
	_ error          = new(ErrWithMessage)
)

//NewErrWithMessage 构建错误返回
//code:用于匹配统一消息错误编码 eRR:填充嵌套错误
func NewErrWithMessage(code int, eRR ...error) *ErrWithMessage {
	r := &ErrWithMessage{
		ErrorCode: SearchErr(code),
	}
	if len(eRR) > 0 {
		r.Err = eRR[0]
	}
	return r
}

//Error 实现接口error 中的方法
//将ErrorCode转为json数据，建议用于日志记录
func (e ErrWithMessage) Error() string {
	bt, _ := json.Marshal(e.ErrorCode)
	return string(bt)
}

//Unwrap 接口实现
func (e ErrWithMessage) Unwrap() error {
	return e.Err
}

//ParseToMessage 实现CustomErrParse的接口
func (e ErrWithMessage) ParseToMessage() *ResponseMessage {
	return &ResponseMessage{
		ErrorCode: e.ErrorCode,
		Data:      nil,
	}
}

func SearchErr(code int) ErrorCode {
	return errmessge.Search(code)
}
func NewResponseMessageData(data interface{}, err error) *ResponseMessage {
	var msg *ResponseMessage
	if err == nil {
		msg = NewMesage(0)
		msg.Data = data
		return msg
	}
	//log.Error("服务错误:" + err.Error())
	if x, ok := err.(CustomErrParse); ok {
		msg = x.ParseToMessage()
		msg.Data = data
		return msg
	}
	return NewMesage(1)
}

func NewResponseMessage(code int, err string) *ResponseMessage {
	return &ResponseMessage{
		ErrorCode: ErrorCode{
			Errno:  code,
			Errmsg: err,
		},
		Data: struct {
		}{},
	}
}

func BadRequestParam(code int) *ResponseMessage {
	return NewMesage(code)
}

func NewSuccessWithMessage(msg string) *ErrWithMessage {
	return &ErrWithMessage{
		Err:       nil,
		ErrorCode: ErrorCode{0, msg},
	}
}

func NewCustomMessage(code int, msg string) *ErrWithMessage {
	return &ErrWithMessage{
		Err:       nil,
		ErrorCode: ErrorCode{code, msg},
	}
}
` +
	`
//ErrorCode 统一错误结构
type ErrorCode struct {
	Errno  int    ` + "`" + `json:"code"` + "`\n" +
	"	Errmsg string `json:\"msg\"`" + "\n" + `
}

//ResponseMessage 统一返回消息结构体
type ResponseMessage struct {
	ErrorCode
	Data interface{} ` + "`json:\"data\"`" + `
}

//ErrWithMessage  自定义错误结构
type ErrWithMessage struct {
	Err error ` + "`json:\"-\"`" + `
	ErrorCode
}
` + `
var errmessge ErrorMap = map[int]string{
	0:   "成功",
	1:   "系统异常",
	2:   "参数错误",
	113: "签名验证失败",
}

type RequestHeader struct {
	UserId      int64 //UserId 唯一标识
    BodyKeys []string //键值
}

`
