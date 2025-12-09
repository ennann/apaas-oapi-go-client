# 背景

aPaaS 平台有完整的 Open API 能力，但是目前这些能力全都以单独接口的形式提供给开发者，不方便开发与调试。基于 Node.js SDK 的实践，我们提供了 Go 版本的 aPaaS OpenAPI SDK，封装公共能力、内置限流与 token 缓存，让 Go 应用可以更高效地接入平台。

## ✨ **功能特性**

- ✅ 获取 accessToken，自动管理 token 有效期
- ✅ record 单条查询、批量查询（支持分页迭代）
- ✅ record 单条创建、批量创建（支持分页迭代）
- ✅ record 单条更新、批量更新
- ✅ record 单条删除、批量删除
- ✅ 页面、附件、全局变量、自动化流程等模块封装
- ✅ 基于 `golang.org/x/time/rate` 的限流能力
- ✅ 可自定义日志实现和日志等级
- ……





**📦 安装**

```bash
go get github.com/ennann/apaas-oapi-go-client/apaas
```

***



# **🚀 快速开始**

```go
package main

import (
	"context"
	"log"

	"github.com/ennann/apaas-oapi-go-client/apaas"
)

func main() {
	client, err := apaas.NewClient(apaas.ClientOptions{
		ClientID:     "your_client_id",
		ClientSecret: "your_client_secret",
		Namespace:    "app_xxx",
	})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}

	client.SetLoggerLevel(apaas.LoggerLevelInfo) // 0-5

	ctx := context.Background()
	if err := client.Init(ctx); err != nil {
		log.Fatalf("init client: %v", err)
	}

	log.Printf("Access Token: %s", client.Token())
	if ttl, ok := client.TokenExpiresIn(); ok {
		log.Printf("Token expires in: %s", ttl)
	}
	log.Printf("Namespace: %s", client.Namespace())
}
```

***



## **🔐 认证**

### **初始化 Client**

| **参数** | **类型** | **说明** |
| :-- | :-- | :-- |
| ClientID | string | 应用 clientId |
| ClientSecret | string | 应用 clientSecret |
| Namespace | string | 命名空间 |
| DisableTokenCache | bool | 是否禁用 token 缓存，默认 false |
| BaseURL | string | 可选，覆盖默认网关地址 |
| HTTPClient | *http.Client | 可选，自定义 HTTP 客户端 |
| Logger | apaas.Logger | 可选，自定义日志实现 |
| LimiterOptions | *apaas.LimiterOptions | 可选，自定义限流参数 |

***



## **📝 日志等级**

调用 `client.SetLoggerLevel(level)` 设置日志等级。

| **Level** | **名称** | **说明** |
| :-- | :-- | :-- |
| apaas.LoggerLevelFatal | fatal | 严重错误 |
| apaas.LoggerLevelError | error | 错误 |
| apaas.LoggerLevelWarn | warn | 警告 |
| apaas.LoggerLevelInfo | info（默认） | 一般信息 |
| apaas.LoggerLevelDebug | debug | 调试信息 |
| apaas.LoggerLevelTrace | trace | 详细追踪 |

***



# 💾 **数据模块**

## **📋 对象列表接口**

### **获取所有对象（数据表）**

```go
res, err := client.Object.List(ctx, apaas.ObjectListParams{
	Offset: 0,
	Limit:  100,
	Filter: &apaas.ObjectListFilter{
		Type:       "custom",
		QuickQuery: "store",
	},
})
if err != nil {
	log.Fatal(err)
}

var payload struct {
	Items []map[string]any `json:"items"`
	Total int              `json:"total"`
}
if err := res.DecodeData(&payload); err != nil {
	log.Fatal(err)
}
log.Printf("code=%s total=%d items=%d", res.Code, payload.Total, len(payload.Items))
```

***

## **🔍 查询接口**

查询条件请根据实际需求自行拼装。详情参考 API 接口文档示例。

### **单条查询**

```go
res, err := client.Object.Search.Record(ctx, apaas.ObjectSearchRecordParams{
	ObjectName: "object_store",
	RecordID:   "your_record_id",
	Select:     []string{"field1", "field2"},
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

***

### **批量查询**

每次查询最多返回 100 条记录。

```go
res, err := client.Object.Search.Records(ctx, apaas.ObjectSearchRecordsParams{
	ObjectName: "object_store",
	Data: map[string]any{
		"need_total_count": true,
		"page_size":        100,
		"offset":           0,
	},
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

***

### **分页查询所有记录**

在上一个请求的基础上，封装每次查询最多返回 100 条记录。

```go
result, err := client.Object.Search.RecordsWithIterator(ctx, apaas.ObjectRecordsIteratorParams{
	ObjectName: "object_store",
	Data: map[string]any{
		"need_total_count": true,
		"page_size":        100,
	},
})
if err != nil {
	log.Fatal(err)
}
log.Printf("Total: %d, Items: %d", result.Total, len(result.Items))
```

***



## **➕ 创建接口**

### **单条创建**

```go
res, err := client.Object.Create.Record(ctx, apaas.ObjectCreateRecordParams{
	ObjectName: "object_event_log",
	Record: map[string]any{
		"name":    "Sample text",
		"content": "Sample text",
	},
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

### **批量创建**

> ⚠️ 每次最多创建 100 条，SDK 提供 `Records` 与 `RecordsWithIterator` 两种方式。

```go
res, err := client.Object.Create.Records(ctx, apaas.ObjectCreateRecordsParams{
	ObjectName: "object_event_log",
	Records: []map[string]any{
		{"name": "Sample text 1", "content": "Sample text 1"},
		{"name": "Sample text 2", "content": "Sample text 2"},
	},
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

### **批量创建（自动拆分）**

支持超过 100 条数据，SDK 已自动分组限流。返回详细的成功/失败统计。

```go
result, err := client.Object.Create.RecordsWithIterator(ctx, apaas.ObjectCreateRecordsIteratorParams{
	ObjectName: "object_event_log",
	Records: []map[string]any{
		{"name": "Sample text 1"},
		{"name": "Sample text 2"},
		// ... 可以超过 100 条
	},
	Limit: 100, // 可选，默认 100
})
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Total: %d\n", result.Total)
fmt.Printf("Success: %d\n", result.SuccessCount)
fmt.Printf("Failed: %d\n", result.FailedCount)

// 查看失败的记录
for _, failed := range result.Failed {
	fmt.Printf("Failed ID: %s, Error: %s\n", failed.ID, failed.Error)
}
```

***



## **✏️ 更新接口**

### **单条更新**

```go
res, err := client.Object.Update.Record(ctx, apaas.ObjectUpdateRecordParams{
	ObjectName: "object_store",
	RecordID:   "your_record_id",
	Record: map[string]any{
		"field1": "newValue",
	},
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

***

### **批量更新**

> ⚠️ 每次最多更新 100 条，超出请使用自动拆分方法。

```go
res, err := client.Object.Update.Records(ctx, apaas.ObjectUpdateRecordsParams{
	ObjectName: "object_store",
	Records: []map[string]any{
		{"_id": "id1", "field1": "value1"},
		{"_id": "id2", "field1": "value2"},
	},
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

***

### **批量更新（自动拆分）**

支持超过 100 条数据，SDK 已自动分组限流。返回详细的成功/失败统计。

```go
result, err := client.Object.Update.RecordsWithIterator(ctx, apaas.ObjectUpdateRecordsIteratorParams{
	ObjectName: "object_store",
	Records: []map[string]any{
		{"_id": "id1", "field1": "value1"},
		{"_id": "id2", "field1": "value2"},
		// ... 可以超过 100 条
	},
	Limit: 100, // 可选，默认 100
})
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Total: %d\n", result.Total)
fmt.Printf("Success: %d\n", result.SuccessCount)
fmt.Printf("Failed: %d\n", result.FailedCount)

// 查看失败的记录
for _, failed := range result.Failed {
	fmt.Printf("Failed ID: %s, Error: %s\n", failed.ID, failed.Error)
}
```

***



## **🗑️ 删除接口**

### **单条删除**

```go
res, err := client.Object.Delete.Record(ctx, apaas.ObjectDeleteRecordParams{
	ObjectName: "object_store",
	RecordID:   "your_record_id",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

***

### **批量删除**

> ⚠️ 每次最多删除 100 条，超出请使用自动拆分方法。

```go
res, err := client.Object.Delete.Records(ctx, apaas.ObjectDeleteRecordsParams{
	ObjectName: "object_store",
	IDs:        []string{"id1", "id2", "id3"},
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

***

### **批量删除（自动拆分）**

支持超过 100 条数据，SDK 已自动分组限流。返回详细的成功/失败统计。

```go
result, err := client.Object.Delete.RecordsWithIterator(ctx, apaas.ObjectDeleteRecordsIteratorParams{
	ObjectName: "object_store",
	IDs:        []string{"id1", "id2", "id3", /* ... 可以超过 100 个 */},
	Limit:      100, // 可选，默认 100
})
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Total: %d\n", result.Total)
fmt.Printf("Success: %d\n", result.SuccessCount)
fmt.Printf("Failed: %d\n", result.FailedCount)

// 查看失败的记录
for _, failed := range result.Failed {
	fmt.Printf("Failed ID: %s, Error: %s\n", failed.ID, failed.Error)
}
```

***



## **📊 对象元数据接口**

### **获取指定对象字段元数据**

```go
res, err := client.Object.Metadata.Field(ctx, apaas.ObjectMetadataFieldParams{
	ObjectName: "_user",
	FieldName:  "_id",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

### **获取指定对象所有字段信息**

```go
res, err := client.Object.Metadata.Fields(ctx, apaas.ObjectMetadataFieldsParams{
	ObjectName: "object_store",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

***



# **📎 附件模块**

## **文件操作**

### **上传文件**

```go
file, err := os.Open("/path/to/file.zip")
if err != nil {
	log.Fatal(err)
}
defer file.Close()

res, err := client.Attachment.File.Upload(ctx, apaas.AttachmentFileUploadParams{
	FileName:    "file.zip",
	Reader:      file,
	ContentType: "application/zip",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

### **下载文件**

```go
data, err := client.Attachment.File.Download(ctx, apaas.AttachmentFileDownloadParams{
	FileID: "625d2f602af94d46972073db32a99ed2",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("downloaded=%d bytes", len(data))
```

### **删除文件**

```go
res, err := client.Attachment.File.Delete(ctx, apaas.AttachmentFileDeleteParams{
	FileID: "625d2f602af94d46972073db32a99ed2",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

## **头像图片操作**

### **上传头像图片**

```go
image, err := os.Open("/path/to/avatar.jpg")
if err != nil {
	log.Fatal(err)
}
defer image.Close()

res, err := client.Attachment.Avatar.Upload(ctx, apaas.AttachmentAvatarUploadParams{
	FileName:    "avatar.jpg",
	Reader:      image,
	ContentType: "image/jpeg",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

### **下载头像图片**

```go
data, err := client.Attachment.Avatar.Download(ctx, apaas.AttachmentAvatarDownloadParams{
	ImageID: "c70d03b21d3c40468ee710d984cfb7a8_o",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("downloaded=%d bytes", len(data))
```

***



# **💽 全局数据模块**

## **全局选项**

### **查询全局选项详情**

```go
res, err := client.Global.Options.Detail(ctx, "global_option_abc")
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

### **查询全局选项列表**

```go
res, err := client.Global.Options.List(ctx, 10, 0, map[string]any{
	"quickQuery": "Sample Text",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

### **分页查询所有全局选项**

```go
result, err := client.Global.Options.ListWithIterator(ctx, 100, map[string]any{
	"quickQuery": "Sample Text",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("Total: %d, Items: %d", result.Total, len(result.Items))
```

## **环境变量**

### **查询环境变量详情**

```go
res, err := client.Global.Variables.Detail(ctx, "global_variable_abc")
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

### **查询环境变量列表**

```go
res, err := client.Global.Variables.List(ctx, 10, 0, map[string]any{
	"quickQuery": "Sample Text",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

### **分页查询所有环境变量**

```go
result, err := client.Global.Variables.ListWithIterator(ctx, 100, map[string]any{
	"quickQuery": "Sample Text",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("Total: %d, Items: %d", result.Total, len(result.Items))
```

***



# **📄 页面模块**

### **获取所有页面**

```go
res, err := client.Page.List(ctx, apaas.PageListParams{
	Limit:  10,
	Offset: 0,
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

### **分页查询所有页面**

```go
result, err := client.Page.ListWithIterator(ctx, &apaas.PageListWithIteratorParams{
	Limit: 100,
})
if err != nil {
	log.Fatal(err)
}
log.Printf("Total: %d, Items: %d", result.Total, len(result.Items))
```

### **获取页面详情**

```go
res, err := client.Page.Detail(ctx, apaas.PageDetailParams{
	PageID: "appPage_page",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

### **获取页面访问地址**

```go
res, err := client.Page.URL(ctx, apaas.PageURLParams{
	PageID: "appPage_page",
	PageParams: map[string]any{
		"var_page": "1234567890",
	},
	ParentPageParams: map[string]any{
		"navId":       "page_nav_id",
		"pageApiName": "page_name",
	},
	NavID: "page_nav_id",
	TabID: "tab_id",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

***



# **🏢 部门模块**

## **部门 ID 交换**

### **单个部门 ID 交换**

```go
result, err := client.Department.Exchange(ctx, apaas.DepartmentExchangeParams{
	DepartmentIDType: "external_department_id",
	DepartmentID:     "Y806608904",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("result=%v", result)
```

### **批量部门 ID 交换**

每次最多 100 个，SDK 已自动拆分限流。

```go
results, err := client.Department.BatchExchange(ctx, apaas.DepartmentBatchExchangeParams{
	DepartmentIDType: "external_department_id",
	DepartmentIDs:    []string{"id1", "id2", "id3"},
})
if err != nil {
	log.Fatal(err)
}
log.Printf("count=%d", len(results))
```

***

<br>

# **🔄 自动化流程模块**

## **V1 版本 - 执行流程**

```go
res, err := client.Automation.V1.Execute(ctx, apaas.AutomationV1ExecuteParams{
	FlowAPIName: "automation_cd05fdab67d",
	Operator: apaas.FlowOperator{
		ID:    100,
		Email: "sample@feishu.cn",
	},
	Params: map[string]any{
		"varRecord_ab67d031d44": map[string]any{
			"_id": 100,
		},
	},
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

## **V2 版本 - 执行流程**

V2 版本支持流程重新提交功能。

```go
isResubmit := true

res, err := client.Automation.V2.Execute(ctx, apaas.AutomationV2ExecuteParams{
	FlowAPIName: "automation_a9ec6ee5fb1",
	Operator: apaas.FlowOperator{
		ID:    100,
		Email: "sample@feishu.cn",
	},
	Params: map[string]any{
		"storeId": 100,
	},
	IsResubmit:    &isResubmit,
	PreInstanceID: "1835957428957195",
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

***

<br>

# **☁️ 云函数模块**

```go
res, err := client.Function.Invoke(ctx, apaas.FunctionInvokeParams{
	Name: "StoreMemberUpdate",
	Params: map[string]any{
		"key": "value",
	},
})
if err != nil {
	log.Fatal(err)
}
log.Printf("code=%s", res.Code)
```

***

<br>

## **🛠️ 高级**

### **获取当前 token**

```go
log.Println(client.Token())
```

### **获取 token 过期时间**

```go
if ttl, ok := client.TokenExpiresIn(); ok {
	log.Printf("token expires in %s", ttl)
} else {
	log.Println("no valid token cached")
}
```

### **获取当前 namespace**

```go
log.Println(client.Namespace())
```

***



## **💡 备注**

- 本 SDK 默认使用标准库 `net/http` 发起请求，可通过 `ClientOptions.HTTPClient` 自定义。
- 基于 `golang.org/x/time/rate` 实现请求限流，可通过 `ClientOptions.LimiterOptions` 调整。
- 默认日志实现输出到标准输出，支持自定义 `Logger` 接口以满足更多需求。


***



> 由 aPaaS OpenAPI Go Client SDK 提供支持，如有问题请提交 Issue 反馈。

---
