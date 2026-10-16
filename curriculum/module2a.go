package curriculum

// module2 builds Module 2: the Go you need to write a backend service, in a
// deliberately gentle order. Concurrency first (because it shapes everything
// else), then the HTTP stack, then persistence and operations.
func module2() Module {
	return Module{
		ID:      "m2",
		Title:   "Module 2 · Backend Engineering in Go",
		Summary: "Concurrency, HTTP, JSON, routing, middleware, databases, structured logging and testing — the standard library plus the packages you meet in real services (chi, sqlx).",
		Topics: []Topic{
			m2t1(),  // goroutines & WaitGroup
			m2t2(),  // channels
			m2t3(),  // select, timeouts, mutexes
			m2t4(),  // context
			m2t5(),  // JSON & struct tags
			m2t6(),  // net/http handlers & routing
			m2t7(),  // chi & REST patterns
			m2t8(),  // middleware
			m2t9(),  // HTTP clients
			m2t10(), // database/sql
			m2t11(), // sqlx & transactions
			m2t12(), // slog, config & backend testing
		},
	}
}

func m2t1() Topic {
	return Topic{
		ID:      "m2t1",
		Title:   "Goroutines and sync.WaitGroup",
		Summary: "Concurrent execution, and waiting for it to finish.",
		Theory: `
A **goroutine** is a function running concurrently with the code that launched it. The
§go§ statement starts one. Goroutines are cheap — a couple of kilobytes of stack that
grows on demand — so starting one per request or per item is normal.

## Starting one

The §go§ keyword in front of a call is the whole syntax:

§§§go
go doWork()
§§§

There is no §Thread§ class, no §new Thread(runnable)§ and no §async§/§await§.

For comparison: a Java thread costs roughly a megabyte of stack plus a system call to
create. The Go runtime multiplexes thousands of goroutines onto a handful of OS threads,
each starting with a ~2 KB stack, so one per request or per item is normal.

§go§ is a *statement*, so there is nothing to assign. §x := go f()§ is a compile error.

The function may be an ordinary one or a function literal:

§§§go
go sendEmail(user.ID)

go func() {
	fmt.Println("running in the background")
}()
§§§

Note the §()§ at the end of the second form: §func() { … }§ is a *value*; the §()§ calls
it. §go func() { … }§ without the final §()§ does not compile.

## main does not wait for them

The moment §main§ returns, the program exits — and every goroutine dies with it, even
if it was halfway through something:

§§§go
package main

import "fmt"

func main() {
	go fmt.Println("from the goroutine")
	fmt.Println("from main")
}
§§§

Possible output:

§§§text
from main
§§§

The goroutine *may* get a chance to run, or it may not — the program does not wait for
it. This is the single most common beginner surprise.

## sync.WaitGroup

§sync.WaitGroup§ (from the §sync§ package) is a counter of running jobs.

* §wg.Add(1)§ — "one more job has started".
* §wg.Done()§ — "one job has finished". (§Done§ is short for §Add(-1)§.)
* §wg.Wait()§ — "block until the counter reaches zero".

Put together:

§§§go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1) // register BEFORE starting the goroutine
		go func(n int) {
			defer wg.Done() // guarantee the counter decrements
			fmt.Println("worker", n)
		}(i)
	}

	wg.Wait() // block here until all three have called Done
	fmt.Println("all done")
}
§§§

One possible output (the order of the three lines is not guaranteed):

§§§text
worker 3
worker 1
worker 2
all done
§§§

§all done§ is *always* last. That is the guarantee §Wait§ gives you.

### Why §Add§ goes before §go§

If you wrote §wg.Add(1)§ *inside* the goroutine, §Wait§ could run before any goroutine
started, see a counter of zero, and return immediately — the classic race. Rule:
**call §Add§ before you launch.**

### Why §defer wg.Done()§

Because §Done§ must run on every path out of the function, including an early §return§
or a panic. §defer§ is exactly the tool for that, as you saw in Module 1.

### The loop-variable question

You will see both of these shapes:

§§§go
// Modern Go (1.22+): each iteration has its own i, so this is safe.
for i := 1; i <= 3; i++ {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println(i)
	}()
}

// Explicit and safe on every Go version: pass the value in.
for i := 1; i <= 3; i++ {
	wg.Add(1)
	go func(n int) {
		defer wg.Done()
		fmt.Println(n)
	}(i)
}
§§§

Before Go 1.22 the first version was a famous bug: every goroutine might print the final
value of §i§. Passing the value as an argument (or shadowing it with §i := i§) avoided
that. The second form is still worth writing when you want to be explicit about what
the goroutine captures — and when you need to support older toolchains.

## Sharing data between goroutines

If two goroutines both write a variable, you have a **data race**, and the result is
undefined. Go's tool for finding them is the race detector:

§§§bash
go test -race ./...
§§§

It is worth the slowdown; run it in CI. The tools that fix races — mutexes and
channels — are the subjects of the next two topics.

> **Note:** a variable declared inside a goroutine is private to it. Only variables
> from an enclosing scope are shared.

## A complete example

§§§go
package main

import (
	"fmt"
	"sync"
)

func main() {
	words := []string{"alpha", "beta", "gamma", "delta"}

	var wg sync.WaitGroup
	for i, w := range words {
		wg.Add(1)
		go func(index int, word string) {
			defer wg.Done()
			fmt.Printf("job %d: %s has %d letters\n", index, word, len(word))
		}(i, w)
	}
	wg.Wait()

	fmt.Println("every job finished")
}
§§§

Typical output (the four job lines can come out in any order):

§§§text
job 3: delta has 5 letters
job 0: alpha has 5 letters
job 1: beta has 4 letters
job 2: gamma has 5 letters
every job finished
§§§

## Rules of thumb

* Use goroutines for work that is **independent** and **waits on something** (network,
  disk, timers). For pure CPU work, more goroutines than CPU cores does not help.
* Always have a way for a goroutine to *finish*. A goroutine blocked forever is a leak
  that holds its memory for the life of the process.
* §WaitGroup§ when you only need "wait for these to end". §channels§ when the goroutines
  need to hand data to each other — that is the next topic.
`,
		Quiz: []Question{
			{
				ID:     "m2t1q1",
				Prompt: "What is the output of this program?",
				Code: `package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Print(i, " ")
		}(i)
	}
	wg.Wait()
	fmt.Print("done")
}`,
				Options: []Option{
					{Text: "§0 1 2 § in some order, then §done§", Correct: true, Rationale: "The three goroutines run concurrently, so their order is not fixed, but §Wait§ guarantees they all finish before §done§."},
					{Text: "§0 1 2 done§ always, in exactly that order.", Correct: false, Rationale: "Concurrent goroutines have no guaranteed ordering."},
					{Text: "§done§ only — the goroutines are killed when §main§ ends.", Correct: false, Rationale: "§wg.Wait()§ blocks §main§ until all three call §Done§."},
					{Text: "It does not compile.", Correct: false, Rationale: "Every part of it is valid."},
				},
				Explanation: "The *order* is unspecified; the *completeness* is guaranteed by §WaitGroup§.",
			},
			{
				ID:     "m2t1q2",
				Prompt: "What is wrong with this code?",
				Code: `var wg sync.WaitGroup
for i := 0; i < 5; i++ {
	go func() {
		wg.Add(1)
		defer wg.Done()
		doWork(i)
	}()
}
wg.Wait()`,
				Options: []Option{
					{Text: "§wg.Add(1)§ is inside the goroutine, so §Wait§ can reach zero before any of them start — it may return immediately.", Correct: true, Rationale: "§Add§ must be called before §go§, on the goroutine that launches the work."},
					{Text: "§defer§ cannot be used inside a goroutine.", Correct: false, Rationale: "It can, and it is the right tool for §Done§."},
					{Text: "§wg.Wait()§ blocks forever.", Correct: false, Rationale: "The risk is the opposite: it returns too early."},
					{Text: "Nothing — the code is correct.", Correct: false, Rationale: "It is a race, and §go test -race§ would not even need to try hard to catch it."},
				},
				Explanation: "Rule: §Add§ before §go§. Register the job on the launcher's side, then start it.",
			},
			{
				ID:     "m2t1q3",
				Prompt: "Which statements about goroutines are true? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "A goroutine costs far less memory to start than an OS thread.", Correct: true, Rationale: "About 2 KB of stack that grows on demand, versus roughly a megabyte."},
					{Text: "§go f()§ returns a value you can assign.", Correct: false, Rationale: "§go§ is a statement, not an expression. Results travel through channels."},
					{Text: "When §main§ returns, running goroutines are stopped.", Correct: true, Rationale: "The process exits, so everything ends."},
					{Text: "Two goroutines may safely write the same variable without synchronisation.", Correct: false, Rationale: "That is a data race. Use a mutex or a channel."},
					{Text: "§wg.Done()§ must be called once for each §wg.Add(1)§.", Correct: true, Rationale: "The counter must return to zero, or §Wait§ blocks forever."},
				},
				Explanation: "Cheap to start, no return values, killed with the process, and never share memory without synchronisation.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m2t1e1",
				Title: "Warm-up: run some jobs concurrently",
				Prompt: `One goroutine per job.


§§§go
func RunAll(jobs []func()) int
§§§

Starts one goroutine per job, waits for all of them, and returns how many jobs ran.

* Every job must actually run exactly once.
* §RunAll(nil)§ returns §0§ and must not block.
* The order in which the jobs run is not specified — but they must all have finished
  before §RunAll§ returns. The test relies on that: each job increments a shared
  counter *guarded by a mutex it provides itself*, and checks the total afterwards.

The important detail is doing §Add§ **before** §go§.
`,
				Starter: `package main

import (
	"fmt"
	"sync"
)

// RunAll runs every job in its own goroutine and returns how many ran.
func RunAll(jobs []func()) int {
	// TODO
	_ = sync.WaitGroup{}
	return 0
}

func main() {
	count := 0
	var mu sync.Mutex
	jobs := make([]func(), 3)
	for i := range jobs {
		jobs[i] = func() {
			mu.Lock()
			count++
			mu.Unlock()
		}
	}
	fmt.Println(RunAll(jobs), count)
}
`,
				Tests: []File{{
					Name: "runall_test.go",
					Content: `package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunAll(t *testing.T) {
	var counter atomic.Int64
	jobs := make([]func(), 25)
	for i := range jobs {
		jobs[i] = func() { counter.Add(1) }
	}

	if got := RunAll(jobs); got != 25 {
		t.Fatalf("RunAll returned %d, want 25", got)
	}
	if counter.Load() != 25 {
		t.Fatalf("%d jobs ran, want 25 — did you wait for them?", counter.Load())
	}
}

func TestRunAllEmpty(t *testing.T) {
	if got := RunAll(nil); got != 0 {
		t.Fatalf("RunAll(nil) = %d, want 0", got)
	}
	if got := RunAll([]func(){}); got != 0 {
		t.Fatalf("RunAll(empty) = %d, want 0", got)
	}
}

func TestRunAllWaits(t *testing.T) {
	// Every job blocks until all of them have started. That is only possible
	// if they really do run concurrently — a serial implementation deadlocks.
	const n = 5
	var started sync.WaitGroup
	started.Add(n)

	jobs := make([]func(), n)
	for i := range jobs {
		jobs[i] = func() { started.Done(); started.Wait() }
	}

	done := make(chan int, 1)
	go func() { done <- RunAll(jobs) }()

	select {
	case got := <-done:
		if got != n {
			t.Fatalf("RunAll = %d, want %d", got, n)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunAll ran the jobs one at a time instead of concurrently")
	}
}
`,
				}},
				Hints: []string{
					"Create §var wg sync.WaitGroup§, then for each job: §wg.Add(1)§, §go func(){ defer wg.Done(); job() }()§.",
					"Pass the job into the goroutine as a parameter, or capture it in a loop variable declared inside the loop — the safest habit.",
					"After the loop, §wg.Wait()§, then §return len(jobs)§.",
					"§RunAll(nil)§ works automatically: the loop body never runs, so §Wait§ returns at once.",
				},
				Solution: `package main

import (
	"fmt"
	"sync"
)

func RunAll(jobs []func()) int {
	var wg sync.WaitGroup
	for _, job := range jobs {
		job := job
		wg.Add(1)
		go func() {
			defer wg.Done()
			job()
		}()
	}
	wg.Wait()
	return len(jobs)
}

func main() {
	count := 0
	var mu sync.Mutex
	jobs := make([]func(), 3)
	for i := range jobs {
		jobs[i] = func() {
			mu.Lock()
			count++
			mu.Unlock()
		}
	}
	fmt.Println(RunAll(jobs), count)
}
`,
			},
			{
				ID:    "m2t1e2",
				Title: "Split the work: concurrent sum",
				Prompt: `Collect a result from each goroutine — something a §WaitGroup§ alone cannot do.

§§§go
func SumConcurrent(nums []int, workers int) int
§§§

* Split §nums§ into at most §workers§ contiguous chunks.
* Sum each chunk in its own goroutine.
* Combine the partial sums safely. The idiomatic first choice is §sync/atomic§:

§§§go
var total atomic.Int64
total.Add(int64(chunkSum))   // safe from many goroutines at once
value := total.Load()
§

§atomic.Int64§ is a counter the hardware updates without a lock. Use it when you only
need to add to a number.

* §workers < 1§ behaves like §workers == 1§. §workers§ larger than the number of
  elements is clamped.
* An empty input returns §0§.

The test compares your answer against a sequential sum for several worker counts, so
any lost update shows up as a wrong total.
`,
				Starter: `package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// SumConcurrent sums nums using at most §workers§ goroutines.
func SumConcurrent(nums []int, workers int) int {
	// TODO
	var total atomic.Int64
	_ = total
	_ = sync.WaitGroup{}
	return 0
}

func main() {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8}
	fmt.Println(SumConcurrent(nums, 4))
}
`,
				Tests: []File{{
					Name: "sumconcurrent_test.go",
					Content: `package main

import (
	"math/rand"
	"testing"
)

func TestSumConcurrent(t *testing.T) {
	for _, workers := range []int{1, 2, 3, 8, 100} {
		nums := make([]int, 1000)
		want := 0
		for i := range nums {
			nums[i] = rand.Intn(100)
			want += nums[i]
		}
		if got := SumConcurrent(nums, workers); got != want {
			t.Fatalf("workers=%d: got %d, want %d", workers, got, want)
		}
	}
}

func TestSumConcurrentEdges(t *testing.T) {
	if got := SumConcurrent(nil, 4); got != 0 {
		t.Fatalf("SumConcurrent(nil) = %d, want 0", got)
	}
	if got := SumConcurrent([]int{5}, 0); got != 5 {
		t.Fatalf("SumConcurrent(single, 0) = %d, want 5", got)
	}
	if got := SumConcurrent([]int{1, 2}, -3); got != 3 {
		t.Fatalf("SumConcurrent(negative workers) = %d, want 3", got)
	}
	if got := SumConcurrent([]int{1, 2, 3}, 2); got != 6 {
		t.Fatalf("SumConcurrent = %d, want 6", got)
	}
}
`,
				}},
				Hints: []string{
					"Guard and clamp first: §if len(nums) == 0 { return 0 }§, §if workers < 1 { workers = 1 }§, §if workers > len(nums) { workers = len(nums) }§.",
					"Chunk size: §size := (len(nums) + workers - 1) / workers§ — integer division rounded up.",
					"Loop §for start := 0; start < len(nums); start += size§, compute §end := min(start+size, len(nums))§, and launch a goroutine with §nums[start:end]§.",
					"Inside the goroutine, sum the chunk into a local variable and then §total.Add(int64(sum))§ — one atomic write, not one per element.",
					"Finish with §wg.Wait()§ and §return int(total.Load())§.",
				},
				Solution: `package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func SumConcurrent(nums []int, workers int) int {
	if len(nums) == 0 {
		return 0
	}
	if workers < 1 {
		workers = 1
	}
	if workers > len(nums) {
		workers = len(nums)
	}

	size := (len(nums) + workers - 1) / workers
	var total atomic.Int64
	var wg sync.WaitGroup

	for start := 0; start < len(nums); start += size {
		end := min(start+size, len(nums))
		wg.Add(1)
		go func(chunk []int) {
			defer wg.Done()
			sum := 0
			for _, v := range chunk {
				sum += v
			}
			total.Add(int64(sum))
		}(nums[start:end])
	}

	wg.Wait()
	return int(total.Load())
}

func main() {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8}
	fmt.Println(SumConcurrent(nums, 4))
}
`,
			},
		},
	}
}

func m2t2() Topic {
	return Topic{
		ID:      "m2t2",
		Title:   "Channels",
		Summary: "Pipes that carry values between goroutines.",
		Theory: `
A **channel** is a typed conduit between goroutines: one sends, another receives. Since
§go§ provides no way to return a value, a channel is how concurrently-launched work
reports its result back.

If you have used Unix pipes (§ls | grep go§) you already have the right mental model —
except that a Go channel carries *typed values*, not text, and it is safe for any
number of goroutines at once.

## Creating one and using it

§§§go
ch := make(chan int)   // a channel carrying ints

ch <- 42               // SEND: push 42 in
v := <-ch              // RECEIVE: take a value out
§§§

The arrow points the way the value travels. §chan int§ is the type; §<-chan int§ means
"receive-only"; §chan<- int§ means "send-only".

A complete program:

§§§go
package main

import "fmt"

func main() {
	ch := make(chan string)

	go func() {
		ch <- "hello from the goroutine"
	}()

	msg := <-ch
	fmt.Println(msg)
}
§§§

Output:

§§§text
hello from the goroutine
§§§

The §<-ch§ in §main§ blocks until a value arrives. That is what makes the program
correct: §main§ cannot finish before the message is delivered.

## Unbuffered channels rendezvous

§make(chan int)§ creates an **unbuffered** channel. A send blocks until a receiver is
ready, and a receive blocks until a sender arrives. The two goroutines meet.

That synchronisation is a feature: an unbuffered channel is a handshake. If you want
the sender to be able to get ahead, give the channel a buffer:

§§§go
ch := make(chan int, 3)   // room for 3 values

ch <- 1   // does not block: the buffer takes it
ch <- 2
ch <- 3
// ch <- 4  // would block until someone receives

fmt.Println(<-ch) // 1
§§§

| | Unbuffered §make(chan T)§ | Buffered §make(chan T, n)§ |
| --- | --- | --- |
| Send blocks | until a receiver is ready | only when the buffer is full |
| Receive blocks | until a sender is ready | only when the buffer is empty |
| Good for | handshakes, signalling, back-pressure | decoupling producer from consumer |

A common idiom is a buffer of 1 for "at most one pending message".

## Closing a channel

§close(ch)§ means **"no more values will ever be sent"**. It is how a producer tells
consumers it has finished:

§§§go
ch := make(chan int, 3)
ch <- 1
ch <- 2
close(ch)

for {
	v, ok := <-ch
	if !ok {
		break // the channel is closed and empty
	}
	fmt.Println(v)
}
§§§

The §ok§ is §false§ once the channel is closed *and* drained — the same comma-ok shape
you have used for maps and lookups.

Because that loop is so common, §range§ over a channel does it for you:

§§§go
for v := range ch {
	fmt.Println(v)   // runs until the channel is closed and empty
}
§§§

Three rules, and breaking any of them is a panic:

1. **Only the sender closes.** A receiver that closes can cause a "send on closed
   channel" panic in the sender.
2. **Never close twice.** §close§ of a closed channel panics.
3. **A closed channel still yields values.** Receiving from a closed, drained channel
   gives you the zero value immediately, forever.

A §nil§ channel is the third case worth knowing: sending or receiving on it blocks
forever. That sounds useless, but it is a handy trick in §select§ (next topic).

## Directional channels document intent

A function signature can state which way a channel will be used:

§§§go
func produce(out chan<- int) {   // this function only sends
	out <- 1
}

func consume(in <-chan int) int { // this function only receives
	return <-in
}
§§§

Only the *creating* function needs the bidirectional §chan int§. Everywhere else,
narrow it. It makes the code self-documenting and impossible to misuse.

## A producer/consumer example

§§§go
package main

import "fmt"

func produce(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // the producer closes
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

func main() {
	for v := range square(produce(2, 3, 4)) {
		fmt.Println(v)
	}
}
§§§

Output:

§§§text
4
9
16
§§§

This is a **pipeline**: §produce§ feeds §square§, which feeds the loop in §main§. Each
stage runs in its own goroutine and closes its own output channel when it is done.

The §defer close(out)§ at the top of each goroutine guarantees the channel is closed on
every exit path, which is what ends the §range§ in the next stage.

## A worker pool

The other fundamental pattern: several workers pulling from one job channel.

§§§go
func ParallelSquare(nums []int, workers int) []int {
	out := make([]int, len(nums))
	jobs := make(chan int)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				out[i] = nums[i] * nums[i]
			}
		}()
	}

	for i := range nums {
		jobs <- i
	}
	close(jobs)   // no more jobs; workers finish their range loops
	wg.Wait()
	return out
}
§§§

Why is writing to §out[i]§ from several goroutines safe? Because each goroutine writes a
**different index**, and nobody reads the slice until §Wait§ returns. A data race needs
two goroutines touching the *same* memory, with at least one writing.

## Deadlock

§§§go
func main() {
	ch := make(chan int)
	ch <- 1 // nobody is receiving — this blocks forever
}
§§§

Go prints:

§§§text
fatal error: all goroutines are asleep - deadlock!
§§§

That message means every goroutine is blocked and nothing can ever wake them. The two
everyday causes:

* sending on an unbuffered channel with no receiver,
* receiving from a channel nobody ever sends on.

## Putting it together

| You want | Use |
| --- | --- |
| To wait for goroutines, no data | §sync.WaitGroup§ |
| To hand values between goroutines | a channel |
| To signal "no more values" | §close§ from the sender |
| To let a producer get ahead | a buffered channel |
| To iterate until a producer stops | §for v := range ch§ |

Go's proverb: **"Do not communicate by sharing memory; instead, share memory by
communicating."** Channels are the communicating half of that sentence; mutexes (next
topic) are for the cases where sharing is genuinely simpler.
`,
		Quiz: []Question{
			{
				ID:     "m2t2q1",
				Prompt: "What does this print?",
				Code: `package main

import "fmt"

func main() {
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	close(ch)

	for v := range ch {
		fmt.Print(v, " ")
	}

	v, ok := <-ch
	fmt.Print(v, ok)
}`,
				Options: []Option{
					{Text: "§1 2 0 false§", Correct: true, Rationale: "§range§ drains the buffer and stops at the close; the extra receive then returns the zero value with §ok == false§."},
					{Text: "§1 2§ and then it blocks forever.", Correct: false, Rationale: "A closed, drained channel returns immediately."},
					{Text: "§1 2 0 true§", Correct: false, Rationale: "§ok§ is false once the channel is closed and empty."},
					{Text: "It panics: receiving from a closed channel.", Correct: false, Rationale: "Only *sending* to a closed channel panics."},
				},
				Explanation: "Receiving from a closed channel never blocks: you get buffered values first, then the zero value with §ok == false§.",
			},
			{
				ID:     "m2t2q2",
				Prompt: "Which lines panic at run time? Select all that apply.",
				Multi:  true,
				Code: `ch := make(chan int, 1)
ch <- 1
close(ch)
close(ch)
ch <- 2
<-ch`,
				Options: []Option{
					{Text: "The second §close(ch)§", Correct: true, Rationale: "Closing a closed channel panics: \"close of closed channel\"."},
					{Text: "§ch <- 2§", Correct: true, Rationale: "Sending on a closed channel panics: \"send on closed channel\"."},
					{Text: "§<-ch§", Correct: false, Rationale: "Receiving from a closed channel is fine — it returns the buffered 1, then zeros."},
					{Text: "The first §close(ch)§", Correct: false, Rationale: "That is the legitimate close."},
				},
				Explanation: "Close exactly once, and only from the sender. Sends after a close are a panic, receives are not.",
			},
			{
				ID:     "m2t2q3",
				Prompt: "What is the difference between §make(chan int)§ and §make(chan int, 5)§?",
				Options: []Option{
					{Text: "The first is unbuffered: every send waits for a receiver. The second buffers five values, so sends only wait when the buffer is full.", Correct: true, Rationale: "The number is the buffer size."},
					{Text: "The second can hold five *goroutines*.", Correct: false, Rationale: "It holds five values."},
					{Text: "The first is faster.", Correct: false, Rationale: "They do different jobs rather than being fast or slow."},
					{Text: "The second never blocks.", Correct: false, Rationale: "It blocks once the buffer is full."},
				},
				Explanation: "An unbuffered channel is a rendezvous; a buffered one is a small queue.",
			},
			{
				ID:     "m2t2q4",
				Prompt: "Who should call §close§ on a channel?",
				Options: []Option{
					{Text: "The sending side, when it is sure there are no more values to send.", Correct: true, Rationale: "Closing is the sender saying \"finished\". Only the sender knows that."},
					{Text: "The receiving side, when it has read everything it wants.", Correct: false, Rationale: "Then a still-running sender would panic on its next send."},
					{Text: "Whoever created the channel.", Correct: false, Rationale: "Creation and completion are different moments."},
					{Text: "Nobody — channels are garbage collected without closing.", Correct: false, Rationale: "Without a close, a §range§ over the channel never ends."},
				},
				Explanation: "Sender closes. Receiver ranges. That division is why §for v := range ch§ is safe.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m2t2e1",
				Title: "Warm-up: a counting producer",
				Prompt: `Build the smallest useful channel pair.

§§§go
func CountTo(n int) <-chan int
func SumAll(in <-chan int) int
§§§

* §CountTo(n)§ returns a channel that yields §1, 2, …, n§ and then closes. §n <= 0§
  produces an immediately closed channel.
* §SumAll(in)§ drains the channel and returns the total — a plain §for v := range in§
  loop.

§CountTo§ must do its work in a goroutine, because the caller has not started receiving
yet. §defer close(out)§ inside that goroutine is what ends §SumAll§'s loop.
`,
				Starter: `package main

import "fmt"

// CountTo returns a channel yielding 1..n, then closing.
func CountTo(n int) <-chan int {
	out := make(chan int)
	// TODO: launch a goroutine, defer close(out), send 1..n
	return out
}

// SumAll drains in and returns the total.
func SumAll(in <-chan int) int {
	// TODO
	return 0
}

func main() {
	fmt.Println(SumAll(CountTo(5)))
	fmt.Println(SumAll(CountTo(0)))
}
`,
				Tests: []File{{
					Name: "count_test.go",
					Content: `package main

import "testing"

func TestCountTo(t *testing.T) {
	var got []int
	for v := range CountTo(5) {
		got = append(got, v)
	}
	want := []int{1, 2, 3, 4, 5}
	if len(got) != len(want) {
		t.Fatalf("CountTo(5) yielded %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("CountTo(5) yielded %v, want %v", got, want)
		}
	}
}

func TestCountToZero(t *testing.T) {
	count := 0
	for range CountTo(0) {
		count++
	}
	if count != 0 {
		t.Fatalf("CountTo(0) yielded %d values, want 0", count)
	}
	for range CountTo(-3) {
		count++
	}
	if count != 0 {
		t.Fatalf("CountTo(-3) yielded %d values, want 0", count)
	}
}

func TestSumAll(t *testing.T) {
	if got := SumAll(CountTo(4)); got != 10 {
		t.Fatalf("SumAll(CountTo(4)) = %d, want 10", got)
	}
	if got := SumAll(CountTo(100)); got != 5050 {
		t.Fatalf("SumAll(CountTo(100)) = %d, want 5050", got)
	}
	if got := SumAll(CountTo(0)); got != 0 {
		t.Fatalf("SumAll(CountTo(0)) = %d, want 0", got)
	}
}

func TestSumAllOnClosedChannel(t *testing.T) {
	ch := make(chan int)
	close(ch)
	if got := SumAll(ch); got != 0 {
		t.Fatalf("SumAll(closed) = %d, want 0", got)
	}
}
`,
				}},
				Hints: []string{
					"§go func() { defer close(out); for i := 1; i <= n; i++ { out <- i } }()§.",
					"The §defer close(out)§ must come first inside the goroutine so it runs on every exit path.",
					"For §n <= 0§ the loop body never runs, so the channel closes immediately — no special case needed.",
					"§SumAll§: §total := 0; for v := range in { total += v }; return total§.",
				},
				Solution: `package main

import "fmt"

func CountTo(n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 1; i <= n; i++ {
			out <- i
		}
	}()
	return out
}

func SumAll(in <-chan int) int {
	total := 0
	for v := range in {
		total += v
	}
	return total
}

func main() {
	fmt.Println(SumAll(CountTo(5)))
	fmt.Println(SumAll(CountTo(0)))
}
`,
			},
			{
				ID:    "m2t2e2",
				Title: "A worker pool that keeps the order",
				Prompt: `§§§go
func ParallelMap(nums []int, workers int, f func(int) int) []int
§§§

Apply §f§ to every element using at most §workers§ goroutines pulling from a single job
channel, and keep the **result in input order**.

* §f§ must be called exactly once per element.
* §workers < 1§ behaves like §1§; §workers§ larger than the input is clamped.
* Empty input returns an empty (non-nil) slice.

The test uses a deliberately slow §f§ that sleeps a different amount per element, so the
work finishes out of order — but the result must not. That is why each worker writes into
§out[i]§ at the index it was given, rather than appending.

Why is that safe? Because each index has exactly one writer and nobody reads the slice
until every worker has finished.
`,
				Starter: `package main

import (
	"fmt"
	"sync"
)

// ParallelMap applies f to every element of nums using a pool of workers.
func ParallelMap(nums []int, workers int, f func(int) int) []int {
	// TODO
	_ = sync.WaitGroup{}
	return nil
}

func main() {
	out := ParallelMap([]int{1, 2, 3, 4}, 2, func(x int) int { return x * 10 })
	fmt.Println(out)
}
`,
				Tests: []File{{
					Name: "parallelmap_test.go",
					Content: `package main

import (
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestParallelMapOrder(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8}
	got := ParallelMap(nums, 4, func(x int) int {
		time.Sleep(time.Duration(9-x) * time.Millisecond) // finishes out of order
		return x * 10
	})
	want := []int{10, 20, 30, 40, 50, 60, 70, 80}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParallelMap = %v, want %v", got, want)
	}
}

func TestParallelMapCallsOnce(t *testing.T) {
	var calls atomic.Int64
	ParallelMap([]int{1, 2, 3}, 2, func(x int) int {
		calls.Add(1)
		return x
	})
	if calls.Load() != 3 {
		t.Fatalf("f called %d times, want 3", calls.Load())
	}
}

func TestParallelMapEdges(t *testing.T) {
	got := ParallelMap(nil, 4, func(x int) int { return x })
	if got == nil || len(got) != 0 {
		t.Fatalf("ParallelMap(nil) = %v, want an empty non-nil slice", got)
	}
	if got := ParallelMap([]int{7}, 0, func(x int) int { return x + 1 }); !reflect.DeepEqual(got, []int{8}) {
		t.Fatalf("workers=0 = %v, want [8]", got)
	}
	if got := ParallelMap([]int{1, 2}, 99, func(x int) int { return -x }); !reflect.DeepEqual(got, []int{-1, -2}) {
		t.Fatalf("workers > len = %v, want [-1 -2]", got)
	}
}
`,
				}},
				Hints: []string{
					"Preallocate the output with the right length: §out := make([]int, len(nums))§.",
					"Send *indices* on the jobs channel, not values: §for i := range nums { jobs <- i }§.",
					"Each worker loops §for i := range jobs { out[i] = f(nums[i]) }§ — no append, no lock.",
					"Remember §close(jobs)§ after the producer loop, then §wg.Wait()§ before returning.",
				},
				Solution: `package main

import (
	"fmt"
	"sync"
)

func ParallelMap(nums []int, workers int, f func(int) int) []int {
	out := make([]int, len(nums))
	if len(nums) == 0 {
		return out
	}
	if workers < 1 {
		workers = 1
	}
	if workers > len(nums) {
		workers = len(nums)
	}

	jobs := make(chan int)
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				out[i] = f(nums[i])
			}
		}()
	}

	for i := range nums {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return out
}

func main() {
	out := ParallelMap([]int{1, 2, 3, 4}, 2, func(x int) int { return x * 10 })
	fmt.Println(out)
}
`,
			},
			{
				ID:    "m2t2e3",
				Title: "Fan-in: merge many channels into one",
				Prompt: `When you have several producers, you often want to treat them as one stream.

§§§go
func FanIn(chans ...<-chan int) <-chan int
§§§

Returns a single channel that yields every value from every input channel, then closes
once **all** of them have closed.

* Values from a single channel must keep their relative order.
* No value may be lost or duplicated.
* §FanIn()§ with no channels returns a channel that closes immediately.
* §FanIn§ must not touch the input channels' buffering assumptions: it should work on
  unbuffered channels too.

The shape of the solution is the interesting part: one goroutine *per* input channel,
each ranging over its own channel and sending into the shared output, plus one extra
goroutine that waits for the group and then closes the output. That last step cannot
happen on the calling goroutine, because closing must wait for the producers.
`,
				Starter: `package main

import (
	"fmt"
	"sync"
)

// FanIn merges chans into a single channel.
func FanIn(chans ...<-chan int) <-chan int {
	out := make(chan int)
	// TODO: one copier goroutine per channel, plus a closer goroutine
	_ = sync.WaitGroup{}
	return out
}

func main() {
	a := make(chan int, 2)
	a <- 1
	a <- 2
	close(a)

	b := make(chan int, 1)
	b <- 3
	close(b)

	for v := range FanIn(a, b) {
		fmt.Print(v, " ")
	}
	fmt.Println()
}
`,
				Tests: []File{{
					Name: "fanin_test.go",
					Content: `package main

import (
	"slices"
	"testing"
	"time"
)

func collect(ch <-chan int) []int {
	var out []int
	for v := range ch {
		out = append(out, v)
	}
	return out
}

func TestFanIn(t *testing.T) {
	a := make(chan int, 3)
	a <- 1
	a <- 2
	a <- 3
	close(a)

	b := make(chan int, 2)
	b <- 10
	b <- 20
	close(b)

	got := collect(FanIn(a, b))
	slices.Sort(got)

	want := []int{1, 2, 3, 10, 20}
	if !slices.Equal(got, want) {
		t.Fatalf("FanIn yielded %v, want the same values as %v", got, want)
	}
}

func TestFanInOrderWithinChannel(t *testing.T) {
	a := make(chan int, 4)
	for i := 1; i <= 4; i++ {
		a <- i
	}
	close(a)

	got := collect(FanIn(a))
	want := []int{1, 2, 3, 4}
	if !slices.Equal(got, want) {
		t.Fatalf("order within a channel was not preserved: %v", got)
	}
}

func TestFanInEmpty(t *testing.T) {
	ch := FanIn()
	select {
	case v, ok := <-ch:
		if ok {
			t.Fatalf("FanIn() yielded %d, want a closed channel", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("FanIn() never closed its channel")
	}
}

func TestFanInUnbuffered(t *testing.T) {
	mk := func(vals ...int) <-chan int {
		ch := make(chan int) // deliberately unbuffered
		go func() {
			defer close(ch)
			for _, v := range vals {
				ch <- v
			}
		}()
		return ch
	}

	got := collect(FanIn(mk(1, 2), mk(3, 4), mk(5)))
	slices.Sort(got)
	if !slices.Equal(got, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("FanIn on unbuffered channels yielded %v", got)
	}
}
`,
				}},
				Hints: []string{
					"Declare §var wg sync.WaitGroup§ before the loop over §chans§.",
					"For each input: §wg.Add(1)§, then §go func(c <-chan int) { defer wg.Done(); for v := range c { out <- v } }(c)§.",
					"§go func() { wg.Wait(); close(out) }()§ — a *separate* goroutine, because §close§ must wait and the caller must not block.",
					"Return §out§, which is a bidirectional §chan int§ that satisfies the declared §<-chan int§ result.",
					"With no inputs, §wg.Wait()§ returns instantly and the closer goroutine closes §out§ — so the empty case works for free.",
				},
				Solution: `package main

import (
	"fmt"
	"sync"
)

func FanIn(chans ...<-chan int) <-chan int {
	out := make(chan int)

	var wg sync.WaitGroup
	for _, c := range chans {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	a := make(chan int, 2)
	a <- 1
	a <- 2
	close(a)

	b := make(chan int, 1)
	b <- 3
	close(b)

	for v := range FanIn(a, b) {
		fmt.Print(v, " ")
	}
	fmt.Println()
}
`,
			},
		},
	}
}

func m2t3() Topic {
	return Topic{
		ID:      "m2t3",
		Title:   "select, timeouts and mutexes",
		Summary: "Waiting on several channels at once, and the other way to share data.",
		Theory: `
Two tools this time: §select§, which lets one goroutine wait on *several* channels, and
the §sync§ mutex, which protects data that goroutines share directly.

## select: wait on whichever channel is ready

§§§go
select {
case v := <-a:
	fmt.Println("got", v, "from a")
case b <- 42:
	fmt.Println("sent 42 into b")
case <-time.After(time.Second):
	fmt.Println("timed out")
}
§§§

§select§ looks like a §switch§, but each case is a **channel operation**. It blocks
until one of them can proceed, then runs that case. If several are ready, it picks one
**at random**.

Two details worth noticing:

* A case may be a send (§b <- 42§) or a receive (§<-a§), not just a receive.
* §time.After(d)§ returns a channel that receives a value after the duration — so it
  gives you a *timeout* in a §select§.

### The default case

Add §default§ and §select§ never blocks:

§§§go
select {
case v := <-ch:
	fmt.Println("got", v)
default:
	fmt.Println("nothing waiting right now")
}
§§§

This is the idiom for **"try this, but do not wait"** — polling, or a non-blocking
enqueue. You met it in the theory of the last topic and will implement it in this
topic's exercises.

### nil channels in select

A case whose channel is §nil§ is never ready. Setting a variable to §nil§ therefore
*disables* that case:

§§§go
for a != nil || b != nil {
	select {
	case v, ok := <-a:
		if !ok {
			a = nil // this case is now permanently disabled
			continue
		}
		fmt.Println("a:", v)
	case v, ok := <-b:
		if !ok {
			b = nil
			continue
		}
		fmt.Println("b:", v)
	}
}
§§§

That is the tidy way to drain two channels and stop when both are exhausted.

## time.After and time.Tick

§§§go
select {
case v := <-work:
	handle(v)
case <-time.After(2 * time.Second):
	return errors.New("timed out")
}
§§§

§time.After§ is convenient but allocates a timer each call. In a hot loop, create one
§time.Timer§ instead and §defer timer.Stop()§. §time.Tick§ returns a channel that fires
on a schedule; like §After§ it leaks unless you use §time.NewTicker§ and stop it.

## Running something with a deadline

§§§go
func WithTimeout(f func() int, d time.Duration) (int, bool) {
	done := make(chan int, 1) // buffered, so the goroutine never blocks
	go func() {
		done <- f()
	}()

	select {
	case v := <-done:
		return v, true
	case <-time.After(d):
		return 0, false
	}
}
§§§

That §buffer of 1§ matters: if the timeout wins, §f§ finishes later and tries to send
with nobody listening. With a buffer of 1 the send never blocks, so the goroutine can
die quietly instead of leaking.

## Mutexes: the other way to share

Channels are for passing values *between* goroutines. Sometimes, though, the natural
thing is a piece of data that several goroutines read and update — a counter, a cache, a
map of sessions. For that, §sync.Mutex§ ("mutual exclusion") is simpler.

§§§go
type Counter struct {
	mu    sync.Mutex
	value int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++ // only one goroutine is ever in here
}

func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}
§§§

* §Lock()§ waits until no other goroutine holds the lock, then takes it.
* §Unlock()§ releases it. **Every §Lock§ must have exactly one §Unlock§.**
* §defer c.mu.Unlock()§ guarantees the unlock on every path, which is the pattern you
  saw with §defer§ in Module 1.

### RWMutex: many readers, one writer

§§§go
type Cache struct {
	mu    sync.RWMutex
	items map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()         // shared: several readers at once
	defer c.mu.RUnlock()
	v, ok := c.items[key]
	return v, ok
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()          // exclusive: one writer, nobody reading
	defer c.mu.Unlock()
	c.items[key] = value
}
§§§

§RLock§ is for readers. Any number of them may hold it at the same time; a writer waits
for all of them, and blocks new readers while it holds §Lock§. For a read-heavy cache —
which is most caches — this is a real win.

Two rules:

1. **Never copy a struct containing a mutex.** Always pass §*Counter§, never §Counter§.
   This is why such a struct is usually created with a §NewX() *X§ function and used
   through a pointer.
2. **Keep critical sections short.** Do not do I/O while holding a lock.

### A map is not safe for concurrent use

Even with different keys! Two goroutines writing to the same map is a data race, and the
runtime can detect it and crash the program:

§§§text
fatal error: concurrent map writes
§§§

Guard it with a §sync.RWMutex§ (as above), use §sync.Map§ for a specific write-once
cache pattern, or use channels to serialise access. There is no fourth option.

## Which should I use, channels or mutexes?

| Use channels when | Use a mutex when |
| --- | --- |
| You are passing ownership of a value along | You are protecting a small piece of shared state |
| You want to sequence stages of a pipeline | You have a struct with fields many goroutines touch |
| One goroutine owns the data and others ask it | A counter, a cache, a registry |

Both are correct tools. The prose advice in the Go community is "use channels to
orchestrate, mutexes to guard", and in practice you will use both in the same file.

## The race detector, again

§§§bash
go test -race ./...
§§§

It finds real races by watching memory accesses at run time and reports both stacks —
the write and the read that conflict. Make it part of CI; do not rely on noticing a
wrong number occasionally.

## A complete example

§§§go
package main

import (
	"fmt"
	"sync"
	"time"
)

type SafeCounter struct {
	mu sync.RWMutex
	n  int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

func (c *SafeCounter) Get() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.n
}

func main() {
	c := &SafeCounter{}
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Inc()
		}()
	}

	// Meanwhile, watch a channel with a timeout.
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
		fmt.Println("all increments finished:", c.Get())
	case <-time.After(2 * time.Second):
		fmt.Println("timed out at", c.Get())
	}
}
§§§

Output:

§§§text
all increments finished: 100
§§§

Note the §SafeCounter§ is used through a pointer (§&SafeCounter{}§) — copying a struct
with a mutex inside would copy the lock, which is exactly what the mutex is there to
prevent.
`,
		Quiz: []Question{
			{
				ID:     "m2t3q1",
				Prompt: "What happens when several cases in a §select§ are ready at the same time?",
				Options: []Option{
					{Text: "Go picks one at random.", Correct: true, Rationale: "Random selection stops code from depending on an accidental ordering."},
					{Text: "The first ready case in source order wins.", Correct: false, Rationale: "Order in the source does not determine priority."},
					{Text: "All ready cases run, one after another.", Correct: false, Rationale: "Exactly one case runs per §select§."},
					{Text: "It panics.", Correct: false, Rationale: "It is well-defined behaviour."},
				},
				Explanation: "One case runs per §select§, chosen at random among the ready ones. Use a loop to handle all of them.",
			},
			{
				ID:     "m2t3q2",
				Prompt: "Why is the §done§ channel buffered in this snippet?",
				Code: `done := make(chan int, 1)
go func() { done <- f() }()

select {
case v := <-done:
	return v, true
case <-time.After(d):
	return 0, false
}`,
				Options: []Option{
					{Text: "So that if the timeout wins, the goroutine can still send without blocking forever — otherwise it would leak.", Correct: true, Rationale: "With an unbuffered channel and nobody listening, the send blocks for the life of the program."},
					{Text: "Because buffered channels are faster.", Correct: false, Rationale: "Speed is not the point."},
					{Text: "Because §time.After§ requires a buffered channel.", Correct: false, Rationale: "The two are unrelated."},
					{Text: "It is unnecessary — the program works either way.", Correct: false, Rationale: "It works, but it leaks a goroutine. Under load that adds up."},
				},
				Explanation: "A buffer of one is the standard fix for 'maybe nobody is listening'.",
			},
			{
				ID:     "m2t3q3",
				Prompt: "What is the difference between §sync.Mutex§ and §sync.RWMutex§?",
				Options: []Option{
					{Text: "§RWMutex§ allows any number of concurrent readers, but a writer excludes everyone.", Correct: true, Rationale: "§RLock§ is shared; §Lock§ is exclusive."},
					{Text: "§RWMutex§ is always faster.", Correct: false, Rationale: "For write-heavy code the extra bookkeeping is a cost, not a win."},
					{Text: "§Mutex§ cannot be embedded in a struct.", Correct: false, Rationale: "Both are commonly the first field of a struct."},
					{Text: "§RWMutex§ protects data, §Mutex§ protects code.", Correct: false, Rationale: "Both protect data via critical sections."},
				},
				Explanation: "Read-heavy data (caches, registries) benefits from §RWMutex§. Write-heavy data is usually fine with a plain §Mutex§.",
			},
			{
				ID:     "m2t3q4",
				Prompt: "Which of these are true about sharing data between goroutines? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "Two goroutines writing the same map, even with different keys, is a data race.", Correct: true, Rationale: "The map's internal structure is shared, so the runtime can crash with \"concurrent map writes\"."},
					{Text: "A struct containing a §sync.Mutex§ must not be copied.", Correct: true, Rationale: "Copying the struct copies the lock, so two copies would guard different data. Pass a pointer."},
					{Text: "§go test -race§ can detect data races at run time.", Correct: true, Rationale: "It instruments memory access and reports conflicting ones with both stacks."},
					{Text: "Reading a variable from two goroutines while nobody writes is a data race.", Correct: false, Rationale: "Races require at least one write."},
					{Text: "Holding a mutex while doing a slow network call is good practice.", Correct: false, Rationale: "It blocks every other goroutine that needs the lock. Do I/O outside the critical section."},
				},
				Explanation: "Guard shared mutable state, keep critical sections short, and let §-race§ find what you miss.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m2t3e1",
				Title: "Warm-up: run something with a deadline",
				Prompt: `§§§go
func WithTimeout(f func() int, d time.Duration) (int, bool)
§§§

Runs §f§ in a goroutine and waits at most §d§ for it.

* If §f§ finishes in time, return its value and §true§.
* If the deadline passes first, return §0§ and §false§ — and do **not** wait for §f§ any
  longer.
* Use §select§, §time.After§ and a **buffered** result channel so the goroutine can
  finish without leaking.

The test runs a fast function with a generous deadline (expect §true§) and a slow one
with a tiny deadline (expect §false§).
`,
				Starter: `package main

import (
	"fmt"
	"time"
)

// WithTimeout runs f, giving up after d.
func WithTimeout(f func() int, d time.Duration) (int, bool) {
	// TODO
	return 0, false
}

func main() {
	fmt.Println(WithTimeout(func() int { return 7 }, time.Second))
	fmt.Println(WithTimeout(func() int { time.Sleep(200 * time.Millisecond); return 7 }, time.Millisecond))
}
`,
				Tests: []File{{
					Name: "withtimeout_test.go",
					Content: `package main

import (
	"testing"
	"time"
)

func TestWithTimeoutFast(t *testing.T) {
	got, ok := WithTimeout(func() int { return 42 }, time.Second)
	if !ok || got != 42 {
		t.Fatalf("WithTimeout(fast) = %d, %v; want 42, true", got, ok)
	}
}

func TestWithTimeoutSlow(t *testing.T) {
	start := time.Now()
	got, ok := WithTimeout(func() int {
		time.Sleep(500 * time.Millisecond)
		return 42
	}, 20*time.Millisecond)

	if ok {
		t.Fatal("WithTimeout reported success even though the deadline passed")
	}
	if got != 0 {
		t.Fatalf("WithTimeout returned %d, want 0", got)
	}
	if elapsed := time.Since(start); elapsed > 300*time.Millisecond {
		t.Fatalf("WithTimeout waited for the slow function (%v) instead of the deadline", elapsed)
	}
}

func TestWithTimeoutZeroValue(t *testing.T) {
	got, ok := WithTimeout(func() int { return 0 }, time.Second)
	if !ok || got != 0 {
		t.Fatalf("WithTimeout returning 0 legitimately = %d, %v; want 0, true", got, ok)
	}
}
`,
				}},
				Hints: []string{
					"§done := make(chan int, 1)§ — the §1§ is what prevents a leak.",
					"§go func() { done <- f() }()§",
					"Then §select { case v := <-done: return v, true; case <-time.After(d): return 0, false }§.",
					"Note that returning 0 and §true§ is different from returning 0 and §false§ — boolean §ok§ carries the meaning, exactly like every other §(v, ok)§ in Go.",
				},
				Solution: `package main

import (
	"fmt"
	"time"
)

func WithTimeout(f func() int, d time.Duration) (int, bool) {
	done := make(chan int, 1)
	go func() {
		done <- f()
	}()

	select {
	case v := <-done:
		return v, true
	case <-time.After(d):
		return 0, false
	}
}

func main() {
	fmt.Println(WithTimeout(func() int { return 7 }, time.Second))
	fmt.Println(WithTimeout(func() int { time.Sleep(200 * time.Millisecond); return 7 }, time.Millisecond))
}
`,
			},
			{
				ID:    "m2t3e2",
				Title: "A thread-safe counter",
				Prompt: `Implement the canonical guarded struct.

§§§go
type Counter struct {
	// your fields: a mutex and an int
}

func NewCounter() *Counter
func (c *Counter) Inc()
func (c *Counter) Add(n int)
func (c *Counter) Value() int
func (c *Counter) Reset()
§§§

* §Inc§ adds 1, §Add§ adds §n§, §Reset§ sets it back to 0.
* Every method must be safe to call from many goroutines at once.
* §NewCounter§ returns a **pointer** — a struct containing a mutex must never be copied,
  which is exactly what returning it by value would invite.

The test hammers the counter with 50 goroutines × 200 increments and expects exactly
10000. A missing lock shows up as a smaller number (or a race-detector failure).
`,
				Starter: `package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu sync.Mutex
	n  int
}

func NewCounter() *Counter {
	// TODO
	return nil
}

func (c *Counter) Inc() {
	// TODO
}

func (c *Counter) Add(n int) {
	// TODO
}

func (c *Counter) Value() int {
	// TODO
	return 0
}

func (c *Counter) Reset() {
	// TODO
}

func main() {
	c := NewCounter()
	c.Inc()
	c.Add(4)
	fmt.Println(c.Value())
}
`,
				Tests: []File{{
					Name: "counter_test.go",
					Content: `package main

import (
	"sync"
	"testing"
)

func TestCounterBasics(t *testing.T) {
	c := NewCounter()
	if c == nil {
		t.Fatal("NewCounter returned nil")
	}
	if c.Value() != 0 {
		t.Fatalf("new counter = %d, want 0", c.Value())
	}
	c.Inc()
	c.Inc()
	c.Add(5)
	if got := c.Value(); got != 7 {
		t.Fatalf("Value = %d, want 7", got)
	}
	c.Reset()
	if c.Value() != 0 {
		t.Fatalf("Value after Reset = %d, want 0", c.Value())
	}
}

func TestCounterConcurrent(t *testing.T) {
	c := NewCounter()
	const goroutines = 50
	const each = 200

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < each; i++ {
				c.Inc()
			}
		}()
	}
	wg.Wait()

	if got := c.Value(); got != goroutines*each {
		t.Fatalf("Value = %d, want %d — a missing lock loses updates", got, goroutines*each)
	}
}

func TestCounterConcurrentReaders(t *testing.T) {
	c := NewCounter()
	c.Add(100)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = c.Value()
			}
		}()
	}
	wg.Wait()
}
`,
				}},
				Hints: []string{
					"§NewCounter§ returns §&Counter{}§.",
					"Every method that touches §n§ needs §c.mu.Lock()§ and §defer c.mu.Unlock()§.",
					"§Inc§ can just call §c.Add(1)§ — but then only §Add§ takes the lock, so avoid double-locking.",
					"§Value§ could use a §sync.RWMutex§ and §RLock§ for extra credit; a plain §Mutex§ passes the tests.",
				},
				Solution: `package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu sync.Mutex
	n  int
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) Inc() {
	c.Add(1)
}

func (c *Counter) Add(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n += n
}

func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

func (c *Counter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n = 0
}

func main() {
	c := NewCounter()
	c.Inc()
	c.Add(4)
	fmt.Println(c.Value())
}
`,
			},
			{
				ID:    "m2t3e3",
				Title: "A guarded cache, and a fair merge",
				Prompt: `Two further uses of §RWMutex§ and §select§.

**1. A read-mostly cache (§sync.RWMutex§).**

§§§go
type Cache struct { /* a mutex and a map */ }

func NewCache() *Cache
func (c *Cache) Set(key, value string)
func (c *Cache) Get(key string) (string, bool)
func (c *Cache) Delete(key string) bool
func (c *Cache) Len() int
§§§

Reads use §RLock§ (many at once); writes use §Lock§ (exclusive). §NewCache§ must return a
pointer with the map already created — a nil map panics on the first write.

**2. A fair merge of two channels (§select§).**

§§§go
func Merge(a, b <-chan int) <-chan int
§§§

Yields every value from both channels and closes when **both** are exhausted. Unlike
§FanIn§, you do not know how many channels there are in advance, and the test feeds both
simultaneously — so this is the §select§-in-a-loop shape, with the trick of setting a
channel to §nil§ once it is closed so that §select§ stops choosing it.
`,
				Starter: `package main

import (
	"fmt"
	"sync"
)

type Cache struct {
	mu    sync.RWMutex
	items map[string]string
}

func NewCache() *Cache {
	// TODO
	return nil
}

func (c *Cache) Set(key, value string) {
	// TODO
}

func (c *Cache) Get(key string) (string, bool) {
	// TODO
	return "", false
}

func (c *Cache) Delete(key string) bool {
	// TODO
	return false
}

func (c *Cache) Len() int {
	// TODO
	return 0
}

// Merge interleaves a and b, closing when both are exhausted.
func Merge(a, b <-chan int) <-chan int {
	out := make(chan int)
	// TODO
	return out
}

func main() {
	c := NewCache()
	c.Set("a", "1")
	v, ok := c.Get("a")
	fmt.Println(v, ok, c.Delete("a"), c.Len())
}
`,
				Tests: []File{{
					Name: "cache_merge_test.go",
					Content: `package main

import (
	"slices"
	"sync"
	"testing"
)

func TestCache(t *testing.T) {
	c := NewCache()
	if c == nil {
		t.Fatal("NewCache returned nil")
	}
	c.Set("a", "1") // must not panic on a nil map

	if v, ok := c.Get("a"); !ok || v != "1" {
		t.Fatalf("Get = %q, %v; want \"1\", true", v, ok)
	}
	if _, ok := c.Get("missing"); ok {
		t.Fatal("Get on a missing key reported ok")
	}
	c.Set("a", "2")
	if v, _ := c.Get("a"); v != "2" {
		t.Fatalf("Set did not overwrite: %q", v)
	}
	if !c.Delete("a") {
		t.Error("Delete of an existing key returned false")
	}
	if c.Delete("a") {
		t.Error("Delete of a missing key returned true")
	}
	if c.Len() != 0 {
		t.Errorf("Len = %d, want 0", c.Len())
	}
}

func TestCacheConcurrent(t *testing.T) {
	c := NewCache()
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				c.Set("shared", "v")
				c.Get("shared")
				c.Len()
			}
		}(i)
	}
	wg.Wait()

	if c.Len() != 1 {
		t.Fatalf("Len = %d, want 1", c.Len())
	}
}

func collect(ch <-chan int) []int {
	var out []int
	for v := range ch {
		out = append(out, v)
	}
	return out
}

func TestMerge(t *testing.T) {
	a := make(chan int, 3)
	b := make(chan int, 2)
	for i := 1; i <= 3; i++ {
		a <- i
	}
	for i := 10; i <= 11; i++ {
		b <- i
	}
	close(a)
	close(b)

	got := collect(Merge(a, b))
	slices.Sort(got)
	if !slices.Equal(got, []int{1, 2, 3, 10, 11}) {
		t.Fatalf("Merge = %v, want every value from both channels", got)
	}
}

func TestMergeOneSideEmpty(t *testing.T) {
	a := make(chan int, 2)
	a <- 1
	a <- 2
	close(a)

	b := make(chan int)
	close(b)

	got := collect(Merge(a, b))
	slices.Sort(got)
	if !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("Merge = %v, want [1 2]", got)
	}
}
`,
				}},
				Hints: []string{
					"Cache: §NewCache§ returns §&Cache{items: map[string]string{}}§.",
					"§Get§ and §Len§ use §RLock§/§RUnlock§; §Set§ and §Delete§ use §Lock§/§Unlock§.",
					"Merge: launch one goroutine, then loop §for a != nil || b != nil§ with a §select§.",
					"In the select's §a§ case, §if !ok { a = nil; continue }§ — that disables the case so §select§ never picks it again.",
					"Close §out§ with §defer close(out)§ at the top of the goroutine.",
				},
				Solution: `package main

import (
	"fmt"
	"sync"
)

type Cache struct {
	mu    sync.RWMutex
	items map[string]string
}

func NewCache() *Cache {
	return &Cache{items: map[string]string{}}
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.items[key]
	return v, ok
}

func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.items[key]; !ok {
		return false
	}
	delete(c.items, key)
	return true
}

func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func Merge(a, b <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for a != nil || b != nil {
			select {
			case v, ok := <-a:
				if !ok {
					a = nil
					continue
				}
				out <- v
			case v, ok := <-b:
				if !ok {
					b = nil
					continue
				}
				out <- v
			}
		}
	}()
	return out
}

func main() {
	c := NewCache()
	c.Set("a", "1")
	v, ok := c.Get("a")
	fmt.Println(v, ok, c.Delete("a"), c.Len())
}
`,
			},
		},
	}
}

func m2t4() Topic {
	return Topic{
		ID:      "m2t4",
		Title:   "context.Context: cancellation and deadlines",
		Summary: "How a long request tells everything below it to stop.",
		Theory: `
A request that queries a database, calls a third-party API and writes a log line can be
abandoned at any point: the client disconnects, or a deadline passes. §context.Context§
carries that cancellation signal down the call chain, so each layer knows to stop.

## The shape

Every function that might block for a while takes a §context.Context§ as its **first
parameter**, named §ctx§:

§§§go
func FetchUser(ctx context.Context, id int) (*User, error)
§§§

The context flows *down*: §main§ creates one, hands it to the handler, which hands it to
the service, which hands it to the database driver. Cancelling at the top unblocks
everything at the bottom.

**Do not** store a context in a struct field. Pass it. (**A rare exception:** §*http.Request§
carries one for you, reachable as §r.Context()§.)

## Creating one

§§§go
ctx := context.Background()                 // the root: never cancelled
ctx := context.TODO()                       // "I will wire this up later"

ctx, cancel := context.WithCancel(parent)   // cancellable by hand
defer cancel()

ctx, cancel := context.WithTimeout(parent, 3*time.Second) // deadline
defer cancel()

ctx, cancel := context.WithDeadline(parent, someTime)
defer cancel()
§§§

**Always §defer cancel()§.** Even when the work finishes early. It releases the timer
and the child's resources; §go vet§'s §lostcancel§ check will tell you off if you forget.

## Detecting cancellation

Inside a long job, either listen on §ctx.Done()§:

§§§go
select {
case v := <-resultCh:
	return v, nil
case <-ctx.Done():
	return nil, ctx.Err()   // context.Canceled, or context.DeadlineExceeded
}
§§§

or poll the cheap non-blocking check in a loop:

§§§go
for _, row := range rows {
	if err := ctx.Err(); err != nil {
		return err
	}
	process(row)
}
§§§

§ctx.Err()§ is §nil§ while the context is alive. When it is finished it returns exactly
one of two sentinel errors:

| Value | Means |
| --- | --- |
| §context.Canceled§ | somebody called §cancel()§ |
| §context.DeadlineExceeded§ | the deadline passed |

You test them with §errors.Is§, the tool from Module 1:

§§§go
if errors.Is(err, context.DeadlineExceeded) {
	// tell the user 504 rather than 500
}
§§§

## The real reason it matters

§*http.Request§ already carries a context that is cancelled the moment the client
disconnects. Use it as the parent and a closed browser tab stops your database query:

§§§go
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	user, err := h.store.GetUser(ctx, id) // cancelled with the request
	if err != nil {
		// distinguish a timeout from a real failure
	}
}
§§§

Every blocking standard-library call has a §...Context§ variant: §db.QueryContext§,
§http.NewRequestWithContext§, §exec.CommandContext§. **Prefer them.** The
non-context versions cannot be cancelled, so a hung dependency hangs your handler
forever.

## Values: the §WithValue§ escape hatch

A context can also carry a small amount of data across API boundaries — a request ID, an
authenticated user, a tracing span.

§§§go
type ctxKey int // an unexported type, so no other package can collide

const userIDKey ctxKey = 0

func WithUserID(ctx context.Context, id int) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserID(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIDKey).(int)
	return id, ok
}
§§§

Rules that matter:

* The key must be an **unexported type** (§ctxKey§ above), never a §string§. If two
  packages both used §context.WithValue(ctx, "user", …)§ they would silently overwrite
  each other.
* Always provide typed accessors, so callers never see §any§.
* Use it for **request-scoped** data only. Configuration, database handles and loggers
  are function parameters or struct fields, not context values.

## A complete example

§§§go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// work pretends to do something slow that respects cancellation.
func work(ctx context.Context, steps int) (int, error) {
	done := 0
	for i := 0; i < steps; i++ {
		select {
		case <-ctx.Done():
			return done, ctx.Err()
		case <-time.After(20 * time.Millisecond):
			done++
		}
	}
	return done, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done, err := work(ctx, 100)
	fmt.Println("steps completed:", done)
	fmt.Println("error:", err)
	fmt.Println("was a timeout:", errors.Is(err, context.DeadlineExceeded))
}
§§§

Output:

§§§text
steps completed: 5
error: context deadline exceeded
was a timeout: true
§§§

The worker stopped *itself* the moment the deadline passed — no goroutine was left
running, and the caller learned exactly why.

## Summary

* Contexts carry cancellation and deadlines **down** a call chain.
* Always pass §ctx§ first; never store it in a struct.
* Always §defer cancel()§.
* Prefer the §...Context§ variant of any blocking API.
* §WithValue§ is for request-scoped data with an unexported key type — not for
  configuration.
`,
		Quiz: []Question{
			{
				ID:     "m2t4q1",
				Prompt: "Which statements about §context§ are correct? Select all that apply.",
				Multi:  true,
				Options: []Option{
					{Text: "The §cancel§ returned by §WithTimeout§ should be called even on the success path.", Correct: true, Rationale: "It releases the timer and the child's resources. §defer cancel()§ right after creation is the idiom."},
					{Text: "A context should be stored in a struct field for later use.", Correct: false, Rationale: "Pass it explicitly as the first parameter."},
					{Text: "Cancelling a parent cancels every context derived from it.", Correct: true, Rationale: "That is the propagation guarantee that makes the whole design work."},
					{Text: "§context.WithValue§ is the recommended way to pass a logger or a database handle.", Correct: false, Rationale: "Those belong in struct fields or parameters. §WithValue§ is for request-scoped data."},
					{Text: "Passing §nil§ as a §Context§ is equivalent to §context.Background()§.", Correct: false, Rationale: "It is a bug; the receiver will panic. Use §context.TODO()§ if you must defer the decision."},
				},
				Explanation: "Contexts flow down, cancel downward, and hold only request-scoped values.",
			},
			{
				ID:     "m2t4q2",
				Prompt: "What is printed?",
				Code: `package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	var err error
	select {
	case <-time.After(time.Second):
		err = errors.New("finished")
	case <-ctx.Done():
		err = ctx.Err()
	}
	fmt.Println(errors.Is(err, context.DeadlineExceeded))
}`,
				Options: []Option{
					{Text: "§true§", Correct: true, Rationale: "The 10 ms deadline fires long before the one-second timer, and §ctx.Err()§ is exactly §DeadlineExceeded§."},
					{Text: "§false§", Correct: false, Rationale: "§ctx.Err()§ returns that sentinel when the deadline passes, and §errors.Is§ matches it."},
					{Text: "§<nil>§ — §ctx.Err()§ is nil until something reads it.", Correct: false, Rationale: "Once the deadline has passed, §Err()§ is non-nil."},
					{Text: "It does not compile.", Correct: false, Rationale: "Everything here is valid."},
				},
				Explanation: "§ctx.Err()§ is §Canceled§ after an explicit cancel, §DeadlineExceeded§ after a timeout, and §nil§ while active.",
			},
			{
				ID:     "m2t4q3",
				Prompt: "Why should §store.GetUser(ctx, id)§ take a context rather than creating its own timeout?",
				Options: []Option{
					{Text: "So the caller's deadline and request lifetime propagate into the database call.", Correct: true, Rationale: "The caller knows the overall budget; the callee should honour it rather than invent a nested policy."},
					{Text: "Because the database driver requires it.", Correct: false, Rationale: "The non-context methods exist; they simply cannot be cancelled."},
					{Text: "To make the function faster.", Correct: false, Rationale: "It is about cancellation and resource release, not speed."},
					{Text: "Because §context.Context§ replaces the §error§ return value.", Correct: false, Rationale: "You still return errors; context only adds cancellation."},
				},
				Explanation: "One deadline covers the whole path: HTTP → service → database → third-party API.",
			},
		},
		Exercises: []Exercise{
			{
				ID:    "m2t4e1",
				Title: "Warm-up: a cancellable countdown",
				Prompt: `§§§go
func CountDown(ctx context.Context, n int) <-chan int
func Collect(ctx context.Context, in <-chan int) ([]int, error)
§§§

* §CountDown§ returns a channel yielding §n, n-1, …, 1§. It must be **closed** when the
  countdown finishes *or* when §ctx§ is done — the goroutine must never leak.
* §Collect§ drains the channel into a slice. If the context finishes first, it returns
  the values gathered so far **and** §ctx.Err()§.
* §n <= 0§ produces an immediately closed channel.

Use a §select§ over §ctx.Done()§ in both directions of the pipe. This is §select§ and
§context§ working together, which is how production code looks.
`,
				Starter: `package main

import (
	"context"
	"fmt"
	"time"
)

// CountDown yields n, n-1, ..., 1, stopping if ctx is cancelled.
func CountDown(ctx context.Context, n int) <-chan int {
	out := make(chan int)
	// TODO
	return out
}

// Collect drains in, returning early if ctx is done.
func Collect(ctx context.Context, in <-chan int) ([]int, error) {
	// TODO
	return nil, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	vals, err := Collect(ctx, CountDown(ctx, 1000000))
	fmt.Println("collected", len(vals), "values, err:", err)
}
`,
				Tests: []File{{
					Name: "countdown_test.go",
					Content: `package main

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestCountDown(t *testing.T) {
	got, err := Collect(context.Background(), CountDown(context.Background(), 5))
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if !reflect.DeepEqual(got, []int{5, 4, 3, 2, 1}) {
		t.Fatalf("CountDown(5) = %v", got)
	}
}

func TestCountDownZero(t *testing.T) {
	got, err := Collect(context.Background(), CountDown(context.Background(), 0))
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("CountDown(0) = %v, want empty", got)
	}
}

func TestCountDownCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := CountDown(ctx, 1000000)

	if first := <-ch; first != 1000000 {
		t.Fatalf("first value = %d, want 1000000", first)
	}
	cancel()

	// The producer must notice the cancellation and close the channel.
	drained := make(chan struct{})
	go func() {
		for range ch {
		}
		close(drained)
	}()

	select {
	case <-drained:
	case <-time.After(2 * time.Second):
		t.Fatal("CountDown leaked: the channel was never closed after cancel")
	}
}

func TestCollectCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// A channel that never delivers: the only ready case is ctx.Done().
	blocked := make(chan int)
	if _, err := Collect(ctx, blocked); err == nil {
		t.Fatal("Collect with a cancelled context should return an error")
	}
}

func TestCollectTimeout(t *testing.T) {
	slow := make(chan int)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := Collect(ctx, slow)
	if err != context.DeadlineExceeded {
		t.Fatalf("Collect err = %v, want exactly context.DeadlineExceeded", err)
	}
}
`,
				}},
				Hints: []string{
					"§CountDown§: §go func() { defer close(out); for i := n; i > 0; i-- { select { case out <- i: case <-ctx.Done(): return } } }()§.",
					"§defer close(out)§ at the top is what guarantees no leak on the §return§ path.",
					"§Collect§: loop with §select { case <-ctx.Done(): return out, ctx.Err(); case v, ok := <-in: ... }§.",
					"Return §ctx.Err()§ itself, unwrapped — the test compares with §==§, not §errors.Is§.",
				},
				Solution: `package main

import (
	"context"
	"fmt"
	"time"
)

func CountDown(ctx context.Context, n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := n; i > 0; i-- {
			select {
			case out <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func Collect(ctx context.Context, in <-chan int) ([]int, error) {
	var out []int
	for {
		select {
		case <-ctx.Done():
			return out, ctx.Err()
		case v, ok := <-in:
			if !ok {
				return out, nil
			}
			out = append(out, v)
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	vals, err := Collect(ctx, CountDown(ctx, 1000000))
	fmt.Println("collected", len(vals), "values, err:", err)
}
`,
			},
			{
				ID:    "m2t4e2",
				Title: "Request-scoped values, done properly",
				Prompt: `Implement the typed accessor pair that most services need.

§§§go
type User struct {
	ID    int
	Email string
}

func WithRequestID(ctx context.Context, id string) context.Context
func RequestID(ctx context.Context) (string, bool)

func WithUser(ctx context.Context, u User) context.Context
func UserFrom(ctx context.Context) (User, bool)
§§§

Requirements:

* The context keys must be of an **unexported type**, so no other package can collide
  with them. Two §const§ values of a §type ctxKey int§ are the standard shape.
* §RequestID§ and §UserFrom§ must never panic — not on §context.Background()§, and not
  on a §nil§ context.
* They return §false§ when the value is absent *or* of the wrong type. A value stored
  under a plain §string§ key must **not** be found.
* Storing two values must not lose either one — contexts form a chain.
`,
				Starter: `package main

import (
	"context"
	"fmt"
)

type ctxKey int

const (
	requestIDKey ctxKey = iota
	userKey
)

type User struct {
	ID    int
	Email string
}

func WithRequestID(ctx context.Context, id string) context.Context {
	// TODO
	return ctx
}

func RequestID(ctx context.Context) (string, bool) {
	// TODO
	return "", false
}

func WithUser(ctx context.Context, u User) context.Context {
	// TODO
	return ctx
}

func UserFrom(ctx context.Context) (User, bool) {
	// TODO
	return User{}, false
}

func main() {
	ctx := WithRequestID(context.Background(), "req-1")
	ctx = WithUser(ctx, User{ID: 7, Email: "a@b.c"})
	fmt.Println(RequestID(ctx))
	fmt.Println(UserFrom(ctx))
	fmt.Println(RequestID(context.Background()))
}
`,
				Tests: []File{{
					Name: "ctxvalue_test.go",
					Content: `package main

import (
	"context"
	"testing"
)

func TestRequestID(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req-42")
	id, ok := RequestID(ctx)
	if !ok || id != "req-42" {
		t.Fatalf("RequestID = %q, %v", id, ok)
	}
	if _, ok := RequestID(context.Background()); ok {
		t.Fatal("RequestID on an empty context should report false")
	}
}

func TestUser(t *testing.T) {
	u := User{ID: 7, Email: "a@b.c"}
	ctx := WithUser(context.Background(), u)

	got, ok := UserFrom(ctx)
	if !ok || got != u {
		t.Fatalf("UserFrom = %+v, %v", got, ok)
	}
	if _, ok := UserFrom(context.Background()); ok {
		t.Fatal("UserFrom on an empty context should report false")
	}
}

func TestBothValuesCoexist(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req-1")
	ctx = WithUser(ctx, User{ID: 1})

	if id, _ := RequestID(ctx); id != "req-1" {
		t.Fatalf("RequestID was lost: %q", id)
	}
	if u, ok := UserFrom(ctx); !ok || u.ID != 1 {
		t.Fatalf("UserFrom was lost: %+v %v", u, ok)
	}
}

func TestForeignKeyDoesNotMatch(t *testing.T) {
	ctx := context.WithValue(context.Background(), "requestID", 12345)
	if _, ok := RequestID(ctx); ok {
		t.Fatal("a value stored under a string key must not be found by RequestID")
	}

	ctx2 := context.WithValue(context.Background(), "user", "not a User")
	if _, ok := UserFrom(ctx2); ok {
		t.Fatal("a value of the wrong type must not be found by UserFrom")
	}
}

func TestNilContextDoesNotPanic(t *testing.T) {
	if _, ok := RequestID(nil); ok {
		t.Error("RequestID(nil) should report false")
	}
	if _, ok := UserFrom(nil); ok {
		t.Error("UserFrom(nil) should report false")
	}
}
`,
				}},
				Hints: []string{
					"§return context.WithValue(ctx, requestIDKey, id)§ and the same shape for the user.",
					"Look up with the comma-ok type assertion: §id, ok := ctx.Value(requestIDKey).(string)§.",
					"Guard the nil context: §if ctx == nil { return ..., false }§ — §ctx.Value§ on a nil context panics.",
					"The key type being unexported is what makes a value stored under §\"requestID\"§ invisible to you.",
				},
				Solution: `package main

import (
	"context"
	"fmt"
)

type ctxKey int

const (
	requestIDKey ctxKey = iota
	userKey
)

type User struct {
	ID    int
	Email string
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestID(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}

func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func UserFrom(ctx context.Context) (User, bool) {
	if ctx == nil {
		return User{}, false
	}
	u, ok := ctx.Value(userKey).(User)
	return u, ok
}

func main() {
	ctx := WithRequestID(context.Background(), "req-1")
	ctx = WithUser(ctx, User{ID: 7, Email: "a@b.c"})
	fmt.Println(RequestID(ctx))
	fmt.Println(UserFrom(ctx))
	fmt.Println(RequestID(context.Background()))
}
`,
			},
		},
	}
}
