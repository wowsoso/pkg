package breaker

import (
	"testing"
	"time"
)

func TestBucketMethodReset(t *testing.T) {
	b := newBucket()

	b.succeed = 1
	b.failed = 1
	b.timeout = 1
	b.reject = 1

	b.reset()

	if b.succeed != 0 && b.failed != 0 && b.timeout != 0 && b.reject != 0 {
		t.Fatalf("reset error")
	}
}

func TestBreaker(t *testing.T) {
	metricsOptions := MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	NewBreaker(metricsOptions, IsOpen, IsClosed)
}

func TestBreakerMethodchangeStateWhenClosed2Open(t *testing.T) {
	metricsOptions := MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b := NewBreakerWithDefault(metricsOptions)

	b.state = CLOSED
	b.buckets[0].succeed = 1
	b.buckets[4].failed = 1
	b.changeState(OPEN)

	if b.state != OPEN {
		t.Fatalf("state should be open, got: %v", b.state)
	}

	for i := 0; i < int(b.MetricsRollingCount); i++ {
		if (b.buckets[i].succeed + b.buckets[i].failed + b.buckets[i].timeout + b.buckets[i].reject) != 0 {
			t.Fatal("set state open should be reset all bucket")
		}
	}

}

func TestBreakerMethodchangeStateWhenOpen2HalfOpen(t *testing.T) {
	metricsOptions := MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b := NewBreaker(metricsOptions, IsOpen, IsClosed)

	b.state = OPEN
	b.changeState(HALFOPEN)

	if b.state != HALFOPEN {
		t.Fatalf("state should be halfopen, got: %v", b.state)
	}
}

func TestBreakerMethodchangeStateWhenHalfOpen2Open(t *testing.T) {
	metricsOptions := MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b := NewBreaker(metricsOptions, IsOpen, IsClosed)

	b.state = HALFOPEN
	b.changeState(OPEN)

	if b.state != OPEN {
		t.Fatalf("state should be halfopen, got: %v", b.state)
	}
}

func TestBreakerMethodchangeStateWhenHalfOpen2Closed(t *testing.T) {
	metricsOptions := MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b := NewBreaker(metricsOptions, IsOpen, IsClosed)

	b.state = HALFOPEN
	b.changeState(CLOSED)

	if b.state != CLOSED {
		t.Fatalf("state should be halfopen, got: %v", b.state)
	}
}

func TestBreakerMethodReceiving(t *testing.T) {
	metricsOptions := MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b := NewBreaker(metricsOptions, IsOpen, IsClosed)

	tick := time.NewTicker(100 * time.Second)
	c := make(chan time.Time)
	tick.C = c
	go b.receiving(tick)

	b.c <- SUCCEED
	b.c <- FAILED
	b.c <- TIMEOUT
	b.c <- REJECT

	c <- time.Now()

	if b.buckets[4].succeed != 1 && b.buckets[4].failed != 1 && b.buckets[4].timeout != 1 && b.buckets[4].reject != 1 {
		t.Fatalf("breaker method error. succeed: %d failed: %d timeout: %d reject: %d", b.buckets[4].succeed, b.buckets[4].failed, b.buckets[4].timeout, b.buckets[4].reject)
	}

	b.Stop()
	select {
	case b.c <- SUCCEED:
		t.Fatalf("b.c should be blocked")
	default:
	}

	metricsOptions = MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b = NewBreaker(metricsOptions, IsOpen, IsClosed)

	tick = time.NewTicker(100 * time.Second)
	c = make(chan time.Time)
	tick.C = c
	b.state = HALFOPEN
	go b.receiving(tick)

	b.c <- SUCCEED
	b.c <- FAILED
	b.c <- TIMEOUT
	b.c <- REJECT

	c <- time.Now()

	if b.recoveryBucket.succeed != 1 && b.recoveryBucket.failed != 1 && b.recoveryBucket.timeout != 1 && b.recoveryBucket.reject != 1 {
		t.Fatalf("breaker method error. succeed: %d failed: %d timeout: %d reject: %d", b.recoveryBucket.succeed, b.recoveryBucket.failed, b.recoveryBucket.timeout, b.recoveryBucket.reject)
	}

	b.state = OPEN
	b.c <- SUCCEED

}

func TestBreakerMethodUpdateState(t *testing.T) {
	metricsOptions := MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b := NewBreaker(metricsOptions, func([]bucket) uint8 { return TRUE }, IsClosed)

	tick := time.NewTicker(100 * time.Second)
	c := make(chan time.Time)
	tick.C = c

	go b.updateState(tick)

	c <- time.Now()

	if b.state != OPEN {
		t.Fatalf("state should be OPEN, got: %v", b.state)
	}

	metricsOptions = MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b = NewBreaker(metricsOptions, func([]bucket) uint8 { return HOLD }, IsClosed)

	tick = time.NewTicker(100 * time.Second)
	c = make(chan time.Time)
	tick.C = c

	go b.updateState(tick)

	c <- time.Now()

	if b.state != CLOSED {
		t.Fatalf("state should be CLOSED, got: %v", b.state)
	}

	metricsOptions = MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b = NewBreaker(metricsOptions, IsOpen, func(bucket) uint8 { return HOLD })
	b.state = HALFOPEN

	tick = time.NewTicker(100 * time.Second)
	c = make(chan time.Time)
	tick.C = c

	go b.updateState(tick)

	c <- time.Now()

	if b.state != HALFOPEN {
		t.Fatalf("state should be HALFOPEN, got: %v", b.state)
	}

	metricsOptions = MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b = NewBreaker(metricsOptions, IsOpen, func(bucket) uint8 { return TRUE })
	b.state = HALFOPEN

	tick = time.NewTicker(100 * time.Second)
	c = make(chan time.Time)
	tick.C = c

	go b.updateState(tick)

	c <- time.Now()

	if b.state != CLOSED {
		t.Fatalf("state should be CLOSED, got: %v", b.state)
	}

	metricsOptions = MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b = NewBreaker(metricsOptions, IsOpen, func(bucket) uint8 { return FALSE })
	b.state = HALFOPEN

	tick = time.NewTicker(100 * time.Second)
	c = make(chan time.Time)
	tick.C = c

	go b.updateState(tick)

	c <- time.Now()

	if b.state != OPEN {
		t.Fatalf("state should be OPEN, got: %v", b.state)
	}

	metricsOptions = MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b = NewBreaker(metricsOptions, IsOpen, func(bucket) uint8 { return FALSE })
	b.state = OPEN

	tick = time.NewTicker(100 * time.Second)
	c = make(chan time.Time)
	tick.C = c

	go b.updateState(tick)

	c <- time.Now()

	if b.state != OPEN {
		t.Fatalf("state should be OPEN, got: %v", b.state)
	}

	b.Stop()

	select {
	case c <- time.Now():
		t.Fatalf("channel should be blocked")
	default:
	}

}

func TestBreakerMethodChan(t *testing.T) {
	metricsOptions := MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b := NewBreaker(metricsOptions, func([]bucket) uint8 { return TRUE }, IsClosed)

	b.c = make(chan uint8)

	go func() {
		b.Chan() <- 1
	}()

	res := <-b.c

	if res != 1 {
		t.Fatalf("result should be 1, got: %v", res)
	}
}

func TestBreakerMethodActive(t *testing.T) {
	metricsOptions := MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b := NewBreaker(metricsOptions, func([]bucket) uint8 { return TRUE }, IsClosed)

	b.state = CLOSED
	result := b.Active()
	if result != false {
		t.Fatalf("result should be false, got: %v", result)
	}

	b.state = HALFOPEN
	result = b.Active()
	if result != false {
		t.Fatalf("result should be false, got: %v", result)
	}

	b.state = OPEN
	result = b.Active()
	if result != true {
		t.Fatalf("result should be true, got: %v", result)
	}
}

func TestBreakerMethodRecovery(t *testing.T) {
	metricsOptions := MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1 * time.Second,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b := NewBreaker(metricsOptions, func([]bucket) uint8 { return TRUE }, IsClosed)

	tick := time.NewTicker(1 * time.Second)
	c := make(chan time.Time)
	tick.C = c
	b.state = OPEN
	go b.recovery(tick)
	c <- time.Now()

	if b.state != HALFOPEN {
		t.Fatalf("state should be HALFOPEN, got: %v", b.state)
	}

	b.state = OPEN
	go b.recovery(time.NewTicker(1 * time.Second))
	b.Stop()

	if b.state != OPEN {
		t.Fatalf("state should be OPEN, got: %v", b.state)
	}

}

func TestBreakerMethodStart(t *testing.T) {
	metricsOptions := MetricsOptions{
		MetricsRollingCount: 5,
		MetricsInterval:     1,
		ReceiveInterval:     1 * time.Second,
		UpdateStateInterval: 1 * time.Second,
		RecoverInterval:     5 * time.Second,
	}

	b := NewBreaker(metricsOptions, func([]bucket) uint8 { return TRUE }, IsClosed)
	go b.Start()
	time.Sleep(100 * time.Millisecond)

	b.Stop()
}
