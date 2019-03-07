package zsec_go_examples_test

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestMarshalEmptyMap(t *testing.T) {
	out, err := json.Marshal([]interface{}{})
	_ = err
	fmt.Printf("Empty slice marshals to: %v\n", string(out))

	var ss []interface{}
	out, err = json.Marshal(ss)
	_ = err
	fmt.Printf("Null slice marshals to: %v\n", string(out))
}

func TestMarshalFormatted(t *testing.T) {

	mm := make(map[string]string)
	for _, aa := range []string{"a", "b", "c"} {
		mm[aa] = aa
	}

	var records []interface{}
	for ii := 0; ii < 3; ii++ {
		records = append(records, mm)
	}

	out, _ := json.Marshal(records)
	fmt.Printf("Without indent, records marshals to:\n%v\n", string(out))

	println()
	out, _ = json.MarshalIndent(records, "", "    ")
	fmt.Printf("With indent, records marshals to:\n%v\n", string(out))
}
