package main

import (
	"strings"
	"testing"
)

func TestParseTestsCapturesFailureOutput(t *testing.T) {
	out := `=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
=== RUN   TestDivide
    divide_test.go:12: Divide(1, 0) = 0, want error
    divide_test.go:13: second failure
--- FAIL: TestDivide (0.00s)
=== RUN   TestSkip
    skip_test.go:4: not ready
--- SKIP: TestSkip (0.00s)
FAIL
exit status 1
FAIL	exercise	0.003s
`

	got := parseTests(out)
	if len(got) != 3 {
		t.Fatalf("got %d tests, want 3: %#v", len(got), got)
	}

	if got[0].Name != "TestAdd" || !got[0].Pass || got[0].Output != "" {
		t.Errorf("TestAdd = %#v, want pass with no output", got[0])
	}
	if got[1].Name != "TestDivide" || got[1].Pass {
		t.Errorf("TestDivide = %#v, want fail", got[1])
	}
	if !strings.Contains(got[1].Output, "Divide(1, 0) = 0, want error") {
		t.Errorf("TestDivide output = %q, want the failure message", got[1].Output)
	}
	if !strings.Contains(got[1].Output, "second failure") {
		t.Errorf("TestDivide output = %q, want all failure lines", got[1].Output)
	}
	if strings.Contains(got[1].Output, "--- FAIL") {
		t.Errorf("TestDivide output = %q, should not contain test markers", got[1].Output)
	}
	if got[2].Name != "TestSkip" || got[2].Pass {
		t.Errorf("TestSkip = %#v, want a non-passing result", got[2])
	}
}

func TestParseTestsHandlesSubtests(t *testing.T) {
	out := `=== RUN   TestGreeting
=== RUN   TestGreeting/empty
    greeting_test.go:9: Greeting("") = "Hello, !", want "Hello!"
--- FAIL: TestGreeting (0.00s)
    --- FAIL: TestGreeting/empty (0.00s)
=== RUN   TestGreeting/name
--- PASS: TestGreeting/name (0.00s)
FAIL
`
	got := parseTests(out)
	if len(got) != 3 {
		t.Fatalf("got %d tests, want 3: %#v", len(got), got)
	}
	if got[0].Name != "TestGreeting" || got[0].Pass {
		t.Errorf("parent test = %#v, want fail", got[0])
	}
	if got[1].Name != "TestGreeting/empty" || got[1].Pass {
		t.Errorf("failing subtest = %#v, want fail", got[1])
	}
	if !strings.Contains(got[1].Output, `Greeting("")`) {
		t.Errorf("subtest output = %q, want its failure message", got[1].Output)
	}
	if got[2].Name != "TestGreeting/name" || !got[2].Pass {
		t.Errorf("passing subtest = %#v, want pass", got[2])
	}
}

func TestParseTestsReturnsNothingForBuildFailure(t *testing.T) {
	out := "# exercise\n./main.go:5:2: undefined: nope\nFAIL\texercise [build failed]\n"
	if got := parseTests(out); len(got) != 0 {
		t.Fatalf("got %#v, want no tests", got)
	}
}
