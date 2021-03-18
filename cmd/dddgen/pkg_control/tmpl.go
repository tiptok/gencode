package pkg_control

const tmplConstantRedis = `package constant

import "os"

var REDIS_HOST = "127.0.0.1"
var REDIS_PORT = "6379"
var REDIS_AUTH = "123456"

func init() {
	if os.Getenv("REDIS_HOST") != "" {
		REDIS_HOST = os.Getenv("REDIS_HOST")
		REDIS_AUTH = os.Getenv("REDIS_AUTH")
	}
	if os.Getenv("REDIS_PORT") != "" {
		REDIS_PORT = os.Getenv("REDIS_PORT")
	}
	if _, ok := os.LookupEnv("REDIS_AUTH"); ok {
		REDIS_AUTH = os.Getenv("REDIS_AUTH")
	}
}`

const tmplRedisInit = `package redis

import (
	"fmt"
	"github.com/tiptok/gocomm/pkg/cache"
	"github.com/tiptok/gocomm/pkg/log"
	"github.com/tiptok/gocomm/pkg/redis"
)

func init() {
	redisSource := fmt.Sprintf("%v:%v", constant.REDIS_HOST, constant.REDIS_PORT)
	err := redis.InitWithDb(100, redisSource, constant.REDIS_AUTH, "0")
	if err != nil {
		log.Error(err)
	}
	cache.InitDefault(cache.WithDefaultRedisPool(redis.GetRedisPool()))
}`
