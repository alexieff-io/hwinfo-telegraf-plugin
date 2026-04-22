package hwinfoShMem

import "testing"

func TestStartsWithLower(t *testing.T) {
	tests := []struct {
		name, str, substr string
		want              bool
	}{
		{"exact_lower", "cpu", "cpu", true},
		{"mixed_case_str", "CPU [#0]", "cpu", true},
		{"mixed_case_substr", "cpu", "CPU", true},
		{"no_match", "gpu", "cpu", false},
		{"prefix_only", "cpuabc", "cpu", true},
		{"substring_not_prefix", "abc cpu", "cpu", false},
		{"empty_substr", "cpu", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := StartsWithLower(tc.str, tc.substr); got != tc.want {
				t.Errorf("StartsWithLower(%q, %q) = %v, want %v", tc.str, tc.substr, got, tc.want)
			}
		})
	}
}
