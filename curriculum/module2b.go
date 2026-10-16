package curriculum

func m2t5() Topic {
	return Topic{
		ID:      "m2t5",
		Title:   "JSON and struct tags",
		Summary: "Turning structs into JSON and back, and controlling the field names.",
		Theory: `
Almost every HTTP API speaks JSON. In Go that means the §encoding/json§ package, which
converts between Go values and JSON text using **reflection** — it looks at your struct
at run time and works out the field names.

## Marshaling: struct → JSON

*To marshal* means to convert a Go value **into** JSON text.

§§§go
package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	ID    int
	Name  string
	Email string
}

func main() {
	u := User{ID: 1, Name: "Ada", Email: "ada@x.io"}

	data, err := json.Marshal(u)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
§§§

Output:

§§§text
{"ID":1,"Name":"Ada","Email":"ada@x.io"}
§§§

Two things to notice:

1. §json.Marshal§ returns §([]byte, error)§ — you must check the error and convert to a
   §string§ to print it.
2. **Field names are copied exactly, capital letters and all.** JSON APIs do not
   normally use capitals, which is what tags are for.

Only **exported** fields (capital letter) are included. An unexported field is silently
skipped — a trap when you wonder why a field "disappears".

## Struct tags: telling json what to call each field

A **struct tag** is a little string in backticks after a field, read by packages like
§encoding/json§:

§§§go
type User struct {
	ID        int    §json:"id"§
	Name      string §json:"name"§
	Email     string §json:"email"§
	Password  string §json:"-"§
	CreatedAt string §json:"created_at"§
}
§§§

| Tag | Effect |
| --- | --- |
| §json:"id"§ | the JSON key becomes §id§ |
| §json:"name,omitempty"§ | the key is left out when the value is the zero value |
| §json:"-"§ | this field is never included |
| no tag | the Go field name is used as-is |

The tag's shape is §json:"NAME,OPTION"§ — the name, then a comma, then options. It is
just a string; forgetting the quotes or the comma is a silent no-op, not an error.

Output with the tags above:

§§§text
{"id":1,"name":"Ada","email":"ada@x.io","created_at":"2024-01-01"}
§§§

The §password§ field is gone entirely.

### omitempty

§,omitempty§ drops the field when it holds a zero value:

§§§go
type Payload struct {
	A int    §json:"a,omitempty"§
	B string §json:"b,omitempty"§
	C []int  §json:"c"§
}

data, _ := json.Marshal(Payload{})
fmt.Println(string(data))
§§§

Output:

§§§text
{"c":null}
§§§

§A§ and §B§ were zero and vanished. §C§ has no §omitempty§, so it appears as §null§.

> **The omitempty trap:** §0§, §""§, §false§ and §nil§ are all "empty". For a PATCH
> endpoint where §""§ is a *meaningful* value ("clear the name"), §omitempty§ cannot
> tell "not provided" from "provided as empty". Use a §*string§ / §*int§ field, which
> is §nil§ when absent and points at the value when present.

## Unmarshaling: JSON → struct

*To unmarshal* means to go the other way.

§§§go
var u User
err := json.Unmarshal([]byte(payload), &u)
§§§

The §&§ is essential: §Unmarshal§ needs to know where to write. Forgetting it is a
compile error (§cannot use u as *User§), which is a friendly failure for once.

Notes:

* Unknown JSON keys are **ignored by default**. Typos in client requests disappear
  silently.
* A value of the wrong type is an **error**: sending §"age": "42"§ into an §int§ fails.
* Missing keys leave the field at its zero value.

## Decoding from a body, strictly

In an HTTP handler you read from §r.Body§, so use the streaming §Decoder§ — it saves
copying the whole body into memory first:

§§§go
var req CreateUserRequest

dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields() // turn silent typos into a 400
if err := dec.Decode(&req); err != nil {
	http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
	return
}
§§§

§DisallowUnknownFields()§ is worth enabling on every request struct: it turns
§"emial":"a@b.c"§ from "the field is mysteriously empty" into a clear error.

## Writing a response

§§§go
w.Header().Set("Content-Type", "application/json; charset=utf-8")
w.WriteHeader(http.StatusCreated)
if err := json.NewEncoder(w).Encode(user); err != nil {
	log.Printf("write response: %v", err)
}
§§§

§NewEncoder(w).Encode(v)§ writes straight to the response, and appends a newline (which
is legal JSON and friendly to §curl§). The §Content-Type§ header must be set *before*
§WriteHeader§ — remember the ordering rule from the earlier HTTP topic.

## The §any§ gotcha: numbers become float64

Decoding into §map[string]any§ or §[]any§ gives you §float64§ for every number,
regardless of what was in the text:

§§§go
var m map[string]any
json.Unmarshal([]byte(§{"id": 1}§), &m)
fmt.Printf("%T\n", m["id"]) // float64
§§§

For large IDs (Snowflake IDs, database keys above 2^53) that loses precision. Either
decode into a typed struct, or ask the decoder for §json.Number§:

§§§go
dec := json.NewDecoder(r.Body)
dec.UseNumber()
§§§

## Custom marshalers

Implement §json.Marshaler§ to control the output:

§§§go
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	switch str {
	case "active":
		*s = StatusActive
	case "banned":
		*s = StatusBanned
	default:
		return fmt.Errorf("unknown status %q", str)
	}
	return nil
}
§§§

That turns §Status(0)§ into §"active"§ on the wire, and rejects anything unexpected on
the way in.

### The §type alias§ trick

If you want to *add* a field rather than replace the whole object:

§§§go
func (u User) MarshalJSON() ([]byte, error) {
	type alias User // a new type with the same fields but NO methods
	return json.Marshal(struct {
		alias
		FullName string §json:"full_name"§
	}{alias(u), u.First + " " + u.Last})
}
§§§

§type alias User§ creates a distinct type whose method set is empty, so calling
§json.Marshal§ on it does **not** call §MarshalJSON§ again. Without that trick you get
infinite recursion and a stack overflow.

## A complete example

§§§go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type CreateUserRequest struct {
	Name string §json:"name"§
	Age  *int   §json:"age,omitempty"§
}

type User struct {
	ID       int    §json:"id"§
	Name     string §json:"name"§
	Email    string §json:"email"§
	Password string §json:"-"§
}

func main() {
	u := User{ID: 1, Name: "Ada", Email: "ada@x.io", Password: "hunter2"}
	data, _ := json.Marshal(u)
	fmt.Println(string(data))

	var req CreateUserRequest
	dec := json.NewDecoder(bytes.NewBufferString(§{"name":"Bo","age":30,"typo":1}§))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		fmt.Println("rejected:", err)
	}
}
§§§

Output:

§§§text
{"id":1,"name":"Ada","email":"ada@x.io"}
rejected: json: unknown field "typo"
§§§
`,
		Quiz: []Question{
			{
				ID:     "m2t5q1",
				Prompt: "What is the JSON output?",
				Code: `type Payload struct {
	A int    §json:"a"§
	B string §json:"b,omitempty"§
	C bool   §json:"c,omitempty"§
	D []int  §json:"d"§
	E string §json:"-"§
	F string
}

b, _ := json.Marshal(Payload{A: 1, D: nil, E: "secret", F: "visible"})
fmt.Println(string(b))`,
				Options: []Option{
					{Text: "§{\"a\":1,\"d\":null,\"F\":\"visible\"}§", Correct: true, Rationale: "§B§ and §C§ are omitted as zero values, §E§ is hidden by §-§, and §F§ has no tag so it keeps its Go name."},
					{Text: "§{\"a\":1,\"b\":\"\",\"c\":false,\"d\":null,\"F\":\"visible\"}§", Correct: false, Rationale: "§omitempty§ removed §B§ and §C§."},
					{Text: "§{\"a\":1,\"d\":[],\"F\":\"visible\"}§", Correct: false, Rationale: "A nil slice marshals to §null§, not §[]§. An empty but non-nil slice gives §[]§."},
					{Text: "§{\"a\":1,\"d\":null,\"E\":\"secret\",\"F\":\"visible\"}§", Correct: false, Rationale: "§E§ has the §-§ tag."},
				},
				Explanation: "§omitempty§ drops zero values, §-§ hides a field, and a field with no tag keeps its exact Go name.",
			},
			{
				ID:     "m2t5q2",
				Prompt: "What happens when you unmarshal §{\"age\": \"42\"}§ into §struct{ Age int §json:\"age\"§ }§?",
				Options: []Option{
					{Text: "§Decode§ returns an error: cannot unmarshal string into Go struct field of type int.", Correct: true, Rationale: "Go is strict about JSON value types unless the target implements §UnmarshalJSON§."},
					{Text: "It succeeds and Age becomes 42.", Correct: false, Rationale: "There is no implicit string-to-number coercion."},
					{Text: "It succeeds and Age stays 0, ignoring the field.", Correct: false, Rationale: "A type mismatch is an error, not a silent skip. Only *unknown field names* are skipped by default."},
					{Text: "It panics.", Correct: false, Rationale: "§Decode§ returns an error value."},
				},
				Explanation: "Use the §,string§ tag option (§json:\"age,string\"§) or a custom unmarshaler when the client sends numbers as strings.",
			},
			{
				ID:     "m2t5q3",
				Prompt: "Why is §DisallowUnknownFields()§ worth enabling on request structs?",
				Options: []Option{
					{Text: "It turns a misspelled key like §\"emial\"§ into a 400 error instead of a silently missing field.", Correct: true, Rationale: "Strict decoding surfaces client mistakes immediately."},
					{Text: "It makes decoding faster.", Correct: false, Rationale: "It adds a lookup; the benefit is correctness."},
					{Text: "It is required for nested structs.", Correct: false, Rationale: "Nested decoding works either way."},
					{Text: "It validates the *values* of the fields.", Correct: false, Rationale: "It only checks that the key names are known."},
				},
				Explanation: "Strict input at the API boundary prevents a whole class of \"why is this field empty?\" support tickets.",
			},
			{
				ID:     "m2t5q4",
				Prompt: "Why does a custom §MarshalJSON§ usually start with §type alias User§?",
				Options: []Option{
					{Text: "Because a defined type does not inherit methods, so the copy cannot call §MarshalJSON§ on itself and recurse forever.", Correct: true, Rationale: "§alias§ has the same fields but an empty method set."},
					{Text: "Because unexported fields need an alias.", Correct: false, Rationale: "The alias has exactly the same fields."},
					{Text: "Because §json.Marshal§ cannot take a pointer.", Correct: false, Rationale: "It accepts values and pointers."},
					{Text: "It is purely documentation.", Correct: false, Rationale: "Without it you get infinite recursion and a stack overflow."},
				},
				Explanation: "The alias trick is the standard way to escape your own custom marshaler.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m2t5e1",
				Title: "Warm-up: control the wire format with tags",
				Prompt: `Define the types and three helpers.

§§§go
type CreateUserRequest struct {
	Name  string   §json:"name"§
	Email string   §json:"email"§
	Age   *int     §json:"age,omitempty"§
	Tags  []string §json:"tags,omitempty"§
}

type User struct {
	ID        int    §json:"id"§
	Name      string §json:"name"§
	Email     string §json:"email"§
	Password  string §json:"-"§
	CreatedAt string §json:"created_at"§
}

func ParseCreateUser(body []byte) (CreateUserRequest, error)
func MustJSON(v any) ([]byte, error)
func Summarize(v any) (map[string]any, error)
§§§

* §MustJSON§ is a thin wrapper around §json.Marshal§.
* §Summarize§ marshals then unmarshals into a §map[string]any§, proving what the tags
  produced.
* §ParseCreateUser§ decodes **strictly** (§DisallowUnknownFields§), then validates:
  a blank §Name§ or an §Email§ without an §@§ is an error. The body arrives as a §[]byte§,
  so wrap it with §bytes.NewReader(body)§ first: §json.NewDecoder§ wants an §io.Reader§,
  and §DisallowUnknownFields§ exists only on §Decoder§, not on §json.Unmarshal§.

The hidden test asserts the exact JSON for a §User§ — including that the password never
appears — and that a body with an unknown field is rejected.
`,
				Starter: `package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type CreateUserRequest struct {
	Name  string   §json:"name"§
	Email string   §json:"email"§
	Age   *int     §json:"age,omitempty"§
	Tags  []string §json:"tags,omitempty"§
}

type User struct {
	ID        int    §json:"id"§
	Name      string §json:"name"§
	Email     string §json:"email"§
	Password  string §json:"-"§
	CreatedAt string §json:"created_at"§
}

// ParseCreateUser strictly decodes and validates a create-user body.
func ParseCreateUser(body []byte) (CreateUserRequest, error) {
	// TODO
	return CreateUserRequest{}, nil
}

// MustJSON returns v as compact JSON.
func MustJSON(v any) ([]byte, error) {
	// TODO
	return nil, nil
}

// Summarize decodes v's JSON into a map, showing which keys the tags produced.
func Summarize(v any) (map[string]any, error) {
	// TODO
	return nil, nil
}

func main() {
	u := User{ID: 1, Name: "Ada", Email: "ada@x.io", Password: "s3cret", CreatedAt: "2024-01-01"}
	b, _ := MustJSON(u)
	fmt.Println(string(b))

	req, err := ParseCreateUser([]byte(§{"name":"Bo","email":"bo@x.io"}§))
	fmt.Println(req, err)

	_, err = ParseCreateUser([]byte(§{"name":"Bo","email":"bo@x.io","nope":1}§))
	fmt.Println(err)

	_ = bytes.NewReader
	_ = strings.Contains
	_ = errors.Join
	_ = json.Marshal
}
`,
				Tests: []File{{
					Name: "json_test.go",
					Content: `package main

import (
	"strings"
	"testing"
)

func TestMustJSONOmitsSecrets(t *testing.T) {
	u := User{ID: 1, Name: "Ada", Email: "ada@x.io", Password: "s3cret", CreatedAt: "2024-01-01"}
	b, err := MustJSON(u)
	if err != nil {
		t.Fatalf("MustJSON: %v", err)
	}
	got := string(b)
	want := "{\"id\":1,\"name\":\"Ada\",\"email\":\"ada@x.io\",\"created_at\":\"2024-01-01\"}"
	if got != want {
		t.Fatalf("JSON = %s, want %s", got, want)
	}
	if strings.Contains(got, "s3cret") {
		t.Fatal("the password leaked into JSON")
	}
}

func TestParseCreateUserValid(t *testing.T) {
	req, err := ParseCreateUser([]byte("{\"name\":\"Bo\",\"email\":\"bo@x.io\",\"age\":30,\"tags\":[\"a\"]}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Name != "Bo" || req.Email != "bo@x.io" {
		t.Fatalf("parsed %+v", req)
	}
	if req.Age == nil || *req.Age != 30 {
		t.Fatalf("Age = %v, want 30", req.Age)
	}
	if len(req.Tags) != 1 {
		t.Fatalf("Tags = %v", req.Tags)
	}
}

func TestParseCreateUserOptionalAge(t *testing.T) {
	req, err := ParseCreateUser([]byte("{\"name\":\"Bo\",\"email\":\"bo@x.io\"}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Age != nil {
		t.Fatalf("Age should be nil when the key is absent, got %v", *req.Age)
	}

	zero := 0
	req, err = ParseCreateUser([]byte("{\"name\":\"Bo\",\"email\":\"bo@x.io\",\"age\":0}"))
	_ = zero
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Age == nil || *req.Age != 0 {
		t.Fatal("a pointer field must distinguish \"0\" from \"absent\"")
	}
}

func TestParseCreateUserRejectsUnknownField(t *testing.T) {
	if _, err := ParseCreateUser([]byte("{\"name\":\"Bo\",\"email\":\"bo@x.io\",\"nope\":1}")); err == nil {
		t.Fatal("an unknown field was accepted")
	}
}

func TestParseCreateUserValidation(t *testing.T) {
	if _, err := ParseCreateUser([]byte("{\"email\":\"bo@x.io\"}")); err == nil {
		t.Fatal("an empty name was accepted")
	}
	if _, err := ParseCreateUser([]byte("{\"name\":\"  \",\"email\":\"bo@x.io\"}")); err == nil {
		t.Fatal("a whitespace-only name was accepted")
	}
	if _, err := ParseCreateUser([]byte("{\"name\":\"Bo\",\"email\":\"nope\"}")); err == nil {
		t.Fatal("an email without @ was accepted")
	}
	if _, err := ParseCreateUser([]byte("not json")); err == nil {
		t.Fatal("malformed JSON was accepted")
	}
}

func TestSummarize(t *testing.T) {
	u := User{ID: 2, Name: "Cy", Email: "cy@x.io", Password: "x", CreatedAt: "2024-05-05"}
	m, err := Summarize(u)
	if err != nil {
		t.Fatalf("Summarize: %v", err)
	}
	if _, ok := m["created_at"]; !ok {
		t.Fatalf("Summarize keys = %v, want created_at", m)
	}
	if _, ok := m["password"]; ok {
		t.Fatal("the hidden field should not appear")
	}
	if m["id"].(float64) != 2 {
		t.Fatalf("id = %v", m["id"])
	}
}
`,
				}},
				Hints: []string{
					"§MustJSON§: §return json.Marshal(v)§.",
					"§ParseCreateUser§: §dec := json.NewDecoder(bytes.NewReader(body)); dec.DisallowUnknownFields(); dec.Decode(&req)§.",
					"Validation: §strings.TrimSpace(req.Name) == \"\"§ and §!strings.Contains(req.Email, \"@\")§.",
					"Collect the validation problems with §errors.Join§, or return the first one — either passes.",
					"§Summarize§: marshal, then §json.Unmarshal(raw, &m)§ into a §map[string]any§.",
				},
				Solution: `package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type CreateUserRequest struct {
	Name  string   §json:"name"§
	Email string   §json:"email"§
	Age   *int     §json:"age,omitempty"§
	Tags  []string §json:"tags,omitempty"§
}

type User struct {
	ID        int    §json:"id"§
	Name      string §json:"name"§
	Email     string §json:"email"§
	Password  string §json:"-"§
	CreatedAt string §json:"created_at"§
}

func ParseCreateUser(body []byte) (CreateUserRequest, error) {
	var req CreateUserRequest
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return CreateUserRequest{}, fmt.Errorf("decode: %w", err)
	}

	var errs []error
	if strings.TrimSpace(req.Name) == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if !strings.Contains(req.Email, "@") {
		errs = append(errs, errors.New("email is invalid"))
	}
	if err := errors.Join(errs...); err != nil {
		return CreateUserRequest{}, err
	}
	return req, nil
}

func MustJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

func Summarize(v any) (map[string]any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func main() {
	u := User{ID: 1, Name: "Ada", Email: "ada@x.io", Password: "s3cret", CreatedAt: "2024-01-01"}
	b, _ := MustJSON(u)
	fmt.Println(string(b))

	req, err := ParseCreateUser([]byte(§{"name":"Bo","email":"bo@x.io"}§))
	fmt.Println(req, err)

	_, err = ParseCreateUser([]byte(§{"name":"Bo","email":"bo@x.io","nope":1}§))
	fmt.Println(err)
}
`,
			},
			{
				ID:    "m2t5e2",
				Title: "Encode to a response, decode from a request",
				Prompt: `These two operations appear inside every HTTP handler; practise them on their own.

§§§go
func WriteJSON(w io.Writer, v any) error
func ReadJSON(r io.Reader, v any) error
§§§

* §WriteJSON§ uses §json.NewEncoder(w).Encode(v)§ — streaming straight to the writer, no
  intermediate byte slice.
* §ReadJSON§ uses a §json.Decoder§ with §DisallowUnknownFields()§ so typos are rejected.
* Both must return the underlying error so the caller can decide the status code: a
  decode failure is a client error (400), an encode failure is a server error (500).
* §ReadJSON§ must reject a body that is empty, malformed, or has extra keys.

The test drives both with a §bytes.Buffer§, which is exactly what §httptest§ gives you
for a response body.
`,
				Starter: `package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// WriteJSON encodes v as JSON to w.
func WriteJSON(w io.Writer, v any) error {
	// TODO
	return nil
}

// ReadJSON strictly decodes JSON from r into v.
func ReadJSON(r io.Reader, v any) error {
	// TODO
	return nil
}

func main() {
	var buf bytes.Buffer
	_ = WriteJSON(&buf, map[string]int{"a": 1})
	fmt.Print(buf.String())

	var out struct {
		A int §json:"a"§
	}
	err := ReadJSON(bytes.NewReader([]byte(§{"a":2}§)), &out)
	fmt.Println(out.A, err)
	_ = json.Marshal
}
`,
				Tests: []File{{
					Name: "encdec_test.go",
					Content: `package main

import (
	"bytes"
	"strings"
	"testing"
)

type payload struct {
	Name string §json:"name"§
	Age  int    §json:"age"§
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, payload{Name: "Ada", Age: 36}); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	want := "{\"name\":\"Ada\",\"age\":36}"
	if got != want {
		t.Fatalf("WriteJSON produced %q, want %q", got, want)
	}
}

func TestReadJSON(t *testing.T) {
	var p payload
	if err := ReadJSON(strings.NewReader("{\"name\":\"Ada\",\"age\":36}"), &p); err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	if p.Name != "Ada" || p.Age != 36 {
		t.Fatalf("decoded %+v", p)
	}
}

func TestReadJSONRejectsUnknownFields(t *testing.T) {
	var p payload
	if err := ReadJSON(strings.NewReader("{\"name\":\"Ada\",\"nope\":1}"), &p); err == nil {
		t.Fatal("an unknown field was accepted")
	}
}

func TestReadJSONRejectsBadInput(t *testing.T) {
	var p payload
	for name, body := range map[string]string{
		"empty":     "",
		"malformed": "{oops",
		"wrong type": "{\"age\":\"not a number\"}",
	} {
		if err := ReadJSON(strings.NewReader(body), &p); err == nil {
			t.Errorf("%s: ReadJSON accepted the input", name)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	in := payload{Name: "Bo", Age: 10}
	if err := WriteJSON(&buf, in); err != nil {
		t.Fatal(err)
	}
	var out payload
	if err := ReadJSON(&buf, &out); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Fatalf("round trip gave %+v, want %+v", out, in)
	}
}
`,
				}},
				Hints: []string{
					"§WriteJSON§ is §return json.NewEncoder(w).Encode(v)§.",
					"§ReadJSON§: §dec := json.NewDecoder(r); dec.DisallowUnknownFields(); return dec.Decode(v)§.",
					"An empty body makes §Decode§ return §io.EOF§ — that counts as an error, which is what the test wants.",
					"The writer parameter is §io.Writer§, not §http.ResponseWriter§, so the same function works for a buffer in a test and for the network in a handler.",
				},
				Solution: `package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func WriteJSON(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

func ReadJSON(r io.Reader, v any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func main() {
	var buf bytes.Buffer
	_ = WriteJSON(&buf, map[string]int{"a": 1})
	fmt.Print(buf.String())

	var out struct {
		A int §json:"a"§
	}
	err := ReadJSON(bytes.NewReader([]byte(§{"a":2}§)), &out)
	fmt.Println(out.A, err)
	_ = json.Marshal
}
`,
			},
		},
	}
}

func m2t6() Topic {
	return Topic{
		ID:      "m2t6",
		Title:   "HTTP handlers and net/http routing",
		Summary: "Serving a request with nothing but the standard library.",
		Theory: `
Go ships a production HTTP server in the standard library. There is no framework to
install — you import §net/http§ and you have one.

## A handler is anything with ServeHTTP

The entire abstraction is one interface:

§§§go
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
§§§

A **handler** is anything that can produce a response for a request. §http.HandlerFunc§
adapts an ordinary function so it satisfies that interface:

§§§go
func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello, %s", r.URL.Path)
}

http.Handle("/hello", http.HandlerFunc(hello))
§§§

## The two arguments

§w http.ResponseWriter§ is where you write the response.

| You write | Result |
| --- | --- |
| §w.Header().Set("Content-Type", …)§ | a response header (must come **before** the status) |
| §w.WriteHeader(http.StatusCreated)§ | the status code |
| §w.Write([]byte("hi"))§ | the body |
| §http.Error(w, "nope", 400)§ | status + plain-text body in one call |

The ordering rule is the single most common HTTP bug in Go:

> **Headers → status → body.** The first write freezes the status at 200. Call
> §WriteHeader§ afterwards and you get a log line saying §superfluous WriteHeader call§
> and the wrong status in the response.

§r *http.Request§ is the incoming request.

| Expression | Gives you |
| --- | --- |
| §r.Method§ | §"GET"§, §"POST"§, … |
| §r.URL.Path§ | the decoded path |
| §r.URL.Query().Get("page")§ | a query parameter |
| §r.Header.Get("Authorization")§ | a header |
| §r.PathValue("id")§ | a URL wildcard (Go 1.22+) |
| §r.Context()§ | the request's context — cancelled when the client disconnects |
| §r.Body§ | the request body (an §io.ReadCloser§) |

## The Go 1.22 ServeMux

Before Go 1.22 the standard mux only understood paths. Since 1.22 it understands
**methods** and **wildcards**, which removes the main reason people reached for a third
-party router:

§§§go
mux := http.NewServeMux()

mux.HandleFunc("GET /users/{id}", getUser)      // {id} matches one segment
mux.HandleFunc("POST /users", createUser)
mux.HandleFunc("GET /files/{path...}", serveFile) // {path...} matches the rest
mux.HandleFunc("GET /{$}", homePage)             // {$} anchors: matches ONLY "/"

func getUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Fprintf(w, "user %s", id)
}
§§§

Pattern rules:

* §/users/{id}§ matches exactly one path segment: §/users/7§ yes, §/users/7/posts§ no.
* §/users/{path...}§ matches the remainder, including slashes.
* A pattern ending in §/§ is a **subtree**: §/api/§ matches §/api/§ and everything below.
* **The most specific pattern wins.** §GET /users/me§ beats §GET /users/{id}§, so a
  literal beats a wildcard, and a wildcard beats a subtree.

A method-less pattern (§"/users"§) matches every method. Use methods — a §POST§ handler
that also answers §GET§ is a surprise nobody wants.

## Reading the body safely

§§§go
func createUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Bound how much we will read: a client could stream gigabytes at us.
	body := http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB

	var req CreateUserRequest
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	...
}
§§§

§http.MaxBytesReader§ is a one-line defence against a request that never ends.

## Writing JSON back

§§§go
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
§§§

Those two helpers are 90% of most APIs' response code, and they guarantee consistent
headers and error shapes.

## Mapping errors to status codes

One §switch§ per handler is the standard pattern:

§§§go
switch {
case errors.Is(err, ErrNotFound):
	writeError(w, http.StatusNotFound, "not found")
case errors.Is(err, ErrBadInput):
	writeError(w, http.StatusBadRequest, err.Error())
case errors.Is(err, context.DeadlineExceeded):
	writeError(w, http.StatusGatewayTimeout, "upstream timed out")
default:
	log.Printf("internal error: %v", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}
§§§

Never send a raw internal error string to a client: it leaks file paths, SQL and
occasionally credentials. Log the detail, return a generic message.

## The server itself

§§§go
srv := &http.Server{
	Addr:              ":8080",
	Handler:           mux,
	ReadHeaderTimeout: 5 * time.Second,
	ReadTimeout:       15 * time.Second,
	WriteTimeout:      15 * time.Second,
	IdleTimeout:       60 * time.Second,
}
log.Fatal(srv.ListenAndServe())
§§§

Those timeouts matter. The default zero means "no limit", and a single slow client can
hold a connection open forever.

## A complete example

§§§go
package main

import (
	"encoding/json"
	"net/http"
)

type Item struct {
	ID   int    §json:"id"§
	Name string §json:"name"§
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		writeJSON(w, http.StatusOK, Item{ID: 1, Name: "item " + id})
	})

	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		var it Item
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&it); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if it.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}
		writeJSON(w, http.StatusCreated, it)
	})

	http.ListenAndServe(":8080", mux)
}
§§§

Testing a handler never needs a real network — the next exercises use
§httptest.NewRecorder()§, which captures the response in memory:

§§§go
req := httptest.NewRequest(http.MethodGet, "/items/7", nil)
rec := httptest.NewRecorder()
mux.ServeHTTP(rec, req)

fmt.Println(rec.Code)        // 200
fmt.Println(rec.Body.String()) // {"id":1,"name":"item 7"}
§§§

## Status codes

| Operation | Method + path | Success |
| --- | --- | --- |
| list | §GET /items§ | 200 |
| read | §GET /items/{id}§ | 200, or 404 |
| create | §POST /items§ | 201 |
| replace | §PUT /items/{id}§ | 200 or 204 |
| partial update | §PATCH /items/{id}§ | 200 |
| delete | §DELETE /items/{id}§ | 204 |
`,
		Quiz: []Question{
			{
				ID:     "m2t6q1",
				Prompt: "What is wrong with this handler?",
				Code: `func handler(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "boom", http.StatusInternalServerError)
		return
	}
	w.Write(data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}`,
				Options: []Option{
					{Text: "The header and status are set *after* the body was written, so §w.Write(data)§ already sent 200 and the header is ignored.", Correct: true, Rationale: "The first write freezes the status and headers. Order must be: headers, status, body."},
					{Text: "§json.Marshal§ returns a §string§, not §[]byte§.", Correct: false, Rationale: "It returns §[]byte§."},
					{Text: "§http.Error§ cannot be used inside a handler.", Correct: false, Rationale: "It is the standard helper."},
					{Text: "Nothing — the response has status 201 and a JSON content type.", Correct: false, Rationale: "The status has already been sent as 200."},
				},
				Explanation: "Ordering is the whole game: §Header().Set§, then §WriteHeader§, then §Write§.",
			},
			{
				ID:     "m2t6q2",
				Prompt: "Which Go 1.22 §ServeMux§ patterns match §GET /api/users/42§? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "§GET /api/users/{id}§", Correct: true, Rationale: "§{id}§ matches exactly one non-empty segment, readable with §r.PathValue(\"id\")§."},
					{Text: "§GET /api/users/{path...}§", Correct: true, Rationale: "§{path...}§ matches the remainder, which here is just §42§."},
					{Text: "§/api/§", Correct: true, Rationale: "A trailing slash makes it a subtree pattern matching everything below §/api/§."},
					{Text: "§GET /api/users§", Correct: false, Rationale: "Without a trailing slash this is an exact match for §/api/users§, so §/42§ does not match."},
					{Text: "§POST /api/users/{id}§", Correct: false, Rationale: "The method has to match too."},
				},
				Explanation: "Literal segments beat wildcards, wildcards beat subtrees, and the method is part of the pattern.",
			},
			{
				ID:     "m2t6q3",
				Prompt: "Why does the handler example wrap the body in §http.MaxBytesReader(w, r.Body, 1<<20)§?",
				Options: []Option{
					{Text: "To cap how many bytes will be read, so a client cannot exhaust the server's memory with a huge or endless body.", Correct: true, Rationale: "Without a limit, §Decode§ will happily read until the process dies."},
					{Text: "To make JSON decoding faster.", Correct: false, Rationale: "It adds a counting wrapper; the benefit is protection."},
					{Text: "To close the body automatically.", Correct: false, Rationale: "You still §defer r.Body.Close()§."},
					{Text: "To reject malformed JSON.", Correct: false, Rationale: "Strict decoding and validation do that."},
				},
				Explanation: "Every public endpoint should bound the request body. §1<<20§ is 1 MiB.",
			},
			{
				ID:     "m2t6q4",
				Prompt: "What does §r.Context()§ give you that a fresh §context.Background()§ does not?",
				Options: []Option{
					{Text: "A context that is cancelled when the client disconnects or the request times out, so downstream work stops.", Correct: true, Rationale: "It ties the whole request chain to the client's lifetime."},
					{Text: "Access to the request headers.", Correct: false, Rationale: "Those are on §r.Header§."},
					{Text: "A longer timeout.", Correct: false, Rationale: "It gives a shorter, request-scoped lifetime."},
					{Text: "Automatic JSON decoding.", Correct: false, Rationale: "Unrelated."},
				},
				Explanation: "§r.Context()§ is the parent you should hand to every database and HTTP call the handler makes.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m2t6e1",
				Title: "Warm-up: a handler you can test without a network",
				Prompt: `Build a small API using only the standard library — no router package.

§§§go
func NewHelloHandler() http.Handler
§§§

The returned mux must answer:

| Request | Response |
| --- | --- |
| §GET /hello/{name}§ | 200, body §"hello, <name>!"§ |
| §GET /hello/{name}?shout=true§ | 200, body uppercased: §"HELLO, ADA!"§ |
| §GET /hello/§ (empty name) | 404 |
| §POST /hello/ada§ | 405, because only GET is registered |

Use §http.NewServeMux§ with Go 1.22 method+wildcard patterns, and §r.PathValue("name")§
to read the segment. §strings.ToUpper§ handles the shout.

The test calls the handler directly with §httptest.NewRequest§ and
§httptest.NewRecorder§ — no sockets, no ports, no waiting.
`,
				Starter: `package main

import (
	"fmt"
	"net/http"
	"strings"
)

// NewHelloHandler returns an http.Handler serving /hello/{name}.
func NewHelloHandler() http.Handler {
	mux := http.NewServeMux()
	// TODO: register the route
	return mux
}

func main() {
	fmt.Println(NewHelloHandler() != nil, strings.ToUpper("x"))
}
`,
				Tests: []File{{
					Name: "hello_test.go",
					Content: `package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func do(t *testing.T, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	NewHelloHandler().ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

func TestHello(t *testing.T) {
	rec := do(t, http.MethodGet, "/hello/ada")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != "hello, ada!" {
		t.Fatalf("body = %q, want %q", got, "hello, ada!")
	}
}

func TestShout(t *testing.T) {
	rec := do(t, http.MethodGet, "/hello/ada?shout=true")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != "HELLO, ADA!" {
		t.Fatalf("body = %q, want %q", got, "HELLO, ADA!")
	}
}

func TestShoutFalseIsQuiet(t *testing.T) {
	rec := do(t, http.MethodGet, "/hello/ada?shout=false")
	if got := rec.Body.String(); got != "hello, ada!" {
		t.Fatalf("body = %q, want %q", got, "hello, ada!")
	}
}

func TestEmptyNameIsNotFound(t *testing.T) {
	rec := do(t, http.MethodGet, "/hello/")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestWrongMethod(t *testing.T) {
	rec := do(t, http.MethodPost, "/hello/ada")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}
`,
				}},
				Hints: []string{
					"§mux.HandleFunc(\"GET /hello/{name}\", func(w http.ResponseWriter, r *http.Request) { ... })§.",
					"Read the segment with §name := r.PathValue(\"name\")§.",
					"§r.URL.Query().Get(\"shout\") == \"true\"§ decides whether to uppercase.",
					"The 405 test passes for free: a pattern registered for GET makes the mux answer 405 for other methods.",
					"Note §/hello/§ with an empty name does not match §/hello/{name}§ (wildcards need a non-empty segment), so you get 404 without any code.",
				},
				Solution: `package main

import (
	"fmt"
	"net/http"
	"strings"
)

func NewHelloHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if r.URL.Query().Get("shout") == "true" {
			fmt.Fprint(w, strings.ToUpper("hello, "+name+"!"))
			return
		}
		fmt.Fprint(w, "hello, "+name+"!")
	})

	return mux
}

func main() {
	fmt.Println(NewHelloHandler() != nil)
}
`,
			},
			{
				ID:    "m2t6e2",
				Title: "A handler with an injectable store",
				Prompt: `Take a narrow interface rather than a concrete database, so the handler is easy to test.

§§§go
type User struct {
	ID    int    §json:"id"§
	Email string §json:"email"§
}

type UserStore interface {
	Get(ctx context.Context, id int) (User, error)
}

var ErrNotFound = errors.New("not found")

func NewUserHandler(store UserStore) http.Handler
§§§

Behaviour of the returned handler:

* §GET /users/{id}§ → 200 with the JSON user, and a §Content-Type§ header.
* A store error that wraps §ErrNotFound§ → 404 with §{"error":"not found"}§.
* Any other store error → 500.
* A non-numeric id → 400.
* Constructing with a §nil§ store must not panic — it only fails when used.

The payoff: the test supplies a tiny §fakeStore§ struct backed by a map and never
touches a database. That is the reason to depend on an interface rather than a §*sql.DB§.
`,
				Starter: `package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type User struct {
	ID    int    §json:"id"§
	Email string §json:"email"§
}

type UserStore interface {
	Get(ctx context.Context, id int) (User, error)
}

var ErrNotFound = errors.New("not found")

// NewUserHandler wires a store into an HTTP handler.
func NewUserHandler(store UserStore) http.Handler {
	mux := http.NewServeMux()
	// TODO: register "GET /users/{id}"
	return mux
}

func main() {
	fmt.Println(NewUserHandler(nil) != nil)
	_ = json.Marshal
	_ = strconv.Atoi
}
`,
				Tests: []File{{
					Name: "handler_test.go",
					Content: `package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
		return User{}, fmt.Errorf("get %d: %w", id, ErrNotFound)
	}
	return u, nil
}

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestGetUserOK(t *testing.T) {
	h := NewUserHandler(fakeStore{users: map[int]User{7: {ID: 7, Email: "a@b.c"}}})
	rec := get(t, h, "/users/7")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct == "" {
		t.Fatal("Content-Type was not set")
	}
	var got User
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("the body is not a User: %v", err)
	}
	if got.ID != 7 || got.Email != "a@b.c" {
		t.Fatalf("user = %+v", got)
	}
}

func TestGetUserNotFound(t *testing.T) {
	h := NewUserHandler(fakeStore{users: map[int]User{}})
	rec := get(t, h, "/users/7")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var e map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil || e["error"] == "" {
		t.Fatalf("404 body = %s, want an error envelope", rec.Body.String())
	}
}

func TestGetUserStoreError(t *testing.T) {
	h := NewUserHandler(fakeStore{err: errors.New("db exploded")})
	rec := get(t, h, "/users/7")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if rec.Body.String() == "db exploded" {
		t.Fatal("the internal error message leaked to the client")
	}
}

func TestGetUserBadID(t *testing.T) {
	h := NewUserHandler(fakeStore{})
	rec := get(t, h, "/users/abc")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestConstructorWithNilStore(t *testing.T) {
	_ = NewUserHandler(nil) // must not panic at construction time
}
`,
				}},
				Hints: []string{
					"Register with §mux.HandleFunc(\"GET /users/{id}\", ...)§ and read §r.PathValue(\"id\")§.",
					"Parse with §strconv.Atoi§; on error write 400 with a JSON error body.",
					"Map the store error with §errors.Is(err, ErrNotFound)§ — it must work through a wrapped error.",
					"For the 500 case, write a generic message and let the real error go to a log, not to the client.",
					"Write the §Content-Type§ header before §WriteHeader§, or the header is ignored.",
				},
				Solution: `package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type User struct {
	ID    int    §json:"id"§
	Email string §json:"email"§
}

type UserStore interface {
	Get(ctx context.Context, id int) (User, error)
}

var ErrNotFound = errors.New("not found")

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func NewUserHandler(store UserStore) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeErr(w, http.StatusBadRequest, "id must be an integer")
			return
		}

		u, err := store.Get(r.Context(), id)
		switch {
		case errors.Is(err, ErrNotFound):
			writeErr(w, http.StatusNotFound, "not found")
		case err != nil:
			writeErr(w, http.StatusInternalServerError, "internal error")
		default:
			writeJSON(w, http.StatusOK, u)
		}
	})

	return mux
}

func main() {
	fmt.Println(NewUserHandler(nil) != nil)
}
`,
			},
		},
	}
}

func m2t7() Topic {
	return Topic{
		ID:      "m2t7",
		Title:   "Routing with chi and REST patterns",
		Summary: "Route groups, shared middleware, and a consistent API shape.",
		Theory: `
The standard §ServeMux§ is enough for small services. Real ones usually want three more
things:

1. **Route groups** — a subtree with its own prefix and middleware.
2. **Middleware stacks** declared in one place.
3. **Sub-routers** you can mount like a file tree.

§github.com/go-chi/chi/v5§ is the most popular answer. It is a *drop-in* §http.Handler§:
it uses the standard §http.HandlerFunc§ signature and plugs into anything that expects
§http.Handler§.

## Installing and creating one

§§§bash
go get github.com/go-chi/chi/v5
§§§

§§§go
import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

r := chi.NewRouter()
r.Use(middleware.RequestID)
r.Use(middleware.Recoverer)
§§§

§r.Use(mw)§ appends a middleware to everything registered **after** it.

## Route groups

§§§go
r := chi.NewRouter()

r.Get("/healthz", health)

r.Route("/api/v1", func(r chi.Router) {
	r.Get("/users", listUsers)
	r.Post("/users", createUser)

	r.Route("/users/{id}", func(r chi.Router) {
		r.Get("/", getUser)
		r.Put("/", updateUser)
		r.Delete("/", deleteUser)
	})
})
§§§

The inner §r§ is a **sub-router**: it carries the §/api/v1§ prefix and any middleware
declared in the outer §Use§ calls. Nesting mirrors the URL structure, which is exactly
what a reader wants.

§Group§ is the variant that adds middleware *without* adding a path prefix:

§§§go
r.Group(func(r chi.Router) {
	r.Use(RequireAPIKey(cfg.APIKey))
	r.Get("/stats", stats)
	r.Get("/admin", admin)
})
§§§

## Reading URL parameters

§§§go
func getUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	...
}
§§§

§chi.URLParam§ works for wildcards defined anywhere in the nested routes. (If you are on
Go 1.22+ and using the *standard* mux instead, it is §r.PathValue("id")§.)

## The example service, end to end

§§§go
package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Item struct {
	ID   int    §json:"id"§
	Name string §json:"name"§
}

type API struct {
	mu     sync.Mutex
	items  map[int]Item
	nextID int
}

func NewAPI() *API {
	return &API{items: map[int]Item{}, nextID: 1}
}

func (a *API) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Get("/items", a.list)
	r.Post("/items", a.create)
	r.Get("/items/{id}", a.get)
	r.Delete("/items/{id}", a.delete)
	return r
}
§§§

Then each handler is a small method on §*API§. Keeping the store on the struct is what
lets the handlers be tested with §httptest§ and a temporary store.

## REST shape that clients can rely on

| Operation | Method + path | Success | Failure |
| --- | --- | --- | --- |
| list | §GET /items§ | 200 + array | |
| read | §GET /items/{id}§ | 200 + object | 404 |
| create | §POST /items§ | 201 + object | 400 |
| replace | §PUT /items/{id}§ | 200 + object | 400 / 404 |
| delete | §DELETE /items/{id}§ | 204, no body | 404 |

Two habits that make an API pleasant:

* **Always return an array, not §null§, for "no results".** Initialise with
  §make([]Item, 0)§ or §[]Item{}§, never §var items []Item§. A JSON §null§ forces every
  client to special-case it.
* **Use one error envelope** — §{"error": "..."}§ — for every failure, so clients have
  one thing to parse.

## Protecting writes

Route groups are the clean way to put auth on a subset:

§§§go
r.Group(func(r chi.Router) {
	r.Use(RequireAuth(sessions))
	r.Post("/items", a.create)
	r.Delete("/items/{id}", a.delete)
})
§§§

The next topic builds exactly that middleware.

## chi's built-in middleware

§github.com/go-chi/chi/v5/middleware§ ships the boring parts:

| Middleware | Does |
| --- | --- |
| §RequestID§ | generates an §X-Request-ID§ and puts it in the context |
| §RealIP§ | fills §r.RemoteAddr§ from §X-Forwarded-For§ |
| §Logger§ | one log line per request |
| §Recoverer§ | turns a panic into a 500 |
| §Timeout(d)§ | cancels the request context after §d§ |
| §Compress§ | gzip response bodies |
| §Throttle§ / §Heartbeat§ | rate limiting and long-poll keep-alives |
| §NoCache§ | the right §Cache-Control§ headers for APIs |

You rarely need more than §RequestID§, §Recoverer§, §Timeout§ and a logger of your own.
`,
		Quiz: []Question{
			{
				ID:     "m2t7q1",
				Prompt: "What does §r.Route(\"/api\", func(r chi.Router) { r.Get(\"/users\", h) })§ register?",
				Options: []Option{
					{Text: "§GET /api/users§, handled by §h§.", Correct: true, Rationale: "§Route§ adds a prefix, so the inner path is joined to the outer one."},
					{Text: "§GET /users§ only.", Correct: false, Rationale: "The prefix applies to everything inside."},
					{Text: "§GET /api§ only.", Correct: false, Rationale: "The route is §/users§ inside the group."},
					{Text: "Nothing — §Route§ is only for middleware.", Correct: false, Rationale: "§Group§ is the middleware-only variant."},
				},
				Explanation: "§Route§ = prefix + nested routes. §Group§ = middleware only, no prefix.",
			},
			{
				ID:     "m2t7q2",
				Prompt: "Why does §r.Use(mw)§ only affect routes registered *after* it?",
				Options: []Option{
					{Text: "Because §Use§ appends to the middleware list, and a route captures the list as it stands when it is registered.", Correct: true, Rationale: "So order matters: §Use§ first, then the routes you want it to wrap."},
					{Text: "Because chi requires middleware to be declared twice.", Correct: false, Rationale: "Once is enough — in the right place."},
					{Text: "Because §Use§ is asynchronous.", Correct: false, Rationale: "It is a simple append."},
					{Text: "It affects all routes, including earlier ones.", Correct: false, Rationale: "Earlier routes keep the middleware list they captured."},
				},
				Explanation: "Declare middleware at the top of a router (or group) and register routes below it.",
			},
			{
				ID:     "m2t7q3",
				Prompt: "A §GET /items§ request finds no items. What should the response body be?",
				Options: []Option{
					{Text: "§[]§ — an empty JSON array.", Correct: true, Rationale: "Clients can then iterate without a null check. Use §make([]Item, 0)§ rather than a nil slice."},
					{Text: "§null§", Correct: false, Rationale: "A nil slice marshals to §null§, forcing every client to special-case it."},
					{Text: "§{}§", Correct: false, Rationale: "The endpoint returns a list, not an object."},
					{Text: "204 No Content", Correct: false, Rationale: "That is for a successful delete, not an empty list."},
				},
				Explanation: "Always encode an empty collection as §[]§. One line of code saves every consumer a branch.",
			},
			{
				ID:     "m2t7q4",
				Prompt: "Why return a 204 with no body for a successful delete?",
				Options: []Option{
					{Text: "The resource is gone, so there is nothing meaningful to return, and a body would just be noise.", Correct: true, Rationale: "204 No Content is precisely defined for this."},
					{Text: "Because HTTP forbids bodies on DELETE responses.", Correct: false, Rationale: "A body is allowed; it is simply not useful."},
					{Text: "Because clients cannot parse a body on a successful delete.", Correct: false, Rationale: "They can; it is a convention, not a rule."},
					{Text: "So that the client's HTTP library follows a redirect.", Correct: false, Rationale: "Unrelated."},
				},
				Explanation: "Match the response to the operation: 200 with data, 201 for creation, 204 for a successful no-content delete.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m2t7e1",
				Title: "Warm-up: a chi router over an in-memory store",
				Prompt: `Build a small REST resource using **chi**.

§§§go
type Item struct {
	ID   int    §json:"id"§
	Name string §json:"name"§
}

type API struct {
	// a mutex, a map[int]Item and a nextID counter
}

func NewAPI() *API
func (a *API) Routes() http.Handler
§§§

Endpoints:

* §GET /items§ → 200 with a JSON array. With no items it must be §[]§, **not** §null§.
* §POST /items§ → 201 and the created item with a server-assigned ID; 400 when
  §name§ is empty.
* §GET /items/{id}§ → 200, or 404 with §{"error":"not found"}§.
* §DELETE /items/{id}§ → 204 with an empty body, or 404.

Guard the map with the mutex on every access — the tests are run with §-race§ enabled
in CI, and a data race fails the build.

chi will be added to the sandbox §go.mod§ automatically.
`,
				Starter: `package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
)

type Item struct {
	ID   int    §json:"id"§
	Name string §json:"name"§
}

type API struct {
	mu     sync.Mutex
	items  map[int]Item
	nextID int
}

func NewAPI() *API {
	// TODO
	return nil
}

func (a *API) Routes() http.Handler {
	r := chi.NewRouter()
	// TODO: register the four endpoints
	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	fmt.Println(NewAPI() != nil, strconv.Itoa(1))
}
`,
				Deps: []string{"github.com/go-chi/chi/v5"},
				Tests: []File{{
					Name: "api_test.go",
					Content: `package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(NewAPI().Routes())
	t.Cleanup(srv.Close)
	return srv
}

func do(t *testing.T, method, url, body string) (*http.Response, []byte) {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = bytes.NewBufferString(body)
	}
	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp, data
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}

func TestEmptyListIsArray(t *testing.T) {
	srv := newServer(t)
	resp, body := do(t, http.MethodGet, srv.URL+"/items", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if string(bytes.TrimSpace(body)) != "[]" {
		t.Fatalf("empty list body = %q, want []", body)
	}
}

func TestCreateAndGet(t *testing.T) {
	srv := newServer(t)

	resp, body := do(t, http.MethodPost, srv.URL+"/items", "{\"name\":\"widget\"}")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want 201 (body %s)", resp.StatusCode, body)
	}
	if resp.Header.Get("Content-Type") == "" {
		t.Fatal("missing Content-Type")
	}
	var created Item
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatalf("create body is not an Item: %v (%s)", err, body)
	}
	if created.ID == 0 || created.Name != "widget" {
		t.Fatalf("created = %+v", created)
	}

	resp, body = do(t, http.MethodGet, srv.URL+"/items/"+itoa(created.ID), "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d, want 200", resp.StatusCode)
	}
	var got Item
	if err := json.Unmarshal(body, &got); err != nil || got.Name != "widget" {
		t.Fatalf("get body = %s (err %v)", body, err)
	}
}

func TestCreateValidation(t *testing.T) {
	srv := newServer(t)
	resp, _ := do(t, http.MethodPost, srv.URL+"/items", "{\"name\":\"\"}")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty name status = %d, want 400", resp.StatusCode)
	}
}

func TestNotFoundAndDelete(t *testing.T) {
	srv := newServer(t)

	resp, body := do(t, http.MethodGet, srv.URL+"/items/999", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("missing status = %d, want 404", resp.StatusCode)
	}
	var e map[string]string
	if err := json.Unmarshal(body, &e); err != nil || e["error"] == "" {
		t.Fatalf("404 body = %s, want an error envelope", body)
	}

	_, created := do(t, http.MethodPost, srv.URL+"/items", "{\"name\":\"x\"}")
	var item Item
	_ = json.Unmarshal(created, &item)

	resp, body = do(t, http.MethodDelete, srv.URL+"/items/"+itoa(item.ID), "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", resp.StatusCode)
	}
	if len(body) != 0 {
		t.Fatalf("204 response had a body: %q", body)
	}

	resp, _ = do(t, http.MethodDelete, srv.URL+"/items/"+itoa(item.ID), "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("second delete status = %d, want 404", resp.StatusCode)
	}
}
`,
				}},
				Hints: []string{
					"§NewAPI§ returns §&API{items: make(map[int]Item), nextID: 1}§.",
					"ch.Register with §r.Get(\"/items\", a.list)§, §r.Post(\"/items\", a.create)§, §r.Get(\"/items/{id}\", a.get)§, §r.Delete(\"/items/{id}\", a.delete)§.",
					"For the list, §out := make([]Item, 0, len(a.items))§ — that is what makes an empty list encode as §[]§.",
					"Read the id with §strconv.Atoi(chi.URLParam(r, \"id\"))§.",
					"For 204: §w.WriteHeader(http.StatusNoContent)§ and write nothing.",
				},
				Solution: `package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Item struct {
	ID   int    §json:"id"§
	Name string §json:"name"§
}

type API struct {
	mu     sync.Mutex
	items  map[int]Item
	nextID int
}

func NewAPI() *API {
	return &API{items: make(map[int]Item), nextID: 1}
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

	r.Get("/items", a.list)
	r.Post("/items", a.create)
	r.Get("/items/{id}", a.get)
	r.Delete("/items/{id}", a.delete)

	return r
}

func (a *API) list(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	out := make([]Item, 0, len(a.items))
	for _, it := range a.items {
		out = append(out, it)
	}
	a.mu.Unlock()
	writeJSON(w, http.StatusOK, out)
}

func (a *API) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string §json:"name"§
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Name == "" {
		writeErr(w, http.StatusBadRequest, "name is required")
		return
	}

	a.mu.Lock()
	it := Item{ID: a.nextID, Name: req.Name}
	a.nextID++
	a.items[it.ID] = it
	a.mu.Unlock()

	writeJSON(w, http.StatusCreated, it)
}

func (a *API) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}

	a.mu.Lock()
	it, ok := a.items[id]
	a.mu.Unlock()

	if !ok {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, it)
}

func (a *API) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}

	a.mu.Lock()
	_, ok := a.items[id]
	if ok {
		delete(a.items, id)
	}
	a.mu.Unlock()

	if !ok {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	fmt.Println(NewAPI() != nil)
}
`,
			},
		},
	}
}

func m2t8() Topic {
	return Topic{
		ID:      "m2t8",
		Title:   "Middleware and the request lifecycle",
		Summary: "Logging, recovery, request IDs and auth, all as wrappers around a handler.",
		Theory: `
**Middleware** is a function that wraps another handler to add behaviour — logging,
authentication, panic recovery — without the handler knowing.

## The signature

Because a handler is just §http.Handler§, middleware is a function that takes one and
returns one:

§§§go
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r) // call the wrapped handler
		fmt.Println(r.Method, r.URL.Path, time.Since(start))
	})
}
§§§

Read it inside-out: §Logging§ returns a *new* handler that does something, calls §next§,
and does something else. That "before / call / after" shape is all middleware ever is.

§http.HandlerFunc(func(w, r) { … })§ converts a plain function into a handler value —
the adaptor you met in the §net/http§ topic.

## Chaining several

§§§go
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

handler := Chain(mux, Logging, Recoverer, RequireToken("s3cret"))
§§§

Why backwards? Because the first middleware you list should run **first**. Applying them
right-to-left nests them so that the outermost wrapper is the one you named first:
§Logging(Recoverer(RequireToken(mux)))§.

Traced through, a request enters Logging, then Recoverer, then RequireToken, then the
mux; and the responses unwind in the opposite order.

Both chi and the standard library have their own way to do this:
§r.Use(mw)§ for chi, and §http.StripPrefix§ / manual wrapping for the standard mux.

## Recording the status code

§http.ResponseWriter§ has no §StatusCode()§ method, so a logging middleware must wrap
it. Two details matter:

§§§go
type statusRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if r.wrote {
		return // guard against a double WriteHeader
	}
	r.status = code
	r.wrote = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.wrote {
		r.WriteHeader(http.StatusOK) // a body implies 200
	}
	return r.ResponseWriter.Write(b)
}
§§§

The §Write§ override is not optional: a handler that only calls §w.Write§ never calls
§WriteHeader§ explicitly, and your recorder would report a status of 0.

## Panic recovery

A panic in a handler takes down the whole process unless something recovers. Wrap every
request:

§§§go
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v\n%s", rec, debug.Stack())
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
§§§

§debug.Stack()§ captures the goroutine's stack so the log line is actually useful. Without
a recoverer, one nil pointer is a production outage.

## Request IDs

§§§go
type ctxKey int

const reqIDKey ctxKey = 0

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newID()
		}
		w.Header().Set("X-Request-ID", id) // echo it to the client
		ctx := context.WithValue(r.Context(), reqIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx)) // pass the new request DOWN
	})
}

func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(reqIDKey).(string)
	return id
}
§§§

§r.WithContext(ctx)§ returns a **shallow copy** of the request with the new context.
Never mutate the request in place — it is shared with other goroutines.

Once every log line includes the request id, one request's whole story becomes findable
in your log store.

## Auth in one function

§§§go
func RequireToken(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if got == "" || got != token {
				w.Header().Set("WWW-Authenticate", "Bearer realm=api")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
§§§

Notes that matter:

* §401 Unauthorized§ means "I do not know who you are" (§403 Forbidden§ means "I know
  you, and no").
* Comparing secrets with §==§ leaks timing information. For real secrets use
  §crypto/subtle.ConstantTimeCompare§.
* **Never log the §Authorization§ header.** Logs are widely readable.

## The order that works

Outermost first:

1. **Recoverer** — so a panic anywhere below becomes a 500.
2. **RequestID** — so everything after it can log the id.
3. **Logger** — wants the id and the final status.
4. **CORS** — must run even for requests auth would reject.
5. **Timeout** — bounds the handler's context.
6. **Auth** — rejects early, cheaply.
7. your routes.

## A complete example

§§§go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello, id=%s", RequestIDFrom(r.Context()))
	})

	h := Chain(mux, RequestID, Recoverer)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.Header.Set("X-Request-ID", "abc-123")
	h.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
	fmt.Println(rec.Header().Get("X-Request-ID"))
}
§§§

Output:

§§§text
200
hello, id=abc-123
abc-123
§§§

That §httptest§ block is also the template for testing your own middleware — no network
required.
`,
		Quiz: []Question{
			{
				ID:     "m2t8q1",
				Prompt: "In §Chain(handler, A, B, C)§, as defined in the theory, which middleware's code before §next.ServeHTTP§ runs first?",
				Options: []Option{
					{Text: "A — it wraps everything else and is therefore outermost.", Correct: true, Rationale: "§Chain§ applies right-to-left, making the first-named middleware the outer wrapper."},
					{Text: "C — it is listed last.", Correct: false, Rationale: "§C§ is innermost, closest to the handler."},
					{Text: "All three run at the same time.", Correct: false, Rationale: "Middleware nesting is synchronous."},
					{Text: "The order is undefined.", Correct: false, Rationale: "It is completely deterministic."},
				},
				Explanation: "§A(B(C(handler)))§: requests go in A → B → C → handler and responses come back C → B → A.",
			},
			{
				ID:     "m2t8q2",
				Prompt: "Why must a status-recording middleware override §Write§ as well as §WriteHeader§?",
				Options: []Option{
					{Text: "Because a handler that only calls §w.Write§ implicitly sends 200, and your wrapper would otherwise record a status of 0.", Correct: true, Rationale: "Many handlers never call WriteHeader explicitly."},
					{Text: "Because §Write§ can change the status.", Correct: false, Rationale: "The status is frozen after the first header write."},
					{Text: "Because §WriteHeader§ is deprecated.", Correct: false, Rationale: "It is current."},
					{Text: "It does not need to — §WriteHeader§ is always called explicitly.", Correct: false, Rationale: "It frequently is not."},
				},
				Explanation: "Override both, and guard against a second §WriteHeader§ call, which logs a superfluous-call warning.",
			},
			{
				ID:     "m2t8q3",
				Prompt: "Which middleware should be outermost?",
				Options: []Option{
					{Text: "Recoverer, so a panic anywhere below — including in other middleware — becomes a 500.", Correct: true, Rationale: "Anything inside the recoverer is protected."},
					{Text: "Auth, so unauthenticated traffic is rejected as early as possible.", Correct: false, Rationale: "Rejecting early is nice, but then a panic in auth itself is unhandled."},
					{Text: "The router.", Correct: false, Rationale: "The router is the innermost handler."},
					{Text: "Logger, because logging is the most important.", Correct: false, Rationale: "Logging wants the final status, so it sits just inside Recoverer."},
				},
				Explanation: "Typical order: Recoverer → RequestID → Logger → CORS → Timeout → Auth → routes.",
			},
			{
				ID:     "m2t8q4",
				Prompt: "A request carries a valid token but the user lacks the §admin§ role. Which status code?",
				Options: []Option{
					{Text: "403 Forbidden", Correct: true, Rationale: "Authentication succeeded; authorisation failed."},
					{Text: "401 Unauthorized", Correct: false, Rationale: "401 means the credentials are missing or invalid."},
					{Text: "404 Not Found", Correct: false, Rationale: "Occasionally used to hide a resource's existence, but not the general answer."},
					{Text: "400 Bad Request", Correct: false, Rationale: "The request is well formed."},
				},
				Explanation: "401 = who are you? 403 = I know you, and no.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m2t8e1",
				Title: "Warm-up: a middleware stack",
				Prompt: `Implement four composable middlewares over the standard library.

§§§go
func Recoverer(next http.Handler) http.Handler
func RequestID(next http.Handler) http.Handler
func RequireToken(token string) func(http.Handler) http.Handler
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler
func RequestIDFrom(ctx context.Context) string
§§§

* §Recoverer§ turns a panic into §500§ with the body §internal server error§.
* §RequestID§ echoes an incoming §X-Request-ID§ or generates one, sets it on the
  response header, and stores it in the context so a handler can read it with
  §RequestIDFrom§.
* §RequireToken§ accepts §Authorization: Bearer <token>§ and returns §401§ otherwise.
* §Chain§ applies the **first** middleware outermost.

The tests use §httptest§ throughout; one of them checks the exact before/handler/after
ordering, which is the part people get wrong.
`,
				Starter: `package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

type ctxKey int

const reqIDKey ctxKey = 0

func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(reqIDKey).(string)
	return id
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func Recoverer(next http.Handler) http.Handler {
	// TODO
	return next
}

func RequestID(next http.Handler) http.Handler {
	// TODO
	return next
}

func RequireToken(token string) func(http.Handler) http.Handler {
	// TODO
	return func(next http.Handler) http.Handler { return next }
}

func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	// TODO
	return h
}

func main() {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok", RequestIDFrom(r.Context()))
	}), RequestID, Recoverer)
	fmt.Println(h != nil, strings.TrimPrefix("x", "x"))
}
`,
				Tests: []File{{
					Name: "mw_test.go",
					Content: `package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoverer(t *testing.T) {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("kaboom")
	}), Recoverer)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "internal server error") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestRequestIDEchoesIncoming(t *testing.T) {
	var seen string
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}), RequestID)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "abc-123")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if seen != "abc-123" {
		t.Fatalf("context request id = %q, want abc-123", seen)
	}
	if got := rec.Header().Get("X-Request-ID"); got != "abc-123" {
		t.Fatalf("response header = %q, want abc-123", got)
	}
}

func TestRequestIDGenerates(t *testing.T) {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), RequestID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("no request id was generated")
	}
}

func TestRequireToken(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	h := Chain(ok, RequireToken("s3cret"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: status = %d, want 401", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token: status = %d, want 401", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer s3cret")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot {
		t.Fatalf("valid token: status = %d, want 418", rec.Code)
	}
}

func TestChainOrder(t *testing.T) {
	var order []string

	first := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "first-in")
			next.ServeHTTP(w, r)
			order = append(order, "first-out")
		})
	}
	second := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "second-in")
			next.ServeHTTP(w, r)
			order = append(order, "second-out")
		})
	}

	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}), first, second)

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	want := []string{"first-in", "second-in", "handler", "second-out", "first-out"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v", order, want)
		}
	}
}
`,
				}},
				Hints: []string{
					"§Chain§ iterates backwards: §for i := len(mws) - 1; i >= 0; i-- { h = mws[i](h) }§.",
					"§Recoverer§: §defer func() { if rec := recover(); rec != nil { http.Error(w, \"internal server error\", 500) } }()§ around §next.ServeHTTP§.",
					"§RequestID§: build the new context, then call §next.ServeHTTP(w, r.WithContext(ctx))§.",
					"§RequireToken§: §strings.TrimPrefix(r.Header.Get(\"Authorization\"), \"Bearer \")§ then compare.",
				},
				Solution: `package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

type ctxKey int

const reqIDKey ctxKey = 0

func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(reqIDKey).(string)
	return id
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
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

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), reqIDKey, id)))
	})
}

func RequireToken(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if got == "" || got != token {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func main() {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok", RequestIDFrom(r.Context()))
	}), RequestID, Recoverer)
	fmt.Println(h != nil)
}
`,
			},
			{
				ID:    "m2t8e2",
				Title: "A logging middleware that records the status",
				Prompt: `Logging middleware is useless without the outcome, and §http.ResponseWriter§ has no
§StatusCode()§ method. So you wrap it.

§§§go
type statusRecorder struct {
	http.ResponseWriter
	Status int
	Wrote  bool
}

func (r *statusRecorder) WriteHeader(code int)
func (r *statusRecorder) Write(b []byte) (int, error)

type LogEntry struct {
	Method    string
	Path      string
	Status    int
	RequestID string
}

func Middleware(log func(LogEntry)) func(http.Handler) http.Handler
§§§

* §statusRecorder§ records the **first** status code only. A second §WriteHeader§ call is
  ignored (that is what the real §net/http§ does too).
* If the handler writes only a body, §Write§ must record status 200 first.
* §Middleware(log)§ calls §log§ exactly once per request, *after* the handler returns,
  with the method, path, final status and the request id from the context.
* The request id is read with §RequestIDFrom(r.Context())§ — the function is provided.

The test asserts the status of both a 200-byte-writing handler and a 418 handler, which
is where a naive implementation reports 0.
`,
				Starter: `package main

import (
	"context"
	"fmt"
	"net/http"
)

type ctxKey int

const reqIDKey ctxKey = 0

func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(reqIDKey).(string)
	return id
}

type statusRecorder struct {
	http.ResponseWriter
	Status int
	Wrote  bool
}

func (r *statusRecorder) WriteHeader(code int) {
	// TODO: record the first status only, then pass it on
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	// TODO: if nothing has been written yet, record 200
	return r.ResponseWriter.Write(b)
}

type LogEntry struct {
	Method    string
	Path      string
	Status    int
	RequestID string
}

// Middleware logs one entry per request, after the handler finishes.
func Middleware(log func(LogEntry)) func(http.Handler) http.Handler {
	// TODO
	return func(next http.Handler) http.Handler { return next }
}

func main() {
	h := Middleware(func(e LogEntry) { fmt.Println(e) })(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hi")
	}))
	fmt.Println(h != nil)
}
`,
				Tests: []File{{
					Name: "log_test.go",
					Content: `package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecordsExplicitStatus(t *testing.T) {
	var got []LogEntry
	h := Middleware(func(e LogEntry) { got = append(got, e) })(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/brew", nil))

	if rec.Code != http.StatusTeapot {
		t.Fatalf("the wrapped handler's status was not passed through: %d", rec.Code)
	}
	if len(got) != 1 {
		t.Fatalf("log called %d times, want 1", len(got))
	}
	if got[0].Status != http.StatusTeapot {
		t.Fatalf("logged status = %d, want 418", got[0].Status)
	}
	if got[0].Method != http.MethodGet || got[0].Path != "/brew" {
		t.Fatalf("logged %+v", got[0])
	}
}

func TestRecordsImplicit200(t *testing.T) {
	var got []LogEntry
	h := Middleware(func(e LogEntry) { got = append(got, e) })(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("hello")) // never calls WriteHeader
		}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/p", nil))

	if len(got) != 1 {
		t.Fatalf("log called %d times, want 1", len(got))
	}
	if got[0].Status != http.StatusOK {
		t.Fatalf("logged status = %d, want 200 (a body implies 200)", got[0].Status)
	}
	if rec.Body.String() != "hello" {
		t.Fatalf("body was lost: %q", rec.Body.String())
	}
}

func TestIgnoresSecondWriteHeader(t *testing.T) {
	var got []LogEntry
	h := Middleware(func(e LogEntry) { got = append(got, e) })(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.WriteHeader(http.StatusInternalServerError) // superfluous
		}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/x", nil))
	if got[0].Status != http.StatusCreated {
		t.Fatalf("logged status = %d, want the first one (201)", got[0].Status)
	}
}

func TestStatusRecorderDefaults(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec}
	if _, err := sr.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	if sr.Status != http.StatusOK {
		t.Fatalf("Status = %d, want 200", sr.Status)
	}
}

func TestLogReadsRequestID(t *testing.T) {
	var got []LogEntry
	base := Middleware(func(e LogEntry) { got = append(got, e) })(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), reqIDKey, "req-9"))
	base.ServeHTTP(httptest.NewRecorder(), req)

	if got[0].RequestID != "req-9" {
		t.Fatalf("logged request id = %q, want req-9", got[0].RequestID)
	}
}
`,
				}},
				Hints: []string{
					"§WriteHeader§: §if r.Wrote { return }; r.Status = code; r.Wrote = true; r.ResponseWriter.WriteHeader(code)§.",
					"§Write§: §if !r.Wrote { r.WriteHeader(http.StatusOK) }§ before delegating.",
					"§Middleware§ returns a §func(http.Handler) http.Handler§ that constructs a §statusRecorder§, calls §next§, then calls §log§.",
					"Read the request id *after* calling §next§ — the context is per-request and already carries it.",
				},
				Solution: `package main

import (
	"context"
	"fmt"
	"net/http"
)

type ctxKey int

const reqIDKey ctxKey = 0

func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(reqIDKey).(string)
	return id
}

type statusRecorder struct {
	http.ResponseWriter
	Status int
	Wrote  bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if r.Wrote {
		return
	}
	r.Status = code
	r.Wrote = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.Wrote {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(b)
}

type LogEntry struct {
	Method    string
	Path      string
	Status    int
	RequestID string
}

func Middleware(log func(LogEntry)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)

			log(LogEntry{
				Method:    r.Method,
				Path:      r.URL.Path,
				Status:    rec.Status,
				RequestID: RequestIDFrom(r.Context()),
			})
		})
	}
}

func main() {
	h := Middleware(func(e LogEntry) { fmt.Println(e) })(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hi")
	}))
	fmt.Println(h != nil)
}
`,
			},
		},
	}
}

func m2t9() Topic {
	return Topic{
		ID:      "m2t9",
		Title:   "HTTP clients and timeouts",
		Summary: "Calling someone else's API without hanging your own service.",
		Theory: `
Your service will call other services. Go's §net/http§ client is excellent, but its
defaults are dangerous, and the mistakes here cause production incidents.

## The default client has no timeout

§§§go
resp, err := http.Get("https://slow.example.com/data")
§§§

§http.Get§ uses §http.DefaultClient§, whose §Timeout§ is **zero — meaning "wait
forever"**. If the remote server accepts your connection and then never replies, that
line blocks for the life of the process, holding a goroutine and a connection.

Always make your own client:

§§§go
client := &http.Client{
	Timeout: 10 * time.Second, // the whole request: dial + write + read
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}
§§§

**Create it once and reuse it.** An §*http.Client§ is safe for concurrent use and
contains a connection pool. Building one per request throws the pool away and leaks
connections — a classic slow death.

## Building a request with a context

For anything where cancellation matters (which is: everything on a request path), build
the request yourself:

§§§go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
if err != nil {
	return nil, fmt.Errorf("build request: %w", err)
}
req.Header.Set("Accept", "application/json")

resp, err := client.Do(req)
if err != nil {
	return nil, fmt.Errorf("get %s: %w", url, err)
}
defer resp.Body.Close()
§§§

§defer resp.Body.Close()§ is not optional. Without it the connection is never returned
to the pool and you leak until the process runs out of file descriptors.

## Checking the status code

A 404 or 500 body will happily decode into your struct — usually producing a zero value
that looks like valid data. Check first:

§§§go
if resp.StatusCode != http.StatusOK {
	// Read a bounded amount for the error message.
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return nil, fmt.Errorf("get %s: status %d: %s", url, resp.StatusCode, bytes.TrimSpace(body))
}
§§§

That §io.LimitReader§ matters: an error response can be enormous, and you do not want to
buffer it all just to log it.

## Decoding the response

§§§go
var out Payload
if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
	return nil, fmt.Errorf("decode %s: %w", url, err)
}
§§§

The decoder streams from the body — no need to §ReadAll§ first.

## PUT, POST and friends

§§§go
body, _ := json.Marshal(payload)

req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
if err != nil {
	return err
}
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+token)
§§§

§bytes.NewReader§ turns a §[]byte§ into an §io.Reader§ for the request body.

## Timeouts in layers

There are three places a deadline can come from, and they compose:

| Layer | Where | Covers |
| --- | --- | --- |
| §client.Timeout§ | your client | the entire exchange including connection setup |
| §context.WithTimeout(ctx, d)§ | the request | everything downstream of that point |
| server-side timeouts | the other service | its own processing |

§client.Timeout§ is the blunt instrument that guarantees you never hang. The context is
the polite one that ties the call to the caller's budget.

## A complete client

§§§go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func postJSON(ctx context.Context, url string, in, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("post %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("post %s: status %d: %s", url, resp.StatusCode, bytes.TrimSpace(detail))
	}

	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response from %s: %w", url, err)
	}
	return nil
}
§§§

Five habits in twenty lines: reusable client, context on the request, bounded error
bodies, status check before decode, and errors wrapped with §%w§ so callers can use
§errors.Is§.

## Testing a client without the internet

§httptest.NewServer§ starts a real HTTP server on a random local port:

§§§go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "{\"ok\":true}")
}))
defer srv.Close()

// srv.URL is something like http://127.0.0.1:54321
§§§

Point your client at §srv.URL§ and you have an end-to-end test of the real HTTP stack —
headers, status codes, body streaming and all — with no network and no mocks.
`,
		Quiz: []Question{
			{
				ID:     "m2t9q1",
				Prompt: "Why is §http.Get(url)§ dangerous in a service?",
				Options: []Option{
					{Text: "It uses §http.DefaultClient§, which has no timeout, so a hanging server blocks the goroutine forever.", Correct: true, Rationale: "A zero §Timeout§ means unlimited waiting."},
					{Text: "It cannot follow redirects.", Correct: false, Rationale: "It follows up to ten redirects by default."},
					{Text: "It does not send an Accept header.", Correct: false, Rationale: "That is a cosmetic issue, not a hazard."},
					{Text: "It opens a new connection per call and closes it.", Correct: false, Rationale: "It shares the default transport's pool — which is why the timeout, not the pooling, is the problem."},
				},
				Explanation: "Always use a client you configured, with a §Timeout§. Better still, use §NewRequestWithContext§ for per-call deadlines.",
			},
			{
				ID:     "m2t9q2",
				Prompt: "What must you always do with a response body?",
				Options: []Option{
					{Text: "§defer resp.Body.Close()§ immediately after a successful §Do§, or the connection is never reused and you leak file descriptors.", Correct: true, Rationale: "Closing returns the connection to the pool."},
					{Text: "Read it completely before closing.", Correct: false, Rationale: "Useful for connection reuse, but not required — and never a substitute for closing."},
					{Text: "Only close it when the status is 200.", Correct: false, Rationale: "Close it whatever the status, including errors."},
					{Text: "Nothing — Go closes it when the function returns.", Correct: false, Rationale: "Go does not close it for you."},
				},
				Explanation: "§defer resp.Body.Close()§ right after the error check is the pattern to memorise.",
			},
			{
				ID:     "m2t9q3",
				Prompt: "Why check §resp.StatusCode§ before decoding the body?",
				Options: []Option{
					{Text: "Because an error page can decode cleanly into your struct as zero values, which looks like valid data.", Correct: true, Rationale: "A 500 with a JSON error body will happily unmarshal into a struct with no matching fields."},
					{Text: "Because §Decode§ fails on non-200 responses.", Correct: false, Rationale: "It knows nothing about status codes."},
					{Text: "Because the status is only set after reading the body.", Correct: false, Rationale: "The status arrives in the headers, before the body."},
					{Text: "It is not necessary; the JSON will be wrong anyway.", Correct: false, Rationale: "It very often is not wrong — just empty."},
				},
				Explanation: "Check the status, read a *bounded* amount of the error body for the message, then return a wrapped error.",
			},
			{
				ID:     "m2t9q4",
				Prompt: "Where should you create the §*http.Client§?",
				Options: []Option{
					{Text: "Once, at package level or on a struct, and reuse it — it is safe for concurrent use and holds the connection pool.", Correct: true, Rationale: "A client per request discards the pool and leaks idle connections."},
					{Text: "Inside each request handler, so each request gets a fresh one.", Correct: false, Rationale: "That is the leak. It is also slower."},
					{Text: "Inside the function that makes the call, using §http.DefaultClient§.", Correct: false, Rationale: "The default has no timeout."},
					{Text: "You should not create one; always use §http.Get§.", Correct: false, Rationale: "§http.Get§ cannot be cancelled and has no timeout."},
				},
				Explanation: "One long-lived client with a tuned §Transport§ is the standard shape.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m2t9e1",
				Title: "Warm-up: GET JSON into a struct",
				Prompt: `§§§go
func GetJSON(ctx context.Context, client *http.Client, url string, out any) error
§§§

* Builds the request with §http.NewRequestWithContext§ so the caller's deadline applies.
* Checks the status: anything outside 200–299 is an error that **includes the status
  code and a bounded slice of the body** (use §io.LimitReader§ so a huge error page does
  not blow up memory).
* Closes the body with §defer resp.Body.Close()§.
* Decodes into §out§ with §json.NewDecoder(resp.Body)§.
* Every failure is wrapped with §fmt.Errorf("...: %w", err)§ so the cause survives.

The test starts a real server with §httptest.NewServer§ and points the client at
§srv.URL§ — so this exercises the actual HTTP stack, not a mock.
`,
				Starter: `package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GetJSON fetches url and decodes the JSON body into out.
func GetJSON(ctx context.Context, client *http.Client, url string, out any) error {
	// TODO
	return nil
}

func main() {
	resp := struct {
		Name string §json:"name"§
	}{}
	_ = GetJSON(context.Background(), http.DefaultClient, "http://example.com", &resp)
	fmt.Println(resp)
	_ = io.LimitReader
	_ = json.Marshal
}
`,
				Tests: []File{{
					Name: "getjson_test.go",
					Content: `package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type payload struct {
	Name string §json:"name"§
	Age  int    §json:"age"§
}

func TestGetJSONSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "{\"name\":\"Ada\",\"age\":36}")
	}))
	defer srv.Close()

	var got payload
	if err := GetJSON(context.Background(), srv.Client(), srv.URL, &got); err != nil {
		t.Fatalf("GetJSON: %v", err)
	}
	if got.Name != "Ada" || got.Age != 36 {
		t.Fatalf("decoded %+v", got)
	}
}

func TestGetJSONStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "something exploded")
	}))
	defer srv.Close()

	err := GetJSON(context.Background(), srv.Client(), srv.URL, &payload{})
	if err == nil {
		t.Fatal("a 500 response should be an error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("the error should mention the status code: %v", err)
	}
}

func TestGetJSONBadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not json at all")
	}))
	defer srv.Close()

	if err := GetJSON(context.Background(), srv.Client(), srv.URL, &payload{}); err == nil {
		t.Fatal("a body that is not JSON should be an error")
	}
}

func TestGetJSONRespectsContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	start := time.Now()
	if err := GetJSON(ctx, srv.Client(), srv.URL, &payload{}); err == nil {
		t.Fatal("a cancelled context should produce an error")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("the call ignored the context deadline (%v)", elapsed)
	}
}

func TestGetJSONUnreachable(t *testing.T) {
	if err := GetJSON(context.Background(), &http.Client{Timeout: time.Second}, "http://127.0.0.1:1/", &payload{}); err == nil {
		t.Fatal("an unreachable server should produce an error")
	}
}
`,
				}},
				Hints: []string{
					"§req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)§.",
					"§resp, err := client.Do(req)§ then §defer resp.Body.Close()§.",
					"Status check: §if resp.StatusCode < 200 || resp.StatusCode > 299 { detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512)); return fmt.Errorf(\"get %s: status %d: %s\", url, resp.StatusCode, detail) }§.",
					"Then §return json.NewDecoder(resp.Body).Decode(out)§, wrapped with context.",
				},
				Solution: `package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetJSON(ctx context.Context, client *http.Client, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("get %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("get %s: status %d: %s", url, resp.StatusCode, bytes.TrimSpace(detail))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response from %s: %w", url, err)
	}
	return nil
}

func main() {
	resp := struct {
		Name string §json:"name"§
	}{}
	_ = GetJSON(context.Background(), http.DefaultClient, "http://example.com", &resp)
	fmt.Println(resp)
}
`,
			},
			{
				ID:    "m2t9e2",
				Title: "Post JSON and fetch several URLs concurrently",
				Prompt: `Two further client techniques.

**1. A JSON POST.**

§§§go
func PostJSON(ctx context.Context, client *http.Client, url string, in any, out any) error
§§§

* Marshals §in§, sends it as §application/json§, and decodes the response into §out§
  when §out§ is not §nil§.
* Same discipline as the GET: context, status check with a bounded error body,
  §defer Close§, wrapped errors.

**2. Concurrent fetches.**

§§§go
func FetchAll(ctx context.Context, client *http.Client, urls []string) ([]string, error)
§§§

* Fetches every URL **concurrently** (one goroutine each) and returns the raw response
  bodies in the order of the input slice. The bodies are not JSON, so read them with
  §io.ReadAll(resp.Body)§.
* If **any** fetch fails, return an error (and do not return a partial slice).
* An empty url list returns an **empty non-nil** slice and a nil error.
* Preserve order — so each goroutine must write into its own index, exactly like the
  §ParallelMap§ exercise from the channels topic.

Combine "one goroutine per job" with "collect the results safely": that is the shape of
almost every fan-out in a backend.
`,
				Starter: `package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// PostJSON sends in as JSON and, if out is non-nil, decodes the reply into it.
func PostJSON(ctx context.Context, client *http.Client, url string, in any, out any) error {
	// TODO
	return nil
}

// FetchAll fetches every url concurrently, preserving order.
func FetchAll(ctx context.Context, client *http.Client, urls []string) ([]string, error) {
	// TODO
	_ = sync.WaitGroup{}
	return nil, nil
}

func main() {
	_ = bytes.NewReader
	_ = json.Marshal
	_ = io.LimitReader
	fmt.Println(FetchAll(context.Background(), http.DefaultClient, nil))
}
`,
				Tests: []File{{
					Name: "post_fetch_test.go",
					Content: `package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPostJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Errorf("Content-Type = %q", ct)
		}

		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "{\"echo\":%q}", in["name"])
	}))
	defer srv.Close()

	var out map[string]string
	err := PostJSON(context.Background(), srv.Client(), srv.URL, map[string]string{"name": "ada"}, &out)
	if err != nil {
		t.Fatalf("PostJSON: %v", err)
	}
	if out["echo"] != "ada" {
		t.Fatalf("response = %v", out)
	}
}

func TestPostJSONNilOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	if err := PostJSON(context.Background(), srv.Client(), srv.URL, map[string]string{"a": "b"}, nil); err != nil {
		t.Fatalf("PostJSON with nil out: %v", err)
	}
}

func TestPostJSONStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "{\"error\":\"bad\"}")
	}))
	defer srv.Close()

	err := PostJSON(context.Background(), srv.Client(), srv.URL, map[string]string{}, nil)
	if err == nil {
		t.Fatal("a 400 response should be an error")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("the error should mention the status: %v", err)
	}
}

func TestFetchAllOrder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/slow":
			time.Sleep(60 * time.Millisecond)
			fmt.Fprint(w, "slow-body")
		case "/fast":
			fmt.Fprint(w, "fast-body")
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	got, err := FetchAll(context.Background(), srv.Client(), []string{srv.URL + "/slow", srv.URL + "/fast"})
	if err != nil {
		t.Fatalf("FetchAll: %v", err)
	}
	want := []string{"slow-body", "fast-body"}
	if len(got) != len(want) {
		t.Fatalf("FetchAll = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("FetchAll = %v, want %v (results must stay in input order)", got, want)
		}
	}
}

func TestFetchAllRunsConcurrently(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		fmt.Fprint(w, "x")
	}))
	defer srv.Close()

	urls := []string{srv.URL + "/1", srv.URL + "/2", srv.URL + "/3", srv.URL + "/4"}
	start := time.Now()
	if _, err := FetchAll(context.Background(), srv.Client(), urls); err != nil {
		t.Fatalf("FetchAll: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 350*time.Millisecond {
		t.Fatalf("FetchAll took %v — the fetches look sequential", elapsed)
	}
}

func TestFetchAllError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bad" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, "ok")
	}))
	defer srv.Close()

	if _, err := FetchAll(context.Background(), srv.Client(), []string{srv.URL + "/ok", srv.URL + "/bad"}); err == nil {
		t.Fatal("a failing URL should make FetchAll return an error")
	}
}

func TestFetchAllEmpty(t *testing.T) {
	got, err := FetchAll(context.Background(), http.DefaultClient, nil)
	if err != nil {
		t.Fatalf("FetchAll(nil) error = %v, want nil", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("FetchAll(nil) = %v, want an empty non-nil slice", got)
	}
}
`,
				}},
				Hints: []string{
					"§PostJSON§: §body, _ := json.Marshal(in)§, then §http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))§, and set the §Content-Type§ header.",
					"Only decode when §out != nil§ — the test posts with a nil out and expects no error.",
					"§FetchAll§: preallocate §bodies := make([]string, len(urls))§ and an §errs := make([]error, len(urls))§, then one goroutine per index.",
					"Each goroutine writes only its own index — that is why no lock is needed.",
					"After §wg.Wait()§, return the first non-nil error (or collect them with §errors.Join§).",
				},
				Solution: `package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

func PostJSON(ctx context.Context, client *http.Client, url string, in any, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("post %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("post %s: status %d: %s", url, resp.StatusCode, bytes.TrimSpace(detail))
	}

	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response from %s: %w", url, err)
	}
	return nil
}

func FetchAll(ctx context.Context, client *http.Client, urls []string) ([]string, error) {
	bodies := make([]string, len(urls))
	errs := make([]error, len(urls))

	var wg sync.WaitGroup
	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string) {
			defer wg.Done()
			bodies[i], errs[i] = fetchBody(ctx, client, u)
		}(i, u)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return bodies, nil
}

func fetchBody(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("get %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("get %s: status %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", url, err)
	}
	return string(data), nil
}

func main() {
	fmt.Println(FetchAll(context.Background(), http.DefaultClient, nil))
}
`,
			},
		},
	}
}
