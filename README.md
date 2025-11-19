# 木犀招新系统 v2

基于 `go-zero` 的木犀招新系统后端仓库

## 服务

- auth：身份认证
- user：用户信息
- task：作业
- review：审阅
- schedule：进度
- form：报名表
- test：测验

## 运行

### 1. 配置

 复制 `~/etc/app-example.yaml` 文件为 `~/etc/app.yaml`，并根据需要进行配置。

### 2. 构建运行

- `rpc`服务

  进入`rpc`服务目录，执行

  ```go
  go run rpc.go -f etc/rpc.yaml
  ```

- `api`服务

  进入`api`服务目录，执行

  ```go
  go run api.go -f etc/api.yaml
  ```

ps：运行整个项目时，user服务需要在task，review，form，test服务之前启动。

