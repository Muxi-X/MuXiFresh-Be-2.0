# 木犀招新系统 v2

Nacos 的服务配置与基础设施配置约定见 [deploy/nacos/README.md](deploy/nacos/README.md)。

基于 `go-zero` 的木犀招新系统后端仓库

## 依赖说明：go-zero fork

本项目通过 `go.mod replace` 使用 go-zero 的团队 fork（**含 etcd 认证补丁**）。

- **为什么**：etcd 启用用户名密码认证后，token 默认 5 分钟过期，而 clientv3 的 watch 不自动刷新 token（[etcd#12385](https://github.com/etcd-io/etcd/issues/12385)），go-zero 又无限紧密重试同一 client，导致 `invalid auth token` 周期性刷屏、消耗 CPU 与日志磁盘。etcd 官方明确不修（[#17384](https://github.com/etcd-io/etcd/issues/17384)），go-zero 的修复 [PR #5709](https://github.com/zeromicro/go-zero/pull/5709) 未合并，只能 fork 打补丁
- **引用**：`replace github.com/zeromicro/go-zero => github.com/Muxi-X/go-zero v1.4.5-muxi.2`
- **补丁仓库**：[Muxi-X/go-zero](https://github.com/Muxi-X/go-zero)（分支 `muxi-patch`，tag `v1.4.5-muxi.2`，[diff](https://github.com/Muxi-X/go-zero/compare/v1.4.5...muxi-patch)）
- **补丁要点**：watch/keepalive 失败时移除缓存 client 惰性重建（新 token）+ 防 goroutine 泄漏/死锁，思路对齐上游 PR #5709，以 `[Muxi Patch]` 标记
- **升级注意**：升级 go-zero 需重新 apply 补丁；若上游 #5709 合并可评估替换回官方

## 服务

- auth：身份认证
- user：用户信息
- task：作业
- review：审阅
- schedule：进度
- form：报名表
- test：测验

## 开发

常用命令封装在 [Makefile](Makefile)：

```bash
make fmt      # gofmt -w .
make vet      # go vet ./...
make test     # go test ./...
make lint     # fmt-check + vet + test
```

CI（gofmt 校验、`go vet`、`go test`）在 PR 与 main push 时运行，见 [.github/workflows/ci.yaml](.github/workflows/ci.yaml)。

## 运行

### 1. 配置

运行时配置由 Nacos 统一管理，不再使用仓库内的 yaml 文件。需要先准备连接 Nacos 所需的环境变量：

| 变量 | 说明 |
| --- | --- |
| `NACOS_ADDR` | Nacos 地址，端口固定 `8848` |
| `NACOS_NAMESPACE` | Nacos namespace ID |
| `NACOS_USERNAME` / `NACOS_PASSWORD` | Nacos 账号密码 |

服务配置（`infra` 及各服务 Data ID）的字段约定与导入顺序见 [deploy/nacos/README.md](deploy/nacos/README.md)。

### 2. 构建运行

go-zero 的 `-f` 配置 flag 已不再使用，入口直接读取 Nacos。进入对应服务目录执行，入口文件名以该目录 `main` 包的文件为准（如 form 为 `form.go`、review 为 `review.go`、userauth 为 `user-auth.go`）：

- `rpc` 服务（例如 form/rpc）

  ```bash
  go run form.go
  ```

- `api` 服务（例如 form/api）

  ```bash
  go run form.go
  ```

ps：运行整个项目时，user 服务需要在 task，review，form，test 服务之前启动。

