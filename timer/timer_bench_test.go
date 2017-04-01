package timer

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkCreateTask(b *testing.B) {
	timer := NewTimer(1000, 59, time.Millisecond)
	go timer.Start()

	for i := 0; i < b.N; i++ {
		timer.Task(uint(rand.Intn(10000)))
	}
}
