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
		{"TestSimple", ConditionValueConfig{ConditionValue{"a > 1", float64(2)}}, args{map[string]interface{}{"a": 2}}, float64(2)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := tt.cvc.GetResult(tt.args.parameters); err != nil || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConditionValueConfig_GetResultFromJson(t *testing.T) {
	cvc, err := ParseConditionValueConfigFile("test.json")
	// 覆盖率测试
	traceMap := make(map[string]bool)
	traceToStr := func(trace []int) string {
		str := ""
		for _, i := range trace {
			str += fmt.Sprintf("%d-", i)
		}
		return str
	}

	t.Run("", func(t *testing.T) {
		for mmr := 0; mmr < 2000; mmr += 30 {
			for mmr_lower := 0; mmr_lower < 2000; mmr_lower += 40 {
				for mmr_upper := mmr_lower; mmr_upper < 2300; mmr_upper += 35 {
					if err != nil {
						t.Errorf("ParseConditionValueConfig() error = %v", err)
					}
					parameters := map[string]interface{}{"mmr": mmr, "mmr_upper": mmr_upper, "mmr_lower": mmr_lower}
					if got, trace, err := cvc.GetResultWithTrace(parameters); err != nil {
						t.Errorf("err: %v, GetResult() = %v", err, got)
					} else {
						traceMap[traceToStr(trace)] = true
					}
				}
			}
		}
		fmt.Printf("traceMap: %v\n", traceMap)
		if len(traceMap) != 4*9 {
			t.Errorf("traceMap: %v", traceMap)
		}
	})
}

func FuzzConditionValueConfig_GetResultFromJson(f *testing.F) {
	content, err := os.ReadFile("test.json")
	if err != nil {
		panic(err)
	}
	cvc, err := ParseConditionValueConfig(content)
	f.Add(1078, 1200, 1000)
	f.Fuzz(func(t *testing.T, mmr, mmr_upper, mmr_lower int) {
		parameters := map[string]interface{}{"mmr": mmr, "mmr_upper": mmr_upper, "mmr_lower": mmr_lower}
		out, trace, err := cvc.GetResultWithTrace(parameters)
		_, ok := out.(float64)

		if err != nil || !ok {
			t.Errorf("err: %v", err)
		} else {
			t.Logf("trace: %v", trace)
		}
	})
}
