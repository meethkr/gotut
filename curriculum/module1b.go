package curriculum

func m1t6() Topic {
	return Topic{
		ID:      "m1t6",
		Title:   "Loops: for is Go's only loop",
		Summary: "Three ways to write the same keyword, plus break and continue.",
		Theory: `
Java has §for§, §while§ and §do…while§. Python has §for§ and §while§. Go has exactly
one keyword — §for§ — and it can express all of them.

## The three-clause loop

§§§go
for i := 0; i < 5; i++ {
	fmt.Println(i)
}
§§§

Output:

§§§text
0
1
2
3
4
§§§

Read it as three parts separated by semicolons:

| Part | Meaning | Optional? |
| --- | --- | --- |
| §i := 0§ | runs once, before the loop | yes |
| §i < 5§ | checked before every iteration; stop when false | yes |
| §i++§ | runs after every iteration | yes |

§i++§ means "add one to §i§". §i--§ subtracts one. Both are statements, not
expressions, so §x := i++§ is not valid Go.

## Shape 2: the while loop

Leave out the first and last parts and you get a §while§:

§§§go
n := 1024
for n > 1 {
	n = n / 2
}
fmt.Println(n)
§§§

Output:

§§§text
1
§§§

This is exactly Python's §while n > 1:§ or Java's §while (n > 1)§. Go just spells it
§for§.

## Shape 3: forever

Leave out everything:

§§§go
for {
	fmt.Println("once")
	break // without this the program never stops
}
§§§

§for { }§ is an infinite loop. You leave it with §break§, §return§, or by making the
program exit. A server usually has one of these at the bottom of §main§.

Try removing the §break§ in the editor and pressing **Run**: our sandbox kills the
program after a few seconds and tells you it timed out instead of hanging forever.

## Shape 4: range — visiting a collection

§§§go
for i, name := range []string{"Ada", "Bo", "Cy"} {
	fmt.Println(i, name)
}
§§§

Output:

§§§text
0 Ada
1 Bo
2 Cy
§§§

§range§ hands you two values: the **index** and a **copy** of the element. If you only
want the index, write §for i := range xs§. If you do not need the index at all, use the
blank identifier:

§§§go
for _, name := range names {
	fmt.Println(name)
}
§§§

The underscore §_§ means "I am deliberately throwing this away". Go will not let you
leave a variable unused, so §_§ is the way to say "not interested".

Since Go 1.22 you can also range over an integer, which simply counts:

§§§go
for i := range 3 {
	fmt.Print(i, " ") // 0 1 2
}
§§§

That is a tidy replacement for §for i := 0; i < 3; i++§ when you do not need §i§ inside
the body.

## break and continue

* §break§ leaves the loop entirely.
* §continue§ abandons the current iteration and jumps to the next one.

§§§go
for i := 1; i <= 10; i++ {
	if i%2 == 0 {
		continue // skip even numbers
	}
	if i > 7 {
		break // stop entirely once we pass 7
	}
	fmt.Print(i, " ")
}
§§§

Output:

§§§text
1 3 5 7
§§§

## A worked example: summing and counting

§§§go
package main

import "fmt"

func main() {
	nums := []int{4, 8, 15, 16, 23, 42}

	sum := 0
	for _, n := range nums {
		sum += n
	}
	fmt.Println("sum:", sum)
	fmt.Println("average:", float64(sum)/float64(len(nums)))

	// count how many are even, stopping early if we find one over 40
	evens := 0
	for _, n := range nums {
		if n > 40 {
			fmt.Println("stopping early at", n)
			break
		}
		if n%2 == 0 {
			evens++
		}
	}
	fmt.Println("evens before 40:", evens)
}
§§§

Output:

§§§text
sum: 108
average: 18
stopping early at 42
evens before 40: 3
§§§

## The nested-loop escape hatch

If you have two loops and want to break out of both, label the outer one:

§§§go
outer:
for i := 0; i < 3; i++ {
	for j := 0; j < 3; j++ {
		if i*j == 4 {
			fmt.Println("found at", i, j)
			break outer // leaves BOTH loops
		}
	}
}
§§§

Output:

§§§text
found at 2 2
§§§

A plain §break§ would only leave the inner loop. Labels also work with §continue§.

## Two things people trip over

* **Off-by-one.** §for i := 0; i <= len(xs); i++§ crashes with §index out of range§,
  because indices run §0 … len(xs)-1§. Use §<§.
* **Modifying a slice while ranging over it** is a common bug: §range§ fixes the length
  once, before the first iteration. To remove elements, build a new slice (the slices
  topic practises that).
`,
		Quiz: []Question{
			{
				ID:     "m1t6q1",
				Prompt: "What does this loop print?",
				Code: `for i := 0; i < 5; i++ {
	if i == 2 {
		continue
	}
	if i == 4 {
		break
	}
	fmt.Print(i, " ")
}`,
				Options: []Option{
					{Text: "§0 1 3 §", Correct: true, Rationale: "§continue§ skips printing §2§; §break§ stops before printing §4§."},
					{Text: "§0 1 2 3 4 §", Correct: false, Rationale: "§continue§ and §break§ both change the flow."},
					{Text: "§0 1 3 4 §", Correct: false, Rationale: "§break§ at §i == 4§ stops before printing 4."},
					{Text: "§2 4 §", Correct: false, Rationale: "The conditions *skip* those values; they do not select them."},
				},
				Explanation: "§continue§ jumps to the next iteration; §break§ leaves the loop. The §i == 4§ check never runs for §i == 4§ because §break§ fires first.",
			},
			{
				ID:     "m1t6q2",
				Prompt: "What is printed?",
				Code: `package main

import "fmt"

func main() {
	xs := []string{"a", "b", "c"}
	for i, v := range xs {
		v = v + "!"
		xs[i] = v
	}
	fmt.Println(xs)
}`,
				Options: []Option{
					{Text: "§[a! b! c!]§", Correct: true, Rationale: "§v§ is a copy, but the loop writes the changed value back through §xs[i]§."},
					{Text: "§[a b c]§", Correct: false, Rationale: "The explicit write through §xs[i]§ does change the slice."},
					{Text: "§[a! a! a!]§", Correct: false, Rationale: "Each iteration reads a different element."},
					{Text: "It does not compile: §v§ cannot be assigned.", Correct: false, Rationale: "§v§ is an ordinary local variable and may be assigned."},
				},
				Explanation: "§range§ gives you a *copy* of each element. Changing the copy does nothing — but the loop body also writes back via the index, so the slice does change.",
			},
			{
				ID:     "m1t6q3",
				Prompt: "How many times does the body of this loop run?",
				Code: `count := 0
for count < 3 {
	count++
}
fmt.Println(count)`,
				Options: []Option{
					{Text: "Three times; it prints §3§.", Correct: true, Rationale: "The body runs while §count < 3§: for count = 0, 1 and 2, after which count is 3 and the loop stops."},
					{Text: "Four times; it prints §3§.", Correct: false, Rationale: "The condition is checked *before* the body. When §count§ becomes 3, the condition is false and the body is skipped."},
					{Text: "It never terminates.", Correct: false, Rationale: "§count§ grows, so the condition eventually fails."},
					{Text: "It does not compile.", Correct: false, Rationale: "§for cond { }§ is a valid while-loop."},
				},
				Explanation: "A §for cond {}§ loop checks the condition first, exactly like a while loop.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t6e1",
				Title: "Warm-up: add up the numbers from 1 to n",
				Prompt: `§§§go
func SumTo(n int) int
§§§

Returns §1 + 2 + … + n§.

* §SumTo(5)§ → §15§
* §SumTo(1)§ → §1§
* §SumTo(0)§ → §0§
* §SumTo(-3)§ → §0§ (no numbers to add)

Write it with an explicit loop — no clever formula please, the loop is the point.
`,
				Starter: `package main

import "fmt"

// SumTo returns 1 + 2 + ... + n, or 0 when n < 1.
func SumTo(n int) int {
	// TODO
	return 0
}

func main() {
	fmt.Println(SumTo(5), SumTo(0), SumTo(-3))
}
`,
				Tests: []File{{
					Name: "sumto_test.go",
					Content: `package main

import "testing"

func TestSumTo(t *testing.T) {
	cases := map[int]int{0: 0, 1: 1, 5: 15, 10: 55, 100: 5050, -3: 0}
	for n, want := range cases {
		if got := SumTo(n); got != want {
			t.Errorf("SumTo(%d) = %d, want %d", n, got, want)
		}
	}
}
`,
				}},
				Hints: []string{
					"Start with §total := 0§ and add inside the loop.",
					"§for i := 1; i <= n; i++ { total += i }§ handles §n <= 0§ for free, because the loop body never runs.",
					"Remember to §return total§, not §return 0§.",
				},
				Solution: `package main

import "fmt"

func SumTo(n int) int {
	total := 0
	for i := 1; i <= n; i++ {
		total += i
	}
	return total
}

func main() {
	fmt.Println(SumTo(5), SumTo(0), SumTo(-3))
}
`,
			},
			{
				ID:    "m1t6e2",
				Title: "FizzBuzz",
				Prompt: `A well-known exercise that turns on **case order**.

§§§go
func FizzBuzz(n int) string
§§§

| Condition | Return |
| --- | --- |
| divisible by both 3 and 5 | §"FizzBuzz"§ |
| divisible by 3 | §"Fizz"§ |
| divisible by 5 | §"Buzz"§ |
| otherwise | the number written out, e.g. §"7"§ |

For example: §1 → "1"§, §3 → "Fizz"§, §5 → "Buzz"§, §15 → "FizzBuzz"§, §30 → "FizzBuzz"§.

§strconv.Itoa(x)§ turns an integer into its decimal text. You will need to import
§"strconv"§.
`,
				Starter: `package main

import (
	"fmt"
	"strconv"
)

// FizzBuzz returns the FizzBuzz word for n.
func FizzBuzz(n int) string {
	// TODO
	return ""
}

func main() {
	for i := 1; i <= 20; i++ {
		fmt.Print(FizzBuzz(i), " ")
	}
	fmt.Println()
	_ = strconv.Itoa // you can delete this line once you use strconv
}
`,
				Tests: []File{{
					Name: "fizzbuzz_test.go",
					Content: `package main

import "testing"

func TestFizzBuzz(t *testing.T) {
	cases := map[int]string{
		1: "1", 2: "2", 3: "Fizz", 4: "4", 5: "Buzz",
		9: "Fizz", 10: "Buzz", 14: "14", 15: "FizzBuzz", 30: "FizzBuzz",
		45: "FizzBuzz", 100: "Buzz", 101: "101",
	}
	for n, want := range cases {
		if got := FizzBuzz(n); got != want {
			t.Errorf("FizzBuzz(%d) = %q, want %q", n, got, want)
		}
	}
}
`,
				}},
				Hints: []string{
					"The divisible-by-15 case must be checked FIRST, or 15 will match the divisible-by-3 case and return §\"Fizz\"§.",
					"§n%15 == 0§ is a compact way to say \"divisible by 3 and 5\".",
					"§strconv.Itoa(n)§ converts an int to its decimal string.",
				},
				Solution: `package main

import (
	"fmt"
	"strconv"
)

func FizzBuzz(n int) string {
	switch {
	case n%15 == 0:
		return "FizzBuzz"
	case n%3 == 0:
		return "Fizz"
	case n%5 == 0:
		return "Buzz"
	default:
		return strconv.Itoa(n)
	}
}

func main() {
	for i := 1; i <= 20; i++ {
		fmt.Print(FizzBuzz(i), " ")
	}
	fmt.Println()
}
`,
			},
			{
				ID:    "m1t6e3",
				Title: "Prime numbers with an early exit",
				Prompt: `§§§go
func IsPrime(n int) bool
func CountPrimes(limit int) int
§§§

* §IsPrime§ returns §false§ for anything below 2, and §true§ for 2, 3, 5, 7, 11…
* You do **not** need to test every divisor up to §n§. If §n§ has a factor, its
  smallest factor is at most §√n§, so looping while §i*i <= n§ is enough — and much
  faster for large numbers.
* Return §false§ the moment you find a divisor; no need to keep looking.
* §CountPrimes(limit)§ counts the primes in §[2, limit]§. §CountPrimes(100)§ is §25§.

This is your first taste of writing a loop whose number of iterations depends on the
value being examined.
`,
				Starter: `package main

import "fmt"

// IsPrime reports whether n is a prime number.
func IsPrime(n int) bool {
	// TODO
	return false
}

// CountPrimes counts the primes in [2, limit].
func CountPrimes(limit int) int {
	// TODO
	return 0
}

func main() {
	fmt.Println(IsPrime(97), CountPrimes(100))
}
`,
				Tests: []File{{
					Name: "primes_test.go",
					Content: `package main

import "testing"

func TestIsPrime(t *testing.T) {
	primes := []int{2, 3, 5, 7, 11, 13, 97, 7919}
	for _, n := range primes {
		if !IsPrime(n) {
			t.Errorf("IsPrime(%d) = false, want true", n)
		}
	}
	composites := []int{-7, 0, 1, 4, 9, 25, 100, 7920}
	for _, n := range composites {
		if IsPrime(n) {
			t.Errorf("IsPrime(%d) = true, want false", n)
		}
	}
}

func TestCountPrimes(t *testing.T) {
	cases := map[int]int{1: 0, 2: 1, 10: 4, 100: 25, 1000: 168}
	for limit, want := range cases {
		if got := CountPrimes(limit); got != want {
			t.Errorf("CountPrimes(%d) = %d, want %d", limit, got, want)
		}
	}
}
`,
				}},
				Hints: []string{
					"§if n < 2 { return false }§ handles negatives, 0 and 1 in one go.",
					"Loop with §for i := 2; i*i <= n; i++ { if n%i == 0 { return false } }§.",
					"After the loop finishes without finding a divisor, §return true§.",
					"§CountPrimes§ is a loop over §n§ calling §IsPrime(n)§ and counting the trues.",
				},
				Solution: `package main

import "fmt"

func IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func CountPrimes(limit int) int {
	count := 0
	for n := 2; n <= limit; n++ {
		if IsPrime(n) {
			count++
		}
	}
	return count
}

func main() {
	fmt.Println(IsPrime(97), CountPrimes(100))
}
`,
			},
		},
	}
}

func m1t7() Topic {
	return Topic{
		ID:      "m1t7",
		Title:   "Functions, multiple return values and variadics",
		Summary: "Multiple return values, named results and variadic parameters.",
		Theory: `
Three Go function features have no direct equivalent in Java or C: multiple return
values, variadic parameters, and functions used as ordinary values.

## The shape of a function

§§§go
func Add(a int, b int) int {
	return a + b
}
§§§

Read the parts as a sentence: *"§func§ Add, taking an §a§ that is an §int§ and a §b§
that is an §int§, gives back an §int§."*

When several consecutive parameters share a type you may group them:

§§§go
func Add(a, b int) int { // a and b are both int
	return a + b
}
§§§

## Returning more than one value

This is the feature that makes Go feel different from Java and C. A function can
return a tuple of values, and there is **no** need for out-parameters or wrapper
classes:

§§§go
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("cannot divide by zero")
	}
	return a / b, nil
}
§§§

Callers unwrap it with §:=§:

§§§go
result, err := Divide(10, 2)
if err != nil {
	fmt.Println("problem:", err)
	return
}
fmt.Println(result)
§§§

The convention from the first topic — **the error comes last** — comes from
here. A function either returns §(value, nil)§ on success or §(zeroValue, err)§ on
failure.

You can also return two values that are not a value-and-error pair, which is handy for
lookups:

§§§go
func FirstLast(xs []int) (int, int, bool) {
	if len(xs) == 0 {
		return 0, 0, false
	}
	return xs[0], xs[len(xs)-1], true
}
§§§

The trailing §bool§ answers "did that work?", exactly like the §ok§ in §m[k]§ for maps.

## Named results

You may name the results in the signature:

§§§go
func MinMax(xs []int) (min, max int) {
	min, max = xs[0], xs[0]
	for _, v := range xs[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return // bare return: "return whatever min and max currently are"
}
§§§

A bare §return§ is allowed only with named results. Use named results when they make
the signature clearer, but do not use them just to save typing — an explicit §return
min, max§ is easier to read.

Named results have one genuinely useful power we will meet again in the errors topic: a
§defer§-ed function can change them after the body returns.

## Variadic functions: any number of arguments

Put §...§ before the type of the last parameter and the function accepts zero or more
of that type:

§§§go
func Sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func main() {
	fmt.Println(Sum())            // 0
	fmt.Println(Sum(1, 2, 3))     // 6
	xs := []int{10, 20}
	fmt.Println(Sum(xs...))       // 30  <- spread an existing slice with "..."
}
§§§

Output:

§§§text
0
6
30
§§§

Inside the function, §nums§ is an ordinary §[]int§. Three rules:

* Only the **last** parameter may be variadic.
* Calling it with zero arguments is legal; §nums§ is then an empty slice.
* To pass an existing slice, spread it: §Sum(xs...)§. Writing §Sum(xs)§ is a compile
  error — a §[]int§ is not an §int§.

§fmt.Println§ itself is variadic: §func Println(a ...any)§. That is why it takes
anything you throw at it.

## Recursion

A function may call itself. Go has no special syntax for it:

§§§go
func Factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * Factorial(n-1)
}
§§§

Each call gets its own copies of the parameters, so §n§ in the outer call is unaffected
by the inner one. There is no tail-call optimisation: deep recursion (hundreds of
thousands of frames) will eventually exhaust the goroutine stack.

## A complete example

§§§go
package main

import (
	"errors"
	"fmt"
)

// DivideAndRemainder returns both parts of integer division.
func DivideAndRemainder(a, b int) (quotient, remainder int) {
	quotient = a / b
	remainder = a % b
	return
}

// Average returns the mean of any number of numbers.
func Average(nums ...float64) (float64, error) {
	if len(nums) == 0 {
		return 0, errors.New("average of no numbers")
	}
	total := 0.0
	for _, n := range nums {
		total += n
	}
	return total / float64(len(nums)), nil
}

func main() {
	q, r := DivideAndRemainder(17, 5)
	fmt.Println(q, r)

	avg, err := Average(2, 4, 6)
	fmt.Println(avg, err)

	_, err = Average()
	fmt.Println(err)
}
§§§

Output:

§§§text
3 2
4 <nil>
average of no numbers
§§§

The §_§ in the last call discards the value we do not need. That is the idiomatic way
to say "I know there is a second result and I am
ignoring it on purpose".
`,
		Quiz: []Question{
			{
				ID:     "m1t7q1",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

func f() (a int, b int) {
	a = 1
	b = 2
	return
}

func main() {
	x, y := f()
	fmt.Println(x, y)
}`,
				Options: []Option{
					{Text: "§1 2§", Correct: true, Rationale: "The bare §return§ returns the current values of the named results §a§ and §b§."},
					{Text: "§0 0§", Correct: false, Rationale: "The named results were assigned before the return."},
					{Text: "It does not compile: §return§ needs values.", Correct: false, Rationale: "A bare §return§ is legal when every result is named."},
					{Text: "It does not compile: a function cannot name its results.", Correct: false, Rationale: "Named results are a standard Go feature."},
				},
				Explanation: "With named results, §a§ and §b§ are real variables. A bare §return§ sends back whatever they hold.",
			},
			{
				ID:     "m1t7q2",
				Prompt: "Which calls to §func Sum(nums ...int) int§ compile? Select all that apply.",
				Multi:  true,
				Code: `xs := []int{1, 2, 3}
Sum(1, 2, 3)
Sum(xs)
Sum(xs...)
Sum()
Sum("a")`,
				Options: []Option{
					{Text: "§Sum(1, 2, 3)§", Correct: true, Rationale: "Individual arguments."},
					{Text: "§Sum(xs)§", Correct: false, Rationale: "A §[]int§ is not an §int§; the compiler rejects it."},
					{Text: "§Sum(xs...)§", Correct: true, Rationale: "Spreading a slice into a variadic parameter is exactly what §...§ is for."},
					{Text: "§Sum()§", Correct: true, Rationale: "Zero arguments are allowed; §nums§ is then an empty slice."},
					{Text: "§Sum(\"a\")§", Correct: false, Rationale: "A string is not an §int§."},
				},
				Explanation: "Variadic parameters accept zero or more values, and an existing slice must be spread with §...§.",
			},
			{
				ID:     "m1t7q3",
				Prompt: "Why does §Divide§ return §(float64, error)§ instead of panicking on division by zero?",
				Code: `func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("cannot divide by zero")
	}
	return a / b, nil
}`,
				Options: []Option{
					{Text: "Because dividing by zero is a foreseeable situation the caller can handle, and Go prefers returning errors over crashing for those.", Correct: true, Rationale: "Go reserves panics for programmer mistakes, not for expected conditions."},
					{Text: "Because Go cannot panic on floating-point division.", Correct: false, Rationale: "It can panic; that is simply the wrong tool here."},
					{Text: "Because §error§ is required for functions with two results.", Correct: false, Rationale: "Two results can be any types, e.g. §(int, int)§."},
					{Text: "Because §float64§ cannot be returned directly.", Correct: false, Rationale: "It can."},
				},
				Explanation: "\"Value plus error\" is Go's way of pushing failures to the caller, who is the only one who knows what to do about them.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t7e1",
				Title: "Warm-up: quotient and remainder together",
				Prompt: `One function, two results.

§§§go
func DivMod(a, b int) (int, int)
§§§

Returns the quotient §a / b§ and the remainder §a % b§.

* §DivMod(17, 5)§ → §3, 2§
* §DivMod(10, 2)§ → §5, 0§
* §DivMod(3, 7)§ → §0, 3§

Name the results in the signature and use a bare §return§ — it is the tidiest way to
write this particular function.
`,
				Starter: `package main

import "fmt"

// DivMod returns the quotient and remainder of a / b.
func DivMod(a, b int) (quotient, remainder int) {
	// TODO
	return
}

func main() {
	fmt.Println(DivMod(17, 5))
	fmt.Println(DivMod(3, 7))
}
`,
				Tests: []File{{
					Name: "divmod_test.go",
					Content: `package main

import "testing"

func TestDivMod(t *testing.T) {
	cases := []struct{ a, b, q, r int }{
		{17, 5, 3, 2},
		{10, 2, 5, 0},
		{3, 7, 0, 3},
		{0, 4, 0, 0},
		{100, 10, 10, 0},
	}
	for _, c := range cases {
		q, r := DivMod(c.a, c.b)
		if q != c.q || r != c.r {
			t.Errorf("DivMod(%d, %d) = %d, %d; want %d, %d", c.a, c.b, q, r, c.q, c.r)
		}
	}
}
`,
				}},
				Hints: []string{
					"§quotient = a / b§ and §remainder = a % b§.",
					"Because the results are named, §return§ on its own sends them back.",
					"§%§ is the remainder operator: §17 % 5§ is §2§.",
				},
				Solution: `package main

import "fmt"

func DivMod(a, b int) (quotient, remainder int) {
	quotient = a / b
	remainder = a % b
	return
}

func main() {
	fmt.Println(DivMod(17, 5))
	fmt.Println(DivMod(3, 7))
}
`,
			},
			{
				ID:    "m1t7e2",
				Title: "Min and max from any number of arguments",
				Prompt: `Use a variadic parameter to accept any number of values.

§§§go
func MinMax(nums ...int) (min, max int, ok bool)
§§§

* With values, return the smallest and largest, and §ok == true§.
* With **no** values, return §0, 0, false§ — this is the "there was nothing to look at"
  signal, the same idea as the §ok§ from a map lookup.
* A single value is both the min and the max.

§MinMax(3, 1, 4, 1, 5)§ → §1, 5, true§
`,
				Starter: `package main

import "fmt"

// MinMax returns the smallest and largest of nums, and whether nums was non-empty.
func MinMax(nums ...int) (min, max int, ok bool) {
	// TODO
	return 0, 0, false
}

func main() {
	fmt.Println(MinMax(3, 1, 4, 1, 5))
	fmt.Println(MinMax())
}
`,
				Tests: []File{{
					Name: "minmax_test.go",
					Content: `package main

import "testing"

func TestMinMax(t *testing.T) {
	cases := []struct {
		nums     []int
		min, max int
		ok       bool
	}{
		{[]int{3, 1, 4, 1, 5}, 1, 5, true},
		{[]int{7}, 7, 7, true},
		{[]int{-3, -9, -1}, -9, -1, true},
		{[]int{0, 0}, 0, 0, true},
		{nil, 0, 0, false},
	}
	for _, c := range cases {
		min, max, ok := MinMax(c.nums...)
		if min != c.min || max != c.max || ok != c.ok {
			t.Errorf("MinMax(%v) = %d, %d, %v; want %d, %d, %v", c.nums, min, max, ok, c.min, c.max, c.ok)
		}
	}
}
`,
				}},
				Hints: []string{
					"Guard the empty case first: §if len(nums) == 0 { return 0, 0, false }§.",
					"Seed both results from the first element: §min, max = nums[0], nums[0]§.",
					"Loop over §nums[1:]§ — the first element is already in place.",
					"Return §min, max, true§ at the end.",
				},
				Solution: `package main

import "fmt"

func MinMax(nums ...int) (min, max int, ok bool) {
	if len(nums) == 0 {
		return 0, 0, false
	}
	min, max = nums[0], nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return min, max, true
}

func main() {
	fmt.Println(MinMax(3, 1, 4, 1, 5))
	fmt.Println(MinMax())
}
`,
			},
			{
				ID:    "m1t7e3",
				Title: "A safe average that reports failure",
				Prompt: `Implement the most common Go signature: a value **and** an error.

§§§go
func Average(nums ...float64) (float64, error)
§§§

* With one or more numbers, return their arithmetic mean and §nil§ for the error.
* With **no** numbers, return §0§ and an error created with
  §errors.New("average of no numbers")§.
* Remember that §float64(len(nums))§ is required — §int / int§ would truncate.

Import §"errors"§ for this one. You will use §errors.New§ constantly from now on.
`,
				Starter: `package main

import (
	"errors"
	"fmt"
)

// Average returns the arithmetic mean of nums, or an error if nums is empty.
func Average(nums ...float64) (float64, error) {
	// TODO
	return 0, nil
}

func main() {
	avg, err := Average(2, 4, 6)
	fmt.Println(avg, err)

	_, err = Average()
	fmt.Println(err)

	// errors.New is imported for you.
	_ = errors.New
}
`,
				Tests: []File{{
					Name: "average_test.go",
					Content: `package main

import (
	"math"
	"testing"
)

func TestAverage(t *testing.T) {
	cases := []struct {
		nums []float64
		want float64
	}{
		{[]float64{2, 4, 6}, 4},
		{[]float64{1.5}, 1.5},
		{[]float64{0, 0}, 0},
		{[]float64{-2, 2}, 0},
		{[]float64{1, 2}, 1.5},
	}
	for _, c := range cases {
		got, err := Average(c.nums...)
		if err != nil {
			t.Errorf("Average(%v) returned an unexpected error: %v", c.nums, err)
			continue
		}
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("Average(%v) = %v, want %v", c.nums, got, c.want)
		}
	}
}

func TestAverageEmpty(t *testing.T) {
	got, err := Average()
	if err == nil {
		t.Fatal("Average() should return an error for empty input")
	}
	if got != 0 {
		t.Errorf("Average() value = %v, want 0 alongside the error", got)
	}
}
`,
				}},
				Hints: []string{
					"Start with §if len(nums) == 0 { return 0, errors.New(\"average of no numbers\") }§.",
					"Sum in a loop: §total += n§.",
					"§return total / float64(len(nums)), nil§",
				},
				Solution: `package main

import (
	"errors"
	"fmt"
)

func Average(nums ...float64) (float64, error) {
	if len(nums) == 0 {
		return 0, errors.New("average of no numbers")
	}
	total := 0.0
	for _, n := range nums {
		total += n
	}
	return total / float64(len(nums)), nil
}

func main() {
	avg, err := Average(2, 4, 6)
	fmt.Println(avg, err)

	_, err = Average()
	fmt.Println(err)
}
`,
			},
		},
	}
}

func m1t8() Topic {
	return Topic{
		ID:      "m1t8",
		Title:   "Function Values, Closures and defer",
		Summary: "First-class functions, closures and deferred cleanup.",
		Theory: `
In Go a function is a value — like an §int§ or a §string§. It can be stored in a
variable, passed to another function, returned from one, and compared to §nil§.

## Functions as values

§§§go
add := func(a, b int) int { return a + b }
fmt.Println(add(2, 3)) // 5
§§§

That §func(a, b int) int { … }§ with no name is a **function literal** (other languages
call it a lambda or anonymous function). You can also point a variable at a named
function:

§§§go
func double(n int) int { return n * 2 }

f := double
fmt.Println(f(21)) // 42
§§§

Functions are also perfectly normal things to pass as arguments. This is how sorting,
HTTP handlers and callbacks all work:

§§§go
func Apply(xs []int, f func(int) int) []int {
	out := make([]int, len(xs))
	for i, v := range xs {
		out[i] = f(v)
	}
	return out
}

func main() {
	doubled := Apply([]int{1, 2, 3}, func(n int) int { return n * 2 })
	fmt.Println(doubled) // [2 4 6]
}
§§§

Read the parameter type §f func(int) int§ as *"f is a function taking an int and
returning an int"*. It is just a type, written out.

## Closures

A function literal can use variables from the surrounding function. When it does, it
**keeps them alive** even after the outer function has returned. That combination is a
**closure**.

§§§go
func Counter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

func main() {
	next := Counter()
	fmt.Println(next()) // 1
	fmt.Println(next()) // 2
	fmt.Println(next()) // 3

	other := Counter()
	fmt.Println(other()) // 1  <- a different, independent count
}
§§§

Output:

§§§text
1
2
3
1
§§§

§count§ is an ordinary local variable of §Counter§. Because the returned closure refers
to it, the compiler moves it to the heap, and it lives as long as the closure does. Each call to §Counter()§ creates a *fresh* §count§.

This is more useful than it looks. A closure is the standard way to give a function
private state without inventing a struct:

§§§go
func Memoize(f func(int) int) func(int) int {
	cache := map[int]int{}
	return func(n int) int {
		if v, ok := cache[n]; ok {
			return v
		}
		v := f(n)
		cache[n] = v
		return v
	}
}
§§§

§Memoize§ wraps a slow function and remembers its answers. The §cache§ belongs to the
closure — nobody else can see or corrupt it.

## defer: "run this when the function finishes"

§defer§ schedules a call to happen when the surrounding function returns — however it
returns, including on an early §return§.

§§§go
func read(name string) {
	fmt.Println("opening", name)
	defer fmt.Println("closing", name)

	fmt.Println("reading", name)
}

func main() {
	read("file.txt")
	fmt.Println("done")
}
§§§

Output:

§§§text
opening file.txt
reading file.txt
closing file.txt
done
§§§

The deferred line runs after the body of §read§ but before §main§ continues.
It is the perfect place for cleanup: closing files, unlocking mutexes, ending
transactions. You write the cleanup immediately after the thing that needs it, and you
cannot forget it on any of the five early-§return§ paths.

### Three rules of defer

**1. Defers run last-in-first-out.** The last one scheduled runs first.

§§§go
for i := 1; i <= 3; i++ {
	defer fmt.Print(i, " ")
}
fmt.Print("body ")
§§§

Output:

§§§text
body 3 2 1
§§§

**2. Arguments are evaluated immediately**, when you write the §defer§. Only the *call*
is postponed.

§§§go
for i := 1; i <= 3; i++ {
	defer fmt.Print(i, " ") // i is copied HERE, not at the end
}
§§§

That is why the loop above prints §3 2 1§ and not §4 4 4§: each §defer§ captured its
own value of §i§.

**3. Defers run even if the function panics.** That is what makes them reliable for
cleanup. (We meet §panic§ properly in the errors topic.)

### One trap

§defer§ inside a loop defers until the **function** returns, not the iteration. This is
a bug:

§§§go
for _, name := range names {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close() // all files stay open until the whole function ends!
}
§§§

With a long list you run out of file descriptors. The fix is to put the loop body in
its own function so the §defer§ fires every iteration.

## A complete example

§§§go
package main

import "fmt"

func main() {
	add := func(a, b int) int { return a + b }
	mul := func(a, b int) int { return a * b }

	// A closure that remembers which operation it was given.
	makeOp := func(name string, op func(int, int) int) func(int, int) {
		return func(a, b int) int {
			fmt.Printf("applying %s(%d, %d)\n", name, a, b)
			return op(a, b)
		}
	}

	doAdd := makeOp("add", add)
	doMul := makeOp("mul", mul)

	fmt.Println(doAdd(2, 3))
	fmt.Println(doMul(2, 3))
}
§§§

Output:

§§§text
applying add(2, 3)
5
applying mul(2, 3)
6
§§§

§makeOp§ is a *function that builds functions*. Once you are comfortable with this
shape, a lot of the standard library stops looking magical.
`,
		Quiz: []Question{
			{
				ID:     "m1t8q1",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

func makeCounter() func() int {
	n := 100
	return func() int {
		n++
		return n
	}
}

func main() {
	a := makeCounter()
	b := makeCounter()
	fmt.Println(a(), a(), b())
}`,
				Options: []Option{
					{Text: "§101 102 101§", Correct: true, Rationale: "Each call to §makeCounter§ creates a fresh §n§; §a§ and §b§ have independent counters."},
					{Text: "§101 102 103§", Correct: false, Rationale: "§b§ has its own §n§, starting at 100 again."},
					{Text: "§101 101 101§", Correct: false, Rationale: "Each closure remembers its own §n§ and increments it."},
					{Text: "§100 100 100§", Correct: false, Rationale: "The closure increments before returning."},
				},
				Explanation: "A closure captures the *variable*. Two calls to the enclosing function create two different variables.",
			},
			{
				ID:     "m1t8q2",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

func run() {
	defer fmt.Print("deferred ")
	fmt.Print("body ")
}

func main() {
	run()
	fmt.Print("main")
}`,
				Options: []Option{
					{Text: "§body deferred main§", Correct: true, Rationale: "The deferred call runs when §run§ returns — before §main§ continues."},
					{Text: "§body main deferred§", Correct: false, Rationale: "§defer§ is scoped to §run§, not to the whole program."},
					{Text: "§deferred body main§", Correct: false, Rationale: "The body of §run§ executes first; the defer waits for the return."},
					{Text: "§body main§ — the defer never runs.", Correct: false, Rationale: "§defer§ always runs when the function returns."},
				},
				Explanation: "§defer§ schedules a call for the end of the *enclosing function*, not the end of the program.",
			},
			{
				ID:     "m1t8q3",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

func main() {
	for i := 1; i <= 3; i++ {
		defer fmt.Print(i, " ")
	}
	fmt.Print("body ")
}`,
				Options: []Option{
					{Text: "§body 3 2 1 §", Correct: true, Rationale: "Defers run LIFO, and each captured its own value of §i§ when the §defer§ statement executed."},
					{Text: "§body 1 2 3 §", Correct: false, Rationale: "Deferred calls run in reverse order of scheduling."},
					{Text: "§body 4 4 4 §", Correct: false, Rationale: "The arguments are evaluated at the moment of the §defer§, so each §i§ is different. (And §i§ is 4 only after the loop.)"},
					{Text: "§1 2 3 body §", Correct: false, Rationale: "The loop body is not the deferred part."},
				},
				Explanation: "Two rules at once: LIFO ordering, and arguments captured at the §defer§ statement rather than at the call.",
			},
			{
				ID:     "m1t8q4",
				Prompt: "Why is opening files inside a loop with §defer f.Close()§ a bug?",
				Options: []Option{
					{Text: "Because the deferred closes all pile up until the surrounding function returns, so many files stay open at once.", Correct: true, Rationale: "You can exhaust the process's file-descriptor limit before the function ends."},
					{Text: "Because §defer§ inside a loop is a compile error.", Correct: false, Rationale: "It compiles — that is what makes it a trap."},
					{Text: "Because §Close§ cannot be deferred.", Correct: false, Rationale: "Deferring §Close§ is the standard idiom."},
					{Text: "Because the loop variable §f§ is reused.", Correct: false, Rationale: "Each iteration declares a new §f§."},
				},
				Explanation: "Extract the loop body into a function (or a closure that you call immediately) so the §defer§ runs every iteration.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t8e1",
				Title: "Warm-up: apply a function twice",
				Prompt: `Pass a function as a value.

§§§go
func ApplyTwice(x int, f func(int) int) int
§§§

Returns §f(f(x))§ — the function applied twice.

* §ApplyTwice(3, func(n int) int { return n * 2 })§ → §12§
* §ApplyTwice(3, func(n int) int { return n + 10 })§ → §23§

Notice that §f§ is a *parameter whose type is a function*. That is all "higher-order
function" means.
`,
				Starter: `package main

import "fmt"

// ApplyTwice returns f(f(x)).
func ApplyTwice(x int, f func(int) int) int {
	// TODO
	return 0
}

func main() {
	double := func(n int) int { return n * 2 }
	fmt.Println(ApplyTwice(3, double))
	fmt.Println(ApplyTwice(3, func(n int) int { return n + 10 }))
}
`,
				Tests: []File{{
					Name: "applytwice_test.go",
					Content: `package main

import "testing"

func TestApplyTwice(t *testing.T) {
	double := func(n int) int { return n * 2 }
	inc := func(n int) int { return n + 1 }

	cases := []struct {
		x    int
		f    func(int) int
		want int
	}{
		{3, double, 12},
		{0, double, 0},
		{-2, double, -8},
		{3, inc, 5},
		{10, func(n int) int { return n / 2 }, 2},
	}
	for _, c := range cases {
		if got := ApplyTwice(c.x, c.f); got != c.want {
			t.Errorf("ApplyTwice(%d, f) = %d, want %d", c.x, got, c.want)
		}
	}
}
`,
				}},
				Hints: []string{
					"Call §f§ once on §x§, then again on the result.",
					"§return f(f(x))§ — one line does it.",
					"The order matters for operations that are not commutative, like §n / 2§.",
				},
				Solution: `package main

import "fmt"

func ApplyTwice(x int, f func(int) int) int {
	return f(f(x))
}

func main() {
	double := func(n int) int { return n * 2 }
	fmt.Println(ApplyTwice(3, double))
	fmt.Println(ApplyTwice(3, func(n int) int { return n + 10 }))
}
`,
			},
			{
				ID:    "m1t8e2",
				Title: "A counter that keeps its own state",
				Prompt: `Implement a closure.

§§§go
func NewCounter(start int) func() int
§§§

* Returns a function that, each time it is called, adds 1 to a private counter and
  returns the new value.
* §start§ is the value *before* the first increment: §NewCounter(10)§ then calling the
  returned function gives §11§, §12§, §13§…
* Two counters made from two calls must not affect each other.

This demonstrates that a closure keeps a variable alive after the
enclosing function has returned.
`,
				Starter: `package main

import "fmt"

// NewCounter returns a function that counts upwards, starting after start.
func NewCounter(start int) func() int {
	// TODO
	return func() int { return 0 }
}

func main() {
	next := NewCounter(10)
	fmt.Println(next(), next(), next())
}
`,
				Tests: []File{{
					Name: "counter_test.go",
					Content: `package main

import "testing"

func TestCounter(t *testing.T) {
	next := NewCounter(10)
	for i, want := range []int{11, 12, 13} {
		if got := next(); got != want {
			t.Fatalf("call %d returned %d, want %d", i+1, got, want)
		}
	}
}

func TestCountersAreIndependent(t *testing.T) {
	a := NewCounter(0)
	b := NewCounter(100)

	if got := a(); got != 1 {
		t.Errorf("a() = %d, want 1", got)
	}
	if got := a(); got != 2 {
		t.Errorf("a() = %d, want 2", got)
	}
	if got := b(); got != 101 {
		t.Errorf("b() = %d, want 101 (independent of a)", got)
	}
	if got := a(); got != 3 {
		t.Errorf("a() = %d, want 3 (independent of b)", got)
	}
}

func TestCounterFromZero(t *testing.T) {
	next := NewCounter(0)
	if got := next(); got != 1 {
		t.Errorf("first call = %d, want 1", got)
	}
}
`,
				}},
				Hints: []string{
					"Declare §count := start§ *outside* the returned function but *inside* §NewCounter§.",
					"The returned closure increments §count§ and returns it.",
					"Returning §count++§ is not valid Go — incrementation is a statement, not an expression. Write §count++§ on its own line, then §return count§.",
				},
				Solution: `package main

import "fmt"

func NewCounter(start int) func() int {
	count := start
	return func() int {
		count++
		return count
	}
}

func main() {
	next := NewCounter(10)
	fmt.Println(next(), next(), next())
}
`,
			},
			{
				ID:    "m1t8e3",
				Title: "Memoize, and prove defer runs last",
				Prompt: `Two tasks in one file.

**Part 1 — a memoising closure.**

§§§go
func Memoize(f func(int) int) func(int) int
§§§

Returns a function that calls §f§ at most once per distinct argument, remembering
answers in a map that only it can see.

**Part 2 — defer ordering.**

§§§go
func DeferOrder() []string
§§§

Returns a slice built by three §defer§ statements that append the strings §"first"§,
§"second"§ and §"third"§, scheduled in that order. Because defers run
last-in-first-out, the test expects §["third", "second", "first"]§ — a concrete reminder
of the ordering rule.

Two tools from later topics, in brief:

* §append§ takes a slice and a value and returns the (possibly longer) slice, so you
  must assign the result back: §out = append(out, "first")§. Starting from
  §var out []string§ is fine — §append§ grows a nil slice.
* §defer§ must be followed by a *call*, not a statement. To defer an assignment like the
  one above, wrap it in an anonymous function and call it immediately:

  §§§go
  defer func() { out = append(out, "first") }()
  §§§
`,
				Starter: `package main

import "fmt"

// Memoize wraps f and caches its results.
func Memoize(f func(int) int) func(int) int {
	// TODO
	return f
}

// DeferOrder schedules three deferred appends and returns their final order.
//
// The result is a NAMED result, which is what lets the deferred closures
// change the value this function returns. (With an unnamed result, §return out§
// would copy the slice before the defers ran, and you would get an empty slice.)
func DeferOrder() (out []string) {
	// TODO: three deferred appends
	return
}

func main() {
	calls := 0
	square := Memoize(func(n int) int {
		calls++
		return n * n
	})
	fmt.Println(square(4), square(4), square(5), calls)
	fmt.Println(DeferOrder())
}
`,
				Tests: []File{{
					Name: "memoize_test.go",
					Content: `package main

import (
	"reflect"
	"testing"
)

func TestMemoizeCaches(t *testing.T) {
	calls := 0
	slow := func(n int) int {
		calls++
		return n * n
	}
	square := Memoize(slow)

	if got := square(4); got != 16 {
		t.Fatalf("square(4) = %d, want 16", got)
	}
	if got := square(4); got != 16 {
		t.Fatalf("square(4) again = %d, want 16", got)
	}
	if calls != 1 {
		t.Fatalf("f was called %d times, want 1", calls)
	}

	square(5)
	square(5)
	square(4)
	if calls != 2 {
		t.Fatalf("f was called %d times, want 2", calls)
	}
}

func TestMemoizeIsIndependentPerWrapper(t *testing.T) {
	a := Memoize(func(n int) int { return n + 1 })
	b := Memoize(func(n int) int { return n + 2 })
	if a(1) != 2 || b(1) != 3 {
		t.Fatalf("wrappers share state: %d %d", a(1), b(1))
	}
}

func TestDeferOrder(t *testing.T) {
	got := DeferOrder()
	want := []string{"third", "second", "first"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DeferOrder() = %v, want %v", got, want)
	}
}
`,
				}},
				Hints: []string{
					"§cache := map[int]int{}§ inside §Memoize§, then the usual §if v, ok := cache[n]; ok { return v }§.",
					"On a miss: §v := f(n); cache[n] = v; return v§.",
					"Write the three deferred closures from the prompt, in the order §\"first\"§, §\"second\"§, §\"third\"§.",
					"Notice the signature is already §(out []string)§ — a **named** result. §return§ on its own sends back whatever §out§ holds *after* the defers have run.",
				},
				Solution: `package main

import "fmt"

func Memoize(f func(int) int) func(int) int {
	cache := map[int]int{}
	return func(n int) int {
		if v, ok := cache[n]; ok {
			return v
		}
		v := f(n)
		cache[n] = v
		return v
	}
}

func DeferOrder() (out []string) {
	defer func() { out = append(out, "first") }()
	defer func() { out = append(out, "second") }()
	defer func() { out = append(out, "third") }()
	return
}

func main() {
	calls := 0
	square := Memoize(func(n int) int {
		calls++
		return n * n
	})
	fmt.Println(square(4), square(4), square(5), calls)
	fmt.Println(DeferOrder())
}
`,
			},
		},
	}
}

func m1t9() Topic {
	return Topic{
		ID:      "m1t9",
		Title:   "Pointers",
		Summary: "Addresses, why Go passes copies, and how to change the caller's value.",
		Theory: `
A pointer is a value that holds the address of another value. Its main use in Go is
letting a function modify the caller's variable.

## Why pointers exist

Go passes **everything by value**. When you hand a function a variable, the function
receives a copy:

§§§go
func broken(n int) {
	n = n + 1 // changes the copy only
}

func main() {
	x := 5
	broken(x)
	fmt.Println(x) // still 5
}
§§§

That is deliberate — copies are easy to reason about and cannot be changed behind your
back. But sometimes you *want* to change the original. That is what a pointer is for.

## The two operators

* §&x§ means "the address of §x§". Its type is §*int§ when §x§ is an §int§. Read §*int§
  as *"pointer to int"*.
* §*p§ means "the value at address §p§". This is called **dereferencing**.

§§§go
package main

import "fmt"

func main() {
	x := 42
	p := &x         // p is *int
	fmt.Println(p)  // 0xc000018030  (an address — yours will differ)
	fmt.Println(*p) // 42

	*p = 99         // write through the pointer
	fmt.Println(x)  // 99
}
§§§

The address printed by the first line will not match mine. It is a real memory
location, and it changes between runs. Never depend on its value.

## Fixing the broken function

§§§go
func addOne(n *int) { // takes a pointer
	*n = *n + 1 // change the value being pointed at
}

func main() {
	x := 5
	addOne(&x) // pass the address
	fmt.Println(x) // 6
}
§§§

Compare the two versions:

| | value parameter | pointer parameter |
| --- | --- | --- |
| Declared as | §n int§ | §n *int§ |
| Call with | §f(x)§ | §f(&x)§ |
| Changing §n§ affects the caller? | no | yes (via §*n§) |

## Pointers to structs are friendlier than they look

Go automatically inserts the dereference when you use §.§ on a pointer to a struct. You
never write §(*p).Field§:

§§§go
type Point struct{ X, Y int }

func move(p *Point, dx int) {
	p.X += dx // same as (*p).X += dx
}

func main() {
	pt := Point{X: 1}
	move(&pt, 10)
	fmt.Println(pt.X) // 11
}
§§§

There is also a shortcut for creating a pointer to a brand new struct: §&Point{}§ is
allowed even though §&42§ is not (composite literals are *addressable*; constants are
not).

## nil: the pointer that points nowhere

The zero value of a pointer is §nil§. Dereferencing a nil pointer panics, so check:

§§§go
var p *int
fmt.Println(p)        // <nil>
fmt.Println(p == nil) // true
// fmt.Println(*p)    // panic: runtime error: invalid memory address or nil pointer dereference

if p != nil {
	fmt.Println(*p)
}
§§§

The same §if x != nil§ guard you will write for file handles, error values and optional
config fields is exactly this check.

## Returning a pointer to a local variable is fine

In C this is a classic bug. In Go it is normal, because the garbage collector keeps the
value alive:

§§§go
func newCounter() *int {
	n := 0
	return &n // completely safe in Go
}
§§§

The compiler's **escape analysis** decides whether the value lives on the stack or the
heap. You do not have to manage this; the compiler does.

## What Go does NOT have

* **No pointer arithmetic.** §p + 1§ is a compile error. You cannot walk off the end of
  an array by accident.
* **No §->§ operator.** §p.Field§, not §p->Field§.
* **No §null§.** There is exactly one nil, shared by pointers, slices, maps, channels,
  functions and interfaces — and it is a *value*, not a keyword.

## When to use a pointer

Use one when you need to:

1. change the caller's value (the §addOne§ example),
2. avoid copying something large,
3. say "this value may be absent" (§nil§ means not there — very common for optional JSON
   fields),
4. share one instance among many owners (a mutex, a database pool, a logger).

Otherwise pass the value. Immutable copies are easier to read, and often faster.

## A complete example

§§§go
package main

import "fmt"

type Account struct {
	Owner   string
	Balance int
}

func Deposit(a *Account, amount int) {
	a.Balance += amount
}

func TryWithdraw(a *Account, amount int) bool {
	if a.Balance < amount {
		return false
	}
	a.Balance -= amount
	return true
}

func main() {
	acc := &Account{Owner: "Ada", Balance: 100}

	Deposit(acc, 50)
	fmt.Println(acc.Balance) // 150

	fmt.Println(TryWithdraw(acc, 500)) // false
	fmt.Println(TryWithdraw(acc, 20))  // true
	fmt.Println(acc.Balance)           // 130
}
§§§

Output:

§§§text
150
false
true
130
§§§

Everything here mutates one shared §Account§ through a pointer. This is the shape of
almost every real Go program that stores state.
`,
		Quiz: []Question{
			{
				ID:     "m1t9q1",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

func change(x int)   { x = 100 }
func changeP(x *int) { *x = 100 }

func main() {
	a := 1
	change(a)
	b := 1
	changeP(&b)
	fmt.Println(a, b)
}`,
				Options: []Option{
					{Text: "§1 100§", Correct: true, Rationale: "§change§ receives a copy and changes nothing the caller can see; §changeP§ writes through the pointer."},
					{Text: "§100 100§", Correct: false, Rationale: "Plain value parameters are always copies."},
					{Text: "§1 1§", Correct: false, Rationale: "§*x = 100§ really does change §b§."},
					{Text: "It does not compile: cannot pass an int where *int is needed.", Correct: false, Rationale: "§changeP(&b)§ passes the address, which *is* a §*int§."},
				},
				Explanation: "Value parameters are copies. Pointers are how you reach back into the caller's variable.",
			},
			{
				ID:     "m1t9q2",
				Prompt: "Which lines are legal Go? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "§p := &Point{1, 2}§", Correct: true, Rationale: "A composite literal is addressable, so you may take its address."},
					{Text: "§p := &42§", Correct: false, Rationale: "A constant has no address. The compiler says: §cannot take the address of 42§."},
					{Text: "§func f() *int { x := 1; return &x }§", Correct: true, Rationale: "Escape analysis moves §x§ to the heap, so the pointer stays valid."},
					{Text: "§p := new(int); *p = 7§", Correct: true, Rationale: "§new(int)§ returns a §*int§ pointing at a zeroed int."},
					{Text: "§p := &p§", Correct: false, Rationale: "Type mismatch: §&p§ is a §**int§ inside the declaration of the §*int§ §p§ itself."},
				},
				Explanation: "Only addressable operands can have their address taken, and returning a pointer to a local is safe in Go.",
			},
			{
				ID:     "m1t9q3",
				Prompt: "What happens when this runs?",
				Code: `package main

import "fmt"

type Box struct{ V int }

func (b *Box) Set(v int) { b.V = v }

func main() {
	var b *Box
	b.Set(5)
	fmt.Println(b.V)
}`,
				Options: []Option{
					{Text: "It panics: nil pointer dereference inside §Set§.", Correct: true, Rationale: "Calling a pointer-receiver method on nil is allowed, but §b.V = v§ dereferences §b§."},
					{Text: "It prints §5§.", Correct: false, Rationale: "There is no Box to store the 5 in."},
					{Text: "It does not compile.", Correct: false, Rationale: "It compiles; the failure is at run time."},
					{Text: "It prints §0§.", Correct: false, Rationale: "The panic happens before any print."},
				},
				Explanation: "A nil pointer is only a problem when you *dereference* it. Methods that guard with §if b == nil { return }§ are safe to call.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t9e1",
				Title: "Warm-up: increment and swap through pointers",
				Prompt: `§§§go
func Increment(p *int)
func Swap(a, b *int)
func ResetIfNegative(p *float64)
§§§

* §Increment§ adds 1 to the value the pointer points at.
* §Swap§ exchanges the two pointed-at values.
* §ResetIfNegative§ sets the value to §0§ when it is negative, and leaves it alone
  otherwise.

All three must tolerate a §nil§ pointer by doing nothing — the tests pass §nil§ on
purpose, and a panic there fails the exercise.
`,
				Starter: `package main

import "fmt"

func Increment(p *int) {
	// TODO
}

func Swap(a, b *int) {
	// TODO
}

func ResetIfNegative(p *float64) {
	// TODO
}

func main() {
	x := 1
	Increment(&x)
	y := 2
	Swap(&x, &y)
	f := -3.5
	ResetIfNegative(&f)
	fmt.Println(x, y, f)
}
`,
				Tests: []File{{
					Name: "pointers_test.go",
					Content: `package main

import "testing"

func TestIncrement(t *testing.T) {
	x := 41
	Increment(&x)
	if x != 42 {
		t.Fatalf("Increment: got %d, want 42", x)
	}
	Increment(nil) // must not panic
}

func TestSwap(t *testing.T) {
	a, b := 1, 2
	Swap(&a, &b)
	if a != 2 || b != 1 {
		t.Fatalf("Swap: got a=%d b=%d, want a=2 b=1", a, b)
	}
	Swap(nil, &a)
	Swap(&a, nil)
}

func TestResetIfNegative(t *testing.T) {
	f := -3.5
	ResetIfNegative(&f)
	if f != 0 {
		t.Fatalf("ResetIfNegative: got %v, want 0", f)
	}
	g := 2.5
	ResetIfNegative(&g)
	if g != 2.5 {
		t.Fatalf("ResetIfNegative changed a positive value to %v", g)
	}
	ResetIfNegative(nil)
}
`,
				}},
				Hints: []string{
					"Guard first: §if p == nil { return }§.",
					"§*p§ is the value; §*p = x§ writes it.",
					"Swap can use the multi-assignment trick: §*a, *b = *b, *a§.",
					"§if p != nil && *p < 0 { *p = 0 }§",
				},
				Solution: `package main

import "fmt"

func Increment(p *int) {
	if p == nil {
		return
	}
	*p++
}

func Swap(a, b *int) {
	if a == nil || b == nil {
		return
	}
	*a, *b = *b, *a
}

func ResetIfNegative(p *float64) {
	if p != nil && *p < 0 {
		*p = 0
	}
}

func main() {
	x := 1
	Increment(&x)
	y := 2
	Swap(&x, &y)
	f := -3.5
	ResetIfNegative(&f)
	fmt.Println(x, y, f)
}
`,
			},
			{
				ID:    "m1t9e2",
				Title: "Returning pointers to values",
				Prompt: `Implement two functions that return pointers.

§§§go
func NewInt(v int) *int
func MinMaxPtr(nums ...int) (min *int, max *int)
§§§

* §NewInt(v)§ returns a pointer to a freshly made int holding §v§.
* §MinMaxPtr§ returns pointers to the smallest and largest values, or §nil, nil§ when
  it is given nothing at all. That §nil§ is the "no answer" signal, just like the
  §false§ you returned in the previous topic.
* A single value gives two pointers to the same number.

Because Go keeps returned locals alive, both of these are safe. In C they would be
use-after-free bugs.
`,
				Starter: `package main

import "fmt"

// NewInt returns a pointer to a new int holding v.
func NewInt(v int) *int {
	// TODO
	return nil
}

// MinMaxPtr returns pointers to the smallest and largest of nums.
func MinMaxPtr(nums ...int) (*int, *int) {
	// TODO
	return nil, nil
}

func main() {
	p := NewInt(7)
	fmt.Println(*p)

	lo, hi := MinMaxPtr(3, 1, 4, 1, 5)
	fmt.Println(*lo, *hi)
	fmt.Println(MinMaxPtr())
}
`,
				Tests: []File{{
					Name: "newint_test.go",
					Content: `package main

import "testing"

func TestNewInt(t *testing.T) {
	p := NewInt(7)
	if p == nil {
		t.Fatal("NewInt returned nil")
	}
	if *p != 7 {
		t.Fatalf("NewInt(7) points at %d, want 7", *p)
	}
	// Each call must produce an independent value.
	q := NewInt(7)
	*p = 99
	if *q != 7 {
		t.Fatalf("two NewInt calls share memory: *q = %d, want 7", *q)
	}
}

func TestMinMaxPtr(t *testing.T) {
	lo, hi := MinMaxPtr(3, 1, 4, 1, 5)
	if lo == nil || hi == nil {
		t.Fatal("MinMaxPtr returned nil for non-empty input")
	}
	if *lo != 1 || *hi != 5 {
		t.Fatalf("MinMaxPtr = %d, %d; want 1, 5", *lo, *hi)
	}

	same, same2 := MinMaxPtr(42)
	if *same != 42 || *same2 != 42 {
		t.Fatalf("single value gave %d, %d", *same, *same2)
	}

	none, none2 := MinMaxPtr()
	if none != nil || none2 != nil {
		t.Fatalf("MinMaxPtr() = %v, %v; want nil, nil", none, none2)
	}
}

func TestMinMaxPtrNegatives(t *testing.T) {
	lo, hi := MinMaxPtr(-5, -1, -9)
	if *lo != -9 || *hi != -1 {
		t.Fatalf("got %d, %d; want -9, -1", *lo, *hi)
	}
}
`,
				}},
				Hints: []string{
					"§NewInt§ is two lines: §v := v§ then §return &v§ — or simply §return &v§ since it is a parameter.",
					"§MinMaxPtr§ returns early with §return nil, nil§ when there is nothing to look at.",
					"Seed from the first element, loop over the rest comparing, then §return &min, &max§.",
					"Careful: if you keep the result in loop variables you end up returning pointers to the *last* value. Store into named local variables declared before the loop.",
				},
				Solution: `package main

import "fmt"

func NewInt(v int) *int {
	return &v
}

func MinMaxPtr(nums ...int) (*int, *int) {
	if len(nums) == 0 {
		return nil, nil
	}
	min, max := nums[0], nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return &min, &max
}

func main() {
	p := NewInt(7)
	fmt.Println(*p)

	lo, hi := MinMaxPtr(3, 1, 4, 1, 5)
	fmt.Println(*lo, *hi)
	fmt.Println(MinMaxPtr())
}
`,
			},
			{
				ID:    "m1t9e3",
				Title: "An account that only changes through pointers",
				Prompt: `An account mutated in one place and read in another.

§§§go
type Account struct {
	Owner   string
	Balance int
}

func NewAccount(owner string, balance int) *Account
func Deposit(a *Account, amount int) int
func Withdraw(a *Account, amount int) (int, bool)
func Balance(a *Account) int
§§§

* §NewAccount§ returns a pointer to a ready account.
* §Deposit§ adds and returns the new balance.
* §Withdraw§ subtracts if there is enough money and returns the new balance and §true§;
  otherwise it changes nothing and returns the current balance and §false§.
* §Balance§ returns the current balance. Because the receiver is an ordinary parameter
  here, it takes a pointer.

Notice how the §*Account§ parameter makes the intent obvious: this function is allowed
to change the account.
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

func Deposit(a *Account, amount int) int {
	// TODO
	return 0
}

func Withdraw(a *Account, amount int) (int, bool) {
	// TODO
	return 0, false
}

func Balance(a *Account) int {
	// TODO
	return 0
}

func main() {
	acc := NewAccount("Ada", 100)
	Deposit(acc, 50)
	fmt.Println(Balance(acc))
	fmt.Println(Withdraw(acc, 500))
	fmt.Println(Withdraw(acc, 20))
	fmt.Println(Balance(acc))
}
`,
				Tests: []File{{
					Name: "account_test.go",
					Content: `package main

import "testing"

func TestNewAccount(t *testing.T) {
	acc := NewAccount("Ada", 100)
	if acc == nil {
		t.Fatal("NewAccount returned nil")
	}
	if acc.Owner != "Ada" || acc.Balance != 100 {
		t.Fatalf("NewAccount = %+v", acc)
	}
	if Balance(acc) != 100 {
		t.Fatalf("Balance = %d, want 100", Balance(acc))
	}
}

func TestDeposit(t *testing.T) {
	acc := NewAccount("Ada", 100)
	if got := Deposit(acc, 50); got != 150 {
		t.Fatalf("Deposit returned %d, want 150", got)
	}
	if Balance(acc) != 150 {
		t.Fatalf("the account itself was not changed: %d", Balance(acc))
	}
}

func TestWithdraw(t *testing.T) {
	acc := NewAccount("Ada", 100)

	if bal, ok := Withdraw(acc, 500); ok || bal != 100 {
		t.Fatalf("Withdraw too much = %d, %v; want 100, false", bal, ok)
	}
	if Balance(acc) != 100 {
		t.Fatalf("a failed withdrawal changed the balance: %d", Balance(acc))
	}

	if bal, ok := Withdraw(acc, 40); !ok || bal != 60 {
		t.Fatalf("Withdraw(40) = %d, %v; want 60, true", bal, ok)
	}
	if Balance(acc) != 60 {
		t.Fatalf("Balance = %d, want 60", Balance(acc))
	}
}

func TestSharedPointer(t *testing.T) {
	acc := NewAccount("Ada", 0)
	alias := acc // same pointer
	Deposit(alias, 25)
	if Balance(acc) != 25 {
		t.Fatalf("aliasing does not share state: %d", Balance(acc))
	}
}
`,
				}},
				Hints: []string{
					"§NewAccount§ returns §&Account{Owner: owner, Balance: balance}§.",
					"§a.Balance += amount§ works on a pointer without any explicit §*§ — Go inserts it for you.",
					"§Withdraw§: §if a.Balance < amount { return a.Balance, false }§, then subtract and return §true§.",
					"§Balance§ is just §return a.Balance§.",
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

func Deposit(a *Account, amount int) int {
	a.Balance += amount
	return a.Balance
}

func Withdraw(a *Account, amount int) (int, bool) {
	if a.Balance < amount {
		return a.Balance, false
	}
	a.Balance -= amount
	return a.Balance, true
}

func Balance(a *Account) int {
	return a.Balance
}

func main() {
	acc := NewAccount("Ada", 100)
	Deposit(acc, 50)
	fmt.Println(Balance(acc))
	fmt.Println(Withdraw(acc, 500))
	fmt.Println(Withdraw(acc, 20))
	fmt.Println(Balance(acc))
}
`,
			},
		},
	}
}

func m1t10() Topic {
	return Topic{
		ID:      "m1t10",
		Title:   "Structs",
		Summary: "Grouping related values into your own type.",
		Theory: `
A **struct** is a collection of named fields — your own type, made of other types. It is
the closest thing Go has to a Java class, except it holds *only* data. Behaviour is
added separately, in the next topic.

## Defining one

§§§go
type User struct {
	Name  string
	Email string
	Age   int
}
§§§

§type User struct { … }§ says: *"there is now a type called §User§, and every §User§
contains these three fields."*

By convention, type names are §CamelCase§ and exported (capitalised) if other packages
should use them.

## Creating one

There are three common ways:

§§§go
// 1. A struct literal with named fields — the way you should almost always do it.
u1 := User{Name: "Ada", Email: "ada@x.io", Age: 36}

// 2. Field by field (the zero value to start).
var u2 User
u2.Name = "Bo"

// 3. Positional: every field, in declaration order. Fragile — avoid.
u3 := User{"Cy", "cy@x.io", 21}
§§§

The named form (§u := User{Name: "Ada"}§) has two big advantages: you may leave fields
out (they get their zero value, so §u.Age§ is §0§), and adding a field to the struct
later does not silently break every literal in your program. The positional form
breaks the moment you reorder the fields.

## Reading and writing fields

§§§go
u := User{Name: "Ada"}
fmt.Println(u.Name) // Ada
u.Age = 36          // writes the field
fmt.Println(u.Age)  // 36
§§§

## Structs are values

A struct is copied when you assign it or pass it to a function — exactly like an §int§:

§§§go
type Point struct{ X, Y int }

func move(p Point) {
	p.X = 100 // changes the copy
}

func main() {
	pt := Point{X: 1, Y: 2}
	move(pt)
	fmt.Println(pt.X) // 1 — unchanged
}
§§§

If you want the function to change the caller's struct, pass a pointer:
§func move(p *Point)§ and call it as §move(&pt)§. You met this in the pointers topic;
structs are where it matters most in practice.

## Printing structs

| Verb | What you get |
| --- | --- |
| §%v§ | §{1 2}§ — values only |
| §%+v§ | §{X:1 Y:2}§ — with field names (**use this when debugging**) |
| §%#v§ | §main.Point{X:1, Y:2}§ — Go syntax that could rebuild the value |

## Structs are comparable

Two structs of the same type can be compared with §==§, as long as every field is
comparable:

§§§go
fmt.Println(Point{1, 2} == Point{1, 2}) // true
fmt.Println(Point{1, 2} == Point{1, 3}) // false
§§§

That also means a struct can be used as a **map key**:

§§§go
grid := map[Point]string{}
grid[Point{0, 0}] = "origin"
§§§

(A struct containing a slice, map or function is *not* comparable, so it cannot be used
with §==§ or as a key.)

## Nested structs

Fields can themselves be structs:

§§§go
type Address struct {
	City    string
	Country string
}

type Customer struct {
	Name    string
	Address Address
}

c := Customer{
	Name:    "Ada",
	Address: Address{City: "Oslo", Country: "NO"},
}
fmt.Println(c.Address.City) // Oslo
§§§

Note the difference between the two worlds: the literal uses §Address: Address{…}§
(field name, then the value), while reading uses §c.Address.City§ (field, field).

## A complete example

§§§go
package main

import "fmt"

type Book struct {
	Title  string
	Author string
	Pages  int
	Read   bool
}

func (b Book) Summary() string { // (methods get their own topic — this is a preview)
	return fmt.Sprintf("%s by %s, %d pages", b.Title, b.Author, b.Pages)
}

func main() {
	b := Book{Title: "The Go Programming Language", Author: "Donovan & Kernighan", Pages: 380}

	fmt.Println(b.Summary())
	fmt.Printf("%v\n", b)
	fmt.Printf("%+v\n", b)

	// A copy, then a change to the copy.
	c := b
	c.Read = true
	fmt.Println(b.Read, c.Read) // false true

	if b == (Book{Title: "The Go Programming Language", Author: "Donovan & Kernighan", Pages: 380}) {
		fmt.Println("structs compare with ==")
	}
}
§§§

Output:

§§§text
The Go Programming Language by Donovan & Kernighan, 380 pages
{The Go Programming Language Donovan & Kernighan 380 false}
{Title:The Go Programming Language Author:Donovan & Kernighan Pages:380 Read:false}
false true
structs compare with ==
§§§

## The thing to watch for

Assignment copies the whole struct. If two parts of your program should see the *same*
data, share a §*Book§, not a §Book§. If they should each have their own, copy — and
remember that a copy containing a slice or a map shares those inner values anyway.
`,
		Quiz: []Question{
			{
				ID:     "m1t10q1",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

type P struct{ X, Y int }

func main() {
	a := P{1, 2}
	b := a
	b.X = 99
	fmt.Println(a.X, b.X)

	var c P
	fmt.Printf("%+v\n", c)
}`,
				Options: []Option{
					{Text: "§1 99§ then §{X:0 Y:0}§", Correct: true, Rationale: "§b := a§ copies the whole struct, and an unset struct has zero values in every field."},
					{Text: "§99 99§ then §{X:0 Y:0}§", Correct: false, Rationale: "Structs are values: assigning one copies it."},
					{Text: "§1 99§ then §{X: Y:}§", Correct: false, Rationale: "§%+v§ prints the values too, including the zeros."},
					{Text: "It does not compile: §c§ is never assigned.", Correct: false, Rationale: "§var c P§ gives a fully zero-valued struct, ready to use."},
				},
				Explanation: "Assignment copies. §%+v§ adds field names to the default format.",
			},
			{
				ID:     "m1t10q2",
				Prompt: "Which structs can be used as a map key? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "§type A struct{ X, Y int }§", Correct: true, Rationale: "All fields are comparable, so §A§ is comparable."},
					{Text: "§type B struct{ Name string }§", Correct: true, Rationale: "Strings are comparable."},
					{Text: "§type C struct{ Tags []string }§", Correct: false, Rationale: "Slices are not comparable, so §C§ is not either."},
					{Text: "§type D struct{ Inner A }§", Correct: true, Rationale: "A struct of comparable fields is comparable, recursively."},
					{Text: "§type E struct{ F func() int }§", Correct: false, Rationale: "Functions compare only to nil, so §E§ cannot be a key."},
				},
				Explanation: "Map keys need §==§. Any struct whose fields all support §==§ does too.",
			},
			{
				ID:     "m1t10q3",
				Prompt: "Why is §User{Name: \"Ada\"}§ usually better than §User{\"Ada\", \"\", 0}§?",
				Options: []Option{
					{Text: "Because adding a new field to §User§ will not break the named form, and readers can see which field holds what.", Correct: true, Rationale: "Named literals are robust to field reordering and additions."},
					{Text: "Because the positional form does not compile.", Correct: false, Rationale: "It compiles, as long as you supply every field in order."},
					{Text: "Because the named form is faster.", Correct: false, Rationale: "The generated code is the same."},
					{Text: "Because fields cannot have zero values.", Correct: false, Rationale: "Every field has a zero value."},
				},
				Explanation: "Positional literals are a maintenance trap: reorder or add a field and every literal becomes wrong — or, worse, still compiles with the values swapped.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t10e1",
				Title: "Warm-up: a point and its translation",
				Prompt: `Define a struct and move it.

§§§go
type Point struct {
	X, Y int
}

func Translate(p *Point, dx, dy int)
func Same(a, b Point) bool
§§§

* §Translate§ adds §dx§ to §X§ and §dy§ to §Y§ — in place, so the caller sees the
  change. That means a pointer parameter.
* §Same§ reports whether two points are equal.

§X, Y int§ on a single line declares two fields of the same type — the same grouping
shortcut you saw for function parameters.
`,
				Starter: `package main

import "fmt"

type Point struct {
	X, Y int
}

// Translate moves p by dx, dy in place.
func Translate(p *Point, dx, dy int) {
	// TODO
}

// Same reports whether a and b are the same point.
func Same(a, b Point) bool {
	// TODO
	return false
}

func main() {
	p := Point{X: 1, Y: 2}
	Translate(&p, 10, -1)
	fmt.Println(p)
	fmt.Println(Same(p, Point{X: 11, Y: 1}))
}
`,
				Tests: []File{{
					Name: "point_test.go",
					Content: `package main

import "testing"

func TestTranslate(t *testing.T) {
	p := Point{X: 1, Y: 2}
	Translate(&p, 10, -1)
	if p.X != 11 || p.Y != 1 {
		t.Fatalf("after Translate: %+v, want {X:11 Y:1}", p)
	}

	q := Point{}
	Translate(&q, 0, 0)
	if q.X != 0 || q.Y != 0 {
		t.Fatalf("translating by zero changed the point: %+v", q)
	}

	r := Point{X: -5, Y: -5}
	Translate(&r, -5, 5)
	if r.X != -10 || r.Y != 0 {
		t.Fatalf("translate with negatives: %+v, want {X:-10 Y:0}", r)
	}
}

func TestSame(t *testing.T) {
	if !Same(Point{1, 2}, Point{1, 2}) {
		t.Error("identical points reported as different")
	}
	if Same(Point{1, 2}, Point{2, 1}) {
		t.Error("different points reported as the same")
	}
	if !Same(Point{}, Point{}) {
		t.Error("two zero points should be equal")
	}
}
`,
				}},
				Hints: []string{
					"§p.X += dx§ and §p.Y += dy§ — Go dereferences the pointer for you.",
					"§Same§ is a single expression: §return a == b§.",
					"Structs compare field by field, so you do not need to compare X and Y separately.",
				},
				Solution: `package main

import "fmt"

type Point struct {
	X, Y int
}

func Translate(p *Point, dx, dy int) {
	p.X += dx
	p.Y += dy
}

func Same(a, b Point) bool {
	return a == b
}

func main() {
	p := Point{X: 1, Y: 2}
	Translate(&p, 10, -1)
	fmt.Println(p)
	fmt.Println(Same(p, Point{X: 11, Y: 1}))
}
`,
			},
			{
				ID:    "m1t10e2",
				Title: "A book with a formatted description",
				Prompt: `§§§go
type Book struct {
	Title  string
	Author string
	Pages  int
	Read   bool
}

func Describe(b Book) string
func MarkRead(b Book) Book
func Longer(a, b Book) Book
§§§

* §Describe§ returns §"<Title> by <Author> (<Pages> pages)"§, and appends
  §" - read"§ when §Read§ is true.
    * §Book{Title:"Go", Author:"Ada", Pages: 300}§ → §"Go by Ada (300 pages)"§
    * The same book with §Read: true§ → §"Go by Ada (300 pages) - read"§
* §MarkRead§ returns a **copy** with §Read§ set to §true§. It must not change the
  original — the test checks this, because structs are values and this is the perfect
  place to see that.
* §Longer§ returns whichever book has more pages.

§Describe§ needs to build one of two strings. A small §if§ and §fmt.Sprintf§ is the
clearest way.
`,
				Starter: `package main

import "fmt"

type Book struct {
	Title  string
	Author string
	Pages  int
	Read   bool
}

// Describe returns a one-line human-readable summary.
func Describe(b Book) string {
	// TODO
	return ""
}

// MarkRead returns a copy of b with Read set to true.
func MarkRead(b Book) Book {
	// TODO
	return b
}

// Longer returns the book with more pages.
func Longer(a, b Book) Book {
	// TODO
	return a
}

func main() {
	goBook := Book{Title: "Go", Author: "Ada", Pages: 300}
	fmt.Println(Describe(goBook))
	fmt.Println(Describe(MarkRead(goBook)))
	fmt.Println(goBook.Read)
}
`,
				Tests: []File{{
					Name: "book_test.go",
					Content: `package main

import "testing"

func TestDescribe(t *testing.T) {
	cases := []struct {
		book Book
		want string
	}{
		{Book{Title: "Go", Author: "Ada", Pages: 300}, "Go by Ada (300 pages)"},
		{Book{Title: "Go", Author: "Ada", Pages: 300, Read: true}, "Go by Ada (300 pages) - read"},
		{Book{Title: "X", Author: "Y", Pages: 0}, "X by Y (0 pages)"},
	}
	for _, c := range cases {
		if got := Describe(c.book); got != c.want {
			t.Errorf("Describe(%+v) = %q, want %q", c.book, got, c.want)
		}
	}
}

func TestMarkReadDoesNotMutate(t *testing.T) {
	original := Book{Title: "Go", Author: "Ada", Pages: 300}
	marked := MarkRead(original)

	if !marked.Read {
		t.Error("MarkRead did not set Read on the copy")
	}
	if original.Read {
		t.Error("MarkRead changed the original book — structs are copied, so this should be impossible")
	}
}

func TestLonger(t *testing.T) {
	short := Book{Title: "Short", Pages: 100}
	long := Book{Title: "Long", Pages: 500}

	if got := Longer(short, long); got.Title != "Long" {
		t.Errorf("Longer(short, long) = %q, want Long", got.Title)
	}
	if got := Longer(long, short); got.Title != "Long" {
		t.Errorf("Longer(long, short) = %q, want Long", got.Title)
	}
	if got := Longer(short, short); got.Title != "Short" {
		t.Errorf("Longer with a tie = %q, want Short", got.Title)
	}
}
`,
				}},
				Hints: []string{
					"§fmt.Sprintf(\"%s by %s (%d pages)\", b.Title, b.Author, b.Pages)§ gives the first form.",
					"Append the read marker with §if b.Read { return base + \" - read\" }§.",
					"§MarkRead§: set §b.Read = true§ and §return b§ — since §b§ is a copy, the caller's book is untouched.",
					"§Longer§: §if b.Pages > a.Pages { return b }; return a§.",
				},
				Solution: `package main

import "fmt"

type Book struct {
	Title  string
	Author string
	Pages  int
	Read   bool
}

func Describe(b Book) string {
	base := fmt.Sprintf("%s by %s (%d pages)", b.Title, b.Author, b.Pages)
	if b.Read {
		return base + " - read"
	}
	return base
}

func MarkRead(b Book) Book {
	b.Read = true
	return b
}

func Longer(a, b Book) Book {
	if b.Pages > a.Pages {
		return b
	}
	return a
}

func main() {
	goBook := Book{Title: "Go", Author: "Ada", Pages: 300}
	fmt.Println(Describe(goBook))
	fmt.Println(Describe(MarkRead(goBook)))
	fmt.Println(goBook.Read)
}
`,
			},
			{
				ID:    "m1t10e3",
				Title: "An inventory with nested structs",
				Prompt: `Build a slightly bigger picture: a warehouse map keyed by location.

§§§go
type Location struct {
	Aisle int
	Shelf int
}

type Item struct {
	SKU      string
	Name     string
	Quantity int
	Where    Location
}

func NewItem(sku, name string, quantity int, aisle, shelf int) Item
func Restock(it Item, amount int) Item
func TotalQuantity(items []Item) int
func FindBySKU(items []Item, sku string) (Item, bool)
§§§

* §NewItem§ builds an §Item§ with the nested §Location§ filled in.
* §Restock§ returns a copy with more of the item.
* §TotalQuantity§ adds up every item's quantity. §nil§/empty gives §0§.
* §FindBySKU§ returns the matching item and §true§, or the zero §Item§ and §false§.
  This "value plus ok" shape is one you have now met three times.

The nested field is written as §Where: Location{Aisle: aisle, Shelf: shelf}§.
`,
				Starter: `package main

import "fmt"

type Location struct {
	Aisle int
	Shelf int
}

type Item struct {
	SKU      string
	Name     string
	Quantity int
	Where    Location
}

func NewItem(sku, name string, quantity int, aisle, shelf int) Item {
	// TODO
	return Item{}
}

func Restock(it Item, amount int) Item {
	// TODO
	return it
}

func TotalQuantity(items []Item) int {
	// TODO
	return 0
}

func FindBySKU(items []Item, sku string) (Item, bool) {
	// TODO
	return Item{}, false
}

func main() {
	a := NewItem("A1", "Mug", 10, 3, 2)
	b := NewItem("B2", "Plate", 5, 4, 1)
	fmt.Println(a)
	fmt.Println(Restock(a, 5))
	fmt.Println(TotalQuantity([]Item{a, b}))
	fmt.Println(FindBySKU([]Item{a, b}, "B2"))
}
`,
				Tests: []File{{
					Name: "inventory_test.go",
					Content: `package main

import "testing"

func TestNewItem(t *testing.T) {
	it := NewItem("A1", "Mug", 10, 3, 2)
	if it.SKU != "A1" || it.Name != "Mug" || it.Quantity != 10 {
		t.Fatalf("NewItem = %+v", it)
	}
	if it.Where.Aisle != 3 || it.Where.Shelf != 2 {
		t.Fatalf("nested Location = %+v, want {Aisle:3 Shelf:2}", it.Where)
	}
}

func TestRestockCopies(t *testing.T) {
	it := NewItem("A1", "Mug", 10, 3, 2)
	more := Restock(it, 5)
	if more.Quantity != 15 {
		t.Fatalf("Restock = %d, want 15", more.Quantity)
	}
	if it.Quantity != 10 {
		t.Fatalf("Restock changed the original: %d", it.Quantity)
	}
}

func TestTotalQuantity(t *testing.T) {
	items := []Item{
		NewItem("A1", "Mug", 10, 1, 1),
		NewItem("B2", "Plate", 5, 1, 2),
		NewItem("C3", "Bowl", 2, 1, 3),
	}
	if got := TotalQuantity(items); got != 17 {
		t.Fatalf("TotalQuantity = %d, want 17", got)
	}
	if got := TotalQuantity(nil); got != 0 {
		t.Fatalf("TotalQuantity(nil) = %d, want 0", got)
	}
}

func TestFindBySKU(t *testing.T) {
	items := []Item{
		NewItem("A1", "Mug", 10, 1, 1),
		NewItem("B2", "Plate", 5, 1, 2),
	}
	got, ok := FindBySKU(items, "B2")
	if !ok || got.Name != "Plate" {
		t.Fatalf("FindBySKU(B2) = %+v, %v", got, ok)
	}
	if _, ok := FindBySKU(items, "ZZ"); ok {
		t.Fatal("FindBySKU found a missing SKU")
	}
	if _, ok := FindBySKU(nil, "A1"); ok {
		t.Fatal("FindBySKU found something in an empty list")
	}
}
`,
				}},
				Hints: []string{
					"§NewItem§ returns §Item{SKU: sku, Name: name, Quantity: quantity, Where: Location{Aisle: aisle, Shelf: shelf}}§.",
					"§Restock§: §it.Quantity += amount§ then §return it§.",
					"§TotalQuantity§ loops with §for _, it := range items { total += it.Quantity }§.",
					"§FindBySKU§ loops and returns early on a match, falling through to §return Item{}, false§.",
				},
				Solution: `package main

import "fmt"

type Location struct {
	Aisle int
	Shelf int
}

type Item struct {
	SKU      string
	Name     string
	Quantity int
	Where    Location
}

func NewItem(sku, name string, quantity int, aisle, shelf int) Item {
	return Item{
		SKU:      sku,
		Name:     name,
		Quantity: quantity,
		Where:    Location{Aisle: aisle, Shelf: shelf},
	}
}

func Restock(it Item, amount int) Item {
	it.Quantity += amount
	return it
}

func TotalQuantity(items []Item) int {
	total := 0
	for _, it := range items {
		total += it.Quantity
	}
	return total
}

func FindBySKU(items []Item, sku string) (Item, bool) {
	for _, it := range items {
		if it.SKU == sku {
			return it, true
		}
	}
	return Item{}, false
}

func main() {
	a := NewItem("A1", "Mug", 10, 3, 2)
	b := NewItem("B2", "Plate", 5, 4, 1)
	fmt.Println(a)
	fmt.Println(Restock(a, 5))
	fmt.Println(TotalQuantity([]Item{a, b}))
	fmt.Println(FindBySKU([]Item{a, b}, "B2"))
}
`,
			},
		},
	}
}
