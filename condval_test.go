package condval

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestParseConditionValueConfig(t *testing.T) {
	type args struct {
		raw []byte
	}
	tests := []struct {
		name    string
		args    args
		want    ConditionValueConfig
		wantErr bool
	}{
		{"TestSimple", args{[]byte(`[{"condition":"a > 1","result":2}]`)}, ConditionValueConfig{ConditionValue{"a > 1", float64(2)}}, false},
		{"TestNested", args{[]byte(`[{"condition":"a > 1","result": [{"condition":"a > 2","result": 2}] }]`)}, ConditionValueConfig{ConditionValue{"a > 1", ConditionValueConfig{ConditionValue{"a > 2", float64(2)}}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConditionValueConfig(tt.args.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConditionValueConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("ParseConditionValueConfig() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestConditionValueConfig_GetResult(t *testing.T) {
	type args struct {
		parameters map[string]interface{}
	}
	tests := []struct {
		name string
		cvc  ConditionValueConfig
		args args
		want interface{}
	}{
		{"TestSimple", ConditionValueConfig{ConditionValue{"a > 1", 2}}, args{map[string]interface{}{"a": 2}}, 2},
		{"TestAlwaysTrue", ConditionValueConfig{ConditionValue{"true", 2}}, args{map[string]interface{}{}}, 2},
		{"TestAndCondition", ConditionValueConfig{ConditionValue{"a > 1 && b < 0", 2}}, args{map[string]interface{}{"a": 2, "b": -1}}, 2},
		{"TestVarInResult", ConditionValueConfig{ConditionValue{"a > 1 && b < 0", "a*2"}}, args{map[string]interface{}{"a": 2, "b": -1}}, 4},
		{"TestNested", ConditionValueConfig{ConditionValue{"a > 1 && b < 0",
			ConditionValueConfig{ConditionValue{"a >=2", 4}},
		}}, args{map[string]interface{}{"a": 2, "b": -1}}, 4},
		{"TestNoCondition", ConditionValueConfig{ConditionValue{"a > 1", 2}}, args{map[string]interface{}{"a": 0}}, ErrNoCond},
		{"TestStringResult", ConditionValueConfig{ConditionValue{"a > 1", "abc"}}, args{map[string]interface{}{"a": 2}}, "abc"},
		{"TestStringResult2", ConditionValueConfig{ConditionValue{"a > 1", "\"a\""}}, args{map[string]interface{}{"a": 2}}, "a"},
		{"TestIntFloat1", ConditionValueConfig{ConditionValue{"true", "a"}}, args{map[string]interface{}{"a": 2200}}, 2200},
		// 下边两个测试用例，会将 a 转换成float64 进行计算
		//{"TestIntFloat2", ConditionValueConfig{ConditionValue{"true", "a/1000"}}, args{map[string]interface{}{"a": 2200}}, 2200 / 1000},
		//{"TestIntFloat2", ConditionValueConfig{ConditionValue{"true", "int(a)/int(1000)"}}, args{map[string]interface{}{"a": 2200}}, 2200 / 1000},
		{"TestIntFloat2", ConditionValueConfig{ConditionValue{"true", "a/1000"}}, args{map[string]interface{}{"a": 2200}}, 2200.0 / 1000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cvc.GetResult(tt.args.parameters)
			if wantedErr, ok := tt.want.(error); (ok && wantedErr != err) || (!ok && (err != nil || !reflect.DeepEqual(got, tt.want))) {
				t.Errorf("GetResult() = %v(%T), %v, want %v", got, got, err, tt.want)
			}
		})
	}
}
func demo() {
	cvc, err := ParseConditionValueConfigFile("demo/demo.json")
	if err != nil {
		panic(err)
	}

	parameters := map[string]interface{}{"va": 1000, "va_upper": 1200, "va_lower": 800}
	got, trace, err := cvc.GetResultWithTrace(parameters)
	fmt.Printf("err: %v, result: %v, trace: %v", err, got, trace)
}

func TestConditionValueConfig_GetResultFromJson(t *testing.T) {
	cvc, err := ParseConditionValueConfigFile("demo/demo.json")
	// 覆盖率测试
	traceMap := make(map[string]bool)
	resultMap := make(map[interface{}]bool)
	traceToStr := func(trace []int) string {
		str := ""
		for _, i := range trace {
			str += fmt.Sprintf("%d-", i)
		}
		return str
	}

	t.Run("", func(t *testing.T) {
		for va := 0; va < 2000; va += 30 {
			for vaLower := 0; vaLower < 2000; vaLower += 40 {
				for vaUpper := vaLower; vaUpper < 2300; vaUpper += 35 {
					if err != nil {
						t.Errorf("ParseConditionValueConfig() error = %v", err)
					}
					parameters := map[string]interface{}{"va": va, "va_upper": vaUpper, "va_lower": vaLower}
					if got, trace, err := cvc.GetResultWithTrace(parameters); err != nil {
						t.Errorf("err: %v, GetResult() = %v", err, got)
					} else {
						traceMap[traceToStr(trace)] = true
						resultMap[got] = true
					}
				}
			}
		}
		t.Logf("traceMap: %v\n", traceMap)
		t.Logf("resultMap: %v\n", resultMap)
	})
}

func FuzzConditionValueConfig_GetResultFromJson(f *testing.F) {
	content, err := os.ReadFile("demo/demo.json")
	if err != nil {
		panic(err)
	}
	cvc, err := ParseConditionValueConfig(content)
	f.Add(1078, 1200, 1000)
	f.Fuzz(func(t *testing.T, va, va_upper, va_lower int) {
		parameters := map[string]interface{}{"va": va, "va_upper": va_upper, "va_lower": va_lower}
		out, _, err := cvc.GetResultWithTrace(parameters)
		_, ok1 := out.(float64)
		_, ok2 := out.(int)

		if err != nil || (!ok1 && !ok2) {
			t.Errorf("err: %v, out: %v(%T)", err, out, out)
		}
	})
}
