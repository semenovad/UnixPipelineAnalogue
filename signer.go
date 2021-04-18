package main

import (
	"fmt"
	"sort"
	"sync"
)

const (
	iterationsNum = 6
)

func startWorker(in, out chan interface{}, job func(in, out chan interface{}), waiter *sync.WaitGroup) {
	defer waiter.Done()
	job(in, out)
	close(out)
}

//ExecutePipeline pipe
func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	out := make(chan interface{})
	wg.Add(len(jobs))
	for _, item := range jobs {
		go startWorker(in, out, item, wg)
		in = out
		out = make(chan interface{})
	}
	wg.Wait()
}

func printSingleHash(data, md5, crc32md5, crc32, str string) {
	fmt.Printf("%s SingleHash data %s\n", data, data)
	fmt.Printf("%s SinlgeHash md5(data) %s\n", data, md5)
	fmt.Printf("%s SinlgeHash crc32(md5(data)) %s\n", data, crc32md5)
	fmt.Printf("%s SinlgeHash crc32(data) %s\n", data, crc32)
	fmt.Printf("%s SinlgeHash result %s\n", data, str)
}
func crc32Worker(data string, ch chan string) {
	ch <- DataSignerCrc32(data)
}

// var start int
func singleHashWorker(data string, out chan interface{}, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	// start := time.Now()
	// fmt.Printf("   LOG:singleHashWorker:STEP1:%s\n", data)
	var md5 string
	crc32md5Chan := make(chan string)
	go func() {
		mu.Lock()
		md5 = DataSignerMd5(data)
		mu.Unlock()
		crc32Worker(md5, crc32md5Chan)
	}()

	crc32Chan := make(chan string)
	go crc32Worker(data, crc32Chan)

	crc32 := <-crc32Chan
	crc32md5 := <-crc32md5Chan
	str := crc32 + "~" + crc32md5
	// go printSingleHash(data, md5, crc32md5, crc32, str)
	// fmt.Printf("   time:%s:%s\n", data, time.Since(start))

	out <- str
}

// SingleHash cool
func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for data := range in {
		wg.Add(1)
		str := fmt.Sprintf("%v", data)
		go singleHashWorker(str, out, wg, mu)
	}

	wg.Wait()
}

func multiHashWorker(data string, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

	var chans = []chan string{
		make(chan string),
		make(chan string),
		make(chan string),
		make(chan string),
		make(chan string),
		make(chan string),
	}

	waiter := &sync.WaitGroup{}
	waiter.Add(iterationsNum)
	for th := 0; th < iterationsNum; th++ {
		go func(ch chan string, iter int) {
			defer waiter.Done()
			th := fmt.Sprintf("%v", iter)
			go crc32Worker(th+data, ch)
		}(chans[th], th)
	}
	waiter.Wait()

	var result string
	for th := 0; th < iterationsNum; th++ {
		tmp := <-chans[th]
		result += tmp
		// fmt.Printf("%s MultiHash: crc32(th+step1)) %d %s\n", data, th, tmp)
	}
	// fmt.Printf("%s MultiHash: result %s\n", data, result)
	out <- result
}

// MultiHash cool
func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for data := range in {
		wg.Add(1)
		str := fmt.Sprintf("%v", data)
		go multiHashWorker(str, out, wg)
	}

	wg.Wait()
}

// CombineResults cool
func CombineResults(in, out chan interface{}) {
	var data []string
	for item := range in {
		tmp := fmt.Sprintf("%v", item)
		data = append(data, tmp)
	}
	sort.Strings(data)
	size := len(data)
	var result = data[0]
	for i := 1; i < size; i++ {
		result += "_" + data[i]
	}
	out <- result
}

func main() {
}
