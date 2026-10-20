// Package curriculum holds the learning content for the GoTut tutorial.
//
// All human readable text in this package is written as Go raw string
// literals. Because raw strings cannot contain a backtick, the section sign
// "§" is used as a stand-in for a backtick everywhere in this package (markdown
// inline code, fenced code blocks, and Go struct tags). The HTTP layer swaps
// "§" back to a real backtick right before the JSON is written to the client.
package curriculum

import "strings"

// File is a single source file belonging to an exercise.
type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// Option is one choice in a multiple choice question.
type Option struct {
	Text      string `json:"text"`
	Correct   bool   `json:"correct"`
	Rationale string `json:"rationale"`
}

// Question is a single multiple choice / code prediction question.
type Question struct {
	ID          string   `json:"id"`
	Prompt      string   `json:"prompt"`
	Code        string   `json:"code"`
	Options     []Option `json:"options"`
	Multi       bool     `json:"multi"`
	Explanation string   `json:"explanation"`
}

// Exercise is a hands on coding task. Exercises always run inside a fresh
// temporary Go module. "Tests" are never shown in the editor, they are appended
// server side when the learner presses "Test" or "Submit".
type Exercise struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Prompt   string   `json:"prompt"`
	Starter  string   `json:"starter"`
	Tests    []File   `json:"-"`
	Hints    []string `json:"hints"`
	Solution string   `json:"solution"`
	Stdin    bool     `json:"stdin"`
	SampleIn string   `json:"sampleIn"`
	// Deps lists module paths that must be present in the sandbox go.mod
	// (for example "github.com/go-chi/chi/v5 v5.0.12").
	Deps []string `json:"deps"`
}

// Topic is one lesson: theory, a quiz and one or more exercises.
type Topic struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Summary   string     `json:"summary"`
	Theory    string     `json:"theory"`
	Quiz      []Question `json:"quiz"`
	Exercises []Exercise `json:"exercises"`
}

// Module groups topics.
type Module struct {
	ID      string  `json:"id"`
	Title   string  `json:"title"`
	Summary string  `json:"summary"`
	Topics  []Topic `json:"topics"`
}

// All returns the complete curriculum.
func All() []Module {
	return []Module{module1(), module2(), module3()}
}

// Render converts the "§" backtick placeholder used throughout this package
// into a real backtick. Call it on any content that is about to leave Go:
// the JSON sent to the browser and the source files written to the sandbox.
func Render(s string) string {
	if !strings.Contains(s, "§") {
		return s
	}
	return strings.ReplaceAll(s, "§", "`")
}

func exerciseIndex() map[string]*Exercise {
	idx := map[string]*Exercise{}
	for _, m := range All() {
		for ti := range m.Topics {
			for ei := range m.Topics[ti].Exercises {
				ex := &m.Topics[ti].Exercises[ei]
				idx[ex.ID] = ex
			}
		}
	}
	return idx
}

// FindExercise looks up an exercise by id.
func FindExercise(id string) (*Exercise, bool) {
	ex, ok := exerciseIndex()[id]
	return ex, ok
}
