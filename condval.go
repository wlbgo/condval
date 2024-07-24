// Package condval provides functionality to evaluate conditions and return corresponding results.
package condval

import (
	"encoding/json"
	"fmt"
	"github.com/expr-lang/expr"
	"os"
)

// Error definitions
var (
	ErrNoCond    = fmt.Errorf("no condition matched")
	ErrCompile   = fmt.Errorf("failed to compile expression")
	ErrRunCode   = fmt.Errorf("failed to run expression")
	ErrLoadFile  = fmt.Errorf("failed to read file")
	ErrParseJson = fmt.Errorf("failed to unmarshal JSON")
	ErrSubResult = fmt.Errorf("failed to marshal sub-result")
)

// ConditionValue represents a condition and its corresponding result.
type ConditionValue struct {
	ConditionExpr string      `json:"condition"`
	Result        interface{} `json:"result"`
}

// ConditionValueConfig is a slice of ConditionValue.
type ConditionValueConfig []ConditionValue

// GetResult evaluates the conditions in the configuration and returns the result of the first matched condition.
func (cvc ConditionValueConfig) GetResult(parameters map[string]interface{}) (interface{}, error) {
	ret, _, err := cvc.GetResultWithTrace(parameters)
	return ret, err
}

// GetResultWithTrace evaluates the conditions in the configuration and returns the result of the first matched condition along with the trace of evaluated conditions.
func (cvc ConditionValueConfig) GetResultWithTrace(parameters map[string]interface{}) (interface{}, []int, error) {
	trace := make([]int, 0)

	tryRunCode := func(code string) (interface{}, error) {
		program, err := expr.Compile(code, expr.Env(parameters))
		if err != nil {
			return nil, ErrCompile
		}
		result, err := expr.Run(program, parameters)
		if err != nil {
			return nil, ErrRunCode
		}
		return result, nil
	}

	for i, cv := range cvc {
		code := cv.ConditionExpr
		result, err := tryRunCode(code)
		if err != nil {
			return nil, nil, err
		}
		if result.(bool) {
			trace = append(trace, i)
			if subCvc, ok := cv.Result.(ConditionValueConfig); ok {
				subRet, subTrace, subErr := subCvc.GetResultWithTrace(parameters)
				if subErr != nil {
					return nil, nil, subErr
				}
				trace = append(trace, subTrace...)
				return subRet, trace, nil
			} else {
				if subExpr, ok := cv.Result.(string); ok {
					result, err := tryRunCode(subExpr)
					if err == nil {
						return result, trace, nil
					}
				}
				return cv.Result, trace, nil
			}
		}
	}
	return nil, nil, ErrNoCond
}

// Equal checks if two ConditionValueConfig instances are equal.
func (cvc ConditionValueConfig) Equal(another ConditionValueConfig) bool {
	if len(cvc) != len(another) {
		fmt.Println("lengths not equal")
		return false
	}

	for i, cv := range cvc {
		if cv.ConditionExpr != another[i].ConditionExpr {
			fmt.Println("conditions not equal")
			return false
		}
		if v1, ok := cv.Result.(ConditionValueConfig); ok {
			v2, ok := another[i].Result.(ConditionValueConfig)
			if !ok {
				fmt.Println("result types not equal")
				return false
			}
			if !v1.Equal(v2) {
				fmt.Println("nested results not equal")
				return false
			}
		} else if cv.Result != another[i].Result {
			fmt.Printf("results not equal: %v != %v\n", cv.Result, another[i].Result)
			return false
		}
	}
	return true
}

// ParseConditionValueConfigFile reads a JSON file and parses it into a ConditionValueConfig.
func ParseConditionValueConfigFile(path string) (ConditionValueConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, ErrLoadFile
	}
	return ParseConditionValueConfig(raw)
}

// ParseConditionValueConfig parses a JSON byte slice into a ConditionValueConfig.
func ParseConditionValueConfig(raw []byte) (ConditionValueConfig, error) {
	var cvc ConditionValueConfig

	rawCvc := make([]struct {
		ConditionExpr string      `json:"condition"`
		Result        interface{} `json:"result"`
	}, 0)

	err := json.Unmarshal(raw, &rawCvc)
	if err != nil {
		return nil, ErrParseJson
	}

	for _, cv := range rawCvc {
		if val, ok := cv.Result.([]interface{}); ok {
			subJson, err := json.Marshal(val)
			if err != nil {
				return nil, ErrSubResult
			}
			subRet, err := ParseConditionValueConfig(subJson)
			if err != nil {
				return nil, ErrSubResult
			}
			cvc = append(cvc, ConditionValue{
				ConditionExpr: cv.ConditionExpr,
				Result:        subRet,
			})
		} else {
			cvc = append(cvc, ConditionValue{
				ConditionExpr: cv.ConditionExpr,
				Result:        cv.Result,
			})
		}
	}

	return cvc, nil
}
