package infra

import (
	"reflect"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// Config contains connection settings shared by services. It is loaded from a
// dedicated Nacos data ID so infrastructure changes do not require editing
// every service configuration.
type Config struct {
	MongoDB       MongoDBConf
	Redis         redis.RedisConf
	Etcd          EtcdConf
	Kafka         KafkaConf
	SMTP          SMTPConf
	ObjectStorage ObjectStorageConf
	Middlewares   MiddlewaresConf
	RpcAuth       RpcAuthConf
}

// RpcAuthConf 是服务间 RPC 调用的共享密钥配置。
// Server 侧未配置（空）时启动失败（fail closed），防止静默裸奔。
type RpcAuthConf struct {
	Token string
}

// MiddlewaresConf mirrors go-zero's server middleware switches
// (rest.RestConf.Middlewares and zrpc.RpcServerConf.Middlewares). Only
// switches set to true here are injected into services; existing per-service
// configuration is never cleared.
type MiddlewaresConf struct {
	Trace      bool
	Log        bool
	Prometheus bool
	MaxConns   bool
	Breaker    bool
	Shedding   bool
	Timeout    bool
	Recover    bool
	Metrics    bool
	MaxBytes   bool
	Gunzip     bool
	Duration   bool
	Stat       bool
}

type MongoDBConf struct {
	URL string
	DB  string
}

// EtcdConf contains only the shared Etcd connection. Registry keys remain in
// each service configuration because they identify individual services.
type EtcdConf struct {
	Hosts              []string
	User               string
	Pass               string
	CertFile           string
	CertKeyFile        string
	CACertFile         string
	InsecureSkipVerify bool
}

type KafkaConf struct {
	Brokers  []string
	Username string
	Password string
}

type SMTPConf struct {
	Host     string
	Port     string
	UserName string
	Password string
}

type ObjectStorageConf struct {
	AccessKey  string
	SecretKey  string
	BucketName string
	DomainName string
}

func (c Config) ApplyEtcd(targets ...*discov.EtcdConf) {
	// go-zero's etcd client does not automatically use the User/Pass in config;
	// must call RegisterAccount explicitly, otherwise connecting to an
	// authenticated etcd fails with "user name is empty".
	if c.Etcd.User != "" && c.Etcd.Pass != "" {
		discov.RegisterAccount(c.Etcd.Hosts, c.Etcd.User, c.Etcd.Pass)
	}

	for _, target := range targets {
		if target == nil {
			continue
		}

		target.Hosts = append([]string(nil), c.Etcd.Hosts...)
		target.User = c.Etcd.User
		target.Pass = c.Etcd.Pass
		target.CertFile = c.Etcd.CertFile
		target.CertKeyFile = c.Etcd.CertKeyFile
		target.CACertFile = c.Etcd.CACertFile
		target.InsecureSkipVerify = c.Etcd.InsecureSkipVerify
	}
}

func (c Config) ApplyKafka(targets ...*kq.KqConf) {
	for _, target := range targets {
		if target == nil {
			continue
		}

		target.Brokers = append([]string(nil), c.Kafka.Brokers...)
		target.Username = c.Kafka.Username
		target.Password = c.Kafka.Password
	}
}

// ApplyMiddlewares injects infrastructure middleware switches into the
// services' config structs. It matches the `Middlewares` field (present in both
// rest.RestConf and zrpc.RpcServerConf) by field name and only ever enables a
// switch (true in infra), never disabling an explicitly configured one.
func (c Config) ApplyMiddlewares(targets ...interface{}) {
	mwType := reflect.TypeOf(c.Middlewares)
	mwValue := reflect.ValueOf(c.Middlewares)

	for _, target := range targets {
		rv := reflect.ValueOf(target)
		if rv.Kind() != reflect.Ptr || rv.IsNil() {
			continue
		}
		elem := rv.Elem()
		if elem.Kind() != reflect.Struct {
			continue
		}
		mw := elem.FieldByName("Middlewares")
		if !mw.IsValid() || mw.Kind() != reflect.Struct || !mw.CanSet() {
			continue
		}

		for i := 0; i < mwType.NumField(); i++ {
			if !mwValue.Field(i).Bool() {
				continue
			}
			field := mw.FieldByName(mwType.Field(i).Name)
			if field.IsValid() && field.CanSet() && field.Kind() == reflect.Bool {
				field.SetBool(true)
			}
		}
	}
}
