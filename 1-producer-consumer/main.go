//////////////////////////////////////////////////////////////////////
//
// Given is a producer-consumer scenario, where a producer reads in
// tweets from a mockstream and a consumer is processing the
// data. Your task is to change the code so that the producer as well
// as the consumer can run concurrently
//

package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

func producer(stream Stream, tweetChan chan *Tweet, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		tweet, err := stream.Next()
		if errors.Is(err, ErrEOF) {
			close(tweetChan)
			return
		}

		tweetChan <- tweet
	}
}

func consumer(tweetChan chan *Tweet, wg *sync.WaitGroup) {
	defer wg.Done()
	for t := range tweetChan {
		if t.IsTalkingAboutGo() {
			fmt.Println(t.Username, "\ttweets about golang")
		} else {
			fmt.Println(t.Username, "\tdoes not tweet about golang")
		}
	}
}

func main() {
	start := time.Now()
	stream := GetMockStream()

	tweetChan := make(chan *Tweet)

	var wg sync.WaitGroup
	wg.Add(2)

	// Producer
	go producer(stream, tweetChan, &wg)

	// Consumer
	go consumer(tweetChan, &wg)

	wg.Wait()

	fmt.Printf("Process took %s\n", time.Since(start))
}
