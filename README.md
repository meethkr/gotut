# GoTut

A local, self-hosted, interactive Go tutorial: for every concept you get **theory, multiple-choice questions, and coding exercises that are actually compiled and tested on your machine**.

No dependencies. No build step for the UI. One `go run .` and you are learning.

```bash
cd gotut
go run .
# open http://127.0.0.1:8080
```

or, equivalently:

```bash
./run.sh          # foreground
./run.sh --bg     # background, logging to /tmp/gotut.log
```

`run.sh` also tells you what to do if port 8080 is already in use.

The UI files are embedded in the binary, but if a `static/` directory exists next to the process they are served from disk instead — so when you edit `app.js`, `styles.css` or `index.html` you only need to reload the page, not restart. A restart is only needed for changes to the Go code or to the curriculum. Static responses are sent with `Cache-Control: no-store`, so a plain reload always shows the current UI.

The first time an exercise uses a third-party package (`chi`, `sqlx`, SQLite) the Go toolchain downloads it into your module cache, which takes a few seconds. Everything else runs offline.

---

## What is inside

| Module                      | Topics   | What you learn                                                                                                                                                                                                                                                              |
| --------------------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **1 · Go Fundamentals**     | 18       | packages, variables & zero values, types & conversions, formatted I/O, `if`/`switch`, loops, functions, closures & `defer`, pointers, structs, arrays & slices, slices under the hood, strings/bytes/runes, maps, methods & embedding, interfaces, errors & panic, generics |
| **2 · Backend Engineering** | 12       | goroutines & `WaitGroup`, channels, `select`/timeouts/mutexes, `context`, JSON & struct tags, `net/http` handlers & routing, `chi` & REST patterns, middleware, HTTP clients, `database/sql`, `sqlx` & transactions, config + `slog` + testing                              |
| **3 · Capstone Project**    | 5 stages | a URL shortener + analytics API: concurrent store → REST API → middleware/auth → SQL persistence → async worker pool                                                                                                                                                        |

That is **35 topics and 83 graded exercises**. Every topic is a sequence of steps: **theory → quiz → exercise(s)**. Progress and your code are stored in the browser's `localStorage`.

---

## How an exercise is graded

- **Run** executes your `main.go` with `go run .` and shows stdout / stderr.
- **Check** runs automated tests (`go test -v`) against your solution.

Both happen in a throwaway temporary module on this machine with a timeout, so a runaway `for {}` cannot hang the app (it is killed after ~25 s).

The reference solution is one click away behind a confirmation, and each exercise has progressive hints.

---

## Layout

```text
.
├── main.go                 # HTTP server: static assets, /api/curriculum, /api/execute
├── runner.go               # sandboxed `go run` / `go test` execution
├── curriculum/
│   ├── types.go            # content model + the "§" → backtick renderer
│   ├── module1a.go         # Module 1, topics 1–5   (hello/packages, variables, types, fmt, if/switch)
│   ├── module1b.go         # Module 1, topics 6–10  (loops, functions, closures/defer, pointers, structs)
│   ├── module1c.go         # Module 1, topics 11–14 (slices, slice internals, strings, maps)
│   ├── module1d.go         # Module 1, topics 15–18 (methods/embedding, interfaces, errors, generics)
│   ├── module2a.go         # Module 2, topics 1–4   (concurrency: goroutines, channels, select, context)
│   ├── module2b.go         # Module 2, topics 5–9   (JSON, net/http, chi, middleware, clients)
│   ├── module2c.go         # Module 2, topics 10–12 (database/sql, sqlx, config/slog/testing)
│   └── module3.go          # the 5-stage capstone
└── static/
    ├── index.html
    ├── styles.css
    └── app.js              # navigation, markdown, quizzes, editor, runner UI
```

---

## Requirements

Go 1.22 or newer (1.22 is the minimum because the exercises use the method-aware `http.ServeMux` patterns and `for range <int>`), on macOS or Linux. The execution sandbox puts each run in its own process group so a runaway program can be killed together with the toolchain child it spawns, which relies on a Unix `syscall`.
