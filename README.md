# apaas-oapi-go-client

🚀 **aPaaS OpenAPI Go 客户端 SDK**

封装 aPaaS 平台 RESTful API 的 Go SDK，简化接口调用，内置限流与 token 缓存功能。

---

## ✨ **功能特性**

- ✅ 获取 accessToken，自动刷新与缓存
- ✅ records 查询（支持分页迭代）
- ✅ record 单条创建、更新、删除
- ✅ 批量创建 / 更新 / 删除（自动分片）
- ✅ 页面、附件、全局变量等模块能力
- ✅ 内置基于 `golang.org/x/time/rate` 的限流器
- ✅ 可自定义日志等级

---

## 📦 **安装**

```bash
go get github.com/apaas/apaas-sdk/go-client/apaas
```

---

## 🚀 **快速开始**

```go
package main

import (
    "context"
    "log"

    "github.com/apaas/apaas-sdk/go-client/apaas"
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

    client.SetLoggerLevel(apaas.LoggerLevelInfo)

    ctx := context.Background()
    if err := client.Init(ctx); err != nil {
        log.Fatalf("init client: %v", err)
    }

    res, err := client.Object.List(ctx, apaas.ObjectListParams{
        Offset: 0,
        Limit:  20,
    })
    if err != nil {
        log.Fatalf("list objects: %v", err)
    }

    log.Printf("request finished, code=%s, msg=%s", res.Code, res.Msg)
}
```

更多使用示例请查阅 `UserManual.md` 与 `examples/`。

