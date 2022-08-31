## 项目结构

项目整体分为 web，controller 和 repository 三层。

web 相关的组件和 handler 在[internal/web/](../internal/web) 文件夹中。

controller 相关在[internal/ctrl/](../internal/ctrl) 文件夹，按照读写区分。

service 和 repository 在[internal/domain/](../internal/domain)定义为 interface，在各个子文件夹(如[internal/subject/](internal/subject)) 中实现。

## 数据库

`bangumi/dev-env` 仓库仅导出了部分表和数据。如果对应表定义和数据缺失请联系 @Trim21 进行导出。

## 依赖注入

项目使用 https://github.com/uber-go/fx 进行依赖注入。

在 [main.go](../main.go) 中提供接口的构造函数。在 [internal/web/handler/new.go](../internal/web/handler/new.go) 中添加函数参数和结构体字段即可。

为了方便在测试中进行 mock，请使用 interface，而非具体的类型。

## 测试

测试使用 go 自带的单元测试框架 https://pkg.go.dev/testing

### 测试文件的命名和包名

导出函数和方法应当在 `${filename}_test.go`文件中进行测试，包名应该添加 `_test` 后缀。

对于 **需要进行测试的** 的内部函数和方法，在`${filename}_internal_test.go`文件中进行测试，并且使用原本的包名。

例如：

在 `github.com/bangumi/server/blob/master/pkg/wiki` 包中，[parser.go](../pkg/wiki/parser.go) 文件的测试分别位于[parser_internal_test.go](../pkg/wiki/parser_internal_test.go) 和 [parser_test.go](../pkg/wiki/parser_test.go) 两个文件中。

`parser_internal_test.go` 属于 `package wiki`，可以使用 `wiki` 包内的非导出函数和方法。

`parser_test.go` 属于 `package wiki_test`，只能使用 `wiki` 包的导出函数和方法。

### mock 外部资源

大多数测试应该对外部依赖进行 mock。

mock 使用 mockery 生成，可以查看 [internal/mocks](../internal/mocks/) 文件夹。

每一个 `web/handler.Handler` 用到的接口，在测试时都应该有 `test.Mock{}` 提供的选项进行 mock。

如:

```golang
package handler_test

import (
	"testing"

	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestHandler_Mock_Example(t *testing.T) {
	t.Parallel()

	app := test.GetWebApp(t, test.Mock{
		SessionManager: mocks.NewSessionManager(t),
		AuthService:    mocks.NewAuthService(t),
	})
	// test app now
}
```

没有进行 mock 的依赖会由 `test.GetWebApp` 提供一个空实现，空实现可以满足 go 编译器的类型约束，但是在测试过程中调用空实现任何方法都会报错。

在 `internal/domain` 添加了对应的接口后，应该使用 mockery 生成对应的 mock 实现。并且在 `internal/test.Mock{}` 类型添加对应的字段。

### 对于外部资源进行测试

如果你测试的是外部依赖本身，如对应的 mysql_repository，请对测试进行标记。

[internal/auth/mysql_repository_test.go](../internal/auth/mysql_repository_test.go)：

```golang
package auth_test

import (
	"testing"

	"github.com/bangumi/server/internal/pkg/test"
)


func TestMysqlRepo_GetByToken(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()
	// then write your tests
}
```

## 代码生成

目前有两部分代码是自动生成的，gorm-gen 和 mocks。

设置数据库相关的环境变量（或者用`.env`）后使用 `task gorm` 生成 gorm 相关的 dal。

使用 `task mock` 在自动生成用到的 mock，相关的 task 定义在 [etc/mock.task.yaml](../etc/mock.task.yaml) 中。

## Code Style

[Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

例外：

## 99 character line length

120。

每个 tab 计 2 个字符宽。

### Local Variable Declarations

使用 `var` 或者 `:=` 均可。

只有初始化变量为零值的时候应该用 `var`

#### bad

```golang
var v = uint32(0)
```

#### good

```golang
var v uint32
```

### Import Group Ordering

import 应该分为 std, external, internal 三部分。

### Import Aliasing

go mod 在 v2 以上的版本会自动添加大版本后缀，不需要针对此情况添加 alias。

### nil is a valid slice

JSON encoder 不会把 `nil` 序列化为空数组，所以在 web 响应空数组时时不应该使用 `nil`。

其他情况则应该使用 `nil` 代替返回空数组。
