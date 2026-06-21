# AGENTS.md

AI 编码 agent 的项目指引。通用开发文档见 [readme.md](readme.md)。

## 架构约定

分层：`web/handler → ctrl → internal/<domain> → dal`

```
web/handler/     HTTP 层，只做参数解析/验证，调用 ctrl，不直接调 repo
web/handler/common/  auth 中间件、AccessToken 解析
ctrl/            Controller 层，编排写操作，聚合多个 repo/cache，不含 HTTP 细节
internal/        领域层，每个子包（auth/ subject/ user/ ...）内定义 Service/Repo 接口并实现
dal/             GORM Gen 自动生成，只被 internal 层调用，web/ctrl 不直接使用
```

- **接口定义位置**: `internal/<domain>/domain.go` 或 `internal/<domain>/domain/`（如 collections）
- **不要跨层调用**: handler 不调 dal，ctrl 不碰 HTTP request/response
- **新增功能**: 从下往上，先在 internal 定义接口和实现，再在 ctrl 编排，最后在 handler 暴露

## 依赖注入 (uber-go/fx)

构造函数统一通过 `fx.Provide` 注册，不要用全局变量或 init 初始化依赖。

模块注册顺序（在 `cmd/web/cmd.go` 和 `web/fx.go`）：

```go
// cmd/web/cmd.go: 驱动 + 基础设施
fx.Provide(driver.NewRueidisClient, driver.NewMysqlDriver, ...)
dal.Module

// cmd/web/cmd.go: 领域层
fx.Provide(user.NewMysqlRepo, subject.NewMysqlRepo, auth.NewService, ...)

// ctrl/fx.go: ctrl.Module (fx.Provide(New))

// web/fx.go: web.Module
//   → handler.Module
//   → fx.Provide(New, session.NewMysqlRepo, session.New)
//   → fx.Invoke(AddRouters)
```

**新增依赖时的步骤**:
1. 在对应构造函数的参数中添加需要的接口/类型
2. 如果还没有对应的 `fx.Provide`，在 `cmd/web/cmd.go` 的对应位置添加
3. 新增接口后在 `.mockery.yaml` 添加配置，运行 `task mock` 生成 mock

## 测试约定

### 文件命名

- `xxx_test.go` — 包名 `xxx_test`，只能测试导出 API
- `xxx_internal_test.go` — 与源文件同包，测试内部函数/方法

### Mock 测试（默认）

```go
// handler 测试
func TestXxx(t *testing.T) {
    t.Parallel()
    app := test.GetWebApp(t, test.Mock{
        AuthService: mocks.NewAuthService(t),
        // 未填的会自动注入空实现，被调用时 panic
    })
    // 用 htest 发请求
}

// 其他层测试
func TestXxx(t *testing.T) {
    t.Parallel()
    mockRepo := mocks.NewSubjectRepo(t)
    mockRepo.EXPECT().Get(gomock.Any(), uint32(1)).Return(...)
    // ...
}
```

- mock 由 mockery 自动生成到 `internal/mocks/`，配置在 `.mockery.yaml`
- 新增领域接口时：在 `.mockery.yaml` 添加 → `task mock` 生成 → 在 `test.Mock{}` 中添加字段
- mock 用 [testify/mock](https://github.com/stretchr/testify#mock-package)，不是 gomock（上面示例中的 gomock 是泛指）

### 集成测试

需要 mysql/redis 的测试用 `test.RequireEnv` 标记：

```go
func TestMysqlRepo(t *testing.T) {
    test.RequireEnv(t, test.EnvMysql)
    t.Parallel()
}
```

### 不要做的事

- 不要在测试中硬编码 `time.Sleep` 等固定等待
- 空 slice 在 JSON 响应中不能用 `nil`（会被序列化为 `null`），用 `make([]T, 0)`
## 代码风格

- 行宽上限 120 字符，每个 tab 宽 2 字符
- import 分组：stdlib → external → internal（各一组，空行分隔）
- 零值变量用 `var v uint32`，不要 `var v = uint32(0)`
- JSON 序列化使用 `sonic`（在 `web/json.go` 中配置）
- Logger 使用 zap，从 fx 注入，不要用 `log.Println`
- 可见性校验用 `go-playground/validator` tag
- 用户可见错误 -> `gerr.Err*`；内部错误 -> `errgo.Wrap`
- GORM `ErrRecordNotFound` 用 `gerr.WrapGormError` 包裹

## 常用命令

```bash
task web        # 构建+启动 HTTP server（需 config.toml）
task lint       # golangci-lint（提交前必须通过）
task test       # 运行 mock 测试（无需外部依赖）
task test-db    # 运行需要 mysql/redis 的测试
task gen        # task gorm + task mock
task mock       # 重新生成 mock 文件
```

## 注意点

- `pkg/` 是唯一可被外部（非本仓库）导入的包，新增可复用工具放 `internal/pkg/` 而非 `pkg/`
- `dal/query/` 由 `task gorm` 从数据库自动生成，禁止手改
- `internal/mocks/` 由 `task mock` 自动生成，禁止手改
- canal 消费者依赖 Kafka + Debezium，测试时不需要启动
- 非 Go 文件（yaml/json/md）用 prettier 格式化
