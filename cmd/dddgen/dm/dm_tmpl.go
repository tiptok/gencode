package dm

const tmplProtocolDomainModel = `package domain

// {{.Desc}}
type {{.Model}} struct {
{{.Items}}
}
{{if .IsDomainModel}}
type {{.Model}}Repository interface {
	Save(dm *{{.Model}}) (*{{.Model}}, error)
	Remove(dm *{{.Model}}) (*{{.Model}}, error)
	FindOne(queryOptions map[string]interface{}) (*{{.Model}}, error)
	Find(queryOptions map[string]interface{}) (int64, []*{{.Model}}, error)
}

func (m *{{.Model}}) Identify() interface{} {
	if m.Id == 0 {
		return nil
	}
	return m.Id
}

func (m *{{.Model}}) Update(data map[string]interface{}) error {
{{range .Fields}}	if v, ok := data["{{.Column}}"]; ok {
		m.{{.Name}} = v.({{.Type}})
	}
{{end}}
	return nil
}
{{end}}
`

const tmplProtocolDomainPgRepository = `package repository

import (
	"fmt"
	"{{.Module}}/pkg/domain"
	"{{.Module}}/pkg/constant"
	"{{.Module}}/pkg/infrastructure/pg/models"
	"{{.Module}}/pkg/infrastructure/pg/transaction"
	. "github.com/tiptok/gocomm/pkg/orm/pgx"
	"github.com/tiptok/gocomm/common"
	"github.com/tiptok/gocomm/pkg/cache"
)

var (
	cache{{.Model}}IdKey = func(id int64)string{
		return fmt.Sprintf("%v:cache:{{.Model}}:id:%v",{{.DBName}},id)
 		// 不需要执行缓存时,key设置为空
		// return ""
	}
)

type {{.Model}}Repository struct {
	*cache.CachedRepository
	transactionContext *transaction.TransactionContext
}

func (repository *{{.Model}}Repository) Save(dm *domain.{{.Model}}) (*domain.{{.Model}}, error) {
	var (
		err error
		m   = &models.{{.Model}}{}
		tx  = repository.transactionContext.PgTx
	)
	if err = common.GobModelTransform(m, dm); err != nil {
		return nil, err
	}
	if dm.Identify() == nil {
		if _,err = tx.Model(m).Insert(); err != nil {
			return nil, err
		}
		dm.Id = m.Id
		return dm, nil
	}
	queryFunc:=func()(interface{},error){
		return tx.Model(m).WherePK().Update()
	}
	if _, err = repository.Query(queryFunc,cache{{.Model}}IdKey(dm.Id)); err != nil {
		return nil, err
	}
	return dm, nil
}

func (repository *{{.Model}}Repository) Remove(dm *domain.{{.Model}}) (*domain.{{.Model}}, error) {
	var (
		tx          = repository.transactionContext.PgTx
		m = &models.{{.Model}}{Id: dm.Identify().(int64)}
	)
	queryFunc:=func()(interface{},error){
		return tx.Model(m).Where("id = ?", dm.Id).Delete()
	}
	if _,err:=repository.Query(queryFunc,cache{{.Model}}IdKey(dm.Id));err!=nil{
		return dm, err
	}
	return dm, nil
}

func (repository *{{.Model}}Repository) FindOne(queryOptions map[string]interface{}) (*domain.{{.Model}}, error) {
	tx := repository.transactionContext.PgDd
	m := new(models.{{.Model}})
    queryFunc:=func()(interface{},error){
		query := NewQuery(tx.Model(m), queryOptions)
		query.SetWhere("id = ?", "id")
		if err := query.First(); err != nil {
			return nil, fmt.Errorf("query row not found")
		}
		return m,nil
	}
	var options []cache.QueryOption
	if _,ok:=queryOptions["id"];!ok{
		options = append(options,cache.WithNoCacheFlag())
	}else {
		m.Id = queryOptions["id"].(int64)
	}
	if err:=repository.QueryCache(cache{{.Model}}IdKey(m.Id),m,queryFunc,options...);err!=nil{
		return nil, err
	}
	if m.Id == 0 {
		return nil, fmt.Errorf("query row not found")
	}
	return repository.transformPgModelToDomainModel(m)
}

func (repository *{{.Model}}Repository) Find(queryOptions map[string]interface{}) (int64, []*domain.{{.Model}}, error) {
	tx := repository.transactionContext.PgTx
	var mList []*models.{{.Model}}
	dmList := make([]*domain.{{.Model}}, 0)
	query := NewQuery(tx.Model(&mList), queryOptions).
		SetOrder("create_time", "sortByCreateTime").
		SetOrder("update_time", "sortByUpdateTime")
	var err error
	if query.AffectRow, err = query.SelectAndCount(); err != nil {
		return 0, dmList, err
	}
	for _, m := range mList {
		if {{.Model}}, err := repository.transformPgModelToDomainModel(m); err != nil {
			return 0, dmList, err
		} else {
			dmList = append(dmList, {{.Model}})
		}
	}
	return int64(query.AffectRow), dmList, nil
}

func (repository *{{.Model}}Repository) transformPgModelToDomainModel({{.Model}}Model *models.{{.Model}}) (*domain.{{.Model}}, error) {
	m := &domain.{{.Model}}{}
	err := common.GobModelTransform(m, {{.Model}}Model)
	return m, err
}

func New{{.Model}}Repository(transactionContext *transaction.TransactionContext) (*{{.Model}}Repository, error) {
	if transactionContext == nil {
		return nil,fmt.Errorf("transactionContext参数不能为nil")
	}
	return &{{.Model}}Repository{transactionContext: transactionContext, CachedRepository: cache.NewDefaultCachedRepository()}, nil
}
`

const tmplProtocolPgModel = `package models

// {{.Desc}}
type {{.Model}} struct {
{{.Items}}
}
`

const tmplConstantPg = `package constant

import "os"

var POSTGRESQL_DB_NAME = "postgres"
var POSTGRESQL_USER = "postgres"      
var POSTGRESQL_PASSWORD = "123456"  
var POSTGRESQL_HOST = "127.0.0.1"  
var POSTGRESQL_PORT = "5432"          
var DISABLE_CREATE_TABLE = false
var DISABLE_SQL_GENERATE_PRINT = false

func init() {
	if os.Getenv("POSTGRESQL_DB_NAME") != "" {
		POSTGRESQL_DB_NAME = os.Getenv("POSTGRESQL_DB_NAME")
	}
	if os.Getenv("POSTGRESQL_USER") != "" {
		POSTGRESQL_USER = os.Getenv("POSTGRESQL_USER")
	}
	if os.Getenv("POSTGRESQL_PASSWORD") != "" {
		POSTGRESQL_PASSWORD = os.Getenv("POSTGRESQL_PASSWORD")
	}
	if os.Getenv("POSTGRESQL_HOST") != "" {
		POSTGRESQL_HOST = os.Getenv("POSTGRESQL_HOST")
	}
	if os.Getenv("POSTGRESQL_PORT") != "" {
		POSTGRESQL_PORT = os.Getenv("POSTGRESQL_PORT")
	}
	if os.Getenv("DISABLE_CREATE_TABLE") != "" {
		DISABLE_CREATE_TABLE = true
	}
	if os.Getenv("DISABLE_SQL_GENERATE_PRINT") != "" {
		DISABLE_SQL_GENERATE_PRINT = true
	}
}
`

const tmplPgInit = `package pg

import (
	"context"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"{{.Module}}/pkg/constant"
	"{{.Module}}/pkg/infrastructure/pg/models"
)

var DB *pg.DB

func init() {
	DB = pg.Connect(&pg.Options{
		User:     constant.POSTGRESQL_USER,
		Password: constant.POSTGRESQL_PASSWORD,
		Database: constant.POSTGRESQL_DB_NAME,
		Addr:     fmt.Sprintf("%s:%s", constant.POSTGRESQL_HOST, constant.POSTGRESQL_PORT),
	})
	if !constant.DISABLE_SQL_GENERATE_PRINT {
		DB.AddQueryHook(SqlGeneratePrintHook{})
	}
	//orm.RegisterTable((*models.OrderGood)(nil))
	if !constant.DISABLE_CREATE_TABLE {
		for _, model := range []interface{}{
{{.models}}
		} {
			err := DB.Model(model).CreateTable(&orm.CreateTableOptions{
				Temp:          false,
				IfNotExists:   true,
				FKConstraints: true,
			})
			if err != nil {
				panic(err)
			}
		}
	}
}

type SqlGeneratePrintHook struct{}

func (hook SqlGeneratePrintHook) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	return c, nil
}

func (hook SqlGeneratePrintHook) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	data, err := q.FormattedQuery()
	if len(string(data)) > 8 { //BEGIN COMMIT
		//log.Debug(string(data))
	}
	return err
}

`

const tmplPgTransaction = `package transaction

import "github.com/go-pg/pg/v10"

type TransactionContext struct {
	PgDd *pg.DB
	PgTx *pg.Tx
}

func (transactionContext *TransactionContext) StartTransaction() error {
	tx, err := transactionContext.PgDd.Begin()
	if err != nil {
		return err
	}
	transactionContext.PgTx = tx
	return nil
}

func (transactionContext *TransactionContext) CommitTransaction() error {
	err := transactionContext.PgTx.Commit()
	return err
}

func (transactionContext *TransactionContext) RollbackTransaction() error {
	err := transactionContext.PgTx.Rollback()
	return err
}

func NewPGTransactionContext(pgDd *pg.DB) *TransactionContext {
	return &TransactionContext{
		PgDd: pgDd,
	}
}
`
