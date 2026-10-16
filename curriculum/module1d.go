package curriculum

func m1t15() Topic {
	return Topic{
		ID:      "m1t15",
		Title:   "Methods and embedding",
		Summary: "Receivers, method sets, and composition by embedding.",
		Theory: `
Go has no classes. Instead you may attach a function to a named type, and that function
becomes a **method** of the type.

## A method is a function with a receiver

§§§go
type Rectangle struct {
	Width  float64
	Height float64
}

//            vvvvvvvv  <- the receiver: "this method belongs to Rectangle"
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func main() {
	r := Rectangle{Width: 4, Height: 3}
	fmt.Println(r.Area()) // 12  <- called with a dot
}
§§§

Written as a plain function it would be §func Area(r Rectangle) float64§. The receiver
form is the same thing with nicer calling syntax: §r.Area()§ instead of §Area(r)§.

You may only define methods on a type declared **in your own package**. So
§func (s string) Shout()§ is illegal — §string§ belongs to someone else. The workaround
is a named type of your own:

§§§go
type Name string

func (n Name) Shout() string {
	return strings.ToUpper(string(n)) + "!"
}
§§§

## Value receivers vs pointer receivers

This is the decision you make for every method, and there are only two rules.

**A value receiver gets a copy.** It cannot change the original:

§§§go
func (r Rectangle) DoubleBroken() {
	r.Width *= 2 // changes the copy
}
§§§

**A pointer receiver gets access to the original.** It can change it:

§§§go
func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor // changes the real thing
	r.Height *= factor
}
§§§

Callers do not have to think about it, because Go takes the address for you when the
value is addressable:

§§§go
r := Rectangle{2, 3}
r.Scale(2)   // Go rewrites this as (&r).Scale(2)
fmt.Println(r.Width) // 4
§§§

### Which do I use?

| Use a **pointer** receiver when | Use a **value** receiver when |
| --- | --- |
| The method changes the receiver | The receiver is small and immutable |
| The struct is large (copying is wasteful) | The type is conceptually a value (§time.Time§, §url.URL§) |
| The struct contains a §sync.Mutex§ (never copy one) | The method only reads and the struct is tiny |

**Be consistent.** If one method on a type needs a pointer receiver, give them all
pointer receivers. Mixed method sets are legal but confusing, and they change which
interfaces the type satisfies (the next topic shows exactly how).

## The method set rule, in one table

| You have | Value-receiver methods callable | Pointer-receiver methods callable |
| --- | --- | --- |
| §T§ | yes | no |
| §*T§ | yes | yes |

So §*T§ can do everything §T§ can, and more. When you pass values into an interface
later, this table is what decides whether the compiler accepts it.

## Composition with embedding

Go has no inheritance. Instead you can **embed** one type inside another, and the inner
type's fields and methods are *promoted* — reachable as if they belonged to the outer
type.

§§§go
type Animal struct {
	Name string
}

func (a Animal) Describe() string {
	return "I am " + a.Name
}

type Dog struct {
	Animal // <- embedded: no field name, just the type
	Breed  string
}

func main() {
	d := Dog{Animal: Animal{Name: "Rex"}, Breed: "Husky"}

	fmt.Println(d.Name)           // promoted field
	fmt.Println(d.Describe())     // promoted method
	fmt.Println(d.Animal.Name)    // the explicit path, always available
}
§§§

Output:

§§§text
Rex
I am Rex
Rex
§§§

Read embedding as **"has a, and lends its surface"**. A §Dog§ is not an §Animal§ in the
type system — you cannot assign a §Dog§ to an §Animal§ variable. It simply carries an
§Animal§ inside it and exposes its members.

### Overriding a promoted method

Declare a method with the same name on the outer type and it wins:

§§§go
func (d Dog) Describe() string {
	return d.Animal.Describe() + " (" + d.Breed + ")"
}
§§§

Calling §d.Animal.Describe()§ inside is the explicit way to reach the inner version —
Go has no §super§ keyword because the path is always spelled out.

### Shadowing with a field

If the outer struct declares a field with the same name as a promoted one, the outer
field wins:

§§§go
type Outer struct {
	Inner     // has a field Name
	Name string // shadows it
}
§§§

Generally, avoid this. Two ways to reach the same-looking name is a reading tax.

## A realistic example

§§§go
package main

import "fmt"

type Base struct {
	ID        int
	CreatedAt string
}

func (b Base) Describe() string {
	return fmt.Sprintf("#%d created %s", b.ID, b.CreatedAt)
}

type User struct {
	Base
	Username string
}

func (u User) Describe() string {
	return u.Base.Describe() + " (user " + u.Username + ")"
}

func main() {
	u := User{Base: Base{ID: 1, CreatedAt: "2024-01-01"}, Username: "ada"}
	fmt.Println(u.Describe())
	fmt.Println(u.Base.Describe()) // the inner version still exists
	fmt.Println(u.ID, u.CreatedAt) // promoted fields
}
§§§

Output:

§§§text
#1 created 2024-01-01 (user ada)
#1 created 2024-01-01
1 2024-01-01
§§§

This "embed common fields, override the interesting method" pattern is everywhere in
real Go code — it is how you share an ID and a timestamp across every entity in a
service without repeating yourself.
`,
		Quiz: []Question{
			{
				ID:     "m1t15q1",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

type Counter struct{ n int }

func (c Counter) IncBroken() { c.n++ }
func (c *Counter) Inc()     { c.n++ }

func main() {
	c := Counter{}
	c.IncBroken()
	fmt.Println(c.n)

	c2 := Counter{}
	c2.Inc()
	fmt.Println(c2.n)
}`,
				Options: []Option{
					{Text: "§0§ then §1§", Correct: true, Rationale: "The value receiver works on a copy, so the first counter never changes. The pointer receiver does change it."},
					{Text: "§1§ then §1§", Correct: false, Rationale: "§c.n++§ on a value receiver modifies the copy."},
					{Text: "§0§ then §0§", Correct: false, Rationale: "§c2.Inc()§ is rewritten by the compiler as §(&c2).Inc()§."},
					{Text: "It does not compile.", Correct: false, Rationale: "Both forms are valid; only their effect differs."},
				},
				Explanation: "Value receivers copy; pointer receivers share. This is the single most important method rule.",
			},
			{
				ID:     "m1t15q2",
				Prompt: "Which calls are legal? Select all that apply.",
				Multi:  true,
				Code: `type T struct{ N int }

func (t T) Get() int   { return t.N }
func (t *T) Set(n int) { t.N = n }

v := T{}
p := &T{}
T{}.Set(3)`,
				Options: []Option{
					{Text: "§v.Get()§", Correct: true, Rationale: "Value receiver on a value."},
					{Text: "§v.Set(1)§", Correct: true, Rationale: "§v§ is addressable, so Go rewrites it as §(&v).Set(1)§."},
					{Text: "§p.Get()§", Correct: true, Rationale: "The pointer is dereferenced automatically."},
					{Text: "§p.Set(2)§", Correct: true, Rationale: "Pointer receiver on a pointer."},
					{Text: "§T{}.Set(3)§", Correct: false, Rationale: "A composite literal is not addressable, so the address cannot be taken implicitly."},
				},
				Explanation: "Method calls work on addressable values for both receiver kinds. Non-addressable values only support value receivers.",
			},
			{
				ID:     "m1t15q3",
				Prompt: "Given §type Dog struct { Animal; Breed string }§, what is true?",
				Options: []Option{
					{Text: "A §Dog§ value can call §Animal§'s methods directly, but a §Dog§ cannot be assigned to a variable of type §Animal§.", Correct: true, Rationale: "Embedding promotes methods; it does not create a subtype relationship."},
					{Text: "A §Dog§ *is* an §Animal§, so assignment works.", Correct: false, Rationale: "Go has no inheritance. §var a Animal = d§ is a compile error."},
					{Text: "§Animal§'s fields become inaccessible once embedded.", Correct: false, Rationale: "They are promoted — even easier to reach, and still reachable as §d.Animal.X§."},
					{Text: "Embedding copies Animal into Dog at run time.", Correct: false, Rationale: "It is a compile-time, structural relationship; the field is stored inline."},
				},
				Explanation: "Embedding is composition that lends its surface: promoted members, no subtype relationship.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t15e1",
				Title: "Warm-up: a rectangle with methods",
				Prompt: `§§§go
type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64
func (r Rectangle) Perimeter() float64
func (r *Rectangle) Scale(factor float64)
func (r Rectangle) IsSquare() bool
func (r Rectangle) String() string
§§§

* §Area§ is §Width * Height§.
* §Perimeter§ is §2 * (Width + Height)§.
* §Scale§ multiplies both sides **in place** — this is why it needs a pointer receiver.
  The test checks that the change survives, which only a pointer receiver can achieve.
* §IsSquare§ is §Width == Height§.
* §String§ returns the shape as §"4.00x3.00"§ — width and height with two decimals.

A method called §String§ returning a §string§ is special: §fmt§ calls it automatically
when you print the value with §%v§ or §Println§. Give it a try in the editor.
`,
				Starter: `package main

import "fmt"

type Rectangle struct {
	Width, Height float64
}

// Area returns width * height.
func (r Rectangle) Area() float64 {
	// TODO
	return 0
}

// Perimeter returns 2*(width+height).
func (r Rectangle) Perimeter() float64 {
	// TODO
	return 0
}

// Scale multiplies both sides by factor, in place.
func (r *Rectangle) Scale(factor float64) {
	// TODO
}

// IsSquare reports whether the rectangle is a square.
func (r Rectangle) IsSquare() bool {
	// TODO
	return false
}

// String renders the rectangle as "WIDTHxHEIGHT" with two decimals.
func (r Rectangle) String() string {
	// TODO
	return ""
}

func main() {
	r := Rectangle{4, 3}
	fmt.Println(r.Area(), r.Perimeter(), r.IsSquare())
	fmt.Println(r)
	r.Scale(2)
	fmt.Println(r)
}
`,
				Tests: []File{{
					Name: "rect_test.go",
					Content: `package main

import (
	"math"
	"testing"
)

func TestRectangle(t *testing.T) {
	r := Rectangle{Width: 4, Height: 3}
	if math.Abs(r.Area()-12) > 1e-9 {
		t.Errorf("Area = %v, want 12", r.Area())
	}
	if math.Abs(r.Perimeter()-14) > 1e-9 {
		t.Errorf("Perimeter = %v, want 14", r.Perimeter())
	}
	if r.IsSquare() {
		t.Error("4x3 reported as a square")
	}
	if !(Rectangle{2, 2}).IsSquare() {
		t.Error("2x2 reported as not a square")
	}
}

func TestScaleMutates(t *testing.T) {
	r := Rectangle{4, 3}
	r.Scale(2.5)
	if math.Abs(r.Width-10) > 1e-9 || math.Abs(r.Height-7.5) > 1e-9 {
		t.Fatalf("after Scale(2.5): %+v, want {10 7.5}", r)
	}
}

func TestString(t *testing.T) {
	r := Rectangle{4, 3}
	if got := r.String(); got != "4.00x3.00" {
		t.Fatalf("String() = %q, want %q", got, "4.00x3.00")
	}
	if got := (Rectangle{1.5, 2}).String(); got != "1.50x2.00" {
		t.Fatalf("String() = %q, want %q", got, "1.50x2.00")
	}
}
`,
				}},
				Hints: []string{
					"§return r.Width * r.Height§ and §return 2 * (r.Width + r.Height)§.",
					"§Scale§: §r.Width *= factor§ and §r.Height *= factor§. The pointer receiver is what makes these stick.",
					"§String§: §fmt.Sprintf(\"%.2fx%.2f\", r.Width, r.Height)§.",
				},
				Solution: `package main

import "fmt"

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64 { return r.Width * r.Height }

func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

func (r Rectangle) IsSquare() bool { return r.Width == r.Height }

func (r Rectangle) String() string {
	return fmt.Sprintf("%.2fx%.2f", r.Width, r.Height)
}

func main() {
	r := Rectangle{4, 3}
	fmt.Println(r.Area(), r.Perimeter(), r.IsSquare())
	fmt.Println(r)
	r.Scale(2)
	fmt.Println(r)
}
`,
			},
			{
				ID:    "m1t15e2",
				Title: "A bank account with methods",
				Prompt: `Same behaviour as the pointers exercise, but expressed as methods. This is how
you will actually write it.

§§§go
type Account struct {
	Owner   string
	Balance int
}

func NewAccount(owner string, balance int) *Account
func (a *Account) Deposit(amount int) int
func (a *Account) Withdraw(amount int) (int, bool)
func (a Account) Summary() string
func (a *Account) Transfer(to *Account, amount int) bool
§§§

* §Deposit§ adds and returns the new balance.
* §Withdraw§ subtracts if there is enough money, returning the new balance and §true§;
  otherwise it changes nothing and returns the balance and §false§.
* §Summary§ returns §"<Owner>: <Balance>"§. Note the **value** receiver — it only reads,
  so a copy is fine and it works on both §Account§ and §*Account§.
* §Transfer§ moves money between two accounts. It must fail **atomically**: if the
  amount is not positive, if §to§ cannot receive, or if §a§ cannot pay, nothing changes
  on either side. Think about the order: check first, then mutate.

These are all pointer receivers except §Summary§ — a deliberate, and very common,
mix.
`,
				Starter: `package main

import "fmt"

type Account struct {
	Owner   string
	Balance int
}

func NewAccount(owner string, balance int) *Account {
	// TODO
	return nil
}

func (a *Account) Deposit(amount int) int {
	// TODO
	return 0
}

func (a *Account) Withdraw(amount int) (int, bool) {
	// TODO
	return 0, false
}

func (a Account) Summary() string {
	// TODO
	return ""
}

func (a *Account) Transfer(to *Account, amount int) bool {
	// TODO
	return false
}

func main() {
	ada := NewAccount("Ada", 100)
	bo := NewAccount("Bo", 0)
	fmt.Println(ada.Summary())
	ada.Deposit(50)
	fmt.Println(ada.Withdraw(500))
	fmt.Println(ada.Transfer(bo, 30))
	fmt.Println(ada.Summary(), bo.Summary())
}
`,
				Tests: []File{{
					Name: "methods_test.go",
					Content: `package main

import "testing"

func TestAccountMethods(t *testing.T) {
	a := NewAccount("Ada", 100)
	if a.Balance != 100 {
		t.Fatalf("NewAccount balance = %d, want 100", a.Balance)
	}
	if got := a.Summary(); got != "Ada: 100" {
		t.Errorf("Summary = %q, want %q", got, "Ada: 100")
	}
}

func TestDepositAndWithdraw(t *testing.T) {
	a := NewAccount("Ada", 100)

	if got := a.Deposit(50); got != 150 {
		t.Errorf("Deposit returned %d, want 150", got)
	}
	if a.Balance != 150 {
		t.Errorf("the account was not changed: %d", a.Balance)
	}

	if bal, ok := a.Withdraw(500); ok || bal != 150 {
		t.Errorf("Withdraw too much = %d, %v; want 150, false", bal, ok)
	}
	if a.Balance != 150 {
		t.Errorf("a failed withdrawal changed the balance: %d", a.Balance)
	}
	if bal, ok := a.Withdraw(50); !ok || bal != 100 {
		t.Errorf("Withdraw(50) = %d, %v; want 100, true", bal, ok)
	}
}

func TestTransfer(t *testing.T) {
	a := NewAccount("Ada", 100)
	b := NewAccount("Bo", 10)

	if !a.Transfer(b, 30) {
		t.Fatal("a valid transfer returned false")
	}
	if a.Balance != 70 || b.Balance != 40 {
		t.Fatalf("after transfer: %d, %d; want 70, 40", a.Balance, b.Balance)
	}
}

func TestTransferIsAtomic(t *testing.T) {
	a := NewAccount("Ada", 100)
	b := NewAccount("Bo", 0)

	if a.Transfer(b, 500) {
		t.Fatal("a transfer of more than the balance succeeded")
	}
	if a.Balance != 100 || b.Balance != 0 {
		t.Fatalf("a failed transfer changed something: %d, %d", a.Balance, b.Balance)
	}

	if a.Transfer(b, 0) {
		t.Fatal("a transfer of zero should not be accepted")
	}
	if a.Transfer(b, -10) {
		t.Fatal("a negative transfer should not be accepted")
	}
}

func TestSummaryWorksThroughPointer(t *testing.T) {
	a := NewAccount("Ada", 5)
	// Summary has a value receiver, so it is callable on a *Account too.
	if got := a.Summary(); got != "Ada: 5" {
		t.Errorf("Summary via pointer = %q", got)
	}
}
`,
				}},
				Hints: []string{
					"§NewAccount§ returns §&Account{Owner: owner, Balance: balance}§.",
					"§Deposit§: §a.Balance += amount; return a.Balance§.",
					"§Withdraw§: guard with §if a.Balance < amount§, then subtract.",
					"§Transfer§: check everything FIRST — §if amount <= 0 || a.Balance < amount { return false }§ — and only then mutate both sides.",
					"§Summary§: §fmt.Sprintf(\"%s: %d\", a.Owner, a.Balance)§.",
				},
				Solution: `package main

import "fmt"

type Account struct {
	Owner   string
	Balance int
}

func NewAccount(owner string, balance int) *Account {
	return &Account{Owner: owner, Balance: balance}
}

func (a *Account) Deposit(amount int) int {
	a.Balance += amount
	return a.Balance
}

func (a *Account) Withdraw(amount int) (int, bool) {
	if a.Balance < amount {
		return a.Balance, false
	}
	a.Balance -= amount
	return a.Balance, true
}

func (a Account) Summary() string {
	return fmt.Sprintf("%s: %d", a.Owner, a.Balance)
}

func (a *Account) Transfer(to *Account, amount int) bool {
	if amount <= 0 || a.Balance < amount {
		return false
	}
	a.Balance -= amount
	to.Balance += amount
	return true
}

func main() {
	ada := NewAccount("Ada", 100)
	bo := NewAccount("Bo", 0)
	fmt.Println(ada.Summary())
	ada.Deposit(50)
	fmt.Println(ada.Withdraw(500))
	fmt.Println(ada.Transfer(bo, 30))
	fmt.Println(ada.Summary(), bo.Summary())
}
`,
			},
			{
				ID:    "m1t15e3",
				Title: "Composition with embedding",
				Prompt: `Build a shared §Base§ and two entities that embed it.

§§§go
type Base struct {
	ID        int
	CreatedAt string
}

func (b Base) Describe() string   // "#<ID> created <CreatedAt>"
func (b Base) GetID() int

type User struct {
	Base
	Username string
}

func (u User) Describe() string   // Base.Describe() + " (user <Username>)"

type Product struct {
	Base
	Title string
	Price float64
}

func AllIDs(items ...interface{ GetID() int }) []int
§§§

* §Base.Describe§ returns e.g. §"#1 created 2024-01-01"§.
* §Base.GetID§ returns §b.ID§.
* §User.Describe§ **overrides** the promoted one and reuses §u.Base.Describe()§ so the
  logic is not duplicated.
* §Product§ does not override §Describe§, so it inherits §Base§'s version — the test
  checks exactly that — the defining behaviour of promotion.
* §AllIDs§ accepts anything with a §GetID() int§ method. This is your first look at an
  interface: §interface{ GetID() int }§ means "any type that has this method". Because
  §Base.GetID§ has a **value** receiver, it is promoted to §User§ and §Product§ as
  values, so both satisfy it.
`,
				Starter: `package main

import "fmt"

type Base struct {
	ID        int
	CreatedAt string
}

// Describe returns "#<ID> created <CreatedAt>".
func (b Base) Describe() string {
	// TODO
	return ""
}

// GetID returns the identifier.
func (b Base) GetID() int {
	// TODO
	return 0
}

type User struct {
	Base
	Username string
}

// Describe overrides Base.Describe.
func (u User) Describe() string {
	// TODO: reuse u.Base.Describe()
	return ""
}

type Product struct {
	Base
	Title string
	Price float64
}

// AllIDs collects the ID of anything with a GetID method.
func AllIDs(items ...interface{ GetID() int }) []int {
	// TODO
	return nil
}

func main() {
	u := User{Base{1, "2024-01-01"}, "ada"}
	p := Product{Base{2, "2024-02-02"}, "Keyboard", 99.0}
	fmt.Println(u.Describe())
	fmt.Println(p.Describe()) // promoted from Base
	fmt.Println(u.ID, p.CreatedAt)
	fmt.Println(AllIDs(u, p))
}
`,
				Tests: []File{{
					Name: "embed_test.go",
					Content: `package main

import (
	"reflect"
	"testing"
)

func TestPromotedFieldsAndMethods(t *testing.T) {
	u := User{Base{ID: 1, CreatedAt: "2024-01-01"}, "ada"}

	if u.ID != 1 {
		t.Errorf("promoted field u.ID = %d, want 1", u.ID)
	}
	if u.GetID() != 1 {
		t.Errorf("promoted method u.GetID() = %d, want 1", u.GetID())
	}
	if u.Base.GetID() != 1 {
		t.Errorf("explicit u.Base.GetID() = %d, want 1", u.Base.GetID())
	}

	p := Product{Base{ID: 2, CreatedAt: "2024-02-02"}, "Keyboard", 99}
	if p.GetID() != 2 {
		t.Errorf("promoted p.GetID() = %d, want 2", p.GetID())
	}
}

func TestDescribeOverrideAndPromotion(t *testing.T) {
	u := User{Base{ID: 1, CreatedAt: "2024-01-01"}, "ada"}
	if got, want := u.Describe(), "#1 created 2024-01-01 (user ada)"; got != want {
		t.Errorf("User.Describe() = %q, want %q", got, want)
	}

	p := Product{Base{ID: 2, CreatedAt: "2024-02-02"}, "Keyboard", 99}
	if got, want := p.Describe(), "#2 created 2024-02-02"; got != want {
		t.Errorf("Product.Describe() = %q, want %q (promoted from Base)", got, want)
	}

	if got, want := (Base{ID: 9, CreatedAt: "X"}).Describe(), "#9 created X"; got != want {
		t.Errorf("Base.Describe() = %q, want %q", got, want)
	}
}

func TestAllIDs(t *testing.T) {
	u := User{Base{ID: 1}, "ada"}
	p := Product{Base{ID: 2}, "Keyboard", 99}

	if got := AllIDs(u, p); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("AllIDs = %v, want [1 2]", got)
	}
	if got := AllIDs(); len(got) != 0 {
		t.Fatalf("AllIDs() = %v, want empty", got)
	}
}
`,
				}},
				Hints: []string{
					"§Base.Describe§: §fmt.Sprintf(\"#%d created %s\", b.ID, b.CreatedAt)§.",
					"§User.Describe§: §return u.Base.Describe() + \" (user \" + u.Username + \")\"§.",
					"§AllIDs§: §ids := make([]int, 0, len(items))§ then append §it.GetID()§ for each.",
					"Because §GetID§ has a value receiver, both §User§ and §Product§ satisfy §interface{ GetID() int }§ as values — no pointers needed.",
				},
				Solution: `package main

import "fmt"

type Base struct {
	ID        int
	CreatedAt string
}

func (b Base) Describe() string {
	return fmt.Sprintf("#%d created %s", b.ID, b.CreatedAt)
}

func (b Base) GetID() int { return b.ID }

type User struct {
	Base
	Username string
}

func (u User) Describe() string {
	return u.Base.Describe() + " (user " + u.Username + ")"
}

type Product struct {
	Base
	Title string
	Price float64
}

func AllIDs(items ...interface{ GetID() int }) []int {
	ids := make([]int, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.GetID())
	}
	return ids
}

func main() {
	u := User{Base{1, "2024-01-01"}, "ada"}
	p := Product{Base{2, "2024-02-02"}, "Keyboard", 99.0}
	fmt.Println(u.Describe())
	fmt.Println(p.Describe())
	fmt.Println(u.ID, p.CreatedAt)
	fmt.Println(AllIDs(u, p))
}
`,
			},
		},
	}
}

func m1t16() Topic {
	return Topic{
		ID:      "m1t16",
		Title:   "Interfaces and type assertions",
		Summary: "Implicit satisfaction, type assertions, and the nil-interface trap.",
		Theory: `
An **interface** specifies behaviour rather than data: a set of method signatures. Any
type whose method set contains them satisfies the interface. That is how Go gets
polymorphism without inheritance.

## Defining and satisfying an interface

§§§go
type Shape interface {
	Area() float64
	Name() string
}

type Circle struct{ Radius float64 }

func (c Circle) Area() float64 { return math.Pi * c.Radius * c.Radius }
func (c Circle) Name() string  { return "circle" }

type Square struct{ Side float64 }

func (s Square) Area() float64 { return s.Side * s.Side }
func (s Square) Name() string  { return "square" }
§§§

There is no §implements§ keyword. A type satisfies an interface by having the methods
the interface lists — *implicit*, or *structural*, satisfaction.

§§§go
shapes := []Shape{Circle{1}, Square{3}}
for _, s := range shapes {
	fmt.Printf("%s: %.2f\n", s.Name(), s.Area())
}
§§§

Output:

§§§text
circle: 3.14
square: 9.00
§§§

The slice holds §Shape§ values, but each one really is a §Circle§ or a §Square§. When we
call §s.Area()§, the correct method runs. That is polymorphism.

**Java comparison:** no §implements§, no abstract base class. And because satisfaction is
structural, you can make a type satisfy an interface *that someone else defined*, long
after the fact, without touching either of them.

## Where to define the interface

Go's proverb: **"accept interfaces, return structs."**

Describe the interface where it is *used*, not where the concrete type lives. A function
should ask for the smallest interface it needs:

§§§go
// Good: this function needs exactly one thing.
func CountBytes(r io.Reader) (int, error) {
	data, err := io.ReadAll(r)
	return len(data), err
}
§§§

That function works with a file, a network connection, a string, or a byte buffer —
anything with a §Read§ method — and it never had to know about any of them.

The standard library's most useful interfaces are tiny:

| Interface | Method | Satisfied by |
| --- | --- | --- |
| §error§ | §Error() string§ | every error type |
| §fmt.Stringer§ | §String() string§ | anything you want printed nicely |
| §io.Reader§ | §Read([]byte) (int, error)§ | files, sockets, buffers, §strings.Reader§ |
| §io.Writer§ | §Write([]byte) (int, error)§ | files, §bytes.Buffer§, §http.ResponseWriter§ |
| §http.Handler§ | §ServeHTTP(w, r)§ | every HTTP handler in Module 2 |

## The method set rule, revisited

This is where the pointer/value receiver decision comes back to bite:

| Type | Satisfies an interface requiring value receivers? | ...pointer receivers? |
| --- | --- | --- |
| §T§ | yes | **no** |
| §*T§ | yes | yes |

§§§go
type Speaker interface{ Speak() string }

type Dog struct{}

func (d *Dog) Speak() string { return "woof" } // pointer receiver

var s Speaker = Dog{}  // COMPILE ERROR: Dog does not implement Speaker
var s Speaker = &Dog{} // fine
§§§

That is why you will often see §[]*User§ rather than §[]User§ when the methods have
pointer receivers.

## The two-word interface value

An interface value holds a **pair**: a type and a value.

§§§text
+------------+-----------------+
| type: *Dog | value: 0xc000... |
+------------+-----------------+
§§§

This explains the most notorious trap in Go:

§§§go
type MyError struct{ msg string }

func (e *MyError) Error() string { return e.msg }

func broken() error {
	var e *MyError // nil pointer
	return e       // returns (type=*MyError, value=nil)
}

func main() {
	err := broken()
	fmt.Println(err == nil) // false!
}
§§§

The interface is **not nil**, because its type slot is filled in. It only *looks* nil.
The fix is to return a literal §nil§ when there is nothing to report:

§§§go
func fixed() error {
	var e *MyError
	if e == nil {
		return nil // an interface with both slots empty — the only true nil error
	}
	return e
}
§§§

This shows up again in Module 2, when a handler "returns no error" yet still reports
one.

## Type assertions: getting the concrete value back

An interface value hides the concrete type. To get it back, use a **type assertion**:

§§§go
var v any = "hello"
s := v.(string)     // panics if v is not a string
n, ok := v.(int)    // the safe form: ok is false if the type does not match
§§§

Always prefer the comma-ok form when you are not certain — a failed assertion panics.

§any§ (Go 1.18+) is an alias for §interface{}§: the interface with no methods, which
*every* type satisfies. Use it when you genuinely need to hold anything (like
§fmt.Println§ does), and not as a way to avoid designing types.

## Type switches: assertions for several types at once

§§§go
func describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("int %d", x) // x is an int here
	case string:
		return "string of length " + strconv.Itoa(len(x))
	case error:
		return "an error: " + x.Error()
	default:
		return fmt.Sprintf("%T", x)
	}
}
§§§

Inside a case naming exactly one type, §x§ has that concrete type. With several types in
one case (§case int, int64:§), §x§ stays §any§.

## A complete example

§§§go
package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
	Name() string
}

type Circle struct{ Radius float64 }
type Square struct{ Side float64 }

func (c Circle) Area() float64 { return math.Pi * c.Radius * c.Radius }
func (c Circle) Name() string  { return "circle" }
func (s Square) Area() float64 { return s.Side * s.Side }
func (s Square) Name() string  { return "square" }

func TotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

func main() {
	shapes := []Shape{Square{2}, Circle{1}}

	fmt.Printf("total: %.2f\n", TotalArea(shapes))
	fmt.Printf("first: %s\n", shapes[0].Name())

	// Downcast: the concrete type is still there if you need it.
	if sq, ok := shapes[0].(Square); ok {
		fmt.Println("side length:", sq.Side)
	}
}
§§§

Output:

§§§text
total: 7.14
first: square
side length: 2
§§§

## Design advice

* **Prefer many small interfaces to one big one.** The bigger the interface, the fewer
  types can satisfy it.
* **Do not create an interface "just in case".** The Go proverb again: "the bigger the
  interface, the weaker the abstraction."
* If a function needs only one method, ask for only that method — or define a local
  one-method interface. That makes it trivial to test.
`,
		Quiz: []Question{
			{
				ID:     "m1t16q1",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

type MyErr struct{}

func (e *MyErr) Error() string { return "boom" }

func bad() error {
	var e *MyErr
	return e
}

func main() {
	fmt.Println(bad() == nil)
}`,
				Options: []Option{
					{Text: "§false§", Correct: true, Rationale: "The interface holds (type=*MyErr, value=nil). A non-empty type slot makes the interface non-nil."},
					{Text: "§true§", Correct: false, Rationale: "A true nil interface needs both slots empty."},
					{Text: "It panics.", Correct: false, Rationale: "Nothing is dereferenced here."},
					{Text: "It does not compile.", Correct: false, Rationale: "It compiles — and that is what makes the trap nasty."},
				},
				Explanation: "The typed-nil-interface trap. Always return a literal nil when there is no error.",
			},
			{
				ID:     "m1t16q2",
				Prompt: "Which types satisfy §interface{ Speak() string }§? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "§type A struct{}; func (a A) Speak() string { return \"a\" }§ as an §A§", Correct: true, Rationale: "A value receiver means both §A§ and §*A§ satisfy the interface."},
					{Text: "§type B struct{}; func (b *B) Speak() string { return \"b\" }§ as a §B§", Correct: false, Rationale: "The method set of §B§ excludes pointer-receiver methods."},
					{Text: "§type B struct{}; func (b *B) Speak() string { return \"b\" }§ as a §*B§", Correct: true, Rationale: "The pointer's method set includes it."},
					{Text: "§func (i int) Speak() string { return \"int\" }§ added to the predeclared §int§", Correct: false, Rationale: "You may only declare methods on a named type defined in the same package. §int§ is not yours."},
					{Text: "§type MyInt int§ with §func (i MyInt) Speak() string { return \"myint\" }§", Correct: true, Rationale: "§MyInt§ is a named type in your package, so methods are allowed."},
				},
				Explanation: "Method sets: §T§ gets value receivers only; §*T§ gets both. And methods live in the package that defines the type.",
			},
			{
				ID:     "m1t16q3",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

func describe(v any) string {
	switch x := v.(type) {
	case int:
		return fmt.Sprintf("int %d", x*2)
	case string:
		return "str " + x
	default:
		return fmt.Sprintf("%T", x)
	}
}

func main() {
	fmt.Println(describe(21))
	fmt.Println(describe("hi"))
	fmt.Println(describe(3.5))
}`,
				Options: []Option{
					{Text: "§int 42§ / §str hi§ / §float64§", Correct: true, Rationale: "In each single-type case §x§ has that concrete type, so §x*2§ works on the int. The default reports the type."},
					{Text: "§int 42§ / §str hi§ / §<nil>§", Correct: false, Rationale: "§x§ in §default§ is the original §any§ value, so §%T§ prints §float64§."},
					{Text: "It does not compile: §x*2§ where §x§ is §any§.", Correct: false, Rationale: "The type switch narrows §x§ to §int§ inside that case."},
					{Text: "§int 21§ / §str hi§ / §float64§", Correct: false, Rationale: "The case body multiplies by 2."},
				},
				Explanation: "A type switch gives you the concrete type inside each single-type case. In §default§ you keep §any§.",
			},
			{
				ID:     "m1t16q4",
				Prompt: "Why does the standard library prefer tiny interfaces like §io.Reader§?",
				Options: []Option{
					{Text: "Because the fewer methods an interface requires, the more types can satisfy it — so the function works with more inputs.", Correct: true, Rationale: "A one-method interface is satisfied by files, sockets, buffers, strings and any custom type you like."},
					{Text: "Because large interfaces are slower.", Correct: false, Rationale: "Method dispatch cost is the same."},
					{Text: "Because Go limits interfaces to three methods.", Correct: false, Rationale: "There is no such limit."},
					{Text: "Because small interfaces cannot be satisfied by pointers.", Correct: false, Rationale: "Pointers satisfy them fine."},
				},
				Explanation: "\"The bigger the interface, the weaker the abstraction.\" Ask for the minimum you actually need.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t16e1",
				Title: "Warm-up: shapes behind an interface",
				Prompt: `§§§go
type Shape interface {
	Area() float64
	Name() string
}

type Circle struct{ Radius float64 }
type Square struct{ Side float64 }

func TotalArea(shapes []Shape) float64
func Largest(shapes []Shape) Shape
func DescribeAll(shapes []Shape) []string
§§§

* §Circle.Area§ is §math.Pi * Radius * Radius§ and its name is §"circle"§.
* §Square.Area§ is §Side * Side§ and its name is §"square"§.
* §TotalArea§ adds up every area; an empty slice gives §0§.
* §Largest§ returns the shape with the biggest area, or **§nil§** for an empty slice.
  Returning §nil§ here takes care: returning a typed nil would be a non-nil interface
  (the trap from the theory section).
* §DescribeAll§ returns one string per shape, formatted §"<name>: <area to 2 dp>"§:
  §["square: 4.00" "circle: 3.14"]§.

Both concrete types should use **value** receivers, so plain values satisfy §Shape§.
`,
				Starter: `package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
	Name() string
}

type Circle struct{ Radius float64 }

func (c Circle) Area() float64 { return math.Pi * c.Radius * c.Radius }
func (c Circle) Name() string  { return "circle" }

type Square struct{ Side float64 }

func (s Square) Area() float64 {
	// TODO
	return 0
}

func (s Square) Name() string {
	// TODO
	return ""
}

// TotalArea sums the area of every shape.
func TotalArea(shapes []Shape) float64 {
	// TODO
	return 0
}

// Largest returns the shape with the greatest area, or nil.
func Largest(shapes []Shape) Shape {
	// TODO
	return nil
}

// DescribeAll returns "name: area" for each shape.
func DescribeAll(shapes []Shape) []string {
	// TODO
	return nil
}

func main() {
	shapes := []Shape{Circle{1}, Square{3}}
	fmt.Println(TotalArea(shapes))
	fmt.Println(Largest(shapes).Name())
	fmt.Println(DescribeAll(shapes))
}
`,
				Tests: []File{{
					Name: "shape_test.go",
					Content: `package main

import (
	"math"
	"testing"
)

func TestTotalArea(t *testing.T) {
	if got := TotalArea([]Shape{Square{2}, Square{3}}); math.Abs(got-13) > 1e-9 {
		t.Errorf("TotalArea = %v, want 13", got)
	}
	if got := TotalArea(nil); got != 0 {
		t.Errorf("TotalArea(nil) = %v, want 0", got)
	}
}

func TestLargest(t *testing.T) {
	small := Square{2}
	big := Circle{Radius: 10}
	if got := Largest([]Shape{small, big}); got.Name() != "circle" {
		t.Errorf("Largest returned %q, want circle", got.Name())
	}
	if got := Largest([]Shape{small}); got.Name() != "square" {
		t.Errorf("Largest(single) = %q", got.Name())
	}
	if got := Largest(nil); got != nil {
		t.Errorf("Largest(nil) = %v, want a true nil interface", got)
	}
}

func TestDescribeAll(t *testing.T) {
	got := DescribeAll([]Shape{Square{2}, Circle{Radius: 1}})
	want := []string{"square: 4.00", "circle: 3.14"}
	if len(got) != len(want) {
		t.Fatalf("DescribeAll = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("DescribeAll = %v, want %v", got, want)
		}
	}
	if len(DescribeAll(nil)) != 0 {
		t.Error("DescribeAll(nil) should be empty")
	}
}
`,
				}},
				Hints: []string{
					"§Square.Area§ is §s.Side * s.Side§ and §Square.Name§ returns §\"square\"§.",
					"§TotalArea§: the usual accumulator loop over the shapes.",
					"§Largest§: guard §len(shapes) == 0§ and §return nil§ — the literal nil, not a typed nil variable.",
					"§DescribeAll§: §fmt.Sprintf(\"%s: %.2f\", s.Name(), s.Area())§.",
				},
				Solution: `package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
	Name() string
}

type Circle struct{ Radius float64 }

func (c Circle) Area() float64 { return math.Pi * c.Radius * c.Radius }
func (c Circle) Name() string  { return "circle" }

type Square struct{ Side float64 }

func (s Square) Area() float64 { return s.Side * s.Side }
func (s Square) Name() string  { return "square" }

func TotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

func Largest(shapes []Shape) Shape {
	if len(shapes) == 0 {
		return nil
	}
	best := shapes[0]
	for _, s := range shapes[1:] {
		if s.Area() > best.Area() {
			best = s
		}
	}
	return best
}

func DescribeAll(shapes []Shape) []string {
	out := make([]string, 0, len(shapes))
	for _, s := range shapes {
		out = append(out, fmt.Sprintf("%s: %.2f", s.Name(), s.Area()))
	}
	return out
}

func main() {
	shapes := []Shape{Circle{1}, Square{3}}
	fmt.Println(TotalArea(shapes))
	fmt.Println(Largest(shapes).Name())
	fmt.Println(DescribeAll(shapes))
}
`,
			},
			{
				ID:    "m1t16e2",
				Title: "Make your own Stringer",
				Prompt: `§fmt§ has a tiny interface it looks for whenever it prints a value:

§§§go
type Stringer interface {
	String() string
}
§§§

If your type has a §String() string§ method, §fmt.Println§, §fmt.Printf("%v")§ and
§fmt.Sprintf§ all use it instead of printing the raw fields.

§§§go
type Temperature struct {
	Celsius float64
}

func (t Temperature) String() string
func (t Temperature) Fahrenheit() float64
func Hottest(ts []Temperature) (Temperature, bool)
§§§

* §String§ renders §"21.5°C (70.7°F)"§ — °C to one decimal, °F to one decimal.
* §Fahrenheit§ is §Celsius*9/5 + 32§.
* §Hottest§ returns the warmest temperature and §true§, or the zero temperature and
  §false§ for an empty slice.

The test checks §fmt.Sprint(t)§, which is the exact call that makes §Stringer§ matter.
`,
				Starter: `package main

import "fmt"

type Temperature struct {
	Celsius float64
}

// String renders the temperature in both scales.
func (t Temperature) String() string {
	// TODO
	return ""
}

// Fahrenheit converts to the other scale.
func (t Temperature) Fahrenheit() float64 {
	// TODO
	return 0
}

// Hottest returns the warmest temperature, and whether there was one.
func Hottest(ts []Temperature) (Temperature, bool) {
	// TODO
	return Temperature{}, false
}

func main() {
	t := Temperature{Celsius: 21.5}
	fmt.Println(t)          // uses String()
	fmt.Println(Hottest([]Temperature{{10}, {21.5}, {-5}}))
}
`,
				Tests: []File{{
					Name: "temperature_test.go",
					Content: `package main

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	cases := []struct {
		c    float64
		want string
	}{
		{21.5, "21.5°C (70.7°F)"},
		{0, "0.0°C (32.0°F)"},
		{100, "100.0°C (212.0°F)"},
	}
	for _, c := range cases {
		if got := (Temperature{Celsius: c.c}).String(); got != c.want {
			t.Errorf("String() = %q, want %q", got, c.want)
		}
	}
}

func TestFmtUsesStringer(t *testing.T) {
	tv := Temperature{Celsius: 0}
	if got := fmt.Sprint(tv); got != "0.0°C (32.0°F)" {
		t.Fatalf("fmt.Sprint = %q; is your String method on a value receiver?", got)
	}
}

func TestFahrenheit(t *testing.T) {
	if got := (Temperature{Celsius: 100}).Fahrenheit(); got != 212 {
		t.Errorf("Fahrenheit = %v, want 212", got)
	}
	if got := (Temperature{Celsius: -40}).Fahrenheit(); got != -40 {
		t.Errorf("Fahrenheit = %v, want -40", got)
	}
}

func TestHottest(t *testing.T) {
	got, ok := Hottest([]Temperature{{10}, {21.5}, {-5}})
	if !ok || got.Celsius != 21.5 {
		t.Fatalf("Hottest = %+v, %v", got, ok)
	}
	if _, ok := Hottest(nil); ok {
		t.Error("Hottest(nil) reported ok")
	}
}
`,
				}},
				Hints: []string{
					"§Fahrenheit§: §return t.Celsius*9/5 + 32§.",
					"§String§: §fmt.Sprintf(\"%.1f°C (%.1f°F)\", t.Celsius, t.Fahrenheit())§.",
					"§Hottest§: guard the empty slice, seed from the first element, compare with §>§.",
					"Keep §String§ on a **value** receiver so both §Temperature§ and §*Temperature§ print nicely.",
				},
				Solution: `package main

import "fmt"

type Temperature struct {
	Celsius float64
}

func (t Temperature) String() string {
	return fmt.Sprintf("%.1f°C (%.1f°F)", t.Celsius, t.Fahrenheit())
}

func (t Temperature) Fahrenheit() float64 {
	return t.Celsius*9/5 + 32
}

func Hottest(ts []Temperature) (Temperature, bool) {
	if len(ts) == 0 {
		return Temperature{}, false
	}
	best := ts[0]
	for _, t := range ts[1:] {
		if t.Celsius > best.Celsius {
			best = t
		}
	}
	return best, true
}

func main() {
	t := Temperature{Celsius: 21.5}
	fmt.Println(t)
	fmt.Println(Hottest([]Temperature{{10}, {21.5}, {-5}}))
}
`,
			},
			{
				ID:    "m1t16e3",
				Title: "A payment processor, one method at a time",
				Prompt: `This is the shape you will meet constantly in backend work: a single-method
interface, several implementations, and code that does not care which one it has.

§§§go
type Processor interface {
	Pay(amount int) (string, error)
	Label() string
}

type Card struct{ Last4 string }
type BankTransfer struct{ IBAN string }
type Fake struct{ Fail bool }

func (c Card) Pay(amount int) (string, error)
func (b BankTransfer) Pay(amount int) (string, error)
func (f Fake) Pay(amount int) (string, error)

func Checkout(p Processor, amount int) (string, error)
func CheckoutAll(ps []Processor, amount int) []string
§§§

Rules:

* §Card.Pay§ returns §"charged <amount> to card ****<Last4>"§.
* §BankTransfer.Pay§ returns §"transferred <amount> to <IBAN>"§.
* §Fake.Pay§ returns an error when §Fail§ is true, and §"paid <amount>"§ otherwise.
  Errors are made with §errors.New§.
* §Checkout§ rejects §amount <= 0§ with its own error before calling §p.Pay§, then
  returns whatever §p.Pay§ returned.
* §CheckoutAll§ calls §Checkout§ for each processor and returns the results, using the
  error text in place of a result when one fails. With a §nil§ slice it returns an empty
  slice.
* §Label§ just returns §"card"§, §"bank"§ or §"fake"§.

Note how §Checkout§ takes the *interface* and has never heard of §Card§,
§BankTransfer§ or §Fake§. That is the payoff.
`,
				Starter: `package main

import (
	"errors"
	"fmt"
)

type Processor interface {
	Pay(amount int) (string, error)
	Label() string
}

type Card struct{ Last4 string }

func (c Card) Pay(amount int) (string, error) {
	// TODO
	return "", nil
}

func (c Card) Label() string { return "card" }

type BankTransfer struct{ IBAN string }

func (b BankTransfer) Pay(amount int) (string, error) {
	// TODO
	return "", nil
}

func (b BankTransfer) Label() string { return "bank" }

type Fake struct{ Fail bool }

func (f Fake) Pay(amount int) (string, error) {
	// TODO
	return "", nil
}

func (f Fake) Label() string { return "fake" }

// Checkout validates the amount and delegates to the processor.
func Checkout(p Processor, amount int) (string, error) {
	// TODO
	return "", nil
}

// CheckoutAll runs Checkout for each processor.
func CheckoutAll(ps []Processor, amount int) []string {
	// TODO
	return nil
}

func main() {
	fmt.Println(Checkout(Card{Last4: "4242"}, 1000))
	fmt.Println(CheckoutAll([]Processor{Card{Last4: "4242"}, Fake{Fail: true}}, 50))

	// errors.New is imported for you.
	_ = errors.New
}
`,
				Tests: []File{{
					Name: "processor_test.go",
					Content: `package main

import "testing"

func TestCardPay(t *testing.T) {
	got, err := Card{Last4: "4242"}.Pay(1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "charged 1000 to card ****4242" {
		t.Fatalf("Card.Pay = %q", got)
	}
}

func TestBankPay(t *testing.T) {
	got, err := BankTransfer{IBAN: "NO93"}.Pay(250)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "transferred 250 to NO93" {
		t.Fatalf("BankTransfer.Pay = %q", got)
	}
}

func TestFakePay(t *testing.T) {
	got, err := Fake{}.Pay(10)
	if err != nil || got != "paid 10" {
		t.Fatalf("Fake.Pay = %q, %v", got, err)
	}
	if _, err := (Fake{Fail: true}).Pay(10); err == nil {
		t.Fatal("Fake{Fail:true}.Pay should return an error")
	}
}

func TestCheckoutValidates(t *testing.T) {
	for _, amount := range []int{0, -5} {
		if _, err := Checkout(Card{Last4: "4242"}, amount); err == nil {
			t.Errorf("Checkout(amount=%d) should be rejected before reaching the processor", amount)
		}
	}
}

func TestCheckoutPropagatesErrors(t *testing.T) {
	if _, err := Checkout(Fake{Fail: true}, 10); err == nil {
		t.Fatal("Checkout swallowed the processor's error")
	}
}

func TestCheckoutAll(t *testing.T) {
	ps := []Processor{
		Card{Last4: "1111"},
		Fake{Fail: true},
		BankTransfer{IBAN: "X"},
	}
	got := CheckoutAll(ps, 100)
	if len(got) != 3 {
		t.Fatalf("CheckoutAll returned %d results, want 3: %v", len(got), got)
	}
	if got[0] != "charged 100 to card ****1111" {
		t.Errorf("result 0 = %q", got[0])
	}
	if got[2] != "transferred 100 to X" {
		t.Errorf("result 2 = %q", got[2])
	}
	if got[1] == "" {
		t.Error("a failed processor should still produce the error text")
	}
	if len(CheckoutAll(nil, 10)) != 0 {
		t.Error("CheckoutAll(nil) should be empty")
	}
}

func TestInterfaceHoldsAllThree(t *testing.T) {
	var ps []Processor = []Processor{Card{Last4: "1"}, BankTransfer{IBAN: "2"}, Fake{}}
	for _, p := range ps {
		if p.Label() == "" {
			t.Errorf("%T has an empty label", p)
		}
	}
}
`,
				}},
				Hints: []string{
					"§Card.Pay§: §fmt.Sprintf(\"charged %d to card ****%s\", amount, c.Last4)§.",
					"§BankTransfer.Pay§: §fmt.Sprintf(\"transferred %d to %s\", amount, b.IBAN)§.",
					"§Fake.Pay§: §if f.Fail { return \"\", errors.New(\"fake failure\") }§ then §return fmt.Sprintf(\"paid %d\", amount), nil§.",
					"§Checkout§: §if amount <= 0 { return \"\", errors.New(\"amount must be positive\") }§, then §return p.Pay(amount)§.",
					"§CheckoutAll§: a loop with §if res, err := Checkout(p, amount); err != nil { out = append(out, err.Error()) } else { out = append(out, res) }§.",
				},
				Solution: `package main

import (
	"errors"
	"fmt"
)

type Processor interface {
	Pay(amount int) (string, error)
	Label() string
}

type Card struct{ Last4 string }

func (c Card) Pay(amount int) (string, error) {
	return fmt.Sprintf("charged %d to card ****%s", amount, c.Last4), nil
}

func (c Card) Label() string { return "card" }

type BankTransfer struct{ IBAN string }

func (b BankTransfer) Pay(amount int) (string, error) {
	return fmt.Sprintf("transferred %d to %s", amount, b.IBAN), nil
}

func (b BankTransfer) Label() string { return "bank" }

type Fake struct{ Fail bool }

func (f Fake) Pay(amount int) (string, error) {
	if f.Fail {
		return "", errors.New("fake failure")
	}
	return fmt.Sprintf("paid %d", amount), nil
}

func (f Fake) Label() string { return "fake" }

func Checkout(p Processor, amount int) (string, error) {
	if amount <= 0 {
		return "", errors.New("amount must be positive")
	}
	return p.Pay(amount)
}

func CheckoutAll(ps []Processor, amount int) []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		res, err := Checkout(p, amount)
		if err != nil {
			out = append(out, err.Error())
			continue
		}
		out = append(out, res)
	}
	return out
}

func main() {
	fmt.Println(Checkout(Card{Last4: "4242"}, 1000))
	fmt.Println(CheckoutAll([]Processor{Card{Last4: "4242"}, Fake{Fail: true}}, 50))
}
`,
			},
		},
	}
}

func m1t17() Topic {
	return Topic{
		ID:      "m1t17",
		Title:   "Errors, wrapping and panic",
		Summary: "Errors are values. Inspecting them, adding context, and when to give up entirely.",
		Theory: `
Go has no exceptions — no §try§, no §catch§, no §throw§. An error is an ordinary value
returned by a function, and the caller decides how to handle it.

§Cannot call it "a value" without showing one:§

§§§go
type error interface {
	Error() string
}
§§§

§error§ is just an interface with one method. Anything with an §Error() string§ method
is an error.

## Creating errors

§§§go
errors.New("something went wrong")
fmt.Errorf("could not open %s", filename)
fmt.Errorf("could not open %s: %w", filename, err) // %w wraps
§§§

§fmt.Errorf§ is §fmt.Sprintf§ for errors: the same verbs, plus §%w§.

## The conventions

1. **The error is the last return value.** §func f() (Result, error)§.
2. **Return §nil§ for success** — an untyped §nil§, not a typed nil (you know why from
   the interfaces topic).
3. **Handle errors immediately.** No deferred error queues.
4. **Add context when you wrap**, but do not just wrap endlessly — the message should
   explain *what you were trying to do*.

§§§go
func LoadUser(id int) (*User, error) {
	u, err := store.Get(id)
	if err != nil {
		return nil, fmt.Errorf("load user %d: %w", id, err)
	}
	return u, nil
}
§§§

Now an error from deep inside the storage layer reads like a trail:
§load user 7: get from database: connection refused§.

## Sentinel errors

A **sentinel error** is a package-level value that callers compare against:

§§§go
var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)
§§§

Naming convention: §ErrSomething§, exported, documented. They are part of your API. The
standard library has §os.ErrNotExist§, §sql.ErrNoRows§, §io.EOF§ and friends.

## Custom error types

When an error needs to carry data, use a struct:

§§§go
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Reason)
}
§§§

Return it as a **pointer** — §&ValidationError{…}§. (A pointer to a struct with an
§Error§ method is what satisfies the interface.)

## Inspecting errors: Is and As

Never compare error strings, and never use §==§ on a wrapped error. Use these two:

§§§go
if errors.Is(err, ErrNotFound) { … }        // walks the chain looking for a match

var ve *ValidationError
if errors.As(err, &ve) { … }                // finds a matching TYPE in the chain
§§§

§errors.Is§ answers "is this (or anything it wraps) that sentinel?" §errors.As§ answers
"is there a value of this type anywhere in the chain, and if so, please put it in my
variable".

A demonstration:

§§§go
package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type ValidationError struct {
	Field string
}

func (e *ValidationError) Error() string { return "invalid " + e.Field }

func lookup(id int) error {
	if id == 0 {
		return fmt.Errorf("lookup %d: %w", id, ErrNotFound) // wrapped
	}
	if id < 0 {
		return &ValidationError{Field: "id"}
	}
	return nil
}

func main() {
	err := lookup(0)
	fmt.Println(err)                       // lookup 0: not found
	fmt.Println(errors.Is(err, ErrNotFound)) // true

	err = lookup(-1)
	var ve *ValidationError
	fmt.Println(errors.As(err, &ve))       // true
	fmt.Println(ve.Field)                  // id

	fmt.Println(errors.Is(lookup(1), ErrNotFound)) // false, no error at all
}
§§§

Output:

§§§text
lookup 0: not found
true
true
id
false
§§§

Note §errors.Is(lookup(1), ErrNotFound)§ is §false§, not a panic: §Is§ on a §nil§ error
is always §false§.

## §%w§ versus §%v§

This distinction matters and is easy to get wrong:

| Verb | Effect | §errors.Is§ still finds the sentinel? |
| --- | --- | --- |
| §%w§ | wraps: the original error becomes part of the chain | **yes** |
| §%v§ | formats: the original becomes just text | no |
| §%s§ | the same as §%v§ for errors | no |

It is common to break an error chain by accident with §%v§. When the caller might want
to test the error, use §%w§.

## §errors.Join§

To report several problems at once:

§§§go
err := errors.Join(validateName(), validateEmail())
if err != nil {
	fmt.Println(err) // one line per error
}
§§§

§Join§ produces an error that §Is§ and §As§ search through for every element.

## §panic§ and §recover§

A §panic§ aborts the current goroutine and unwinds the stack, running deferred calls on
the way. It is for **programmer errors**: a broken invariant, an impossible state, a nil
map write, an index out of range. It is not for "the file was missing" or "the user
typed nonsense".

§§§go
func mustPositive(n int) int {
	if n <= 0 {
		panic(fmt.Sprintf("n must be positive, got %d", n))
	}
	return n
}
§§§

§recover§ stops a panic, but only from inside a deferred function:

§§§go
func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from %v", r)
		}
	}()
	return a / b, nil // panics when b == 0
}
§§§

Three rules for §recover§:

1. It only works inside a **deferred** call.
2. It only recovers panics **in the same goroutine**. A panic in a goroutine with no
   recovery takes the whole program down.
3. Panics should not cross package boundaries. Convert them at the boundary — that is
   exactly what a web server does with a "recovery middleware" in Module 2.

## When to use which

| Situation | Tool |
| --- | --- |
| Anything the caller could plausibly handle | return an §error§ |
| "This should never happen" / broken invariant | §panic§ |
| Failure during §init()§ — the program cannot run | §panic§ |
| Bad user input | return an §error§ |
| Index out of range, nil map write | the runtime panics for you |

## The style that Go programmers expect

Guard clauses and early returns, so the happy path stays level:

§§§go
func Process(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	return validate(data)
}
§§§

Every error is handled where it happens, and the main flow of the function is the
columns on the left.
`,
		Quiz: []Question{
			{
				ID:     "m1t17q1",
				Prompt: "Which of these correctly detects a wrapped §ErrNotFound§? Select all that apply.",
				Multi:  true,
				Code: `base := ErrNotFound
err1 := fmt.Errorf("loading user: %w", base)
err2 := fmt.Errorf("loading user: %v", base)`,
				Options: []Option{
					{Text: "§errors.Is(err1, ErrNotFound)§", Correct: true, Rationale: "§%w§ preserves the chain, so §Is§ finds the sentinel."},
					{Text: "§errors.Is(err2, ErrNotFound)§", Correct: false, Rationale: "§%v§ flattened the error to a string; the chain is gone."},
					{Text: "§err1 == ErrNotFound§", Correct: false, Rationale: "§==§ does not look inside a wrapped error."},
					{Text: "§strings.Contains(err1.Error(), \"not found\")§", Correct: false, Rationale: "It happens to work, and breaks the moment anybody rewords the message. Never do this."},
				},
				Explanation: "Wrap with §%w§, inspect with §errors.Is§ and §errors.As§ — never string matching.",
			},
			{
				ID:     "m1t17q2",
				Prompt: "What does this print?",
				Code: `package main

import (
	"errors"
	"fmt"
)

type ValidationError struct{ Field string }

func (e *ValidationError) Error() string { return "invalid " + e.Field }

func main() {
	err := fmt.Errorf("outer: %w", &ValidationError{Field: "age"})

	var ve *ValidationError
	fmt.Println(errors.As(err, &ve), ve.Field)
	fmt.Println(errors.Is(err, nil))
}`,
				Options: []Option{
					{Text: "§true age§ and §false§", Correct: true, Rationale: "§As§ finds the wrapped §*ValidationError§ and fills §ve§. §Is(nonNilErr, nil)§ is false."},
					{Text: "§false <nil>§ and §false§", Correct: false, Rationale: "§As§ does look through wrapped errors."},
					{Text: "§true age§ and §true§", Correct: false, Rationale: "§errors.Is(err, nil)§ is true only when §err§ itself is nil."},
					{Text: "It panics because §ve§ is nil.", Correct: false, Rationale: "§As§ assigns into §ve§ only on a match, and it matched."},
				},
				Explanation: "§errors.As§ walks the wrap chain looking for a type. §errors.Is(err, nil)§ is a nil-check.",
			},
			{
				ID:     "m1t17q3",
				Prompt: "What happens?",
				Code: `package main

import "fmt"

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered:", r)
		}
	}()
	nums := []int{1, 2}
	fmt.Println(nums[5])
}`,
				Options: []Option{
					{Text: "It prints §recovered: runtime error: index out of range [5] with length 2§ and then exits normally.", Correct: true, Rationale: "§recover§ inside the deferred function stops the panic and lets the program continue."},
					{Text: "The program crashes with a panic.", Correct: false, Rationale: "The deferred §recover§ catches it."},
					{Text: "It prints nothing and exits silently.", Correct: false, Rationale: "The deferred function prints."},
					{Text: "It does not compile.", Correct: false, Rationale: "An out-of-range index is a runtime panic, not a compile error."},
				},
				Explanation: "§recover()§ must be called directly inside a deferred function; it returns the panic value and stops the unwinding.",
			},
			{
				ID:     "m1t17q4",
				Prompt: "Which of these are appropriate uses of §panic§? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "A missing required environment variable detected at startup.", Correct: true, Rationale: "The program cannot run correctly; failing fast is idiomatic."},
					{Text: "A user typed a bad date into a form.", Correct: false, Rationale: "Expected failure — return an error."},
					{Text: "A §switch§ over an enum hits a case that cannot happen.", Correct: true, Rationale: "A broken invariant is exactly what panic is for."},
					{Text: "A database query found no rows.", Correct: false, Rationale: "Expected — return a sentinel error like §ErrNotFound§."},
					{Text: "A test helper cannot build its fixture.", Correct: true, Rationale: "In test code, failing fast on a broken fixture is normal and helpful."},
				},
				Explanation: "Panic for programmer errors and impossible states; return errors for anything a caller can reasonably handle.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t17e1",
				Title: "Warm-up: divide with a sentinel error",
				Prompt: `§§§go
var ErrDivideByZero = errors.New("cannot divide by zero")

func Divide(a, b float64) (float64, error)
§§§

* Returns §a / b§ and §nil§.
* Returns §0§ and §ErrDivideByZero§ when §b§ is zero.

The test compares with §errors.Is§, so the sentinel must be the exact value you return
(not a new error with the same text).
`,
				Starter: `package main

import (
	"errors"
	"fmt"
)

var ErrDivideByZero = errors.New("cannot divide by zero")

// Divide returns a/b, or ErrDivideByZero when b is zero.
func Divide(a, b float64) (float64, error) {
	// TODO
	return 0, nil
}

func main() {
	fmt.Println(Divide(10, 4))
	fmt.Println(Divide(1, 0))
}
`,
				Tests: []File{{
					Name: "divide_test.go",
					Content: `package main

import (
	"errors"
	"math"
	"testing"
)

func TestDivide(t *testing.T) {
	got, err := Divide(10, 4)
	if err != nil {
		t.Fatalf("Divide(10, 4) returned an error: %v", err)
	}
	if math.Abs(got-2.5) > 1e-9 {
		t.Fatalf("Divide(10, 4) = %v, want 2.5", got)
	}
}

func TestDivideByZero(t *testing.T) {
	got, err := Divide(1, 0)
	if !errors.Is(err, ErrDivideByZero) {
		t.Fatalf("Divide(1, 0) error = %v, want ErrDivideByZero", err)
	}
	if got != 0 {
		t.Fatalf("Divide(1, 0) value = %v, want 0", got)
	}
}

func TestDivideNegative(t *testing.T) {
	got, err := Divide(-9, 3)
	if err != nil || math.Abs(got+3) > 1e-9 {
		t.Fatalf("Divide(-9, 3) = %v, %v; want -3, nil", got, err)
	}
}
`,
				}},
				Hints: []string{
					"Guard first: §if b == 0 { return 0, ErrDivideByZero }§.",
					"Then §return a / b, nil§.",
					"Return the sentinel itself, not §errors.New(\"cannot divide by zero\")§ — a fresh error would not be §Is§-equal to the sentinel.",
				},
				Solution: `package main

import (
	"errors"
	"fmt"
)

var ErrDivideByZero = errors.New("cannot divide by zero")

func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return a / b, nil
}

func main() {
	fmt.Println(Divide(10, 4))
	fmt.Println(Divide(1, 0))
}
`,
			},
			{
				ID:    "m1t17e2",
				Title: "A store that wraps and a validator that types",
				Prompt: `Two kinds of error in one package — a sentinel and a custom type.

§§§go
var ErrNotFound = errors.New("not found")
var ErrBadInput = errors.New("bad input")

type User struct {
	ID   int
	Name string
}

type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string

type Store struct {
	users map[int]User
}

func NewStore() *Store
func (s *Store) Add(u User) error
func (s *Store) Get(id int) (User, error)
func (s *Store) Delete(id int) error
§§§

* §Add§ returns a §*ValidationError§ when §ID <= 0§ (field §"id"§) or §Name == ""§
  (field §"name"§). The test extracts it with §errors.As§, so it must be a pointer to
  the type.
* §Get§ and §Delete§ return an error that **wraps** §ErrNotFound§ with §%w§ and mentions
  the id. The test checks the error is *not* the bare sentinel (so it must be wrapped)
  and that §errors.Is§ still finds it.
* §Error§ on the validation type renders §"invalid <Field>: <Reason>"§.

This is the standard error toolkit of a real Go package.
`,
				Starter: `package main

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("not found")
	ErrBadInput = errors.New("bad input")
)

type User struct {
	ID   int
	Name string
}

// ValidationError carries which field was wrong and why.
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	// TODO
	return ""
}

type Store struct {
	users map[int]User
}

func NewStore() *Store {
	// TODO
	return nil
}

func (s *Store) Add(u User) error {
	// TODO
	return nil
}

func (s *Store) Get(id int) (User, error) {
	// TODO
	return User{}, nil
}

func (s *Store) Delete(id int) error {
	// TODO
	return nil
}

func main() {
	s := NewStore()
	fmt.Println(s.Add(User{ID: 1, Name: "Ada"}))
	fmt.Println(s.Add(User{ID: 0}))
	_, err := s.Get(99)
	fmt.Println(err, errors.Is(err, ErrNotFound))
}
`,
				Tests: []File{{
					Name: "store_test.go",
					Content: `package main

import (
	"errors"
	"strings"
	"testing"
)

func TestAddValidation(t *testing.T) {
	s := NewStore()

	var ve *ValidationError
	if err := s.Add(User{ID: 0, Name: "x"}); !errors.As(err, &ve) {
		t.Fatalf("Add(id=0) = %v, want a *ValidationError", err)
	} else if ve.Field != "id" {
		t.Errorf("Field = %q, want \"id\"", ve.Field)
	}

	ve = nil
	if err := s.Add(User{ID: 1, Name: ""}); !errors.As(err, &ve) {
		t.Fatalf("Add(empty name) = %v, want a *ValidationError", err)
	} else if ve.Field != "name" {
		t.Errorf("Field = %q, want \"name\"", ve.Field)
	}

	if err := s.Add(User{ID: 1, Name: "Ada"}); err != nil {
		t.Fatalf("Add(valid) = %v, want nil", err)
	}
}

func TestValidationErrorMessage(t *testing.T) {
	err := (&ValidationError{Field: "id", Reason: "must be positive"}).Error()
	if err != "" && !strings.Contains(err, "id") {
		t.Errorf("Error() = %q, should mention the field", err)
	}
	if err != "" && !strings.Contains(err, "must be positive") {
		t.Errorf("Error() = %q, should mention the reason", err)
	}
}

func TestGetWrapsNotFound(t *testing.T) {
	s := NewStore()
	_, err := s.Get(42)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(missing) = %v, want an error wrapping ErrNotFound", err)
	}
	if err == ErrNotFound {
		t.Fatal("the error should be wrapped with %w, not returned bare")
	}
}

func TestRoundTrip(t *testing.T) {
	s := NewStore()
	if err := s.Add(User{ID: 7, Name: "Grace"}); err != nil {
		t.Fatal(err)
	}
	u, err := s.Get(7)
	if err != nil || u.Name != "Grace" {
		t.Fatalf("Get(7) = %+v, %v", u, err)
	}
	if err := s.Delete(7); err != nil {
		t.Fatalf("Delete(7) = %v", err)
	}
	if err := s.Delete(7); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second Delete = %v, want ErrNotFound", err)
	}
}
`,
				}},
				Hints: []string{
					"§Error§: §fmt.Sprintf(\"invalid %s: %s\", e.Field, e.Reason)§.",
					"§NewStore§: §&Store{users: make(map[int]User)}§.",
					"§Add§: return §&ValidationError{Field: \"id\", Reason: \"must be positive\"}§ (pointer!) or the name variant.",
					"§Get§: §return User{}, fmt.Errorf(\"get user %d: %w\", id, ErrNotFound)§ — §%w§ is what preserves the chain.",
				},
				Solution: `package main

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("not found")
	ErrBadInput = errors.New("bad input")
)

type User struct {
	ID   int
	Name string
}

type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Reason)
}

type Store struct {
	users map[int]User
}

func NewStore() *Store {
	return &Store{users: make(map[int]User)}
}

func (s *Store) Add(u User) error {
	if u.ID <= 0 {
		return &ValidationError{Field: "id", Reason: "must be positive"}
	}
	if u.Name == "" {
		return &ValidationError{Field: "name", Reason: "must not be empty"}
	}
	s.users[u.ID] = u
	return nil
}

func (s *Store) Get(id int) (User, error) {
	u, ok := s.users[id]
	if !ok {
		return User{}, fmt.Errorf("get user %d: %w", id, ErrNotFound)
	}
	return u, nil
}

func (s *Store) Delete(id int) error {
	if _, ok := s.users[id]; !ok {
		return fmt.Errorf("delete user %d: %w", id, ErrNotFound)
	}
	delete(s.users, id)
	return nil
}

func main() {
	s := NewStore()
	fmt.Println(s.Add(User{ID: 1, Name: "Ada"}))
	fmt.Println(s.Add(User{ID: 0}))
	_, err := s.Get(99)
	fmt.Println(err, errors.Is(err, ErrNotFound))
}
`,
			},
			{
				ID:    "m1t17e3",
				Title: "Turn a panic into an error",
				Prompt: `Two functions, one of which must not be allowed to crash the program.

§§§go
func ParseAge(s string) (int, error)
func SafeDivide(a, b int) (result int, err error)
func SafeIndex(xs []int, i int) (value int, err error)
§§§

* §ParseAge§ uses §strconv.Atoi§ and **wraps** the failure so the message mentions the
  input: §fmt.Errorf("parse age %q: %w", s, err)§. It also rejects negative ages with its
  own error.
* §SafeDivide§ uses §defer§ + §recover§ so that §SafeDivide(1, 0)§ returns an error
  instead of crashing. The named results are essential here: the deferred function
  writes to §err§ and §result§ *after* the panic.
* §SafeIndex§ does the same for an out-of-range index. Guard it with a length check
  *and* wrap the whole thing in a deferred recover — the guard is the good code, and the
  recover is the safety net.

With named results you can assign to §err§ and §value§ inside the deferred closure; with
plain results you could not.
`,
				Starter: `package main

import (
	"fmt"
	"strconv"
)

// ParseAge converts a string to a non-negative age.
func ParseAge(s string) (int, error) {
	// TODO
	return 0, nil
}

// SafeDivide returns a/b, recovering from a division by zero.
func SafeDivide(a, b int) (result int, err error) {
	// TODO: defer + recover
	return 0, nil
}

// SafeIndex returns xs[i], recovering from an out-of-range index.
func SafeIndex(xs []int, i int) (value int, err error) {
	// TODO: guard, then defer + recover
	return 0, nil
}

func main() {
	fmt.Println(ParseAge("42"))
	fmt.Println(ParseAge("banana"))
	d1, e1 := SafeDivide(10, 2)
	d2, e2 := SafeDivide(1, 0)
	fmt.Println(d1, e1, d2, e2)
	fmt.Println(SafeIndex([]int{1, 2, 3}, 5))

	// strconv is imported for you.
	_ = strconv.Atoi
}
`,
				Tests: []File{{
					Name: "safe_test.go",
					Content: `package main

import (
	"errors"
	"strings"
	"testing"
)

func TestParseAge(t *testing.T) {
	got, err := ParseAge("42")
	if err != nil || got != 42 {
		t.Fatalf("ParseAge(\"42\") = %d, %v", got, err)
	}

	_, err = ParseAge("banana")
	if err == nil {
		t.Fatal("ParseAge(\"banana\") should fail")
	}
	if !strings.Contains(err.Error(), "banana") {
		t.Errorf("the error should mention the input: %v", err)
	}

	if _, err := ParseAge("-5"); err == nil {
		t.Error("ParseAge should reject a negative age")
	}
	if _, err := ParseAge("0"); err != nil {
		t.Errorf("ParseAge(\"0\") = %v, want nil", err)
	}
}

func TestSafeDivide(t *testing.T) {
	got, err := SafeDivide(10, 2)
	if err != nil || got != 5 {
		t.Fatalf("SafeDivide(10, 2) = %d, %v; want 5, nil", got, err)
	}

	got, err = SafeDivide(1, 0)
	if err == nil {
		t.Fatal("SafeDivide(1, 0) should return an error rather than panic")
	}
	if got != 0 {
		t.Errorf("SafeDivide(1, 0) value = %d, want 0", got)
	}
}

func TestSafeIndex(t *testing.T) {
	got, err := SafeIndex([]int{10, 20, 30}, 1)
	if err != nil || got != 20 {
		t.Fatalf("SafeIndex(_, 1) = %d, %v; want 20, nil", got, err)
	}

	for _, i := range []int{-1, 3, 100} {
		got, err := SafeIndex([]int{10, 20, 30}, i)
		if err == nil {
			t.Errorf("SafeIndex(_, %d) should fail", i)
		}
		if got != 0 {
			t.Errorf("SafeIndex(_, %d) value = %d, want 0", i, got)
		}
	}

	if _, err := SafeIndex(nil, 0); err == nil {
		t.Error("SafeIndex(nil, 0) should fail")
	}
}

func TestErrorsAreNotSentinel(t *testing.T) {
	_, err := SafeDivide(1, 0)
	if errors.Is(err, nil) {
		t.Fatal("a non-nil error should not be Is(nil)")
	}
}
`,
				}},
				Hints: []string{
					"§ParseAge§: §n, err := strconv.Atoi(s)§; on failure §return 0, fmt.Errorf(\"parse age %q: %w\", s, err)§; if §n < 0§ return your own error.",
					"§SafeDivide§: §defer func() { if r := recover(); r != nil { err = fmt.Errorf(\"divide: %v\", r); result = 0 } }()§ then §return a / b, nil§.",
					"§SafeIndex§: guard with §if i < 0 || i >= len(xs) { return 0, fmt.Errorf(\"index %d out of range (len %d)\", i, len(xs)) }§ and still keep the deferred recover as a net.",
					"Named results are what let the deferred closure set the values that get returned.",
				},
				Solution: `package main

import (
	"errors"
	"fmt"
	"strconv"
)

func ParseAge(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parse age %q: %w", s, err)
	}
	if n < 0 {
		return 0, errors.New("age must not be negative")
	}
	return n, nil
}

func SafeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("divide: %v", r)
			result = 0
		}
	}()
	return a / b, nil
}

func SafeIndex(xs []int, i int) (value int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("index: %v", r)
			value = 0
		}
	}()
	if i < 0 || i >= len(xs) {
		return 0, fmt.Errorf("index %d out of range (len %d)", i, len(xs))
	}
	return xs[i], nil
}

func main() {
	fmt.Println(ParseAge("42"))
	fmt.Println(ParseAge("banana"))
	d1, e1 := SafeDivide(10, 2)
	d2, e2 := SafeDivide(1, 0)
	fmt.Println(d1, e1, d2, e2)
	fmt.Println(SafeIndex([]int{1, 2, 3}, 5))
}
`,
			},
		},
	}
}

func m1t18() Topic {
	return Topic{
		ID:      "m1t18",
		Title:   "Generics: Type Parameters",
		Summary: "Writing code once that works for many types, and knowing when not to.",
		Theory: `
Generics let a function or type work over a *set* of types, with the compiler checking
every instantiation.

They arrived in Go 1.18 — late by design. The Go team's advice is worth internalising
before you use them:

> **Write the concrete version first. Reach for type parameters when you find yourself
> writing the same code a second and third time for different types.**

## A generic function

§§§go
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
§§§

Read §[T cmp.Ordered]§ as: *"for any type §T§ that supports ordering"*. Then §a§, §b§ and
the result are all §T§.

§§§go
fmt.Println(Max(3, 7))         // 7       — T is int
fmt.Println(Max("b", "a"))     // b       — T is string
fmt.Println(Max(1.5, 0.5))     // 1.5     — T is float64
§§§

One function, three types, all checked at compile time. §T§ stands for "type"; §[T any]§
is how you declare it. You will also see §K§ and §V§ for map-like code.

## Constraints: what T is allowed to be

The thing in the brackets is the **constraint** — the set of types §T§ may be.

| Constraint | Means |
| --- | --- |
| §any§ | any type at all (the default; same as §interface{}§) |
| §comparable§ | types supporting §==§ and §!=§ — usable as map keys |
| §cmp.Ordered§ | types supporting §<§ §>§ §<=§ §>=§ (numbers and strings) |
| §int \| float64§ | exactly these two types |
| §~int§ | any type whose *underlying* type is §int§ |

§cmp.Ordered§ and friends live in the §cmp§ package (Go 1.21+).

### The tilde: §~§

§~int§ means "int, or any named type based on int":

§§§go
type Celsius float64

func Sum[T ~int | ~float64](xs []T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}

fmt.Println(Sum([]int{1, 2, 3}))         // 6
fmt.Println(Sum([]Celsius{1.5, 2.5}))    // 4     <- works because of ~float64
§§§

Without the tilde, §Celsius§ would be rejected: it is not literally §float64§, it only
has §float64§ as its underlying type.

§§§go
type Adder interface {
	~int | ~float64
}
§§§

A constraint is just an interface. Naming it makes signatures readable.

### The §var zero T§ idiom

You cannot write §var x int§ when the type is unknown. Use:

§§§go
func First[T any](xs []T) (T, bool) {
	var zero T
	if len(xs) == 0 {
		return zero, false
	}
	return xs[0], true
}
§§§

§var zero T§ is "the zero value of whatever §T§ turns out to be". §T§ is a real type at
compile time — Go does not erase it as Java does, and there is no boxing overhead.

## Generic functions you will actually write

§§§go
func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

func Filter[T any](xs []T, keep func(T) bool) []T {
	out := make([]T, 0, len(xs))
	for _, x := range xs {
		if keep(x) {
			out = append(out, x)
		}
	}
	return out
}

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
§§§

Note §Keys§: §K comparable§ because map keys must be comparable, and §V any§ because a map
value can be anything.

## Generic types

A struct can also take type parameters:

§§§go
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

func (s *Stack[T]) Len() int { return len(s.items) }
§§§

The receiver writes §(s *Stack[T])§ — you must repeat the type parameters.

Using it, the compiler infers the argument from the receiver:

§§§go
s := &Stack[int]{}
s.Push(1)
s.Push(2)
v, _ := s.Pop()   // v is an int
fmt.Println(v, s.Len()) // 2 1
§§§

Real examples in the wild: §sync.Map[K, V]§ (Go 1.23), §slices.IndexFunc§,
§container/heap§ wrappers, and every generic data structure you will build in Module 2.

## When NOT to use generics

* **When an interface does the job.** If every type you care about already has the method
  you need, take the interface. §io.Reader§ beats §[T io.Reader]§.
* **When you would only instantiate it once.** Generics with one use are just
  indirection.
* **To avoid a type switch.** A generic function cannot look inside §T§; if you need to
  behave differently per type, you want an interface, not a type parameter.
* **For "generic" containers of unrelated things.** That is §any§, and it throws away
  type safety.
`,
		Quiz: []Question{
			{
				ID:     "m1t18q1",
				Prompt: "What may §T§ be in §func Max[T cmp.Ordered](a, b T) T§?",
				Options: []Option{
					{Text: "Any type whose operators are covered by §cmp.Ordered§ — all the integer and float types, and §string§.", Correct: true, Rationale: "§cmp.Ordered§ is a union of ordered types, including named types with those underlying types."},
					{Text: "Only §int§ and §float64§.", Correct: false, Rationale: "§string§, §uint§, §byte§ and others are included."},
					{Text: "Any type at all — constraints are documentation only.", Correct: false, Rationale: "The compiler enforces constraints."},
					{Text: "Any type with a §Less§ method.", Correct: false, Rationale: "§cmp.Ordered§ is about operators, not methods."},
				},
				Explanation: "A constraint is the set of types §T§ may be. §cmp.Ordered§ covers everything that supports ordering operators.",
			},
			{
				ID:     "m1t18q2",
				Prompt: "Why does §Sum[T ~int | ~float64]§ accept §Celsius§ (= §type Celsius float64§) but §Sum[T int | float64]§ would not?",
				Options: []Option{
					{Text: "Because §~§ matches any type whose *underlying* type is one of those, while the bare form requires the exact named type.", Correct: true, Rationale: "§Celsius§ is not literally §float64§; it merely has §float64§ as its underlying type."},
					{Text: "Because §~§ means \"pointer to\".", Correct: false, Rationale: "§~§ is about underlying types."},
					{Text: "Because named types are always excluded from generics.", Correct: false, Rationale: "They are included when you use §~§."},
					{Text: "There is no difference between the two.", Correct: false, Rationale: "There is: the tilde widens the set."},
				},
				Explanation: "§~T§ admits named types based on §T§. Without it, only the exact type qualifies.",
			},
			{
				ID:     "m1t18q3",
				Prompt: "What does §var zero T§ do inside a generic function?",
				Options: []Option{
					{Text: "Declares a variable holding the zero value of whatever type §T§ is at that instantiation.", Correct: true, Rationale: "You cannot name the type, so this is the idiom for \"the zero value\"."},
					{Text: "Declares an untyped nil.", Correct: false, Rationale: "§nil§ is not valid for a §T§ constrained to §int§."},
					{Text: "It is not allowed in Go.", Correct: false, Rationale: "It is the standard idiom."},
					{Text: "It resolves to §0§ regardless of §T§.", Correct: false, Rationale: "It is the zero value of §T§ — §\"\"§ for a string, §nil§ for a slice, and so on."},
				},
				Explanation: "Generics are instantiated at compile time, so §zero§ is a real typed value with no boxing.",
			},
			{
				ID:     "m1t18q4",
				Prompt: "When is an interface a better choice than a type parameter?",
				Options: []Option{
					{Text: "When the types you care about already share the behaviour you need — e.g. anything with a §Read§ method.", Correct: true, Rationale: "If one method describes the requirement, ask for the interface. Generics are for algorithms over types, not for polymorphism."},
					{Text: "Never — generics are always preferable.", Correct: false, Rationale: "Generics cannot look inside §T§, so they cannot replace dynamic dispatch."},
					{Text: "Only when performance does not matter.", Correct: false, Rationale: "Performance is rarely the deciding factor."},
					{Text: "When you want to avoid writing tests.", Correct: false, Rationale: "Unrelated."},
				},
				Explanation: "\"Accept interfaces\" still wins whenever the behaviour, not the type, is what you need.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t18e1",
				Title: "Warm-up: generic Min and Max",
				Prompt: `§§§go
func Min[T cmp.Ordered](a, b T) T
func Max[T cmp.Ordered](a, b T) T
func Clamp[T cmp.Ordered](value, low, high T) T
§§§

* §Min§ and §Max§ return the smaller and larger of two values.
* §Clamp§ returns §value§, but never below §low§ and never above §high§. Assume
  §low <= high§.
* One implementation must work for §int§, §float64§ **and** §string§ — the test calls
  all three.

Import §"cmp"§ for the constraint.
`,
				Starter: `package main

import (
	"cmp"
	"fmt"
)

// Min returns the smaller of a and b.
func Min[T cmp.Ordered](a, b T) T {
	// TODO
	return a
}

// Max returns the larger of a and b.
func Max[T cmp.Ordered](a, b T) T {
	// TODO
	return a
}

// Clamp keeps value within [low, high].
func Clamp[T cmp.Ordered](value, low, high T) T {
	// TODO
	return value
}

func main() {
	fmt.Println(Min(3, 7), Max(3, 7))
	fmt.Println(Min("b", "a"), Max("b", "a"))
	fmt.Println(Clamp(15, 0, 10), Clamp(-5, 0, 10), Clamp(5, 0, 10))
}
`,
				Tests: []File{{
					Name: "minmax_test.go",
					Content: `package main

import "testing"

func TestMinMaxInt(t *testing.T) {
	if Min(3, 7) != 3 || Max(3, 7) != 7 {
		t.Errorf("Min/Max(3, 7) = %d/%d", Min(3, 7), Max(3, 7))
	}
	if Min(-1, -9) != -9 || Max(-1, -9) != -1 {
		t.Errorf("Min/Max(-1, -9) = %d/%d", Min(-1, -9), Max(-1, -9))
	}
	if Min(5, 5) != 5 || Max(5, 5) != 5 {
		t.Errorf("Min/Max with equal values = %d/%d", Min(5, 5), Max(5, 5))
	}
}

func TestMinMaxString(t *testing.T) {
	if Min("b", "a") != "a" || Max("b", "a") != "b" {
		t.Errorf("Min/Max(\"b\", \"a\") = %q/%q", Min("b", "a"), Max("b", "a"))
	}
}

func TestMinMaxFloat(t *testing.T) {
	if Min(1.5, 0.5) != 0.5 || Max(1.5, 0.5) != 1.5 {
		t.Errorf("Min/Max(1.5, 0.5) = %v/%v", Min(1.5, 0.5), Max(1.5, 0.5))
	}
}

func TestClamp(t *testing.T) {
	if got := Clamp(15, 0, 10); got != 10 {
		t.Errorf("Clamp(15, 0, 10) = %d, want 10", got)
	}
	if got := Clamp(-5, 0, 10); got != 0 {
		t.Errorf("Clamp(-5, 0, 10) = %d, want 0", got)
	}
	if got := Clamp(5, 0, 10); got != 5 {
		t.Errorf("Clamp(5, 0, 10) = %d, want 5", got)
	}
	if got := Clamp("m", "a", "z"); got != "m" {
		t.Errorf("Clamp string = %q", got)
	}
	if got := Clamp("A", "a", "z"); got != "a" {
		t.Errorf("Clamp string low = %q, want \"a\"", got)
	}
}
`,
				}},
				Hints: []string{
					"§Min§: §if a < b { return a }; return b§.",
					"§Max§: same with §>§.",
					"§Clamp§ composes them: §return Min(Max(value, low), high)§ — a nice demonstration that generic functions call each other.",
					"§import \"cmp\"§ is enough; §cmp.Ordered§ is the constraint.",
				},
				Solution: `package main

import (
	"cmp"
	"fmt"
)

func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Clamp[T cmp.Ordered](value, low, high T) T {
	return Min(Max(value, low), high)
}

func main() {
	fmt.Println(Min(3, 7), Max(3, 7))
	fmt.Println(Min("b", "a"), Max("b", "a"))
	fmt.Println(Clamp(15, 0, 10), Clamp(-5, 0, 10), Clamp(5, 0, 10))
}
`,
			},
			{
				ID:    "m1t18e2",
				Title: "Map, Filter and Sum over any type",
				Prompt: `§§§go
func Map[T, U any](xs []T, f func(T) U) []U
func Filter[T any](xs []T, keep func(T) bool) []T
func Sum[T ~int | ~float64](xs []T) T
func Keys[K comparable, V any](m map[K]V) []K
§§§

* §Map§ transforms each element — note the **two** type parameters: the element type
  in, the element type out.
* §Filter§ keeps the elements for which §keep§ is true.
* §Sum§ adds numbers, and must work for named types thanks to §~§. An empty slice gives
  the zero value of §T§.
* §Keys§ returns the map's keys in **unsorted** order — the test sorts before comparing,
  because map iteration order is random.

§Map§ and §Filter§ must return non-nil slices even for nil input.
`,
				Starter: `package main

import (
	"fmt"
	"slices"
)

// Map applies f to every element of xs.
func Map[T, U any](xs []T, f func(T) U) []U {
	// TODO
	return nil
}

// Filter keeps the elements for which keep returns true.
func Filter[T any](xs []T, keep func(T) bool) []T {
	// TODO
	return nil
}

// Sum adds up numbers of any numeric type.
func Sum[T ~int | ~float64](xs []T) T {
	// TODO
	var zero T
	return zero
}

// Keys returns the keys of a map.
func Keys[K comparable, V any](m map[K]V) []K {
	// TODO
	return nil
}

func main() {
	fmt.Println(Map([]int{1, 2}, func(n int) string { return fmt.Sprint(n * 2) }))
	fmt.Println(Filter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 }))
	fmt.Println(Sum([]int{1, 2, 3}))
	fmt.Println(Keys(map[string]int{"a": 1, "b": 2}))
	keys := Keys(map[string]int{"b": 1, "a": 2})
	slices.Sort(keys)
	fmt.Println(keys)
}
`,
				Tests: []File{{
					Name: "generic_test.go",
					Content: `package main

import (
	"slices"
	"testing"
)

type Celsius float64

func TestMap(t *testing.T) {
	got := Map([]int{1, 2, 3}, func(n int) int { return n * 10 })
	if !slices.Equal(got, []int{10, 20, 30}) {
		t.Fatalf("Map = %v", got)
	}

	// The element type may change.
	strs := Map([]int{1, 2}, func(n int) string { return string(rune('a' - 1 + n)) })
	if !slices.Equal(strs, []string{"a", "b"}) {
		t.Fatalf("Map to strings = %v", strs)
	}

	if got := Map[int, int](nil, func(n int) int { return n }); got == nil || len(got) != 0 {
		t.Fatalf("Map(nil) = %v, want an empty non-nil slice", got)
	}
}

func TestFilter(t *testing.T) {
	if got := Filter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 }); !slices.Equal(got, []int{2, 4}) {
		t.Fatalf("Filter = %v", got)
	}
	if got := Filter([]string{"a", "bb"}, func(s string) bool { return len(s) > 1 }); !slices.Equal(got, []string{"bb"}) {
		t.Fatalf("Filter strings = %v", got)
	}
	if got := Filter[int](nil, func(int) bool { return true }); got == nil || len(got) != 0 {
		t.Fatalf("Filter(nil) = %v, want an empty non-nil slice", got)
	}
}

func TestSum(t *testing.T) {
	if got := Sum([]int{1, 2, 3}); got != 6 {
		t.Errorf("Sum(int) = %v, want 6", got)
	}
	if got := Sum([]float64{1.5, 2.5}); got != 4 {
		t.Errorf("Sum(float64) = %v, want 4", got)
	}
	// The named type works because of the ~ in the constraint.
	if got := Sum([]Celsius{2.5, 2.5}); got != 5 {
		t.Errorf("Sum(Celsius) = %v, want 5", got)
	}
	if got := Sum[int](nil); got != 0 {
		t.Errorf("Sum(nil) = %v, want 0", got)
	}
}

func TestKeys(t *testing.T) {
	got := Keys(map[string]int{"b": 1, "a": 2, "c": 3})
	slices.Sort(got)
	if !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Fatalf("Keys = %v", got)
	}
	if got := Keys(map[int]string{}); len(got) != 0 {
		t.Fatalf("Keys(empty) = %v", got)
	}
}
`,
				}},
				Hints: []string{
					"§Map§: §out := make([]U, 0, len(xs))§ then §out = append(out, f(x))§.",
					"§Filter§: same, with §if keep(x)§.",
					"§Sum§: keep §var total T§ and add in a loop, returning §total§ — not §zero§.",
					"§Keys§: §keys := make([]K, 0, len(m))§ then §for k := range m { keys = append(keys, k) }§.",
				},
				Solution: `package main

import (
	"fmt"
	"slices"
)

func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

func Filter[T any](xs []T, keep func(T) bool) []T {
	out := make([]T, 0, len(xs))
	for _, x := range xs {
		if keep(x) {
			out = append(out, x)
		}
	}
	return out
}

func Sum[T ~int | ~float64](xs []T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func main() {
	fmt.Println(Map([]int{1, 2}, func(n int) string { return fmt.Sprint(n * 2) }))
	fmt.Println(Filter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 }))
	fmt.Println(Sum([]int{1, 2, 3}))
	keys := Keys(map[string]int{"a": 1, "b": 2})
	slices.Sort(keys)
	fmt.Println(keys)
}
`,
			},
			{
				ID:    "m1t18e3",
				Title: "A generic Stack and Pair",
				Prompt: `Build two generic containers.

**1. A stack.**

§§§go
type Stack[T any] struct { /* your field */ }

func NewStack[T any]() *Stack[T]
func (s *Stack[T]) Push(v T)
func (s *Stack[T]) Pop() (T, bool)
func (s *Stack[T]) Peek() (T, bool)
func (s *Stack[T]) Len() int
§§§

* §Push§ adds to the top.
* §Pop§ removes and returns the top, or the zero value and §false§ when empty.
* §Peek§ returns the top without removing it.
* §Len§ must tolerate a **nil** receiver by returning 0 — the test calls it on a nil
  §*Stack[int]§.

**2. A pair.**

§§§go
type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

func (p Pair[K, V]) String() string
func NewPair[K comparable, V any](k K, v V) Pair[K, V]
func ToMap[K comparable, V any](pairs []Pair[K, V]) map[K]V
§§§

* §String§ renders §"(<key> => <value>)"§ using §%v§ for both.
* §ToMap§ folds a slice of pairs into a map; later duplicates overwrite earlier ones.
`,
				Starter: `package main

import "fmt"

type Stack[T any] struct {
	items []T
}

func NewStack[T any]() *Stack[T] {
	// TODO
	return nil
}

func (s *Stack[T]) Push(v T) {
	// TODO
}

func (s *Stack[T]) Pop() (T, bool) {
	// TODO
	var zero T
	return zero, false
}

func (s *Stack[T]) Peek() (T, bool) {
	// TODO
	var zero T
	return zero, false
}

func (s *Stack[T]) Len() int {
	// TODO
	return 0
}

type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

func NewPair[K comparable, V any](k K, v V) Pair[K, V] {
	// TODO
	return Pair[K, V]{}
}

func (p Pair[K, V]) String() string {
	// TODO
	return ""
}

func ToMap[K comparable, V any](pairs []Pair[K, V]) map[K]V {
	// TODO
	return nil
}

func main() {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	v, _ := s.Pop()
	fmt.Println(v, s.Len())
	fmt.Println(NewPair("a", 1))
	fmt.Println(ToMap([]Pair[string, int]{NewPair("a", 1), NewPair("b", 2)}))
}
`,
				Tests: []File{{
					Name: "stack_test.go",
					Content: `package main

import "testing"

func TestStack(t *testing.T) {
	s := NewStack[string]()
	if s == nil {
		t.Fatal("NewStack returned nil")
	}
	if s.Len() != 0 {
		t.Fatal("a new stack is not empty")
	}
	if _, ok := s.Pop(); ok {
		t.Fatal("Pop on an empty stack reported ok")
	}
	if _, ok := s.Peek(); ok {
		t.Fatal("Peek on an empty stack reported ok")
	}

	s.Push("a")
	s.Push("b")
	if s.Len() != 2 {
		t.Fatalf("Len = %d, want 2", s.Len())
	}
	if v, ok := s.Peek(); !ok || v != "b" {
		t.Fatalf("Peek = %q, %v; want \"b\", true", v, ok)
	}
	if s.Len() != 2 {
		t.Fatalf("Peek changed the length: %d", s.Len())
	}
	if v, ok := s.Pop(); !ok || v != "b" {
		t.Fatalf("Pop = %q, %v; want \"b\", true", v, ok)
	}
	if s.Len() != 1 {
		t.Fatalf("Len after Pop = %d, want 1", s.Len())
	}
}

func TestStackZeroValue(t *testing.T) {
	var s *Stack[int]
	if s.Len() != 0 {
		t.Fatal("Len on a nil stack should be 0")
	}
}

func TestStackOfInts(t *testing.T) {
	s := NewStack[int]()
	for i := 1; i <= 3; i++ {
		s.Push(i)
	}
	var got []int
	for {
		v, ok := s.Pop()
		if !ok {
			break
		}
		got = append(got, v)
	}
	want := []int{3, 2, 1}
	if len(got) != 3 {
		t.Fatalf("popped %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("popped %v, want %v", got, want)
		}
	}
}

func TestPair(t *testing.T) {
	p := NewPair("a", 1)
	if p.Key != "a" || p.Value != 1 {
		t.Fatalf("NewPair = %+v", p)
	}
	if got := p.String(); got == "" {
		t.Error("Pair.String returned an empty string")
	}
	if got := p.String(); len(got) < 5 || got[0] != '(' {
		t.Errorf("Pair.String() = %q, expected something like \"(a => 1)\"", got)
	}
}

func TestToMap(t *testing.T) {
	m := ToMap([]Pair[string, int]{NewPair("a", 1), NewPair("b", 2)})
	if len(m) != 2 || m["a"] != 1 || m["b"] != 2 {
		t.Fatalf("ToMap = %v", m)
	}

	// Later duplicates win.
	m = ToMap([]Pair[string, int]{NewPair("a", 1), NewPair("a", 5)})
	if m["a"] != 5 {
		t.Fatalf("ToMap with duplicates = %v, want a:5", m)
	}

	if len(ToMap[string, int](nil)) != 0 {
		t.Error("ToMap(nil) should be empty")
	}
}
`,
				}},
				Hints: []string{
					"§NewStack§ returns §&Stack[T]{}§.",
					"§Push§: §s.items = append(s.items, v)§.",
					"§Pop§: guard §len(s.items) == 0§, take the last element with §s.items[len(s.items)-1]§, then shrink with §s.items = s.items[:len(s.items)-1]§.",
					"§Len§ must handle a nil receiver: §if s == nil { return 0 }§.",
					"§String§: §fmt.Sprintf(\"(%v => %v)\", p.Key, p.Value)§.",
					"§ToMap§: §out := map[K]V{}§ then §for _, p := range pairs { out[p.Key] = p.Value }§.",
				},
				Solution: `package main

import "fmt"

type Stack[T any] struct {
	items []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if s == nil || len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

func (s *Stack[T]) Peek() (T, bool) {
	var zero T
	if s == nil || len(s.items) == 0 {
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Len() int {
	if s == nil {
		return 0
	}
	return len(s.items)
}

type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

func NewPair[K comparable, V any](k K, v V) Pair[K, V] {
	return Pair[K, V]{Key: k, Value: v}
}

func (p Pair[K, V]) String() string {
	return fmt.Sprintf("(%v => %v)", p.Key, p.Value)
}

func ToMap[K comparable, V any](pairs []Pair[K, V]) map[K]V {
	out := map[K]V{}
	for _, p := range pairs {
		out[p.Key] = p.Value
	}
	return out
}

func main() {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	v, _ := s.Pop()
	fmt.Println(v, s.Len())
	fmt.Println(NewPair("a", 1))
	fmt.Println(ToMap([]Pair[string, int]{NewPair("a", 1), NewPair("b", 2)}))
}
`,
			},
		},
	}
}
