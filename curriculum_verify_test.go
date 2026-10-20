package main

import (
	"os"
	"strings"
	"testing"

	"gotut/curriculum"
)

// TestReferenceSolutions runs every exercise's reference solution against its
// hidden tests. It is opt-in because a full pass compiles 83 throwaway modules
// and downloads the third-party dependencies used by module 2.
//
//	GOTUT_VERIFY=1 go test -run TestReferenceSolutions -v .
//	GOTUT_VERIFY=1 GOTUT_VERIFY_MATCH=m1t3 go test -run TestReferenceSolutions -v .
func TestReferenceSolutions(t *testing.T) {
	if os.Getenv("GOTUT_VERIFY") == "" {
		t.Skip("set GOTUT_VERIFY=1 to run every reference solution")
	}

	match := os.Getenv("GOTUT_VERIFY_MATCH")
	matches := strings.Split(match, ",")
	for _, mod := range curriculum.All() {
		for _, topic := range mod.Topics {
			for _, ex := range topic.Exercises {
				if match != "" && !matchesExercise(ex.ID, topic.ID, matches) {
					continue
				}
				ex := ex
				t.Run(ex.ID, func(t *testing.T) {
					res := executeExercise(&ex, []curriculum.File{{Name: "main.go", Content: ex.Solution}}, "test", "", 0)
					if res.OK {
						return
					}

					var b strings.Builder
					b.WriteString("reference solution did not pass\n")
					if res.Note != "" {
						b.WriteString("note: " + res.Note + "\n")
					}
					if len(res.Tests) > 0 {
						for _, test := range res.Tests {
							status := "PASS"
							if !test.Pass {
								status = "FAIL"
							}
							b.WriteString(status + " " + test.Name + "\n")
							if test.Output != "" {
								b.WriteString(test.Output + "\n")
							}
						}
					}
					if res.Stdout != "" {
						b.WriteString("stdout:\n" + res.Stdout + "\n")
					}
					if res.Stderr != "" {
						b.WriteString("stderr:\n" + res.Stderr + "\n")
					}
					t.Fatal(b.String())
				})
			}
		}
	}
}

func TestStarterSolutionsFail(t *testing.T) {
	if os.Getenv("GOTUT_VERIFY") == "" {
		t.Skip("set GOTUT_VERIFY=1 to run every starter solution")
	}

	match := os.Getenv("GOTUT_VERIFY_MATCH")
	matches := strings.Split(match, ",")
	for _, mod := range curriculum.All() {
		for _, topic := range mod.Topics {
			for _, ex := range topic.Exercises {
				if match != "" && !matchesExercise(ex.ID, topic.ID, matches) {
					continue
				}
				ex := ex
				t.Run(ex.ID, func(t *testing.T) {
					res := executeExercise(&ex, []curriculum.File{{Name: "main.go", Content: ex.Starter}}, "test", "", 0)
					if res.OK {
						t.Fatalf("starter code passed the hidden tests; the exercise has no work to do")
					}
				})
			}
		}
	}
}

func matchesExercise(exID, topicID string, matches []string) bool {
	for _, m := range matches {
		m = strings.TrimSpace(m)
		if m != "" && (strings.Contains(exID, m) || strings.Contains(topicID, m)) {
			return true
		}
	}
	return false
}
