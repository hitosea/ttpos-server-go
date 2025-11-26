# 类型使用示例

> Go Main 模块中的类型使用规范示例代码

---

## 使用 any 替代 interface{}

从 Go 1.18 开始，`any` 是 `interface{}` 的内置类型别名，代码更简洁易读。

### ✅ 正确示例

```go
// 函数参数使用 any
func Process(data any) error {
    return nil
}

// 变量声明使用 any
var result any = "hello"

// map 值使用 any
var config map[string]any = map[string]any{
    "host": "localhost",
    "port": 3306,
}

// 返回值使用 any
func GetValue(key string) (any, error) {
    return "value", nil
}

// 结构体字段使用 any
type Message struct {
    Type    string `json:"type"`
    Payload any    `json:"payload"`
}
```

### ❌ 错误示例

```go
// 函数参数使用 interface{}
func Process(data interface{}) error {
    return nil
}

// 变量声明使用 interface{}
var result interface{} = "hello"

// map 值使用 interface{}
var config map[string]interface{} = map[string]interface{}{
    "host": "localhost",
    "port": 3306,
}

// 返回值使用 interface{}
func GetValue(key string) (interface{}, error) {
    return "value", nil
}

// 结构体字段使用 interface{}
type Message struct {
    Type    string      `json:"type"`
    Payload interface{} `json:"payload"`
}
```

---

## 实际应用场景

### 场景 1: JSON 响应中的动态数据

```go
type ApiResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data"` // 使用 any 而不是 interface{}
}

func Success(data any) ApiResponse {
    return ApiResponse{
        Code:    0,
        Message: "success",
        Data:    data,
    }
}
```

### 场景 2: 配置解析

```go
type Config struct {
    Database map[string]any `json:"database"` // 使用 any 而不是 interface{}
    Redis    map[string]any `json:"redis"`    // 使用 any 而不是 interface{}
}

func LoadConfig(path string) (*Config, error) {
    var config Config
    // 解析配置文件
    return &config, nil
}
```

### 场景 3: 泛型数据处理

```go
// 处理任意类型的数据
func ToJSON(data any) ([]byte, error) { // 使用 any 而不是 interface{}
    return json.Marshal(data)
}

func FromJSON(jsonData []byte, target any) error { // 使用 any 而不是 interface{}
    return json.Unmarshal(jsonData, target)
}
```

---

## 相关文档

- [Go Main 核心约束](../../../../.cursor/rules/go-main.mdc) - 类型使用规范
- [常用模式示例](./common-patterns-examples.md) - 其他开发模式

---

**最后更新**: 2025-11-26  
**维护者**: TTPOS Team

