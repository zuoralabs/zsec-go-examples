package zsec_go_examples_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func doWork(from string, mm map[string]string, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 3; i++ {
		fmt.Println(from, ":", i)
	}
	mm[from] = "hi"
}

func TestBasicWaitGroups(t *testing.T) {
	wg := new(sync.WaitGroup)

	mm := make(map[string]string)

	wg.Add(1)
	doWork("direct", mm, wg)

	wg.Add(1)
	go doWork("goroutine", mm, wg)

	wg.Add(1)
	go func(msg string, mm map[string]string, wg *sync.WaitGroup) {
		defer wg.Done()
		fmt.Println(msg)
		mm[msg] = "hi"
	}("going", mm, wg)

	_, _ = fmt.Scanln()
	wg.Wait()
	fmt.Println("done")

	fmt.Printf("%v\n", mm)
}

func startAGoroutine(from string, wg *sync.WaitGroup) chan string {
	defer wg.Done()
	cc := make(chan string)

	wg.Add(1)
	go func() {
		defer wg.Done()
		println("started work")
		time.Sleep(time.Millisecond * 200)
		println("finishing work")
		cc <- from
	}()

	return cc
}

func TestBasicChannels(t *testing.T) {
	var wg sync.WaitGroup
	//wg := new(sync.WaitGroup)  // could also use new to ensure you use pointers to wg

	cc := make(chan string)
	go func() { cc <- "hi\n" }()
	print(<-cc)

	wg.Add(1)
	cc2 := startAGoroutine("hi2\n", &wg) // you must use a pointer to wg (due to pass by value -- 60% sure)
	print(<-cc2)                         // goroutine waits for us to read from channel.
	wg.Wait()

	fmt.Println("done")
}

func TestLimitConcurrentRoutines(t *testing.T) {
	maxConcurrent := 2

	// limits to 2
	canGo := make(chan bool, maxConcurrent)

	results := make(chan int, maxConcurrent)

	for i := 0; i < maxConcurrent; i++ {
		canGo <- true
	}

	numGoRoutines := 0

	for i := 0; i < 100; i++ {
		numGoRoutines += 1
		//println(i)
		go func(i int) {
			<-canGo
			time.Sleep(time.Second)
			results <- i
			canGo <- true
		}(i)
	}

	for i := 0; i < numGoRoutines; i++ {
		fmt.Printf("result: %d\n", <-results)
	}
}

func TestMapReducePattern(t *testing.T) {
	NUM_SOURCES := 100
	NUM_WORKERS := 100 /*
		in this example all the workers are the same, so it's redundant to have
		NUM_WORKERS > MAX_CONCURRENT.
		However, you could imagine a scenario with multiple pipelines and workers
		specific to each pipeline.
	*/
	MAX_CONCURRENT := 5 // in case we want to limit the amount of resources used without limiting number of workers

	canGo := make(chan bool, MAX_CONCURRENT) // Used to limit number of concurrent workers

	for i := 0; i < MAX_CONCURRENT; i++ {
		canGo <- true
	}

	workerInput := make(chan int, 5)
	workerOutput := make(chan int, 5)

	sources := new(sync.WaitGroup)
	workers := new(sync.WaitGroup)

	// start sources. Just to make it interesting, I have two layers of
	// goroutines. Note both inner and outer ones are waited on.
	sources.Add(1)
	go func() {
		defer sources.Done()
		for i := 0; i < NUM_SOURCES; i++ {
			sources.Add(1)
			go func(id int, input int) {
				defer sources.Done()
				fmt.Printf("Source %v started.\n", id)
				workerInput <- input
			}(i, i)
		}
	}()

	// start workers
	for i := 0; i < NUM_WORKERS; i++ {
		workers.Add(1)
		go func(id int) {
			defer workers.Done()
			for input := range workerInput {
				<-canGo
				fmt.Printf("Worker %v started.\n", id)
				time.Sleep(time.Millisecond * 200) // simulate doing work
				workerOutput <- input * input
				canGo <- true
			}
		}(i)
	}

	go func() {
		sources.Wait()
		close(workerInput) // workers will only stop when workerInput is closed
		workers.Wait()
		close(workerOutput) // the collector loop (below) will only stop when workerInput is closed
	}()

	// start collector loop
	count := 0
	total := 0
	for ii := range workerOutput {
		count += 1
		total += ii
	}

	finalResult := total
	assert.Equal(t, NUM_SOURCES, count)
	assert.Equal(t, 328350, finalResult)
	fmt.Printf("final result is %v.\n", finalResult)
}

func Test4(t *testing.T) {
	maxConcurrent := 2
	numWorkers := 50
	numJobs := 10

	cc := make(chan int, 5)
	canGo := make(chan bool, maxConcurrent)
	for i := 0; i < maxConcurrent; i++ {
		canGo <- true
	}

	jobs := new(sync.WaitGroup)
	jobs.Add(1)
	go func() {
		defer jobs.Done()
		for i := 0; i < numJobs; i++ {
			jobs.Add(1)
			cc <- i
		}
	}()

	workers := new(sync.WaitGroup)
	for i := 0; i < numWorkers; i++ {
		workers.Add(1)
		go func(i int) {
			defer workers.Done()
			fmt.Printf("starting: w%v\n", i)
			for input := range cc {
				<-canGo
				fmt.Printf("working:  w%v on j%v\n", i, input)
				time.Sleep(time.Millisecond * 200)
				jobs.Done()
				canGo <- true
			}
			time.Sleep(time.Millisecond * 200)
			fmt.Printf("stopping: w%v\n", i)
		}(i)
	}

	jobs.Wait()
	println("jobs complete")
	close(cc)
	workers.Wait()
	println("workers complete")
}

func TestLoopVariableCapture(t *testing.T) {
	// Gotcha: loop variables are captured by the closure (Python programmers will recognize this gotcha)

	numWorkers := 5
	workers := new(sync.WaitGroup)
	var mm1 sync.Map
	var mm2 sync.Map
	for i := 0; i < numWorkers; i++ {
		workers.Add(1)
		go func(k int) {
			defer workers.Done()
			time.Sleep(time.Millisecond * 200)
			fmt.Printf("part 1: working: %v\n", i)
			mm1.Store(k, i)
		}(i)
	}

	// NOTE: this range style is frowned upon; don't use in production
	for i := range make([]int, numWorkers) {
		workers.Add(1)
		go func(k int) {
			defer workers.Done()
			time.Sleep(time.Millisecond * 200)
			fmt.Printf("part 2: working: %v\n", i)
			mm2.Store(k, i)
		}(i)
	}

	workers.Wait()
	for count, mm := range []sync.Map{mm1, mm2} {
		// NOTE: this range style is frowned upon
		for i := range make([]int, numWorkers) {
			v, _ := mm.Load(i)
			v2 := v.(int)
			fmt.Printf("part %v: worker %v got %v\n", count+1, i, v2)
		}
	}
}
