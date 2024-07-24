# Condition Value Go Pacakge

`condval` 是一个用于根据条件表达式计算结果的Go包。它支持动态条件、嵌套条件和动态计算结果。

可以应用于规则易变，而输入变量固定的场景，不需要构建代码，只需要修改配置就可以更新对应逻辑。

## 使用方法

1. 从JSON文本（文件或字节数组）中加载条件值的配置
2. 使用 `GetResult` 方法计算结果，或使用 `GetResultWithTrace` 方法计算结果并记录命中条件的路径

配置有单独的小节进行说明，使用之前请认真阅读[注意事项](#注意事项)。

### 示例

```go
package main

import (
	"fmt"
	"github.com/wlbgo/condval"
)

func main() {
	cvc, err := condval.ParseConditionValueConfig([]byte(`
[
  { "condition": "va<=1000",
    "result": [ { "condition": "va - 300 > va_upper", "result": 101 },
      { "condition": "va - 200 < va_lower", "result": 102 },
      { "condition": "va - 300 >= va_lower", "result": "103" },
      { "condition": "va - 300 < va_lower", "result": 104 }
    ]
  },
  { "condition": "true",
    "result": [ { "condition": "va - 200 > va_upper", "result": 201 },
      { "condition": "va - 100 < va_lower", "result": 202 },
      { "condition": "va - 200 >= va_lower", "result": 203 },
      { "condition": "va - 200 < va_lower", "result": 204 }
    ]
  }
]

`))
	if err != nil {
		panic(err)
	}

	parameters := map[string]interface{}{"va": 1000, "va_upper": 1200, "va_lower": 800}
	got, trace, err := cvc.GetResultWithTrace(parameters)
	fmt.Printf("err: %v, result: %v, trace: %v", err, got, trace)
}

```

运行上述代码，将会得到下边的输出:

```
err: <nil>, result: 104, trace: [0 3]
```

输出的含义错误为 `nil`，结果为 `104`，路径为 `[0 3]` 即命中 [demo.json](demo/demo.json) 第一个条件中的第三条。

---

可以直接在 demo 中找到一个完整demo，运行下边的命令即可：

```shell
cd demo
go run .
```

## 配置说明

配置 schema 详见 [condition-value-config-schema.json](condition-value-config-schema.json)。

这里做简单的说明，根节点的是一个数组，我们把他叫做 `ConditionValueConfig`，每个元素是一个 `ConditionValue` 对象，包含两个属性：

- `condition`：条件表达式，使用字符串表示。
- `result`：
    - 条件满足时返回的结果，可以是任意类型
    - 如果存在条件嵌套，`result` 应该使用是一个 `ConditionValueConfig`，即嵌套一个配置数组。

### 配置示例

#### 1. 简单条件

```json
[
  {
    "condition": "a > 1",
    "result": 2
  }
]
```

#### 2. 始终为真

```json
[
  {
    "condition": "true",
    "result": 2
  }
]
```

#### 3. 与条件

```json
[
  {
    "condition": "a > 1 && b < 0",
    "result": 2
  }
]
```

#### 4. 结果中使用变量

```json
[
  {
    "condition": "a > 1 && b < 0",
    "result": "a*2"
  }
]
```

#### 5. 嵌套条件

```json
[
  {
    "condition": "a > 1 && b < 0",
    "result": [
      {
        "condition": "a >= 2",
        "result": 4
      }
    ]
  }
]
```

#### 6. 字符串

如果返回结果为字符串，会先对其进行运算操作，如果存在变量，会替换为变量的值。

因此，如果要使用跟传入变量冲突的字符串，需要将其进行转义：

```json
[
  {
    "condition": "a > 1",
    "result": "abc"
  },
  {
    "condition": "true",
    "result": "\"a\""
  }
]
```

## 注意事项

### 条件匹配只命中第一个

我们把配置数组中的每个元素叫做 `ConditionValue`，在计算结果时，只会命中第一个满足条件的 `ConditionValue`。

特殊情况下，如果命中 condition1 之后，其结果为嵌套的 `ConditionValueConfig`，但是子结果中没有满足条件的 `ConditionValue`
，则会错误 `ErrNoCond`。

即使外层的有满足条件的 `ConditionValue`，也不会继续。

这种处理逻辑跟 ``if`` 的嵌套类似：

```
if case1 {
  if subCase1 {
    return result1
  }
} else if case2 {
    return result2
} else {
    return result3
}
return resultUnfound
```

### 字符串的处理

配置中的result如果是字符串，则会优先进行**运算**，即：将其当做一个表达式进行计算。
比如 "true" 会被计算为布尔型 `true`；又如表达式中有变量，则会进行对应的计算得到结果。

如果要返回确定的字符串，则需要进行转义，如 `"\"a\""`，会返回字符串 `"a"`，而不是变量 `a` 的值。

### 整数的处理

配置中的 result 如果是整数或者是整数的表达式，会被当做 `float64` 处理，因此如果进行整数运算，可能不符合预期，需要主动进行干预。

下边的结果会得到 `1.05`，而不是 `1`：

```go
package main

import (
	"fmt"
	"github.com/wlbgo/condval"
)

func main() {
	cvc, err := condval.ParseConditionValueConfig([]byte(`[{"condition": "true", "result": "a/20"}]`))
	if err != nil {
		panic(err)
	}

	parameters := map[string]interface{}{"a": 21}
	got, trace, err := cvc.GetResultWithTrace(parameters)
	fmt.Printf("err: %v, result: %v, trace: %v", err, got, trace)
}
```

## TODO benchmark



