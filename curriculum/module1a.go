package curriculum

// module1 builds Module 1: Go fundamentals, in the order a beginner should meet
// the ideas. Each topic is deliberately small: one core mental model, several
// worked examples with their real output, and two or three graded exercises.
//
// Reminder: "§" stands for a backtick everywhere in this package (see types.go).
func module1() Module {
	return Module{
		ID:      "m1",
		Title:   "Module 1 · Go Fundamentals",
		Summary: "Learn Go one idea at a time, starting from an empty file. Every concept is introduced with a complete example and its output before you are asked to write anything.",
		Topics: []Topic{
			m1t1(),  // hello + packages
			m1t2(),  // variables & zero values
			m1t3(),  // types, conversions, constants
			m1t4(),  // formatted I/O
			m1t5(),  // if & switch
			m1t6(),  // loops
			m1t7(),  // functions & multiple returns
			m1t8(),  // functions as values, closures, defer
			m1t9(),  // pointers
			m1t10(), // structs
			m1t11(), // arrays & slice basics
			m1t12(), // slices under the hood
			m1t13(), // strings, bytes & runes
			m1t14(), // maps
			m1t15(), // methods & embedding
			m1t16(), // interfaces
			m1t17(), // errors
			m1t18(), // generics
		},
	}
}

func m1t1() Topic {
	return Topic{
		ID:      "m1t1",
		Title:   "Packages and Program Entry",
		Summary: "How Go organises code, and where execution begins.",
		Theory: `
Go organises code into **packages**. A package is one directory of §.go§ files compiled
together, sharing a single scope: a function, type, constant or variable declared in one
file is visible to every other file in that directory, with no import required.

Packages are grouped into a **module**, a set of packages versioned together and rooted
at a §go.mod§ file. A module is the unit you build, test and publish; a package is the
unit you import.

A module and its packages on disk:

§§§text
myapp/
├── go.mod             says "this folder is one module called myapp"
├── main.go            starts with: package main
└── mathutil/
    ├── add.go         starts with: package mathutil
    └── multiply.go    starts with: package mathutil
§§§

Three rules govern the layout:

1. **Every §.go§ file begins with a §package§ clause** naming the folder it lives in.
2. **Two files in the same folder must declare the same package name.** §add.go§
   and §multiply.go§ above are both §package mathutil§ — that is what lets §add.go§
   call a function defined in §multiply.go§.
3. **Two files in different folders must declare different package names.** §main.go§
   cannot say §package mathutil§, because it is not in the §mathutil§ folder.

In short, a package is a directory whose files share a scope.

**If you are coming from Java:** a Go package plays the role of a Java *package* and
a *compilation unit* at once. There is no §class§ wrapper around everything, and the
folder name is the package name.

**If you are coming from Python:** a package is like a directory with §__init__.py§,
except the directory name *is* the package name — you never invent a different one.

**A package is not:** a class, a single file, or a whole project. A project is made of
many packages.

## The entry point: package main

An **executable** program must live in a package called §main§, and that package must
contain a function called §main§.

§§§go
package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}
§§§

Output:

§§§text
Hello, Go!
§§§

Reading it line by line:

| Line | What it means |
| --- | --- |
| §package main§ | "This folder is a program, not a library." |
| §import "fmt"§ | "I want to use code from Go's §fmt§ package." §fmt§ = format. |
| §func main() {§ | The function Go runs when you start the program. |
| §fmt.Println(...)§ | §Println§ = "print line". The §fmt.§ prefix means "the §Println§ from package §fmt§". |

§func main§ takes no arguments and returns nothing; that signature is fixed. Command-line
arguments arrive through the §os.Args§ slice instead.

## Calling across packages: the dotted prefix

You will have noticed §fmt.Println§. In Go, whenever you use something that lives in
*another* package, you must write §packagename.Thing§:

§§§go
package main

import (
	"fmt"     // imported, so we can write fmt.Sprintf
	"strings" // imported, so we can write strings.ToUpper
)

func main() {
	fmt.Println(strings.ToUpper("quiet")) // the strings package
	fmt.Println(strings.Repeat("ab", 3))  // repeat "ab" three times
}
§§§

Output:

§§§text
QUIET
ababab
§§§

One import is written §import "fmt"§; several are written in a parenthesised block,
one per line, which is the normal style once you import more than one. §goimports§ and
most editors add them for you.

## Exported names: the capital letter rule

Go has **no** §public§ and **no** §private§. Visibility is decided by the first letter
of a name:

| Name | Who can use it |
| --- | --- |
| §Greeting§ (capital G) | Anyone, in any package. This is called **exported**. |
| §greeting§ (lower-case g) | Only code in the *same* package. This is called **unexported**. |

That is why every §fmt§ function you have seen starts with a capital letter:
§fmt.Println§, §fmt.Sprintf§. The §fmt§ package also has plenty of lower-case helpers
that only it can use.

**Java comparison:** §Greeting§ is §public§, §greeting§ is *package-private*. There is
no §protected§.

## Comments

§§§go
// A line comment, ignored by the compiler.

/*
   A block comment.
   Useful for longer explanations.
*/

// Greeting returns a friendly hello for the given name.
func Greeting(name string) string {
	return "Hello, " + name
}
§§§

The convention is to start a comment that documents a declaration with the name of the
thing being documented — §// Greeting returns ...§. The §go doc§ command reads those
comments to build documentation automatically, so it is worth the small effort.

## Running the program

From the folder containing §main.go§:

| Command | What it does |
| --- | --- |
| §go run .§ | Compiles and runs immediately. Use this while learning. |
| §go build -o myapp .§ | Produces a standalone binary called §myapp§. |
| §go fmt .§ | Rewrites your files in the official formatting style. |

The §.§ means "the package in this folder". Go installs a formatter (§gofmt§) that
ends all arguments about how code should look — there is one true style, and §go fmt§
applies it.

## The one thing people trip over

A file in package §main§ is a **program**. A file in any other package is a
**library**: it cannot be run on its own, and it cannot have a §func main§.

§§§go
// in folder myapp/mathutil/, file add.go
package mathutil

// Add is exported because it starts with a capital A.
func Add(a, b int) int { return a + b }

// helper is unexported: only files in package mathutil may call it.
func helper() string { return "only visible inside mathutil" }
§§§

If you try §go run mathutil/§, Go reports that it is not a §main§ package. That is not
an error in your code — it is Go telling you "this is a library, there is nothing to
run".
`,
		Quiz: []Question{
			{
				ID:     "m1t1q1",
				Prompt: "Which statements about Go packages are true? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "A package is a folder of §.go§ files that all declare the same package name.", Correct: true, Rationale: "That is the whole definition. Same folder, same package clause, shared namespace."},
					{Text: "Two files in the same folder may declare different package names.", Correct: false, Rationale: "That is a compile error: §found packages a and b in the same directory§."},
					{Text: "A file in package §main§ with §func main§ can be executed with §go run .§", Correct: true, Rationale: "That is exactly what makes a program executable."},
					{Text: "A package named §mathutil§ can be executed directly with §go run mathutil/§.", Correct: false, Rationale: "Only package §main§ produces an executable."},
					{Text: "Package names always start with a capital letter.", Correct: false, Rationale: "Package names are conventionally lower-case, like §fmt§ or §strings§."},
				},
				Explanation: "A package = a folder + a shared §package§ clause. Only §package main§ containing §func main§ is a program; everything else is a library.",
			},
			{
				ID:     "m1t1q2",
				Prompt: "This file lives in the folder §myapp/store/§. Which line must it start with?",
				Code: `???

import "fmt"

func Save(name string) {
	fmt.Println("saving", name)
}`,
				Options: []Option{
					{Text: "§package store§", Correct: true, Rationale: "The package name must agree with the folder, and all files in that folder must agree with each other."},
					{Text: "§package main§", Correct: false, Rationale: "§main§ is reserved for programs, and this folder is a library."},
					{Text: "§package Save§", Correct: false, Rationale: "Package names are normally lower-case and unrelated to function names."},
					{Text: "Nothing — a package clause is optional.", Correct: false, Rationale: "Every §.go§ file must begin with a package clause."},
				},
				Explanation: "All files in §myapp/store/§ declare §package store§. Folder name and package name match.",
			},
			{
				ID:     "m1t1q3",
				Prompt: "Which names can code in a *different* package use? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "§func Greeting(name string) string§", Correct: true, Rationale: "Capital G → exported."},
					{Text: "§func greet(name string) string§", Correct: false, Rationale: "Lower-case g → visible only inside its own package."},
					{Text: "§const MaxRetries = 3§", Correct: true, Rationale: "Exported constant."},
					{Text: "§var errNotFound = errors.New(\"nope\")§", Correct: false, Rationale: "Lower-case → unexported. The standard library names *exported* sentinels §ErrXxx§."},
					{Text: "§type Client struct{}§", Correct: true, Rationale: "Exported type."},
				},
				Explanation: "Capital first letter = exported = usable from any package. There are no §public§/§private§ keywords in Go.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t1e1",
				Title: "Warm-up: print a line",
				Prompt: `Print one line.

Write a program whose §main§ function prints exactly this line, including the exclamation mark:

§§§text
Hello, Go!
§§§

Use §fmt.Println§. When you press **Run**, you will see your program's output.
§Check§ runs a hidden test that captures what your program prints — so the text must
match exactly, with no extra characters.
`,
				Starter: `package main

import "fmt"

func main() {
	// fmt is imported for you — you do not need to add it.
	// TODO: print "Hello, Go!" with fmt.Println, then delete this line.
	_ = fmt.Sprint
}
`,
				Tests: []File{{
					Name: "hello_test.go",
					Content: `package main

import (
	"os"
	"strings"
	"testing"
)

var capturedOutput string

// TestMain runs the program's main() once and captures everything it prints.
func TestMain(m *testing.M) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		os.Exit(m.Run())
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf []byte
		tmp := make([]byte, 256)
		for {
			n, readErr := r.Read(tmp)
			buf = append(buf, tmp[:n]...)
			if readErr != nil {
				break
			}
		}
		done <- string(buf)
	}()

	main()

	w.Close()
	os.Stdout = old
	capturedOutput = <-done

	os.Exit(m.Run())
}

func TestGreetingPrinted(t *testing.T) {
	if got := strings.TrimSpace(capturedOutput); got != "Hello, Go!" {
		t.Fatalf("your program printed %q, want %q", got, "Hello, Go!")
	}
}
`,
				}},
				Hints: []string{
					"§fmt.Println(...)§ prints whatever you give it, then moves to a new line.",
					"The text must be exactly §Hello, Go!§ — capital H, capital G.",
					"§fmt.Println(\"Hello, Go!\")§",
				},
				Solution: `package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}
`,
			},
			{
				ID:    "m1t1e2",
				Title: "Building a greeting function",
				Prompt: `A program lives in §package main§. The hidden test, however, is a separate
file — and to reach it, our file has to declare a package too. That is why everything
we write from now on sits in package §main§: the test file joins the same package and
can call our functions directly.

Write a function:

§§§go
func Greeting(name string) string
§§§

that returns the text §Hello, <name>!§.

* §Greeting("Gopher")§ must return §Hello, Gopher!§
* §Greeting("world")§ must return §Hello, world!§

Notice the capital §G§ in §Greeting§. That makes the function **exported**, which
matters as soon as another package (or another file in a test) wants to use it.

To build the text, remember that §fmt.Println§ *prints* a string. Because §Greeting§
must return one instead, use its sibling §fmt.Sprintf§ — same formatting, but it hands
the result back as a §string§:

§§§go
message := fmt.Sprintf("Goodbye, %s.", who) // "Goodbye, Ada." when who is "Ada"
§§§

§%s§ is a placeholder that a string argument fills in; the values after the format
string fill the placeholders left to right. (Format verbs get a full topic later.)

Then call it from §main§ so that pressing **Run** prints the greeting too.
`,
				Starter: `package main

import "fmt"

// Greeting returns a friendly greeting for name.
func Greeting(name string) string {
	// TODO
	return ""
}

func main() {
	fmt.Println(Greeting("Gopher"))
}
`,
				Tests: []File{{
					Name: "greeting_test.go",
					Content: `package main

import "testing"

func TestGreeting(t *testing.T) {
	cases := map[string]string{
		"Gopher": "Hello, Gopher!",
		"world":  "Hello, world!",
		"":       "Hello, !",
		"Ada":    "Hello, Ada!",
	}
	for in, want := range cases {
		if got := Greeting(in); got != want {
			t.Errorf("Greeting(%q) = %q, want %q", in, got, want)
		}
	}
}
`,
				}},
				Hints: []string{
					"§Greeting§ returns its text rather than printing it, so use §fmt.Sprintf§ from the prompt.",
					"Build the format string §\"Hello, %s!\"§ and pass §name§ as the one value after it.",
					"§return fmt.Sprintf(\"Hello, %s!\", name)§",
				},
				Solution: `package main

import "fmt"

func Greeting(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func main() {
	fmt.Println(Greeting("Gopher"))
}
`,
			},
		},
	}
}

func m1t2() Topic {
	return Topic{
		ID:      "m1t2",
		Title:   "Variables and zero values",
		Summary: "Declaration, assignment, and the zero value every type starts at.",
		Theory: `
Go is statically typed: a variable's type is fixed at compile time, and a variable is
never left uninitialised. Declaring one without a value zeroes its memory rather than
leaving it undefined. That default — §0§, §""§, §false§ or §nil§, depending on the type —
is the type's **zero value**.

## Declaring a variable

There are two forms, and both are used constantly.

**Way 1 — §var§**, when you want to state the type (or just want the zero value):

§§§go
var age int              // a box named age, holding an int
var name string          // a box named name, holding a string
var price float64 = 9.99 // declare with a type AND a starting value
§§§

**Way 2 — §:=§**, the *short declaration*, when the type is obvious from the value:

§§§go
count := 10      // Go infers: count is an int
greeting := "hi" // Go infers: greeting is a string
ratio := 1.5     // Go infers: ratio is a float64
§§§

§:=§ means "create this variable and put this value in it". It is *only* allowed inside
a function (inside §func main§, for example). At the top level of a file you must use
§var§.

Both lines below do the same thing — the second is just shorter:

§§§go
var count int = 10
count := 10
§§§

A complete program:

§§§go
package main

import "fmt"

func main() {
	var city string = "Oslo"
	population := 709037

	fmt.Println(city, "has", population, "people")
}
§§§

Output:

§§§text
Oslo has 709037 people
§§§

The commas in §fmt.Println(a, b, c)§ put a single space between each value. We will do
much better than that in the formatted-output topic.

## Zero values

Where Java leaves an uninitialised §Integer§ field as §null§ — and throws when you read
it — Go defines a value for every type:

> **Every variable is initialised. If you do not give a value, Go gives you the zero
> value of that type.**

So §var age int§ does not leave §age§ "empty" — it sets it to §0§. This is one of the
reasons Go programs crash less: there is no "uninitialised" state to trip over.

| Type | Zero value | Notes |
| --- | --- | --- |
| §int§, §int8§, §int64§, §uint§ … | §0§ | all number types start at zero |
| §float64§, §float32§ | §0§ | the same number as §0.0§ |
| §bool§ | §false§ | |
| §string§ | §""§ | an empty string, not §nil§ |
| pointer (§*T§) | §nil§ | "points at nothing" |
| slice, map, channel | §nil§ | an empty collection |
| struct | a struct whose every field is the zero value | |
| array of N things | N zero values | |

Every type has one:

§§§go
package main

import "fmt"

func main() {
	var i int
	var f float64
	var b bool
	var s string

	fmt.Printf("%d %v %v %q\n", i, f, b, s)
}
§§§

Output:

§§§text
0 0 false ""
§§§

§%q§ prints a string with quotation marks so you can *see* that it is empty rather
than missing. We meet all the percent verbs in the next topic.

## Declaring several at once

Grouped declarations are common at file level:

§§§go
var (
	host    string
	port    int
	verbose bool
)
§§§

And on one line:

§§§go
var x, y int = 3, 4
a, b := "left", "right"
§§§

## Reassigning vs redeclaring

§§§go
count := 1 // creates count
count = 2  // assigns a new value to the existing count  (one equals sign)
count := 3 // ERROR: no new variables on left side of :=
§§§

The error message is Go's way of saying "§:=§ creates, §=§ assigns". If you are
surprised by it, look for a variable you thought was new but is not.

One useful exception: §:=§ is allowed if *at least one* variable on the left is new:

§§§go
result, err := doThing() // both new
next, err := doOther()   // next is new, so := is fine and err is reused
§§§

## Why zero values matter

Because every type has a usable zero value, structs are ready to use the moment they
are declared. There is no constructor required just to avoid a crash:

§§§go
package main

import "fmt"

type Config struct {
	Host string
	Port int
}

func main() {
	var c Config                   // no constructor, no nil
	fmt.Println(c.Host, c.Port)    // prints an empty string and 0
	c.Port = 8080
	fmt.Println(c.Port)
}
§§§

Output:

§§§text
 0
8080
§§§

## Two rules

* §:=§ creates; §=§ assigns. §:=§ only works inside a function.
* A declared-but-unused **local** variable is a compile error: §declared and not used§.
  Go refuses to compile code with dead variables.
`,
		Quiz: []Question{
			{
				ID:     "m1t2q1",
				Prompt: "What does this program print?",
				Code: `package main

import "fmt"

func main() {
	var i int
	var s string
	var b bool
	var f float64
	fmt.Println(i, s == "", b, f)
}`,
				Options: []Option{
					{Text: "§0 true false 0§", Correct: true, Rationale: "Zero values: §0§, §\"\"§, §false§, §0§. We print §s == \"\"§ to show the empty string compares equal to §\"\"§."},
					{Text: "§0 <nil> false 0§", Correct: false, Rationale: "A string's zero value is §\"\"§, never §nil§. Only pointers, slices, maps and channels can be §nil§."},
					{Text: "It does not compile: the variables are never assigned.", Correct: false, Rationale: "Declaring and then using a variable in §Println§ is enough; Go initialises it for you."},
					{Text: "§<nil> <nil> <nil> <nil>§", Correct: false, Rationale: "§int§, §string§, §bool§ and §float64§ cannot be §nil§ at all."},
				},
				Explanation: "Every Go type has a defined zero value. Numbers are 0, booleans false, strings \"\" — and the program compiles because the variables are used.",
			},
			{
				ID:     "m1t2q2",
				Prompt: "Which line is a compile error?",
				Code: `func main() {
	a := 1
	a = 2
	a := 3
	b, c := 4, 5
	d, c := 6, 7
	_ = b
	_ = d
	_ = c
	_ = a
}`,
				Options: []Option{
					{Text: "§a := 3§", Correct: true, Rationale: "§a§ already exists and no *new* variable is introduced, so §:=§ is rejected."},
					{Text: "§a = 2§", Correct: false, Rationale: "Plain assignment to an existing variable is fine."},
					{Text: "§d, c := 6, 7§", Correct: false, Rationale: "§d§ is new, so §:=§ is allowed even though §c§ already exists."},
					{Text: "§_ = a§", Correct: false, Rationale: "§_§ is the blank identifier and may always be assigned to."},
				},
				Explanation: "§:=§ needs at least one new variable on the left-hand side. §a := 3§ has none.",
			},
			{
				ID:     "m1t2q3",
				Prompt: "A §var xs []int§ is declared and never given a value. What can you safely do with it?",
				Options: []Option{
					{Text: "Read its length (0), range over it (zero iterations), and §xs = append(xs, 1)§ to add an element.", Correct: true, Rationale: "A nil slice behaves like an empty one for reading, and append on a nil slice is legal — it allocates."},
					{Text: "Nothing — it is §nil§ and any use panics.", Correct: false, Rationale: "Reading is safe; only writing by index would panic."},
					{Text: "§xs[0] = 1§ works because Go grows it automatically.", Correct: false, Rationale: "Indexing past the length panics; use §append§."},
					{Text: "§len(xs)§ panics because the slice is nil.", Correct: false, Rationale: "§len§ of a nil slice is 0."},
				},
				Explanation: "Nil slices (and nil maps) are readable. A nil *map* panics on write, but a nil slice is fine because §append§ returns a new slice.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t2e1",
				Title: "Warm-up: return the zero values",
				Prompt: `Declare four variables with §var§ — an §int§, a §string§, a §bool§ and a §float64§
— and return them *without ever assigning a value*. Let Go's zero values do the work.

§§§go
func ZeroValues() (int, string, bool, float64)
§§§

The test expects §0, "", false, 0§.
`,
				Starter: `package main

import "fmt"

// ZeroValues returns the zero value of each of the four types.
func ZeroValues() (int, string, bool, float64) {
	// TODO: declare the variables with var and return them
	return 0, "", false, 0
}

func main() {
	fmt.Println(ZeroValues())
}
`,
				Tests: []File{{
					Name: "zerovalues_test.go",
					Content: `package main

import "testing"

func TestZeroValues(t *testing.T) {
	i, s, b, f := ZeroValues()
	if i != 0 {
		t.Errorf("int zero value = %d, want 0", i)
	}
	if s != "" {
		t.Errorf("string zero value = %q, want empty", s)
	}
	if b != false {
		t.Errorf("bool zero value = %v, want false", b)
	}
	if f != 0 {
		t.Errorf("float64 zero value = %v, want 0", f)
	}
}
`,
				}},
				Hints: []string{
					"§var i int§ declares an int and Go sets it to 0 for you.",
					"You can return all four at once: §return i, s, b, f§.",
					"Do not write §i := 0§ — this exercise is about the zero value.",
				},
				Solution: `package main

import "fmt"

func ZeroValues() (int, string, bool, float64) {
	var i int
	var s string
	var b bool
	var f float64
	return i, s, b, f
}

func main() {
	fmt.Println(ZeroValues())
}
`,
			},
			{
				ID:    "m1t2e2",
				Title: "Swap two values",
				Prompt: `Write a function that takes two strings and returns them in the opposite order.

§§§go
func Swap(a, b string) (string, string)
§§§

§Swap("left", "right")§ must return §"right", "left"§.

Go has a neat trick here: you can assign several variables at once, and the right-hand
side is fully evaluated *before* anything is written. That means §a, b = b, a§ swaps
without a temporary variable.
`,
				Starter: `package main

import "fmt"

// Swap returns its two arguments in the opposite order.
func Swap(a, b string) (string, string) {
	// TODO
	return a, b
}

func main() {
	fmt.Println(Swap("left", "right"))
}
`,
				Tests: []File{{
					Name: "swap_test.go",
					Content: `package main

import "testing"

func TestSwap(t *testing.T) {
	a, b := Swap("left", "right")
	if a != "right" || b != "left" {
		t.Fatalf("Swap(\"left\", \"right\") = %q, %q; want \"right\", \"left\"", a, b)
	}

	x, y := Swap("", "x")
	if x != "x" || y != "" {
		t.Fatalf("Swap with an empty string = %q, %q", x, y)
	}
}
`,
				}},
				Hints: []string{
					"§a, b = b, a§ swaps the two variables in one statement.",
					"Then §return a, b§.",
					"Use §=§, not §:=§ — the parameters §a§ and §b§ already exist.",
				},
				Solution: `package main

import "fmt"

func Swap(a, b string) (string, string) {
	a, b = b, a
	return a, b
}

func main() {
	fmt.Println(Swap("left", "right"))
}
`,
			},
			{
				ID:    "m1t2e3",
				Title: "Read two numbers and print their sum",
				Prompt: `This exercise reads from **standard input**, using §fmt.Scan§.

§§§go
package main

import "fmt"

func main() {
	var a, b int
	fmt.Scan(&a, &b)
	fmt.Println(a + b)
}
§§§

§&a§ means "the address of §a§" — §Scan§ needs to know *where* to put the number it
reads. We explain §&§ properly in the pointers topic; for now, read §&a§ as
"fill in §a§".

**Your task:** put the addition in its own function so the test can call it directly,

§§§go
func SumTwo(a, b int) int
§§§

and keep §main§ reading the two numbers and printing §SumTwo(a, b)§.

Use the **Input** box under the editor to type something like §7 35§ and press **Run**.
`,
				Stdin:    true,
				SampleIn: "7 35",
				Starter: `package main

import "fmt"

// SumTwo returns a + b.
func SumTwo(a, b int) int {
	// TODO
	return 0
}

func main() {
	var a, b int
	fmt.Scan(&a, &b)
	fmt.Println(SumTwo(a, b))
}
`,
				Tests: []File{{
					Name: "sumtwo_test.go",
					Content: `package main

import "testing"

func TestSumTwo(t *testing.T) {
	cases := []struct{ a, b, want int }{
		{7, 35, 42},
		{0, 0, 0},
		{-5, 5, 0},
		{-3, -4, -7},
		{100, 250, 350},
	}
	for _, c := range cases {
		if got := SumTwo(c.a, c.b); got != c.want {
			t.Errorf("SumTwo(%d, %d) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
`,
				}},
				Hints: []string{
					"Your function body is one line: §return a + b§.",
					"§fmt.Scan(&a, &b)§ fills both variables from the input, separated by whitespace.",
					"Try the Input box with §7 35§ — you should see §42§ in the Output tab.",
				},
				Solution: `package main

import "fmt"

func SumTwo(a, b int) int {
	return a + b
}

func main() {
	var a, b int
	fmt.Scan(&a, &b)
	fmt.Println(SumTwo(a, b))
}
`,
			},
		},
	}
}

func m1t3() Topic {
	return Topic{
		ID:      "m1t3",
		Title:   "Basic types, conversions and constants",
		Summary: "The number types, converting between them, and values that never change.",
		Theory: `
Go has a small set of built-in types, and it never converts between them implicitly.
The strictness is deliberate: it turns numeric mistakes into compile errors rather than
silent truncation.

## The basic types

**Whole numbers:**

| Type | Size | Use it for |
| --- | --- | --- |
| §int§ | 32 or 64 bits, whichever the machine prefers | **the default. Use this.** |
| §int8§ §int16§ §int32§ §int64§ | exactly 8/16/32/64 bits | a specific width (files, network protocols) |
| §uint§, §uint8§ … §uint64§ | the same, but never negative | bit manipulation, byte buffers |
| §byte§ | exactly 8 bits | raw data; a **nickname for §uint8§** |
| §rune§ | exactly 32 bits | one Unicode character; a **nickname for §int32§** |

**Numbers with a decimal point:**

| Type | Notes |
| --- | --- |
| §float64§ | 64-bit. **Use this one.** It is the default for §1.5§ |
| §float32§ | 32-bit, less precise, half the memory |

**Other:**

| Type | Notes |
| --- | --- |
| §bool§ | §true§ or §false§ — nothing else, no 0/1 |
| §string§ | text; it gets its own topic later |
| §complex64§, §complex128§ | complex numbers with §i§; rarely needed |

Two nicknames worth memorising, because the standard library uses them constantly:

* §byte§ is exactly §uint8§.
* §rune§ is exactly §int32§.

## Conversion: Go never does it for you

This is the rule that catches everybody coming from another language:

> **Go will not convert between numeric types automatically. Not even §int§ → §int64§.**

§§§go
var small int32 = 7
var big int64 = small // COMPILE ERROR
§§§

The error reads §cannot use small (variable of type int32) as int64 value§. You must
say what you want:

§§§go
var small int32 = 7
var big int64 = int64(small) // fine
§§§

A conversion is written like a function call: §int64(x)§, §float64(y)§, §int(z)§.

**Why so strict?** Because the alternative (C's implicit narrowing) silently loses
data: §int64(300)§ stuffed into a §byte§ becomes 44, and C lets that happen without a
word. Go asks you to write the conversion so the truncation is *your* decision.

### Integer division

§§§go
fmt.Println(7 / 2)          // 3
fmt.Println(7.0 / 2.0)      // 3.5
fmt.Println(float64(7) / 2) // 3.5
§§§

Whenever an average comes out as a whole number, you forgot a conversion.

## A worked example: converting both ways

§§§go
package main

import "fmt"

func main() {
	price := 19.99 // float64
	quantity := 3  // int

	// total := price * quantity        // COMPILE ERROR: mismatched types

	total := price * float64(quantity) // convert the int up to float64
	rounded := int(total)              // converting down truncates

	fmt.Println(total)
	fmt.Println(rounded)
	fmt.Println(total - float64(rounded))
}
§§§

Output:

§§§text
59.97
59
0.9699999999999989
§§§

Three things to notice:

1. §float64(quantity)§ was required. Without it the program does not compile.
2. §int(59.97)§ is §59§, not §60§. **Conversion truncates toward zero.** To round you
   need §math.Round§.
3. §0.9699...§ instead of §0.97§ is not a Go bug. Binary floating point cannot
   represent §19.99§ exactly, so tiny errors like this are normal. Never compare floats
   with §==§; check that the difference is small.

## Constants: values that never change

§§§go
const Pi = 3.14159
const Greeting = "hello"
const Debug = true

// Grouped, the common style at the top of a file:
const (
	MaxRetries = 3
	Timeout    = 30
)
§§§

A constant must be computable at *compile time* — §const t = time.Now()§ is rejected,
because §time.Now()§ only has a value once the program runs.

### Constants are "untyped" until you use them

This is a genuine Go speciality. A constant written without a type adapts to whatever
context it lands in:

§§§go
const Num = 10 // untyped constant

var a int = Num     // fine, Num acts like an int
var b float64 = Num // also fine, Num acts like a float64
var c int32 = Num   // also fine
§§§

A *variable* could never do that. This is why §const§ is more flexible than §var§ for
fixed values, and why §total := price * 3§ just works.

## iota: numbering things in a row

An **enumeration** is a set of named values — the days of the week, the states of an
order. Go has no §enum§ keyword; the idiom is a §const§ block using §iota§.

§iota§ is a counter that starts at 0 in every §const§ block and goes up by one on every
line:

§§§go
package main

import "fmt"

type Weekday int

const (
	Sunday Weekday = iota // 0
	Monday                // 1  (the expression "Weekday = iota" is implied)
	Tuesday               // 2
	Wednesday             // 3
)

func main() {
	fmt.Println(Sunday, Monday, Wednesday)
}
§§§

Output:

§§§text
0 1 3
§§§

Notice §type Weekday int§: we invented a **named type**. §Weekday§ is still an §int§
underneath, but the compiler now treats it as its own type, which makes function
signatures self-documenting and stops you passing a plain §int§ by mistake.

§iota§ can also do arithmetic:

§§§go
const (
	KB = 1 << (10 * (iota + 1)) // 1 shifted left 10 places = 1024
	MB                          // 1 shifted left 20 places = 1048576
	GB                          // 1 shifted left 30 places
)
§§§

§<<§ is "shift left", which multiplies by 2 that many times. If that is new to you,
just read §1 << 10§ as "1 shifted ten places left", which is 1024.

## Summary of the surprises

* §int§ is the default whole number; §float64§ is the default decimal number.
* §byte§ = §uint8§, §rune§ = §int32§.
* Conversion is always explicit: §float64(x)§.
* §int(9.99)§ gives §9§ — conversion truncates.
* Constants are untyped and adapt to context; variables do not.
* §iota§ builds enumerations inside a §const§ block.
`,
		Quiz: []Question{
			{
				ID:     "m1t3q1",
				Prompt: "Which lines compile? Select all that apply.",
				Multi:  true,
				Code: `var a int32 = 5
var x int = 10
var y float64 = 2.5

var p int64 = a
var q float64 = x
var r int = y
var s float64 = float64(a)
var t float32 = 1.5`,
				Options: []Option{
					{Text: "§var p int64 = a§", Correct: false, Rationale: "§int32§ → §int64§ is a widening conversion, and Go still refuses to do it implicitly."},
					{Text: "§var q float64 = x§", Correct: false, Rationale: "§x§ is a typed §int§, so §float64(x)§ is required."},
					{Text: "§var r int = y§", Correct: false, Rationale: "§float64§ → §int§ loses data, so Go makes you write §int(y)§."},
					{Text: "§var s float64 = float64(a)§", Correct: true, Rationale: "The conversion is explicit, so this is fine."},
					{Text: "§var t float32 = 1.5§", Correct: true, Rationale: "§1.5§ is an **untyped constant**, so it can become a §float32§ without any conversion."},
				},
				Explanation: "Variables never convert implicitly. Untyped constants do adapt to whatever type you need.",
			},
			{
				ID:     "m1t3q2",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

func main() {
	total := 100
	parts := 7
	fmt.Println(total / parts)
	fmt.Println(float64(total) / float64(parts))
	fmt.Println(total / parts * 2)
}`,
				Options: []Option{
					{Text: "§14§, §14.285714285714286§, §28§", Correct: true, Rationale: "Integer division gives 14, the float version gives the real ratio, and the last line does integer division *then* multiplies."},
					{Text: "§14.285714285714286§ three times", Correct: false, Rationale: "§total§ and §parts§ are both §int§."},
					{Text: "§14§, §14.5§, §28§", Correct: false, Rationale: "The second line converts before dividing, so nothing is rounded."},
					{Text: "It does not compile.", Correct: false, Rationale: "All three lines are valid Go."},
				},
				Explanation: "§int / int§ truncates. Convert to §float64§ *before* dividing. And §total / parts * 2§ is §(100/7)*2 = 14*2 = 28§ — order of operations matters.",
			},
			{
				ID:     "m1t3q3",
				Prompt: "What is the value of §Wednesday§?",
				Code: `const (
	Sunday = iota * 10
	Monday
	_
	Wednesday
)`,
				Options: []Option{
					{Text: "§30§", Correct: true, Rationale: "§iota§ counts 0,1,2,3 — including the blank line — so Wednesday uses the 4th index: §3 * 10§."},
					{Text: "§20§", Correct: false, Rationale: "That ignores the §_§ line, which still consumes an §iota§ value."},
					{Text: "§40§", Correct: false, Rationale: "§iota§ starts at 0, so the four lines are 0,1,2,3."},
					{Text: "It does not compile.", Correct: false, Rationale: "§_§ skips a value and is perfectly legal in a §const§ block."},
				},
				Explanation: "§iota§ is the line index within the §const§ block, starting at 0. §_§ still counts.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t3e1",
				Title: "Warm-up: Celsius to Fahrenheit",
				Prompt: `Implement the conversion §f = c*9/5 + 32§.

§§§go
func CelsiusToFahrenheit(c float64) float64
§§§

This is a small exercise with a big lesson hidden in it. If §c§ is a §float64§, then
§c * 9 / 5§ is all floating point and everything is fine. But if you ever write that
calculation with two §int§ values, §9 / 5§ is §1§ and the answer is wrong.

Worked values: §0 → 32§, §100 → 212§, §-40 → -40§.
`,
				Starter: `package main

import "fmt"

// CelsiusToFahrenheit converts a temperature from Celsius to Fahrenheit.
func CelsiusToFahrenheit(c float64) float64 {
	// TODO
	return 0
}

func main() {
	fmt.Println(CelsiusToFahrenheit(100)) // 212
	fmt.Println(CelsiusToFahrenheit(0))   // 32
}
`,
				Tests: []File{{
					Name: "temperature_test.go",
					Content: `package main

import (
	"math"
	"testing"
)

func almost(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestCelsiusToFahrenheit(t *testing.T) {
	cases := []struct{ in, want float64 }{
		{0, 32},
		{100, 212},
		{-40, -40},
		{37, 98.6},
		{21.5, 70.7},
	}
	for _, c := range cases {
		if got := CelsiusToFahrenheit(c.in); !almost(got, c.want) {
			t.Errorf("CelsiusToFahrenheit(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}
`,
				}},
				Hints: []string{
					"The literals §9§, §5§ and §32§ are untyped constants, so they adapt to §float64§ automatically.",
					"§return c*9/5 + 32§",
					"If you got a whole number, check that §c§ really is a §float64§ — integer division only happens between two ints.",
				},
				Solution: `package main

import "fmt"

func CelsiusToFahrenheit(c float64) float64 {
	return c*9/5 + 32
}

func main() {
	fmt.Println(CelsiusToFahrenheit(100))
	fmt.Println(CelsiusToFahrenheit(0))
}
`,
			},
			{
				ID:    "m1t3e2",
				Title: "Money: converting between types on purpose",
				Prompt: `Implement two helpers that show both directions of conversion.

§§§go
func TotalCents(price float64, quantity int) int
func AveragePerItem(totalCents int, items int) float64
§§§

Rules:

* §TotalCents§ takes a price in **dollars** (§19.99§) and a whole-number quantity, and
  returns the total in **whole cents** (§3998§). Convert the quantity to §float64§,
  multiply, then convert to §int§.
* §AveragePerItem§ divides whole cents by a number of items and returns **dollars** as
  a §float64§.
* If §items§ is §0§, return §0§ — never divide by zero.

An §if§ in Go looks like this — no parentheses around the condition, and the braces are
always required:

§§§go
if x < 0 {
	return 0
}
§§§

(There is a whole topic on §if§ later; this one shape is all you need for the guard.)

Watch the rounding. §19.99 * 2§ is §39.98§ mathematically, but as a float it is
§39.979999999999997§, so §int§ gives §3997§ — one cent short. Adding §0.5§ before
converting rounds to the nearest whole number instead of truncating. That is the
standard fix for money.
`,
				Starter: `package main

import "fmt"

// TotalCents converts a dollar price and a quantity into whole cents.
func TotalCents(price float64, quantity int) int {
	// TODO
	return 0
}

// AveragePerItem returns the average price in dollars for the given items.
func AveragePerItem(totalCents int, items int) float64 {
	// TODO
	return 0
}

func main() {
	fmt.Println(TotalCents(19.99, 2))    // 3998
	fmt.Println(AveragePerItem(3998, 2)) // 19.99
}
`,
				Tests: []File{{
					Name: "money_test.go",
					Content: `package main

import (
	"math"
	"testing"
)

func closeEnough(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestTotalCents(t *testing.T) {
	cases := []struct {
		price    float64
		quantity int
		want     int
	}{
		{19.99, 2, 3998},
		{1.00, 1, 100},
		{0.05, 3, 15},
		{9.99, 0, 0},
		{10.10, 10, 10100},
	}
	for _, c := range cases {
		if got := TotalCents(c.price, c.quantity); got != c.want {
			t.Errorf("TotalCents(%v, %d) = %d, want %d", c.price, c.quantity, got, c.want)
		}
	}
}

func TestAveragePerItem(t *testing.T) {
	if got := AveragePerItem(3998, 2); !closeEnough(got, 19.99) {
		t.Errorf("AveragePerItem(3998, 2) = %v, want 19.99", got)
	}
	if got := AveragePerItem(100, 0); got != 0 {
		t.Errorf("AveragePerItem(100, 0) = %v, want 0 (guard the division)", got)
	}
	if got := AveragePerItem(0, 5); got != 0 {
		t.Errorf("AveragePerItem(0, 5) = %v, want 0", got)
	}
}
`,
				}},
				Hints: []string{
					"§float64(quantity)§ converts the int so the multiplication stays in floating point.",
					"§int(price*float64(quantity)*100 + 0.5)§ — the §+ 0.5§ turns truncation into rounding.",
					"§AveragePerItem§: use the §if§ guard shape from the prompt to return early for §0§ items, then §return float64(totalCents) / float64(items) / 100§.",
				},
				Solution: `package main

import "fmt"

func TotalCents(price float64, quantity int) int {
	return int(price*float64(quantity)*100 + 0.5)
}

func AveragePerItem(totalCents int, items int) float64 {
	if items == 0 {
		return 0
	}
	return float64(totalCents) / float64(items) / 100
}

func main() {
	fmt.Println(TotalCents(19.99, 2))
	fmt.Println(AveragePerItem(3998, 2))
}
`,
			},
			{
				ID:    "m1t3e3",
				Title: "An iota enumeration for order status",
				Prompt: `Build an §iota§ enumeration and a function that turns each value
back into a readable word.

§§§go
type Status int

const (
	Pending Status = iota
	Paid
	Shipped
	Cancelled
)

func StatusName(s Status) string
§§§

§StatusName§ returns §"pending"§, §"paid"§, §"shipped"§ or §"cancelled"§. For any other
value, including negatives, return §"unknown"§.

Branch on the value with a §switch§ — Go's multi-way branch. Each §case§ stops on its
own (there is no fall-through, so no §break§ is needed), and §default§ catches
everything else:

§§§go
switch day {
case "sat", "sun":
	return "weekend"
default:
	return "unknown"
}
§§§

(§switch§ gets its own topic later; this shape is all you need here.)
`,
				Starter: `package main

import "fmt"

type Status int

const (
	Pending Status = iota
	Paid
	Shipped
	Cancelled
)

// StatusName returns the lower-case name of a status.
func StatusName(s Status) string {
	// TODO
	return ""
}

func main() {
	fmt.Println(StatusName(Pending), StatusName(Shipped), StatusName(99))
}
`,
				Tests: []File{{
					Name: "status_test.go",
					Content: `package main

import "testing"

func TestStatusName(t *testing.T) {
	cases := map[Status]string{
		Pending:      "pending",
		Paid:         "paid",
		Shipped:      "shipped",
		Cancelled:    "cancelled",
		Status(-1):   "unknown",
		Status(4):    "unknown",
		Status(1000): "unknown",
	}
	for in, want := range cases {
		if got := StatusName(in); got != want {
			t.Errorf("StatusName(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestIotaOrder(t *testing.T) {
	if Pending != 0 || Paid != 1 || Shipped != 2 || Cancelled != 3 {
		t.Fatalf("iota values wrong: %d %d %d %d", Pending, Paid, Shipped, Cancelled)
	}
}
`,
				}},
				Hints: []string{
					"§iota§ gives Pending 0, Paid 1, Shipped 2, Cancelled 3 — you get that for free from the §const§ block.",
					"One §case§ per named status, each returning its lower-case word; the §default§ from the prompt handles everything else.",
					"§Status(99)§ and §Status(-1)§ match none of the named cases, so they fall to §default§.",
				},
				Solution: `package main

import "fmt"

type Status int

const (
	Pending Status = iota
	Paid
	Shipped
	Cancelled
)

func StatusName(s Status) string {
	switch s {
	case Pending:
		return "pending"
	case Paid:
		return "paid"
	case Shipped:
		return "shipped"
	case Cancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

func main() {
	fmt.Println(StatusName(Pending), StatusName(Shipped), StatusName(99))
}
`,
			},
		},
	}
}

func m1t4() Topic {
	return Topic{
		ID:      "m1t4",
		Title:   "Formatted output and input with fmt",
		Summary: "The printf verbs, and reading values from the user.",
		Theory: `
§fmt§ (short for *format*) handles printing and scanning. It is the most-used package in
the standard library.

## Printing: the three you will actually use

§§§go
fmt.Println("a", 1, true)         // prints each value, spaces between, newline at the end
fmt.Print("a", 1, true)           // same, but NO spaces between values and no newline
fmt.Printf("%s is %d\n", "x", 1)  // full control; does NOT add a newline for you
§§§

Output:

§§§text
a 1 true
a1 true
x is 1
§§§

Everything after §Printf§ is a **format string** plus the values to pour into it.
§Printf§ does not add a newline — that is what §\n§ is for. Forgetting it is the single
most common beginner bug.

## The format string

A format string is ordinary text with **verbs** in it. A verb starts with §%§ and says
"put a value here, formatted like this".

| Verb | Meaning | Example input | Output |
| --- | --- | --- | --- |
| §%v§ | default format, works for anything | §[]int{1,2}§ | §[1 2]§ |
| §%d§ | integer, decimal | §42§ | §42§ |
| §%s§ | string | §"hi"§ | §hi§ |
| §%q§ | string, quoted and escaped | §"hi"§ | §"hi"§ |
| §%t§ | boolean | §true§ | §true§ |
| §%f§ | float, 6 decimals by default | §3.5§ | §3.500000§ |
| §%.2f§ | float with 2 decimals | §3.14159§ | §3.14§ |
| §%8.2f§ | float, width 8, right-aligned | §3.14159§ | §    3.14§ |
| §%-8s§ | string, width 8, LEFT-aligned | §"ab"§ | §ab      § |
| §%T§ | the type of the value | §3.5§ | §float64§ |
| §%+v§ | like §%v§ but with struct field names | §Point{1,2}§ | §{X:1 Y:2}§ |
| §%%§ | a literal percent sign | — | §%§ |

The common verbs in one program:

§§§go
package main

import "fmt"

type Point struct{ X, Y int }

func main() {
	name := "Ada"
	age := 36
	price := 19.9876
	ok := true

	fmt.Printf("%s is %d, spent %.2f, ok=%t\n", name, age, price, ok)
	fmt.Printf("%q vs %s\n", name, name)
	fmt.Printf("type of price is %T\n", price)
	fmt.Printf("%v | %+v\n", Point{1, 2}, Point{1, 2})
	fmt.Printf("|%-6s|%6s|\n", "left", "right")
}
§§§

Output (the spacing matters — count the characters):

§§§text
Ada is 36, spent 19.99, ok=true
"Ada" vs Ada
type of price is float64
{1 2} | {X:1 Y:2}
|left  | right|
§§§

Notice §%.2f§ printed §19.99§, not §19.9876§ — it rounds. And §%-6s§ padded §left§ to
six characters on the right so the column lines up.

## Sprintf: formatting into a string instead of printing

When you want the text as a *value* (to return it, store it, compare it), use
§Sprintf§ — the §S§ stands for *string*:

§§§go
func Receipt(item string, qty int, unit float64) string {
	return fmt.Sprintf("%d x %s = $%.2f", qty, item, float64(qty)*unit)
}
§§§

§Sprintf§ prints nothing. It hands you back the string and lets you decide what to do
with it. This is the function used in the first exercise.

## Reading input

§fmt.Scan§ reads whitespace-separated values from **standard input** — the text a user
types, or, in our editor, whatever you put in the **Input** box.

§§§go
package main

import "fmt"

func main() {
	var name string
	var age int
	var height float64

	fmt.Scan(&name, &age, &height)

	fmt.Printf("%s is %d years old and %.1f m tall\n", name, age, height)
}
§§§

With the input §Ada 36 1.7§, the output is:

§§§text
Ada is 36 years old and 1.7 m tall
§§§

Each §&§ means "put the value here". Read §&age§ as "the place where §age§ lives". The
compiler will not accept §fmt.Scan(name)§ without it, because §Scan§ returns values
*through* the pointers you hand it. Pointers get a full topic later; this is the first
place you use them.

Two practical notes:

* §Scan§ splits on spaces, tabs and newlines, so you do not need to know how the input
  is laid out across lines.
* If robustness matters, check the returned values: §n, err := fmt.Scan(&a)§. Given a
  word where a number belongs, §err§ is non-nil and §a§ keeps its zero value.

## Sizing things up

| Situation | Tool |
| --- | --- |
| A human-readable line of output | §fmt.Println§ |
| You need the text back as a string | §fmt.Sprintf§ |
| Columns, decimals, padding | §fmt.Printf§ |
| Reading typed-in values | §fmt.Scan§ plus §&§ for each target |
`,
		Quiz: []Question{
			{
				ID:     "m1t4q1",
				Prompt: "What does this print? (Count the spaces carefully.)",
				Code: `package main

import "fmt"

func main() {
	fmt.Printf("[%-5d]", 42)
	fmt.Printf("[%5d]", 42)
	fmt.Printf("[%.3f]", 3.14159)
	fmt.Printf("[%.2f]", 2.005)
}`,
				Options: []Option{
					{Text: "§[42   ][   42][3.142][2.00]§", Correct: true, Rationale: "§%-5d§ pads on the right, §%5d§ pads on the left, §%.3f§ rounds §3.14159§ to §3.142§, and §%.2f§ gives exactly two decimals."},
					{Text: "§[42   ][   42][3.141][2.00]§", Correct: false, Rationale: "§%.3f§ rounds the third decimal up: §3.14159 → 3.142§. Precision verbs round, they do not truncate."},
					{Text: "§[42][42][3.142][2.01]§", Correct: false, Rationale: "The width numbers in §%-5d§ and §%5d§ really do pad the output."},
					{Text: "§[42   ][   42][3.142][2.0]§", Correct: false, Rationale: "§%.2f§ always shows two decimal places, even when they are zero."},
				},
				Explanation: "Width is part of the verb: §%5d§ is 5 characters wide, §%-5d§ is the same but left-aligned. Precision (§.3f§) rounds.",
			},
			{
				ID:     "m1t4q2",
				Prompt: "Which line reads a name and an age from the user?",
				Options: []Option{
					{Text: "§fmt.Scan(&name, &age)§", Correct: true, Rationale: "§Scan§ must be given the addresses of the variables it should fill in."},
					{Text: "§fmt.Scan(name, age)§", Correct: false, Rationale: "Without §&§, §Scan§ cannot write into your variables. The compiler rejects it: §cannot use name as *string§."},
					{Text: "§fmt.Print(&name, &age)§", Correct: false, Rationale: "§Print§ writes to the screen; it never reads input."},
					{Text: "§name, age = fmt.Scan()§", Correct: false, Rationale: "§Scan§ does not return values that way."},
				},
				Explanation: "Input functions need a *place to put the answer*. §&name§ says \"here\". You will understand exactly why in the pointers topic.",
			},
			{
				ID:     "m1t4q3",
				Prompt: "Why does this print everything on one line?",
				Code: `package main

import "fmt"

func main() {
	fmt.Printf("one ")
	fmt.Printf("two ")
	fmt.Printf("three")
}`,
				Options: []Option{
					{Text: "Because §Printf§ never adds a newline — you must include §\\n§ in the format string yourself.", Correct: true, Rationale: "§Println§ adds one; §Print§ and §Printf§ do not."},
					{Text: "Because the format strings are missing verbs.", Correct: false, Rationale: "A format string with no verbs is perfectly legal — it just prints its text."},
					{Text: "Because §fmt§ buffers output until the program ends.", Correct: false, Rationale: "There is no such buffering here."},
					{Text: "It does not compile.", Correct: false, Rationale: "It compiles and runs."},
				},
				Explanation: "§Println§ = print line (adds §\\n§). §Print§ and §Printf§ leave newlines to you.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t4e1",
				Title: "Warm-up: a formatted receipt line",
				Prompt: `Return a single receipt line, formatted exactly like this:

§§§text
3 x Coffee = $7.50
§§§

§§§go
func ReceiptLine(item string, qty int, unitPrice float64) string
§§§

* The total is §float64(qty) * unitPrice§.
* The price is shown with exactly **two decimals** and a dollar sign.
* Format: §<qty> x <item> = $<total>§

Examples:

| Call | Result |
| --- | --- |
| §ReceiptLine("Coffee", 3, 2.5)§ | §3 x Coffee = $7.50§ |
| §ReceiptLine("Tea", 1, 3)§ | §1 x Tea = $3.00§ |
| §ReceiptLine("Muffin", 0, 4.25)§ | §0 x Muffin = $0.00§ |
`,
				Starter: `package main

import "fmt"

// ReceiptLine formats one line of a receipt.
func ReceiptLine(item string, qty int, unitPrice float64) string {
	// TODO
	return ""
}

func main() {
	fmt.Println(ReceiptLine("Coffee", 3, 2.5))
	fmt.Println(ReceiptLine("Tea", 1, 3))
}
`,
				Tests: []File{{
					Name: "receipt_test.go",
					Content: `package main

import "testing"

func TestReceiptLine(t *testing.T) {
	cases := []struct {
		item  string
		qty   int
		price float64
		want  string
	}{
		{"Coffee", 3, 2.5, "3 x Coffee = $7.50"},
		{"Tea", 1, 3, "1 x Tea = $3.00"},
		{"Muffin", 0, 4.25, "0 x Muffin = $0.00"},
		{"Bagel", 10, 1.05, "10 x Bagel = $10.50"},
	}
	for _, c := range cases {
		if got := ReceiptLine(c.item, c.qty, c.price); got != c.want {
			t.Errorf("ReceiptLine(%q, %d, %v) = %q, want %q", c.item, c.qty, c.price, got, c.want)
		}
	}
}
`,
				}},
				Hints: []string{
					"§fmt.Sprintf§ returns the formatted text; §fmt.Printf§ would print it instead.",
					"§%.2f§ gives exactly two decimals.",
					"§return fmt.Sprintf(\"%d x %s = $%.2f\", qty, item, float64(qty)*unitPrice)§",
				},
				Solution: `package main

import "fmt"

func ReceiptLine(item string, qty int, unitPrice float64) string {
	return fmt.Sprintf("%d x %s = $%.2f", qty, item, float64(qty)*unitPrice)
}

func main() {
	fmt.Println(ReceiptLine("Coffee", 3, 2.5))
	fmt.Println(ReceiptLine("Tea", 1, 3))
}
`,
			},
			{
				ID:    "m1t4e2",
				Title: "Read input and report it back, formatted",
				Prompt: `Read two values from standard input — a name (text) and an age (whole
number) — and print a one-line summary with the person's age next year.

Input:

§§§text
Ada 36
§§§

Output:

§§§text
Ada is 36, and will be 37 next year.
§§§

Put the formatting in a function so it can be tested:

§§§go
func Summarize(name string, age int) string
§§§

§main§ reads the two values with §fmt.Scan§ (remember the §&§) and prints
§Summarize(name, age)§.
`,
				Stdin:    true,
				SampleIn: "Ada 36",
				Starter: `package main

import "fmt"

// Summarize builds the one-line summary.
func Summarize(name string, age int) string {
	// TODO
	return ""
}

func main() {
	var name string
	var age int
	fmt.Scan(&name, &age)
	fmt.Println(Summarize(name, age))
}
`,
				Tests: []File{{
					Name: "summarize_test.go",
					Content: `package main

import "testing"

func TestSummarize(t *testing.T) {
	cases := []struct {
		name string
		age  int
		want string
	}{
		{"Ada", 36, "Ada is 36, and will be 37 next year."},
		{"Bo", 0, "Bo is 0, and will be 1 next year."},
		{"Cy", 99, "Cy is 99, and will be 100 next year."},
	}
	for _, c := range cases {
		if got := Summarize(c.name, c.age); got != c.want {
			t.Errorf("Summarize(%q, %d) = %q, want %q", c.name, c.age, got, c.want)
		}
	}
}
`,
				}},
				Hints: []string{
					"§fmt.Scan(&name, &age)§ reads two whitespace-separated values.",
					"The sentence ends with a full stop, and there is a comma after the age.",
					"§return fmt.Sprintf(\"%s is %d, and will be %d next year.\", name, age, age+1)§",
				},
				Solution: `package main

import "fmt"

func Summarize(name string, age int) string {
	return fmt.Sprintf("%s is %d, and will be %d next year.", name, age, age+1)
}

func main() {
	var name string
	var age int
	fmt.Scan(&name, &age)
	fmt.Println(Summarize(name, age))
}
`,
			},
		},
	}
}

func m1t5() Topic {
	return Topic{
		ID:      "m1t5",
		Title:   "Conditionals: if and switch",
		Summary: "Branching, and the two shapes of switch.",
		Theory: `
Go's §if§ and §switch§ differ from their C and Java equivalents in three ways: the
condition is not parenthesised, the braces are mandatory even for a single statement, and
both statements may declare a variable scoped to their branches.

## if

The simplest form:

§§§go
if score >= 90 {
	fmt.Println("A")
}
§§§

With alternatives:

§§§go
if score >= 90 {
	fmt.Println("A")
} else if score >= 80 {
	fmt.Println("B")
} else {
	fmt.Println("try again")
}
§§§

Two rules differ from most languages:

1. **The braces are mandatory.** §if x > 0 fmt.Println("yes")§ does not compile.
2. **The condition must be a §bool§.** §if 1 {§ and §if name {§ are compile errors.
   There is no truthiness: no "0 is false", no "empty string is false". Write
   §if name != "" {§ and §if count > 0 {§.

### if with a short statement

An §if§ may begin with a small statement, separated from the condition by a semicolon.
Anything declared there exists **only inside the §if§**:

§§§go
if remainder := 10 % 3; remainder == 0 {
	fmt.Println("10 is divisible by 3")
} else {
	fmt.Println("remainder is", remainder) // visible here, in the else branch
}
// fmt.Println(remainder) // ERROR: undefined outside the if
§§§

Output:

§§§text
remainder is 1
§§§

Why bother? Because it keeps a value's lifetime exactly as long as it is needed. You
will use this constantly for errors and lookups in a few topics' time.

## switch

§switch§ compares one value against several possibilities. Its most important
property: **a matching case stops there — you do not need §break§.**

§§§go
switch day {
case "sat", "sun":
	fmt.Println("weekend")
case "mon", "tue", "wed", "thu", "fri":
	fmt.Println("weekday")
default:
	fmt.Println("not a day I know")
}
§§§

* A case can list several values separated by commas.
* Only the **first** matching case runs.
* §default§ runs when nothing matched. It can appear anywhere, but readers expect last.

**Java comparison:** there you must remember §break§, and forgetting it gives you the
notorious fall-through bug. Go inverts the problem: the fall-through must be requested
explicitly with §fallthrough§.

### switch with no value

Drop the value after §switch§ and each case becomes a boolean condition. This is Go's
clean replacement for long §else if§ chains:

§§§go
switch {
case n < 0:
	fmt.Println("negative")
case n == 0:
	fmt.Println("zero")
default:
	fmt.Println("positive")
}
§§§

§switch {§ is short for §switch true {§ — the first case whose condition is true wins.

### switch with a short statement, too

§§§go
switch grade := score / 10; grade {
case 10, 9:
	fmt.Println("A")
case 8:
	fmt.Println("B")
default:
	fmt.Println("C or below")
}
§§§

## A complete example

§§§go
package main

import "fmt"

func classify(n int) string {
	switch {
	case n < 0:
		return "negative"
	case n == 0:
		return "zero"
	case n%2 == 0:
		return "even"
	default:
		return "odd"
	}
}

func main() {
	for _, n := range []int{-4, 0, 6, 7} {
		fmt.Printf("%d is %s\n", n, classify(n))
	}
}
§§§

Output:

§§§text
-4 is negative
0 is zero
6 is even
7 is odd
§§§

The order of the cases matters: §n < 0§ is checked first, so §-4§ is reported as
negative rather than even.

## Combining conditions

Comparisons work on strings too, and §&&§ (and), §||§ (or) and §!§ (not) combine them:

§§§go
if user != "" && (age >= 18 || hasConsent) {
	fmt.Println("allowed")
}
§§§

Two style notes:

* Use §&&§ and §||§, not the words §and§/§or§.
* Prefer positive conditions. §if !notReady§ is a puzzle; §if ready§ is not.

## if vs switch: which one?

| Use §if§ when | Use §switch§ when |
| --- | --- |
| There is one condition, or two | You are comparing one value against several |
| The conditions are unrelated | The cases are mutually exclusive ranges or values |
| You need short-circuit logic with §&&§/§||§ | You want a table of possibilities, one per line |

Both are always available; the compiler does not care. Pick the one a tired reader
understands fastest.
`,
		Quiz: []Question{
			{
				ID:     "m1t5q1",
				Prompt: "Which of these compile? Select all that apply.",
				Multi:  true,
				Code: `n := 5
s := "hi"
b := true

if n > 3 { }
if b { }
if s { }
if n { }
if n > 3 && s != "" { }`,
				Options: []Option{
					{Text: "§if n > 3 { }§", Correct: true, Rationale: "The condition is a bool."},
					{Text: "§if b { }§", Correct: true, Rationale: "§b§ is already a bool."},
					{Text: "§if s { }§", Correct: false, Rationale: "§s§ is a string. Go has no truthiness; a non-empty string is not \"true\"."},
					{Text: "§if n { }§", Correct: false, Rationale: "§n§ is an int. 0 is not false in Go."},
					{Text: "§if n > 3 && s != \"\" { }§", Correct: true, Rationale: "Both sides are booleans, combined with §&&§."},
				},
				Explanation: "A Go §if§ needs a §bool§, full stop. Java's §if (x)§ and Python's §if x:§ habits both fail here — write the comparison out.",
			},
			{
				ID:     "m1t5q2",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

func main() {
	n := 15
	switch {
	case n%3 == 0:
		fmt.Print("Fizz")
	case n%5 == 0:
		fmt.Print("Buzz")
	case n%3 == 0 && n%5 == 0:
		fmt.Print("FizzBuzz")
	default:
		fmt.Print(n)
	}
	fmt.Println()
}`,
				Options: []Option{
					{Text: "§Fizz§", Correct: true, Rationale: "The first matching case wins and the switch exits — the §FizzBuzz§ case is never even reached."},
					{Text: "§FizzBuzz§", Correct: false, Rationale: "That is the *goal* of the game, but this code checks the wrong case first."},
					{Text: "§FizzBuzzFizz§", Correct: false, Rationale: "Go does not fall through."},
					{Text: "§15§", Correct: false, Rationale: "§15 % 3 == 0§, so the first case matches."},
				},
				Explanation: "The §n%3==0 && n%5==0§ check must come FIRST. This is a classic ordering bug, and it is exactly what the FizzBuzz exercise asks you to get right.",
			},
			{
				ID:     "m1t5q3",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

func main() {
	if x := 10; x > 5 {
		fmt.Println("big", x)
	} else {
		fmt.Println("small", x)
	}
}`,
				Options: []Option{
					{Text: "§big 10§", Correct: true, Rationale: "§x§ is created by the init statement and is visible in both branches."},
					{Text: "§small 10§", Correct: false, Rationale: "§10 > 5§ is true, so the §if§ branch runs."},
					{Text: "It does not compile: §x§ is undefined in the body.", Correct: false, Rationale: "The init statement's variables are in scope for the whole §if§/§else§."},
					{Text: "It does not compile: §x§ is declared and not used.", Correct: false, Rationale: "§x§ *is* used, in the §Println§."},
				},
				Explanation: "§if stmt; cond { }§ scopes §stmt§'s variables to the entire if/else chain — and to nothing beyond it.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t5e1",
				Title: "Warm-up: letter grades",
				Prompt: `Turn a score into a letter grade.

§§§go
func Grade(score int) string
§§§

| Score | Grade |
| --- | --- |
| 90 or more | §"A"§ |
| 80–89 | §"B"§ |
| 70–79 | §"C"§ |
| 60–69 | §"D"§ |
| below 60 | §"F"§ |

Use an §if / else if / else§ chain. The cases are checked top to bottom, so start with
the highest.
`,
				Starter: `package main

import "fmt"

// Grade maps a score to a letter grade.
func Grade(score int) string {
	// TODO
	return ""
}

func main() {
	for _, s := range []int{95, 85, 75, 65, 40, 0} {
		fmt.Println(s, Grade(s))
	}
}
`,
				Tests: []File{{
					Name: "grade_test.go",
					Content: `package main

import "testing"

func TestGrade(t *testing.T) {
	cases := map[int]string{
		100: "A", 90: "A", 89: "B", 80: "B", 79: "C",
		70: "C", 69: "D", 60: "D", 59: "F", 0: "F", -10: "F",
	}
	for score, want := range cases {
		if got := Grade(score); got != want {
			t.Errorf("Grade(%d) = %q, want %q", score, got, want)
		}
	}
}
`,
				}},
				Hints: []string{
					"Order the branches from highest to lowest: §if score >= 90 { return \"A\" } else if score >= 80 { ... }§.",
					"The final §else§ (or a bare §return \"F\"§) catches everything below 60, including negatives.",
					"Only the *first* true branch runs, so you never need §score >= 80 && score < 90§.",
				},
				Solution: `package main

import "fmt"

func Grade(score int) string {
	if score >= 90 {
		return "A"
	} else if score >= 80 {
		return "B"
	} else if score >= 70 {
		return "C"
	} else if score >= 60 {
		return "D"
	}
	return "F"
}

func main() {
	for _, s := range []int{95, 85, 75, 65, 40, 0} {
		fmt.Println(s, Grade(s))
	}
}
`,
			},
			{
				ID:    "m1t5e2",
				Title: "Days of the week with switch",
				Prompt: `Classify a day name.

§§§go
func DayType(day string) string
§§§

| Input | Output |
| --- | --- |
| §"sat"§, §"sun"§ | §"weekend"§ |
| §"mon"§ … §"fri"§ | §"weekday"§ |
| anything else, including §""§ | §"unknown"§ |

The five weekdays do not need five separate cases:

§§§go
switch day {
case "sat", "sun":
	return "weekend"
case "mon", "tue", "wed", "thu", "fri":
	return "weekday"
default:
	return "unknown"
}
§§§
`,
				Starter: `package main

import "fmt"

// DayType reports whether a day name is a weekday, a weekend day, or unknown.
func DayType(day string) string {
	// TODO
	return ""
}

func main() {
	for _, d := range []string{"sat", "mon", "friday", ""} {
		fmt.Println(d, DayType(d))
	}
}
`,
				Tests: []File{{
					Name: "daytype_test.go",
					Content: `package main

import "testing"

func TestDayType(t *testing.T) {
	cases := map[string]string{
		"sat": "weekend", "sun": "weekend",
		"mon": "weekday", "tue": "weekday", "wed": "weekday",
		"thu": "weekday", "fri": "weekday",
		"": "unknown", "friday": "unknown", "SAT": "unknown", "xyz": "unknown",
	}
	for day, want := range cases {
		if got := DayType(day); got != want {
			t.Errorf("DayType(%q) = %q, want %q", day, got, want)
		}
	}
}
`,
				}},
				Hints: []string{
					"Multiple values in one case are separated by commas: §case \"sat\", \"sun\":§.",
					"The comparison is case-sensitive, so §SAT§ falls into §default§.",
					"There is no need for §break§ — Go stops at the first matching case.",
				},
				Solution: `package main

import "fmt"

func DayType(day string) string {
	switch day {
	case "sat", "sun":
		return "weekend"
	case "mon", "tue", "wed", "thu", "fri":
		return "weekday"
	default:
		return "unknown"
	}
}

func main() {
	for _, d := range []string{"sat", "mon", "friday", ""} {
		fmt.Println(d, DayType(d))
	}
}
`,
			},
			{
				ID:    "m1t5e3",
				Title: "The largest of three, written twice",
				Prompt: `Return the largest of three integers.

§§§go
func Max3(a, b, c int) int
§§§

Write it twice in the same file to feel the difference:

1. §Max3§ with a chain of §if§ statements.
2. §Max3Switch§ using §switch {§ with boolean cases — the expression-less switch.

Both must behave identically. Think about which one you would rather read next year.
`,
				Starter: `package main

import "fmt"

// Max3 returns the largest of a, b and c, using if statements.
func Max3(a, b, c int) int {
	// TODO
	return 0
}

// Max3Switch returns the largest of a, b and c, using a switch with no value.
func Max3Switch(a, b, c int) int {
	// TODO
	return 0
}

func main() {
	fmt.Println(Max3(1, 9, 4), Max3Switch(1, 9, 4))
}
`,
				Tests: []File{{
					Name: "max3_test.go",
					Content: `package main

import "testing"

func TestMax3(t *testing.T) {
	fns := map[string]func(int, int, int) int{"Max3": Max3, "Max3Switch": Max3Switch}
	cases := []struct{ a, b, c, want int }{
		{1, 9, 4, 9},
		{9, 1, 4, 9},
		{1, 4, 9, 9},
		{5, 5, 5, 5},
		{-1, -7, -3, -1},
		{0, 0, -1, 0},
	}
	for name, fn := range fns {
		for _, c := range cases {
			if got := fn(c.a, c.b, c.c); got != c.want {
				t.Errorf("%s(%d, %d, %d) = %d, want %d", name, c.a, c.b, c.c, got, c.want)
			}
		}
	}
}
`,
				}},
				Hints: []string{
					"§if§ version: start with §max := a§, then §if b > max { max = b }§ and the same for §c§.",
					"§switch§ version: §switch { case a >= b && a >= c: return a ... }§",
					"Both must handle equal values — use §>=§ rather than §>§ in the switch conditions.",
				},
				Solution: `package main

import "fmt"

func Max3(a, b, c int) int {
	max := a
	if b > max {
		max = b
	}
	if c > max {
		max = c
	}
	return max
}

func Max3Switch(a, b, c int) int {
	switch {
	case a >= b && a >= c:
		return a
	case b >= a && b >= c:
		return b
	default:
		return c
	}
}

func main() {
	fmt.Println(Max3(1, 9, 4), Max3Switch(1, 9, 4))
}
`,
			},
		},
	}
}
