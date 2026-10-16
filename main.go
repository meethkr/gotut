// GoTut is a small local, self-hosted Go tutorial.
// It serves a single page app and executes the learner's code on this machine.
package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"time"

	"gotut/curriculum"
)

//go:embed all:static
var staticFS embed.FS

func main() {
	// Go's mime table does not know woff2 on every platform; without this the
	// font would be served as application/octet-stream.
	_ = mime.AddExtensionType(".woff2", "font/woff2")
	_ = mime.AddExtensionType(".woff", "font/woff")

	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("static assets: %v", err)
	}

	// When you are working on the UI it is convenient to edit the files under
	// static/ and just reload the page. If that directory exists next to the
	// process we serve it from disk; otherwise we fall back to the copy that was
	// embedded into the binary at build time.
	var assets http.FileSystem = http.FS(static)
	if info, err := os.Stat("static"); err == nil && info.IsDir() {
		assets = http.Dir("static")
		log.Printf("serving UI from ./static (live edits)")
	}

	mux := http.NewServeMux()
	mux.Handle("/", noStore(http.FileServer(assets)))
	mux.HandleFunc("/api/curriculum", handleCurriculum)
	mux.HandleFunc("/api/execute", handleExecute)

	addr := "127.0.0.1:8080"
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("GoTut is running — open http://%s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// handleCurriculum streams the whole curriculum as JSON. The "§" backtick
// placeholder is swapped for a real backtick here (see curriculum/types.go).
func handleCurriculum(w http.ResponseWriter, r *http.Request) {
	raw, err := json.Marshal(curriculum.All())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	raw = []byte(curriculum.Render(string(raw)))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(raw)
}

// noStore stops the browser caching app.js / styles.css. Without it a plain
// reload can keep showing yesterday's UI, which looks exactly like "my changes
// did not apply".
func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, must-revalidate")
		next.ServeHTTP(w, r)
	})
}

type executeRequest struct {
	ExerciseID string            `json:"exerciseId"`
	Files      []curriculum.File `json:"files"`
	Mode       string            `json:"mode"`
	Stdin      string            `json:"stdin"`
}

func handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	var req executeRequest
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	ex, ok := curriculum.FindExercise(req.ExerciseID)
	if !ok {
		http.Error(w, "unknown exercise: "+req.ExerciseID, http.StatusNotFound)
		return
	}

	if req.Mode == "" {
		req.Mode = "run"
	}
	if req.Mode != "run" && req.Mode != "test" {
		http.Error(w, "mode must be run or test", http.StatusBadRequest)
		return
	}

	files := req.Files
	if len(files) == 0 && req.Mode == "run" {
		files = []curriculum.File{{Name: "main.go", Content: ex.Starter}}
	}
	result := executeExercise(ex, files, req.Mode, req.Stdin, 0)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(result)
}
