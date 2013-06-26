package rmutex

import (
	"runtime"
	"testing"
)

type expectState struct {
	owner     Token
	recursion int32
}

// a helper function
func checkState(t *testing.T, r *Rmutex, expect *expectState) {
	if r.owner != int32(expect.owner) {
		t.Fatalf("owner %d is not the expected one(%d)\n",
			r.owner, expect.owner)
	}
	if r.recursion != expect.recursion {
		t.Fatalf("recursion %d is not the expected one(%d)\n",
			r.recursion, expect.recursion)
	}
}

func TestSingle(t *testing.T) {
	const level = 1000
	lock := NewRmutex()
	token := NewToken()

	// recursion lock
	for i := 0; i < level; i++ {
		lock.Lock(token)
		lock.Lock(token)
	}
	checkState(t, lock, &expectState{token, 2 * level})

	// recursion unlock
	for i := 0; i < level; i++ {
		lock.Unlock(token)
	}
	checkState(t, lock, &expectState{token, level})
	for i := 0; i < level; i++ {
		lock.Unlock(token)
	}
	checkState(t, lock, &expectState{0, 0})
}

func TestMultiply(t *testing.T) {
	const (
		loop  = 1000
		procs = 4
	)
	runtime.GOMAXPROCS(procs)
	lock := NewRmutex()
	done := make(chan bool)
	for i := 0; i < procs; i++ {
		go func() {
			token := NewToken()
			for i := 0; i < loop; i++ {
				lock.Lock(token)
				lock.Lock(token)
				// do some work
				for j := 0; j < loop; j++ {
					i += 0
				}
				lock.Unlock(token)
				lock.Lock(token)
				lock.Unlock(token)
				lock.Unlock(token)
			}
			done <- true
		}()
	}
	// wait done
	for i := 0; i < procs; i++ {
		<-done
	}
	checkState(t, lock, &expectState{0, 0})
}
