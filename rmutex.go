package rmutex

import (
	"fmt"
	_ "sync"
	"sync/atomic"
)

// Token used to identify the caller
type Token int32

var tokenGender int32 = 0

func getToken() int32 {
	return atomic.AddInt32(&tokenGender, 1)
}

// Get a newly generated token
func NewToken() Token {
	return Token(getToken())
}

type Rmutex struct {
	sema      chan bool
	owner     Token // current lock owner
	recursion int32 //current resursion level
	counter   int32
}

func NewRmutex() *Rmutex {
	return &Rmutex{
		sema: make(chan bool),
	}
}

func (r *Rmutex) Lock(token Token) {
	if atomic.AddInt32(&r.counter, 1) > 1 {
		if r.owner != token {
			<-r.sema
		}
	}
	// we are now inside the lock
	r.owner = token
	r.recursion++
}

func (r *Rmutex) Unlock(token Token) {
	if token != r.owner {
		panic(fmt.Sprintf("you are not the owner(%d): %d!", r.owner, token))
	}
	r.recursion--
	recur := r.recursion
	if recur == 0 {
		r.owner = 0 //default init value
	}
	if atomic.AddInt32(&r.counter, -1) > 0 {
		if recur == 0 {
			r.sema <- true
		}
	}
	// we are outside the lock
}

func (r *Rmutex) Trylock(token Token) bool {
	if token == r.owner {
		// already inside the lock
		atomic.AddInt32(&r.counter, 1)
	} else {
		if !atomic.CompareAndSwapInt32(&r.counter, 0, 1) {
			// we are not the first one to grasp the lock
			return false
		}
		// we are the first one
		r.owner = token
	}
	r.recursion++
	return true
}
