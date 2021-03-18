package model

type CustomerModel struct {
	Name      string  `json:"name"`
	ValueType string  `json:"value_type"`
	Fields    []Field `json:"fields"`
}

type Field struct {
	Name      string `json:"name"`
	TypeValue string `json:"type"`
	Desc      string `json:"desc"`
	Required  bool   `json:"required"`
}

type SvrOptions struct {
	// mod 路径
	ModulePath string
	// 项目路径
	ProjectPath string
	// 保存路径
	SaveTo string
	// 服务语言
	Language string
	// 框架
	Lib string
	// 持久化
	DataPersistence string
}
