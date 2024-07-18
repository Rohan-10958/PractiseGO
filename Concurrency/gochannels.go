package Concurrency

import (
	"sync"
	"time"
)

func findDigits(num int, dg chan int) {
	for num != 0 {
		var digit int = num % 10
		dg <- digit
		num = num / 10
	}
	close(dg)
}

func workerFindSum(dg chan int, wg *sync.WaitGroup, sumc chan int) {
	localsum := 0
	for v := range dg {
		localsum += v
	}
	sumc <- localsum
	wg.Done()
}
func allocateWorkersFindSum(done chan int, dg chan int, numOfWorkers int) {
	var wg sync.WaitGroup
	sumc := make(chan int)
	finalsum := make(chan int)
	go func(finallsum chan int) {
		sum := 0
		for val := range sumc {
			sum += val
		}
		finalsum <- sum
		close(finallsum)
	}(finalsum)

	for i := 0; i < numOfWorkers; i++ {
		wg.Add(1)
		go workerFindSum(dg, &wg, sumc)
	}

	wg.Wait()
	close(sumc)
	done <- <-finalsum
}

func FindSumUsingNWorkers(num int, numOfWorkers int) (int, time.Duration) {
	start := time.Now()
	dg := make(chan int)
	sum := make(chan int)
	go findDigits(num, dg)

	go allocateWorkersFindSum(sum, dg, numOfWorkers)

	return <-sum, time.Since(start)

}
