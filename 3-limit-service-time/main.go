//////////////////////////////////////////////////////////////////////
//
// Your video processing service has a freemium model. Everyone has 10
// sec of free processing time on your service. After that, the
// service will kill your process, unless you are a paid premium user.
//
// Beginner Level: 10s max per request
// Advanced Level: 10s max per user (accumulated)
//

package main

import (
	"sync/atomic"
	"time"
)

// User defines the UserModel. Use this to check whether a User is a
// Premium user or not
type User struct {
	ID        int
	IsPremium bool
	TimeUsed  atomic.Int64 // in seconds
}

// HandleRequest runs the processes requested by users. Returns false
// if process had to be killed
func HandleRequest(process func(), u *User) bool {
	done := make(chan int64, 1)
	go func() {
		start := time.Now()
		process()
		timePassed := int64(time.Now().Sub(start).Seconds())

		done <- timePassed
	}()

	select {
	case consumedTime := <-done:
		u.TimeUsed.Add(consumedTime)

		return u.TimeUsed.Load() <= 10 || u.IsPremium
	}
}

func main() {
	RunMockServer()
}
