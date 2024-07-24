package main

import (
	"fmt"
	"github.com/wlbgo/condval"
)

func main() {
	cvc, err := condval.ParseConditionValueConfigFile("demo.json")
	if err != nil {
		panic(err)
	}

	parameters := map[string]interface{}{"va": 1000, "va_upper": 1200, "va_lower": 800}
	got, trace, err := cvc.GetResultWithTrace(parameters)
	fmt.Printf("err: %v, result: %v, trace: %v", err, got, trace)
}
