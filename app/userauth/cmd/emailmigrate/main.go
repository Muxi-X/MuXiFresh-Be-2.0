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
)

// 一次性邮箱存量治理命令。默认 dry-run，仅加 -apply 才写入；不进任何服务启动
// 路径，必须由运维手动执行一次（治理完成即不再使用）。详见 pr-notes 设计文档。
func main() {
	apply := flag.Bool("apply", false, "真正写入改动；默认只做 dry-run 打印计划")
	mongoURL := flag.String("mongo", "", "直连 MongoDB URL（覆盖 Nacos，便于本地/测试）")
	dbName := flag.String("db", "", "数据库名（覆盖 Nacos，需与 -mongo 同时使用）")
	flag.Parse()

	url, db, err := resolveTarget(*mongoURL, *dbName)
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

// resolveTarget 返回治理目标；优先使用 -mongo/-db 直连覆盖，否则从 Nacos 的
// infra 读取。直接覆盖必须成对给出。
func resolveTarget(mongoURL, dbName string) (string, string, error) {
	if mongoURL != "" || dbName != "" {
		if mongoURL == "" || dbName == "" {
			return "", "", fmt.Errorf("-mongo 与 -db 必须同时提供")
		}
		return mongoURL, dbName, nil
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
