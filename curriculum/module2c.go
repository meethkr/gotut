package curriculum

func m2t10() Topic {
	return Topic{
		ID:      "m2t10",
		Title:   "Databases with database/sql",
		Summary: "Connecting, querying, and never building SQL by string concatenation.",
		Theory: `
§database/sql§ is Go's **generic** database interface. It does not know about Postgres
or SQLite; a **driver** package teaches it. You import the driver for its side effects
only (a blank import) and then talk to the standard API.

§§§go
import (
	"database/sql"
	_ "modernc.org/sqlite" // registers the driver named "sqlite" (pure Go, no cgo)
	// _ "github.com/jackc/pgx/v5/stdlib" // "pgx" for PostgreSQL
	// _ "github.com/go-sql-driver/mysql" // "mysql"
)

db, err := sql.Open("sqlite", "file:app.db")
if err != nil {
	return err
}
defer db.Close()
§§§

## sql.Open does not connect

This catches people out once:

* §sql.Open§ validates the driver name and the DSN, and creates a **pool object**.
* It does **not** open a connection.

So §err§ being nil tells you nothing about whether the database exists. To find out,
§Ping§:

§§§go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := db.PingContext(ctx); err != nil {
	return fmt.Errorf("database unreachable: %w", err)
}
§§§

## *sql.DB is a pool, not a connection

It is safe for concurrent use and should be created **once** and shared. Four settings
matter:

§§§go
db.SetMaxOpenConns(25)                  // never exceed what the DB can take
db.SetMaxIdleConns(25)                  // keep some warm
db.SetConnMaxLifetime(5 * time.Minute)  // recycle, so restarts and DNS changes are seen
db.SetConnMaxIdleTime(2 * time.Minute)  // close truly idle connections
§§§

For SQLite specifically, keep §SetMaxOpenConns(1)§: SQLite serialises writers anyway, and
a single connection avoids "database is locked" surprises.

## Repositories: keep SQL in one place

A common structure is a type that owns the §*sql.DB§ and exposes methods:

§§§go
type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }
§§§

Everything SQL-ish lives in these methods. Handlers call §repo.Get(ctx, id)§ and never
see a query string — which makes the handler testable with a fake and the SQL changeable
without touching HTTP.

## Querying one row

§§§go
var name string
err := db.QueryRowContext(ctx,
	"SELECT name FROM users WHERE id = ?", id).Scan(&name)
if errors.Is(err, sql.ErrNoRows) {
	return ErrNotFound // your own sentinel from Module 1
}
if err != nil {
	return fmt.Errorf("get user %d: %w", id, err)
}
§§§

§QueryRow§ returns a §*sql.Row§ that holds the error until you call §Scan§. That is why
the error check happens *after* §Scan§ — a very common source of confusion.

## Querying many rows

§§§go
rows, err := db.QueryContext(ctx, "SELECT id, name FROM users WHERE active = ?", true)
if err != nil {
	return nil, err
}
defer rows.Close() // ALWAYS

var users []User
for rows.Next() {
	var u User
	if err := rows.Scan(&u.ID, &u.Name); err != nil {
		return nil, err
	}
	users = append(users, u)
}
return users, rows.Err() // check this too
§§§

Three things that bite:

1. **§defer rows.Close()§** — without it you hold a connection from the pool until the
   process ends. Under load the service hangs.
2. **§rows.Err()§** — §Next§ returns false both at the end *and* on a mid-iteration
   failure. §rows.Err()§ tells you which. Returning §nil§ without checking it hides
   errors.
3. **Initialise the slice** (§var users []User§ gives §nil§) if it is going into JSON —
   otherwise an empty result marshals to §null§ instead of §[]§.

## Parameter placeholders: never concatenate

§§§go
// NEVER DO THIS — SQL injection
q := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name)

// DO THIS — the driver sends the value out of band
db.QueryContext(ctx, "SELECT * FROM users WHERE name = ?", name)
§§§

§?§ is the placeholder for SQLite and MySQL. PostgreSQL (§pgx§, §lib/pq§) uses §$1§,
§$2§. The placeholder style is the *only* thing that changes between drivers, which is
why portable SQL uses §?§ and §sqlx.Rebind§ converts.

Parameters can never change the shape of a query — that is the entire point. Identifiers
(table names, column names, §ORDER BY§ targets) cannot be parameterised, so validate them
against a fixed allow-list instead.

## Writing data

§§§go
res, err := db.ExecContext(ctx,
	"INSERT INTO users (name, email) VALUES (?, ?)", u.Name, u.Email)
if err != nil {
	return err
}

id, err := res.LastInsertId() // SQLite and MySQL
n, err := res.RowsAffected()  // useful for UPDATE and DELETE
§§§

§RowsAffected§ is how you turn "no rows matched" into §ErrNotFound§ without a second
query:

§§§go
res, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
if err != nil {
	return err
}
n, _ := res.RowsAffected()
if n == 0 {
	return ErrNotFound
}
§§§

## NULL values

A SQL §NULL§ cannot scan into an §int§ or a §string§. Three ways out:

§§§go
// 1. The sql.Null* wrappers
var name sql.NullString
rows.Scan(&name)
if name.Valid {
	fmt.Println(name.String)
}

// 2. A pointer
var email *string
rows.Scan(&email) // nil when the column is NULL

// 3. COALESCE, so the database supplies a default
db.QueryRow("SELECT COALESCE(bio, '') FROM users WHERE id = ?", id).Scan(&bio)
§§§

## Transactions

§§§go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
	return err
}
defer tx.Rollback() // a no-op after a successful Commit

if _, err := tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, from); err != nil {
	return err
}
if _, err := tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, to); err != nil {
	return err
}
return tx.Commit()
§§§

The §defer tx.Rollback()§ idiom guarantees the transaction is released on every error
path. After a successful §Commit§ it returns §sql.ErrTxDone§ and is harmless.

Two rules of thumb: keep transactions **short** (they hold locks), and always touch rows
in the **same order** everywhere to avoid deadlocks.

## A complete repository

§§§go
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("not found")

type Note struct {
	ID    int64
	Title string
}

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

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS notes (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL)")
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func CreateNote(ctx context.Context, db *sql.DB, title string) (int64, error) {
	res, err := db.ExecContext(ctx, "INSERT INTO notes (title) VALUES (?)", title)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetNote(ctx context.Context, db *sql.DB, id int64) (Note, error) {
	var n Note
	err := db.QueryRowContext(ctx, "SELECT id, title FROM notes WHERE id = ?", id).Scan(&n.ID, &n.Title)
	if errors.Is(err, sql.ErrNoRows) {
		return Note{}, ErrNotFound
	}
	if err != nil {
		return Note{}, fmt.Errorf("get note %d: %w", id, err)
	}
	return n, nil
}
§§§

You will write those three functions, in some form, in every Go service you ever build.
`,
		Quiz: []Question{
			{
				ID:     "m2t10q1",
				Prompt: "What does §sql.Open§ actually do?",
				Options: []Option{
					{Text: "Validates the driver name and DSN and creates a connection pool; it does not connect yet.", Correct: true, Rationale: "The first query, or an explicit §Ping§, establishes a real connection."},
					{Text: "Opens one connection and keeps it for the process lifetime.", Correct: false, Rationale: "§*sql.DB§ is a pool, not a connection."},
					{Text: "Runs the schema migrations.", Correct: false, Rationale: "It knows nothing about your schema."},
					{Text: "Nothing at all.", Correct: false, Rationale: "It does validate the driver registration and DSN."},
				},
				Explanation: "This lazy behaviour is why a nil error from §sql.Open§ does not mean the database is reachable.",
			},
			{
				ID:     "m2t10q2",
				Prompt: "Which are safe against SQL injection? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "§db.Query(\"SELECT * FROM users WHERE id = ?\", id)§", Correct: true, Rationale: "The value is sent as a parameter, separate from the query text."},
					{Text: "§db.Query(fmt.Sprintf(\"SELECT * FROM users WHERE name = '%s'\", name))§", Correct: false, Rationale: "Input can change the query structure. This is the classic injection."},
					{Text: "§db.Query(\"SELECT * FROM users WHERE name = ?\", name)§", Correct: true, Rationale: "Same as the first — parameters are the fix."},
					{Text: "§db.Query(\"SELECT * FROM users ORDER BY \" + sortField)§", Correct: false, Rationale: "Identifiers cannot be parameterised, so this must be validated against an allow-list instead."},
				},
				Explanation: "Bind *values* with placeholders; validate *identifiers* against a fixed set.",
			},
			{
				ID:     "m2t10q3",
				Prompt: "Why is §defer rows.Close()§ so important right after §db.Query§?",
				Options: []Option{
					{Text: "Because the rows hold a connection from the pool; leaving them open exhausts the pool and the service stops responding.", Correct: true, Rationale: "It is a failure that appears only under load, which makes it especially nasty."},
					{Text: "Because §Next§ panics if rows are not closed.", Correct: false, Rationale: "It does not."},
					{Text: "Because the query is not actually executed until Close.", Correct: false, Rationale: "The query runs immediately."},
					{Text: "It is optional when you consume every row.", Correct: false, Rationale: "Still required — and §rows.Err()§ still needs checking."},
				},
				Explanation: "§defer rows.Close()§ plus checking §rows.Err()§ after the loop is the complete pattern.",
			},
			{
				ID:     "m2t10q4",
				Prompt: "A column is SQL §NULL§ and you scan it into a §string§. What happens?",
				Options: []Option{
					{Text: "§Scan§ returns an error like \"converting NULL to string is unsupported\".", Correct: true, Rationale: "Use §sql.NullString§, a §*string§, or §COALESCE§ in the query."},
					{Text: "The string becomes empty.", Correct: false, Rationale: "It errors rather than coercing."},
					{Text: "The string becomes the literal text §NULL§.", Correct: false, Rationale: "No."},
					{Text: "It panics.", Correct: false, Rationale: "§Scan§ returns an error value."},
				},
				Explanation: "Nullability is part of the schema, so it must be handled explicitly in Go.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m2t10e1",
				Title: "Warm-up: open, migrate, then CRUD",
				Prompt: `Build a repository over SQLite using the pure-Go §modernc.org/sqlite§ driver.

§§§go
var ErrNoteNotFound = errors.New("note not found")

type Note struct {
	ID    int64
	Title string
	Body  string
}

func Open(dsn string) (*sql.DB, error)   // open, ping, migrate
func CreateNote(db *sql.DB, n Note) (int64, error)
func GetNote(db *sql.DB, id int64) (Note, error)
func ListNotes(db *sql.DB) ([]Note, error)
func DeleteNote(db *sql.DB, id int64) error
func SearchNotes(db *sql.DB, term string) ([]Note, error)
§§§

Schema:

§§§sql
notes(id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, body TEXT NOT NULL)
§§§

Rules:

* §Open§ sets sane pool settings, pings, and creates the table with §CREATE TABLE IF
  NOT EXISTS§.
* §GetNote§ maps §sql.ErrNoRows§ to §ErrNoteNotFound§ using §errors.Is§.
* §ListNotes§ returns an **empty non-nil** slice when there is nothing. That means
  §notes := []Note{}§ rather than §var notes []Note§. Every function takes the §*sql.DB§
  as a parameter, so the test can hand each test its own temporary database.
* §SearchNotes§ matches §title LIKE %term%§. The parameter is the *value* §"%"+term+"%"§
  — never concatenated into the SQL text.
* §DeleteNote§ deletes by id and returns §ErrNoteNotFound§ when no row matched — use the
  result's §RowsAffected()§ to tell.
* Do not forget §defer rows.Close()§ and the final §rows.Err()§.
`,
				Starter: `package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

var ErrNoteNotFound = errors.New("note not found")

type Note struct {
	ID    int64
	Title string
	Body  string
}

// Open opens the database, verifies it, and creates the schema.
func Open(dsn string) (*sql.DB, error) {
	// TODO
	return nil, nil
}

func CreateNote(db *sql.DB, n Note) (int64, error) {
	// TODO
	return 0, nil
}

func GetNote(db *sql.DB, id int64) (Note, error) {
	// TODO
	return Note{}, nil
}

func ListNotes(db *sql.DB) ([]Note, error) {
	// TODO
	return nil, nil
}

func DeleteNote(db *sql.DB, id int64) error {
	// TODO
	return nil
}

func SearchNotes(db *sql.DB, term string) ([]Note, error) {
	// TODO
	return nil, nil
}

func main() {
	db, err := Open("file:demo.db?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	id, _ := CreateNote(db, Note{Title: "hello", Body: "world"})
	n, _ := GetNote(db, id)
	fmt.Println(n, strings.ToUpper("ok"))
}
`,
				Deps: []string{"modernc.org/sqlite"},
				Tests: []File{{
					Name: "repo_test.go",
					Content: `package main

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func newDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open("file:" + filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCreateAndGet(t *testing.T) {
	db := newDB(t)

	id, err := CreateNote(db, Note{Title: "first", Body: "body"})
	if err != nil {
		t.Fatalf("CreateNote: %v", err)
	}
	if id == 0 {
		t.Fatal("CreateNote returned id 0")
	}

	n, err := GetNote(db, id)
	if err != nil {
		t.Fatalf("GetNote: %v", err)
	}
	if n.ID != id || n.Title != "first" || n.Body != "body" {
		t.Fatalf("GetNote = %+v", n)
	}
}

func TestGetMissing(t *testing.T) {
	db := newDB(t)
	if _, err := GetNote(db, 999); !errors.Is(err, ErrNoteNotFound) {
		t.Fatalf("GetNote(missing) = %v, want ErrNoteNotFound", err)
	}
}

func TestGetMissingBoundaryIDs(t *testing.T) {
	db := newDB(t)
	for _, id := range []int64{0, -1, -999, 999} {
		if _, err := GetNote(db, id); !errors.Is(err, ErrNoteNotFound) {
			t.Errorf("GetNote(%d) = %v, want ErrNoteNotFound", id, err)
		}
	}
}

func TestCreateAssignsDistinctIDs(t *testing.T) {
	db := newDB(t)

	first, err := CreateNote(db, Note{Title: "one", Body: "1"})
	if err != nil {
		t.Fatalf("CreateNote first: %v", err)
	}
	second, err := CreateNote(db, Note{Title: "two", Body: "2"})
	if err != nil {
		t.Fatalf("CreateNote second: %v", err)
	}
	if first == second {
		t.Fatalf("both notes got id %d; ids must be distinct", first)
	}

	n, err := GetNote(db, second)
	if err != nil {
		t.Fatalf("GetNote: %v", err)
	}
	if n.ID != second || n.Title != "two" || n.Body != "2" {
		t.Fatalf("GetNote = %+v, want the second note", n)
	}
}

func TestCreateEmptyStrings(t *testing.T) {
	db := newDB(t)

	id, err := CreateNote(db, Note{})
	if err != nil {
		t.Fatalf("CreateNote(empty): %v", err)
	}
	n, err := GetNote(db, id)
	if err != nil {
		t.Fatalf("GetNote: %v", err)
	}
	if n.Title != "" || n.Body != "" {
		t.Fatalf("GetNote = %+v, want empty title and body", n)
	}
}

func TestListOrdered(t *testing.T) {
	db := newDB(t)
	for _, title := range []string{"a", "b", "c"} {
		if _, err := CreateNote(db, Note{Title: title, Body: "x"}); err != nil {
			t.Fatal(err)
		}
	}

	notes, err := ListNotes(db)
	if err != nil {
		t.Fatalf("ListNotes: %v", err)
	}
	if len(notes) != 3 {
		t.Fatalf("len = %d, want 3", len(notes))
	}
	for i, want := range []string{"a", "b", "c"} {
		if notes[i].Title != want {
			t.Fatalf("notes[%d] = %q, want %q", i, notes[i].Title, want)
		}
	}
}

func TestListEmptyIsNonNil(t *testing.T) {
	db := newDB(t)
	notes, err := ListNotes(db)
	if err != nil {
		t.Fatalf("ListNotes: %v", err)
	}
	if notes == nil {
		t.Fatal("ListNotes returned nil; return an empty slice so JSON gives [] not null")
	}
	if len(notes) != 0 {
		t.Fatalf("len = %d, want 0", len(notes))
	}
}

func TestListAfterDeleteKeepsOrder(t *testing.T) {
	db := newDB(t)

	var ids []int64
	for _, title := range []string{"a", "b", "c", "d"} {
		id, err := CreateNote(db, Note{Title: title, Body: "x"})
		if err != nil {
			t.Fatalf("CreateNote(%q): %v", title, err)
		}
		ids = append(ids, id)
	}
	if err := DeleteNote(db, ids[1]); err != nil {
		t.Fatalf("DeleteNote: %v", err)
	}
	if err := DeleteNote(db, ids[3]); err != nil {
		t.Fatalf("DeleteNote: %v", err)
	}

	notes, err := ListNotes(db)
	if err != nil {
		t.Fatalf("ListNotes: %v", err)
	}
	if len(notes) != 2 {
		t.Fatalf("len = %d, want 2: %+v", len(notes), notes)
	}
	if notes[0].ID != ids[0] || notes[0].Title != "a" {
		t.Errorf("notes[0] = %+v, want id %d title a", notes[0], ids[0])
	}
	if notes[1].ID != ids[2] || notes[1].Title != "c" {
		t.Errorf("notes[1] = %+v, want id %d title c", notes[1], ids[2])
	}
}

func TestOpenIsIdempotentAndPersists(t *testing.T) {
	dsn := "file:" + filepath.Join(t.TempDir(), "persist.db")

	first, err := Open(dsn)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	id, err := CreateNote(first, Note{Title: "kept", Body: "body"})
	if err != nil {
		first.Close()
		t.Fatalf("CreateNote: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second, err := Open(dsn)
	if err != nil {
		t.Fatalf("second Open (CREATE TABLE IF NOT EXISTS should tolerate an existing table): %v", err)
	}
	defer second.Close()

	n, err := GetNote(second, id)
	if err != nil {
		t.Fatalf("GetNote after reopen: %v", err)
	}
	if n.Title != "kept" || n.Body != "body" {
		t.Fatalf("GetNote after reopen = %+v", n)
	}
	notes, err := ListNotes(second)
	if err != nil || len(notes) != 1 {
		t.Fatalf("ListNotes after reopen = %+v, %v; want one note", notes, err)
	}
}

func TestDelete(t *testing.T) {
	db := newDB(t)
	id, _ := CreateNote(db, Note{Title: "gone", Body: "x"})

	if err := DeleteNote(db, id); err != nil {
		t.Fatalf("DeleteNote: %v", err)
	}
	if err := DeleteNote(db, id); !errors.Is(err, ErrNoteNotFound) {
		t.Fatalf("second DeleteNote = %v, want ErrNoteNotFound", err)
	}
}

func TestDeleteMissingIDs(t *testing.T) {
	db := newDB(t)
	kept, err := CreateNote(db, Note{Title: "kept", Body: "x"})
	if err != nil {
		t.Fatalf("CreateNote: %v", err)
	}

	for _, id := range []int64{0, -1, 999} {
		if err := DeleteNote(db, id); !errors.Is(err, ErrNoteNotFound) {
			t.Errorf("DeleteNote(%d) = %v, want ErrNoteNotFound", id, err)
		}
	}
	if _, err := GetNote(db, kept); err != nil {
		t.Fatalf("a failed delete removed another note: %v", err)
	}
}

func TestSearch(t *testing.T) {
	db := newDB(t)
	_, _ = CreateNote(db, Note{Title: "golang tips", Body: "x"})
	_, _ = CreateNote(db, Note{Title: "rust tips", Body: "y"})
	_, _ = CreateNote(db, Note{Title: "cooking", Body: "z"})

	notes, err := SearchNotes(db, "tips")
	if err != nil {
		t.Fatalf("SearchNotes: %v", err)
	}
	if len(notes) != 2 {
		t.Fatalf("len = %d, want 2", len(notes))
	}
	if _, err := SearchNotes(db, "%"); err != nil {
		t.Fatalf("SearchNotes with a wildcard character should still be safe: %v", err)
	}
}

func TestSearchMultipleMatchesAndNoMatch(t *testing.T) {
	db := newDB(t)
	for _, title := range []string{"tips one", "other", "tips two"} {
		if _, err := CreateNote(db, Note{Title: title, Body: "x"}); err != nil {
			t.Fatalf("CreateNote: %v", err)
		}
	}

	notes, err := SearchNotes(db, "tips")
	if err != nil {
		t.Fatalf("SearchNotes: %v", err)
	}
	if len(notes) != 2 {
		t.Fatalf("len = %d, want 2: %+v", len(notes), notes)
	}
	got := map[string]bool{}
	for _, n := range notes {
		got[n.Title] = true
	}
	if !got["tips one"] || !got["tips two"] {
		t.Fatalf("SearchNotes = %+v, want tips one and tips two", notes)
	}

	none, err := SearchNotes(db, "no-such-term")
	if err != nil {
		t.Fatalf("SearchNotes(no match): %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("SearchNotes(no match) = %+v, want no rows", none)
	}
}

func TestSearchOnlyMatchesTitle(t *testing.T) {
	db := newDB(t)
	_, _ = CreateNote(db, Note{Title: "alpha", Body: "secret keyword"})

	notes, err := SearchNotes(db, "keyword")
	if err != nil {
		t.Fatalf("SearchNotes: %v", err)
	}
	if len(notes) != 0 {
		t.Fatalf("SearchNotes matched the body: %+v", notes)
	}
}

func TestSearchHostileInputDoesNotBreakSchema(t *testing.T) {
	db := newDB(t)
	if _, err := CreateNote(db, Note{Title: "survivor", Body: "x"}); err != nil {
		t.Fatalf("CreateNote: %v", err)
	}

	notes, err := SearchNotes(db, "'; DROP TABLE notes; --")
	if err != nil {
		t.Fatalf("SearchNotes(hostile term) = %v; the term must be a parameter, not SQL", err)
	}
	if len(notes) != 0 {
		t.Fatalf("SearchNotes(hostile term) = %+v, want no rows", notes)
	}

	all, err := ListNotes(db)
	if err != nil {
		t.Fatalf("ListNotes after a hostile search: %v", err)
	}
	if len(all) != 1 || all[0].Title != "survivor" {
		t.Fatalf("the notes table was damaged: %+v", all)
	}
}
`,
				}},
				Hints: []string{
					"§sql.Open(\"sqlite\", dsn)§, then §db.SetMaxOpenConns(1)§, then §db.PingContext(ctx)§, then the §CREATE TABLE§.",
					"§CreateNote§: §res, err := db.Exec(§INSERT INTO notes (title, body) VALUES (?, ?)§, n.Title, n.Body)§ then §res.LastInsertId()§.",
					"§GetNote§: §db.QueryRow(§SELECT id, title, body FROM notes WHERE id = ?§, id).Scan(&n.ID, &n.Title, &n.Body)§, and map §sql.ErrNoRows§.",
					"§ListNotes§: §notes := []Note{}§ (not §var notes []Note§), then the §for rows.Next()§ loop and §return notes, rows.Err()§.",
					"§DeleteNote§ uses §RowsAffected()§: if it is 0, nothing matched, so return §ErrNoteNotFound§.",
					"§SearchNotes§: §WHERE title LIKE ?§ with the *parameter* §\"%\"+term+\"%\"§.",
				},
				Solution: `package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var ErrNoteNotFound = errors.New("note not found")

type Note struct {
	ID    int64
	Title string
	Body  string
}

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

	_, err = db.ExecContext(ctx, §CREATE TABLE IF NOT EXISTS notes (
		id    INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		body  TEXT NOT NULL
	)§)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func CreateNote(db *sql.DB, n Note) (int64, error) {
	res, err := db.Exec(§INSERT INTO notes (title, body) VALUES (?, ?)§, n.Title, n.Body)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetNote(db *sql.DB, id int64) (Note, error) {
	var n Note
	err := db.QueryRow(§SELECT id, title, body FROM notes WHERE id = ?§, id).
		Scan(&n.ID, &n.Title, &n.Body)
	if errors.Is(err, sql.ErrNoRows) {
		return Note{}, ErrNoteNotFound
	}
	if err != nil {
		return Note{}, err
	}
	return n, nil
}

func ListNotes(db *sql.DB) ([]Note, error) {
	rows, err := db.Query(§SELECT id, title, body FROM notes ORDER BY id§)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := []Note{}
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Body); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func DeleteNote(db *sql.DB, id int64) error {
	res, err := db.Exec(§DELETE FROM notes WHERE id = ?§, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNoteNotFound
	}
	return nil
}

func SearchNotes(db *sql.DB, term string) ([]Note, error) {
	rows, err := db.Query(§SELECT id, title, body FROM notes WHERE title LIKE ? ORDER BY id§, "%"+term+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := []Note{}
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Body); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func main() {
	db, err := Open("file:demo.db?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	id, _ := CreateNote(db, Note{Title: "hello", Body: "world"})
	n, _ := GetNote(db, id)
	fmt.Println(n, strings.ToUpper("ok"))
}
`,
			},
			{
				ID:    "m2t10e2",
				Title: "Move money in a transaction",
				Prompt: `Transactions are where correctness gets real: a half-finished transfer is worse than
a failed one.

§§§go
type Account struct {
	ID      int
	Owner   string
	Balance int
}

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrAccountNotFound   = errors.New("account not found")
	ErrBadAmount         = errors.New("amount must be positive")
)

func Open(db *sql.DB) error                          // create the schema
func CreateAccount(db *sql.DB, owner string, balance int) (int, error)
func GetAccount(db *sql.DB, id int) (Account, error) // ErrAccountNotFound
func Transfer(db *sql.DB, from, to, amount int) error
func TotalBalance(db *sql.DB) (int, error)           // SUM over every account
§§§

§Transfer§ must:

* run inside a transaction (§db.BeginTx§, §defer tx.Rollback()§, §tx.Commit()§),
* reject §amount <= 0§ with §ErrBadAmount§,
* return §ErrAccountNotFound§ if either account is missing, changing nothing,
* return §ErrInsufficientFunds§ and change nothing if the source is short,
* leave the **total** balance across all accounts unchanged on success.

Read both rows *inside* the transaction, check the balance, then update — that order is
what makes the failure atomic.

§TotalBalance§ sums the balances of every account. SQL's §SUM§ over an empty table is
§NULL§, which will not scan into an §int§ — use §COALESCE(SUM(balance), 0)§ so an empty
table gives §0§.
`,
				Starter: `package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

type Account struct {
	ID      int
	Owner   string
	Balance int
}

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrAccountNotFound   = errors.New("account not found")
	ErrBadAmount         = errors.New("amount must be positive")
)

// Open creates the schema on an already-opened *sql.DB.
func Open(db *sql.DB) error {
	// TODO
	return nil
}

func CreateAccount(db *sql.DB, owner string, balance int) (int, error) {
	// TODO
	return 0, nil
}

func GetAccount(db *sql.DB, id int) (Account, error) {
	// TODO
	return Account{}, nil
}

func Transfer(db *sql.DB, from, to, amount int) error {
	// TODO
	return nil
}

func TotalBalance(db *sql.DB) (int, error) {
	// TODO
	return 0, nil
}

func main() {
	db, err := sql.Open("sqlite", "file:bank.db?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	_ = Open(db)
	a, _ := CreateAccount(db, "ada", 100)
	b, _ := CreateAccount(db, "bo", 0)
	fmt.Println(Transfer(db, a, b, 30))
	fmt.Println(TotalBalance(db))
}
`,
				Deps: []string{"modernc.org/sqlite"},
				Tests: []File{{
					Name: "transfer_test.go",
					Content: `package main

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func newBank(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "bank.db"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	if err := Open(db); err != nil {
		t.Fatalf("Open: %v", err)
	}
	return db
}

func TestCreateAndGet(t *testing.T) {
	db := newBank(t)

	id, err := CreateAccount(db, "ada", 100)
	if err != nil || id == 0 {
		t.Fatalf("CreateAccount = %d, %v", id, err)
	}
	a, err := GetAccount(db, id)
	if err != nil {
		t.Fatalf("GetAccount: %v", err)
	}
	if a.Owner != "ada" || a.Balance != 100 {
		t.Fatalf("account = %+v", a)
	}
	if _, err := GetAccount(db, 999); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("missing account = %v, want ErrAccountNotFound", err)
	}
}

func TestTransferSuccess(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 100)
	b, _ := CreateAccount(db, "bo", 20)

	before, _ := TotalBalance(db)
	if err := Transfer(db, a, b, 30); err != nil {
		t.Fatalf("Transfer: %v", err)
	}

	gotA, _ := GetAccount(db, a)
	gotB, _ := GetAccount(db, b)
	if gotA.Balance != 70 || gotB.Balance != 50 {
		t.Fatalf("balances = %d, %d; want 70, 50", gotA.Balance, gotB.Balance)
	}
	after, _ := TotalBalance(db)
	if before != after {
		t.Fatalf("the total changed: %d -> %d", before, after)
	}
}

func TestTransferInsufficientFunds(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 10)
	b, _ := CreateAccount(db, "bo", 0)

	if err := Transfer(db, a, b, 999); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("Transfer = %v, want ErrInsufficientFunds", err)
	}
	gotA, _ := GetAccount(db, a)
	gotB, _ := GetAccount(db, b)
	if gotA.Balance != 10 || gotB.Balance != 0 {
		t.Fatalf("a failed transfer changed something: %d, %d", gotA.Balance, gotB.Balance)
	}
}

func TestTransferMissingAccount(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 100)

	if err := Transfer(db, a, 999, 10); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("Transfer to a missing account = %v, want ErrAccountNotFound", err)
	}
	gotA, _ := GetAccount(db, a)
	if gotA.Balance != 100 {
		t.Fatalf("the balance changed on a failed transfer: %d", gotA.Balance)
	}
}

func TestTransferBadAmount(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 100)
	b, _ := CreateAccount(db, "bo", 0)

	for _, amount := range []int{0, -5} {
		if err := Transfer(db, a, b, amount); err == nil {
			t.Fatalf("Transfer(amount=%d) should fail", amount)
		}
	}
}

func TestGetAccountMissingBoundaryIDs(t *testing.T) {
	db := newBank(t)
	for _, id := range []int{0, -1, 999} {
		if _, err := GetAccount(db, id); !errors.Is(err, ErrAccountNotFound) {
			t.Errorf("GetAccount(%d) = %v, want ErrAccountNotFound", id, err)
		}
	}
}

func TestTotalBalanceEmptyAndMultiple(t *testing.T) {
	db := newBank(t)

	total, err := TotalBalance(db)
	if err != nil {
		t.Fatalf("TotalBalance(empty): %v", err)
	}
	if total != 0 {
		t.Fatalf("TotalBalance(empty) = %d, want 0", total)
	}

	for _, in := range []struct {
		owner   string
		balance int
	}{{"ada", 100}, {"bo", 20}, {"cy", 5}} {
		if _, err := CreateAccount(db, in.owner, in.balance); err != nil {
			t.Fatalf("CreateAccount(%s): %v", in.owner, err)
		}
	}
	total, err = TotalBalance(db)
	if err != nil {
		t.Fatalf("TotalBalance: %v", err)
	}
	if total != 125 {
		t.Fatalf("TotalBalance = %d, want 125", total)
	}
}

func TestTransferExactBalanceAndOneOver(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 50)
	b, _ := CreateAccount(db, "bo", 10)

	if err := Transfer(db, a, b, 51); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("Transfer(51 from 50) = %v, want ErrInsufficientFunds", err)
	}
	gotA, _ := GetAccount(db, a)
	gotB, _ := GetAccount(db, b)
	if gotA.Balance != 50 || gotB.Balance != 10 {
		t.Fatalf("the failed boundary transfer changed balances: %d, %d", gotA.Balance, gotB.Balance)
	}

	if err := Transfer(db, a, b, 50); err != nil {
		t.Fatalf("Transfer(exact balance): %v", err)
	}
	gotA, _ = GetAccount(db, a)
	gotB, _ = GetAccount(db, b)
	if gotA.Balance != 0 || gotB.Balance != 60 {
		t.Fatalf("balances = %d, %d; want 0, 60", gotA.Balance, gotB.Balance)
	}
	if total, err := TotalBalance(db); err != nil || total != 60 {
		t.Fatalf("TotalBalance = %d, %v; want 60, nil", total, err)
	}
}

func TestTransferMissingSource(t *testing.T) {
	db := newBank(t)
	b, _ := CreateAccount(db, "bo", 15)

	if err := Transfer(db, 999, b, 5); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("Transfer from a missing account = %v, want ErrAccountNotFound", err)
	}
	gotB, _ := GetAccount(db, b)
	if gotB.Balance != 15 {
		t.Fatalf("the destination changed on a failed transfer: %d", gotB.Balance)
	}
}

func TestTransferBadAmountIsTypedAndAtomic(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 100)
	b, _ := CreateAccount(db, "bo", 25)

	for _, amount := range []int{0, -1, -100} {
		if err := Transfer(db, a, b, amount); !errors.Is(err, ErrBadAmount) {
			t.Errorf("Transfer(amount=%d) = %v, want ErrBadAmount", amount, err)
		}
	}
	gotA, _ := GetAccount(db, a)
	gotB, _ := GetAccount(db, b)
	if gotA.Balance != 100 || gotB.Balance != 25 {
		t.Fatalf("a rejected transfer changed balances: %d, %d", gotA.Balance, gotB.Balance)
	}
	if total, err := TotalBalance(db); err != nil || total != 125 {
		t.Fatalf("TotalBalance = %d, %v; want 125, nil", total, err)
	}
}

func TestFailedTransfersPreserveTotal(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 40)
	b, _ := CreateAccount(db, "bo", 10)
	before, _ := TotalBalance(db)

	if err := Transfer(db, a, b, 41); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("Transfer = %v, want ErrInsufficientFunds", err)
	}
	if err := Transfer(db, a, 999, 5); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("Transfer = %v, want ErrAccountNotFound", err)
	}
	after, err := TotalBalance(db)
	if err != nil {
		t.Fatalf("TotalBalance: %v", err)
	}
	if after != before {
		t.Fatalf("failed transfers changed the total: %d -> %d", before, after)
	}
}

func TestSequentialTransfers(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 100)
	b, _ := CreateAccount(db, "bo", 0)

	if err := Transfer(db, a, b, 30); err != nil {
		t.Fatalf("first Transfer: %v", err)
	}
	if err := Transfer(db, b, a, 10); err != nil {
		t.Fatalf("second Transfer: %v", err)
	}
	gotA, _ := GetAccount(db, a)
	gotB, _ := GetAccount(db, b)
	if gotA.Balance != 80 || gotB.Balance != 20 {
		t.Fatalf("balances = %d, %d; want 80, 20", gotA.Balance, gotB.Balance)
	}
	if total, err := TotalBalance(db); err != nil || total != 100 {
		t.Fatalf("TotalBalance = %d, %v; want 100, nil", total, err)
	}
}
`,
				}},
				Hints: []string{
					"§Open(db)§ runs §CREATE TABLE IF NOT EXISTS accounts (id INTEGER PRIMARY KEY AUTOINCREMENT, owner TEXT NOT NULL, balance INTEGER NOT NULL)§.",
					"§GetAccount§: §QueryRow(...).Scan(&a.ID, &a.Owner, &a.Balance)§, mapping §sql.ErrNoRows§ to §ErrAccountNotFound§.",
					"§Transfer§: guard §amount <= 0§ first, then §tx, err := db.BeginTx(ctx, nil)§ and §defer tx.Rollback()§.",
					"Inside the transaction, read both rows with §tx.QueryRow§. If either is missing, return §ErrAccountNotFound§ (the deferred rollback undoes nothing, since nothing was changed).",
					"Then check §src.Balance < amount§, and only after that run the two §tx.Exec§ updates and §tx.Commit()§.",
					"§TotalBalance§: §SELECT COALESCE(SUM(balance), 0) FROM accounts§ — COALESCE so an empty table gives 0 rather than NULL.",
				},
				Solution: `package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

type Account struct {
	ID      int
	Owner   string
	Balance int
}

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrAccountNotFound   = errors.New("account not found")
	ErrBadAmount         = errors.New("amount must be positive")
)

func Open(db *sql.DB) error {
	_, err := db.Exec(§CREATE TABLE IF NOT EXISTS accounts (
		id      INTEGER PRIMARY KEY AUTOINCREMENT,
		owner   TEXT NOT NULL,
		balance INTEGER NOT NULL
	)§)
	return err
}

func CreateAccount(db *sql.DB, owner string, balance int) (int, error) {
	res, err := db.Exec(§INSERT INTO accounts (owner, balance) VALUES (?, ?)§, owner, balance)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func GetAccount(db *sql.DB, id int) (Account, error) {
	var a Account
	err := db.QueryRow(§SELECT id, owner, balance FROM accounts WHERE id = ?§, id).
		Scan(&a.ID, &a.Owner, &a.Balance)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, ErrAccountNotFound
	}
	if err != nil {
		return Account{}, err
	}
	return a, nil
}

func Transfer(db *sql.DB, from, to, amount int) error {
	if amount <= 0 {
		return ErrBadAmount
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var src Account
	err = tx.QueryRow(§SELECT id, owner, balance FROM accounts WHERE id = ?§, from).
		Scan(&src.ID, &src.Owner, &src.Balance)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAccountNotFound
	}
	if err != nil {
		return err
	}

	var dst Account
	err = tx.QueryRow(§SELECT id, owner, balance FROM accounts WHERE id = ?§, to).
		Scan(&dst.ID, &dst.Owner, &dst.Balance)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAccountNotFound
	}
	if err != nil {
		return err
	}

	if src.Balance < amount {
		return ErrInsufficientFunds
	}

	if _, err := tx.Exec(§UPDATE accounts SET balance = balance - ? WHERE id = ?§, amount, from); err != nil {
		return err
	}
	if _, err := tx.Exec(§UPDATE accounts SET balance = balance + ? WHERE id = ?§, amount, to); err != nil {
		return err
	}
	return tx.Commit()
}

func TotalBalance(db *sql.DB) (int, error) {
	var total int
	err := db.QueryRow(§SELECT COALESCE(SUM(balance), 0) FROM accounts§).Scan(&total)
	return total, err
}

func main() {
	db, err := sql.Open("sqlite", "file:bank.db?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	_ = Open(db)
	a, _ := CreateAccount(db, "ada", 100)
	b, _ := CreateAccount(db, "bo", 0)
	fmt.Println(Transfer(db, a, b, 30))
	fmt.Println(TotalBalance(db))
}
`,
			},
		},
	}
}

func m2t11() Topic {
	return Topic{
		ID:      "m2t11",
		Title:   "Less boilerplate with sqlx",
		Summary: "Scanning rows into structs, and named parameters.",
		Theory: `
§database/sql§ makes you list every column and every §&field§ by hand. §sqlx§ (package
§github.com/jmoiron/sqlx§) keeps the same API and adds:

* §Get§ / §Select§ — scan a row, or many rows, straight into a struct,
* §NamedExec§ — named parameters instead of counted §?§,
* §In§ — expand a slice into an §IN (…)§ list,
* §Rebind§ — switch placeholders between §?§ and §$1§.

It is a thin wrapper, not an ORM. The SQL is still yours.

## Struct tags map columns to fields

§§§go
import "github.com/jmoiron/sqlx"

type User struct {
	ID    int    §db:"id"§
	Name  string §db:"name"§
	Email string §db:"email"§
}
§§§

The §db:"…"§ tag works exactly like the §json:"…"§ tag you already know: it says which
column fills which field, and matches are case-insensitive when you omit it.

## Connecting

§§§go
db, err := sqlx.Connect("sqlite", dsn) // sql.Open + Ping in one call
if err != nil {
	return err
}
defer db.Close()
§§§

§sqlx.Connect§ pings for you, which is the check §database/sql§ makes you write.

You can also wrap an existing handle: §sqlx.NewDb(sqlDB, "sqlite")§.

## Get: one row

§§§go
var u User
err := db.Get(&u, "SELECT id, name, email FROM users WHERE id = ?", id)
if errors.Is(err, sql.ErrNoRows) {
	return ErrNotFound
}
§§§

No §Scan(&u.ID, &u.Name, &u.Email)§. §sqlx§ matches columns to fields by name.

## Select: many rows

§§§go
var users []User
err := db.Select(&users, "SELECT id, name, email FROM users ORDER BY id")
§§§

An empty result leaves §users§ as §nil§. If that is going into JSON, normalise it:

§§§go
if users == nil {
	users = []User{}
}
§§§

## Named parameters

§§§go
_, err := db.NamedExec(
	"INSERT INTO users (name, email) VALUES (:name, :email)",
	u,
)
§§§

The struct's §db§ tags become the §:name§ placeholders. §NamedExec§ also accepts a
§map[string]any§. §sqlx§ rewrites the query into §?§ form internally, so the driver still
sees ordinary parameters — injection is just as impossible.

## IN lists and placeholder rewriting

§§§go
q := "SELECT id, name FROM users WHERE id IN (?)"
q, args, err := sqlx.In(q, []int{1, 2, 3})
// q is now: ... WHERE id IN (?, ?, ?)   and args is []any{1, 2, 3}

q = db.Rebind(q) // converts ? to $1, $2 ... for PostgreSQL
rows, err := db.Queryx(q, args...)
§§§

§sqlx.In§ returns rewritten SQL *and* a flattened argument list — you must use both.
§Rebind§ is what makes the same repository work on SQLite and Postgres.

## Queryx: rows with struct scanning

§§§go
rows, err := db.Queryx("SELECT id, name FROM users")
if err != nil {
	return err
}
defer rows.Close()

for rows.Next() {
	var u User
	if err := rows.StructScan(&u); err != nil {
		return err
	}
	// ... use u
}
return rows.Err()
§§§

Useful when you want to process a large result set one struct at a time instead of
loading it all into a slice with §Select§.

## Transactions still work the same way

§§§go
tx, err := db.Beginx()
if err != nil {
	return err
}
defer tx.Rollback()

if _, err := tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, from); err != nil {
	return err
}
if _, err := tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, to); err != nil {
	return err
}
return tx.Commit()
§§§

A §*sqlx.Tx§ has §Get§, §Select§, §NamedExec§ and §Exec§ too, so you can keep the
convenience inside the transaction. §defer tx.Rollback()§ is still the safety net: it
returns §sql.ErrTxDone§ after a successful commit and is otherwise harmless.

## Which should I use?

| Reach for §database/sql§ when | Reach for §sqlx§ when |
| --- | --- |
| The query is one column, one value | You are scanning into structs |
| You want zero third-party dependencies | You want §IN§ lists and named params |
| The code is tiny and unlikely to grow | The code is a real repository |

They are the same library underneath, so mixing them is fine: §sqlx.DB§ embeds
§*sql.DB§, and §sqlx.NewDb§ wraps an existing handle. Many services use §database/sql§
in one repository and §sqlx§ in another.

## A complete example

§§§go
package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Score struct {
	ID     int    §db:"id"§
	Player string §db:"player"§
	Points int    §db:"points"§
}

type Scores struct {
	db *sqlx.DB
}

func NewScores(db *sqlx.DB) *Scores { return &Scores{db: db} }

func (s *Scores) Add(player string, points int) (int64, error) {
	res, err := s.db.NamedExec(
		"INSERT INTO scores (player, points) VALUES (:player, :points)",
		Score{Player: player, Points: points},
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Scores) Above(min int) ([]Score, error) {
	var out []Score
	err := s.db.Select(&out, "SELECT id, player, points FROM scores WHERE points >= ? ORDER BY points DESC", min)
	if out == nil {
		out = []Score{}
	}
	return out, err
}

func (s *Scores) One(id int) (Score, error) {
	var sc Score
	err := s.db.Get(&sc, "SELECT id, player, points FROM scores WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return Score{}, errors.New("score not found")
	}
	return sc, err
}

func main() {
	fmt.Println("see the exercise")
}
§§§

The scanning boilerplate is gone: no §Scan(&x.ID, &x.Player, …)§, no manual §[]Score{}§
juggling inside the loop.
`,
		Quiz: []Question{
			{
				ID:     "m2t11q1",
				Prompt: "What does §db.Get(&u, query, args…)§ do that §database/sql§ does not?",
				Options: []Option{
					{Text: "It scans the row into the struct's fields by matching column names to §db§ tags, so you do not write §Scan§ by hand.", Correct: true, Rationale: "It is the same query, with reflection doing the field wiring."},
					{Text: "It caches query results.", Correct: false, Rationale: "§sqlx§ has no query cache."},
					{Text: "It builds the SQL for you.", Correct: false, Rationale: "You still write the SQL."},
					{Text: "It runs the query in a transaction automatically.", Correct: false, Rationale: "You control transactions explicitly with §Beginx§."},
				},
				Explanation: "§sqlx§ removes boilerplate around scanning and parameters. The SQL stays yours.",
			},
			{
				ID:     "m2t11q2",
				Prompt: "What must you do with the results of §sqlx.In§?",
				Options: []Option{
					{Text: "Use *both* the rewritten query and the returned argument list, because the placeholders were expanded.", Correct: true, Rationale: "§In§ turns one §?§ into as many as the slice has elements, and flattens the args to match."},
					{Text: "Only the query — the args are unchanged.", Correct: false, Rationale: "The args are rebuilt to match the new placeholders."},
					{Text: "Nothing; it mutates the argument slice in place.", Correct: false, Rationale: "It returns new values."},
					{Text: "Call §Rebind§ first, then §In§.", Correct: false, Rationale: "§Rebind§ is for switching placeholder style; §In§ expands lists. Order depends on the driver."},
				},
				Explanation: "§q, args, err := sqlx.In(…)§ then §q = db.Rebind(q)§ for non-?¥ drivers.",
			},
			{
				ID:     "m2t11q3",
				Prompt: "§db.Select(&users, …)§ finds nothing. What is §users§?",
				Options: []Option{
					{Text: "§nil§ — so if it is going into JSON you should replace it with §[]User{}§.", Correct: true, Rationale: "An empty result leaves the slice untouched, and a nil slice marshals to §null§."},
					{Text: "An empty non-nil slice.", Correct: false, Rationale: "§Select§ does not initialise it for you."},
					{Text: "An error is returned instead.", Correct: false, Rationale: "Finding nothing is not an error."},
					{Text: "It panics.", Correct: false, Rationale: "No."},
				},
				Explanation: "The same nil-slice-versus-empty-JSON problem you met in Module 1's slice topic, now on the database boundary.",
			},
			{
				ID:     "m2t11q4",
				Prompt: "Why is §db.NamedExec§ still safe against SQL injection?",
				Options: []Option{
					{Text: "Because sqlx rewrites §:name§ into ordinary driver placeholders, so the value is still sent as a parameter.", Correct: true, Rationale: "It is sugar over the same parameter mechanism."},
					{Text: "Because it escapes quotes in the values.", Correct: false, Rationale: "Escaping is exactly the fragile approach that parameters replace."},
					{Text: "Because it only accepts trusted structs.", Correct: false, Rationale: "Struct field values come from user input too."},
					{Text: "It is not — you should avoid it.", Correct: false, Rationale: "It is as safe as §?§ placeholders."},
				},
				Explanation: "Named parameters are a convenience layer; the underlying parameter binding is unchanged.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m2t11e1",
				Title: "Warm-up: a repository with sqlx",
				Prompt: `Rewrite a repository using §sqlx§ and notice how much scanning code disappears.

§§§go
type Score struct {
	ID     int    §db:"id"§
	Player string §db:"player"§
	Points int    §db:"points"§
}

func Open(dsn string) (*sqlx.DB, error)   // connect + schema
func AddScore(db *sqlx.DB, player string, points int) (int64, error)
func GetScore(db *sqlx.DB, id int64) (Score, error)
func TopScores(db *sqlx.DB, min int, limit int) ([]Score, error)
func TotalPoints(db *sqlx.DB) (int, error)
§§§

Schema: §scores(id INTEGER PRIMARY KEY AUTOINCREMENT, player TEXT NOT NULL, points
INTEGER NOT NULL)§.

* §AddScore§ uses §db.NamedExec§ with §:player§ and §:points§.
* §GetScore§ uses §db.Get§ and maps §sql.ErrNoRows§ to §ErrScoreNotFound§.
* §TopScores§ returns scores with at least §min§ points, ordered by points descending,
  limited to §limit§. Use §db.Select§, and make sure an empty result is a **non-nil**
  empty slice.
* §TotalPoints§ uses §db.Get§ with §SELECT COALESCE(SUM(points), 0)§.

§sqlx§ will be added to the sandbox §go.mod§ automatically.
`,
				Starter: `package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var ErrScoreNotFound = errors.New("score not found")

type Score struct {
	ID     int    §db:"id"§
	Player string §db:"player"§
	Points int    §db:"points"§
}

// Open connects and creates the schema.
func Open(dsn string) (*sqlx.DB, error) {
	// TODO
	return nil, nil
}

func AddScore(db *sqlx.DB, player string, points int) (int64, error) {
	// TODO
	return 0, nil
}

func GetScore(db *sqlx.DB, id int64) (Score, error) {
	// TODO
	return Score{}, nil
}

func TopScores(db *sqlx.DB, min int, limit int) ([]Score, error) {
	// TODO
	return nil, nil
}

func TotalPoints(db *sqlx.DB) (int, error) {
	// TODO
	return 0, nil
}

func main() {
	db, err := Open("file:demo.db?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	id, _ := AddScore(db, "ada", 10)
	s, _ := GetScore(db, id)
	total, _ := TotalPoints(db)
	fmt.Println(s, total)

	_ = sql.ErrNoRows
}
`,
				Deps: []string{"github.com/jmoiron/sqlx", "modernc.org/sqlite"},
				Tests: []File{{
					Name: "scores_test.go",
					Content: `package main

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
)

func newRepo(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := Open("file:" + filepath.Join(t.TempDir(), "scores.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestAddAndGet(t *testing.T) {
	db := newRepo(t)

	id, err := AddScore(db, "ada", 42)
	if err != nil || id == 0 {
		t.Fatalf("AddScore = %d, %v", id, err)
	}

	s, err := GetScore(db, id)
	if err != nil {
		t.Fatalf("GetScore: %v", err)
	}
	if s.Player != "ada" || s.Points != 42 {
		t.Fatalf("GetScore = %+v", s)
	}
}

func TestGetMissing(t *testing.T) {
	db := newRepo(t)
	if _, err := GetScore(db, 999); !errors.Is(err, ErrScoreNotFound) {
		t.Fatalf("GetScore(missing) = %v, want ErrScoreNotFound", err)
	}
}

func TestTopScores(t *testing.T) {
	db := newRepo(t)
	for _, in := range []struct {
		player string
		points int
	}{{"ada", 10}, {"bo", 50}, {"cy", 30}, {"di", 5}} {
		if _, err := AddScore(db, in.player, in.points); err != nil {
			t.Fatal(err)
		}
	}

	got, err := TopScores(db, 20, 10)
	if err != nil {
		t.Fatalf("TopScores: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("TopScores(20) returned %d rows, want 2: %+v", len(got), got)
	}
	if got[0].Player != "bo" || got[1].Player != "cy" {
		t.Fatalf("TopScores order = %+v, want bo then cy", got)
	}

	limited, err := TopScores(db, 0, 1)
	if err != nil {
		t.Fatalf("TopScores: %v", err)
	}
	if len(limited) != 1 || limited[0].Player != "bo" {
		t.Fatalf("limit not applied: %+v", limited)
	}
}

func TestTopScoresEmptyIsNonNil(t *testing.T) {
	db := newRepo(t)
	got, err := TopScores(db, 1000, 10)
	if err != nil {
		t.Fatalf("TopScores: %v", err)
	}
	if got == nil {
		t.Fatal("TopScores returned nil for an empty result; return an empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestTotalPoints(t *testing.T) {
	db := newRepo(t)

	if got, err := TotalPoints(db); err != nil || got != 0 {
		t.Fatalf("TotalPoints(empty) = %d, %v; want 0, nil", got, err)
	}

	_, _ = AddScore(db, "a", 10)
	_, _ = AddScore(db, "b", 32)
	if got, err := TotalPoints(db); err != nil || got != 42 {
		t.Fatalf("TotalPoints = %d, %v; want 42, nil", got, err)
	}
}

func TestGetScoreMissingBoundaryIDs(t *testing.T) {
	db := newRepo(t)
	for _, id := range []int64{0, -1, 999} {
		if _, err := GetScore(db, id); !errors.Is(err, ErrScoreNotFound) {
			t.Errorf("GetScore(%d) = %v, want ErrScoreNotFound", id, err)
		}
	}
}

func TestAddScoreSamePlayerGetsDistinctRows(t *testing.T) {
	db := newRepo(t)

	first, err := AddScore(db, "ada", 10)
	if err != nil {
		t.Fatalf("AddScore: %v", err)
	}
	second, err := AddScore(db, "ada", 20)
	if err != nil {
		t.Fatalf("AddScore: %v", err)
	}
	if first == second {
		t.Fatalf("both scores got id %d; ids must be distinct", first)
	}

	for _, tc := range []struct {
		id     int64
		points int
	}{{first, 10}, {second, 20}} {
		s, err := GetScore(db, tc.id)
		if err != nil {
			t.Fatalf("GetScore(%d): %v", tc.id, err)
		}
		if s.Player != "ada" || s.Points != tc.points {
			t.Errorf("GetScore(%d) = %+v, want player ada, points %d", tc.id, s, tc.points)
		}
	}
}

func TestTopScoresThresholdIsInclusive(t *testing.T) {
	db := newRepo(t)
	for _, in := range []struct {
		player string
		points int
	}{{"low", 10}, {"edge", 20}, {"high", 30}} {
		if _, err := AddScore(db, in.player, in.points); err != nil {
			t.Fatal(err)
		}
	}

	got, err := TopScores(db, 20, 10)
	if err != nil {
		t.Fatalf("TopScores: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("TopScores(20) = %+v, want the edge and high scores", got)
	}
	if got[0].Player != "high" || got[1].Player != "edge" {
		t.Fatalf("TopScores order = %+v, want high then edge", got)
	}
}

func TestTopScoresLimitAndNegativePoints(t *testing.T) {
	db := newRepo(t)
	for _, in := range []struct {
		player string
		points int
	}{{"neg", -5}, {"zero", 0}, {"pos", 5}} {
		if _, err := AddScore(db, in.player, in.points); err != nil {
			t.Fatal(err)
		}
	}

	all, err := TopScores(db, -10, 100)
	if err != nil {
		t.Fatalf("TopScores: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("TopScores(-10, 100) returned %d rows, want all 3: %+v", len(all), all)
	}
	if all[0].Player != "pos" || all[1].Player != "zero" || all[2].Player != "neg" {
		t.Fatalf("TopScores order = %+v, want pos, zero, neg", all)
	}

	two, err := TopScores(db, -10, 2)
	if err != nil {
		t.Fatalf("TopScores: %v", err)
	}
	if len(two) != 2 || two[0].Player != "pos" || two[1].Player != "zero" {
		t.Fatalf("TopScores limit 2 = %+v, want pos and zero", two)
	}

	if got, err := TotalPoints(db); err != nil || got != 0 {
		t.Fatalf("TotalPoints = %d, %v; want 0, nil", got, err)
	}
}

func TestTopScoresNoMatchWithRowsIsNonNil(t *testing.T) {
	db := newRepo(t)
	_, _ = AddScore(db, "ada", 10)

	got, err := TopScores(db, 11, 10)
	if err != nil {
		t.Fatalf("TopScores: %v", err)
	}
	if got == nil {
		t.Fatal("TopScores returned nil for an empty result; return an empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("TopScores = %+v, want no rows", got)
	}
}

func TestTotalPointsWithNegativeScores(t *testing.T) {
	db := newRepo(t)
	_, _ = AddScore(db, "a", 10)
	_, _ = AddScore(db, "b", -15)

	got, err := TotalPoints(db)
	if err != nil {
		t.Fatalf("TotalPoints: %v", err)
	}
	if got != -5 {
		t.Fatalf("TotalPoints = %d, want -5", got)
	}
}

var _ = sql.ErrNoRows
`,
				}},
				Hints: []string{
					"§sqlx.Connect(\"sqlite\", dsn)§ pings for you. Set §db.SetMaxOpenConns(1)§ for SQLite, then run the §CREATE TABLE§.",
					"§AddScore§: §db.NamedExec(§INSERT INTO scores (player, points) VALUES (:player, :points)§, Score{Player: player, Points: points})§.",
					"§GetScore§: §db.Get(&s, §SELECT id, player, points FROM scores WHERE id = ?§, id)§ then check §errors.Is(err, sql.ErrNoRows)§.",
					"§TopScores§: §var out []Score; err := db.Select(&out, §SELECT id, player, points FROM scores WHERE points >= ? ORDER BY points DESC LIMIT ?§, min, limit)§, then §if out == nil { out = []Score{} }§.",
					"§TotalPoints§: §var total int; err := db.Get(&total, §SELECT COALESCE(SUM(points), 0) FROM scores§)§.",
				},
				Solution: `package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var ErrScoreNotFound = errors.New("score not found")

type Score struct {
	ID     int    §db:"id"§
	Player string §db:"player"§
	Points int    §db:"points"§
}

func Open(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(§CREATE TABLE IF NOT EXISTS scores (
		id     INTEGER PRIMARY KEY AUTOINCREMENT,
		player TEXT NOT NULL,
		points INTEGER NOT NULL
	)§); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func AddScore(db *sqlx.DB, player string, points int) (int64, error) {
	res, err := db.NamedExec(
		§INSERT INTO scores (player, points) VALUES (:player, :points)§,
		Score{Player: player, Points: points},
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetScore(db *sqlx.DB, id int64) (Score, error) {
	var s Score
	err := db.Get(&s, §SELECT id, player, points FROM scores WHERE id = ?§, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Score{}, ErrScoreNotFound
	}
	if err != nil {
		return Score{}, err
	}
	return s, nil
}

func TopScores(db *sqlx.DB, min int, limit int) ([]Score, error) {
	var out []Score
	err := db.Select(&out,
		§SELECT id, player, points FROM scores WHERE points >= ? ORDER BY points DESC LIMIT ?§,
		min, limit)
	if out == nil {
		out = []Score{}
	}
	return out, err
}

func TotalPoints(db *sqlx.DB) (int, error) {
	var total int
	err := db.Get(&total, §SELECT COALESCE(SUM(points), 0) FROM scores§)
	return total, err
}

func main() {
	db, err := Open("file:demo.db?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	id, _ := AddScore(db, "ada", 10)
	s, _ := GetScore(db, id)
	total, _ := TotalPoints(db)
	fmt.Println(s, total)
}
`,
			},
			{
				ID:    "m2t11e2",
				Title: "A transfer with sqlx, and an IN list",
				Prompt: `Two more §sqlx§ features: transactions with §Beginx§, and §In§ for lists.

§§§go
type Account struct {
	ID      int §db:"id"§
	Owner   string §db:"owner"§
	Balance int §db:"balance"§
}

func Open(db *sqlx.DB) error
func CreateAccount(db *sqlx.DB, owner string, balance int) (int, error)
func GetAccount(db *sqlx.DB, id int) (Account, error)
func Transfer(db *sqlx.DB, from, to, amount int) error
func AccountsFor(db *sqlx.DB, ids []int) ([]Account, error)
§§§

* §Transfer§ runs in a §Beginx§ transaction with §defer tx.Rollback()§ and reads both
  accounts with §tx.Get§ before changing anything. §amount <= 0§ is rejected. It must be
  atomic: a failure changes nothing, and the total balance is unchanged on success.
* §AccountsFor§ fetches several accounts by id in a single query. Build the query with
  §sqlx.In§, then §db.Rebind§, then §db.Selectx§… or simply §db.Select§ with the
  expanded arguments. With no ids, return an empty slice without querying at all.
* Results must be ordered by id so the test can compare them.

Because §sqlx.In§ rewrites one §?§ into several, this is the query that a hand-rolled
§database/sql§ version gets wrong by forgetting to spread the arguments.
`,
				Starter: `package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Account struct {
	ID      int    §db:"id"§
	Owner   string §db:"owner"§
	Balance int    §db:"balance"§
}

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrAccountNotFound   = errors.New("account not found")
	ErrBadAmount         = errors.New("amount must be positive")
)

func Open(db *sqlx.DB) error {
	// TODO
	return nil
}

func CreateAccount(db *sqlx.DB, owner string, balance int) (int, error) {
	// TODO
	return 0, nil
}

func GetAccount(db *sqlx.DB, id int) (Account, error) {
	// TODO
	return Account{}, nil
}

func Transfer(db *sqlx.DB, from, to, amount int) error {
	// TODO
	return nil
}

func AccountsFor(db *sqlx.DB, ids []int) ([]Account, error) {
	// TODO
	return nil, nil
}

func main() {
	db, err := sqlx.Connect("sqlite", "file:bank.db?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	_ = Open(db)
	a, _ := CreateAccount(db, "ada", 100)
	b, _ := CreateAccount(db, "bo", 0)
	fmt.Println(Transfer(db, a, b, 30))
	fmt.Println(AccountsFor(db, []int{a, b}))

	_ = sql.ErrNoRows
}
`,
				Deps: []string{"github.com/jmoiron/sqlx", "modernc.org/sqlite"},
				Tests: []File{{
					Name: "sqlxtx_test.go",
					Content: `package main

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func newBank(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Connect("sqlite", "file:"+filepath.Join(t.TempDir(), "bank.db"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	if err := Open(db); err != nil {
		t.Fatalf("Open: %v", err)
	}
	return db
}

func TestCreateAndGet(t *testing.T) {
	db := newBank(t)

	id, err := CreateAccount(db, "ada", 100)
	if err != nil || id == 0 {
		t.Fatalf("CreateAccount = %d, %v", id, err)
	}
	a, err := GetAccount(db, id)
	if err != nil {
		t.Fatalf("GetAccount: %v", err)
	}
	if a.Owner != "ada" || a.Balance != 100 {
		t.Fatalf("account = %+v", a)
	}
	if _, err := GetAccount(db, 999); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("missing = %v, want ErrAccountNotFound", err)
	}
}

func TestTransferSuccess(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 100)
	b, _ := CreateAccount(db, "bo", 20)

	if err := Transfer(db, a, b, 30); err != nil {
		t.Fatalf("Transfer: %v", err)
	}
	gotA, _ := GetAccount(db, a)
	gotB, _ := GetAccount(db, b)
	if gotA.Balance != 70 || gotB.Balance != 50 {
		t.Fatalf("balances = %d, %d; want 70, 50", gotA.Balance, gotB.Balance)
	}
}

func TestTransferAtomic(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 10)
	b, _ := CreateAccount(db, "bo", 0)

	if err := Transfer(db, a, b, 999); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("Transfer = %v, want ErrInsufficientFunds", err)
	}
	gotA, _ := GetAccount(db, a)
	if gotA.Balance != 10 {
		t.Fatalf("a failed transfer changed the balance: %d", gotA.Balance)
	}

	if err := Transfer(db, a, 999, 5); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("Transfer to a missing account = %v, want ErrAccountNotFound", err)
	}
	gotA, _ = GetAccount(db, a)
	if gotA.Balance != 10 {
		t.Fatalf("balance changed on a failed transfer: %d", gotA.Balance)
	}

	for _, amount := range []int{0, -1} {
		if err := Transfer(db, a, b, amount); err == nil {
			t.Fatalf("Transfer(amount=%d) should fail", amount)
		}
	}
}

func TestAccountsFor(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 1)
	b, _ := CreateAccount(db, "bo", 2)
	c, _ := CreateAccount(db, "cy", 3)

	got, err := AccountsFor(db, []int{a, c})
	if err != nil {
		t.Fatalf("AccountsFor: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("AccountsFor returned %d rows, want 2: %+v", len(got), got)
	}
	if got[0].ID != a || got[1].ID != c {
		t.Fatalf("AccountsFor = %+v, want ids %d and %d in order", got, a, c)
	}

	all, err := AccountsFor(db, []int{a, b, c})
	if err != nil || len(all) != 3 {
		t.Fatalf("AccountsFor(all) = %+v, %v", all, err)
	}

	none, err := AccountsFor(db, nil)
	if err != nil {
		t.Fatalf("AccountsFor(nil) error = %v", err)
	}
	if none == nil || len(none) != 0 {
		t.Fatalf("AccountsFor(nil) = %v, want an empty non-nil slice", none)
	}
}

func TestGetAccountMissingBoundaryIDs(t *testing.T) {
	db := newBank(t)
	for _, id := range []int{0, -1, 999} {
		if _, err := GetAccount(db, id); !errors.Is(err, ErrAccountNotFound) {
			t.Errorf("GetAccount(%d) = %v, want ErrAccountNotFound", id, err)
		}
	}
}

func TestTransferExactBalanceAndOneOver(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 50)
	b, _ := CreateAccount(db, "bo", 10)

	if err := Transfer(db, a, b, 51); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("Transfer(51 from 50) = %v, want ErrInsufficientFunds", err)
	}
	gotA, _ := GetAccount(db, a)
	gotB, _ := GetAccount(db, b)
	if gotA.Balance != 50 || gotB.Balance != 10 {
		t.Fatalf("the failed boundary transfer changed balances: %d, %d", gotA.Balance, gotB.Balance)
	}

	if err := Transfer(db, a, b, 50); err != nil {
		t.Fatalf("Transfer(exact balance): %v", err)
	}
	gotA, _ = GetAccount(db, a)
	gotB, _ = GetAccount(db, b)
	if gotA.Balance != 0 || gotB.Balance != 60 {
		t.Fatalf("balances = %d, %d; want 0, 60", gotA.Balance, gotB.Balance)
	}
}

func TestTransferMissingSource(t *testing.T) {
	db := newBank(t)
	b, _ := CreateAccount(db, "bo", 15)

	if err := Transfer(db, 999, b, 5); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("Transfer from a missing account = %v, want ErrAccountNotFound", err)
	}
	gotB, _ := GetAccount(db, b)
	if gotB.Balance != 15 {
		t.Fatalf("the destination changed on a failed transfer: %d", gotB.Balance)
	}
}

func TestTransferBadAmountIsTypedAndAtomic(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 100)
	b, _ := CreateAccount(db, "bo", 25)

	for _, amount := range []int{0, -1, -100} {
		if err := Transfer(db, a, b, amount); !errors.Is(err, ErrBadAmount) {
			t.Errorf("Transfer(amount=%d) = %v, want ErrBadAmount", amount, err)
		}
	}
	gotA, _ := GetAccount(db, a)
	gotB, _ := GetAccount(db, b)
	if gotA.Balance != 100 || gotB.Balance != 25 {
		t.Fatalf("a rejected transfer changed balances: %d, %d", gotA.Balance, gotB.Balance)
	}
}

func TestAccountsForOrderDuplicatesAndMissing(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 1)
	b, _ := CreateAccount(db, "bo", 2)
	c, _ := CreateAccount(db, "cy", 3)

	// Unsorted input with a duplicate: SQL's IN deduplicates, and the result
	// must still come back ordered by id.
	got, err := AccountsFor(db, []int{c, a, c})
	if err != nil {
		t.Fatalf("AccountsFor: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("AccountsFor([c a c]) returned %d rows, want 2: %+v", len(got), got)
	}
	if got[0].ID != a || got[1].ID != c {
		t.Fatalf("AccountsFor([c a c]) = %+v, want ids %d then %d", got, a, c)
	}

	mixed, err := AccountsFor(db, []int{b, 999})
	if err != nil {
		t.Fatalf("AccountsFor(missing id): %v", err)
	}
	if len(mixed) != 1 || mixed[0].ID != b {
		t.Fatalf("AccountsFor([b 999]) = %+v, want only id %d", mixed, b)
	}

	none, err := AccountsFor(db, []int{998, 999})
	if err != nil {
		t.Fatalf("AccountsFor(all missing): %v", err)
	}
	if none == nil || len(none) != 0 {
		t.Fatalf("AccountsFor(all missing) = %v, want an empty non-nil slice", none)
	}

	empty, err := AccountsFor(db, []int{})
	if err != nil {
		t.Fatalf("AccountsFor(empty): %v", err)
	}
	if empty == nil || len(empty) != 0 {
		t.Fatalf("AccountsFor(empty) = %v, want an empty non-nil slice", empty)
	}
}

func TestSequentialTransfersPreserveTotal(t *testing.T) {
	db := newBank(t)
	a, _ := CreateAccount(db, "ada", 100)
	b, _ := CreateAccount(db, "bo", 0)

	if err := Transfer(db, a, b, 30); err != nil {
		t.Fatalf("first Transfer: %v", err)
	}
	if err := Transfer(db, b, a, 10); err != nil {
		t.Fatalf("second Transfer: %v", err)
	}
	gotA, _ := GetAccount(db, a)
	gotB, _ := GetAccount(db, b)
	if gotA.Balance != 80 || gotB.Balance != 20 {
		t.Fatalf("balances = %d, %d; want 80, 20", gotA.Balance, gotB.Balance)
	}

	var total int
	if err := db.Get(&total, "SELECT COALESCE(SUM(balance), 0) FROM accounts"); err != nil {
		t.Fatalf("sum: %v", err)
	}
	if total != 100 {
		t.Fatalf("total balance = %d, want 100", total)
	}
}

var _ = sql.ErrNoRows
`,
				}},
				Hints: []string{
					"§Open§ creates §accounts(id INTEGER PRIMARY KEY AUTOINCREMENT, owner TEXT NOT NULL, balance INTEGER NOT NULL)§.",
					"§GetAccount§: §db.Get(&a, §SELECT id, owner, balance FROM accounts WHERE id = ?§, id)§ and map §sql.ErrNoRows§.",
					"§Transfer§: §tx, err := db.Beginx()§, §defer tx.Rollback()§, then §tx.Get§ both rows, check, §tx.Exec§ twice, §tx.Commit()§.",
					"§AccountsFor§: §if len(ids) == 0 { return []Account{}, nil }§ first, then §q, args, err := sqlx.In(§SELECT id, owner, balance FROM accounts WHERE id IN (?) ORDER BY id§, ids)§, §q = db.Rebind(q)§, then §db.Select(&out, q, args...)§.",
					"Note the §args...§ — §sqlx.In§ returns a slice of arguments that must be spread into the query.",
				},
				Solution: `package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Account struct {
	ID      int    §db:"id"§
	Owner   string §db:"owner"§
	Balance int    §db:"balance"§
}

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrAccountNotFound   = errors.New("account not found")
	ErrBadAmount         = errors.New("amount must be positive")
)

func Open(db *sqlx.DB) error {
	_, err := db.Exec(§CREATE TABLE IF NOT EXISTS accounts (
		id      INTEGER PRIMARY KEY AUTOINCREMENT,
		owner   TEXT NOT NULL,
		balance INTEGER NOT NULL
	)§)
	return err
}

func CreateAccount(db *sqlx.DB, owner string, balance int) (int, error) {
	res, err := db.NamedExec(
		§INSERT INTO accounts (owner, balance) VALUES (:owner, :balance)§,
		Account{Owner: owner, Balance: balance},
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func GetAccount(db *sqlx.DB, id int) (Account, error) {
	var a Account
	err := db.Get(&a, §SELECT id, owner, balance FROM accounts WHERE id = ?§, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, ErrAccountNotFound
	}
	if err != nil {
		return Account{}, err
	}
	return a, nil
}

func Transfer(db *sqlx.DB, from, to, amount int) error {
	if amount <= 0 {
		return ErrBadAmount
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var src Account
	err = tx.Get(&src, §SELECT id, owner, balance FROM accounts WHERE id = ?§, from)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAccountNotFound
	}
	if err != nil {
		return err
	}

	var dst Account
	err = tx.Get(&dst, §SELECT id, owner, balance FROM accounts WHERE id = ?§, to)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAccountNotFound
	}
	if err != nil {
		return err
	}

	if src.Balance < amount {
		return ErrInsufficientFunds
	}

	if _, err := tx.Exec(§UPDATE accounts SET balance = balance - ? WHERE id = ?§, amount, from); err != nil {
		return err
	}
	if _, err := tx.Exec(§UPDATE accounts SET balance = balance + ? WHERE id = ?§, amount, to); err != nil {
		return err
	}
	return tx.Commit()
}

func AccountsFor(db *sqlx.DB, ids []int) ([]Account, error) {
	if len(ids) == 0 {
		return []Account{}, nil
	}

	q, args, err := sqlx.In(
		§SELECT id, owner, balance FROM accounts WHERE id IN (?) ORDER BY id§, ids)
	if err != nil {
		return nil, err
	}
	q = db.Rebind(q)

	var out []Account
	if err := db.Select(&out, q, args...); err != nil {
		return nil, err
	}
	if out == nil {
		out = []Account{}
	}
	return out, nil
}

func main() {
	db, err := sqlx.Connect("sqlite", "file:bank.db?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	_ = Open(db)
	a, _ := CreateAccount(db, "ada", 100)
	b, _ := CreateAccount(db, "bo", 0)
	fmt.Println(Transfer(db, a, b, 30))
	fmt.Println(AccountsFor(db, []int{a, b}))
}
`,
			},
		},
	}
}

func m2t12() Topic {
	return Topic{
		ID:      "m2t12",
		Title:   "Configuration, structured logging and testing",
		Summary: "The operational half of a service: settings, log output, shutdown, and tests that catch bugs.",
		Theory: `
A service that cannot be configured, cannot be observed, or has no tests is not finished.
This topic covers the three things you add before you deploy.

## Configuration from the environment

Twelve-factor style: read settings from the environment, validate them **once** at
startup, and pass a typed struct around. Do not sprinkle §os.Getenv§ through the
codebase — you will not know what settings exist, and nothing will fail when one is
missing.

§§§go
type Config struct {
	Addr         string
	DatabaseURL  string
	LogLevel     slog.Level
	ReadTimeout  time.Duration
}

func LoadConfig(getenv func(string) string) (Config, error) {
	cfg := Config{
		Addr:        ":8080",
		DatabaseURL: getenv("DATABASE_URL"),
		LogLevel:    slog.LevelInfo,
		ReadTimeout: 10 * time.Second,
	}

	var errs []error
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	...
	return cfg, errors.Join(errs...)
}
§§§

Taking the lookup function as a parameter (§getenv func(string) string§) instead of
calling §os.Getenv§ directly is what makes this testable: a test passes a map lookup, and
never mutates the real process environment.

**Fail fast**: if the configuration is invalid, log and exit before binding a port. A
service that starts with half a config is worse than one that refuses to start.

## Structured logging with log/slog

§log/slog§ (Go 1.21+) produces **structured** logs: key/value pairs a log aggregator can
index, instead of prose you have to regex.

§§§go
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

slog.Info("server started", "addr", ":8080", "env", "prod")
log := slog.With("request_id", "abc-123") // every line from here carries the id
log.Error("query failed", "err", err, "table", "users")
§§§

JSON output — one object per line, which is the point:

§§§json
{"time":"2024-05-05T10:00:00Z","level":"INFO","msg":"server started","addr":":8080","env":"prod"}
§§§

| Handler | Use |
| --- | --- |
| §slog.NewTextHandler(w, opts)§ | §key=value§ for local development |
| §slog.NewJSONHandler(w, opts)§ | production, one JSON object per line |

Levels, from chattiest to most severe: §Debug§ < §Info§ < §Warn§ < §Error§.

Two habits that matter more than the API:

* **Log identifiers, never secrets.** No passwords, tokens, cookies, §Authorization§
  headers or whole request bodies.
* **Use §With§ for repeated context** (request id, user id) so you do not thread the same
  argument through every call.

## Graceful shutdown

Deployments, restarts and §Ctrl-C§ all need the same thing: stop taking new work, finish
the work in flight, then exit.

§§§go
func run(ctx context.Context) error {
	cfg, err := LoadConfig(os.Getenv)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	}
}
§§§

Key points:

* §signal.NotifyContext§ gives you a context cancelled by §Ctrl-C§ or §SIGTERM§.
* §srv.Shutdown§ stops accepting, waits for in-flight requests, then returns.
* §ListenAndServe§ returns §http.ErrServerClosed§ after a clean shutdown — **filter it
  out**, or you log a false failure on every deploy.
* Give shutdown its own timeout. A hung request must not hang the deployment.
* Close dependencies **after** the server: defers run last-in-first-out, so
  §defer db.Close()§ registered before §defer analytics.Stop()§ runs after it. That is
  what you want.

## Health and readiness

Two tiny endpoints that an orchestrator needs:

§§§go
mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK) // the process is alive
})

mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
})
§§§

The orchestrator restarts the container on a failing §/healthz§, and withholds traffic
while §/readyz§ fails.

## Testing backend code

The Go test framework is in the standard library: files named §*_test.go§, functions
named §TestXxx(t *testing.T)§, run with §go test§.

### Table-driven tests with subtests

§§§go
func TestGrade(t *testing.T) {
	tests := []struct {
		name  string
		score int
		want  string
	}{
		{"top", 95, "A"},
		{"edge", 90, "A"},
		{"fail", 10, "F"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Grade(tc.score); got != tc.want {
				t.Errorf("Grade(%d) = %q, want %q", tc.score, got, tc.want)
			}
		})
	}
}

go test -run 'TestGrade/edge' ./...
§§§

| Call | Effect |
| --- | --- |
| §t.Errorf§ | record a failure and keep going |
| §t.Fatalf§ | record a failure and stop this test |
| §t.Skip§ | skip (usually behind an env var) |
| §t.Run§ | a named subtest, separately reportable |
| §t.Parallel()§ | run concurrently with other parallel tests |
| §t.Cleanup(f)§ | teardown, runs LIFO at the end |
| §t.Helper()§ | mark a helper so failures report the caller's line |
| §t.TempDir()§ | a fresh directory, removed automatically |
| §t.Setenv(k, v)§ | set an environment variable for this test only |

§t.TempDir()§ and §t.Cleanup§ are how you avoid leaving databases and files behind — each
test gets an isolated filesystem and its own teardown.

### Testing HTTP without a network

§net/http/httptest§ is the standard toolkit:

§§§go
req := httptest.NewRequest(http.MethodGet, "/items/7", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

if rec.Code != http.StatusOK { ... }
// rec.Body.Bytes() is your JSON
§§§

* §httptest.NewRequest§ + §httptest.NewRecorder§ — fast, in-memory, no sockets.
* §httptest.NewServer(handler)§ — a real server on a random port, when you need a real
  client, TLS, redirects or streaming.

### Fakes instead of mocks

Define a narrow interface where it is **used**, then supply a five-line fake in the test:

§§§go
type UserStore interface {
	Get(ctx context.Context, id int) (User, error)
}

type fakeStore struct {
	users map[int]User
	err   error
}

func (f fakeStore) Get(ctx context.Context, id int) (User, error) {
	if f.err != nil {
		return User{}, f.err
	}
	u, ok := f.users[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}
§§§

For databases, prefer a **real** SQLite file in §t.TempDir()§: it runs in milliseconds and
catches SQL mistakes a mock never would. Reserve fakes for genuinely slow or external
things (email, payments).

### Commands worth knowing

§§§bash
go test ./...                    # everything
go test -race ./...              # data-race detection — always in CI
go test -cover ./...             # coverage percentage
go test -run TestGrade -v ./...  # one test, verbose
go test -count=1 ./...           # bypass the test result cache
go test -short ./...             # skip slow tests via testing.Short()
go vet ./...                     # static analysis
§§§

§go test -race§ plus §go vet§ plus the normal run is the standard CI trio.
`,
		Quiz: []Question{
			{
				ID:     "m2t12q1",
				Prompt: "Why is §LoadConfig(getenv func(string) string)§ better than reading §os.Getenv§ directly?",
				Options: []Option{
					{Text: "Tests can inject values without mutating the process environment, and the dependency is explicit.", Correct: true, Rationale: "Pure functions are easy to test, and parallel tests do not leak state into each other."},
					{Text: "§os.Getenv§ is deprecated.", Correct: false, Rationale: "It is current and correct at the edges of the program."},
					{Text: "Because environment variables are not thread-safe.", Correct: false, Rationale: "§os.Getenv§ is safe for concurrent use."},
					{Text: "Because containers cannot set environment variables.", Correct: false, Rationale: "They set them all the time — that is the point of twelve-factor config."},
				},
				Explanation: "Inverting the environment dependency keeps configuration logic pure and testable.",
			},
			{
				ID:     "m2t12q2",
				Prompt: "Which are true about §http.Server.Shutdown§? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "It stops accepting new connections.", Correct: true, Rationale: "The listener closes first."},
					{Text: "It waits for active requests to finish, up to the context deadline.", Correct: true, Rationale: "Then it returns, possibly with §context.DeadlineExceeded§."},
					{Text: "It makes §ListenAndServe§ return §http.ErrServerClosed§.", Correct: true, Rationale: "Filter that error out, or every clean deploy logs a false failure."},
					{Text: "It closes idle connections immediately.", Correct: true, Rationale: "Idle connections are dropped right away."},
					{Text: "It kills in-flight requests immediately.", Correct: false, Rationale: "That is §Close()§; §Shutdown§ is the graceful version."},
				},
				Explanation: "§ErrServerClosed§ from §ListenAndServe§ is the expected signal of a clean shutdown.",
			},
			{
				ID:     "m2t12q3",
				Prompt: "What is the difference between §t.Errorf§ and §t.Fatalf§?",
				Options: []Option{
					{Text: "§Errorf§ records the failure and continues; §Fatalf§ stops the current test function immediately.", Correct: true, Rationale: "§Fatalf§ calls §runtime.Goexit§, so deferred cleanups still run."},
					{Text: "§Fatalf§ fails the whole package; §Errorf§ fails one subtest.", Correct: false, Rationale: "Both fail the test that called them; package failure is just the aggregate."},
					{Text: "They are identical.", Correct: false, Rationale: "They differ in control flow, and that matters."},
					{Text: "§Errorf§ panics.", Correct: false, Rationale: "It records a failure."},
				},
				Explanation: "Use §Fatalf§ only when continuing would be meaningless or dangerous — a nil pointer, a missing fixture.",
			},
			{
				ID:     "m2t12q4",
				Prompt: "How should a test wait for asynchronous work to finish?",
				Options: []Option{
					{Text: "Synchronise deterministically — poll a counter with a deadline, or call a blocking §Stop()§/§Flush()§.", Correct: true, Rationale: "Either is deterministic; a §Stop§ that drains and waits is the cleanest."},
					{Text: "§time.Sleep(100 * time.Millisecond)§.", Correct: false, Rationale: "Sleeps make tests slow *and* flaky."},
					{Text: "It cannot be tested.", Correct: false, Rationale: "It can, with proper synchronisation."},
					{Text: "Read the channel from the test to drain it.", Correct: false, Rationale: "Consuming in the test steals work from the workers."},
				},
				Explanation: "Channels, WaitGroups and counted semaphores are what make concurrency testable.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m2t12e1",
				Title: "Warm-up: a config loader and a JSON logger",
				Prompt: `**Part 1 — configuration.**

§§§go
type Config struct {
	Addr        string
	DatabaseURL string
	LogLevel    slog.Level
	ReadTimeout time.Duration
}

func LoadConfig(getenv func(string) string) (Config, error)
§§§

* §ADDR§ defaults to §":8080"§.
* §DATABASE_URL§ is **required** — its absence is an error.
* §LOG_LEVEL§ accepts §debug§ / §info§ / §warn§ / §error§ (default §info§); anything else
  is an error that lists the accepted values.
* §READ_TIMEOUT§ is parsed with §time.ParseDuration§, default §"10s"§; an invalid value
  is an error.
* All problems must be reported **together** with §errors.Join§, and each message must
  name the offending environment variable (§DATABASE_URL§, §LOG_LEVEL§ or
  §READ_TIMEOUT§) so the operator knows what to fix.

**Part 2 — logging.**

§§§go
func NewJSONLogger(w io.Writer, level slog.Level) *slog.Logger
func LogRequestStart(l *slog.Logger, method, path, requestID string)
func LogRequestEnd(l *slog.Logger, method, path string, status int, d time.Duration)
§§§

* The logger must filter out records below §level§.
* The helpers emit §msg§ values §"request started"§ and §"request finished"§ with fields
  §method§, §path§, §request_id§ and (for the second) §status§ and §duration_ms§.
* §duration_ms§ should be an integer, so use §d.Milliseconds()§.

The test parses your JSON with §encoding/json§, so the field names must match exactly.
`,
				Starter: `package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

type Config struct {
	Addr        string
	DatabaseURL string
	LogLevel    slog.Level
	ReadTimeout time.Duration
}

// LoadConfig reads configuration from the supplied environment lookup.
func LoadConfig(getenv func(string) string) (Config, error) {
	// TODO
	return Config{}, nil
}

func NewJSONLogger(w io.Writer, level slog.Level) *slog.Logger {
	// TODO
	return nil
}

func LogRequestStart(l *slog.Logger, method, path, requestID string) {
	// TODO
}

func LogRequestEnd(l *slog.Logger, method, path string, status int, d time.Duration) {
	// TODO
}

func main() {
	cfg, err := LoadConfig(os.Getenv)
	fmt.Println(cfg, err)

	logger := NewJSONLogger(os.Stdout, slog.LevelInfo)
	LogRequestStart(logger, "GET", "/users", "req-1")

	_ = errors.Join
}
`,
				Tests: []File{{
					Name: "config_test.go",
					Content: `package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := LoadConfig(env(map[string]string{"DATABASE_URL": "file:x.db"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want :8080", cfg.Addr)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want Info", cfg.LogLevel)
	}
	if cfg.ReadTimeout != 10*time.Second {
		t.Errorf("ReadTimeout = %v, want 10s", cfg.ReadTimeout)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	cfg, err := LoadConfig(env(map[string]string{
		"DATABASE_URL": "postgres://x",
		"ADDR":         ":9999",
		"LOG_LEVEL":    "debug",
		"READ_TIMEOUT": "3s",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Addr != ":9999" || cfg.LogLevel != slog.LevelDebug || cfg.ReadTimeout != 3*time.Second {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestLoadConfigErrors(t *testing.T) {
	if _, err := LoadConfig(env(nil)); err == nil {
		t.Fatal("a missing DATABASE_URL should be an error")
	}
	if _, err := LoadConfig(env(map[string]string{"DATABASE_URL": "x", "LOG_LEVEL": "verbose"})); err == nil {
		t.Fatal("an invalid LOG_LEVEL should be an error")
	}
	if _, err := LoadConfig(env(map[string]string{"DATABASE_URL": "x", "READ_TIMEOUT": "banana"})); err == nil {
		t.Fatal("an invalid READ_TIMEOUT should be an error")
	}

	_, err := LoadConfig(env(map[string]string{"LOG_LEVEL": "nope", "READ_TIMEOUT": "banana"}))
	if err == nil {
		t.Fatal("expected a joined error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "LOG_LEVEL") || !strings.Contains(msg, "READ_TIMEOUT") {
		t.Fatalf("errors.Join did not combine both problems: %q", msg)
	}
}

func TestLoadConfigEmptyValuesUseDefaults(t *testing.T) {
	cfg, err := LoadConfig(env(map[string]string{
		"DATABASE_URL": "file:x.db",
		"ADDR":         "",
		"LOG_LEVEL":    "",
		"READ_TIMEOUT": "",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want :8080", cfg.Addr)
	}
	if cfg.DatabaseURL != "file:x.db" {
		t.Errorf("DatabaseURL = %q, want file:x.db", cfg.DatabaseURL)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want Info", cfg.LogLevel)
	}
	if cfg.ReadTimeout != 10*time.Second {
		t.Errorf("ReadTimeout = %v, want 10s", cfg.ReadTimeout)
	}
}

func TestLoadConfigLogLevels(t *testing.T) {
	tests := []struct {
		raw  string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
	}
	for _, tc := range tests {
		t.Run(tc.raw, func(t *testing.T) {
			cfg, err := LoadConfig(env(map[string]string{
				"DATABASE_URL": "file:x.db",
				"LOG_LEVEL":    tc.raw,
			}))
			if err != nil {
				t.Fatalf("LoadConfig: %v", err)
			}
			if cfg.LogLevel != tc.want {
				t.Fatalf("LogLevel = %v, want %v", cfg.LogLevel, tc.want)
			}
		})
	}
}

func TestLoadConfigReadTimeoutVariants(t *testing.T) {
	tests := []struct {
		raw  string
		want time.Duration
	}{
		{"250ms", 250 * time.Millisecond},
		{"1m30s", 90 * time.Second},
		{"0s", 0},
	}
	for _, tc := range tests {
		t.Run(tc.raw, func(t *testing.T) {
			cfg, err := LoadConfig(env(map[string]string{
				"DATABASE_URL": "file:x.db",
				"READ_TIMEOUT": tc.raw,
			}))
			if err != nil {
				t.Fatalf("LoadConfig: %v", err)
			}
			if cfg.ReadTimeout != tc.want {
				t.Fatalf("ReadTimeout = %v, want %v", cfg.ReadTimeout, tc.want)
			}
		})
	}
}

func TestLoadConfigInvalidLogLevelListsAcceptedValues(t *testing.T) {
	_, err := LoadConfig(env(map[string]string{
		"DATABASE_URL": "file:x.db",
		"LOG_LEVEL":    "verbose",
	}))
	if err == nil {
		t.Fatal("an invalid LOG_LEVEL should be an error")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "log_level") {
		t.Fatalf("the error does not name LOG_LEVEL: %q", err)
	}
	for _, want := range []string{"debug", "info", "warn", "error"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the error does not list the accepted value %q: %q", want, err)
		}
	}
}

func TestLoadConfigInvalidValuesAreCaseSensitive(t *testing.T) {
	for _, raw := range []string{"INFO", "Debug", "warning"} {
		if _, err := LoadConfig(env(map[string]string{
			"DATABASE_URL": "file:x.db",
			"LOG_LEVEL":    raw,
		})); err == nil {
			t.Errorf("LOG_LEVEL %q should be rejected; only debug, info, warn and error are accepted", raw)
		}
	}
}

func TestLoadConfigBadTimeoutIsNamed(t *testing.T) {
	for _, raw := range []string{"banana", "10", "-"} {
		_, err := LoadConfig(env(map[string]string{
			"DATABASE_URL": "file:x.db",
			"READ_TIMEOUT": raw,
		}))
		if err == nil {
			t.Errorf("READ_TIMEOUT %q should be an error", raw)
			continue
		}
		if !strings.Contains(err.Error(), "READ_TIMEOUT") {
			t.Errorf("the error for READ_TIMEOUT %q does not name the variable: %q", raw, err)
		}
	}
}

func TestLoadConfigJoinsAllErrors(t *testing.T) {
	_, err := LoadConfig(env(map[string]string{
		"LOG_LEVEL":    "nope",
		"READ_TIMEOUT": "banana",
	}))
	if err == nil {
		t.Fatal("expected a joined error")
	}
	msg := err.Error()
	for _, want := range []string{"DATABASE_URL", "LOG_LEVEL", "READ_TIMEOUT"} {
		if !strings.Contains(msg, want) {
			t.Errorf("errors.Join did not report %s: %q", want, msg)
		}
	}
}

func TestLoadConfigAllOverrides(t *testing.T) {
	cfg, err := LoadConfig(env(map[string]string{
		"DATABASE_URL": "postgres://db/app",
		"ADDR":         "127.0.0.1:1234",
		"LOG_LEVEL":    "error",
		"READ_TIMEOUT": "45s",
	}))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.DatabaseURL != "postgres://db/app" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.Addr != "127.0.0.1:1234" {
		t.Errorf("Addr = %q", cfg.Addr)
	}
	if cfg.LogLevel != slog.LevelError {
		t.Errorf("LogLevel = %v, want Error", cfg.LogLevel)
	}
	if cfg.ReadTimeout != 45*time.Second {
		t.Errorf("ReadTimeout = %v, want 45s", cfg.ReadTimeout)
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := NewJSONLogger(&buf, slog.LevelWarn)

	l.Info("ignored", "k", "v")
	if buf.Len() != 0 {
		t.Fatalf("an Info record leaked through a Warn-level logger: %s", buf.String())
	}

	l.Warn("kept", "k", "v")
	if buf.Len() == 0 {
		t.Fatal("a Warn record was filtered out")
	}
}

func TestLogRequestFields(t *testing.T) {
	var buf bytes.Buffer
	l := NewJSONLogger(&buf, slog.LevelDebug)

	LogRequestStart(l, "GET", "/users", "req-1")
	LogRequestEnd(l, "GET", "/users", 200, 1500*time.Millisecond)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d log lines, want 2: %s", len(lines), buf.String())
	}

	var start map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &start); err != nil {
		t.Fatalf("the first line is not JSON: %v", err)
	}
	if start["msg"] != "request started" || start["method"] != "GET" ||
		start["path"] != "/users" || start["request_id"] != "req-1" {
		t.Fatalf("start line = %v", start)
	}

	var end map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &end); err != nil {
		t.Fatalf("the second line is not JSON: %v", err)
	}
	if end["msg"] != "request finished" || end["status"].(float64) != 200 {
		t.Fatalf("end line = %v", end)
	}
	if end["duration_ms"].(float64) != 1500 {
		t.Fatalf("duration_ms = %v, want 1500", end["duration_ms"])
	}
}

func TestLoggerLevelFilteringAllLevels(t *testing.T) {
	tests := []struct {
		name  string
		level slog.Level
		want  []string
	}{
		{"debug", slog.LevelDebug, []string{"DEBUG", "INFO", "WARN", "ERROR"}},
		{"info", slog.LevelInfo, []string{"INFO", "WARN", "ERROR"}},
		{"warn", slog.LevelWarn, []string{"WARN", "ERROR"}},
		{"error", slog.LevelError, []string{"ERROR"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := NewJSONLogger(&buf, tc.level)

			l.Debug("d")
			l.Info("i")
			l.Warn("w")
			l.Error("e")

			var got []string
			for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
				if line == "" {
					continue
				}
				var rec map[string]any
				if err := json.Unmarshal([]byte(line), &rec); err != nil {
					t.Fatalf("line is not JSON: %q: %v", line, err)
				}
				got = append(got, rec["level"].(string))
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("levels = %v, want %v (buffer: %s)", got, tc.want, buf.String())
			}
		})
	}
}

func TestHelpersRespectLoggerLevel(t *testing.T) {
	var buf bytes.Buffer
	l := NewJSONLogger(&buf, slog.LevelWarn)

	LogRequestStart(l, "GET", "/users", "req-1")
	LogRequestEnd(l, "GET", "/users", 200, time.Second)

	if buf.Len() != 0 {
		t.Fatalf("Info-level helpers leaked through a Warn-level logger: %s", buf.String())
	}
}

func TestLogRequestEndVariants(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		d       time.Duration
		wantMS  float64
	}{
		{"no-content", 204, 0, 0},
		{"sub-millisecond", 404, 250 * time.Microsecond, 0},
		{"server-error", 500, 3 * time.Second, 3000},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := NewJSONLogger(&buf, slog.LevelInfo)
			LogRequestEnd(l, "POST", "/items", tc.status, tc.d)

			line := strings.TrimSpace(buf.String())
			if line == "" {
				t.Fatal("LogRequestEnd wrote nothing")
			}
			if strings.Contains(line, "\n") {
				t.Fatalf("LogRequestEnd wrote more than one line: %q", line)
			}

			var rec map[string]any
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				t.Fatalf("line is not JSON: %q: %v", line, err)
			}
			if rec["msg"] != "request finished" {
				t.Errorf("msg = %v, want request finished", rec["msg"])
			}
			if rec["method"] != "POST" || rec["path"] != "/items" {
				t.Errorf("method/path = %v/%v, want POST//items", rec["method"], rec["path"])
			}
			if rec["status"].(float64) != float64(tc.status) {
				t.Errorf("status = %v, want %d", rec["status"], tc.status)
			}
			if rec["duration_ms"].(float64) != tc.wantMS {
				t.Errorf("duration_ms = %v, want %v", rec["duration_ms"], tc.wantMS)
			}
		})
	}
}

func TestLogRequestStartFields(t *testing.T) {
	var buf bytes.Buffer
	l := NewJSONLogger(&buf, slog.LevelInfo)
	LogRequestStart(l, "DELETE", "/items/7", "req-42")

	var rec map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &rec); err != nil {
		t.Fatalf("line is not JSON: %v", err)
	}
	if rec["msg"] != "request started" {
		t.Errorf("msg = %v, want request started", rec["msg"])
	}
	if rec["method"] != "DELETE" || rec["path"] != "/items/7" || rec["request_id"] != "req-42" {
		t.Errorf("record = %v", rec)
	}
	if _, ok := rec["status"]; ok {
		t.Errorf("a start record should not carry status: %v", rec)
	}
}
`,
				}},
				Hints: []string{
					"§slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})§ wrapped in §slog.New§.",
					"Attributes are alternating key/value arguments: §l.Info(\"request started\", \"method\", method, \"path\", path, \"request_id\", requestID)§.",
					"Pass §d.Milliseconds()§ so the value serialises as a number, not a string.",
					"Accumulate problems in a §[]error§ and finish with §errors.Join(errs...)§.",
					"Build the defaults first, then override from the environment — that keeps the function easy to read.",
				},
				Solution: `package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

type Config struct {
	Addr        string
	DatabaseURL string
	LogLevel    slog.Level
	ReadTimeout time.Duration
}

func LoadConfig(getenv func(string) string) (Config, error) {
	cfg := Config{
		Addr:        ":8080",
		DatabaseURL: getenv("DATABASE_URL"),
		LogLevel:    slog.LevelInfo,
		ReadTimeout: 10 * time.Second,
	}

	var errs []error

	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	if addr := getenv("ADDR"); addr != "" {
		cfg.Addr = addr
	}

	switch lvl := getenv("LOG_LEVEL"); lvl {
	case "", "info":
		cfg.LogLevel = slog.LevelInfo
	case "debug":
		cfg.LogLevel = slog.LevelDebug
	case "warn":
		cfg.LogLevel = slog.LevelWarn
	case "error":
		cfg.LogLevel = slog.LevelError
	default:
		errs = append(errs, fmt.Errorf("LOG_LEVEL %q is invalid (want debug, info, warn or error)", lvl))
	}

	if raw := getenv("READ_TIMEOUT"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			errs = append(errs, fmt.Errorf("READ_TIMEOUT %q: %w", raw, err))
		} else {
			cfg.ReadTimeout = d
		}
	}

	return cfg, errors.Join(errs...)
}

func NewJSONLogger(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
}

func LogRequestStart(l *slog.Logger, method, path, requestID string) {
	l.Info("request started",
		"method", method,
		"path", path,
		"request_id", requestID,
	)
}

func LogRequestEnd(l *slog.Logger, method, path string, status int, d time.Duration) {
	l.Info("request finished",
		"method", method,
		"path", path,
		"status", status,
		"duration_ms", d.Milliseconds(),
	)
}

func main() {
	cfg, err := LoadConfig(os.Getenv)
	fmt.Println(cfg, err)

	logger := NewJSONLogger(os.Stdout, slog.LevelInfo)
	LogRequestStart(logger, "GET", "/users", "req-1")
}
`,
			},
			{
				ID:    "m2t12e2",
				Title: "Fix the buggy statistics package",
				Prompt: `This exercise is about a skill, not a concept: **reading a failing test suite and
fixing the code it describes.**

The three functions below are wrong in four separate ways. Press **Test**, read the
failures carefully, and fix them until everything passes. This is exactly the loop you
will run a hundred times a week.

§§§go
func Mean(nums []float64) (float64, error)
func Median(nums []float64) (float64, error)
func StdDev(nums []float64) (float64, error)
§§§

Specification:

* All three return §ErrEmpty§ for an empty or nil input.
* §Median§ must sort a **copy** — the caller's slice must not be reordered — and must
  average the two middle values when the count is even.
* §StdDev§ is the *population* standard deviation: the square root of the mean of the
  squared deviations from the mean.
* §Mean§ is the arithmetic mean, as a float.

The seeded bugs are: integer division, an in-place sort, a missing square root, and not
handling the even-length case. The tests tell you which is which.
`,
				Starter: `package main

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

var ErrEmpty = errors.New("no data")

func Mean(nums []float64) (float64, error) {
	if len(nums) == 0 {
		return 0, ErrEmpty
	}
	sum := 0.0
	for _, v := range nums {
		sum += v
	}
	// BUG: integer division
	return float64(int(sum) / len(nums)), nil
}

func Median(nums []float64) (float64, error) {
	if len(nums) == 0 {
		return 0, ErrEmpty
	}
	// BUG: this mutates the caller's slice
	sort.Float64s(nums)
	n := len(nums)
	// BUG: does not average the two middle values for an even count
	return nums[n/2], nil
}

func StdDev(nums []float64) (float64, error) {
	if len(nums) == 0 {
		return 0, ErrEmpty
	}
	m, _ := Mean(nums)
	sum := 0.0
	for _, v := range nums {
		sum += (v - m) * (v - m)
	}
	// BUG: missing the square root
	return sum / float64(len(nums)), nil
}

func main() {
	fmt.Println(Mean([]float64{1, 2, 3}))
	_ = math.Sqrt
}
`,
				Tests: []File{{
					Name: "stats_test.go",
					Content: `package main

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func almost(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestEmpty(t *testing.T) {
	for name, fn := range map[string]func([]float64) (float64, error){
		"Mean":   Mean,
		"Median": Median,
		"StdDev": StdDev,
	} {
		if _, err := fn(nil); !errors.Is(err, ErrEmpty) {
			t.Errorf("%s(nil) error = %v, want ErrEmpty", name, err)
		}
	}
}

func TestMean(t *testing.T) {
	got, err := Mean([]float64{1, 2, 3, 4})
	if err != nil || !almost(got, 2.5) {
		t.Fatalf("Mean = %v, %v; want 2.5", got, err)
	}
	got, _ = Mean([]float64{0.5})
	if !almost(got, 0.5) {
		t.Fatalf("Mean(single) = %v", got)
	}
}

func TestMedianOdd(t *testing.T) {
	got, err := Median([]float64{5, 1, 3})
	if err != nil || !almost(got, 3) {
		t.Fatalf("Median = %v, %v; want 3", got, err)
	}
}

func TestMedianEven(t *testing.T) {
	got, err := Median([]float64{4, 1, 3, 2})
	if err != nil || !almost(got, 2.5) {
		t.Fatalf("Median = %v, %v; want 2.5", got, err)
	}
}

func TestMedianDoesNotMutate(t *testing.T) {
	in := []float64{5, 1, 3, 2, 4}
	before := append([]float64(nil), in...)

	if _, err := Median(in); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(in, before) {
		t.Fatalf("Median reordered the caller's slice: %v", in)
	}
}

func TestStdDev(t *testing.T) {
	// The population standard deviation of [2,4,4,4,5,5,7,9] is exactly 2.
	got, err := StdDev([]float64{2, 4, 4, 4, 5, 5, 7, 9})
	if err != nil || !almost(got, 2) {
		t.Fatalf("StdDev = %v, %v; want 2", got, err)
	}
	got, _ = StdDev([]float64{3, 3, 3})
	if !almost(got, 0) {
		t.Fatalf("StdDev of a constant = %v, want 0", got)
	}
}

func TestEmptySlice(t *testing.T) {
	for name, fn := range map[string]func([]float64) (float64, error){
		"Mean":   Mean,
		"Median": Median,
		"StdDev": StdDev,
	} {
		if _, err := fn([]float64{}); !errors.Is(err, ErrEmpty) {
			t.Errorf("%s(empty slice) error = %v, want ErrEmpty", name, err)
		}
	}
}

func TestMeanTable(t *testing.T) {
	tests := []struct {
		name string
		in   []float64
		want float64
	}{
		{"single", []float64{7}, 7},
		{"fraction", []float64{1, 2}, 1.5},
		{"negative fraction", []float64{-1, -2}, -1.5},
		{"negatives cancel", []float64{-5, 0, 5}, 0},
		{"zeros", []float64{0, 0, 0}, 0},
		{"decimals", []float64{0.1, 0.2, 0.3}, 0.2},
		{"unsorted", []float64{9, 1, 5}, 5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Mean(tc.in)
			if err != nil {
				t.Fatalf("Mean(%v) error = %v", tc.in, err)
			}
			if !almost(got, tc.want) {
				t.Fatalf("Mean(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestMedianTable(t *testing.T) {
	tests := []struct {
		name string
		in   []float64
		want float64
	}{
		{"single", []float64{42}, 42},
		{"two values", []float64{2, 1}, 1.5},
		{"even duplicates", []float64{1, 2, 2, 3}, 2},
		{"odd unsorted", []float64{9, 1, 5, 3, 7}, 5},
		{"negative odd", []float64{-5, -1, -3}, -3},
		{"negative even", []float64{-4, -1}, -2.5},
		{"decimal pair", []float64{0.5, 1.5}, 1},
		{"all equal", []float64{2, 2, 2, 2}, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Median(tc.in)
			if err != nil {
				t.Fatalf("Median(%v) error = %v", tc.in, err)
			}
			if !almost(got, tc.want) {
				t.Fatalf("Median(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestMedianDoesNotMutateEven(t *testing.T) {
	in := []float64{4, 1, 3, 2}
	before := append([]float64(nil), in...)

	if _, err := Median(in); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(in, before) {
		t.Fatalf("Median reordered the caller's even-length slice: %v", in)
	}
}

func TestStdDevTable(t *testing.T) {
	tests := []struct {
		name string
		in   []float64
		want float64
	}{
		// Population, not sample: dividing by n rather than n-1 matters here.
		{"two values", []float64{2, 4}, 1},
		{"four values", []float64{1, 2, 3, 4}, math.Sqrt(1.25)},
		{"single value", []float64{7}, 0},
		{"negative values", []float64{-1, 0, 1}, math.Sqrt(2.0 / 3.0)},
		{"decimals", []float64{0.5, 1.5}, 0.5},
		{"zeros", []float64{0, 0}, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := StdDev(tc.in)
			if err != nil {
				t.Fatalf("StdDev(%v) error = %v", tc.in, err)
			}
			if !almost(got, tc.want) {
				t.Fatalf("StdDev(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
`,
				}},
				Hints: []string{
					"§Mean§: §return sum / float64(len(nums)), nil§ — dividing two ints truncates.",
					"§Median§: §sorted := make([]float64, len(nums)); copy(sorted, nums); sort.Float64s(sorted)§.",
					"Even count: §(sorted[n/2-1] + sorted[n/2]) / 2§. Odd count: §sorted[n/2]§.",
					"§StdDev§ wraps the result in §math.Sqrt§.",
					"Press §Test§ after each fix — the failures narrow down as you go, which is the workflow this exercise is teaching.",
				},
				Solution: `package main

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

var ErrEmpty = errors.New("no data")

func Mean(nums []float64) (float64, error) {
	if len(nums) == 0 {
		return 0, ErrEmpty
	}
	sum := 0.0
	for _, v := range nums {
		sum += v
	}
	return sum / float64(len(nums)), nil
}

func Median(nums []float64) (float64, error) {
	if len(nums) == 0 {
		return 0, ErrEmpty
	}
	sorted := make([]float64, len(nums))
	copy(sorted, nums)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2], nil
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2, nil
}

func StdDev(nums []float64) (float64, error) {
	if len(nums) == 0 {
		return 0, ErrEmpty
	}
	m, err := Mean(nums)
	if err != nil {
		return 0, err
	}
	sum := 0.0
	for _, v := range nums {
		sum += (v - m) * (v - m)
	}
	return math.Sqrt(sum / float64(len(nums))), nil
}

func main() {
	fmt.Println(Mean([]float64{1, 2, 3}))
}
`,
			},
			{
				ID:    "m2t12e3",
				Title: "Graceful shutdown",
				Prompt: `Implement §RunServer§ so that cancelling the context stops the server cleanly and
returns §nil§.

§§§go
func RunServer(ctx context.Context, srv *http.Server, ln net.Listener) error
§§§

Requirements:

* Serve on the provided listener, from a goroutine.
* When §ctx§ is cancelled, call §srv.Shutdown§ with a **fresh** 5-second timeout context
  and return its result (§nil§ on a clean stop).
* A real listen error — anything except §http.ErrServerClosed§ — must be returned.
* No goroutine leaks: the function must return once shutdown has completed.

Taking the listener as a parameter is what makes this testable: the test uses
§net.Listen("tcp", "127.0.0.1:0")§ to get a free port, so nothing collides and nothing
waits.
`,
				Starter: `package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// RunServer serves on ln until ctx is cancelled, then shuts down gracefully.
func RunServer(ctx context.Context, srv *http.Server, ln net.Listener) error {
	// TODO
	return nil
}

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(100 * time.Millisecond); cancel() }()

	fmt.Println(RunServer(ctx, &http.Server{Handler: http.NewServeMux()}, ln))
	_ = errors.Is
}
`,
				Tests: []File{{
					Name: "shutdown_test.go",
					Content: `package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestRunServerServesThenShutsDown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- RunServer(ctx, &http.Server{Handler: mux}, ln)
	}()

	url := "http://" + ln.Addr().String() + "/ping"
	var body string
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			body = string(b)
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if body != "pong" {
		t.Fatalf("the server did not serve: body = %q", body)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunServer returned %v, want nil after a clean shutdown", err)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("RunServer did not return after the context was cancelled")
	}

	if _, err := http.Get(url); err == nil {
		t.Fatal("the server is still accepting connections after shutdown")
	}
}

func TestRunServerAlreadyCancelled(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan error, 1)
	go func() {
		done <- RunServer(ctx, &http.Server{Handler: http.NewServeMux()}, ln)
	}()

	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("RunServer = %v, want nil or context.Canceled", err)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("RunServer hung on an already-cancelled context")
	}
}

func TestRunServerContextDeadline(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- RunServer(ctx, &http.Server{Handler: http.NewServeMux()}, ln)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunServer = %v, want nil when the context deadline expires", err)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("RunServer did not return after the context deadline")
	}

	url := "http://" + ln.Addr().String() + "/"
	if resp, err := http.Get(url); err == nil {
		resp.Body.Close()
		t.Fatal("the server is still accepting connections after the deadline")
	}
}

func TestRunServerWaitsForInFlightRequest(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	started := make(chan struct{}, 1)
	release := make(chan struct{})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		started <- struct{}{}
		<-release
		fmt.Fprint(w, "finished")
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- RunServer(ctx, &http.Server{Handler: mux}, ln)
	}()

	type response struct {
		body string
		err  error
	}
	respCh := make(chan response, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/slow")
		if err != nil {
			respCh <- response{err: err}
			return
		}
		defer resp.Body.Close()
		b, readErr := io.ReadAll(resp.Body)
		respCh <- response{body: string(b), err: readErr}
	}()

	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("the handler never started")
	}

	cancel()
	// Let Shutdown begin while the handler is still blocked, so this really
	// tests draining rather than an already-finished request.
	time.Sleep(50 * time.Millisecond)
	close(release)

	select {
	case r := <-respCh:
		if r.err != nil {
			t.Fatalf("the in-flight request failed during shutdown: %v", r.err)
		}
		if r.body != "finished" {
			t.Fatalf("in-flight body = %q, want finished", r.body)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("the in-flight request never completed")
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunServer = %v, want nil after draining the request", err)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("RunServer did not return after the in-flight request finished")
	}
}

func TestRunServerReturnsListenError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	// Serve on a listener that is already closed: a real accept error, not a
	// graceful http.ErrServerClosed.
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- RunServer(context.Background(), &http.Server{Handler: http.NewServeMux()}, ln)
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("RunServer returned nil for a closed listener")
		}
	case <-time.After(6 * time.Second):
		t.Fatal("RunServer did not return after a listen error")
	}
}
`,
				}},
				Hints: []string{
					"Run §srv.Serve(ln)§ in a goroutine and send its error on a **buffered** channel (§make(chan error, 1)§), so the goroutine never blocks.",
					"§select§ on §ctx.Done()§ and the error channel.",
					"Filter §http.ErrServerClosed§ — §Serve§ returns it after §Shutdown§.",
					"Give §Shutdown§ its own context: §sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)§.",
					"After §Shutdown§, drain the error channel so the §Serve§ goroutine cannot leak.",
				},
				Solution: `package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

func RunServer(ctx context.Context, srv *http.Server, ln net.Listener) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	select {
	case <-ctx.Done():
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(sctx); err != nil {
			return err
		}
		// Drain the Serve goroutine so it cannot leak.
		select {
		case err := <-errCh:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
		case <-time.After(time.Second):
		}
		return nil

	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(100 * time.Millisecond); cancel() }()

	fmt.Println(RunServer(ctx, &http.Server{Handler: http.NewServeMux()}, ln))
}
`,
			},
		},
	}
}
