# Condition Value Go Package

[中文简体](README.zh_CN.md)

`condval` is a Go package for evaluating results based on conditional expressions. It supports dynamic conditions,
nested conditions, and dynamic result evaluation.

This package is useful in scenarios where rules are variable but input variables are fixed. You can update the
corresponding logic by modifying the configuration without needing to rebuild the code.

## Usage

1. Load the configuration of condition values from JSON text (file or byte array).
2. Use the `GetResult` method to compute the result, or use the `GetResultWithTrace` method to compute the result and
   record the path of the matched conditions.

The configuration is explained in a separate section. Please read the [Notes](#notes) carefully before using it.

### Example

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

Running the above code will produce the following output:

```
err: <nil>, result: 104, trace: [0 3]
```

The output means the error is `nil`, the result is `104`, and the path is `[0 3]`, which matches the third condition in
the first condition set of [demo.json](demo/demo.json).

---

You can find a complete demo in the `demo` directory. Run the following command to execute it:

```shell
cd demo
go run .
```

## Configuration Explanation

The configuration schema is detailed in [condition-value-config-schema.json](condition-value-config-schema.json).

Here is a brief explanation: the root node is an array, which we call `ConditionValueConfig`. Each element is
a `ConditionValue` object containing two properties:

- `condition`: The condition expression, represented as a string.
- `result`:
    - The result returned when the condition is met, which can be of any type.
    - If there are nested conditions, `result` should be a `ConditionValueConfig`, i.e., a nested configuration array.

### Configuration Examples

#### 1. Simple Condition

```json
[
  {
    "condition": "a > 1",
    "result": 2
  }
]
```

#### 2. Always True

```json
[
  {
    "condition": "true",
    "result": 2
  }
]
```

#### 3. AND Condition

```json
[
  {
    "condition": "a > 1 && b < 0",
    "result": 2
  }
]
```

#### 4. Using Variables in Result

```json
[
  {
    "condition": "a > 1 && b < 0",
    "result": "a*2"
  }
]
```

#### 5. Nested Conditions

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

#### 6. Strings

If the result is a string, it will be evaluated first. If there are variables, they will be replaced with the variable
values.

Therefore, if you want to use a string that conflicts with the input variables, you need to escape it:

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

## Notes

### Only the First Matching Condition is Hit

We call each element in the configuration array a `ConditionValue`. When computing the result, only the first
matching `ConditionValue` will be hit.

In special cases, if `condition1` is hit and its result is a nested `ConditionValueConfig`, but no `ConditionValue` in
the sub-result matches, an `ErrNoCond` error will occur.

Even if there are other matching `ConditionValue` in the outer layer, they will not be evaluated further.

This logic is similar to nested `if` statements:

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

### Handling Strings

If the result in the configuration is a string, it will be evaluated first, i.e., treated as an expression. For
example, "true" will be evaluated as the boolean `true`; if there are variables in the expression, they will be
evaluated accordingly.

To return a specific string, you need to escape it, such as `"\"a\""`, which will return the string `"a"` instead of the
value of the variable `a`.

### Handling Integers

If the result in the configuration is an integer or an integer expression, it will be treated as `float64`. Therefore,
integer operations may not meet expectations and may require manual intervention.

The following result will be `1.05` instead of `1`:

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

## TODO Benchmark
