package curriculum

// module3 builds Module 3: a multi-stage capstone project.
func module3() Module {
	return Module{
		ID:      "m3",
		Title:   "Module 3 · Capstone Project: URL Shortener & Analytics API",
		Summary: "Five stages that grow from an in-memory map to a tested, concurrent HTTP service with SQL persistence and an analytics worker pool. Each stage is graded by a hidden test suite.",
		Topics: []Topic{
			m3stage1(), m3stage2(), m3stage3(), m3stage4(), m3stage5(),
		},
	}
}

func m3stage1() Topic {
	return Topic{
		ID:      "m3s1",
		Title:   "Stage 1 · A concurrent in-memory store",
		Summary: "Thread-safe storage, short-code generation and the repository contract.",
		Theory: `
Every backend starts as a data structure plus a rule about who may touch it.
Here the data is a map from short code to URL, and the rule is: **many goroutines
may read, only one may write at a time**.

## Why not just a §map§?

A bare map is not safe for concurrent use. Two goroutines writing (or one writing
while another reads) is a data race — the runtime detects it and crashes with
§fatal error: concurrent map writes§. §go test -race§ will catch it in development
instead of in production.

The fix is a §sync.RWMutex§:

§§§go
type Shortener struct {
    mu     sync.RWMutex
    urls   map[string]string
    clicks map[string]int
    next   uint64
}

func (s *Shortener) Resolve(code string) (string, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    url, ok := s.urls[code]
    if !ok {
        return "", ErrNotFound
    }
    return url, nil
}

func (s *Shortener) Create(original string) (string, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    code := s.newCode()
    s.urls[code] = original
    return code, nil
}
§§§

§sync.RWMutex§ allows any number of concurrent readers, but writers are exclusive.
Use §Lock§/§Unlock§ for anything that mutates, §RLock§/§RUnlock§ for pure reads.

Note the §defer§: it makes the critical section impossible to leak, even if a
future edit adds an early return.

## Code generation

Requirements for a short code:

* **unique** — collision would silently overwrite a URL,
* **short** — that is the purpose,
* **URL-safe** — no §/§, §?§, §#§ or whitespace.

Base62 (§0-9A-Za-z§) is the classic alphabet. A monotonic counter encoded in
base62 is unique by construction and requires no collision loop:

§§§go
const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func base62(n uint64) string {
    if n == 0 {
        return string(alphabet[0])
    }
    var out []byte
    for n > 0 {
        out = append([]byte{alphabet[n%62]}, out...)
        n /= 62
    }
    return string(out)
}
§§§

Random codes are also fine (§crypto/rand§ or §math/rand/v2§) as long as you check
for collisions under the lock:

§§§go
for {
    code := randomCode(7)
    if _, exists := s.urls[code]; !exists {
        s.urls[code] = original
        return code, nil
    }
}
§§§

Because the map is only touched under the write lock, the check-then-insert is
atomic with respect to other goroutines. (Outside a lock it would be a classic
TOCTOU bug.)

## The contract

Keep the storage behind a small interface so later stages can swap the
implementation without touching the HTTP layer:

§§§go
type Store interface {
    Create(ctx context.Context, url string) (string, error)
    Resolve(ctx context.Context, code string) (string, error)
    Record(ctx context.Context, code string) error
    Clicks(ctx context.Context, code string) (int, error)
}
§§§

Adding §ctx§ from day one costs nothing and means the SQL stage is a drop-in
replacement. This is the single most important design decision in the project.
`,
		Quiz: []Question{
			{
				ID:     "m3s1q1",
				Prompt: "Why is §sync.RWMutex§ a good fit for this store?",
				Options: []Option{
					{Text: "Redirects are reads and vastly outnumber creates, so concurrent readers should not block each other.", Correct: true, Rationale: "§RLock§ allows unlimited simultaneous readers; only writes are exclusive."},
					{Text: "Because §map§ requires an RWMutex to compile.", Correct: false, Rationale: "Maps have no such requirement."},
					{Text: "Because it is faster than a plain Mutex for all workloads.", Correct: false, Rationale: "Under heavy writes, §RWMutex§ overhead can make it slower."},
					{Text: "Because it prevents goroutine leaks.", Correct: false, Rationale: "It only guards memory access."},
				},
				Explanation: "Read-heavy workloads benefit from §RWMutex§; write-heavy ones are usually better with a plain §sync.Mutex§ or sharding.",
			},
			{
				ID:     "m3s1q2",
				Prompt: "Where must the §defer s.mu.RUnlock()§ appear?",
				Options: []Option{
					{Text: "Immediately after §RLock()§, before any logic that can return early.", Correct: true, Rationale: "That guarantees the lock is released on every path."},
					{Text: "Just before the §return§ statement.", Correct: false, Rationale: "Early returns would leak the read lock."},
					{Text: "In the calling function.", Correct: false, Rationale: "The lock must be released by the same goroutine that acquired it, as soon as possible."},
					{Text: "It does not matter — Go unlocks automatically.", Correct: false, Rationale: "Go has no automatic unlocking."},
				},
				Explanation: "§defer§ right after acquiring is the standard pattern, and it is why §mu§ is always the first field of the struct in Go style.",
			},
			{
				ID:     "m3s1q3",
				Prompt: "A counter-based base62 code generator never collides. Why then would anyone choose random codes?",
				Options: []Option{
					{Text: "Because sequential codes are guessable — an attacker can enumerate every short link.", Correct: true, Rationale: "Enumeration is a real privacy problem for a public shortener; random codes add entropy."},
					{Text: "Because counters cannot be encoded in base62.", Correct: false, Rationale: "They can."},
					{Text: "Because random codes are shorter.", Correct: false, Rationale: "They are usually longer for the same collision resistance."},
					{Text: "Because counters are not thread-safe.", Correct: false, Rationale: "A counter under a lock is perfectly safe."},
				},
				Explanation: "Trade-off: sequential codes are cheap and collision-free but enumerable; random codes are unguessable but need a collision check.",
			},
		},
		Exercises: []Exercise{
			{
				ID:     "m3s1e1",
				Title:  "Stage 1 — in-memory shortener",
				Prompt: "**Project stage 1 of 5.**\n\nImplement a thread-safe in-memory URL shortener.\n\n§§§go\ntype Shortener struct{ /* your fields */ }\n\nvar (\n    ErrNotFound = errors.New(\"not found\")\n    ErrEmptyURL = errors.New(\"url must not be empty\")\n)\n\nfunc NewShortener() *Shortener\nfunc (s *Shortener) Create(original string) (string, error)\nfunc (s *Shortener) Resolve(code string) (string, error)\nfunc (s *Shortener) Record(code string) error\nfunc (s *Shortener) Clicks(code string) (int, error)\nfunc (s *Shortener) Len() int\n§§§\n\nRules:\n\n* §Create§ rejects an empty (or whitespace-only) URL with §ErrEmptyURL§ and returns\n  a short code that is unique, ASCII, and safe in a URL path.\n* §Resolve§ and §Clicks§ return §ErrNotFound§ for unknown codes.\n* §Record§ increments the click counter and returns §ErrNotFound§ for unknown codes.\n* The store must be safe for concurrent use — the hidden test hammers it from\n  50 goroutines and runs under §-race§.\n\n**Design tip:** keep everything behind one §sync.RWMutex§, and prefer a monotonic\ncounter encoded in base62 so you never need a collision loop.",
				Starter: `package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	ErrNotFound = errors.New("not found")
	ErrEmptyURL = errors.New("url must not be empty")
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Shortener struct {
	mu     sync.RWMutex
	urls   map[string]string
	clicks map[string]int
	next   uint64
}

func NewShortener() *Shortener {
	// TODO
	return nil
}

func base62(n uint64) string {
	// TODO: encode n using alphabet
	return ""
}

func (s *Shortener) Create(original string) (string, error) {
	// TODO
	return "", nil
}

func (s *Shortener) Resolve(code string) (string, error) {
	// TODO
	return "", nil
}

func (s *Shortener) Record(code string) error {
	// TODO
	return nil
}

func (s *Shortener) Clicks(code string) (int, error) {
	// TODO
	return 0, nil
}

func (s *Shortener) Len() int {
	// TODO
	return 0
}

func main() {
	s := NewShortener()
	code, err := s.Create("https://go.dev")
	fmt.Println(code, err)
	url, err := s.Resolve(code)
	fmt.Println(url, err)
	_ = strings.TrimSpace
}
`,
				Tests: []File{{
					Name: "store_test.go",
					Content: `package main

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestCreateAndResolve(t *testing.T) {
	s := NewShortener()
	code, err := s.Create("https://go.dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if code == "" {
		t.Fatal("Create returned an empty code")
	}
	if strings.ContainsAny(code, "/?# ") {
		t.Fatalf("code %q is not URL-safe", code)
	}
	url, err := s.Resolve(code)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if url != "https://go.dev" {
		t.Fatalf("Resolve = %q", url)
	}
}

func TestCreateValidation(t *testing.T) {
	s := NewShortener()
	for _, bad := range []string{"", "   ", "\t\n"} {
		if _, err := s.Create(bad); !errors.Is(err, ErrEmptyURL) {
			t.Fatalf("Create(%q) = %v, want ErrEmptyURL", bad, err)
		}
	}
}

func TestCodesAreUnique(t *testing.T) {
	s := NewShortener()
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		code, err := s.Create("https://example.com/" + strings.Repeat("x", i%7))
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if seen[code] {
			t.Fatalf("duplicate code %q", code)
		}
		seen[code] = true
	}
	if s.Len() != 500 {
		t.Fatalf("Len = %d, want 500", s.Len())
	}
}

func TestClicks(t *testing.T) {
	s := NewShortener()
	code, _ := s.Create("https://go.dev")

	if _, err := s.Clicks(code); err != nil {
		t.Fatalf("Clicks on a fresh code: %v", err)
	}
	if n, _ := s.Clicks(code); n != 0 {
		t.Fatalf("fresh clicks = %d, want 0", n)
	}
	for i := 0; i < 3; i++ {
		if err := s.Record(code); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
	if n, _ := s.Clicks(code); n != 3 {
		t.Fatalf("clicks = %d, want 3", n)
	}
}

func TestNotFound(t *testing.T) {
	s := NewShortener()
	if _, err := s.Resolve("nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Resolve = %v, want ErrNotFound", err)
	}
	if err := s.Record("nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Record = %v, want ErrNotFound", err)
	}
	if _, err := s.Clicks("nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Clicks = %v, want ErrNotFound", err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := NewShortener()
	const goroutines = 50
	const perGoroutine = 40

	var wg sync.WaitGroup
	codes := make([][]string, goroutines)
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				code, err := s.Create("https://example.com")
				if err != nil {
					t.Errorf("Create: %v", err)
					return
				}
				codes[g] = append(codes[g], code)
			}
		}(g)
	}
	wg.Wait()

	if s.Len() != goroutines*perGoroutine {
		t.Fatalf("Len = %d, want %d", s.Len(), goroutines*perGoroutine)
	}

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for _, code := range codes[g] {
				if _, err := s.Resolve(code); err != nil {
					t.Errorf("Resolve(%q): %v", code, err)
				}
				if err := s.Record(code); err != nil {
					t.Errorf("Record(%q): %v", code, err)
				}
				if _, err := s.Clicks(code); err != nil {
					t.Errorf("Clicks(%q): %v", code, err)
				}
			}
		}(g)
	}
	wg.Wait()
}

func TestCreateSameURLTwiceGivesDistinctCodes(t *testing.T) {
	s := NewShortener()
	const u = "https://example.com/duplicate"

	var codes []string
	for i := 0; i < 3; i++ {
		code, err := s.Create(u)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		codes = append(codes, code)
	}

	seen := map[string]bool{}
	for _, code := range codes {
		if seen[code] {
			t.Fatalf("duplicate code %q for the same URL", code)
		}
		seen[code] = true
		got, err := s.Resolve(code)
		if err != nil || got != u {
			t.Fatalf("Resolve(%q) = %q, %v; want %q", code, got, err, u)
		}
	}
	if s.Len() != len(codes) {
		t.Fatalf("Len = %d, want %d", s.Len(), len(codes))
	}
}

func TestCreateResolveRoundTrip(t *testing.T) {
	s := NewShortener()
	urls := []string{
		"https://go.dev/doc/?q=1#top",
		"https://example.com/a%20b?x=1&y=2",
		"https://例え.jp/パス",
	}

	seen := map[string]bool{}
	for _, u := range urls {
		code, err := s.Create(u)
		if err != nil {
			t.Fatalf("Create(%q): %v", u, err)
		}
		if seen[code] {
			t.Fatalf("duplicate code %q", code)
		}
		seen[code] = true
		got, err := s.Resolve(code)
		if err != nil || got != u {
			t.Fatalf("Resolve(%q) = %q, %v; want %q", code, got, err, u)
		}
	}
}

func TestCodesAreURLSafe(t *testing.T) {
	s := NewShortener()
	for i := 0; i < 200; i++ {
		code, err := s.Create("https://example.com")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if code == "" {
			t.Fatal("Create returned an empty code")
		}
		for _, r := range code {
			if r <= ' ' || r > '~' || strings.ContainsRune("/?#%", r) {
				t.Fatalf("code %q contains unsafe character %q", code, r)
			}
		}
	}
}

func TestRejectedCreateIsNotStored(t *testing.T) {
	s := NewShortener()
	for _, bad := range []string{"", " ", "\t", "\n", "\u00a0"} {
		if _, err := s.Create(bad); !errors.Is(err, ErrEmptyURL) {
			t.Fatalf("Create(%q) = %v, want ErrEmptyURL", bad, err)
		}
	}
	if s.Len() != 0 {
		t.Fatalf("Len = %d after rejected creates, want 0", s.Len())
	}
	if _, err := s.Create("https://go.dev"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if s.Len() != 1 {
		t.Fatalf("Len = %d, want 1", s.Len())
	}
}

func TestClicksArePerCode(t *testing.T) {
	s := NewShortener()
	a, err := s.Create("https://a.dev")
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.Create("https://b.dev")
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		if err := s.Record(a); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
	if n, err := s.Clicks(a); err != nil || n != 5 {
		t.Fatalf("Clicks(a) = %d, %v; want 5", n, err)
	}
	if n, err := s.Clicks(b); err != nil || n != 0 {
		t.Fatalf("Clicks(b) = %d, %v; want 0", n, err)
	}
	if err := s.Record("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Record(missing) = %v, want ErrNotFound", err)
	}
	if n, _ := s.Clicks(a); n != 5 {
		t.Fatalf("Clicks(a) changed to %d after a failed Record", n)
	}
}

func TestConcurrentRecordIsExact(t *testing.T) {
	s := NewShortener()
	code, err := s.Create("https://go.dev")
	if err != nil {
		t.Fatal(err)
	}

	const goroutines = 20
	const each = 25

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < each; i++ {
				if err := s.Record(code); err != nil {
					t.Errorf("Record: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()

	n, err := s.Clicks(code)
	if err != nil {
		t.Fatal(err)
	}
	if n != goroutines*each {
		t.Fatalf("clicks = %d, want %d (lost updates?)", n, goroutines*each)
	}
}

func TestEmptyCodeIsNotFound(t *testing.T) {
	s := NewShortener()
	if _, err := s.Resolve(""); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Resolve(\"\") = %v, want ErrNotFound", err)
	}
	if err := s.Record(""); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Record(\"\") = %v, want ErrNotFound", err)
	}
	if _, err := s.Clicks(""); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Clicks(\"\") = %v, want ErrNotFound", err)
	}
}

func TestConcurrentCreateResolve(t *testing.T) {
	s := NewShortener()
	const goroutines = 10
	const each = 20

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < each; i++ {
				u := "https://example.com/" + strings.Repeat("p", (g+i)%5)
				code, err := s.Create(u)
				if err != nil {
					t.Errorf("Create: %v", err)
					return
				}
				got, err := s.Resolve(code)
				if err != nil || got != u {
					t.Errorf("Resolve(%q) = %q, %v; want %q", code, got, err, u)
					return
				}
			}
		}(g)
	}
	wg.Wait()

	if s.Len() != goroutines*each {
		t.Fatalf("Len = %d, want %d", s.Len(), goroutines*each)
	}
}
`,
				}},
				Hints: []string{
					"§NewShortener§ returns §&Shortener{urls: map[string]string{}, clicks: map[string]int{}}§.",
					"§base62§: build digits least-significant-first, then reverse; handle §n == 0§ as §\"0\"§.",
					"§Create§ under §Lock§: trim the input, reject empty, increment §s.next§, encode, store, return.",
					"Use §strings.TrimSpace(original) == \"\"§ for validation — but store the *original* (untrimmed) URL.",
					"§Len§ needs at least §RLock§.",
				},
				Solution: `package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	ErrNotFound = errors.New("not found")
	ErrEmptyURL = errors.New("url must not be empty")
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Shortener struct {
	mu     sync.RWMutex
	urls   map[string]string
	clicks map[string]int
	next   uint64
}

func NewShortener() *Shortener {
	return &Shortener{
		urls:   make(map[string]string),
		clicks: make(map[string]int),
	}
}

func base62(n uint64) string {
	if n == 0 {
		return string(alphabet[0])
	}
	var out []byte
	for n > 0 {
		out = append(out, alphabet[n%62])
		n /= 62
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}

func (s *Shortener) Create(original string) (string, error) {
	if strings.TrimSpace(original) == "" {
		return "", ErrEmptyURL
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	code := base62(s.next)
	s.urls[code] = original
	return code, nil
}

func (s *Shortener) Resolve(code string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.urls[code]
	if !ok {
		return "", ErrNotFound
	}
	return url, nil
}

func (s *Shortener) Record(code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.urls[code]; !ok {
		return ErrNotFound
	}
	s.clicks[code]++
	return nil
}

func (s *Shortener) Clicks(code string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.urls[code]; !ok {
		return 0, ErrNotFound
	}
	return s.clicks[code], nil
}

func (s *Shortener) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

func main() {
	s := NewShortener()
	code, err := s.Create("https://go.dev")
	fmt.Println(code, err)
	url, err := s.Resolve(code)
	fmt.Println(url, err)
}
`,
			},
		},
	}
}

func m3stage2() Topic {
	return Topic{
		ID:      "m3s2",
		Title:   "Stage 2 · The HTTP API",
		Summary: "Expose the store over REST: shorten, redirect, stats.",
		Theory: `
Now the store becomes a service. The mapping from operations to HTTP is the
contract:

| Operation | Method + path | Success | Failure |
| --- | --- | --- | --- |
| shorten | §POST /shorten§ | §201§ + JSON | §400§ bad JSON/URL |
| redirect | §GET /{code}§ | §302§ + §Location§ | §404§ |
| stats | §GET /stats/{code}§ | §200§ + JSON | §404§ |

## Separating transport from domain

The handler should do exactly three things: parse input, call the store, write
output. No business rules, no SQL, no formatting decisions beyond JSON.

§§§go
func (a *API) shorten(w http.ResponseWriter, r *http.Request) {
    var req struct {
        URL string §json:"url"§
    }
    if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid JSON body")
        return
    }

    code, err := a.store.Create(r.Context(), req.URL)
    switch {
    case errors.Is(err, ErrEmptyURL):
        writeErr(w, http.StatusBadRequest, "url is required")
        return
    case err != nil:
        writeErr(w, http.StatusInternalServerError, "internal error")
        return
    }

    writeJSON(w, http.StatusCreated, map[string]string{
        "code":      code,
        "short_url": a.baseURL + "/" + code,
    })
}
§§§

Domain errors map to status codes in one §switch§. That switch is the entire
translation layer between your domain and HTTP.

## Redirects

§§§go
func (a *API) redirect(w http.ResponseWriter, r *http.Request) {
    code := chi.URLParam(r, "code")
    url, err := a.store.Resolve(r.Context(), code)
    if errors.Is(err, ErrNotFound) {
        http.NotFound(w, r)
        return
    }
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "internal error")
        return
    }
    if err := a.store.Record(r.Context(), code); err != nil {
        // Analytics must not break the redirect: log and continue.
        slog.Error("record click", "code", code, "err", err)
    }
    http.Redirect(w, r, url, http.StatusFound)   // 302
}
§§§

Two decisions matter here:

* §http.Redirect§ sets §Location§ and the status for you.
* 302 (Found) versus 301 (Moved Permanently): 301 is cached by browsers forever,
  so click counting stops working. Use 302 — or 307/308 if you care about method
  preservation.
* A failure to record a click must never fail the redirect. Side concerns degrade;
  the main path survives.

## Why §{code}§ at the root is a design constraint

§GET /{code}§ swallows everything at the root: §/favicon.ico§, §/robots.txt§, and
your own §/stats§ route all collide with it.

Standard resolutions:

1. Prefix the API: §/api/shorten§, §/s/{code}§, §/api/stats/{code}§.
2. Reserve a reserved-word list and 404 those explicitly.
3. Use a host split: §short.example.com/{code}§ vs §api.example.com/...§.

Routing precedence in chi and in Go 1.22's §ServeMux§ favours the more specific
pattern, so §GET /stats/{code}§ wins over §GET /{code}§ automatically. Explicit
patterns beat wildcards.

## Responses that clients can rely on

* Always set §Content-Type§ — §application/json; charset=utf-8§ for JSON.
* 201 for creation, and include the created resource (or a §Location§ header).
* 204 for a successful delete with no body.
* A stable error envelope: §{"error": "..."}§.
* Never return a Go error string straight to the client: it leaks internals.
`,
		Quiz: []Question{
			{
				ID:     "m3s2q1",
				Prompt: "Why is 302 usually better than 301 for a public shortener?",
				Options: []Option{
					{Text: "Browsers cache 301 aggressively, so subsequent clicks never reach the server and analytics silently break.", Correct: true, Rationale: "A permanent redirect is remembered; you lose the request."},
					{Text: "302 is faster.", Correct: false, Rationale: "Both are a single response."},
					{Text: "301 is deprecated.", Correct: false, Rationale: "It is not."},
					{Text: "302 preserves the request method.", Correct: false, Rationale: "That is 307/308."},
				},
				Explanation: "Redirect analytics and caching are in direct tension; 302 keeps the server in the loop.",
			},
			{
				ID:     "m3s2q2",
				Prompt: "Which routes does §GET /{code}§ collide with? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "§GET /stats/{code}§", Correct: true, Rationale: "A request to §/stats/abc§ has two path segments; §/{code}§ matches only one segment, so a *specific* §/stats/...§ pattern wins. It is a near-collision that requires precedence rules."},
					{Text: "§GET /favicon.ico§", Correct: true, Rationale: "Single segment, so it matches §/{code}§ unless you special-case it."},
					{Text: "§POST /shorten§", Correct: false, Rationale: "The method differs, so a method-aware router separates them."},
					{Text: "§GET /s/{code}§", Correct: false, Rationale: "Two segments — no collision."},
				},
				Explanation: "Root-level wildcards are convenient but greedy; prefix your API surface or reserve names.",
			},
			{
				ID:     "m3s2q3",
				Prompt: "The click-recording call fails while resolving a redirect. What should the handler do?",
				Options: []Option{
					{Text: "Log the failure and still redirect — analytics is a secondary concern.", Correct: true, Rationale: "Users must not see an error because a counter failed."},
					{Text: "Return 500 so the failure is visible.", Correct: false, Rationale: "That punishes the user for an internal analytics problem."},
					{Text: "Retry in a loop until it succeeds.", Correct: false, Rationale: "The request would hang; retry belongs in a worker with backoff."},
					{Text: "Ignore it silently.", Correct: false, Rationale: "Log it, at least."},
				},
				Explanation: "Decide deliberately which failures are fatal to a request and which are not — and log the non-fatal ones.",
			},
		},
		Exercises: []Exercise{
			{
				ID:     "m3s2e1",
				Title:  "Stage 2 — REST API over the store",
				Prompt: "**Project stage 2 of 5.**\n\nThe store from stage 1 is provided below (do not change it). Build the HTTP layer.\n\n§§§go\nfunc NewAPI(store Store, baseURL string) *API\nfunc (a *API) Routes() http.Handler\n§§§\n\nEndpoints:\n\n* §POST /shorten§ with §{\"url\": \"https://...\"}§ → §201§ and\n  §{\"code\":\"1\",\"short_url\":\"<baseURL>/1\"}§.\n  Empty/whitespace URL → §400§ with an error envelope.\n* §GET /{code}§ → §302§ with a §Location§ header, and it records a click.\n  Unknown code → §404§.\n* §GET /stats/{code}§ → §200§ and §{\"code\":\"1\",\"url\":\"...\",\"clicks\":2}§.\n  Unknown code → §404§.\n\nUse **chi**. The tests call the handler directly with §httptest§ and **do not follow\nredirects**, so the raw §302§ and §Location§ are asserted.",
				Starter: `package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

var (
	ErrNotFound = errors.New("not found")
	ErrEmptyURL = errors.New("url must not be empty")
)

// ---------- store (from stage 1, unchanged) ----------

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Store interface {
	Create(ctx context.Context, url string) (string, error)
	Resolve(ctx context.Context, code string) (string, error)
	Record(ctx context.Context, code string) error
	Clicks(ctx context.Context, code string) (int, error)
}

type memoryStore struct {
	mu     sync.RWMutex
	urls   map[string]string
	clicks map[string]int
	next   uint64
}

func NewMemoryStore() *memoryStore {
	return &memoryStore{urls: map[string]string{}, clicks: map[string]int{}}
}

func base62(n uint64) string {
	if n == 0 {
		return string(alphabet[0])
	}
	var out []byte
	for n > 0 {
		out = append(out, alphabet[n%62])
		n /= 62
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}

func (s *memoryStore) Create(ctx context.Context, original string) (string, error) {
	if strings.TrimSpace(original) == "" {
		return "", ErrEmptyURL
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	code := base62(s.next)
	s.urls[code] = original
	return code, nil
}

func (s *memoryStore) Resolve(ctx context.Context, code string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.urls[code]
	if !ok {
		return "", ErrNotFound
	}
	return url, nil
}

func (s *memoryStore) Record(ctx context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.urls[code]; !ok {
		return ErrNotFound
	}
	s.clicks[code]++
	return nil
}

func (s *memoryStore) Clicks(ctx context.Context, code string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.urls[code]; !ok {
		return 0, ErrNotFound
	}
	return s.clicks[code], nil
}

// ---------- your work starts here ----------

type API struct {
	store   Store
	baseURL string
}

func NewAPI(store Store, baseURL string) *API {
	// TODO
	return nil
}

func (a *API) Routes() http.Handler {
	r := chi.NewRouter()
	// TODO: POST /shorten, GET /stats/{code}, GET /{code}
	return r
}

func main() {
	_ = chi.NewRouter
	fmt.Println(NewMemoryStore() != nil)
	_ = json.Marshal
}
`,
				Deps: []string{"github.com/go-chi/chi/v5"},
				Tests: []File{{
					Name: "api_test.go",
					Content: `package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fixedStore struct {
	urls   map[string]string
	clicks map[string]int
	next   int
}

func newFixedStore() *fixedStore {
	return &fixedStore{urls: map[string]string{}, clicks: map[string]int{}}
}

func (s *fixedStore) Create(ctx context.Context, url string) (string, error) {
	if strings.TrimSpace(url) == "" {
		return "", ErrEmptyURL
	}
	s.next++
	code := "s" + string(rune('0'+s.next))
	s.urls[code] = url
	return code, nil
}

func (s *fixedStore) Resolve(ctx context.Context, code string) (string, error) {
	u, ok := s.urls[code]
	if !ok {
		return "", ErrNotFound
	}
	return u, nil
}

func (s *fixedStore) Record(ctx context.Context, code string) error {
	if _, ok := s.urls[code]; !ok {
		return ErrNotFound
	}
	s.clicks[code]++
	return nil
}

func (s *fixedStore) Clicks(ctx context.Context, code string) (int, error) {
	if _, ok := s.urls[code]; !ok {
		return 0, ErrNotFound
	}
	return s.clicks[code], nil
}

func newTestAPI(t *testing.T) (*fixedStore, http.Handler) {
	t.Helper()
	st := newFixedStore()
	return st, NewAPI(st, "http://sho.rt").Routes()
}

func TestShorten(t *testing.T) {
	_, h := newTestAPI(t)

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(§{"url":"https://go.dev"}§))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body %s)", rec.Code, rec.Body.String())
	}
	var resp struct {
		Code     string §json:"code"§
		ShortURL string §json:"short_url"§
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("body: %v (%s)", err, rec.Body.String())
	}
	if resp.Code == "" {
		t.Fatal("empty code in response")
	}
	if resp.ShortURL != "http://sho.rt/"+resp.Code {
		t.Fatalf("short_url = %q", resp.ShortURL)
	}
}

func TestShortenBadInput(t *testing.T) {
	_, h := newTestAPI(t)

	for name, body := range map[string]string{
		"empty url":  §{"url":""}§,
		"blank url":  §{"url":"   "}§,
		"bad json":   §{oops§,
	} {
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", name, rec.Code)
		}
	}
}

func TestRedirect(t *testing.T) {
	st, h := newTestAPI(t)
	st.urls["abc"] = "https://example.com"

	req := httptest.NewRequest(http.MethodGet, "/abc", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "https://example.com" {
		t.Fatalf("Location = %q", loc)
	}
	if st.clicks["abc"] != 1 {
		t.Fatalf("click was not recorded: %v", st.clicks)
	}
}

func TestRedirectNotFound(t *testing.T) {
	_, h := newTestAPI(t)
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestStats(t *testing.T) {
	st, h := newTestAPI(t)
	st.urls["abc"] = "https://example.com"
	st.clicks["abc"] = 7

	req := httptest.NewRequest(http.MethodGet, "/stats/abc", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	var got struct {
		Code   string §json:"code"§
		URL    string §json:"url"§
		Clicks int    §json:"clicks"§
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Code != "abc" || got.URL != "https://example.com" || got.Clicks != 7 {
		t.Fatalf("stats = %+v", got)
	}
}

func TestStatsNotFound(t *testing.T) {
	_, h := newTestAPI(t)
	req := httptest.NewRequest(http.MethodGet, "/stats/missing", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestStatsDoesNotRecordAClick(t *testing.T) {
	st, h := newTestAPI(t)
	st.urls["abc"] = "https://example.com"

	req := httptest.NewRequest(http.MethodGet, "/stats/abc", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if st.clicks["abc"] != 0 {
		t.Fatalf("stats recorded a click: %d", st.clicks["abc"])
	}
	_ = io.Discard
	_ = bytes.NewBuffer
}

func doRequest(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestShortenCreatesDistinctCodes(t *testing.T) {
	st, h := newTestAPI(t)
	seen := map[string]bool{}

	for i := 0; i < 3; i++ {
		rec := doRequest(t, h, http.MethodPost, "/shorten", §{"url":"https://go.dev"}§)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create %d: status = %d, want 201 (body %s)", i, rec.Code, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Fatalf("create %d: Content-Type = %q, want JSON", i, ct)
		}
		var resp struct {
			Code     string §json:"code"§
			ShortURL string §json:"short_url"§
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("create %d: body: %v (%s)", i, err, rec.Body.String())
		}
		if resp.Code == "" {
			t.Fatalf("create %d: empty code", i)
		}
		if seen[resp.Code] {
			t.Fatalf("create %d: duplicate code %q", i, resp.Code)
		}
		seen[resp.Code] = true
		if resp.ShortURL != "http://sho.rt/"+resp.Code {
			t.Fatalf("create %d: short_url = %q, want %q", i, resp.ShortURL, "http://sho.rt/"+resp.Code)
		}
	}
	if len(st.urls) != 3 {
		t.Fatalf("store has %d links, want 3", len(st.urls))
	}
}

func TestShortenErrorEnvelope(t *testing.T) {
	_, h := newTestAPI(t)
	cases := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"empty object", §{}§},
		{"null", "null"},
		{"array", "[]"},
		{"wrong type", §{"url":123}§},
		{"blank url", §{"url":" \t "}§},
	}
	for _, c := range cases {
		rec := doRequest(t, h, http.MethodPost, "/shorten", c.body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400 (body %s)", c.name, rec.Code, rec.Body.String())
			continue
		}
		var env map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
			t.Errorf("%s: error body is not JSON: %v (%s)", c.name, err, rec.Body.String())
			continue
		}
		if msg, _ := env["error"].(string); msg == "" {
			t.Errorf("%s: missing error envelope: %s", c.name, rec.Body.String())
		}
	}
}

func TestRedirectAndStatsFlow(t *testing.T) {
	st, h := newTestAPI(t)

	rec := doRequest(t, h, http.MethodPost, "/shorten", §{"url":"https://go.dev"}§)
	if rec.Code != http.StatusCreated {
		t.Fatalf("shorten: status = %d (body %s)", rec.Code, rec.Body.String())
	}
	var created struct {
		Code string §json:"code"§
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Code == "" {
		t.Fatal("shorten returned an empty code")
	}

	for i := 1; i <= 3; i++ {
		got := doRequest(t, h, http.MethodGet, "/"+created.Code, "")
		if got.Code != http.StatusFound {
			t.Fatalf("redirect %d: status = %d, want 302", i, got.Code)
		}
		if loc := got.Header().Get("Location"); loc != "https://go.dev" {
			t.Fatalf("redirect %d: Location = %q", i, loc)
		}
		if st.clicks[created.Code] != i {
			t.Fatalf("redirect %d: clicks = %d, want %d", i, st.clicks[created.Code], i)
		}
	}

	stats := doRequest(t, h, http.MethodGet, "/stats/"+created.Code, "")
	if stats.Code != http.StatusOK {
		t.Fatalf("stats: status = %d (body %s)", stats.Code, stats.Body.String())
	}
	if ct := stats.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("stats: Content-Type = %q, want JSON", ct)
	}
	var got struct {
		Code   string §json:"code"§
		URL    string §json:"url"§
		Clicks int    §json:"clicks"§
	}
	if err := json.Unmarshal(stats.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Code != created.Code || got.URL != "https://go.dev" || got.Clicks != 3 {
		t.Fatalf("stats = %+v, want code %q, url https://go.dev, clicks 3", got, created.Code)
	}
}

func TestStatsNeverRecordsClicks(t *testing.T) {
	st, h := newTestAPI(t)
	st.urls["abc"] = "https://example.com"

	for i := 0; i < 3; i++ {
		rec := doRequest(t, h, http.MethodGet, "/stats/abc", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("stats %d: status = %d (body %s)", i, rec.Code, rec.Body.String())
		}
	}
	if st.clicks["abc"] != 0 {
		t.Fatalf("stats recorded clicks: %d", st.clicks["abc"])
	}
}

func TestUnsupportedMethods(t *testing.T) {
	_, h := newTestAPI(t)
	cases := []struct {
		name   string
		method string
		path   string
	}{
		{"put shorten", http.MethodPut, "/shorten"},
		{"delete shorten", http.MethodDelete, "/shorten"},
		{"post stats", http.MethodPost, "/stats/abc"},
		{"delete redirect", http.MethodDelete, "/abc"},
		{"put redirect", http.MethodPut, "/abc"},
	}
	for _, c := range cases {
		rec := doRequest(t, h, c.method, c.path, "")
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: status = %d, want 405", c.name, rec.Code)
		}
	}
}

func TestUnknownPathsReturn404(t *testing.T) {
	_, h := newTestAPI(t)
	for _, path := range []string{"/", "/a/b", "/stats", "/stats/abc/extra"} {
		rec := doRequest(t, h, http.MethodGet, path, "")
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s: status = %d, want 404", path, rec.Code)
		}
	}
}
`,
				}},
				Hints: []string{
					"Register §GET /stats/{code}§ *before* §GET /{code}§ — chi gives priority to explicit patterns, but registering both is required.",
					"Read the code with §chi.URLParam(r, \"code\")§.",
					"§http.Redirect(w, r, url, http.StatusFound)§ handles the Location header and status.",
					"Order matters for the click: resolve, then record, then redirect.",
					"Define small helpers §writeJSON§ and §writeErr§ for consistency.",
				},
				Solution: `package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	ErrNotFound = errors.New("not found")
	ErrEmptyURL = errors.New("url must not be empty")
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Store interface {
	Create(ctx context.Context, url string) (string, error)
	Resolve(ctx context.Context, code string) (string, error)
	Record(ctx context.Context, code string) error
	Clicks(ctx context.Context, code string) (int, error)
}

type memoryStore struct {
	mu     sync.RWMutex
	urls   map[string]string
	clicks map[string]int
	next   uint64
}

func NewMemoryStore() *memoryStore {
	return &memoryStore{urls: map[string]string{}, clicks: map[string]int{}}
}

func base62(n uint64) string {
	if n == 0 {
		return string(alphabet[0])
	}
	var out []byte
	for n > 0 {
		out = append(out, alphabet[n%62])
		n /= 62
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}

func (s *memoryStore) Create(ctx context.Context, original string) (string, error) {
	if strings.TrimSpace(original) == "" {
		return "", ErrEmptyURL
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	code := base62(s.next)
	s.urls[code] = original
	return code, nil
}

func (s *memoryStore) Resolve(ctx context.Context, code string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.urls[code]
	if !ok {
		return "", ErrNotFound
	}
	return url, nil
}

func (s *memoryStore) Record(ctx context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.urls[code]; !ok {
		return ErrNotFound
	}
	s.clicks[code]++
	return nil
}

func (s *memoryStore) Clicks(ctx context.Context, code string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.urls[code]; !ok {
		return 0, ErrNotFound
	}
	return s.clicks[code], nil
}

type API struct {
	store   Store
	baseURL string
}

func NewAPI(store Store, baseURL string) *API {
	return &API{store: store, baseURL: baseURL}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (a *API) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Post("/shorten", a.shorten)
	r.Get("/stats/{code}", a.stats)
	r.Get("/{code}", a.redirect)
	return r
}

func (a *API) shorten(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string §json:"url"§
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := dec.Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	code, err := a.store.Create(r.Context(), req.URL)
	switch {
	case errors.Is(err, ErrEmptyURL):
		writeErr(w, http.StatusBadRequest, "url is required")
	case err != nil:
		writeErr(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusCreated, map[string]string{
			"code":      code,
			"short_url": a.baseURL + "/" + code,
		})
	}
}

func (a *API) redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	url, err := a.store.Resolve(r.Context(), code)
	if errors.Is(err, ErrNotFound) {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if err := a.store.Record(r.Context(), code); err != nil {
		// Analytics failure must not break the redirect.
		_ = err
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (a *API) stats(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	url, err := a.store.Resolve(r.Context(), code)
	if errors.Is(err, ErrNotFound) {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	clicks, err := a.store.Clicks(r.Context(), code)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"code":   code,
		"url":    url,
		"clicks": clicks,
	})
}

func main() {
	fmt.Println(NewMemoryStore() != nil)
}
`,
			},
		},
	}
}

func m3stage3() Topic {
	return Topic{
		ID:      "m3s3",
		Title:   "Stage 3 · Middleware, auth & observability",
		Summary: "Wrap the API in a middleware stack and protect the write endpoints.",
		Theory: `
The endpoints work. Now make them operable: know who called, protect the writes,
survive panics, and give every request an identity you can trace.

## The stack

Build the router so middleware is declared once, in one place, in the right
order:

§§§go
func (a *API) Routes() http.Handler {
    r := chi.NewRouter()

    r.Use(RequestID)       // every log line and error gains a correlation id
    r.Use(Recoverer)       // a panic becomes a 500, not a dead server
    r.Use(Timeout(5 * time.Second))

    // Public surface
    r.Post("/shorten", a.shorten)
    r.Get("/{code}", a.redirect)

    // Protected surface
    r.Group(func(r chi.Router) {
        r.Use(RequireAPIKey(a.apiKey))
        r.Get("/stats/{code}", a.stats)
    })

    return r
}
§§§

§chi.Router.Group§ and §chi.Router.Route§ let you attach different middleware to
different subtrees without nesting handlers by hand.

## Where the API key goes

For a server-to-server API, a static key in a header is acceptable *if* it is
compared in constant time and never logged:

§§§go
func RequireAPIKey(want string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            got := r.Header.Get("X-API-Key")
            if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
                w.Header().Set("WWW-Authenticate", "ApiKey")
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
§§§

Things that matter more than the mechanism:

* **Load the key from the environment**, never hard-code it.
* **Never log it.** Not in the request log, not in error messages.
* Return §401§ for a missing/invalid credential and §403§ for a valid credential
  that lacks permission. They mean different things.
* For user-facing apps, prefer short-lived tokens (JWT or opaque sessions) over a
  shared static key.

## Correlation IDs

§§§go
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.Header.Get("X-Request-ID")
        if id == "" {
            id = uuid.NewString()   // or crypto/rand + hex
        }
        w.Header().Set("X-Request-ID", id)
        next.ServeHTTP(w, r.WithContext(WithRequestID(r.Context(), id)))
    })
}
§§§

Then every log line includes it, and you can reconstruct one request's story from
a log store. Add the same field to error responses so a user can quote it to
support.

## Logging the outcome, not the input

A request log line wants: method, path, status, duration, request id, and the
response size. It does **not** want the body, the query string values (they can
contain tokens), or the API key.

Build it from a wrapping §ResponseWriter§ (see the Middleware topic) and emit one
line when the handler returns:

§§§go
func Logger(l *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            rec := &statusRecorder{ResponseWriter: w, status: 200}
            next.ServeHTTP(rec, r)
            l.Info("request",
                "method", r.Method,
                "path", r.URL.Path,
                "status", rec.status,
                "duration_ms", time.Since(start).Milliseconds(),
                "request_id", RequestIDFrom(r.Context()),
            )
        })
    }
}
§§§

## Health endpoints

Add them now, while the router is open:

§§§go
r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
§§§

Note the collision with §GET /{code}§ from stage 2: §/healthz§ is a single segment,
so it *would* match the wildcard. Because specific patterns win, registering
§/healthz§ explicitly is enough — but this is exactly the kind of thing the
stage-2 discussion warned about. Prefixing the redirect route (§/s/{code}§) makes
the whole problem disappear.

## Timeouts and bodies

* §middleware.Timeout(5*time.Second)§ cancels the request context; handlers must
  honour it (your store already takes a §ctx§).
* §http.MaxBytesReader§ on every request body.
* §middleware.Recoverer§ or your own — but there must be one, or a single panic
  takes down the process.
`,
		Quiz: []Question{
			{
				ID:     "m3s3q1",
				Prompt: "Which middleware should be outermost (first to run)?",
				Options: []Option{
					{Text: "Recoverer — so panics in every other middleware and handler are caught.", Correct: true, Rationale: "Anything inside it is protected."},
					{Text: "Auth — so unauthenticated traffic is rejected as early as possible.", Correct: false, Rationale: "Rejecting early is nice, but a panic in auth itself would then be unhandled."},
					{Text: "The router.", Correct: false, Rationale: "The router is the innermost handler."},
					{Text: "Logger.", Correct: false, Rationale: "Logging wants to wrap everything to record the final status, so it sits just inside Recoverer."},
				},
				Explanation: "Typical order: Recoverer → RequestID → Logger → CORS → Timeout → Auth → routes.",
			},
			{
				ID:     "m3s3q2",
				Prompt: "A request has §Authorization: Bearer <valid token>§ but the user lacks the §admin§ role. Which status?",
				Options: []Option{
					{Text: "403 Forbidden", Correct: true, Rationale: "Authentication succeeded; authorisation failed."},
					{Text: "401 Unauthorized", Correct: false, Rationale: "401 means 'I do not know who you are' — credentials are missing or invalid."},
					{Text: "404 Not Found", Correct: false, Rationale: "Sometimes used to hide resource existence, but not the generic answer."},
					{Text: "400 Bad Request", Correct: false, Rationale: "The request is well formed."},
				},
				Explanation: "401 = who are you? 403 = I know you, and no.",
			},
			{
				ID:     "m3s3q3",
				Prompt: "What must never appear in your request logs? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "§Authorization§ / §X-API-Key§ header values", Correct: true, Rationale: "Logs are widely readable; credentials must never land there."},
					{Text: "Passwords and full card numbers in bodies", Correct: true, Rationale: "Same reason, plus compliance requirements."},
					{Text: "The HTTP method and path", Correct: false, Rationale: "These are exactly what you want to log."},
					{Text: "The response status and duration", Correct: false, Rationale: "Also wanted."},
					{Text: "Session cookies", Correct: true, Rationale: "A cookie is a credential — anyone with the log can impersonate the user."},
				},
				Explanation: "Log the shape of the request, not its secrets. Redaction belongs in the logging layer, not in every call site.",
			},
		},
		Exercises: []Exercise{
			{
				ID:     "m3s3e1",
				Title:  "Stage 3 — middleware stack with API-key auth",
				Prompt: "**Project stage 3 of 5.**\n\nThe store and the router skeleton are provided. Add the middleware layer.\n\nImplement:\n\n§§§go\nfunc RequestID(next http.Handler) http.Handler\nfunc Recoverer(next http.Handler) http.Handler\nfunc RequireAPIKey(key string) func(http.Handler) http.Handler\nfunc APIKeyFrom(ctx context.Context) string\nfunc WithAPIKey(ctx context.Context, key string) context.Context\n§§§\n\nAnd wire them into §Routes()§ so that:\n\n* every response carries an §X-Request-ID§ header (echoed from the request, or\n  generated if absent — any non-empty string works; there is no UUID package in the\n  sandbox, so use e.g. §crypto/rand§ with §encoding/hex§, or §fmt.Sprintf(\"%d\", time.Now().UnixNano())§),\n* a panic in any handler becomes §500§ with body §internal server error§,\n* §GET /stats/{code}§ requires a valid §X-API-Key§ header (§401§ otherwise) and the\n  handler must be able to read the key from the context via §APIKeyFrom§,\n* §POST /shorten§ and §GET /{code}§ stay public.\n\nFor the protected §/stats/{code}§ route, return §200§ with\n§{\"code\":...,\"url\":...,\"clicks\":...,\"api_key\":\"<key from context>\"}§.\n\n§chi.Router.Group§ is the clean way to attach middleware to a subtree.",
				Starter: `package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

var (
	ErrNotFound = errors.New("not found")
	ErrEmptyURL = errors.New("url must not be empty")
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Store interface {
	Create(ctx context.Context, url string) (string, error)
	Resolve(ctx context.Context, code string) (string, error)
	Record(ctx context.Context, code string) error
	Clicks(ctx context.Context, code string) (int, error)
}

type memoryStore struct {
	mu     sync.RWMutex
	urls   map[string]string
	clicks map[string]int
	next   uint64
}

func NewMemoryStore() *memoryStore {
	return &memoryStore{urls: map[string]string{}, clicks: map[string]int{}}
}

func (s *memoryStore) Create(ctx context.Context, original string) (string, error) {
	if strings.TrimSpace(original) == "" {
		return "", ErrEmptyURL
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	code := fmt.Sprintf("c%d", s.next)
	s.urls[code] = original
	return code, nil
}

func (s *memoryStore) Resolve(ctx context.Context, code string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.urls[code]
	if !ok {
		return "", ErrNotFound
	}
	return url, nil
}

func (s *memoryStore) Record(ctx context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.urls[code]; !ok {
		return ErrNotFound
	}
	s.clicks[code]++
	return nil
}

func (s *memoryStore) Clicks(ctx context.Context, code string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.urls[code]; !ok {
		return 0, ErrNotFound
	}
	return s.clicks[code], nil
}

// ---------- your work ----------

type ctxKey int

const apiKeyCtxKey ctxKey = 0

func WithAPIKey(ctx context.Context, key string) context.Context {
	// TODO
	return ctx
}

func APIKeyFrom(ctx context.Context) string {
	// TODO
	return ""
}

func RequestID(next http.Handler) http.Handler {
	// TODO
	return next
}

func Recoverer(next http.Handler) http.Handler {
	// TODO
	return next
}

func RequireAPIKey(key string) func(http.Handler) http.Handler {
	// TODO
	return func(next http.Handler) http.Handler { return next }
}

type API struct {
	store   Store
	baseURL string
	apiKey  string
}

func NewAPI(store Store, baseURL, apiKey string) *API {
	return &API{store: store, baseURL: baseURL, apiKey: apiKey}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (a *API) Routes() http.Handler {
	r := chi.NewRouter()
	// TODO: r.Use(RequestID, Recoverer)
	// TODO: public POST /shorten and GET /{code}
	// TODO: protected group with RequireAPIKey for GET /stats/{code}
	return r
}

func (a *API) shorten(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string §json:"url"§
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	code, err := a.store.Create(r.Context(), req.URL)
	if errors.Is(err, ErrEmptyURL) {
		writeErr(w, http.StatusBadRequest, "url is required")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"code": code, "short_url": a.baseURL + "/" + code})
}

func (a *API) redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	url, err := a.store.Resolve(r.Context(), code)
	if errors.Is(err, ErrNotFound) {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	_ = a.store.Record(r.Context(), code)
	http.Redirect(w, r, url, http.StatusFound)
}

func (a *API) stats(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	url, err := a.store.Resolve(r.Context(), code)
	if errors.Is(err, ErrNotFound) {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	clicks, _ := a.store.Clicks(r.Context(), code)
	writeJSON(w, http.StatusOK, map[string]any{
		"code":    code,
		"url":     url,
		"clicks":  clicks,
		"api_key": APIKeyFrom(r.Context()),
	})
}

func main() {
	fmt.Println(NewAPI(NewMemoryStore(), "http://x", "k") != nil)
}

// silence the unused import until you implement RequestID
var _ = alphabet
`,
				Deps: []string{"github.com/go-chi/chi/v5"},
				Tests: []File{{
					Name: "middleware_test.go",
					Content: `package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

const testKey = "s3cret-key"

func newAPI(t *testing.T) (*memoryStore, http.Handler) {
	t.Helper()
	st := NewMemoryStore()
	return st, NewAPI(st, "http://sho.rt", testKey).Routes()
}

func TestRequestIDHeader(t *testing.T) {
	_, h := newAPI(t)
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(§{"url":"https://go.dev"}§))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing X-Request-ID on the response")
	}

	req = httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(§{"url":"https://go.dev"}§))
	req.Header.Set("X-Request-ID", "given-id")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("X-Request-ID"); got != "given-id" {
		t.Fatalf("X-Request-ID = %q, want it echoed", got)
	}
}

func TestRecovererTurnsPanicInto500(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	h := RequestID(Recoverer(inner))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "internal server error") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestStatsRequiresAPIKey(t *testing.T) {
	st, h := newAPI(t)
	st.urls["abc"] = "https://example.com"

	req := httptest.NewRequest(http.MethodGet, "/stats/abc", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no key: status = %d, want 401", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/stats/abc", nil)
	req.Header.Set("X-API-Key", "wrong")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong key: status = %d, want 401", rec.Code)
	}
}

func TestStatsWithAPIKey(t *testing.T) {
	st, h := newAPI(t)
	st.urls["abc"] = "https://example.com"
	st.clicks["abc"] = 2

	req := httptest.NewRequest(http.MethodGet, "/stats/abc", nil)
	req.Header.Set("X-API-Key", testKey)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["api_key"] != testKey {
		t.Fatalf("handler could not read the key from the context: %v", got["api_key"])
	}
	if got["clicks"].(float64) != 2 {
		t.Fatalf("clicks = %v", got["clicks"])
	}
}

func TestPublicRoutesDoNotNeedAKey(t *testing.T) {
	_, h := newAPI(t)

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(§{"url":"https://go.dev"}§))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("shorten status = %d, want 201", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/nope", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown code status = %d, want 404", rec.Code)
	}
}

func TestAPIKeyContextHelpers(t *testing.T) {
	ctx := context.Background()
	if got := APIKeyFrom(ctx); got != "" {
		t.Fatalf("APIKeyFrom(empty) = %q, want empty", got)
	}

	ctx = WithAPIKey(ctx, "first")
	if got := APIKeyFrom(ctx); got != "first" {
		t.Fatalf("APIKeyFrom = %q, want first", got)
	}

	child := WithAPIKey(ctx, "second")
	if got := APIKeyFrom(child); got != "second" {
		t.Fatalf("APIKeyFrom(child) = %q, want second", got)
	}
	if got := APIKeyFrom(ctx); got != "first" {
		t.Fatalf("parent context was mutated: %q", got)
	}
}

func TestRequestIDOnEveryResponse(t *testing.T) {
	st, h := newAPI(t)
	st.urls["abc"] = "https://example.com"

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		key    string
	}{
		{"shorten", http.MethodPost, "/shorten", §{"url":"https://go.dev"}§, ""},
		{"bad json", http.MethodPost, "/shorten", "{", ""},
		{"redirect", http.MethodGet, "/abc", "", ""},
		{"unknown redirect", http.MethodGet, "/missing", "", ""},
		{"stats authorized", http.MethodGet, "/stats/abc", "", testKey},
		{"stats unauthorized", http.MethodGet, "/stats/abc", "", ""},
	}
	for _, c := range cases {
		req := httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
		if c.key != "" {
			req.Header.Set("X-API-Key", c.key)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Header().Get("X-Request-ID") == "" {
			t.Errorf("%s: missing X-Request-ID", c.name)
		}
	}
}

func TestRequestIDIsEchoedOnProtectedRoute(t *testing.T) {
	st, h := newAPI(t)
	st.urls["abc"] = "https://example.com"

	req := httptest.NewRequest(http.MethodGet, "/stats/abc", nil)
	req.Header.Set("X-API-Key", testKey)
	req.Header.Set("X-Request-ID", "trace-42")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-Request-ID"); got != "trace-42" {
		t.Fatalf("X-Request-ID = %q, want the supplied id", got)
	}
}

func TestRecovererIsolatesPanicAndKeepsRequestID(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("secret panic value")
	})
	h := RequestID(Recoverer(inner))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "internal server error") {
		t.Fatalf("body = %q", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret panic value") {
		t.Fatalf("panic value leaked to the client: %q", rec.Body.String())
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("request id was lost while recovering from the panic")
	}
}

func TestStatsAPIKeyVariants(t *testing.T) {
	st, h := newAPI(t)
	st.urls["abc"] = "https://example.com"

	bad := []string{"", "wrong", testKey + "-extra", strings.ToUpper(testKey), " " + testKey}
	for _, key := range bad {
		req := httptest.NewRequest(http.MethodGet, "/stats/abc", nil)
		if key != "" {
			req.Header.Set("X-API-Key", key)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("key %q: status = %d, want 401", key, rec.Code)
		}
		if strings.Contains(rec.Body.String(), testKey) {
			t.Errorf("key %q: 401 body leaked the configured key: %q", key, rec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/stats/abc", nil)
	req.Header.Set("X-API-Key", testKey)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid key: status = %d (body %s)", rec.Code, rec.Body.String())
	}
	var got struct {
		Code   string §json:"code"§
		URL    string §json:"url"§
		Clicks int    §json:"clicks"§
		APIKey string §json:"api_key"§
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Code != "abc" || got.URL != "https://example.com" || got.Clicks != 0 || got.APIKey != testKey {
		t.Fatalf("stats = %+v", got)
	}
}

func TestStatsUnknownCodeWithValidKey(t *testing.T) {
	_, h := newAPI(t)
	req := httptest.NewRequest(http.MethodGet, "/stats/missing", nil)
	req.Header.Set("X-API-Key", testKey)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 after auth passes", rec.Code)
	}
}

func TestAuthAppliesOnlyToStats(t *testing.T) {
	st, h := newAPI(t)
	st.urls["abc"] = "https://example.com"

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(§{"url":"https://go.dev"}§))
	req.Header.Set("X-API-Key", "wrong")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("shorten with a bad key: status = %d, want 201", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/abc", nil)
	req.Header.Set("X-API-Key", "wrong")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("redirect with a bad key: status = %d, want 302", rec.Code)
	}
	if st.clicks["abc"] != 1 {
		t.Fatalf("redirect clicks = %d, want 1", st.clicks["abc"])
	}
}

func TestStatsDoesNotRecordClicks(t *testing.T) {
	st, h := newAPI(t)
	st.urls["abc"] = "https://example.com"

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/stats/abc", nil)
		req.Header.Set("X-API-Key", testKey)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("stats %d: status = %d (body %s)", i, rec.Code, rec.Body.String())
		}
	}
	if st.clicks["abc"] != 0 {
		t.Fatalf("stats recorded clicks: %d", st.clicks["abc"])
	}
}

func TestConcurrentStatsRequests(t *testing.T) {
	st, h := newAPI(t)
	st.urls["abc"] = "https://example.com"

	const n = 20
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/stats/abc", nil)
			authorized := i%2 == 0
			if authorized {
				req.Header.Set("X-API-Key", testKey)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if authorized && rec.Code != http.StatusOK {
				t.Errorf("authorized request %d: status = %d", i, rec.Code)
			}
			if !authorized && rec.Code != http.StatusUnauthorized {
				t.Errorf("unauthorized request %d: status = %d", i, rec.Code)
			}
			if rec.Header().Get("X-Request-ID") == "" {
				t.Errorf("request %d: missing X-Request-ID", i)
			}
		}(i)
	}
	wg.Wait()
}
`,
				}},
				Hints: []string{
					"§r.Use(RequestID, Recoverer)§ registers both, in that order.",
					"§chi.Router.Group§: §r.Group(func(r chi.Router) { r.Use(RequireAPIKey(a.apiKey)); r.Get(\"/stats/{code}\", a.stats) })§.",
					"§RequireAPIKey§ should store the key in the context so the handler can echo it: use §WithAPIKey(r.Context(), got)§ and §r.WithContext(...)§.",
					"Use §crypto/rand§ + §encoding/hex§, or just §fmt.Sprintf(\"%d\", time.Now().UnixNano())§ for the generated request id.",
					"§APIKeyFrom§ uses the comma-ok type assertion on §ctx.Value(apiKeyCtxKey)§.",
				},
				Solution: `package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

var (
	ErrNotFound = errors.New("not found")
	ErrEmptyURL = errors.New("url must not be empty")
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Store interface {
	Create(ctx context.Context, url string) (string, error)
	Resolve(ctx context.Context, code string) (string, error)
	Record(ctx context.Context, code string) error
	Clicks(ctx context.Context, code string) (int, error)
}

type memoryStore struct {
	mu     sync.RWMutex
	urls   map[string]string
	clicks map[string]int
	next   uint64
}

func NewMemoryStore() *memoryStore {
	return &memoryStore{urls: map[string]string{}, clicks: map[string]int{}}
}

func (s *memoryStore) Create(ctx context.Context, original string) (string, error) {
	if strings.TrimSpace(original) == "" {
		return "", ErrEmptyURL
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	code := fmt.Sprintf("c%d", s.next)
	s.urls[code] = original
	return code, nil
}

func (s *memoryStore) Resolve(ctx context.Context, code string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.urls[code]
	if !ok {
		return "", ErrNotFound
	}
	return url, nil
}

func (s *memoryStore) Record(ctx context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.urls[code]; !ok {
		return ErrNotFound
	}
	s.clicks[code]++
	return nil
}

func (s *memoryStore) Clicks(ctx context.Context, code string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.urls[code]; !ok {
		return 0, ErrNotFound
	}
	return s.clicks[code], nil
}

type ctxKey int

const (
	apiKeyCtxKey ctxKey = iota
	reqIDCtxKey
)

func WithAPIKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, apiKeyCtxKey, key)
}

func APIKeyFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	key, _ := ctx.Value(apiKeyCtxKey).(string)
	return key
}

func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), reqIDCtxKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func RequireAPIKey(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := r.Header.Get("X-API-Key")
			if got == "" || got != key {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithAPIKey(r.Context(), got)))
		})
	}
}

type API struct {
	store   Store
	baseURL string
	apiKey  string
}

func NewAPI(store Store, baseURL, apiKey string) *API {
	return &API{store: store, baseURL: baseURL, apiKey: apiKey}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (a *API) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(RequestID)
	r.Use(Recoverer)

	r.Post("/shorten", a.shorten)
	r.Get("/{code}", a.redirect)

	r.Group(func(r chi.Router) {
		r.Use(RequireAPIKey(a.apiKey))
		r.Get("/stats/{code}", a.stats)
	})

	return r
}

func (a *API) shorten(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string §json:"url"§
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	code, err := a.store.Create(r.Context(), req.URL)
	if errors.Is(err, ErrEmptyURL) {
		writeErr(w, http.StatusBadRequest, "url is required")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"code": code, "short_url": a.baseURL + "/" + code})
}

func (a *API) redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	url, err := a.store.Resolve(r.Context(), code)
	if errors.Is(err, ErrNotFound) {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	_ = a.store.Record(r.Context(), code)
	http.Redirect(w, r, url, http.StatusFound)
}

func (a *API) stats(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	url, err := a.store.Resolve(r.Context(), code)
	if errors.Is(err, ErrNotFound) {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	clicks, _ := a.store.Clicks(r.Context(), code)
	writeJSON(w, http.StatusOK, map[string]any{
		"code":    code,
		"url":     url,
		"clicks":  clicks,
		"api_key": APIKeyFrom(r.Context()),
	})
}

func main() {
	fmt.Println(NewAPI(NewMemoryStore(), "http://x", "k") != nil)
}
`,
			},
		},
	}
}

func m3stage4() Topic {
	return Topic{
		ID:      "m3s4",
		Title:   "Stage 4 · SQL persistence",
		Summary: "Swap the memory store for SQLite without touching the HTTP layer.",
		Theory: `
This is the payoff of stage 1's decision to hide storage behind an interface. The
HTTP layer does not change at all; only a new §Store§ implementation appears.

## The schema

§§§sql
CREATE TABLE IF NOT EXISTS links (
    code       TEXT PRIMARY KEY,
    url        TEXT NOT NULL,
    clicks     INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_links_created_at ON links (created_at);
§§§

Design notes:

* The **code is the primary key** — it is how every request addresses a link, so the
  lookup is an index seek, not a scan.
* §created_at§ defaults in the database, so no client clock can lie about it.
* A schema change is a *migration*: versioned, applied in order, never edited after
  it has run anywhere.

## Insert or fail

Two write styles, with different semantics:

§§§go
// Fail on duplicates
_, err := db.ExecContext(ctx, §INSERT INTO links (code, url) VALUES (?, ?)§, code, url)

// Upsert (SQLite/Postgres)
_, err := db.ExecContext(ctx, §
    INSERT INTO links (code, url) VALUES (?, ?)
    ON CONFLICT(code) DO UPDATE SET url = excluded.url
§, code, url)
§§§

Choosing correctly matters: an upsert on a shortener silently rewrites someone
else's link if two goroutines generate the same code. §INSERT§ + retry on conflict
is the safer default.

## Counting clicks atomically

Do not read-modify-write from Go. Let the database do it in one statement:

§§§go
// WRONG: two round trips, and a lost update under concurrency
n := selectClicks(code)
n++
updateClicks(code, n)

// RIGHT: atomic in the database
_, err := db.ExecContext(ctx, §UPDATE links SET clicks = clicks + 1 WHERE code = ?§, code)
§§§

Check §RowsAffected()§ to distinguish "updated" from "no such code" — that is your
§ErrNotFound§ without a second query.

## Transactions and consistency

Analytics that spans several statements (increment a counter *and* append to an
event table) belongs in a transaction, so it is all-or-nothing:

§§§go
tx, err := db.BeginTx(ctx, nil)
if err != nil { return err }
defer tx.Rollback()

if _, err := tx.ExecContext(ctx, §UPDATE links SET clicks = clicks + 1 WHERE code = ?§, code); err != nil {
    return err
}
if _, err := tx.ExecContext(ctx, §INSERT INTO events (code, at) VALUES (?, ?)§, code, time.Now()); err != nil {
    return err
}
return tx.Commit()
§§§

## Testing SQL

Run against a **real** database. SQLite in a temp file is fast, needs no server,
and catches real SQL errors:

§§§go
func newDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := Open("file:" + filepath.Join(t.TempDir(), "test.db"))
    if err != nil { t.Fatal(err) }
    t.Cleanup(func() { db.Close() })
    return db
}
§§§

Isolation comes free: every test gets its own database file and its own §t.TempDir()§.

> **SQLite concurrency:** SQLite serialises writers. Keep §SetMaxOpenConns(1)§ (or
> small) for file databases, and enable WAL (§PRAGMA journal_mode=WAL§) if you need
> concurrent readers alongside a writer.

## Postgres in production

The same code works against Postgres with two changes: the driver
(§github.com/jackc/pgx/v5/stdlib§ or §lib/pq§) and the placeholders (§$1§ instead
of §?§). §sqlx.Rebind§ converts between them:

§§§go
q = db.Rebind(§SELECT * FROM links WHERE code = ?§)   // -> $1
§§§

Keep the SQL in one place per repository type, and use §QueryContext§ everywhere
so the request context still cancels the query when a client disconnects.
`,
		Quiz: []Question{
			{
				ID:     "m3s4q1",
				Prompt: "Why is §UPDATE links SET clicks = clicks + 1 WHERE code = ?§ better than reading, incrementing in Go, and writing back?",
				Options: []Option{
					{Text: "It is a single atomic statement, so concurrent clicks cannot lose an update.", Correct: true, Rationale: "Read-modify-write from Go has a race window and needs a transaction or lock."},
					{Text: "It uses fewer bytes on the wire.", Correct: false, Rationale: "Efficiency is a side benefit, not the reason."},
					{Text: "It is the only syntax SQLite supports.", Correct: false, Rationale: "Both are valid SQL."},
					{Text: "It avoids needing an index.", Correct: false, Rationale: "You still want the primary-key index."},
				},
				Explanation: "Push invariants into the database whenever it can enforce them atomically.",
			},
			{
				ID:     "m3s4q2",
				Prompt: "Why does the same §Store§ code work against SQLite and Postgres with only a driver change?",
				Options: []Option{
					{Text: "Because §database/sql§ is a generic interface and the SQL used is standard; only placeholders differ, and §Rebind§ handles that.", Correct: true, Rationale: "Keeping SQL portable is a deliberate design choice."},
					{Text: "Because SQLite speaks the Postgres wire protocol.", Correct: false, Rationale: "It does not."},
					{Text: "Because Go compiles SQL at build time.", Correct: false, Rationale: "SQL is a string sent at runtime."},
					{Text: "Because §database/sql§ translates dialects automatically.", Correct: false, Rationale: "It does not translate SQL; it only manages drivers and connections."},
				},
				Explanation: "Sticking to ANSI-ish SQL plus §Rebind§ is what makes the swap practical.",
			},
			{
				ID:     "m3s4q3",
				Prompt: "A §Store.Create§ uses §INSERT§ and the code already exists. What is the correct response?",
				Options: []Option{
					{Text: "Treat it as a collision: generate another code (or return a conflict) — never overwrite the existing link.", Correct: true, Rationale: "Overwriting silently redirects another user's traffic."},
					{Text: "Use an upsert so the new URL wins.", Correct: false, Rationale: "That is exactly the dangerous behaviour."},
					{Text: "Return 500 and move on.", Correct: false, Rationale: "A collision is expected and recoverable, not an internal error."},
					{Text: "Delete the old row and insert the new one.", Correct: false, Rationale: "Same problem as the upsert."},
				},
				Explanation: "Detecting and handling the unique-violation is a normal part of code generation.",
			},
		},
		Exercises: []Exercise{
			{
				ID:     "m3s4e1",
				Title:  "Stage 4 — SQLite-backed store",
				Prompt: "**Project stage 4 of 5.**\n\nImplement the §Store§ interface on top of SQLite. The interface (and therefore\nthe whole HTTP layer) does not change.\n\n§§§go\nfunc Open(dsn string) (*sql.DB, error)   // open, ping, migrate, sane pool settings\nfunc NewSQLStore(db *sql.DB) *SQLStore\n\nfunc (s *SQLStore) Create(ctx context.Context, url string) (string, error)\nfunc (s *SQLStore) Resolve(ctx context.Context, code string) (string, error)\nfunc (s *SQLStore) Record(ctx context.Context, code string) error\nfunc (s *SQLStore) Clicks(ctx context.Context, code string) (int, error)\nfunc (s *SQLStore) Stats(ctx context.Context) (links int, clicks int, err error)\n§§§\n\nRequirements:\n\n* Schema: §links(code TEXT PRIMARY KEY, url TEXT NOT NULL, clicks INTEGER NOT NULL DEFAULT 0, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)§.\n* §Create§ generates a unique code; on a uniqueness conflict it must retry rather\n  than overwrite. §ErrEmptyURL§ for blank input. There is no sentinel error for a\n  collision — in SQLite a primary-key clash comes back as an ordinary error whose text\n  contains §\"UNIQUE\"§.\n* §Record§ must increment atomically in SQL and return §ErrNotFound§ when no row\n  matched.\n* §Resolve§/§Clicks§ return §ErrNotFound§ for unknown codes.\n* §Stats§ returns the number of links and the total clicks, using §sqlx§.\n* §NewSQLStore§ receives a §*sql.DB§, but the struct field is a §*sqlx.DB§; bridge\n  them with §sqlx.NewDb(db, \"sqlite\")§.\n\nUse **sqlx** for §Stats§ (§db.Get§ with §COALESCE(SUM(clicks),0)§) and plain\n§database/sql§ everywhere else.",
				Starter: `package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var (
	ErrNotFound = errors.New("not found")
	ErrEmptyURL = errors.New("url must not be empty")
)

type Store interface {
	Create(ctx context.Context, url string) (string, error)
	Resolve(ctx context.Context, code string) (string, error)
	Record(ctx context.Context, code string) error
	Clicks(ctx context.Context, code string) (int, error)
}

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomCode(n int) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	v := binary.BigEndian.Uint64(b)
	out := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, alphabet[v%62])
		v /= 62
	}
	return string(out)
}

func Open(dsn string) (*sql.DB, error) {
	// TODO: sql.Open("sqlite", dsn), pool settings, ping, migrate
	return nil, nil
}

type SQLStore struct {
	db *sqlx.DB
}

func NewSQLStore(db *sql.DB) *SQLStore {
	// TODO
	return nil
}

func (s *SQLStore) Create(ctx context.Context, url string) (string, error) {
	// TODO: insert with retry on conflict
	return "", nil
}

func (s *SQLStore) Resolve(ctx context.Context, code string) (string, error) {
	// TODO
	return "", nil
}

func (s *SQLStore) Record(ctx context.Context, code string) error {
	// TODO
	return nil
}

func (s *SQLStore) Clicks(ctx context.Context, code string) (int, error) {
	// TODO
	return 0, nil
}

func (s *SQLStore) Stats(ctx context.Context) (links int, clicks int, err error) {
	// TODO
	return 0, 0, nil
}

func main() {
	db, err := Open("file:demo.db?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	st := NewSQLStore(db)
	code, _ := st.Create(context.Background(), "https://go.dev")
	fmt.Println(code, strings.ToUpper(code) != "", time.Now().Year() > 2000)
}
`,
				Deps: []string{"github.com/jmoiron/sqlx", "modernc.org/sqlite"},
				Tests: []File{{
					Name: "sqlstore_test.go",
					Content: `package main

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func newStore(t *testing.T) *SQLStore {
	t.Helper()
	db, err := Open("file:" + filepath.Join(t.TempDir(), "links.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewSQLStore(db)
}

func TestCreateResolve(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()

	code, err := st.Create(ctx, "https://go.dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if code == "" || strings.ContainsAny(code, "/?# ") {
		t.Fatalf("bad code %q", code)
	}
	got, err := st.Resolve(ctx, code)
	if err != nil || got != "https://go.dev" {
		t.Fatalf("Resolve = %q, %v", got, err)
	}
}

func TestCreateValidation(t *testing.T) {
	st := newStore(t)
	if _, err := st.Create(context.Background(), "   "); !errors.Is(err, ErrEmptyURL) {
		t.Fatalf("Create(blank) = %v, want ErrEmptyURL", err)
	}
}

func TestCreateManyUnique(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		code, err := st.Create(ctx, "https://example.com")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if seen[code] {
			t.Fatalf("duplicate code %q", code)
		}
		seen[code] = true
	}
	links, _, err := st.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if links != 100 {
		t.Fatalf("links = %d, want 100", links)
	}
}

func TestRecordAndClicks(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	code, _ := st.Create(ctx, "https://go.dev")

	for i := 0; i < 5; i++ {
		if err := st.Record(ctx, code); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
	n, err := st.Clicks(ctx, code)
	if err != nil || n != 5 {
		t.Fatalf("Clicks = %d, %v; want 5", n, err)
	}
}

func TestNotFound(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	if _, err := st.Resolve(ctx, "nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Resolve = %v", err)
	}
	if err := st.Record(ctx, "nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Record = %v", err)
	}
	if _, err := st.Clicks(ctx, "nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Clicks = %v", err)
	}
}

func TestStats(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	a, _ := st.Create(ctx, "https://a.dev")
	b, _ := st.Create(ctx, "https://b.dev")
	_ = st.Record(ctx, a)
	_ = st.Record(ctx, a)
	_ = st.Record(ctx, b)

	links, clicks, err := st.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if links != 2 || clicks != 3 {
		t.Fatalf("Stats = %d links, %d clicks; want 2, 3", links, clicks)
	}
}

func TestConcurrentRecord(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	code, _ := st.Create(ctx, "https://go.dev")

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if err := st.Record(ctx, code); err != nil {
					t.Errorf("Record: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()

	n, err := st.Clicks(ctx, code)
	if err != nil {
		t.Fatal(err)
	}
	if n != 200 {
		t.Fatalf("clicks = %d, want 200 (lost updates?)", n)
	}
}

var _ Store = (*SQLStore)(nil)
var _ = sql.ErrNoRows

func openStoreAt(t *testing.T, path string) (*SQLStore, *sql.DB) {
	t.Helper()
	db, err := Open("file:" + path)
	if err != nil {
		t.Fatalf("Open(%q): %v", path, err)
	}
	return NewSQLStore(db), db
}

func TestStatsOnEmptyStore(t *testing.T) {
	st := newStore(t)
	links, clicks, err := st.Stats(context.Background())
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if links != 0 || clicks != 0 {
		t.Fatalf("Stats = %d links, %d clicks; want 0, 0", links, clicks)
	}
}

func TestCreateRejectsBlankVariants(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	for _, bad := range []string{"", " ", "\t\n", "\u00a0"} {
		if _, err := st.Create(ctx, bad); !errors.Is(err, ErrEmptyURL) {
			t.Fatalf("Create(%q) = %v, want ErrEmptyURL", bad, err)
		}
	}
	links, clicks, err := st.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if links != 0 || clicks != 0 {
		t.Fatalf("rejected creates wrote rows: %d links, %d clicks", links, clicks)
	}
}

func TestRoundTripSpecialURLs(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	urls := []string{
		"https://go.dev/doc/?q=1#top",
		"https://example.com/a%20b?x=1&y=2",
		"https://例え.jp/パス",
		"https://example.com/" + strings.Repeat("long", 100),
	}

	seen := map[string]bool{}
	for _, u := range urls {
		code, err := st.Create(ctx, u)
		if err != nil {
			t.Fatalf("Create(%q): %v", u, err)
		}
		if code == "" || strings.ContainsAny(code, "/?# ") {
			t.Fatalf("bad code %q for %q", code, u)
		}
		if seen[code] {
			t.Fatalf("duplicate code %q", code)
		}
		seen[code] = true
		got, err := st.Resolve(ctx, code)
		if err != nil || got != u {
			t.Fatalf("Resolve(%q) = %q, %v; want %q", code, got, err, u)
		}
	}
}

func TestCreateSameURLGetsDistinctCodes(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	const u = "https://example.com/dup"

	seen := map[string]bool{}
	for i := 0; i < 5; i++ {
		code, err := st.Create(ctx, u)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if seen[code] {
			t.Fatalf("duplicate code %q", code)
		}
		seen[code] = true
		got, err := st.Resolve(ctx, code)
		if err != nil || got != u {
			t.Fatalf("Resolve(%q) = %q, %v", code, got, err)
		}
	}
	links, clicks, err := st.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if links != 5 || clicks != 0 {
		t.Fatalf("Stats = %d links, %d clicks; want 5, 0", links, clicks)
	}
}

func TestClicksArePerCode(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	a, err := st.Create(ctx, "https://a.dev")
	if err != nil {
		t.Fatal(err)
	}
	b, err := st.Create(ctx, "https://b.dev")
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		if err := st.Record(ctx, a); err != nil {
			t.Fatalf("Record(a): %v", err)
		}
	}
	if err := st.Record(ctx, b); err != nil {
		t.Fatalf("Record(b): %v", err)
	}

	if n, err := st.Clicks(ctx, a); err != nil || n != 2 {
		t.Fatalf("Clicks(a) = %d, %v; want 2", n, err)
	}
	if n, err := st.Clicks(ctx, b); err != nil || n != 1 {
		t.Fatalf("Clicks(b) = %d, %v; want 1", n, err)
	}
	links, clicks, err := st.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if links != 2 || clicks != 3 {
		t.Fatalf("Stats = %d links, %d clicks; want 2, 3", links, clicks)
	}
}

func TestRecordUnknownLeavesStoreUntouched(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	if err := st.Record(ctx, "ghost"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Record(ghost) = %v, want ErrNotFound", err)
	}
	if _, err := st.Resolve(ctx, "ghost"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Resolve(ghost) = %v, want ErrNotFound", err)
	}
	links, clicks, err := st.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if links != 0 || clicks != 0 {
		t.Fatalf("failed Record created state: %d links, %d clicks", links, clicks)
	}
}

func TestCreateDoesNotOverwriteExistingCode(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	if _, err := st.db.ExecContext(ctx, "INSERT INTO links (code, url) VALUES (?, ?)", "fixed", "https://original.example"); err != nil {
		t.Fatalf("seed row: %v", err)
	}

	code, err := st.Create(ctx, "https://new.example")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if code == "fixed" {
		t.Fatal("Create reused an existing code")
	}
	got, err := st.Resolve(ctx, "fixed")
	if err != nil || got != "https://original.example" {
		t.Fatalf("existing row was overwritten: %q, %v", got, err)
	}
	links, _, err := st.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if links != 2 {
		t.Fatalf("links = %d, want 2", links)
	}
}

func TestPersistenceAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "persist.db")
	ctx := context.Background()

	st, db := openStoreAt(t, path)
	code, err := st.Create(ctx, "https://persist.dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := st.Record(ctx, code); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	st2, db2 := openStoreAt(t, path)
	defer db2.Close()

	got, err := st2.Resolve(ctx, code)
	if err != nil || got != "https://persist.dev" {
		t.Fatalf("Resolve after reopen = %q, %v", got, err)
	}
	n, err := st2.Clicks(ctx, code)
	if err != nil || n != 3 {
		t.Fatalf("Clicks after reopen = %d, %v; want 3", n, err)
	}
	links, clicks, err := st2.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if links != 1 || clicks != 3 {
		t.Fatalf("Stats after reopen = %d links, %d clicks; want 1, 3", links, clicks)
	}
}

func TestConcurrentCreateUnique(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	const goroutines = 10
	const each = 5

	codes := make(chan string, goroutines*each)
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < each; i++ {
				code, err := st.Create(ctx, "https://example.com")
				if err != nil {
					t.Errorf("Create: %v", err)
					return
				}
				codes <- code
			}
		}()
	}
	wg.Wait()
	close(codes)

	seen := map[string]bool{}
	for code := range codes {
		if seen[code] {
			t.Errorf("duplicate code %q", code)
		}
		seen[code] = true
	}
	if len(seen) != goroutines*each {
		t.Fatalf("got %d distinct codes, want %d", len(seen), goroutines*each)
	}
	links, _, err := st.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if links != goroutines*each {
		t.Fatalf("links = %d, want %d", links, goroutines*each)
	}
}

func TestOpenFailsForMissingParentDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-dir", "links.db")
	db, err := Open("file:" + path)
	if err == nil {
		if db != nil {
			db.Close()
		}
		t.Fatal("Open succeeded for a database whose parent directory does not exist")
	}
}
`,
				}},
				Hints: []string{
					"§sql.Open(\"sqlite\", dsn)§, then §db.SetMaxOpenConns(1)§ (SQLite serialises writers), §PingContext§, then the §CREATE TABLE IF NOT EXISTS§.",
					"Code generation with retry: loop up to ~5 times, §INSERT§, and on a unique-constraint error try again.",
					"Detect the conflict by checking §strings.Contains(err.Error(), \"UNIQUE\")§ — pragmatic and driver-agnostic enough for this exercise.",
					"§Record§: §res, err := db.ExecContext(ctx, \"UPDATE links SET clicks = clicks + 1 WHERE code = ?\", code)§ then §n, _ := res.RowsAffected(); if n == 0 { return ErrNotFound }§.",
					"§Stats§ uses sqlx: §db.Get(&clicks, \"SELECT COALESCE(SUM(clicks),0) FROM links\")§ and §db.Get(&links, \"SELECT COUNT(*) FROM links\")§.",
				},
				Solution: `package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var (
	ErrNotFound = errors.New("not found")
	ErrEmptyURL = errors.New("url must not be empty")
)

type Store interface {
	Create(ctx context.Context, url string) (string, error)
	Resolve(ctx context.Context, code string) (string, error)
	Record(ctx context.Context, code string) error
	Clicks(ctx context.Context, code string) (int, error)
}

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomCode(n int) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	v := binary.BigEndian.Uint64(b)
	out := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, alphabet[v%62])
		v /= 62
	}
	return string(out)
}

const schema = §
CREATE TABLE IF NOT EXISTS links (
    code       TEXT PRIMARY KEY,
    url        TEXT NOT NULL,
    clicks     INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)§

func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

type SQLStore struct {
	db *sqlx.DB
}

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: sqlx.NewDb(db, "sqlite")}
}

func (s *SQLStore) Create(ctx context.Context, url string) (string, error) {
	if strings.TrimSpace(url) == "" {
		return "", ErrEmptyURL
	}
	for attempt := 0; attempt < 5; attempt++ {
		code := randomCode(6)
		_, err := s.db.ExecContext(ctx,
			§INSERT INTO links (code, url) VALUES (?, ?)§, code, url)
		if err == nil {
			return code, nil
		}
		if strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
			continue // collision: try another code
		}
		return "", err
	}
	return "", errors.New("could not generate a unique code")
}

func (s *SQLStore) Resolve(ctx context.Context, code string) (string, error) {
	var url string
	err := s.db.QueryRowContext(ctx, §SELECT url FROM links WHERE code = ?§, code).Scan(&url)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *SQLStore) Record(ctx context.Context, code string) error {
	res, err := s.db.ExecContext(ctx,
		§UPDATE links SET clicks = clicks + 1 WHERE code = ?§, code)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLStore) Clicks(ctx context.Context, code string) (int, error) {
	var clicks int
	err := s.db.QueryRowContext(ctx, §SELECT clicks FROM links WHERE code = ?§, code).Scan(&clicks)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	return clicks, nil
}

func (s *SQLStore) Stats(ctx context.Context) (int, int, error) {
	var links int
	if err := s.db.GetContext(ctx, &links, §SELECT COUNT(*) FROM links§); err != nil {
		return 0, 0, err
	}
	var clicks int
	if err := s.db.GetContext(ctx, &clicks, §SELECT COALESCE(SUM(clicks), 0) FROM links§); err != nil {
		return 0, 0, err
	}
	return links, clicks, nil
}

func main() {
	db, err := Open("file:demo.db?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	st := NewSQLStore(db)
	code, _ := st.Create(context.Background(), "https://go.dev")
	fmt.Println(code)
}
`,
			},
		},
	}
}

func m3stage5() Topic {
	return Topic{
		ID:      "m3s5",
		Title:   "Stage 5 · Analytics, worker pools & graceful shutdown",
		Summary: "Move click recording off the request path and shut everything down cleanly.",
		Theory: `
Recording a click on the redirect path costs a database write per request. At
scale you want the redirect fast and the analytics eventually consistent. That is
a classic producer/consumer pipeline.

## The shape

§§§text
HTTP handler ──Record(event)──▶ [ buffered channel ] ──▶ N worker goroutines ──▶ store
   (never blocks)                    (bounded)              (batched writes)
§§§

§§§go
type Analytics struct {
    events  chan ClickEvent
    wg      sync.WaitGroup
    stopped atomic.Bool
    counts  sync.Map   // code -> *atomic.Int64
}

func NewAnalytics(store Store, workers, buffer int) *Analytics {
    a := &Analytics{events: make(chan ClickEvent, buffer)}
    for i := 0; i < workers; i++ {
        a.wg.Add(1)
        go a.worker()
    }
    return a
}

func (a *Analytics) worker() {
    defer a.wg.Done()
    for e := range a.events {          // ends when the channel is closed
        _ = a.store.Record(context.Background(), e.Code)
        a.bump(e.Code)
    }
}
§§§

## Non-blocking enqueue

The request path must not block on analytics. A full buffer should drop (and
count) the event rather than add latency:

§§§go
func (a *Analytics) Record(e ClickEvent) bool {
    if a.stopped.Load() {
        return false
    }
    select {
    case a.events <- e:
        return true
    default:
        a.dropped.Add(1)
        return false
    }
}
§§§

Whether dropping is acceptable is a product decision — but "the redirect took 4
seconds because the analytics buffer was full" never is.

## Shutdown

Order matters, and it is the reverse of startup:

§§§go
func (a *Analytics) Stop() {
    if !a.stopped.CompareAndSwap(false, true) {
        return                       // idempotent
    }
    close(a.events)                  // signal: no more work
    a.wg.Wait()                      // drain every worker
}
§§§

§close(chan)§ + §WaitGroup§ is the canonical "drain and stop" pair:

* Closing tells every §range§ over the channel to finish after the buffer is empty.
* §wg.Wait()§ guarantees the workers have actually returned, so nothing writes to
  the database after §main§ has moved on.

Do **not** close the channel from the receiving side, and do not call §Stop§ twice —
§close§ of a closed channel panics. §CompareAndSwap§ makes it idempotent.

## Batching (the next step)

At higher throughput, workers become a single goroutine that batches:

§§§go
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()
batch := make([]ClickEvent, 0, 256)

flush := func() {
    if len(batch) == 0 { return }
    if err := a.store.RecordBatch(ctx, batch); err != nil {
        slog.Error("flush batch", "err", err, "size", len(batch))
    }
    batch = batch[:0]
}

for {
    select {
    case e, ok := <-a.events:
        if !ok { flush(); return }
        batch = append(batch, e)
        if len(batch) >= 256 { flush() }
    case <-ticker.C:
        flush()
    }
}
§§§

One §INSERT§ per 256 clicks instead of per click. Note the §flush()§ on the
§!ok§ path: shutdown must never lose buffered work.

## Wiring it into main

§§§go
func run(ctx context.Context) error {
    db, err := Open(cfg.DatabaseURL)
    if err != nil { return err }
    defer db.Close()

    store := NewSQLStore(db)
    analytics := NewAnalytics(store, 4, 1024)
    defer analytics.Stop()                      // runs before db.Close()? No — LIFO!

    srv := &http.Server{Handler: NewAPI(store, analytics, cfg.BaseURL).Routes()}

    ...
}
§§§

Careful with §defer§ order: defers run LIFO, so §defer db.Close()§ registered
first runs *last*. That is exactly right here — analytics must stop before the
database closes.

Order of operations on shutdown:

1. stop accepting HTTP traffic (§srv.Shutdown§),
2. drain the analytics queue (§analytics.Stop§),
3. close the database (§db.Close§),
4. exit.

## What the tests check

Because the pipeline is asynchronous, tests must **synchronise**, not sleep:

* poll with a deadline: §for time.Now().Before(deadline) { if total == n { break }; time.Sleep(time.Millisecond) }§
* or expose a §Wait()§/§Flush()§ method that blocks until the queue is empty.

A test that sleeps and hopes is a flaky test. Never do it.
`,
		Quiz: []Question{
			{
				ID:     "m3s5q1",
				Prompt: "Why does §Stop§ close the channel and then call §wg.Wait()§?",
				Options: []Option{
					{Text: "Closing ends every worker's §range§ loop; Wait ensures they have all finished before the caller proceeds.", Correct: true, Rationale: "Without Wait you could close the database while a worker is still writing."},
					{Text: "§wg.Wait()§ closes the channel for you.", Correct: false, Rationale: "It does not touch the channel."},
					{Text: "Closing the channel cancels the workers immediately.", Correct: false, Rationale: "Workers finish the buffered events first."},
					{Text: "It prevents deadlocks caused by §close§.", Correct: false, Rationale: "Order here is about completeness, not deadlock."},
				},
				Explanation: "close signals 'no more sends'; WaitGroup waits for 'all receivers finished'. You need both.",
			},
			{
				ID:     "m3s5q2",
				Prompt: "The analytics buffer is full during a traffic spike. What should §Record§ do?",
				Options: []Option{
					{Text: "Drop the event (and count the drop), so the redirect stays fast.", Correct: true, Rationale: "Latency on the user path is the thing you must protect."},
					{Text: "Block until there is room.", Correct: false, Rationale: "That turns an analytics backlog into user-visible latency."},
					{Text: "Panic so the problem is visible.", Correct: false, Rationale: "Never crash the request path."},
					{Text: "Spawn an unbounded number of goroutines.", Correct: false, Rationale: "That just moves the problem to memory."},
				},
				Explanation: "Choose explicitly between dropping, blocking and scaling. Dropping with a visible metric is usually right for analytics.",
			},
			{
				ID:     "m3s5q3",
				Prompt: "In §run()§, §defer db.Close()§ is registered before §defer analytics.Stop()§. What runs first at return?",
				Options: []Option{
					{Text: "§analytics.Stop()§ — defers run LIFO, so the last registered defer runs first.", Correct: true, Rationale: "Which is what we want: analytics drains before the database closes."},
					{Text: "§db.Close()§ — registration order.", Correct: false, Rationale: "Deferred calls run in reverse order."},
					{Text: "Both simultaneously.", Correct: false, Rationale: "They run sequentially."},
					{Text: "Neither, unless §os.Exit§ is called.", Correct: false, Rationale: "Defers run when the function returns."},
				},
				Explanation: "LIFO defer order is why resource acquisition and its defer are written on adjacent lines.",
			},
			{
				ID:     "m3s5q4",
				Prompt: "How should the test wait for asynchronous work to finish?",
				Options: []Option{
					{Text: "Poll a counter with a deadline, or call a blocking §Stop()§/§Flush()§.", Correct: true, Rationale: "Either is deterministic; Stop both drains and waits."},
					{Text: "§time.Sleep(100 * time.Millisecond)§.", Correct: false, Rationale: "Sleeps make tests slow and flaky."},
					{Text: "It cannot be tested.", Correct: false, Rationale: "It can, with proper synchronisation."},
					{Text: "Read the channel from the test to drain it.", Correct: false, Rationale: "Consuming in the test steals work from the workers."},
				},
				Explanation: "Deterministic synchronisation (channels, WaitGroups, counted semaphores) is what makes concurrency testable.",
			},
		},
		Exercises: []Exercise{
			{
				ID:     "m3s5e1",
				Title:  "Stage 5 — analytics worker pool",
				Prompt: "**Project stage 5 of 5.**\n\nBuild the non-blocking analytics pipeline.\n\n§§§go\ntype ClickEvent struct {\n    Code string\n    At   time.Time\n}\n\ntype Recorder interface {\n    Record(ctx context.Context, code string) error\n}\n\nfunc NewAnalytics(rec Recorder, workers, buffer int) *Analytics\nfunc (a *Analytics) Record(e ClickEvent) bool   // never blocks\nfunc (a *Analytics) Total(code string) int\nfunc (a *Analytics) Dropped() int64\nfunc (a *Analytics) Pending() int               // events queued but not yet processed\nfunc (a *Analytics) Stop()                      // idempotent: drain, wait, return\n§§§\n\nBehaviour:\n\n* §workers < 1§ is treated as 1; §buffer < 0§ as 0 (an unbuffered channel).\n* §Record§ hands the event to a worker through a buffered channel. If the buffer is\n  full, the event is dropped, counted in §Dropped()§, and §Record§ returns false.\n* Every accepted event must eventually reach the §Recorder§ exactly once and be\n  counted in §Total§.\n* After §Stop§, §Record§ must return false and never panic. §Stop§ called twice must\n  not panic either.\n* §Pending()§ reports how many accepted events are still queued. Go's built-in §len§\n  works on channels too: §len(ch)§ is the number of values currently sitting in a\n  buffered channel.\n\nThe test accepts events from many goroutines, then uses §Stop§ as the barrier —\nwhich is precisely why §Stop§ must drain before returning.",
				Starter: `package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ClickEvent struct {
	Code string
	At   time.Time
}

type Recorder interface {
	Record(ctx context.Context, code string) error
}

type Analytics struct {
	events  chan ClickEvent
	rec     Recorder
	wg      sync.WaitGroup
	stopped atomic.Bool
	dropped atomic.Int64

	mu     sync.RWMutex
	counts map[string]*atomic.Int64
}

func NewAnalytics(rec Recorder, workers, buffer int) *Analytics {
	// TODO: create the channel, seed the counts map and start the workers
	return nil
}

func (a *Analytics) Record(e ClickEvent) bool {
	// TODO: non-blocking
	return false
}

func (a *Analytics) Total(code string) int {
	// TODO
	return 0
}

func (a *Analytics) Dropped() int64 {
	// TODO
	return 0
}

func (a *Analytics) Pending() int {
	// TODO
	return 0
}

func (a *Analytics) Stop() {
	// TODO
}

func main() {
	fmt.Println(NewAnalytics(nil, 2, 8) != nil, strings.TrimSpace("x") != "")
}
`,
				Tests: []File{{
					Name: "analytics_test.go",
					Content: `package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type countingRecorder struct {
	calls atomic.Int64
	delay time.Duration
}

func (c *countingRecorder) Record(ctx context.Context, code string) error {
	c.calls.Add(1)
	if c.delay > 0 {
		time.Sleep(c.delay)
	}
	return nil
}

func TestRecordReachesRecorder(t *testing.T) {
	rec := &countingRecorder{}
	a := NewAnalytics(rec, 3, 64)

	const n = 50
	for i := 0; i < n; i++ {
		if !a.Record(ClickEvent{Code: "abc", At: time.Now()}) {
			t.Fatalf("event %d was dropped unexpectedly", i)
		}
	}
	a.Stop()

	if got := a.Total("abc"); got != n {
		t.Fatalf("Total = %d, want %d", got, n)
	}
	if got := rec.calls.Load(); got != int64(n) {
		t.Fatalf("Recorder saw %d events, want %d", got, n)
	}
	if a.Pending() != 0 {
		t.Fatalf("Pending = %d after Stop, want 0", a.Pending())
	}
}

func TestConcurrentRecord(t *testing.T) {
	rec := &countingRecorder{}
	a := NewAnalytics(rec, 8, 4096)

	const goroutines = 20
	const each = 100

	var wg sync.WaitGroup
	var accepted atomic.Int64
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < each; i++ {
				if a.Record(ClickEvent{Code: "x", At: time.Now()}) {
					accepted.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	a.Stop()

	if accepted.Load() == 0 {
		t.Fatal("no events were accepted")
	}
	if got := int64(a.Total("x")); got != accepted.Load() {
		t.Fatalf("Total = %d, want %d", got, accepted.Load())
	}
	if got := rec.calls.Load(); got != accepted.Load() {
		t.Fatalf("Recorder saw %d, want %d", got, accepted.Load())
	}
}

func TestDropsWhenFull(t *testing.T) {
	rec := &countingRecorder{delay: 20 * time.Millisecond}
	a := NewAnalytics(rec, 1, 2)

	dropped := 0
	for i := 0; i < 200; i++ {
		if !a.Record(ClickEvent{Code: "y", At: time.Now()}) {
			dropped++
		}
	}
	a.Stop()

	if dropped == 0 {
		t.Fatal("expected some events to be dropped with a tiny buffer")
	}
	if a.Dropped() != int64(dropped) {
		t.Fatalf("Dropped() = %d, want %d", a.Dropped(), dropped)
	}
	total := int64(a.Total("y")) + a.Dropped()
	if total != 200 {
		t.Fatalf("accepted + dropped = %d, want 200", total)
	}
}

func TestRecordAfterStop(t *testing.T) {
	rec := &countingRecorder{}
	a := NewAnalytics(rec, 2, 4)
	a.Stop()
	if a.Record(ClickEvent{Code: "z"}) {
		t.Fatal("Record accepted an event after Stop")
	}
	a.Stop() // must not panic
}

func TestStopIsIdempotent(t *testing.T) {
	a := NewAnalytics(&countingRecorder{}, 4, 16)
	for i := 0; i < 10; i++ {
		a.Record(ClickEvent{Code: "q"})
	}
	a.Stop()
	first := a.Total("q")
	a.Stop()
	a.Stop()
	if a.Total("q") != first {
		t.Fatalf("Total changed after extra Stop calls: %d -> %d", first, a.Total("q"))
	}
}

func TestWorkerCountFloor(t *testing.T) {
	rec := &countingRecorder{}
	a := NewAnalytics(rec, 0, 64)
	for i := 0; i < 10; i++ {
		if !a.Record(ClickEvent{Code: "w"}) {
			t.Fatalf("event %d was dropped with an ample buffer", i)
		}
	}
	a.Stop()
	if a.Total("w") != 10 {
		t.Fatalf("Total = %d, want 10 with workers clamped to 1", a.Total("w"))
	}
	if rec.calls.Load() != 10 {
		t.Fatalf("Recorder saw %d events, want 10", rec.calls.Load())
	}
}

type gateRecorder struct {
	calls   atomic.Int64
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func newGateRecorder() *gateRecorder {
	return &gateRecorder{started: make(chan struct{}), release: make(chan struct{})}
}

func (g *gateRecorder) Record(ctx context.Context, code string) error {
	g.calls.Add(1)
	g.once.Do(func() { close(g.started) })
	<-g.release
	return nil
}

func TestPendingReportsQueuedEvents(t *testing.T) {
	rec := newGateRecorder()
	a := NewAnalytics(rec, 1, 10)

	if !a.Record(ClickEvent{Code: "p", At: time.Now()}) {
		t.Fatal("first event was dropped")
	}
	select {
	case <-rec.started:
	case <-time.After(2 * time.Second):
		t.Fatal("worker never started")
	}

	for i := 0; i < 3; i++ {
		if !a.Record(ClickEvent{Code: "p", At: time.Now()}) {
			t.Fatalf("event %d was dropped with an ample buffer", i)
		}
	}
	if got := a.Pending(); got != 3 {
		t.Fatalf("Pending = %d, want 3 while the worker is blocked", got)
	}

	close(rec.release)
	a.Stop()

	if got := a.Pending(); got != 0 {
		t.Fatalf("Pending = %d after Stop, want 0", got)
	}
	if got := a.Total("p"); got != 4 {
		t.Fatalf("Total = %d, want 4", got)
	}
	if got := rec.calls.Load(); got != 4 {
		t.Fatalf("Recorder saw %d events, want 4", got)
	}
}

func TestZeroBufferDropsWhileWorkerBusy(t *testing.T) {
	rec := newGateRecorder()
	a := NewAnalytics(rec, 1, 0)

	deadline := time.Now().Add(2 * time.Second)
	accepted := false
	for time.Now().Before(deadline) {
		if a.Record(ClickEvent{Code: "z", At: time.Now()}) {
			accepted = true
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !accepted {
		t.Fatal("could not hand an event to the single worker")
	}
	select {
	case <-rec.started:
	case <-time.After(2 * time.Second):
		t.Fatal("worker never started")
	}

	before := a.Dropped()
	for i := 0; i < 5; i++ {
		if a.Record(ClickEvent{Code: "z", At: time.Now()}) {
			t.Fatalf("event %d was accepted into a full unbuffered queue", i)
		}
	}
	if got := a.Dropped() - before; got != 5 {
		t.Fatalf("dropped %d events, want 5", got)
	}

	close(rec.release)
	a.Stop()
	if got := a.Total("z"); got != 1 {
		t.Fatalf("Total = %d, want 1", got)
	}
	if got := rec.calls.Load(); got != 1 {
		t.Fatalf("Recorder saw %d events, want 1", got)
	}
	if got := a.Pending(); got != 0 {
		t.Fatalf("Pending = %d after Stop, want 0", got)
	}
}

func TestNegativeBufferFloor(t *testing.T) {
	rec := &countingRecorder{}
	a := NewAnalytics(rec, 2, -7)

	const n = 100
	accepted := 0
	for i := 0; i < n; i++ {
		if a.Record(ClickEvent{Code: "nb", At: time.Now()}) {
			accepted++
		}
	}
	a.Stop()

	if got := a.Total("nb"); got != accepted {
		t.Fatalf("Total = %d, want %d", got, accepted)
	}
	if got := rec.calls.Load(); got != int64(accepted) {
		t.Fatalf("Recorder saw %d events, want %d", got, accepted)
	}
	if got := a.Dropped(); got != int64(n-accepted) {
		t.Fatalf("Dropped = %d, want %d", got, n-accepted)
	}
}

func TestWorkerFloorForNonPositiveValues(t *testing.T) {
	for _, workers := range []int{0, -1, -10} {
		rec := &countingRecorder{}
		a := NewAnalytics(rec, workers, 64)
		for i := 0; i < 20; i++ {
			if !a.Record(ClickEvent{Code: "wf", At: time.Now()}) {
				t.Fatalf("workers=%d: event %d was dropped with an ample buffer", workers, i)
			}
		}
		a.Stop()

		if got := a.Total("wf"); got != 20 {
			t.Fatalf("workers=%d: Total = %d, want 20", workers, got)
		}
		if got := rec.calls.Load(); got != 20 {
			t.Fatalf("workers=%d: Recorder saw %d events, want 20", workers, got)
		}
	}
}

func TestTotalsArePerCode(t *testing.T) {
	rec := &countingRecorder{}
	a := NewAnalytics(rec, 4, 128)

	codes := []string{"a", "b", "c"}
	counts := []int{5, 3, 7}
	total := 0
	for i, code := range codes {
		for j := 0; j < counts[i]; j++ {
			if !a.Record(ClickEvent{Code: code, At: time.Now()}) {
				t.Fatalf("event %d for %q was dropped", j, code)
			}
			total++
		}
	}
	a.Stop()

	for i, code := range codes {
		if got := a.Total(code); got != counts[i] {
			t.Fatalf("Total(%q) = %d, want %d", code, got, counts[i])
		}
	}
	if got := a.Total("missing"); got != 0 {
		t.Fatalf("Total(missing) = %d, want 0", got)
	}
	if got := rec.calls.Load(); got != int64(total) {
		t.Fatalf("Recorder saw %d events, want %d", got, total)
	}
}

func TestConcurrentPerCodeTotals(t *testing.T) {
	rec := &countingRecorder{}
	a := NewAnalytics(rec, 8, 4096)

	const goroutines = 10
	const each = 50

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			code := string(rune('a' + g))
			for i := 0; i < each; i++ {
				if !a.Record(ClickEvent{Code: code, At: time.Now()}) {
					t.Errorf("event %d for %q was dropped", i, code)
					return
				}
			}
		}(g)
	}
	wg.Wait()
	a.Stop()

	for g := 0; g < goroutines; g++ {
		code := string(rune('a' + g))
		if got := a.Total(code); got != each {
			t.Errorf("Total(%q) = %d, want %d", code, got, each)
		}
	}
	if got := rec.calls.Load(); got != goroutines*each {
		t.Fatalf("Recorder saw %d events, want %d", got, goroutines*each)
	}
}

func TestStopDrainsBufferedEvents(t *testing.T) {
	rec := &countingRecorder{}
	a := NewAnalytics(rec, 1, 8)

	for i := 0; i < 8; i++ {
		if !a.Record(ClickEvent{Code: "drain", At: time.Now()}) {
			t.Fatalf("event %d was dropped", i)
		}
	}
	a.Stop()

	if got := a.Total("drain"); got != 8 {
		t.Fatalf("Total = %d, want 8", got)
	}
	if got := rec.calls.Load(); got != 8 {
		t.Fatalf("Recorder saw %d events, want 8", got)
	}
	if got := a.Dropped(); got != 0 {
		t.Fatalf("Dropped = %d, want 0", got)
	}
	if got := a.Pending(); got != 0 {
		t.Fatalf("Pending = %d after Stop, want 0", got)
	}
}

func TestConcurrentStopCallsAreSafe(t *testing.T) {
	rec := &countingRecorder{}
	a := NewAnalytics(rec, 4, 32)
	for i := 0; i < 20; i++ {
		if !a.Record(ClickEvent{Code: "cs", At: time.Now()}) {
			t.Fatalf("event %d was dropped", i)
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.Stop()
		}()
	}
	wg.Wait()

	if got := a.Total("cs"); got != 20 {
		t.Fatalf("Total = %d, want 20", got)
	}
	if got := rec.calls.Load(); got != 20 {
		t.Fatalf("Recorder saw %d events, want 20", got)
	}
}

func TestRecordAfterStopManyGoroutines(t *testing.T) {
	a := NewAnalytics(&countingRecorder{}, 2, 8)
	a.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if a.Record(ClickEvent{Code: "late", At: time.Now()}) {
				t.Errorf("goroutine %d: Record accepted an event after Stop", i)
			}
		}(i)
	}
	wg.Wait()

	if got := a.Total("late"); got != 0 {
		t.Fatalf("Total = %d, want 0", got)
	}
}
`,
				}},
				Hints: []string{
					"Fields you need: §events chan ClickEvent§, §rec Recorder§, §wg sync.WaitGroup§, §stopped atomic.Bool§, §dropped atomic.Int64§, and §counts map[string]*atomic.Int64§ guarded by a §sync.RWMutex§.",
					"§Record§: check §stopped.Load()§, then §select { case a.events <- e: return true; default: a.dropped.Add(1); return false }§.",
					"Worker loop: §for e := range a.events { a.rec.Record(context.Background(), e.Code); a.bump(e.Code) }§ with §defer a.wg.Done()§.",
					"§Stop§: §if !a.stopped.CompareAndSwap(false, true) { return }§, then §close(a.events)§, then §a.wg.Wait()§.",
					"§Pending§ = §len(a.events)§.",
					"Make §Record§ safe after §Stop§ — a §send on a closed channel§ panics, so the §stopped§ flag must be checked *before* the select, and §Stop§ must set it before closing.",
				},
				Solution: `package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type ClickEvent struct {
	Code string
	At   time.Time
}

type Recorder interface {
	Record(ctx context.Context, code string) error
}

type Analytics struct {
	events  chan ClickEvent
	rec     Recorder
	wg      sync.WaitGroup
	stopped atomic.Bool
	dropped atomic.Int64

	mu     sync.RWMutex
	counts map[string]*atomic.Int64
}

func NewAnalytics(rec Recorder, workers, buffer int) *Analytics {
	if workers < 1 {
		workers = 1
	}
	if buffer < 0 {
		buffer = 0
	}
	a := &Analytics{
		events: make(chan ClickEvent, buffer),
		rec:    rec,
		counts: make(map[string]*atomic.Int64),
	}
	for i := 0; i < workers; i++ {
		a.wg.Add(1)
		go a.worker()
	}
	return a
}

func (a *Analytics) worker() {
	defer a.wg.Done()
	for e := range a.events {
		if a.rec != nil {
			_ = a.rec.Record(context.Background(), e.Code)
		}
		a.counter(e.Code).Add(1)
	}
}

func (a *Analytics) counter(code string) *atomic.Int64 {
	a.mu.RLock()
	c, ok := a.counts[code]
	a.mu.RUnlock()
	if ok {
		return c
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if c, ok := a.counts[code]; ok {
		return c
	}
	c = &atomic.Int64{}
	a.counts[code] = c
	return c
}

func (a *Analytics) Record(e ClickEvent) bool {
	if a.stopped.Load() {
		return false
	}
	select {
	case a.events <- e:
		return true
	default:
		a.dropped.Add(1)
		return false
	}
}

func (a *Analytics) Total(code string) int {
	a.mu.RLock()
	c, ok := a.counts[code]
	a.mu.RUnlock()
	if !ok {
		return 0
	}
	return int(c.Load())
}

func (a *Analytics) Dropped() int64 { return a.dropped.Load() }

func (a *Analytics) Pending() int { return len(a.events) }

func (a *Analytics) Stop() {
	if !a.stopped.CompareAndSwap(false, true) {
		return
	}
	close(a.events)
	a.wg.Wait()
}

func main() {
	fmt.Println(NewAnalytics(nil, 2, 8) != nil)
}
`,
			},
		},
	}
}
