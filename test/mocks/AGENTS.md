<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# mocks 目录

测试 Mock 实现目录，提供基于 `stretchr/testify/mock` 的接口 Mock 实现，用于单元测试中的依赖注入。

## 目录用途

`mocks` 包提供项目各核心模块的 Repository 接口 Mock 实现，使单元测试能够：

- **隔离测试目标**：无需真实数据库即可测试服务层逻辑
- **控制测试场景**：精确模拟成功、失败、边缘情况
- **验证交互行为**：断言方法调用顺序、参数、次数

## 关键文件

| 文件 | 功能 |
|------|------|
| `doc.go` | 包声明，标识 `mocks` 包及依赖 |
| `repository_mocks.go` | 所有 Repository 接口的 Mock 实现集合 |

## Mock 实现清单

### User Module

| Mock 类 | 接口 | 方法数 |
|---------|------|---------|
| `MockUserRepository` | `UserRepository` | 7 |
| `MockRoleRepository` | `RoleRepository` | 5 |
| `MockTokenBlacklist` | `TokenBlacklist` | 2 |

### Post Module

| Mock 类 | 接口 | 方法数 |
|---------|------|---------|
| `MockPostRepository` | `PostRepository` | 15 |
| `MockCategoryRepository` | `CategoryRepository` | 9 |
| `MockTagRepository` | `TagRepository` | 11 |
| `MockSeriesRepository` | `SeriesRepository` | 5 |
| `MockPostLikeRepository` | `PostLikeRepository` | 6 |

### Comment Module

| Mock 类 | 接口 | 方法数 |
|---------|------|---------|
| `MockCommentRepository` | `CommentRepository` | 11 |
| `MockCommentLikeRepository` | `CommentLikeRepository` | 8 |

### Media Module

| Mock 类 | 接口 | 方法数 |
|---------|------|---------|
| `MockMediaRepository` | `MediaRepository` | 6 |

### Search Module

| Mock 类 | 接口 | 方法数 |
|---------|------|---------|
| `MockSearchRepository` | `SearchRepository` | 4 |

## AI Agent Mock 使用指南

### 1. 基本使用模式

```go
package service_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/google/uuid"

    "rua.plus/cadmus/internal/core/user"
    "rua.plus/cadmus/test/mocks"
)

func TestUserService_GetUser(t *testing.T) {
    // 创建 mock
    mockRepo := new(mocks.MockUserRepository)
    
    // 设置预期：GetByID 被调用时返回预定义用户
    expectedUser := &user.User{
        ID:       uuid.New(),
        Username: "testuser",
        Email:    "test@example.com",
    }
    mockRepo.On("GetByID", mock.Anything, mock.Anything).
        Return(expectedUser, nil)
    
    // 注入 mock 到服务
    svc := NewUserService(mockRepo)
    result, err := svc.GetUser(context.Background(), expectedUser.ID)
    
    // 验证结果
    assert.NoError(t, err)
    assert.Equal(t, expectedUser.Username, result.Username)
    
    // 验证预期调用
    mockRepo.AssertExpectations(t)
}
```

### 2. 模拟错误场景

```go
// 模拟记录不存在
mockRepo.On("GetByID", mock.Anything, mock.Anything).
    Return(nil, errors.New("user not found"))

// 模拟数据库错误
mockRepo.On("Create", mock.Anything, mock.Anything).
    Return(errors.New("database connection failed"))
```

### 3. 验证调用参数

```go
// 验证特定参数被传递
mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *user.User) bool {
    return u.Email == "test@example.com"
})).Return(nil)

// 验证后断言
mockRepo.AssertCalled(t, "Create", mock.Anything, expectedUser)
mockRepo.AssertNotCalled(t, "Delete")
```

### 4. 验证调用次数

```go
// 验证方法被调用的次数
mockRepo.AssertNumberOfCalls(t, "GetByID", 2)

// 验证调用顺序
expectedOrder := []string{"Create", "Update"}
mock.AssertCalledAtLeastOnce(t, mockRepo, expectedOrder...)
```

### 5. 多返回值处理

```go
// List 方法返回切片、总数和错误
mockRepo.On("List", mock.Anything, 0, 10).
    Return([]*user.User{user1, user2}, 25, nil)

// 注意：nil 切片需要显式转换为正确类型
mockRepo.On("List", mock.Anything, 0, 10).
    Return([]*user.User(nil), 0, errors.New("error"))
```

### 6. 常见陷阱

```go
// 错误：返回 nil 导致类型断言失败
// Return(nil, nil) // 第一个 nil 无法断言为 *user.User

// 正确：显式返回 nil 指针
// Return((*user.User)(nil), nil)

// 正确：使用条件判断
mockRepo.On("GetByID", mock.Anything, mock.Anything).
    Run(func(args mock.Arguments) {
        // 可以在这里添加自定义逻辑
    }).
    Return(expectedUser, nil)
```

## 运行测试

```bash
# 运行服务层测试（使用 mocks）
go test ./internal/service/... -v

# 运行所有测试
go test ./... -v

# 带覆盖率
go test ./... -cover
```

## 添加新 Mock

1. 在 `repository_mocks.go` 中添加新的 Mock 结构体：

```go
type MockNewRepository struct {
    mock.Mock
}

func (m *MockNewRepository) Method(ctx context.Context, arg Type) (ReturnType, error) {
    args := m.Called(ctx, arg)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(ReturnType), args.Error(1)
}
```

2. 处理 nil 返回值时需要类型转换以避免断言失败。

## 子目录

无子目录。所有 Mock 实现集中在 `repository_mocks.go` 文件中。
