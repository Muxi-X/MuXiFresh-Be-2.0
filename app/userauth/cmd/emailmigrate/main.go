package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"MuXiFresh-Be-2.0/app/userauth/cmd/emailmigrate/migrate"
	"MuXiFresh-Be-2.0/common/infra"
	"MuXiFresh-Be-2.0/common/nacos"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

// 一次性邮箱存量治理命令：把规范化后重复的账号收敛为唯一 keeper（旧账号摘除邮箱
// 并迁移学号），再把全库邮箱域名拍平为小写，供建 email 唯一索引前使用。
// 默认 dry-run，仅加 -apply 才写入；不进任何服务启动路径，由运维手动执行一次。
//
// 用法（DSN 含凭据，命令不会回显它）：
//
//	go run ./app/userauth/cmd/emailmigrate -dsn "<mongo uri>[:port]/<db>"
//	go run ./app/userauth/cmd/emailmigrate -dsn "<mongo uri>[:port]/<db>" -apply
//
// 不传 -dsn 时回退到 Nacos infra（需能连上 Nacos）。
func main() {
	apply := flag.Bool("apply", false, "真正写入改动；默认只做 dry-run 打印计划")
	dsn := flag.String("dsn", "", "MongoDB 连接串（含数据库名），如 mongodb://user:pass@host:27017/dbname")
	dbName := flag.String("db", "", "数据库名；仅在 DSN 未带库名时使用")
	flag.Parse()

	url, db, err := resolveTarget(*dsn, *dbName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "resolve mongo target:", err)
		os.Exit(1)
	}

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect mongo:", err)
		os.Exit(1)
	}
	defer client.Disconnect(context.Background())

	if !*apply {
		fmt.Println("[dry-run] 仅打印计划，不改动数据。")
	}
	if err := migrate.Run(ctx, client, db, *apply, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "migrate failed:", err)
		os.Exit(1)
	}
	fmt.Println("done.")
}

// resolveTarget 返回治理目标；给出 -dsn 时只连该库（不回显 DSN），否则从 Nacos
// infra 读取。数据库名优先取 -db，其次取 DSN 路径；都缺则报错。
func resolveTarget(dsn, dbName string) (string, string, error) {
	if dsn != "" {
		parsed, err := connstring.Parse(dsn)
		if err != nil {
			return "", "", fmt.Errorf("解析 -dsn 失败: %w", err)
		}
		db := dbName
		if db == "" {
			db = parsed.Database
		}
		if db == "" {
			return "", "", fmt.Errorf("-dsn 未包含数据库名，请用 -db 指定")
		}
		return dsn, db, nil
	}

	if dbName != "" {
		return "", "", fmt.Errorf("-db 只能与 -dsn 一起使用")
	}

	var c infra.Config
	if err := nacos.Load(nacos.Infra(&c)); err != nil {
		return "", "", err
	}
	if c.MongoDB.URL == "" || c.MongoDB.DB == "" {
		return "", "", fmt.Errorf("Nacos infra 缺少 MongoDB.URL 或 DB")
	}
	return c.MongoDB.URL, c.MongoDB.DB, nil
}
