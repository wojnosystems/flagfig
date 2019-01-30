package flagfig

import (
	"strings"
	"testing"
)

func TestArgsAfterArgWithEqualTo(t *testing.T) {
	cases := map[string]struct {
		input    []string
		token    string
		expected []string
	}{
		"split on -v": {
			input:    []string{"/usr/local/go/bin/go", "get", "-t", "-v", "github.com/go-sql-driver/mysql/...", "#gosetup"},
			token:    "-v",
			expected: []string{"github.com/go-sql-driver/mysql/...", "#gosetup"},
		},
		"debugging split": {
			input:    []string{"/usr/local/go/bin/go", "tool", "test2json", "-t", "/home/wojno/.local/share/JetBrains/Toolbox/apps/Goland/ch-0/183.5153.54/plugins/go/lib/dlv/linux/dlv", "--listen=localhost:43235", "--headless=true", "--api-version=2", "exec", "/tmp/___road_test_go", "--", "-test.v", "-test.run", "^TestRunRoadTestSuite$", "#gosetup"},
			token:    "--",
			expected: []string{"-test.v", "-test.run", "^TestRunRoadTestSuite$", "#gosetup"},
		},
		"token was not found": {
			input:    []string{"/usr/local/go/bin/go", "get", "-t", "-v", "github.com/go-sql-driver/mysql/...", "#gosetup"},
			token:    "BLARGH",
			expected: []string{"/usr/local/go/bin/go", "get", "-t", "-v", "github.com/go-sql-driver/mysql/...", "#gosetup"},
		},
	}

	for caseName, c := range cases {
		actual := ArgsAfterArgWithEqualTo(c.token, c.input...)
		for i, ev := range c.expected {
			if strings.Compare(ev, actual[i]) != 0 {
				t.Error("strings should match for case: ", caseName)
			}
		}
	}
}
