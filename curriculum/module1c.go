package curriculum

func m1t11() Topic {
	return Topic{
		ID:      "m1t11",
		Title:   "Arrays and slice basics",
		Summary: "Arrays, slices, and which one to reach for.",
		Theory: `
Go has two sequence types. One of them covers almost every case you will meet:

| | Array | Slice |
| --- | --- | --- |
| Written | §[3]int§ | §[]int§ |
| Length | fixed, part of the type | can grow |
| Declared | §var a [3]int§ | §var s []int§ |
| Use it for | almost nothing directly | everything |

## Arrays: fixed size, part of the type

§§§go
var a [3]int           // three ints, all zero
b := [3]string{"a", "b", "c"}
c := [...]int{1, 2, 3} // the compiler counts: this is [3]int
§§§

§[3]int§ and §[4]int§ are **different types**. You cannot assign one to the other, and a
function that takes §[3]int§ will not accept a §[4]int§. The length is baked in.

An array is a value, like a struct and like an §int§. Assigning it copies all the
elements:

§§§go
a := [3]int{1, 2, 3}
b := a
b[0] = 99
fmt.Println(a[0], b[0]) // 1 99 — the copy changed, the original did not
§§§

Because copying a large array is expensive and the fixed length is awkward, arrays are
mostly used for things of a known, small size — a 16-byte checksum, a 3×3 matrix, a
fixed-size buffer.

## Slices: a window onto an array

§[]int§ (notice: **no number in the brackets**) is a slice. It has a length that can
grow:

§§§go
var s []int                  // nil slice: length 0, usable
t := []int{1, 2, 3}          // slice literal
u := make([]int, 3)          // length 3, all zeros
v := make([]int, 0, 10)      // length 0, but room for 10 (we come back to this)
§§§

The three forms you will actually type:

* §[]int{1, 2, 3}§ — "make a slice with these elements".
* §make([]int, n)§ — "make a slice of n zeros".
* §make([]int, 0, n)§ — "make an empty slice that has room for n" (the preallocation
  idiom; we explain why in the next topic).

## Indexing, length, assignment

Indices start at **0** and run to §len(s)-1§:

§§§go
s := []int{10, 20, 30}
fmt.Println(s[0])    // 10
fmt.Println(s[2])    // 30
fmt.Println(len(s))  // 3
s[1] = 99            // write a position
fmt.Println(s)       // [10 99 30]

// fmt.Println(s[3])  // panic: index out of range [3] with length 3
§§§

Asking for an index that does not exist is a **runtime panic**, not a compile error —
the compiler cannot always know how long your slice is. This is the most common panic
beginners meet.

## Slicing: taking a piece of a slice

§s[low:high]§ gives you the elements from §low§ up to **but not including** §high§:

§§§go
s := []int{0, 1, 2, 3, 4, 5}
fmt.Println(s[1:4]) // [1 2 3]
fmt.Println(s[:3])  // [0 1 2]   — low defaults to 0
fmt.Println(s[3:])  // [3 4 5]   — high defaults to len
fmt.Println(s[:])   // [0 1 2 3 4 5]
§§§

§s[1:4]§ is read as "start at index 1, stop before index 4" — four minus one is three
elements. Half-open ranges like this are used everywhere in Go, and they make
§s[:n]§ (the first n) and §s[n:]§ (everything from n) read naturally.

## Walking a slice with range

§§§go
xs := []string{"Ada", "Bo", "Cy"}

for i := 0; i < len(xs); i++ {
	fmt.Println(i, xs[i])
}

// The idiomatic form:
for i, name := range xs {
	fmt.Println(i, name)
}

// Index not needed? Throw it away with _ . It is required: Go rejects unused variables.
for _, name := range xs {
	fmt.Println(name)
}
§§§

Both loops print the same three lines. §range§ is shorter, cannot run past the end, and
is what Go programmers expect to read.

One warning that the next topic unpacks properly: **§range§ gives you a copy of each
element**. Changing the copy changes nothing:

§§§go
for _, v := range []int{1, 2, 3} {
	v = v * 2 // pointless: v is a copy
}
§§§

## Adding elements with append

Slices grow with §append§. It returns a (possibly new) slice, so **you must assign the
result back**:

§§§go
s := []int{1, 2}
s = append(s, 3)          // s is now [1 2 3]
s = append(s, 4, 5)       // several at once
s = append(s, s...)       // append a whole slice; the ... spreads it
§§§

Forgetting the assignment is a classic bug — §append(s, 3)§ on its own compiles and does
nothing useful.

## The standard library already did the boring parts

Since Go 1.21 there is a §slices§ package with the helpers everyone used to write by
hand:

§§§go
import "slices"

nums := []int{3, 1, 2}

slices.Sort(nums)              // in place: [1 2 3]
fmt.Println(slices.Min(nums))  // 1
fmt.Println(slices.Max(nums))  // 3
fmt.Println(slices.Contains(nums, 2))    // true
fmt.Println(slices.Index(nums, 3))       // 2
fmt.Println(slices.Equal(nums, []int{1, 2, 3})) // true
§§§

Slices **cannot** be compared with §==§ (only to §nil§). §slices.Equal§ is the
replacement — you will see why in the next topic.

## A complete example

§§§go
package main

import (
	"fmt"
	"slices"
)

func main() {
	scores := []int{88, 95, 71, 60}

	fmt.Println("count:", len(scores))
	fmt.Println("first half:", scores[:2])
	fmt.Println("last two:", scores[2:])

	total := 0
	for _, s := range scores {
		total += s
	}
	fmt.Println("average:", float64(total)/float64(len(scores)))

	scores = append(scores, 100)
	slices.Sort(scores)
	fmt.Println("sorted:", scores)
	fmt.Println("slices.Equal:", slices.Equal(scores, []int{60, 71, 88, 95, 100}))
}
§§§

Output:

§§§text
count: 4
first half: [88 95]
last two: [71 60]
average: 78.5
sorted: [60 71 88 95 100]
slices.Equal: true
§§§

## Common mistakes

* **Off by one.** §for i := 0; i <= len(xs); i++§ panics on the last iteration.
* **Ignoring what §append§ returns.** §append(s, x)§ alone is a no-op.
* **Assuming §==§ works.** §[]int{1} == []int{1}§ does not compile.
* **A §nil§ slice is fine.** §var s []int§ has §len(s) == 0§, can be ranged over, and
  §append§ works on it. You only need §make§ when you want to assign by index
  immediately.
`,
		Quiz: []Question{
			{
				ID:     "m1t11q1",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

func main() {
	s := []int{0, 1, 2, 3, 4, 5}
	fmt.Println(s[1:4], s[:2], s[4:], len(s))
}`,
				Options: []Option{
					{Text: "§[1 2 3] [0 1] [4 5] 6§", Correct: true, Rationale: "§s[1:4]§ includes 1, 2, 3 but not index 4; §s[:2]§ is the first two; §s[4:]§ is everything from index 4."},
					{Text: "§[1 2 3 4] [0 1] [4 5] 6§", Correct: false, Rationale: "The high bound is *excluded*."},
					{Text: "§[1 2 3] [0 1 2] [5] 6§", Correct: false, Rationale: "§s[:2]§ stops before index 2, and §s[4:]§ starts at index 4."},
					{Text: "It panics.", Correct: false, Rationale: "Every index is in range."},
				},
				Explanation: "Slicing uses a half-open range: §s[low:high]§ starts at §low§ and stops *before* §high§.",
			},
			{
				ID:     "m1t11q2",
				Prompt: "What is wrong with this code?",
				Code: `xs := []int{1, 2, 3}
append(xs, 4)
fmt.Println(xs)`,
				Options: []Option{
					{Text: "§append§ returns the slice you should use; the result here is discarded, so §xs§ is unchanged.", Correct: true, Rationale: "Write §xs = append(xs, 4)§."},
					{Text: "§append§ needs at least two elements.", Correct: false, Rationale: "One element is fine."},
					{Text: "You cannot append to a slice literal.", Correct: false, Rationale: "You can append to any slice."},
					{Text: "Nothing — it prints §[1 2 3 4]§.", Correct: false, Rationale: "The append result was thrown away."},
				},
				Explanation: "§append§ may need to allocate a new backing array, so it returns the slice to use. Always assign the result.",
			},
			{
				ID:     "m1t11q3",
				Prompt: "Which of these are true about a nil slice (§var s []int§)? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "§len(s)§ is 0.", Correct: true, Rationale: "A nil slice has length 0."},
					{Text: "§range s§ iterates zero times.", Correct: true, Rationale: "There is nothing to visit."},
					{Text: "§s = append(s, 1)§ works and gives a slice of length 1.", Correct: true, Rationale: "§append§ on a nil slice allocates a backing array."},
					{Text: "§s[0] = 1§ works and grows the slice.", Correct: false, Rationale: "Indexing past the length panics, nil or not."},
					{Text: "§s == nil§ is true.", Correct: true, Rationale: "A nil slice is equal to nil."},
				},
				Explanation: "Nil slices are the friendly zero value: readable, rangeable and appendable. Only index-assignment fails.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t11e1",
				Title: "Warm-up: add up a slice",
				Prompt: `§§§go
func Sum(xs []int) int
func Count(xs []int, target int) int
§§§

* §Sum§ adds every element. A §nil§ or empty slice gives §0§ — and this works
  automatically, because ranging over an empty slice performs zero iterations.
* §Count§ reports how many times §target§ appears.

Both take a §[]int§, which means "a slice of ints" — nothing else, no length in the
brackets.
`,
				Starter: `package main

import "fmt"

// Sum returns the total of xs.
func Sum(xs []int) int {
	// TODO
	return 0
}

// Count returns how many times target appears in xs.
func Count(xs []int, target int) int {
	// TODO
	return 0
}

func main() {
	xs := []int{1, 2, 2, 3}
	fmt.Println(Sum(xs), Count(xs, 2), Count(xs, 9))
}
`,
				Tests: []File{{
					Name: "sum_test.go",
					Content: `package main

import "testing"

func TestSum(t *testing.T) {
	cases := []struct {
		xs   []int
		want int
	}{
		{[]int{1, 2, 3}, 6},
		{[]int{-1, 1}, 0},
		{nil, 0},
		{[]int{}, 0},
		{[]int{100}, 100},
		{[]int{0, 0, 0}, 0},
		{[]int{5, 5, 5}, 15},
		{[]int{-5, -10, 3}, -12},
		{[]int{1, 2, 3, 4, 5}, 15},
		{[]int{1000000, -1000000}, 0},
	}
	for _, c := range cases {
		if got := Sum(c.xs); got != c.want {
			t.Errorf("Sum(%v) = %d, want %d", c.xs, got, c.want)
		}
	}

	// A slice with spare capacity must still only sum its length.
	spare := make([]int, 3, 10)
	copy(spare, []int{4, 5, 6})
	if got := Sum(spare); got != 15 {
		t.Errorf("Sum(slice with spare capacity) = %d, want 15", got)
	}

	// A sub-slice shares a backing array; Sum must not read past its length.
	base := []int{1, 2, 3, 4, 5}
	if got := Sum(base[1:4]); got != 9 {
		t.Errorf("Sum(base[1:4]) = %d, want 9", got)
	}
}

func TestCount(t *testing.T) {
	xs := []int{1, 2, 2, 3, 2}
	if got := Count(xs, 2); got != 3 {
		t.Errorf("Count(2) = %d, want 3", got)
	}
	if got := Count(xs, 99); got != 0 {
		t.Errorf("Count(99) = %d, want 0", got)
	}
	if got := Count(nil, 1); got != 0 {
		t.Errorf("Count on nil = %d, want 0", got)
	}

	cases := []struct {
		xs     []int
		target int
		want   int
	}{
		{[]int{}, 0, 0},
		{[]int{}, 7, 0},
		{[]int{7}, 7, 1},
		{[]int{7}, 8, 0},
		{[]int{0, 0, 0}, 0, 3},
		{[]int{-1, -1, 2}, -1, 2},
		{[]int{1, 2, 3}, 9, 0},
		{[]int{2, 2, 2, 2}, 2, 4},
		{[]int{1, 2, 2, 3, 2}, 2, 3},
	}
	for _, c := range cases {
		if got := Count(c.xs, c.target); got != c.want {
			t.Errorf("Count(%v, %d) = %d, want %d", c.xs, c.target, got, c.want)
		}
	}

	// Spare capacity is not part of the slice.
	spare := make([]int, 3, 8)
	copy(spare, []int{3, 3, 4})
	if got := Count(spare, 3); got != 2 {
		t.Errorf("Count(slice with spare capacity) = %d, want 2", got)
	}

	// Shared backing array: only the window counts.
	base := []int{2, 1, 2, 1, 2}
	if got := Count(base[1:4], 2); got != 1 {
		t.Errorf("Count(base[1:4], 2) = %d, want 1", got)
	}
}
`,
				}},
				Hints: []string{
					"Both functions are a single §for _, v := range xs§ loop with an accumulator.",
					"§total += v§ and §if v == target { n++ }§.",
					"Return the accumulator at the end — do not return inside the loop.",
				},
				Solution: `package main

import "fmt"

func Sum(xs []int) int {
	total := 0
	for _, v := range xs {
		total += v
	}
	return total
}

func Count(xs []int, target int) int {
	n := 0
	for _, v := range xs {
		if v == target {
			n++
		}
	}
	return n
}

func main() {
	xs := []int{1, 2, 2, 3}
	fmt.Println(Sum(xs), Count(xs, 2), Count(xs, 9))
}
`,
			},
			{
				ID:    "m1t11e2",
				Title: "First, last and middle",
				Prompt: `Practise reading positions and slicing.

§§§go
func FirstLast(xs []int) (int, int, bool)
func Middle(xs []int) []int
func DropFirst(xs []int) []int
§§§

* §FirstLast§ returns the first element, the last element, and §true§. For an empty or
  nil slice it returns §0, 0, false§ — the fourth time you have used this "value plus
  ok" shape.
* §Middle§ returns everything except the first and last elements. For fewer than three
  elements it returns an empty (length-0) slice.
* §DropFirst§ returns everything from index 1 onwards. For an empty slice, an empty
  result.

Each is one line of slicing once you have guarded the empty case.
`,
				Starter: `package main

import "fmt"

// FirstLast returns the first and last elements, and whether xs was non-empty.
func FirstLast(xs []int) (int, int, bool) {
	// TODO
	return 0, 0, false
}

// Middle returns xs without its first and last elements.
func Middle(xs []int) []int {
	// TODO
	return nil
}

// DropFirst returns xs without its first element.
func DropFirst(xs []int) []int {
	// TODO
	return nil
}

func main() {
	fmt.Println(FirstLast([]int{1, 2, 3}))
	fmt.Println(Middle([]int{1, 2, 3, 4, 5}))
	fmt.Println(DropFirst([]int{1, 2, 3}))
}
`,
				Tests: []File{{
					Name: "edges_test.go",
					Content: `package main

import "testing"

func TestFirstLast(t *testing.T) {
	if a, b, ok := FirstLast([]int{1, 2, 3}); !ok || a != 1 || b != 3 {
		t.Errorf("FirstLast([1 2 3]) = %d, %d, %v; want 1, 3, true", a, b, ok)
	}
	if a, b, ok := FirstLast([]int{7}); !ok || a != 7 || b != 7 {
		t.Errorf("FirstLast([7]) = %d, %d, %v; want 7, 7, true", a, b, ok)
	}
	if _, _, ok := FirstLast(nil); ok {
		t.Error("FirstLast(nil) reported ok")
	}
	if _, _, ok := FirstLast([]int{}); ok {
		t.Error("FirstLast(empty) reported ok")
	}
}

func TestMiddle(t *testing.T) {
	if got := Middle([]int{1, 2, 3, 4, 5}); len(got) != 3 || got[0] != 2 || got[2] != 4 {
		t.Errorf("Middle = %v, want [2 3 4]", got)
	}
	if got := Middle([]int{1, 2, 3}); len(got) != 1 || got[0] != 2 {
		t.Errorf("Middle([1 2 3]) = %v, want [2]", got)
	}
	for _, xs := range [][]int{nil, {}, {1}, {1, 2}} {
		if got := Middle(xs); len(got) != 0 {
			t.Errorf("Middle(%v) = %v, want an empty slice", xs, got)
		}
	}
}

func TestDropFirst(t *testing.T) {
	if got := DropFirst([]int{1, 2, 3}); len(got) != 2 || got[0] != 2 {
		t.Errorf("DropFirst = %v, want [2 3]", got)
	}
	if got := DropFirst([]int{9}); len(got) != 0 {
		t.Errorf("DropFirst([9]) = %v, want empty", got)
	}
	if got := DropFirst(nil); len(got) != 0 {
		t.Errorf("DropFirst(nil) = %v, want empty", got)
	}
}

func sameInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestFirstLastTable(t *testing.T) {
	cases := []struct {
		xs          []int
		first, last int
		ok          bool
	}{
		{nil, 0, 0, false},
		{[]int{}, 0, 0, false},
		{[]int{4}, 4, 4, true},
		{[]int{-3}, -3, -3, true},
		{[]int{1, 2}, 1, 2, true},
		{[]int{-3, -2, -1}, -3, -1, true},
		{[]int{10, 20, 30, 40}, 10, 40, true},
		{[]int{0, 0, 0, 0}, 0, 0, true},
	}
	for _, c := range cases {
		first, last, ok := FirstLast(c.xs)
		if ok != c.ok || first != c.first || last != c.last {
			t.Errorf("FirstLast(%v) = %d, %d, %v; want %d, %d, %v",
				c.xs, first, last, ok, c.first, c.last, c.ok)
		}
	}
}

func TestMiddleContent(t *testing.T) {
	cases := []struct {
		xs   []int
		want []int
	}{
		{nil, []int{}},
		{[]int{}, []int{}},
		{[]int{1}, []int{}},
		{[]int{1, 2}, []int{}},
		{[]int{1, 2, 3}, []int{2}},
		{[]int{1, 2, 3, 4}, []int{2, 3}},
		{[]int{1, 2, 3, 4, 5, 6}, []int{2, 3, 4, 5}},
		{[]int{-2, -1, 0, 1}, []int{-1, 0}},
	}
	for _, c := range cases {
		if got := Middle(c.xs); !sameInts(got, c.want) {
			t.Errorf("Middle(%v) = %v, want %v", c.xs, got, c.want)
		}
	}

	// A slice with spare capacity is still just its length.
	spare := make([]int, 0, 10)
	spare = append(spare, 1, 2, 3, 4, 5)
	if got := Middle(spare); !sameInts(got, []int{2, 3, 4}) {
		t.Errorf("Middle(slice with spare capacity) = %v, want [2 3 4]", got)
	}
}

func TestDropFirstContent(t *testing.T) {
	cases := []struct {
		xs   []int
		want []int
	}{
		{nil, []int{}},
		{[]int{}, []int{}},
		{[]int{9}, []int{}},
		{[]int{1, 2}, []int{2}},
		{[]int{1, 2, 3, 4}, []int{2, 3, 4}},
		{[]int{-1, 0, 1}, []int{0, 1}},
	}
	for _, c := range cases {
		if got := DropFirst(c.xs); !sameInts(got, c.want) {
			t.Errorf("DropFirst(%v) = %v, want %v", c.xs, got, c.want)
		}
	}

	// Only the part after index 0 is returned, even with spare capacity.
	spare := make([]int, 0, 12)
	spare = append(spare, 7, 8, 9)
	if got := DropFirst(spare); !sameInts(got, []int{8, 9}) {
		t.Errorf("DropFirst(slice with spare capacity) = %v, want [8 9]", got)
	}
}
`,
				}},
				Hints: []string{
					"§FirstLast§: §if len(xs) == 0 { return 0, 0, false }§ then §return xs[0], xs[len(xs)-1], true§.",
					"§Middle§: guard §len(xs) < 3§ first, then §return xs[1 : len(xs)-1]§.",
					"§DropFirst§: guard §len(xs) == 0§, then §return xs[1:]§.",
				},
				Solution: `package main

import "fmt"

func FirstLast(xs []int) (int, int, bool) {
	if len(xs) == 0 {
		return 0, 0, false
	}
	return xs[0], xs[len(xs)-1], true
}

func Middle(xs []int) []int {
	if len(xs) < 3 {
		return []int{}
	}
	return xs[1 : len(xs)-1]
}

func DropFirst(xs []int) []int {
	if len(xs) == 0 {
		return []int{}
	}
	return xs[1:]
}

func main() {
	fmt.Println(FirstLast([]int{1, 2, 3}))
	fmt.Println(Middle([]int{1, 2, 3, 4, 5}))
	fmt.Println(DropFirst([]int{1, 2, 3}))
}
`,
			},
			{
				ID:    "m1t11e3",
				Title: "Build a slice with append",
				Prompt: `§§§go
func Evens(limit int) []int
func RunningTotal(xs []int) []int
func Repeat(s string, n int) []string
§§§

* §Evens(limit)§ returns every even number in §[0, limit]§ inclusive:
  §Evens(10)§ → §[0 2 4 6 8 10]§. For a negative limit, return an empty slice.
* §RunningTotal(xs)§ returns a new slice the same length as §xs§, where element §i§ is
  the sum of §xs[0..i]§: §RunningTotal([1 2 3])§ → §[1 3 6]§.
* §Repeat(s, n)§ returns a slice containing §s§ repeated §n§ times:
  §Repeat("x", 3)§ → §[x x x]§. §n <= 0§ gives an empty slice.

All three are §make(…, 0, …)§ plus §append§ — the pattern for "build a new list".
Note that §make§ panics on a negative length or capacity, so any negative input must be
guarded *before* you preallocate.
`,
				Starter: `package main

import "fmt"

// Evens returns every even number from 0 to limit inclusive.
func Evens(limit int) []int {
	// TODO
	return nil
}

// RunningTotal returns the cumulative sums of xs.
func RunningTotal(xs []int) []int {
	// TODO
	return nil
}

// Repeat returns s repeated n times.
func Repeat(s string, n int) []string {
	// TODO
	return nil
}

func main() {
	fmt.Println(Evens(10))
	fmt.Println(RunningTotal([]int{1, 2, 3}))
	fmt.Println(Repeat("x", 3))
}
`,
				Tests: []File{{
					Name: "build_test.go",
					Content: `package main

import "testing"

func TestEvens(t *testing.T) {
	cases := []struct {
		limit int
		want  []int
	}{
		{10, []int{0, 2, 4, 6, 8, 10}},
		{0, []int{0}},
		{-5, []int{}},
		{-1, []int{}},
		{1, []int{0}},
		{2, []int{0, 2}},
		{3, []int{0, 2}},
		{4, []int{0, 2, 4}},
		{5, []int{0, 2, 4}},
		{11, []int{0, 2, 4, 6, 8, 10}},
	}
	for _, c := range cases {
		got := Evens(c.limit)
		if len(got) != len(c.want) {
			t.Fatalf("Evens(%d) = %v, want %v", c.limit, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Fatalf("Evens(%d) = %v, want %v", c.limit, got, c.want)
			}
		}
	}
}

func TestRunningTotal(t *testing.T) {
	got := RunningTotal([]int{1, 2, 3})
	want := []int{1, 3, 6}
	if len(got) != len(want) {
		t.Fatalf("RunningTotal = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("RunningTotal = %v, want %v", got, want)
		}
	}
	if len(RunningTotal(nil)) != 0 {
		t.Error("RunningTotal(nil) should be empty")
	}
}

func TestRepeat(t *testing.T) {
	if got := Repeat("x", 3); len(got) != 3 || got[0] != "x" || got[2] != "x" {
		t.Errorf("Repeat(\"x\", 3) = %v", got)
	}
	if got := Repeat("x", 0); len(got) != 0 {
		t.Errorf("Repeat(\"x\", 0) = %v, want empty", got)
	}
	if got := Repeat("x", -2); len(got) != 0 {
		t.Errorf("Repeat(\"x\", -2) = %v, want empty", got)
	}
}

func sameInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestRunningTotalTable(t *testing.T) {
	cases := []struct {
		xs   []int
		want []int
	}{
		{nil, []int{}},
		{[]int{}, []int{}},
		{[]int{5}, []int{5}},
		{[]int{-1, 2, -3}, []int{-1, 1, -2}},
		{[]int{0, 0, 0}, []int{0, 0, 0}},
		{[]int{10, -10, 10}, []int{10, 0, 10}},
		{[]int{1, 1, 1, 1}, []int{1, 2, 3, 4}},
		{[]int{-5, -5, 5}, []int{-5, -10, -5}},
	}
	for _, c := range cases {
		got := RunningTotal(c.xs)
		if !sameInts(got, c.want) {
			t.Errorf("RunningTotal(%v) = %v, want %v", c.xs, got, c.want)
		}
	}

	in := []int{1, 2, 3}
	_ = RunningTotal(in)
	if !sameInts(in, []int{1, 2, 3}) {
		t.Errorf("RunningTotal mutated its input: %v", in)
	}

	spare := make([]int, 0, 12)
	spare = append(spare, 2, 4, 6)
	if got := RunningTotal(spare); !sameInts(got, []int{2, 6, 12}) {
		t.Errorf("RunningTotal(slice with spare capacity) = %v, want [2 6 12]", got)
	}
}

func TestRepeatTable(t *testing.T) {
	cases := []struct {
		s    string
		n    int
		want []string
	}{
		{"", 3, []string{"", "", ""}},
		{"x", 1, []string{"x"}},
		{"ab", 2, []string{"ab", "ab"}},
		{"é", 3, []string{"é", "é", "é"}},
		{"word", 0, []string{}},
		{"word", -1, []string{}},
	}
	for _, c := range cases {
		if got := Repeat(c.s, c.n); !sameStrings(got, c.want) {
			t.Errorf("Repeat(%q, %d) = %v, want %v", c.s, c.n, got, c.want)
		}
	}
}
`,
				}},
				Hints: []string{
					"§Evens§: §out := make([]int, 0, limit/2+1)§, then §for i := 0; i <= limit; i += 2 { out = append(out, i) }§.",
					"§RunningTotal§: keep a §sum§ outside the loop; append it after each addition.",
					"§Repeat§: guard §n <= 0§ first so §make§ never sees a negative capacity.",
					"Return an empty slice rather than nil where the test checks the length — both have length 0, so either passes here.",
				},
				Solution: `package main

import "fmt"

func Evens(limit int) []int {
	if limit < 0 {
		return []int{}
	}
	out := make([]int, 0, limit/2+2)
	for i := 0; i <= limit; i += 2 {
		out = append(out, i)
	}
	return out
}

func RunningTotal(xs []int) []int {
	out := make([]int, 0, len(xs))
	sum := 0
	for _, x := range xs {
		sum += x
		out = append(out, sum)
	}
	return out
}

func Repeat(s string, n int) []string {
	if n <= 0 {
		return []string{}
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, s)
	}
	return out
}

func main() {
	fmt.Println(Evens(10))
	fmt.Println(RunningTotal([]int{1, 2, 3}))
	fmt.Println(Repeat("x", 3))
}
`,
			},
		},
	}
}

func m1t12() Topic {
	return Topic{
		ID:      "m1t12",
		Title:   "Slices under the hood: len, cap and shared memory",
		Summary: "The three-word slice header, how append grows, and the aliasing trap.",
		Theory: `
A slice value is a small descriptor, not the data itself. Understanding that descriptor
explains nearly every slice bug you will write.

## A slice is a small descriptor, not a list

A slice value is three words long:

§§§text
+------+-----+-----+
| ptr  | len | cap |
+------+-----+-----+
   |       \     \
   |        \     -> capacity: how many elements the backing array holds
   |         -> length: how many you may access with s[i]
   -> pointer to the first element of an array somewhere
§§§

The **backing array** is the actual storage. The slice just describes a window onto it.

§§§go
s := make([]int, 3, 5)
fmt.Println(len(s), cap(s)) // 3 5
§§§

* §len§ — how many elements there are. §s[i]§ is valid for §i < len§.
* §cap§ — how many elements would fit before a new array is needed.

§cap§ is always §>= len§.

## Assignment copies the descriptor, not the data

§§§go
a := []int{1, 2, 3, 4, 5}
b := a          // b is a *separate* 3-word descriptor...
b[0] = 99       // ...pointing at the SAME array
fmt.Println(a)  // [99 2 3 4 5]
§§§

Two variables, one array. This is the single most important thing to internalise.

## Slicing shares the array too

§§§go
a := []int{1, 2, 3, 4, 5}
b := a[1:3]     // b == [2 3]
b[0] = 99
fmt.Println(a)  // [1 99 3 4 5]
§§§

§b§ is not a copy of anything. It is a window that starts one element later in the
*same* array. Writing through the window is visible through §a§.

The numbers behind that slice:

§§§go
a := []int{1, 2, 3, 4, 5}   // len 5, cap 5
b := a[1:3]                 // len 2, cap 4  <- cap counts to the end of the array!
c := a[1:3:3]               // len 2, cap 2  <- the 3-index form limits cap
§§§

The capacity needs a closer look. §b§ has §cap 4§, not 2, because the backing
array continues past where §b§ ends. §b§ is allowed to grow into §a§'s tail. The
three-index form §a[low:high:max]§ sets §cap = max - low§, which *stops* a callee from
writing into your data.

## How append grows

§append§ does one of two things:

* If there is free capacity, it writes into the existing array and returns a slice
  with a bigger length. **Your original slice shares that array.**
* If §len == cap§, it allocates a new, bigger array, copies everything across, and
  returns a slice pointing at the new array. Your original is now independent.

§§§go
package main

import "fmt"

func main() {
	a := make([]int, 0, 2)
	fmt.Println(len(a), cap(a)) // 0 2

	a = append(a, 1, 2)
	fmt.Println(len(a), cap(a)) // 2 2

	b := append(a, 3)           // no room: new array is allocated
	b[0] = 99
	fmt.Println(a, b)           // [1 2] [99 2 3]  <- independent

	c := make([]int, 0, 4)
	c = append(c, 1)
	d := append(c, 2, 3)        // plenty of room: same array
	d[0] = 99
	fmt.Println(c, d)           // [99] [99 2 3]   <- shared!
}
§§§

Output:

§§§text
0 2
2 2
[1 2] [99 2 3]
[99] [99 2 3]
§§§

The safe rule that falls out of all of this:

> **Always write §s = append(s, x)§.** Then you are always working with the slice
> §append§ returned, whether it is the old array or a new one.

## Aliasing in practice

The most famous slice bug in Go, in one program:

§§§go
a := []int{1, 2, 3, 4, 5}
b := a[1:3]        // [2 3], cap 4
b = append(b, 99)  // fits! writes into a[3]
fmt.Println(a)     // [1 2 3 99 5]  <- the 4 is GONE
fmt.Println(b)     // [2 3 99]
§§§

Nobody intended to overwrite the 4. The fix is to be explicit when you mean "a
detached copy":

§§§go
b := make([]int, len(a[1:3]))
copy(b, a[1:3])    // copy returns the number of elements copied
b = append(b, 99)  // now this cannot touch a
§§§

§copy(dst, src)§ copies §min(len(dst), len(src))§ elements and never grows anything.

## A memory-leak footnote

§§§go
huge := make([]byte, 1_000_000)
small := huge[:10]   // small holds a reference to the whole 1 MB array
§§§

§small§ is 10 bytes long but keeps a megabyte alive. To really release it, copy the
part you want into a fresh slice:

§§§go
small := append([]byte(nil), huge[:10]...)
§§§

You will not think about this often, but it explains many "why is my service using
gigabytes?" mysteries.

## The short version

| You write | What happens |
| --- | --- |
| §b := a§ | new descriptor, **same array** |
| §b := a[1:3]§ | new descriptor, **same array**, cap counts to the end |
| §b := a[1:3:3]§ | same array, but §cap§ is clamped so §append§ must reallocate |
| §b := make(…); copy(b, a)§ | a genuinely independent copy |
| §append§ with room left | writes in place, sharing the array |
| §append§ with no room | allocates, copies, returns a slice on the new array |
`,
		Quiz: []Question{
			{
				ID:     "m1t12q1",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

func main() {
	a := []int{1, 2, 3, 4, 5}
	b := a[1:3]
	fmt.Println(len(b), cap(b))
}`,
				Options: []Option{
					{Text: "§2 4§", Correct: true, Rationale: "§b§ has two elements but its capacity runs to the end of §a§'s array: 2 elements plus 2 more."},
					{Text: "§2 2§", Correct: false, Rationale: "Capacity is not the length. It counts to the end of the backing array."},
					{Text: "§3 4§", Correct: false, Rationale: "§a[1:3]§ contains indices 1 and 2 — two elements."},
					{Text: "§2 5§", Correct: false, Rationale: "Capacity is measured from where the slice starts: §5 - 1 = 4§."},
				},
				Explanation: "§cap(s[low:high])§ is §cap(s) - low§. The backing array does not shrink just because you are looking at part of it.",
			},
			{
				ID:     "m1t12q2",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

func main() {
	a := []int{1, 2, 3, 4, 5}
	b := a[1:3]
	b = append(b, 99)
	fmt.Println(a)
	fmt.Println(b)
}`,
				Options: []Option{
					{Text: "§[1 2 3 99 5]§ and §[2 3 99]§", Correct: true, Rationale: "§b§ had spare capacity, so §append§ wrote 99 into §a§'s index 3, overwriting the 4."},
					{Text: "§[1 2 3 4 5]§ and §[2 3 99]§", Correct: false, Rationale: "No new array was allocated, because §cap(b)§ was 4 and §len(b)§ was 2."},
					{Text: "§[1 2 3 4 5]§ and §[2 3]§", Correct: false, Rationale: "The append result is assigned back to §b§."},
					{Text: "It panics.", Correct: false, Rationale: "Everything is in bounds."},
				},
				Explanation: "This is *the* slice aliasing trap. Slicing shares the backing array, and §append§ only reallocates when capacity runs out.",
			},
			{
				ID:     "m1t12q3",
				Prompt: "How do you make a genuinely independent copy of a slice?",
				Multi:  true,
				Options: []Option{
					{Text: "§b := make([]int, len(a)); copy(b, a)§", Correct: true, Rationale: "A fresh array plus an explicit copy. The canonical answer."},
					{Text: "§b := append([]int(nil), a...)§", Correct: true, Rationale: "Appending to a nil slice allocates a new array, so this is a copy. The idiom `append([]T(nil), …)` is common."},
					{Text: "§b := a§", Correct: false, Rationale: "That copies the descriptor, not the data."},
					{Text: "§b := a[:]§", Correct: false, Rationale: "A window over the whole array — still shared."},
					{Text: "§b := slices.Clone(a)§", Correct: true, Rationale: "The standard library's tidy version of the first option."},
				},
				Explanation: "§copy§, §slices.Clone§, or §append([]T(nil), s...)§. Everything else shares memory.",
			},
			{
				ID:     "m1t12q4",
				Prompt: "What does §copy(dst, src)§ do when §len(dst) == 2§ and §len(src) == 5§?",
				Options: []Option{
					{Text: "Copies 2 elements and returns 2. It never grows the destination.", Correct: true, Rationale: "copy transfers min(len(dst), len(src)) elements and returns that count."},
					{Text: "Grows §dst§ to 5 and copies everything.", Correct: false, Rationale: "§copy§ never changes either length."},
					{Text: "Panics with a length mismatch.", Correct: false, Rationale: "A shorter destination simply truncates the copy."},
					{Text: "Copies 0 elements.", Correct: false, Rationale: "It copies as many as fit."},
				},
				Explanation: "§copy§ is deliberately silent about mismatched lengths: it copies what fits and tells you how many.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t12e1",
				Title: "Warm-up: an independent copy",
				Prompt: `§§§go
func Clone(xs []int) []int
func CloneTail(xs []int, from int) []int
§§§

* §Clone§ returns a new slice with the same contents. After §y := Clone(x)§, changing
  §y§ must never affect §x§.
* §CloneTail(xs, from)§ returns an independent copy of §xs[from:]§. If §from§ is out of
  range (negative, or past the end), return an empty slice.
* Both must return **non-nil** slices, even for empty input, so that §len§ and §append§
  behave predictably. There is a difference worth knowing: §var out []int§ (and a §nil§
  literal) is a **nil** slice, while §make([]int, 0)§ and §[]int{}§ are non-nil, length-0
  slices.

The test writes into the result and checks the original is untouched — exactly the bug
from the theory section.
`,
				Starter: `package main

import "fmt"

// Clone returns an independent copy of xs.
func Clone(xs []int) []int {
	// TODO
	return nil
}

// CloneTail returns an independent copy of xs[from:].
func CloneTail(xs []int, from int) []int {
	// TODO
	return nil
}

func main() {
	a := []int{1, 2, 3}
	b := Clone(a)
	b[0] = 99
	fmt.Println(a, b)
	fmt.Println(CloneTail([]int{1, 2, 3, 4}, 2))
}
`,
				Tests: []File{{
					Name: "clone_test.go",
					Content: `package main

import "testing"

func TestCloneIsIndependent(t *testing.T) {
	a := []int{1, 2, 3}
	b := Clone(a)

	if len(b) != 3 || b[0] != 1 || b[2] != 3 {
		t.Fatalf("Clone = %v, want [1 2 3]", b)
	}
	b[0] = 99
	if a[0] != 1 {
		t.Fatalf("Clone shares memory with its input: a[0] = %d", a[0])
	}
}

func TestCloneEmpty(t *testing.T) {
	if got := Clone(nil); got == nil || len(got) != 0 {
		t.Fatalf("Clone(nil) = %v, want an empty non-nil slice", got)
	}
	if got := Clone([]int{}); got == nil || len(got) != 0 {
		t.Fatalf("Clone(empty) = %v, want an empty non-nil slice", got)
	}
}

func TestCloneTail(t *testing.T) {
	a := []int{1, 2, 3, 4}
	got := CloneTail(a, 2)
	if len(got) != 2 || got[0] != 3 || got[1] != 4 {
		t.Fatalf("CloneTail(a, 2) = %v, want [3 4]", got)
	}
	got[0] = 99
	if a[2] != 3 {
		t.Fatalf("CloneTail shares memory: a = %v", a)
	}
}

func TestCloneTailOutOfRange(t *testing.T) {
	a := []int{1, 2, 3}
	for _, from := range []int{-1, 3, 99} {
		if got := CloneTail(a, from); len(got) != 0 {
			t.Errorf("CloneTail(a, %d) = %v, want empty", from, got)
		}
	}
	if got := CloneTail(a, 0); len(got) != 3 {
		t.Errorf("CloneTail(a, 0) = %v, want all three", got)
	}
}

func sameInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestCloneTable(t *testing.T) {
	cases := []struct {
		name string
		xs   []int
	}{
		{"nil", nil},
		{"empty", []int{}},
		{"one", []int{7}},
		{"many", []int{1, 2, 3, 4, 5}},
		{"negative", []int{-3, -2, -1, 0}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Clone(c.xs)
			if got == nil {
				t.Fatalf("Clone(%v) returned nil; want an empty non-nil slice", c.xs)
			}
			if !sameInts(got, c.xs) {
				t.Fatalf("Clone(%v) = %v", c.xs, got)
			}
			before := append([]int(nil), c.xs...)
			if len(got) > 0 {
				got[0] = 999
			}
			if !sameInts(c.xs, before) {
				t.Fatalf("Clone shares memory with its input: got %v, input %v", got, c.xs)
			}
		})
	}
}

func TestCloneWithSpareCapacityAndSubslice(t *testing.T) {
	spare := make([]int, 3, 10)
	copy(spare, []int{4, 5, 6})
	got := Clone(spare)
	got[1] = 99
	if spare[1] != 5 {
		t.Fatalf("Clone of a slice with spare capacity shares memory: %v", spare)
	}

	base := []int{1, 2, 3, 4, 5}
	sub := Clone(base[1:4])
	if !sameInts(sub, []int{2, 3, 4}) {
		t.Fatalf("Clone(base[1:4]) = %v, want [2 3 4]", sub)
	}
	sub[0] = 99
	if base[1] != 2 {
		t.Fatalf("Clone of a sub-slice shares memory: %v", base)
	}
}

func TestCloneInputMutationDoesNotAffectClone(t *testing.T) {
	a := []int{1, 2, 3}
	b := Clone(a)
	a[0] = 99
	if !sameInts(b, []int{1, 2, 3}) {
		t.Fatalf("mutating the input changed the clone: %v", b)
	}
}

func TestCloneTailTable(t *testing.T) {
	xs := []int{1, 2, 3, 4}
	cases := []struct {
		from int
		want []int
	}{
		{0, []int{1, 2, 3, 4}},
		{1, []int{2, 3, 4}},
		{2, []int{3, 4}},
		{3, []int{4}},
		{4, []int{}},
		{-1, []int{}},
		{5, []int{}},
		{99, []int{}},
	}
	for _, c := range cases {
		got := CloneTail(xs, c.from)
		if got == nil {
			t.Errorf("CloneTail(xs, %d) returned nil; want a non-nil slice", c.from)
			continue
		}
		if !sameInts(got, c.want) {
			t.Errorf("CloneTail(xs, %d) = %v, want %v", c.from, got, c.want)
			continue
		}
		if len(got) > 0 {
			got[0] = 999
		}
		if !sameInts(xs, []int{1, 2, 3, 4}) {
			t.Fatalf("CloneTail(xs, %d) shares memory with xs: %v", c.from, xs)
		}
	}
}

func TestCloneTailEmptyInputs(t *testing.T) {
	for _, in := range [][]int{nil, {}} {
		if got := CloneTail(in, 0); got == nil || len(got) != 0 {
			t.Errorf("CloneTail(%v, 0) = %v, want an empty non-nil slice", in, got)
		}
	}

	spare := make([]int, 0, 10)
	spare = append(spare, 10, 20, 30)
	got := CloneTail(spare, 1)
	if !sameInts(got, []int{20, 30}) {
		t.Fatalf("CloneTail(slice with spare capacity, 1) = %v, want [20 30]", got)
	}
	got[0] = 999
	if spare[1] != 20 {
		t.Fatalf("CloneTail shares memory: %v", spare)
	}
}
`,
				}},
				Hints: []string{
					"§out := make([]int, len(xs))§ then §copy(out, xs)§ — and return §out§.",
					"Because §make([]int, len(xs))§ with an empty §xs§ already gives a non-nil slice, §Clone§ handles the empty case for free.",
					"§CloneTail§: validate §from§ with §if from < 0 || from > len(xs) { return []int{} }§, then §return Clone(xs[from:])§.",
				},
				Solution: `package main

import "fmt"

func Clone(xs []int) []int {
	out := make([]int, len(xs))
	copy(out, xs)
	return out
}

func CloneTail(xs []int, from int) []int {
	if from < 0 || from > len(xs) {
		return []int{}
	}
	return Clone(xs[from:])
}

func main() {
	a := []int{1, 2, 3}
	b := Clone(a)
	b[0] = 99
	fmt.Println(a, b)
	fmt.Println(CloneTail([]int{1, 2, 3, 4}, 2))
}
`,
			},
			{
				ID:    "m1t12e2",
				Title: "Filter and remove without surprises",
				Prompt: `§§§go
func Filter(xs []int, keep func(int) bool) []int
func RemoveAt(xs []int, i int) []int
func Without(xs []int, target int) []int
§§§

* §Filter§ returns a new slice containing only the elements for which §keep§ is true.
  It must **not** modify §xs§.
* §RemoveAt§ returns a new slice with the element at index §i§ removed. If §i§ is out
  of range, return a copy of the input unchanged (still a copy — the test mutates the
  result).
* §Without§ returns a new slice with every occurrence of §target§ removed.

Every one of these must return a *detached* slice. The naive §append(xs[:i], xs[i+1:]...)§
trick writes into the caller's array, and the test catches it.
`,
				Starter: `package main

import "fmt"

// Filter returns the elements of xs for which keep returns true.
func Filter(xs []int, keep func(int) bool) []int {
	// TODO
	return nil
}

// RemoveAt returns a copy of xs without the element at index i.
func RemoveAt(xs []int, i int) []int {
	// TODO
	return nil
}

// Without returns a copy of xs with every occurrence of target removed.
func Without(xs []int, target int) []int {
	// TODO
	return nil
}

func main() {
	xs := []int{1, 2, 3, 4}
	fmt.Println(Filter(xs, func(n int) bool { return n%2 == 0 }))
	fmt.Println(RemoveAt(xs, 1))
	fmt.Println(Without([]int{1, 2, 1, 3}, 1))
}
`,
				Tests: []File{{
					Name: "filter_test.go",
					Content: `package main

import "testing"

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestFilter(t *testing.T) {
	in := []int{1, 2, 3, 4}
	got := Filter(in, func(n int) bool { return n%2 == 0 })
	if !equal(got, []int{2, 4}) {
		t.Fatalf("Filter = %v, want [2 4]", got)
	}
	if !equal(in, []int{1, 2, 3, 4}) {
		t.Fatalf("Filter mutated its input: %v", in)
	}
	if len(Filter(nil, func(int) bool { return true })) != 0 {
		t.Error("Filter(nil) should be empty")
	}
}

func TestRemoveAtDetached(t *testing.T) {
	in := []int{10, 20, 30}
	got := RemoveAt(in, 1)
	if !equal(got, []int{10, 30}) {
		t.Fatalf("RemoveAt(_, 1) = %v, want [10 30]", got)
	}
	if !equal(in, []int{10, 20, 30}) {
		t.Fatalf("RemoveAt mutated its input: %v", in)
	}
	got[0] = 999
	if in[0] != 10 {
		t.Fatalf("RemoveAt returned a slice sharing memory: %v", in)
	}

	if !equal(RemoveAt(in, 9), in) {
		t.Error("RemoveAt with an out-of-range index should return the input unchanged")
	}
	if !equal(RemoveAt(in, -1), in) {
		t.Error("RemoveAt with a negative index should return the input unchanged")
	}
}

func TestWithout(t *testing.T) {
	got := Without([]int{1, 2, 1, 3, 1}, 1)
	if !equal(got, []int{2, 3}) {
		t.Fatalf("Without = %v, want [2 3]", got)
	}
	if !equal(Without([]int{5, 5}, 9), []int{5, 5}) {
		t.Error("Without should keep elements that are not the target")
	}
	if len(Without(nil, 1)) != 0 {
		t.Error("Without(nil) should be empty")
	}
}

func TestFilterTable(t *testing.T) {
	cases := []struct {
		name string
		in   []int
		keep func(int) bool
		want []int
	}{
		{"all", []int{1, 2, 3}, func(int) bool { return true }, []int{1, 2, 3}},
		{"none", []int{1, 3, 5}, func(n int) bool { return n%2 == 0 }, []int{}},
		{"positive", []int{-2, -1, 0, 1, 2}, func(n int) bool { return n > 0 }, []int{1, 2}},
		{"nil", nil, func(int) bool { return true }, []int{}},
		{"empty", []int{}, func(int) bool { return true }, []int{}},
		{"duplicates", []int{2, 2, 3, 2}, func(n int) bool { return n == 2 }, []int{2, 2, 2}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			before := append([]int(nil), c.in...)
			got := Filter(c.in, c.keep)
			if !equal(got, c.want) {
				t.Fatalf("Filter(%v) = %v, want %v", c.in, got, c.want)
			}
			if !equal(c.in, before) {
				t.Fatalf("Filter mutated its input: %v", c.in)
			}
			if len(got) > 0 {
				got[0] = 999
				if !equal(c.in, before) {
					t.Fatalf("Filter returned a slice sharing memory with its input: %v", c.in)
				}
			}
		})
	}
}

func TestRemoveAtTable(t *testing.T) {
	in := []int{10, 20, 30, 40}
	cases := []struct {
		i    int
		want []int
	}{
		{0, []int{20, 30, 40}},
		{1, []int{10, 30, 40}},
		{2, []int{10, 20, 40}},
		{3, []int{10, 20, 30}},
		{4, []int{10, 20, 30, 40}},
		{-1, []int{10, 20, 30, 40}},
		{99, []int{10, 20, 30, 40}},
	}
	for _, c := range cases {
		got := RemoveAt(in, c.i)
		if !equal(got, c.want) {
			t.Errorf("RemoveAt(_, %d) = %v, want %v", c.i, got, c.want)
			continue
		}
		if !equal(in, []int{10, 20, 30, 40}) {
			t.Fatalf("RemoveAt(_, %d) mutated its input: %v", c.i, in)
		}
		if len(got) > 0 {
			got[0] = 999
		}
		if !equal(in, []int{10, 20, 30, 40}) {
			t.Fatalf("RemoveAt(_, %d) returned a slice sharing memory: %v", c.i, in)
		}
	}
}

func TestRemoveAtSmallInputs(t *testing.T) {
	one := []int{7}
	if got := RemoveAt(one, 0); len(got) != 0 {
		t.Errorf("RemoveAt([7], 0) = %v, want empty", got)
	}
	if !equal(one, []int{7}) {
		t.Errorf("RemoveAt([7], 0) mutated its input: %v", one)
	}
	if got := RemoveAt(nil, 0); len(got) != 0 {
		t.Errorf("RemoveAt(nil, 0) = %v, want empty", got)
	}
	if got := RemoveAt([]int{}, 0); len(got) != 0 {
		t.Errorf("RemoveAt(empty, 0) = %v, want empty", got)
	}
}

func TestWithoutTable(t *testing.T) {
	cases := []struct {
		name   string
		xs     []int
		target int
		want   []int
	}{
		{"duplicates", []int{1, 2, 1, 3, 1}, 1, []int{2, 3}},
		{"absent", []int{1, 2, 3}, 9, []int{1, 2, 3}},
		{"all", []int{5, 5, 5}, 5, []int{}},
		{"keep when target differs", []int{5, 5}, 9, []int{5, 5}},
		{"nil", nil, 1, []int{}},
		{"empty", []int{}, 1, []int{}},
		{"negative", []int{-1, 0, -1}, -1, []int{0}},
		{"interleaved", []int{1, 2, 1, 2}, 1, []int{2, 2}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			before := append([]int(nil), c.xs...)
			got := Without(c.xs, c.target)
			if !equal(got, c.want) {
				t.Fatalf("Without(%v, %d) = %v, want %v", c.xs, c.target, got, c.want)
			}
			if !equal(c.xs, before) {
				t.Fatalf("Without mutated its input: %v", c.xs)
			}
			if len(got) > 0 {
				got[0] = 999
				if !equal(c.xs, before) {
					t.Fatalf("Without returned a slice sharing memory with its input: %v", c.xs)
				}
			}
		})
	}
}
`,
				}},
				Hints: []string{
					"§Filter§: §out := make([]int, 0, len(xs))§ and append the keepers. Because the output is a brand-new array, there is nothing to detach.",
					"§RemoveAt§: validate the index first; then §out := make([]int, 0, len(xs)-1)§ and append §xs[:i]§ then §xs[i+1:]§.",
					"§Without§: same shape as §Filter§, with §if v != target§ as the condition.",
				},
				Solution: `package main

import "fmt"

func Filter(xs []int, keep func(int) bool) []int {
	out := make([]int, 0, len(xs))
	for _, v := range xs {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

func RemoveAt(xs []int, i int) []int {
	if i < 0 || i >= len(xs) {
		return Clone(xs)
	}
	out := make([]int, 0, len(xs)-1)
	out = append(out, xs[:i]...)
	out = append(out, xs[i+1:]...)
	return out
}

func Without(xs []int, target int) []int {
	out := make([]int, 0, len(xs))
	for _, v := range xs {
		if v != target {
			out = append(out, v)
		}
	}
	return out
}

func Clone(xs []int) []int {
	out := make([]int, len(xs))
	copy(out, xs)
	return out
}

func main() {
	xs := []int{1, 2, 3, 4}
	fmt.Println(Filter(xs, func(n int) bool { return n%2 == 0 }))
	fmt.Println(RemoveAt(xs, 1))
	fmt.Println(Without([]int{1, 2, 1, 3}, 1))
}
`,
			},
			{
				ID:    "m1t12e3",
				Title: "Capacity in the wild: chunking",
				Prompt: `This is the exercise where §cap§ stops being trivia.

§§§go
func Chunk(xs []int, size int) [][]int
func Merge(a, b []int) []int
§§§

* §Chunk§ splits §xs§ into consecutive groups of at most §size§ elements:
  §Chunk([1 2 3 4 5], 2)§ → §[[1 2] [3 4] [5]]§. §size <= 0§ returns §nil§. An empty
  input returns an empty result.
* §Merge§ returns a new slice containing all of §a§'s elements followed by all of
  §b§'s. It must not share memory with either input.

**The important constraint:** each chunk returned by §Chunk§ must be a *detached copy*.
If you write the obvious §xs[i:j]§, then the chunks are windows into §xs§ with spare
capacity — and the test, which writes into a chunk and checks §xs§ is unchanged,
catches it. Use a fresh §make§ plus §copy§, or the three-index form §xs[i:j:j]§.
`,
				Starter: `package main

import "fmt"

// Chunk splits xs into groups of at most size elements.
func Chunk(xs []int, size int) [][]int {
	// TODO
	return nil
}

// Merge returns a new slice with a's elements followed by b's.
func Merge(a, b []int) []int {
	// TODO
	return nil
}

func main() {
	fmt.Println(Chunk([]int{1, 2, 3, 4, 5}, 2))
	fmt.Println(Merge([]int{1, 2}, []int{3, 4}))
}
`,
				Tests: []File{{
					Name: "chunk_test.go",
					Content: `package main

import "testing"

func TestChunk(t *testing.T) {
	got := Chunk([]int{1, 2, 3, 4, 5}, 2)
	want := [][]int{{1, 2}, {3, 4}, {5}}
	if len(got) != len(want) {
		t.Fatalf("Chunk returned %d groups, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if len(got[i]) != len(want[i]) {
			t.Fatalf("group %d = %v, want %v", i, got[i], want[i])
		}
		for j := range want[i] {
			if got[i][j] != want[i][j] {
				t.Fatalf("group %d = %v, want %v", i, got[i], want[i])
			}
		}
	}
}

func TestChunkEdgeCases(t *testing.T) {
	if got := Chunk([]int{1, 2}, 0); got != nil {
		t.Errorf("Chunk(size=0) = %v, want nil", got)
	}
	if got := Chunk([]int{1, 2}, -1); got != nil {
		t.Errorf("Chunk(size=-1) = %v, want nil", got)
	}
	if got := Chunk(nil, 2); len(got) != 0 {
		t.Errorf("Chunk(nil) = %v, want empty", got)
	}
	one := Chunk([]int{1, 2, 3}, 10)
	if len(one) != 1 || len(one[0]) != 3 {
		t.Errorf("Chunk with a size larger than the input = %v", one)
	}
}

func TestChunksAreDetached(t *testing.T) {
	xs := []int{1, 2, 3, 4}
	chunks := Chunk(xs, 2)
	chunks[0][0] = 999
	if xs[0] != 1 {
		t.Fatalf("writing to a chunk corrupted the source slice: %v", xs)
	}
}

func TestMerge(t *testing.T) {
	a := []int{1, 2}
	b := []int{3, 4}
	got := Merge(a, b)
	want := []int{1, 2, 3, 4}
	if len(got) != len(want) {
		t.Fatalf("Merge = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Merge = %v, want %v", got, want)
		}
	}
	got[0] = 999
	if a[0] != 1 || b[0] != 3 {
		t.Fatalf("Merge shares memory with its inputs: a=%v b=%v", a, b)
	}
	if len(Merge(nil, nil)) != 0 {
		t.Error("Merge(nil, nil) should be empty")
	}
}

func sameInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sameChunks(a, b [][]int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !sameInts(a[i], b[i]) {
			return false
		}
	}
	return true
}

func TestChunkTable(t *testing.T) {
	cases := []struct {
		name string
		xs   []int
		size int
		want [][]int
	}{
		{"size one", []int{1, 2, 3}, 1, [][]int{{1}, {2}, {3}}},
		{"exact division", []int{1, 2, 3, 4}, 2, [][]int{{1, 2}, {3, 4}}},
		{"remainder", []int{1, 2, 3, 4, 5, 6, 7}, 3, [][]int{{1, 2, 3}, {4, 5, 6}, {7}}},
		{"size larger than input", []int{1, 2}, 10, [][]int{{1, 2}}},
		{"single element", []int{9}, 2, [][]int{{9}}},
		{"empty", []int{}, 2, [][]int{}},
		{"nil", nil, 2, [][]int{}},
		{"negative values", []int{-2, -1, 0, 1}, 2, [][]int{{-2, -1}, {0, 1}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Chunk(c.xs, c.size); !sameChunks(got, c.want) {
				t.Fatalf("Chunk(%v, %d) = %v, want %v", c.xs, c.size, got, c.want)
			}
		})
	}
}

func TestAllChunksAreDetached(t *testing.T) {
	xs := make([]int, 6, 20)
	copy(xs, []int{1, 2, 3, 4, 5, 6})
	chunks := Chunk(xs, 2)
	if len(chunks) != 3 {
		t.Fatalf("Chunk returned %d groups, want 3", len(chunks))
	}
	for i := range chunks {
		if len(chunks[i]) == 0 {
			continue
		}
		chunks[i][0] = 999
		if !sameInts(xs, []int{1, 2, 3, 4, 5, 6}) {
			t.Fatalf("writing to chunk %d corrupted the source: %v", i, xs)
		}
	}

	// The chunks must also be independent of one another.
	fresh := Chunk(xs, 2)
	fresh[0][0] = 999
	if fresh[1][0] != 3 {
		t.Fatalf("chunks share memory with one another: %v", fresh)
	}
}

func TestMergeTable(t *testing.T) {
	cases := []struct {
		name string
		a, b []int
		want []int
	}{
		{"both nil", nil, nil, []int{}},
		{"nil then values", nil, []int{1}, []int{1}},
		{"values then nil", []int{1}, nil, []int{1}},
		{"both empty", []int{}, []int{}, []int{}},
		{"a empty", []int{}, []int{2, 3}, []int{2, 3}},
		{"b empty", []int{1, 2}, []int{}, []int{1, 2}},
		{"mixed", []int{-1, 0}, []int{1, 2}, []int{-1, 0, 1, 2}},
		{"duplicates", []int{1, 1}, []int{1}, []int{1, 1, 1}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			beforeA := append([]int(nil), c.a...)
			beforeB := append([]int(nil), c.b...)
			got := Merge(c.a, c.b)
			if !sameInts(got, c.want) {
				t.Fatalf("Merge(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
			}
			if !sameInts(c.a, beforeA) || !sameInts(c.b, beforeB) {
				t.Fatalf("Merge mutated its inputs: a=%v b=%v", c.a, c.b)
			}
			if len(got) > 0 {
				got[0] = 999
				if !sameInts(c.a, beforeA) || !sameInts(c.b, beforeB) {
					t.Fatalf("Merge shares memory with its inputs: a=%v b=%v", c.a, c.b)
				}
			}
		})
	}
}

func TestMergeWithSpareCapacity(t *testing.T) {
	a := make([]int, 0, 10)
	a = append(a, 1, 2, 3)
	b := make([]int, 0, 10)
	b = append(b, 4, 5)
	got := Merge(a, b)
	if !sameInts(got, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("Merge = %v, want [1 2 3 4 5]", got)
	}
	got[0] = 99
	got[3] = 99
	if !sameInts(a, []int{1, 2, 3}) || !sameInts(b, []int{4, 5}) {
		t.Fatalf("Merge shares memory with inputs that had spare capacity: a=%v b=%v", a, b)
	}
}
`,
				}},
				Hints: []string{
					"Loop with §for i := 0; i < len(xs); i += size§ and clamp the end: §j := min(i+size, len(xs))§.",
					"§min§ is a builtin since Go 1.21 — no import needed.",
					"Detach a chunk with §chunk := make([]int, j-i); copy(chunk, xs[i:j]); out = append(out, chunk)§.",
					"§Merge§: §out := make([]int, 0, len(a)+len(b))§ then two appends.",
				},
				Solution: `package main

import "fmt"

func Chunk(xs []int, size int) [][]int {
	if size <= 0 {
		return nil
	}
	out := make([][]int, 0, (len(xs)+size-1)/size)
	for i := 0; i < len(xs); i += size {
		j := min(i+size, len(xs))
		chunk := make([]int, j-i)
		copy(chunk, xs[i:j])
		out = append(out, chunk)
	}
	return out
}

func Merge(a, b []int) []int {
	out := make([]int, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out
}

func main() {
	fmt.Println(Chunk([]int{1, 2, 3, 4, 5}, 2))
	fmt.Println(Merge([]int{1, 2}, []int{3, 4}))
}
`,
			},
		},
	}
}

func m1t13() Topic {
	return Topic{
		ID:      "m1t13",
		Title:   "Strings, bytes and runes",
		Summary: "Why len(\"héllo\") is 6, and how to walk text safely.",
		Theory: `
A Go §string§ is an immutable sequence of bytes. That one fact explains the behaviour
behind how non-ASCII text behaves.

## A string is bytes, and it is immutable

A Go §string§ is a **read-only sequence of bytes**. It is usually UTF-8 encoded text,
but the type itself is just bytes.

**A sizeable difference from Java and C#:** there is no §char§ type. §s[i]§ gives you a
§byte§ (§uint8§), not a character. And strings are **immutable**: you cannot do
§s[0] = 'H'§. To change a string you build a new one.

§§§go
s := "hello"
fmt.Println(len(s))      // 5
fmt.Println(s[0])        // 104  <- the *number* 104, which is the byte for 'h'
fmt.Printf("%c\n", s[0]) // h    <- %c prints the byte as a character
// s[0] = 'H'            // COMPILE ERROR: cannot assign to s[0]
§§§

The second line prints a number, not a character, because §s[0]§ is a §byte§. §%c§
tells §fmt§ to render it as a character instead.

## len counts bytes, not characters

This is the fact that causes most text bugs:

§§§go
héllo := "héllo"
fmt.Println(len(héllo))         // 6, not 5!
fmt.Println(len([]rune(héllo))) // 5
§§§

A §rune§ is one **Unicode code point** — one character, roughly. It is a nickname for
§int32§. The letter §é§ is a single character that UTF-8 encodes as **two** bytes, so
the string is 6 bytes and 5 characters.

A table of the moment:

| Expression | Result | Why |
| --- | --- | --- |
| §len("héllo")§ | §6§ | bytes |
| §len([]rune("héllo"))§ | §5§ | characters |
| §"é"[0]§ | §195§ | the first byte of a 2-byte character |
| §utf8.RuneCountInString("héllo")§ | §5§ | characters, without allocating |

## range over a string gives you runes

If you want characters, §range§ is the tool:

§§§go
for i, r := range "héllo" {
	fmt.Printf("position %d holds %c (%d)\n", i, r, r)
}
§§§

Output — and there is a subtlety in the positions:

§§§text
position 0 holds h (104)
position 1 holds é (233)
position 3 holds l (108)
position 4 holds l (108)
position 5 holds o (111)
§§§

§i§ jumps from 1 to 3. It is a **byte offset**, not a character index, because §é§ took
two bytes. A plain §for i := 0; i < len(s); i++§ loop would visit the middle of §é§ and
produce rubbish for anything non-ASCII. **Use §range§ to walk text.**

## Building strings without the O(n²) trap

Strings are immutable, so §s += x§ in a loop copies everything each time:

§§§go
// Slow for long inputs: each += copies the whole string so far.
s := ""
for i := 0; i < 1000; i++ {
	s += "x"
}
§§§

For anything more than a handful of appends, use §strings.Builder§, which accumulates in
a buffer and materialises the string once at the end:

§§§go
var b strings.Builder
for i := 0; i < 5; i++ {
	fmt.Fprintf(&b, "item %d; ", i)
}
result := b.String()
fmt.Println(result) // item 0; item 1; ...
§§§

Note the §&b§ — §Builder§ methods need a pointer, because they modify the builder.

## Functions you will actually use

From §strings§:

| Function | Does |
| --- | --- |
| §strings.ToUpper(s)§ / §ToLower(s)§ | change case |
| §strings.TrimSpace(s)§ | remove leading and trailing whitespace |
| §strings.Trim(s, ".,!")§ | remove any leading/trailing characters in the cutset |
| §strings.Contains(s, sub)§ | substring test |
| §strings.HasPrefix(s, p)§ / §HasSuffix§ | prefix/suffix test |
| §strings.Index(s, sub)§ | position, or §-1§ |
| §strings.Split(s, ",")§ | → §[]string§ |
| §strings.Fields(s)§ | split on *any* whitespace, dropping empties |
| §strings.Join(xs, ", ")§ | the reverse of Split |
| §strings.ReplaceAll(s, old, new)§ | replace every occurrence |
| §strings.Repeat(s, n)§ | repeat |
| §strings.EqualFold(a, b)§ | case-insensitive equality |

From §strconv§ (string ↔ number):

| Function | Does |
| --- | --- |
| §strconv.Atoi("42")§ | string → int, plus an error |
| §strconv.Itoa(42)§ | int → string |
| §strconv.ParseFloat("3.5", 64)§ | string → float64 |
| §strconv.FormatFloat(x, 'f', 2, 64)§ | float → string with 2 decimals |

§Atoi§ and §Itoa§ are the two you will type most. Note that §Atoi§ returns an error,
because §"banana"§ is a perfectly good string and a terrible number.

## A complete example

§§§go
package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func main() {
	text := "  Go 👋 "

	fmt.Printf("bytes: %d, runes: %d\n", len(text), utf8.RuneCountInString(text))

	clean := strings.TrimSpace(text)
	fmt.Println("trimmed:", clean)

	words := strings.Fields("the  quick   brown")
	fmt.Println("fields:", words, len(words))

	fmt.Println("upper:", strings.ToUpper("quiet"))

	var b strings.Builder
	b.WriteString("paths:")
	for i := 1; i <= 3; i++ {
		fmt.Fprintf(&b, " /p%d", i)
	}
	fmt.Println(b.String())

	for i, r := range "añ" {
		fmt.Printf("byte %d = %c\n", i, r)
	}
}
§§§

Output:

§§§text
bytes: 11, runes: 6
trimmed: Go 👋
fields: [the quick brown] 3
upper: QUIET
paths: /p1 /p2 /p3
byte 0 = a
byte 1 = ñ
§Worth noting: the emoji §👋§ is 4 bytes and 1 rune, and §ñ§ is 2 bytes and 1 rune. The
§range§ loop correctly reports §ñ§ at byte position 1 with no half-characters.

## The rules to remember

* A string is bytes. §len§ counts bytes, and §s[i]§ is a §byte§.
* A §rune§ is one character. Convert with §[]rune(s)§ when you need to count or index
  characters.
* Use §range§ to walk text. Use an index loop only for ASCII.
* §strings.Builder§ for repeated concatenation.
* §strconv§ to move between strings and numbers.
`,
		Quiz: []Question{
			{
				ID:     "m1t13q1",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

func main() {
	s := "日本"
	fmt.Println(len(s))
	fmt.Println(len([]rune(s)))
	fmt.Println(s[0])
}`,
				Options: []Option{
					{Text: "§6§, §2§ and §230§", Correct: true, Rationale: "Each character is 3 bytes in UTF-8, so 6 bytes and 2 runes; §s[0]§ is the first byte, 230."},
					{Text: "§2§, §2§ and §日§", Correct: false, Rationale: "§len§ on a string counts bytes, and §s[0]§ is a §byte§, not a character."},
					{Text: "§6§, §6§ and §230§", Correct: false, Rationale: "§[]rune§ conversion yields characters, not bytes."},
					{Text: "It does not compile.", Correct: false, Rationale: "All three lines are valid."},
				},
				Explanation: "§len(string)§ = bytes. §len([]rune(string))§ = characters. §s[i]§ = a byte.",
			},
			{
				ID:     "m1t13q2",
				Prompt: "What is printed?",
				Code: `package main

import "fmt"

func main() {
	for i, r := range "añ" {
		fmt.Print(i, ":", string(r), " ")
	}
}`,
				Options: []Option{
					{Text: "§0:a 1:ñ §", Correct: true, Rationale: "§a§ is one byte at index 0; §ñ§ is two bytes and starts at byte offset 1."},
					{Text: "§0:a 1:ñ 2:? §", Correct: false, Rationale: "§range§ decodes whole runes, so it never yields half a character."},
					{Text: "§0:a 2:ñ §", Correct: false, Rationale: "§range§ reports the *byte offset*, which is 1 for §ñ§, not 2."},
					{Text: "It does not compile: §string(r)§ is invalid.", Correct: false, Rationale: "§string(r)§ converts a rune to a one-character string."},
				},
				Explanation: "§range§ over a string yields byte offsets and whole runes — the safe way to walk text.",
			},
			{
				ID:     "m1t13q3",
				Prompt: "Which statements about strings are true? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "Strings are immutable; you cannot assign to §s[0]§.", Correct: true, Rationale: "Every \"modification\" builds a new string."},
					{Text: "§s += x§ in a loop is quadratic in the number of iterations.", Correct: true, Rationale: "Each append copies the whole string so far; §strings.Builder§ avoids it."},
					{Text: "§strconv.Atoi§ returns an error for §\"banana\"§.", Correct: true, Rationale: "Parsing failures are reported with the §error§ idiom, not a panic."},
					{Text: "Two strings can be compared with §==§.", Correct: true, Rationale: "Strings are comparable, unlike slices."},
					{Text: "§len(s)§ counts characters.", Correct: false, Rationale: "It counts bytes. Use §utf8.RuneCountInString§ or §len([]rune(s))§."},
				},
				Explanation: "Strings are immutable, comparable byte sequences. Length is bytes; characters need decoding.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t13e1",
				Title: "Warm-up: count bytes and characters",
				Prompt: `§§§go
func ByteLen(s string) int
func RuneLen(s string) int
§§§

* §ByteLen§ is §len(s)§ — the number of bytes.
* §RuneLen§ counts **characters**, so §RuneLen("héllo")§ is §5§ and §RuneLen("日本")§ is §2§.

You can write §RuneLen§ as §len([]rune(s))§, or import §"unicode/utf8"§ and use
§utf8.RuneCountInString(s)§, which does not allocate. Both are accepted by the test.
`,
				Starter: `package main

import "fmt"

// ByteLen returns the number of bytes in s.
func ByteLen(s string) int {
	// TODO
	return 0
}

// RuneLen returns the number of characters (runes) in s.
func RuneLen(s string) int {
	// TODO
	return 0
}

func main() {
	fmt.Println(ByteLen("héllo"), RuneLen("héllo"))
	fmt.Println(ByteLen("日本"), RuneLen("日本"))
}
`,
				Tests: []File{{
					Name: "len_test.go",
					Content: `package main

import "testing"

func TestByteLen(t *testing.T) {
	cases := map[string]int{"": 0, "abc": 3, "héllo": 6, "日本": 6, "👋": 4}
	for in, want := range cases {
		if got := ByteLen(in); got != want {
			t.Errorf("ByteLen(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestRuneLen(t *testing.T) {
	cases := map[string]int{"": 0, "abc": 3, "héllo": 5, "日本": 2, "👋": 1}
	for in, want := range cases {
		if got := RuneLen(in); got != want {
			t.Errorf("RuneLen(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestByteAndRuneLenTable(t *testing.T) {
	cases := []struct {
		in    string
		bytes int
		runes int
	}{
		{"", 0, 0},
		{"a", 1, 1},
		{"abc", 3, 3},
		{"héllo", 6, 5},
		{"日本", 6, 2},
		{"👋", 4, 1},
		{"a👋b", 6, 3},
		{"aé日👋", 10, 4},
		{"\u00e9", 2, 1},
		{"e\u0301", 3, 2},
		{"a\nb", 3, 3},
		{" ", 1, 1},
		{"  \t", 3, 3},
		{"\U0001F1EC\U0001F1E7", 8, 2},
		{"\xff", 1, 1},
	}
	for _, c := range cases {
		if got := ByteLen(c.in); got != c.bytes {
			t.Errorf("ByteLen(%q) = %d, want %d", c.in, got, c.bytes)
		}
		if got := RuneLen(c.in); got != c.runes {
			t.Errorf("RuneLen(%q) = %d, want %d", c.in, got, c.runes)
		}
	}
}
`,
				}},
				Hints: []string{
					"§ByteLen§ is literally §return len(s)§.",
					"§RuneLen§: §return len([]rune(s))§.",
					"Or import §\"unicode/utf8\"§ and return §utf8.RuneCountInString(s)§.",
				},
				Solution: `package main

import (
	"fmt"
	"unicode/utf8"
)

func ByteLen(s string) int {
	return len(s)
}

func RuneLen(s string) int {
	return utf8.RuneCountInString(s)
}

func main() {
	fmt.Println(ByteLen("héllo"), RuneLen("héllo"))
	fmt.Println(ByteLen("日本"), RuneLen("日本"))
}
`,
			},
			{
				ID:    "m1t13e2",
				Title: "Reverse a string, characters first",
				Prompt: `§§§go
func Reverse(s string) string
func ReverseASCII(s string) string
§§§

* §Reverse§ reverses the **characters**, so §Reverse("héllo")§ is §"olléh"§.
* §ReverseASCII§ reverses byte by byte. It is *wrong* for non-ASCII input but fine for
  plain English, and writing both shows you exactly where the difference shows up.

Both must handle the empty string.

The trick for the character version: convert to §[]rune§, swap from both ends walking
inwards, then convert back with §string(runes)§. The byte version is the same idea with
§[]byte(s)§ instead — and §string(b)§ converts a byte slice back to a string. It works
on raw bytes, which is exactly why it mangles multi-byte characters.
`,
				Starter: `package main

import "fmt"

// Reverse reverses the characters of s.
func Reverse(s string) string {
	// TODO
	return ""
}

// ReverseASCII reverses the bytes of s (only correct for ASCII input).
func ReverseASCII(s string) string {
	// TODO
	return ""
}

func main() {
	fmt.Println(Reverse("héllo"))
	fmt.Println(ReverseASCII("hello"))
}
`,
				Tests: []File{{
					Name: "reverse_test.go",
					Content: `package main

import "testing"

func TestReverse(t *testing.T) {
	cases := map[string]string{
		"":         "",
		"a":        "a",
		"ab":       "ba",
		"aba":      "aba",
		"aab":      "baa",
		"abcdef":   "fedcba",
		"hello":    "olleh",
		"héllo":    "olléh",
		"日本":       "本日",
		"ab👋":      "👋ba",
		"a b":      "b a",
		"12321":    "12321",
		"  ":       "  ",
		"aé日":      "日éa",
		"👋👋":       "👋👋",
		"e\u0301x": "x\u0301e",
	}
	for in, want := range cases {
		if got := Reverse(in); got != want {
			t.Errorf("Reverse(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestReverseASCII(t *testing.T) {
	for in, want := range map[string]string{
		"":       "",
		"a":      "a",
		"ab":     "ba",
		"abba":   "abba",
		"abcdef": "fedcba",
		"hello":  "olleh",
		"12345":  "54321",
		"a b":    "b a",
		"  ":     "  ",
	} {
		if got := ReverseASCII(in); got != want {
			t.Errorf("ReverseASCII(%q) = %q, want %q", in, got, want)
		}
	}
}
`,
				}},
				Hints: []string{
					"§runes := []rune(s)§, then swap with §for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 { runes[i], runes[j] = runes[j], runes[i] }§.",
					"Finish with §return string(runes)§.",
					"§ReverseASCII§ follows the same swap loop over §[]byte(s)§ instead of runes.",
				},
				Solution: `package main

import "fmt"

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func ReverseASCII(s string) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}

func main() {
	fmt.Println(Reverse("héllo"))
	fmt.Println(ReverseASCII("hello"))
}
`,
			},
			{
				ID:    "m1t13e3",
				Title: "Word statistics with strings and a Builder",
				Prompt: `§§§go
func Normalize(s string) string
func WordStats(text string) []string
func JoinWith(words []string, sep string) string
§§§

* §Normalize§ lower-cases the text, keeps only letters, digits and spaces (drop every
  other character), and trims the result. §Normalize("Hello, World!")§ → §"hello world"§.
* §WordStats§ splits the normalized text on any whitespace and returns
  §"<word>:<count>"§ entries, **sorted alphabetically by word**. With no words, an empty
  slice.
  * §WordStats("the cat and the hat")§ → §["and:1" "cat:1" "hat:1" "the:2"]§
* §JoinWith§ joins words with a separator, built with a §strings.Builder§. The test
  checks the exact output, including that there is no trailing separator.

§strings.Fields§ handles the splitting; §slices.Sort§ handles the ordering.

Three tools you have not met in full yet, in brief:

* The §unicode§ package classifies runes: §unicode.IsLetter(r)§ and §unicode.IsDigit(r)§
  are the two predicates §Normalize§ wants.
* §strings.Builder§ has both §WriteString(s)§ and §WriteRune(r)§, each appending to the
  same accumulating buffer; end with §b.String()§.
* §WordStats§ needs a count per word. The tool for that is a §map[string]int§ — a lookup
  table from key to value. §counts := map[string]int{}§ creates an empty one, and
  §counts[w]++§ adds one to the tally for §w§ (a key that is not there yet reads as §0§
  first). Maps get their own topic next; this is all you need now.
`,
				Starter: `package main

import (
	"fmt"
	"strings"
)

// Normalize lower-cases s and removes everything except letters, digits and spaces.
func Normalize(s string) string {
	// TODO
	return ""
}

// WordStats returns "word:count" entries sorted by word.
func WordStats(text string) []string {
	// TODO
	return nil
}

// JoinWith joins words using sep, built with a strings.Builder.
func JoinWith(words []string, sep string) string {
	// TODO
	return ""
}

func main() {
	fmt.Println(Normalize("Hello, World!"))
	fmt.Println(WordStats("the cat and the hat"))
	fmt.Println(JoinWith([]string{"a", "b", "c"}, ", "))

	// strings is imported for you.
	_ = strings.Fields
}
`,
				Tests: []File{{
					Name: "words_test.go",
					Content: `package main

import "testing"

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"Hello, World!":     "hello world",
		"  Go  ":            "go",
		"a-b_c!":            "abc",
		"":                  "",
		"...":               "",
		"!!!":               "",
		"123 ok":            "123 ok",
		"  Hello,   World!": "hello   world",
		"a\tb\nc":           "abc",
		"Über Café":         "über café",
		"a  b":              "a  b",
		"Hello123":          "hello123",
		" \t\n ":            "",
		"日本 語":             "日本 語",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWordStats(t *testing.T) {
	got := WordStats("the cat and the hat")
	want := []string{"and:1", "cat:1", "hat:1", "the:2"}
	if !sameStrings(got, want) {
		t.Fatalf("WordStats = %v, want %v", got, want)
	}
	if len(WordStats("")) != 0 {
		t.Errorf("WordStats(\"\") = %v, want empty", WordStats(""))
	}
	if len(WordStats(",,,")) != 0 {
		t.Errorf("WordStats(punctuation) = %v, want empty", WordStats(",,,"))
	}
	got = WordStats("b a b")
	if !sameStrings(got, []string{"a:1", "b:2"}) {
		t.Fatalf("WordStats = %v, want [a:1 b:2]", got)
	}
}

func TestJoinWith(t *testing.T) {
	if got := JoinWith([]string{"a", "b", "c"}, ", "); got != "a, b, c" {
		t.Errorf("JoinWith = %q", got)
	}
	if got := JoinWith([]string{"only"}, ", "); got != "only" {
		t.Errorf("JoinWith(single) = %q, want no separator", got)
	}
	if got := JoinWith(nil, ", "); got != "" {
		t.Errorf("JoinWith(nil) = %q, want empty", got)
	}
}

func TestWordStatsTable(t *testing.T) {
	cases := []struct {
		text string
		want []string
	}{
		{"the cat and the hat", []string{"and:1", "cat:1", "hat:1", "the:2"}},
		{"", []string{}},
		{"   \t\n ", []string{}},
		{",,,", []string{}},
		{"one", []string{"one:1"}},
		{"go go GO", []string{"go:3"}},
		{"b a c b a", []string{"a:2", "b:2", "c:1"}},
		{"  spaced   out  ", []string{"out:1", "spaced:1"}},
		{"hello, hello!", []string{"hello:2"}},
		{"café café", []string{"café:2"}},
		{"z z z z", []string{"z:4"}},
		{"a1 b2 a1", []string{"a1:2", "b2:1"}},
		{"bbb aaa", []string{"aaa:1", "bbb:1"}},
	}
	for _, c := range cases {
		got := WordStats(c.text)
		if !sameStrings(got, c.want) {
			t.Errorf("WordStats(%q) = %v, want %v", c.text, got, c.want)
		}
	}
}

func TestJoinWithTable(t *testing.T) {
	cases := []struct {
		words []string
		sep   string
		want  string
	}{
		{nil, ", ", ""},
		{[]string{}, ", ", ""},
		{[]string{"only"}, ", ", "only"},
		{[]string{"a", "b"}, "-", "a-b"},
		{[]string{"a", "b", "c", "d"}, "-", "a-b-c-d"},
		{[]string{"x", "y"}, "", "xy"},
		{[]string{"solo"}, "--", "solo"},
		{[]string{"hé", "日"}, " + ", "hé + 日"},
	}
	for _, c := range cases {
		if got := JoinWith(c.words, c.sep); got != c.want {
			t.Errorf("JoinWith(%v, %q) = %q, want %q", c.words, c.sep, got, c.want)
		}
	}
}
`,
				}},
				Hints: []string{
					"§Normalize§: lower-case, loop over the runes keeping the ones the §unicode§ predicates accept, then §strings.TrimSpace§ the built result.",
					"§WordStats§: §words := strings.Fields(Normalize(text))§, tally each word in a §map[string]int§, format §fmt.Sprintf(\"%s:%d\", w, n)§ per entry, then §slices.Sort§.",
					"§JoinWith§: loop and write §sep§ *before* every word except the first, or use the 'skip index 0' trick.",
				},
				Solution: `package main

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

func Normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func WordStats(text string) []string {
	counts := map[string]int{}
	for _, w := range strings.Fields(Normalize(text)) {
		counts[w]++
	}
	out := make([]string, 0, len(counts))
	for w, n := range counts {
		out = append(out, fmt.Sprintf("%s:%d", w, n))
	}
	slices.Sort(out)
	return out
}

func JoinWith(words []string, sep string) string {
	var b strings.Builder
	for i, w := range words {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(w)
	}
	return b.String()
}

func main() {
	fmt.Println(Normalize("Hello, World!"))
	fmt.Println(WordStats("the cat and the hat"))
	fmt.Println(JoinWith([]string{"a", "b", "c"}, ", "))
}
`,
			},
		},
	}
}

func m1t14() Topic {
	return Topic{
		ID:      "m1t14",
		Title:   "Maps",
		Summary: "Looking things up by key, and the comma-ok idiom.",
		Theory: `
A **map** is a lookup table: you give it a key, it gives you a value. In Java it is a
§HashMap§, in Python a §dict§, in JavaScript a plain object. In Go it is §map[K]V§ —
read as *"a map from §K§ to §V§"*.

## Creating one

§§§go
var m1 map[string]int          // nil map: reads work, WRITES PANIC
m2 := map[string]int{}         // empty but usable
m3 := map[string]int{"a": 1}   // literal with entries
m4 := make(map[string]int)     // make, like slices
§§§

That first line is a trap worth carving into memory:

§§§go
var counts map[string]int
fmt.Println(counts["missing"]) // 0    — reading a nil map is fine
fmt.Println(len(counts))       // 0    — so is len
// counts["x"] = 1             // PANIC: assignment to entry in nil map
§§§

**A nil map is readable but not writable.** Use §map[…]{}§ or §make(…)§ when you intend
to put things in it.

## Reading, writing, deleting

§§§go
ages := map[string]int{"ada": 36}

ages["bo"] = 21          // write (or overwrite)
fmt.Println(ages["ada"]) // 36
fmt.Println(ages["zed"]) // 0   — no error, just the zero value
delete(ages, "bo")       // remove a key (no-op if absent)
fmt.Println(len(ages))   // 1
§§§

Note the third line. Reading a key that does not exist does **not** fail — it hands you
the zero value for the value type. That is convenient, and it is also ambiguous:

§§§go
m := map[string]int{"a": 0}
fmt.Println(m["a"])       // 0
fmt.Println(m["nope"])    // 0    — is that "zero" or "missing"? You cannot tell!
§§§

## The comma-ok idiom

To tell "the value is zero" apart from "the key is absent", use the two-value form:

§§§go
v, ok := m["a"]
fmt.Println(v, ok) // 0 true   — present, and its value is 0

v, ok = m["nope"]
fmt.Println(v, ok) // 0 false  — not present
§§§

This is the *same pattern* you have already met three times (the §ok§ from §FirstLast§,
from §MinMax§, from checking an error), and it is the idiomatic way to ask "is this key
here?".

You will also see the throwaway version, which is very common:

§§§go
if _, ok := m[key]; !ok {
	fmt.Println("not found")
}
§§§

## The counting idiom

Because a missing key reads as zero, counting is a one-liner. This pattern is worth
memorising:

§§§go
counts := map[string]int{}
for _, w := range words {
	counts[w]++ // read 0 if absent, then store 1
}
§§§

§counts[w]++§ desugars to §counts[w] = counts[w] + 1§. The read gives 0 for a new key,
so the write stores 1. No existence check needed.

## Iteration order is deliberately random

§§§go
for key, value := range m {
	fmt.Println(key, value)
}
§§§

Go **randomises** the order on every run. This is on purpose: it stops code from
accidentally depending on an implementation detail. If you need determinism, collect the
keys and sort them:

§§§go
keys := make([]string, 0, len(ages))
for k := range ages {
	keys = append(keys, k)
}
slices.Sort(keys)
for _, k := range keys {
	fmt.Println(k, ages[k])
}
§§§

(Since Go 1.23 you can write §slices.Sorted(maps.Keys(ages))§.)

## Sets, and grouping

Go has no §Set§ type. The idiom is a map with an empty-struct value, because
§struct{}§ occupies zero bytes:

§§§go
seen := map[string]struct{}{}
seen["a"] = struct{}{}
if _, ok := seen["a"]; ok { /* already seen */ }
§§§

A §map[string]bool§ is equally common and slightly more readable.

Grouping — several values per key — relies on the fact that appending to a nil slice
works:

§§§go
byLength := map[int][]string{}
for _, w := range []string{"a", "bb", "cc", "ddd"} {
	byLength[len(w)] = append(byLength[len(w)], w)
}
fmt.Println(byLength) // map[1:[a] 2:[bb cc] 3:[ddd]]
§§§

That works even for a key seen for the first time, because §byLength[len(w)]§ reads as
§nil§ and §append(nil, x)§ gives a one-element slice.

## Keys must be comparable

Slices, maps and functions cannot be keys. Everything else can:

§§§go
map[string]int{}     // fine
map[int]string{}     // fine
map[[2]int]string{}  // arrays work
map[Point]string{}   // structs of comparable fields work
// map[[]int]string{} // COMPILE ERROR: slices are not comparable
§§§

Struct keys are a neat way to build a composite key:

§§§go
type Coord struct{ X, Y int }

grid := map[Coord]string{}
grid[Coord{0, 0}] = "origin"
§§§

## Maps are references (like slices)

Assigning a map copies the descriptor, not the table — both names see the same data:

§§§go
a := map[string]int{"x": 1}
b := a
b["x"] = 2
fmt.Println(a["x"]) // 2
§§§

To copy a map properly, loop over it (§maps.Clone(m)§ does this for you):

§§§go
copyOf := maps.Clone(a)
§§§

## A complete example

§§§go
package main

import (
	"fmt"
	"slices"
)

func main() {
	text := []string{"go", "is", "fun", "go", "is", "fast"}
	counts := map[string]int{}
	for _, w := range text {
		counts[w]++
	}

	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for _, k := range keys {
		fmt.Printf("%-5s %d\n", k, counts[k])
	}

	if _, ok := counts["rust"]; !ok {
		fmt.Println("rust: absent (comma-ok says so)")
	}
}
§§§

Output:

§§§text
fast  1
fun   1
go    2
is    2
rust: absent (comma-ok says so)
§§§

## The one thing people trip over

A map is **not** safe for concurrent use. Two goroutines writing to the same map is a
data race, and Go's runtime detects it and crashes with §fatal error: concurrent map
writes§. We solve that properly in Module 2 with §sync.RWMutex§. For now: one
goroutine, one map.
`,
		Quiz: []Question{
			{
				ID:     "m1t14q1",
				Prompt: "Which line panics at run time?",
				Code: `var m map[string]int
_ = m["a"]
_, _ = m["a"]
m["a"] = 1
delete(m, "a")
_ = len(m)`,
				Options: []Option{
					{Text: "§m[\"a\"] = 1§", Correct: true, Rationale: "Writing to a nil map panics with: assignment to entry in nil map."},
					{Text: "§_ = m[\"a\"]§", Correct: false, Rationale: "Reading a nil map is safe and returns the zero value."},
					{Text: "§delete(m, \"a\")§", Correct: false, Rationale: "§delete§ on a nil map is a no-op."},
					{Text: "§_ = len(m)§", Correct: false, Rationale: "§len§ of a nil map is 0."},
				},
				Explanation: "Only writes panic on a nil map. Initialise with §make§ or a literal before storing anything.",
			},
			{
				ID:     "m1t14q2",
				Prompt: "What is the difference between these two lookups?",
				Code: `m := map[string]int{"a": 0}

v := m["a"]
v2, ok := m["a"]`,
				Options: []Option{
					{Text: "§v§ is §0§; §v2§ is §0§ and §ok§ is §true§. The two-value form tells you the key exists.", Correct: true, Rationale: "Only the comma-ok form distinguishes \"present and zero\" from \"absent\"."},
					{Text: "§v§ is §0§ and §v2§ is also §0§, but §ok§ is §false§ because the value is zero.", Correct: false, Rationale: "§ok§ is about key presence, not the value."},
					{Text: "They are identical in every respect.", Correct: false, Rationale: "The second form gives you the presence information."},
					{Text: "The second line does not compile.", Correct: false, Rationale: "That is exactly what the comma-ok idiom is for."},
				},
				Explanation: "Reading a missing key yields the zero value with no error — so you need the §ok§ flag to know whether it was really there.",
			},
			{
				ID:     "m1t14q3",
				Prompt: "Which types may be used as map keys? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "§string§", Correct: true, Rationale: "Comparable."},
					{Text: "§[2]int§", Correct: true, Rationale: "Arrays of comparable types are comparable."},
					{Text: "§[]int§", Correct: false, Rationale: "Slices have no §==§, so they cannot be keys."},
					{Text: "§struct{ A int }§", Correct: true, Rationale: "Structs whose fields are all comparable are comparable."},
					{Text: "§func()§", Correct: false, Rationale: "Functions compare only against nil."},
					{Text: "§*Point§", Correct: true, Rationale: "Pointers are comparable — but the key is the address, not the pointed-to value."},
				},
				Explanation: "Keys need §==§. Slices, maps and functions are the exceptions.",
			},
			{
				ID:     "m1t14q4",
				Prompt: "Why does §counts[w]++§ work on a map that has never seen §w§?",
				Options: []Option{
					{Text: "Because reading a missing key returns the zero value, and §counts[w]++§ is really §counts[w] = counts[w] + 1§.", Correct: true, Rationale: "The read yields 0, and the write stores 1."},
					{Text: "Because §++§ knows about maps and inserts the key.", Correct: false, Rationale: "§++§ is a plain read-modify-write."},
					{Text: "Because Go initialises missing keys to 1.", Correct: false, Rationale: "To 0 — the zero value of §int§."},
					{Text: "It is not legal: you must check §ok§ first.", Correct: false, Rationale: "It is legal and extremely common on a non-nil map."},
				},
				Explanation: "The zero value is what makes the counting idiom a one-liner.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m1t14e1",
				Title: "Warm-up: count the words",
				Prompt: `§§§go
func WordCount(text string) map[string]int
func Lookup(counts map[string]int, word string) (int, bool)
§§§

* §WordCount§ splits §text§ on whitespace and counts how many times each word appears.
  For empty text, return an **empty, non-nil** map (so the caller can write to it).
* §Lookup§ returns the count for §word§ and whether the word was there at all — the
  comma-ok idiom, wrapped in a function.

§strings.Fields§ does the splitting.
`,
				Starter: `package main

import (
	"fmt"
	"strings"
)

// WordCount counts occurrences of each whitespace-separated word.
func WordCount(text string) map[string]int {
	// TODO
	return map[string]int{}
}

// Lookup returns the count for word, and whether it was present.
func Lookup(counts map[string]int, word string) (int, bool) {
	// TODO
	return 0, false
}

func main() {
	counts := WordCount("the cat and the hat")
	fmt.Println(counts)
	fmt.Println(Lookup(counts, "the"))
	fmt.Println(Lookup(counts, "dog"))
	_ = strings.Fields
}
`,
				Tests: []File{{
					Name: "wordcount_test.go",
					Content: `package main

import "testing"

func TestWordCount(t *testing.T) {
	got := WordCount("the cat and the hat")
	want := map[string]int{"the": 2, "cat": 1, "and": 1, "hat": 1}
	if len(got) != len(want) {
		t.Fatalf("WordCount = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("WordCount[%q] = %d, want %d (full map %v)", k, got[k], v, got)
		}
	}
}

func TestWordCountEmpty(t *testing.T) {
	got := WordCount("")
	if got == nil {
		t.Fatal("WordCount(\"\") returned nil; return an empty map")
	}
	if len(got) != 0 {
		t.Fatalf("WordCount(\"\") = %v, want empty", got)
	}
	got["x"] = 1 // must not panic: the map has to be usable
}

func TestWordCountWhitespace(t *testing.T) {
	got := WordCount("  a\tb\n a  ")
	if got["a"] != 2 || got["b"] != 1 {
		t.Fatalf("WordCount = %v, want a:2 b:1", got)
	}
}

func TestLookup(t *testing.T) {
	counts := WordCount("the cat and the hat")

	if n, ok := Lookup(counts, "the"); !ok || n != 2 {
		t.Errorf("Lookup(the) = %d, %v; want 2, true", n, ok)
	}
	if n, ok := Lookup(counts, "dog"); ok || n != 0 {
		t.Errorf("Lookup(dog) = %d, %v; want 0, false", n, ok)
	}
}

func TestWordCountTable(t *testing.T) {
	cases := []struct {
		text string
		want map[string]int
	}{
		{"", map[string]int{}},
		{"   \t\n ", map[string]int{}},
		{"one", map[string]int{"one": 1}},
		{"a a a", map[string]int{"a": 3}},
		{"go go GO", map[string]int{"go": 2, "GO": 1}},
		{"  a\tb\n a  ", map[string]int{"a": 2, "b": 1}},
		{"one two one two three", map[string]int{"one": 2, "two": 2, "three": 1}},
		{"héllo héllo 日本", map[string]int{"héllo": 2, "日本": 1}},
		{"same same same same", map[string]int{"same": 4}},
	}
	for _, c := range cases {
		got := WordCount(c.text)
		if got == nil {
			t.Errorf("WordCount(%q) returned nil; want a non-nil map", c.text)
			continue
		}
		if len(got) != len(c.want) {
			t.Errorf("WordCount(%q) = %v, want %v", c.text, got, c.want)
			continue
		}
		for word, want := range c.want {
			if got[word] != want {
				t.Errorf("WordCount(%q)[%q] = %d, want %d (full map %v)",
					c.text, word, got[word], want, got)
			}
		}
	}
}

func TestLookupTable(t *testing.T) {
	counts := map[string]int{"zero": 0, "two": 2, "": 1}
	cases := []struct {
		word string
		n    int
		ok   bool
	}{
		{"zero", 0, true},
		{"two", 2, true},
		{"", 1, true},
		{"missing", 0, false},
		{"Zero", 0, false},
	}
	for _, c := range cases {
		if n, ok := Lookup(counts, c.word); n != c.n || ok != c.ok {
			t.Errorf("Lookup(%q) = %d, %v; want %d, %v", c.word, n, ok, c.n, c.ok)
		}
	}
	if len(counts) != 3 {
		t.Errorf("Lookup modified the map: %v", counts)
	}

	// Reading from a nil map is safe and always reports absent.
	if n, ok := Lookup(nil, "anything"); ok || n != 0 {
		t.Errorf("Lookup(nil, anything) = %d, %v; want 0, false", n, ok)
	}
}
`,
				}},
				Hints: []string{
					"Start with §counts := map[string]int{}§ — a literal, so it is never nil.",
					"§for _, w := range strings.Fields(text) { counts[w]++ }§",
					"§Lookup§ is two lines: §v, ok := counts[word]§ then §return v, ok§.",
				},
				Solution: `package main

import (
	"fmt"
	"strings"
)

func WordCount(text string) map[string]int {
	counts := map[string]int{}
	for _, w := range strings.Fields(text) {
		counts[w]++
	}
	return counts
}

func Lookup(counts map[string]int, word string) (int, bool) {
	v, ok := counts[word]
	return v, ok
}

func main() {
	counts := WordCount("the cat and the hat")
	fmt.Println(counts)
	fmt.Println(Lookup(counts, "the"))
	fmt.Println(Lookup(counts, "dog"))
}
`,
			},
			{
				ID:    "m1t14e2",
				Title: "Grouping and inverting",
				Prompt: `§§§go
func GroupByLength(words []string) map[int][]string
func Invert(counts map[string]int) map[int][]string
func SortedKeys(m map[string]int) []string
§§§

* §GroupByLength§ groups words by their length:
  §GroupByLength(["a" "bb" "cc" "ddd"])§ → §map[1:[a] 2:[bb cc] 3:[ddd]]§.
* §Invert§ turns a word→count map into a count→words map. Because Go map iteration order
  is random, the test sorts the inner slices before comparing:
  §Invert({"the":2 "cat":1})§ → §map[1:[cat] 2:[the]]§.
* §SortedKeys§ returns the keys of a map in alphabetical order — the fix for the random
  iteration order.

For §GroupByLength§, remember that appending to a value you have never seen works,
because §m[k]§ reads as §nil§.
`,
				Starter: `package main

import (
	"fmt"
	"slices"
)

// GroupByLength groups words by their length, preserving input order within a group.
func GroupByLength(words []string) map[int][]string {
	// TODO
	return map[int][]string{}
}

// Invert maps each count back to the words with that count.
func Invert(counts map[string]int) map[int][]string {
	// TODO
	return map[int][]string{}
}

// SortedKeys returns the map's keys in alphabetical order.
func SortedKeys(m map[string]int) []string {
	// TODO
	return nil
}

func main() {
	fmt.Println(GroupByLength([]string{"a", "bb", "cc", "ddd"}))
	fmt.Println(Invert(map[string]int{"the": 2, "cat": 1}))
	fmt.Println(SortedKeys(map[string]int{"b": 1, "a": 2}))

	keys := SortedKeys(map[string]int{"b": 1, "a": 2})
	slices.Sort(keys)
	fmt.Println(keys)
}
`,
				Tests: []File{{
					Name: "group_test.go",
					Content: `package main

import (
	"slices"
	"testing"
)

func TestGroupByLength(t *testing.T) {
	got := GroupByLength([]string{"a", "bb", "cc", "ddd"})
	if len(got) != 3 {
		t.Fatalf("GroupByLength returned %d groups, want 3: %v", len(got), got)
	}
	if !slices.Equal(got[1], []string{"a"}) {
		t.Errorf("group 1 = %v, want [a]", got[1])
	}
	if !slices.Equal(got[2], []string{"bb", "cc"}) {
		t.Errorf("group 2 = %v, want [bb cc] (input order)", got[2])
	}
	if !slices.Equal(got[3], []string{"ddd"}) {
		t.Errorf("group 3 = %v, want [ddd]", got[3])
	}
	if len(GroupByLength(nil)) != 0 {
		t.Error("GroupByLength(nil) should be empty")
	}
}

func TestInvert(t *testing.T) {
	counts := map[string]int{"the": 2, "cat": 1, "hat": 1}
	got := Invert(counts)

	for _, group := range got {
		slices.Sort(group)
	}
	if !slices.Equal(got[1], []string{"cat", "hat"}) {
		t.Errorf("Invert group 1 = %v, want [cat hat]", got[1])
	}
	if !slices.Equal(got[2], []string{"the"}) {
		t.Errorf("Invert group 2 = %v, want [the]", got[2])
	}
}

func TestInvertNonDestructive(t *testing.T) {
	counts := map[string]int{"a": 1}
	_ = Invert(counts)
	if len(counts) != 1 || counts["a"] != 1 {
		t.Errorf("Invert modified its input: %v", counts)
	}
}

func TestSortedKeys(t *testing.T) {
	got := SortedKeys(map[string]int{"b": 1, "a": 2, "c": 3})
	want := []string{"a", "b", "c"}
	if !slices.Equal(got, want) {
		t.Fatalf("SortedKeys = %v, want %v", got, want)
	}
	if len(SortedKeys(nil)) != 0 {
		t.Error("SortedKeys(nil) should be empty")
	}
}

func TestGroupByLengthTable(t *testing.T) {
	cases := []struct {
		name  string
		words []string
		want  map[int][]string
	}{
		{"nil", nil, map[int][]string{}},
		{"empty", []string{}, map[int][]string{}},
		{"single", []string{"x"}, map[int][]string{1: []string{"x"}}},
		{"same length", []string{"aa", "bb", "cc"}, map[int][]string{2: []string{"aa", "bb", "cc"}}},
		{"mixed", []string{"a", "bb", "c", "dd", "eee"}, map[int][]string{
			1: []string{"a", "c"},
			2: []string{"bb", "dd"},
			3: []string{"eee"},
		}},
		{"unicode byte lengths", []string{"é", "日", "ab"}, map[int][]string{
			2: []string{"é", "ab"},
			3: []string{"日"},
		}},
		{"duplicates", []string{"a", "a", "bb"}, map[int][]string{
			1: []string{"a", "a"},
			2: []string{"bb"},
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := GroupByLength(c.words)
			if got == nil {
				t.Fatal("GroupByLength returned nil; want a non-nil map")
			}
			if len(got) != len(c.want) {
				t.Fatalf("GroupByLength(%v) = %v, want %v", c.words, got, c.want)
			}
			for length, want := range c.want {
				if !slices.Equal(got[length], want) {
					t.Errorf("group %d = %v, want %v", length, got[length], want)
				}
			}
		})
	}
}

func TestInvertTable(t *testing.T) {
	cases := []struct {
		name   string
		counts map[string]int
		want   map[int][]string
	}{
		{"nil", nil, map[int][]string{}},
		{"empty", map[string]int{}, map[int][]string{}},
		{"unique", map[string]int{"a": 1, "b": 2, "c": 3}, map[int][]string{
			1: []string{"a"},
			2: []string{"b"},
			3: []string{"c"},
		}},
		{"shared counts", map[string]int{"a": 2, "b": 2, "c": 1}, map[int][]string{
			1: []string{"c"},
			2: []string{"a", "b"},
		}},
		{"zero count", map[string]int{"z": 0, "a": 1}, map[int][]string{
			0: []string{"z"},
			1: []string{"a"},
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Invert(c.counts)
			if got == nil {
				t.Fatal("Invert returned nil; want a non-nil map")
			}
			if len(got) != len(c.want) {
				t.Fatalf("Invert(%v) = %v, want %v", c.counts, got, c.want)
			}
			for _, group := range got {
				slices.Sort(group)
			}
			for count, want := range c.want {
				if !slices.Equal(got[count], want) {
					t.Errorf("Invert group %d = %v, want %v", count, got[count], want)
				}
			}
		})
	}
}

func TestSortedKeysTable(t *testing.T) {
	cases := []struct {
		m    map[string]int
		want []string
	}{
		{nil, []string{}},
		{map[string]int{}, []string{}},
		{map[string]int{"a": 1}, []string{"a"}},
		{map[string]int{"b": 2, "a": 1, "c": 3}, []string{"a", "b", "c"}},
		{map[string]int{"c": 1, "b": 2, "a": 3}, []string{"a", "b", "c"}},
		{map[string]int{"10": 1, "2": 2, "1": 3}, []string{"1", "10", "2"}},
		{map[string]int{"é": 1, "z": 2, "a": 3}, []string{"a", "z", "é"}},
	}
	for _, c := range cases {
		if got := SortedKeys(c.m); !slices.Equal(got, c.want) {
			t.Errorf("SortedKeys(%v) = %v, want %v", c.m, got, c.want)
		}
	}
}
`,
				}},
				Hints: []string{
					"§GroupByLength§: §groups[len(w)] = append(groups[len(w)], w)§.",
					"§Invert§: loop over the map, and append the word into §out[count]§.",
					"§SortedKeys§: collect with §make([]string, 0, len(m))§, then §slices.Sort(keys)§.",
					"Return non-nil maps from both grouping functions so callers can write to them.",
				},
				Solution: `package main

import (
	"fmt"
	"slices"
)

func GroupByLength(words []string) map[int][]string {
	groups := map[int][]string{}
	for _, w := range words {
		groups[len(w)] = append(groups[len(w)], w)
	}
	return groups
}

func Invert(counts map[string]int) map[int][]string {
	out := map[int][]string{}
	for word, n := range counts {
		out[n] = append(out[n], word)
	}
	return out
}

func SortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func main() {
	fmt.Println(GroupByLength([]string{"a", "bb", "cc", "ddd"}))
	fmt.Println(Invert(map[string]int{"the": 2, "cat": 1}))
	fmt.Println(SortedKeys(map[string]int{"b": 1, "a": 2}))
}
`,
			},
			{
				ID:    "m1t14e3",
				Title: "A phone book with delete and update",
				Prompt: `§§§go
type PhoneBook struct {
	entries map[string]string
}

func NewPhoneBook() *PhoneBook
func (pb *PhoneBook) Set(name, number string)
func (pb *PhoneBook) Get(name string) (string, bool)
func (pb *PhoneBook) Delete(name string) bool
func (pb *PhoneBook) Names() []string
§§§

* §NewPhoneBook§ returns a ready-to-use, **non-nil** map — otherwise the first §Set§
  panics.
* §Set§ stores or overwrites a number.
* §Get§ returns the number and whether the name was there.
* §Delete§ removes an entry and reports whether anything was removed.
* §Names§ returns every name in alphabetical order.

Yes, these are *methods* on a struct — §func (pb *PhoneBook) Set(…)§. They get the full
treatment in the next topic; here they are just functions that happen to take the book
as a receiver instead of as a parameter. The §*PhoneBook§ receiver is what lets §Set§
modify the book.
`,
				Starter: `package main

import "fmt"

type PhoneBook struct {
	entries map[string]string
}

func NewPhoneBook() *PhoneBook {
	// TODO
	return nil
}

func (pb *PhoneBook) Set(name, number string) {
	// TODO
}

func (pb *PhoneBook) Get(name string) (string, bool) {
	// TODO
	return "", false
}

func (pb *PhoneBook) Delete(name string) bool {
	// TODO
	return false
}

func (pb *PhoneBook) Names() []string {
	// TODO
	return nil
}

func main() {
	pb := NewPhoneBook()
	pb.Set("ada", "555-0100")
	pb.Set("bo", "555-0101")
	fmt.Println(pb.Get("ada"))
	fmt.Println(pb.Delete("bo"), pb.Delete("zed"))
	fmt.Println(pb.Names())
}
`,
				Tests: []File{{
					Name: "phonebook_test.go",
					Content: `package main

import (
	"slices"
	"testing"
)

func TestNewPhoneBook(t *testing.T) {
	pb := NewPhoneBook()
	if pb == nil {
		t.Fatal("NewPhoneBook returned nil")
	}
	pb.Set("ada", "555-0100") // must not panic
	if got, ok := pb.Get("ada"); !ok || got != "555-0100" {
		t.Fatalf("Get after Set = %q, %v", got, ok)
	}
}

func TestSetOverwrites(t *testing.T) {
	pb := NewPhoneBook()
	pb.Set("ada", "111")
	pb.Set("ada", "222")
	if got, _ := pb.Get("ada"); got != "222" {
		t.Fatalf("Get = %q, want 222", got)
	}
}

func TestGetMissing(t *testing.T) {
	pb := NewPhoneBook()
	if got, ok := pb.Get("nobody"); ok || got != "" {
		t.Fatalf("Get(missing) = %q, %v; want \"\", false", got, ok)
	}
}

func TestDelete(t *testing.T) {
	pb := NewPhoneBook()
	pb.Set("ada", "111")

	if !pb.Delete("ada") {
		t.Error("Delete of an existing name returned false")
	}
	if pb.Delete("ada") {
		t.Error("Delete of a missing name returned true")
	}
	if _, ok := pb.Get("ada"); ok {
		t.Error("the entry survived deletion")
	}
}

func TestNames(t *testing.T) {
	pb := NewPhoneBook()
	pb.Set("cy", "3")
	pb.Set("ada", "1")
	pb.Set("bo", "2")

	got := pb.Names()
	want := []string{"ada", "bo", "cy"}
	if !slices.Equal(got, want) {
		t.Fatalf("Names = %v, want %v", got, want)
	}
	if len(NewPhoneBook().Names()) != 0 {
		t.Error("Names on an empty book should be empty")
	}
}

func TestSetOverwritesAndKeepsOneEntry(t *testing.T) {
	pb := NewPhoneBook()
	pb.Set("ada", "111")
	pb.Set("bo", "222")
	pb.Set("ada", "333")
	pb.Set("bo", "444")

	if got := pb.Names(); !slices.Equal(got, []string{"ada", "bo"}) {
		t.Fatalf("Names after overwrites = %v, want [ada bo]", got)
	}
	if got, ok := pb.Get("ada"); !ok || got != "333" {
		t.Errorf("Get(ada) = %q, %v; want 333, true", got, ok)
	}
	if got, ok := pb.Get("bo"); !ok || got != "444" {
		t.Errorf("Get(bo) = %q, %v; want 444, true", got, ok)
	}
}

func TestDeleteOnlyRemovesTarget(t *testing.T) {
	pb := NewPhoneBook()
	pb.Set("cy", "3")
	pb.Set("ada", "1")
	pb.Set("bo", "2")

	if !pb.Delete("bo") {
		t.Error("Delete(bo) = false, want true")
	}
	if _, ok := pb.Get("bo"); ok {
		t.Error("the deleted entry is still present")
	}
	if got := pb.Names(); !slices.Equal(got, []string{"ada", "cy"}) {
		t.Errorf("Names after deleting bo = %v, want [ada cy]", got)
	}
	if pb.Delete("bo") {
		t.Error("second Delete(bo) = true, want false")
	}
	if !pb.Delete("ada") || !pb.Delete("cy") {
		t.Error("deleting the remaining entries should return true")
	}
	if got := pb.Names(); len(got) != 0 {
		t.Errorf("Names after deleting everything = %v, want empty", got)
	}
}

func TestDeleteFromEmptyBook(t *testing.T) {
	pb := NewPhoneBook()
	if pb.Delete("nobody") {
		t.Error("Delete on an empty book returned true")
	}
	if got := pb.Names(); len(got) != 0 {
		t.Errorf("Names on an empty book = %v, want empty", got)
	}
}

func TestSetAfterDelete(t *testing.T) {
	pb := NewPhoneBook()
	pb.Set("ada", "111")
	if !pb.Delete("ada") {
		t.Fatal("Delete(ada) = false, want true")
	}
	pb.Set("ada", "222")
	if got, ok := pb.Get("ada"); !ok || got != "222" {
		t.Errorf("Get after re-adding = %q, %v; want 222, true", got, ok)
	}
	if got := pb.Names(); !slices.Equal(got, []string{"ada"}) {
		t.Errorf("Names after re-adding = %v, want [ada]", got)
	}
}

func TestEmptyName(t *testing.T) {
	pb := NewPhoneBook()
	pb.Set("", "555")
	if got, ok := pb.Get(""); !ok || got != "555" {
		t.Errorf("Get(empty name) = %q, %v; want 555, true", got, ok)
	}
	if !pb.Delete("") {
		t.Error("Delete(empty name) = false, want true")
	}
	pb.Set("", "")
	if got, ok := pb.Get(""); !ok || got != "" {
		t.Errorf("Get(empty name, empty number) = %q, %v; want \"\", true", got, ok)
	}
}

func TestNamesUnicodeSort(t *testing.T) {
	pb := NewPhoneBook()
	pb.Set("é", "1")
	pb.Set("z", "2")
	pb.Set("日", "3")
	pb.Set("a", "4")

	want := []string{"a", "z", "é", "日"}
	if got := pb.Names(); !slices.Equal(got, want) {
		t.Fatalf("Names = %v, want %v", got, want)
	}
}
`,
				}},
				Hints: []string{
					"§NewPhoneBook§ returns §&PhoneBook{entries: map[string]string{}}§.",
					"§Set§: §pb.entries[name] = number§.",
					"§Get§: §v, ok := pb.entries[name]; return v, ok§.",
					"§Delete§ must know whether it removed something: check with §_, ok := pb.entries[name]§ before deleting, then return §ok§.",
					"§Names§ collects the keys then §slices.Sort§ them.",
				},
				Solution: `package main

import (
	"fmt"
	"slices"
)

type PhoneBook struct {
	entries map[string]string
}

func NewPhoneBook() *PhoneBook {
	return &PhoneBook{entries: map[string]string{}}
}

func (pb *PhoneBook) Set(name, number string) {
	pb.entries[name] = number
}

func (pb *PhoneBook) Get(name string) (string, bool) {
	number, ok := pb.entries[name]
	return number, ok
}

func (pb *PhoneBook) Delete(name string) bool {
	if _, ok := pb.entries[name]; !ok {
		return false
	}
	delete(pb.entries, name)
	return true
}

func (pb *PhoneBook) Names() []string {
	names := make([]string, 0, len(pb.entries))
	for name := range pb.entries {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func main() {
	pb := NewPhoneBook()
	pb.Set("ada", "555-0100")
	pb.Set("bo", "555-0101")
	fmt.Println(pb.Get("ada"))
	fmt.Println(pb.Delete("bo"), pb.Delete("zed"))
	fmt.Println(pb.Names())
}
`,
			},
		},
	}
}
