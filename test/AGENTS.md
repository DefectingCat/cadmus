<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# test 目录

测试工具和 Mock 实现目录，为项目提供完整的单元测试基础设施。

## 目录结构

```
test/
├── helpers/          # 测试辅助函数和 fixtures
└── mocks/            # 接口 Mock 实现
```

## 子目录说明

### helpers/

存放测试辅助函数和测试数据 fixtures。

- `doc.go` - 包声明文件，标识 `helpers` 包

### mocks/

存放使用 `testify/mock` 编写的 Mock 实现。

- `doc.go` - 包声明文件，标识 `mocks` 包
- `repository_mocks.go` - Repository 接口的 Mock 实现集合

## Mock 列表

### User Module

| Mock 类 | 接口 |
|---------|------|
| `MockUserRepository` | `UserRepository` |
| `MockRoleRepository` | `RoleRepository` |
| `MockTokenBlacklist` | `TokenBlacklist` |

### Post Module

| Mock 类 | 接口 |
|---------|------|
| `MockPostRepository` | `PostRepository` |
| `MockCategoryRepository` | `CategoryRepository` |
| `MockTagRepository` | `TagRepository` |
| `MockSeriesRepository` | `SeriesRepository` |
| `MockPostLikeRepository` | `PostLikeRepository` |

### Comment Module

| Mock 类 | 接口 |
|---------|------|
| `MockCommentRepository` | `CommentRepository` |
| `MockCommentLikeRepository` | `CommentLikeRepository` |

### Media Module

| Mock 类 | 接口 |
|---------|------|
| `MockMediaRepository` | `MediaRepository` |

### Search Module

| Mock 类 | 接口 |
|---------|------|
| `MockSearchRepository` | `SearchRepository` |

## AI Agent 测试指南

### testify 框架使用

```go
import (
    "testing"
    "github.com/stretchr/testify/mock"
    "rua.plus/cadmus/test/mocks"
)

func TestUserService(t *testing.T) {
    // 1. 创建 mock
    mockRepo := new(mocks.MockUserRepository)
    
    // 2. 设置预期
    mockRepo.On("GetByID", mock.Anything, mock.Anything).
        Return(&user.User{ID: uuid.New()}, nil)
    
    // 3. 调用被测试的函数
    svc := NewUserService(mockRepo)
    result, err := svc.GetUser(ctx, id)
    
    // 4. 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, result)
    
    // 5. 验证预期调用
    mockRepo.AssertExpectations(t)
}
```

### Mock 使用模式

#### 设置返回值

```go
// 成功场景
mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

// 错误场景
mockRepo.On("GetByID", mock.Anything, mock.Anything).
    Return(nil, errors.New("not found"))

// 多返回值
mockRepo.On("List", mock.Anything, mock.Anything, mock.Anything).
    Return([]*user.User{...}, 10, nil)
```

#### 验证调用

```go
// 验证是否被调用
mockRepo.AssertCalled(t, "Create", mock.Anything, mock.Anything)

// 验证调用次数
mockRepo.AssertNumberOfCalls(t, "GetByID", 2)

// 验证调用顺序
mock.AssertCallOrder(t,
    mockRepo.On("Create", ...),
    mockRepo.On("Update", ...),
)
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./test/mocks

# 带覆盖率
go test -cover ./...

# 显示详细输出
go test -v ./...
```
