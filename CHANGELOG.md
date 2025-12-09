# Changelog

## [最新版本] - 2025-12-09

### 🎉 重大更新：对齐 Node.js SDK 数据结构

本次更新将 Go SDK 的批量操作返回结构与 Node.js SDK 完全对齐，提供更一致的开发体验。

### ✨ 新增功能

#### 1. 统一的批量操作返回结构

所有 `RecordsWithIterator` 方法（创建、更新、删除）现在返回统一的 `BatchOperationResult` 结构：

```go
type BatchOperationResult struct {
    Total        int             `json:"total"`        // 总记录数
    Success      []OperationItem `json:"success"`      // 成功的记录
    Failed       []OperationItem `json:"failed"`       // 失败的记录
    SuccessCount int             `json:"successCount"` // 成功数量
    FailedCount  int             `json:"failedCount"`  // 失败数量
}

type OperationItem struct {
    ID      string `json:"_id"`             // 记录 ID
    Success bool   `json:"success"`         // 是否成功
    Error   string `json:"error,omitempty"` // 错误信息（失败时）
}
```

#### 2. 方法返回类型变更对比

| 方法 | 之前返回 | 现在返回 | 变更原因 |
|------|---------|---------|----------|
| `Object.Create.RecordsWithIterator` | `*RecordsIteratorResult` | `*BatchOperationResult` | 提供成功/失败统计 |
| `Object.Update.RecordsWithIterator` | `BatchResponses` | `*BatchOperationResult` | 统一返回结构 |
| `Object.Delete.RecordsWithIterator` | `BatchResponses` | `*BatchOperationResult` | 统一返回结构 |
| `Object.Search.RecordsWithIterator` | `*RecordsIteratorResult` | `*RecordsIteratorResult` | 保持不变（查询场景） |

#### 3. 支持自定义批次大小

所有批量操作现在支持通过 `Limit` 参数自定义每批次处理的记录数量（默认 100）：

```go
// 创建
result, err := client.Object.Create.RecordsWithIterator(ctx, apaas.ObjectCreateRecordsIteratorParams{
    ObjectName: "object_event_log",
    Records:    records,
    Limit:      100, // 新增：可选参数，默认 100
})

// 更新
result, err := client.Object.Update.RecordsWithIterator(ctx, apaas.ObjectUpdateRecordsIteratorParams{
    ObjectName: "object_store",
    Records:    records,
    Limit:      100, // 新增：可选参数，默认 100
})

// 删除
result, err := client.Object.Delete.RecordsWithIterator(ctx, apaas.ObjectDeleteRecordsIteratorParams{
    ObjectName: "object_store",
    IDs:        ids,
    Limit:      100, // 新增：可选参数，默认 100
})
```

### 🔧 改进

#### 1. 增强的错误处理

- 所有批量操作现在在单个批次失败时不会中断整个流程
- 失败的记录会被标记并收集到 `Failed` 数组中
- 提供详细的错误信息，便于调试

#### 2. 代码复用优化

- `Update.RecordsWithIterator` 现在调用 `Update.Records()` 方法（与 Node.js 保持一致）
- `Delete.RecordsWithIterator` 现在调用 `Delete.Records()` 方法（与 Node.js 保持一致）
- 减少代码重复，提高可维护性

#### 3. 参数校验增强

所有批量操作方法都增加了参数校验：

```go
// 空数组校验
if len(params.Records) == 0 {
    return &BatchOperationResult{
        Total: 0, 
        Success: []OperationItem{}, 
        Failed: []OperationItem{},
        SuccessCount: 0,
        FailedCount: 0,
    }, nil
}

// nil 参数校验
if params.Records == nil {
    return nil, fmt.Errorf("参数 records 必须是一个数组")
}
```

#### 4. 日志输出优化

- 增加了更详细的日志信息
- 批次处理进度更清晰
- 成功/失败统计实时输出

### 📝 迁移指南

#### 从旧版本升级

**创建操作**

```go
// 之前
result, err := client.Object.Create.RecordsWithIterator(ctx, params)
fmt.Printf("Created: %d items\n", len(result.Items))

// 现在
result, err := client.Object.Create.RecordsWithIterator(ctx, params)
fmt.Printf("Total: %d, Success: %d, Failed: %d\n", 
    result.Total, result.SuccessCount, result.FailedCount)

// 处理失败的记录
for _, failed := range result.Failed {
    fmt.Printf("Failed ID: %s, Error: %s\n", failed.ID, failed.Error)
}
```

**更新操作**

```go
// 之前
responses, err := client.Object.Update.RecordsWithIterator(ctx, params)
for _, resp := range responses {
    fmt.Printf("Response code: %s\n", resp.Code)
}

// 现在
result, err := client.Object.Update.RecordsWithIterator(ctx, params)
fmt.Printf("Total: %d, Success: %d, Failed: %d\n", 
    result.Total, result.SuccessCount, result.FailedCount)

// 处理失败的记录
for _, failed := range result.Failed {
    fmt.Printf("Failed ID: %s, Error: %s\n", failed.ID, failed.Error)
}
```

**删除操作**

```go
// 之前
responses, err := client.Object.Delete.RecordsWithIterator(ctx, params)
for _, resp := range responses {
    fmt.Printf("Response code: %s\n", resp.Code)
}

// 现在
result, err := client.Object.Delete.RecordsWithIterator(ctx, params)
fmt.Printf("Total: %d, Success: %d, Failed: %d\n", 
    result.Total, result.SuccessCount, result.FailedCount)

// 处理失败的记录
for _, failed := range result.Failed {
    fmt.Printf("Failed ID: %s, Error: %s\n", failed.ID, failed.Error)
}
```

### 🎯 与 Node.js SDK 的一致性

| 特性 | Node.js SDK | Go SDK | 状态 |
|------|-------------|--------|------|
| 创建操作返回结构 | `{ total, success, failed, successCount, failedCount }` | `BatchOperationResult` | ✅ 一致 |
| 更新操作返回结构 | `{ total, success, failed, successCount, failedCount }` | `BatchOperationResult` | ✅ 一致 |
| 删除操作返回结构 | `{ total, success, failed, successCount, failedCount }` | `BatchOperationResult` | ✅ 一致 |
| 查询操作返回结构 | `{ total, items }` | `RecordsIteratorResult` | ✅ 一致 |
| 调用对应的 records 方法 | ✅ 是 | ✅ 是 | ✅ 一致 |
| 错误处理 | ✅ 完善 | ✅ 完善 | ✅ 一致 |
| 支持 limit 参数 | ✅ 支持 | ✅ 支持 | ✅ 一致 |

### 📚 文档更新

- ✅ 更新 `README.md` - 新增批量操作示例
- ✅ 更新 `UserManual.md` - 完整的使用说明和代码示例
- ✅ 更新 `examples/simple/main.go` - 演示新的 API 用法
- ✅ 新增 `CHANGELOG.md` - 记录所有变更

### 🐛 Bug 修复

- 修复批量操作在部分失败时可能丢失数据的问题
- 修复错误处理不够详细的问题
- 修复日志输出格式不一致的问题

### ⚠️ 破坏性变更

以下方法的返回类型发生了变化，需要更新调用代码：

1. `Object.Create.RecordsWithIterator` - 返回类型从 `*RecordsIteratorResult` 改为 `*BatchOperationResult`
2. `Object.Update.RecordsWithIterator` - 返回类型从 `BatchResponses` 改为 `*BatchOperationResult`
3. `Object.Delete.RecordsWithIterator` - 返回类型从 `BatchResponses` 改为 `*BatchOperationResult`

---

## 下一步计划

- [ ] 添加单元测试覆盖新的批量操作功能
- [ ] 添加集成测试
- [ ] 性能优化和基准测试
- [ ] 支持更多的 API 端点

---

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License
