//go:build unit

package threadtest

import (
	"sync/atomic"
	"time"
)

var leafCounter atomic.Uint32
var rootCounter atomic.Uint32

func ResetCounters() {
	leafCounter = atomic.Uint32{}
	rootCounter = atomic.Uint32{}
}

func GetLeafCounter() int {
	return int(leafCounter.Load())
}

func GetRootCounter() int {
	return int(rootCounter.Load())
}

type LeafAtomicSingleton int

func (a *LeafAtomicSingleton) SummonSingleton() *LeafAtomicSingleton {
	val := leafCounter.Add(1)
	time.Sleep(time.Millisecond * 10)
	return new(LeafAtomicSingleton(val))
}

type RootAtomicSingleton struct {
	a *LeafAtomicSingleton
}

func (r *RootAtomicSingleton) SummonSingleton(a *LeafAtomicSingleton) *RootAtomicSingleton {
	rootCounter.Add(1)
	time.Sleep(time.Millisecond * 10)
	return new(RootAtomicSingleton{a})
}
