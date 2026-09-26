# Nacos YAML 配置约定

Nacos 中的配置内容统一使用 YAML，不再接受 JSON。所有运行配置自托管在云端 Nacos，仓库内不保存配置模板或运行配置。

配置分为两类：

- `infra`：共享的 MongoDB、Redis、Etcd、Kafka、SMTP、对象存储连接信息及账密，以及中间件开关和 RPC 鉴权密钥。
- 服务配置：每个进程独立维护，只包含端口、RPC、JWT、Kafka 和业务参数等自身配置。

默认情况下，所有配置位于当前 Nacos namespace 的 `PROD` group。

Nacos 自身的地址和账号仍通过环境变量传入，因为应用必须先连接 Nacos，才能读取 `infra`；它不能依赖自己尚未读取到的配置。

## Data ID

| Data ID | 用途 |
| --- | --- |
| `infra` | MongoDB、Redis、Etcd、Kafka、SMTP、对象存储、中间件开关、RPC 鉴权密钥 |
| `accountCenter` | 账户 RPC |
| `assignment` | 任务 RPC |
| `comment` | 评论 RPC |
| `form-api` | 报名 API |
| `form-rpc` | 报名 RPC |
| `intro-api` | 介绍 API |
| `intro-rpc` | 介绍 RPC |
| `review` | 审核 API |
| `schedule-api` | 进度 API |
| `schedule-rpc` | 进度 RPC |
| `submission` | 提交 RPC |
| `task` | 任务 API |
| `test-api` | 测试 API |
| `test-rpc` | 测试 RPC |
| `user-api` | 用户 API |
| `user-auth` | 认证 API、Kafka Topic/Group、验证码 |
| `user-rpc` | 用户 RPC |

配置中的密码、JWT 密钥、域名、MongoDB 地址等均按环境填入实际值，直接维护在云端 Nacos。所有验证 JWT 的服务应使用与 `user-auth` 相同的 `JwtAuth.AccessSecret`。

服务配置只保留自身语义：Etcd `Key`、Kafka `Topic/Group`、端口、JWT 和业务参数。以下公共信息只能存在于 `infra`：

- MongoDB URL、数据库名
- Redis Host、类型、密码、TLS
- Etcd Hosts、账号、密码、TLS 证书
- Kafka Brokers、账号、密码
- SMTP Host、端口、账号、密码
- 对象存储 AccessKey、SecretKey、Bucket、Domain
- 中间件开关（`Middlewares`，如 `Recover`）
- RPC 鉴权共享密钥（`RpcAuth.Token`）：服务间 RPC 调用鉴权，所有 RPC server/API client 共享；**缺失时全部服务启动失败（fail closed）**，上线前必须先配好

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `NACOS_ADDR` | 无 | Nacos 服务地址，端口固定为 `8848` |
| `NACOS_NAMESPACE` | 无 | Nacos namespace ID |
| `NACOS_USERNAME` | 无 | Nacos 用户名 |
| `NACOS_PASSWORD` | 无 | Nacos 密码 |
| `NACOS_GROUP` | `PROD` | 服务配置所在 group |
| `NACOS_INFRA_GROUP` | `NACOS_GROUP` | infra 配置所在 group |
| `NACOS_INFRA_DATA_ID` | `infra` | infra 配置的 Data ID |

## 导入顺序

1. 在 Nacos 上创建 `infra` Data ID，配置格式选择 YAML。
2. 按表格创建各服务 Data ID，填入对应 YAML。
3. 检查服务配置里的 RPC `Etcd.Key` 与调用方完全一致。
4. 启动服务；确认正常后，删除 Nacos 中遗留的旧 JSON 配置。
