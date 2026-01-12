package vue

import "testing"

func TestIntArrayToString(t *testing.T) {
	got := intArrayToString([]int{1, 2, 3})
	if got != "1,2,3" {
		t.Fatalf("expected '1,2,3', got %s", got)
	}
}

func TestIsNil(t *testing.T) {
	var nilSlice []string
	var nilMap map[string]string
	var nilChan chan int
	var cfgPtr *Config

	cases := []struct {
		name   string
		value  interface{}
		expect bool
	}{
		{"nil_interface", nil, true},
		{"nil_slice", nilSlice, true},
		{"nil_map", nilMap, true},
		{"nil_chan", nilChan, true},
		{"nil_pointer", cfgPtr, true},
		{"non_nil_slice", []string{"x"}, false},
		{"non_nil_map", map[string]string{"a": "b"}, false},
		{"non_nil_pointer", &Config{}, false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := isNil(tc.value); got != tc.expect {
				t.Fatalf("isNil(%v) = %v, want %v", tc.value, got, tc.expect)
			}
		})
	}
}
