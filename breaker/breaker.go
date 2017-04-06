package breaker

import (
	"context"
	// "fmt"
	"sync"
	"time"
)

const (
	SUCCEED = 0
	FAILED  = 1
	TIMEOUT = 2
	REJECT  = 3

	CLOSED   = 0
	OPEN     = 1
	HALFOPEN = 2

	HOLD  = 0
	TRUE  = 1
	FALSE = 2
)

type bucket struct {
	succeed uint32
	failed  uint32
	timeout uint32
	reject  uint32
}

func newBucket() *bucket {
	return &bucket{}
}

func (b *bucket) reset() {
	b.succeed = 0
	b.failed = 0
	b.timeout = 0
	b.reject = 0
}

func (b *bucket) faileds() int {
	return int(b.failed + b.timeout + b.reject)
}

func (b *bucket) all() int {
	return int(b.succeed + b.failed + b.timeout + b.reject)
}

type MetricsOptions struct {
	MetricsRollingCount uint8
	MetricsInterval     time.Duration
	ReceiveInterval     time.Duration
	UpdateStateInterval time.Duration
	RecoverInterval     time.Duration
}

type Breaker struct {
	sync.Mutex

	buckets        []bucket
	recoveryBucket bucket
	state          uint8
	c              chan uint8
	ctx            context.Context
	cancelFunc     context.CancelFunc

	MetricsOptions

	isOpen   func([]bucket) uint8
	isClosed func(bucket) uint8
}

func NewBreaker(metricsOptions MetricsOptions, _isOpen func([]bucket) uint8, _isClosed func(bucket) uint8) *Breaker {
	ctx, cancelFunc := context.WithCancel(context.Background())

	b := &Breaker{
		buckets:    make([]bucket, metricsOptions.MetricsRollingCount),
		state:      CLOSED,
		c:          make(chan uint8),
		ctx:        ctx,
		cancelFunc: cancelFunc,

		MetricsOptions: metricsOptions,
	}

	b.isOpen = _isOpen
	b.isClosed = _isClosed

	go b.Start()
	return b
}

func NewBreakerWithDefault(metricsOptions MetricsOptions) *Breaker {
	ctx, cancelFunc := context.WithCancel(context.Background())

	b := &Breaker{
		buckets:    make([]bucket, metricsOptions.MetricsRollingCount),
		state:      CLOSED,
		c:          make(chan uint8),
		ctx:        ctx,
		cancelFunc: cancelFunc,

		MetricsOptions: metricsOptions,
	}

	b.isOpen = IsOpen
	b.isClosed = IsClosed

	go b.Start()

	return b
}

func (b *Breaker) recovery(tick *time.Ticker) {
	for {
		select {
		case <-b.ctx.Done():
			tick.Stop()
			return
		case <-tick.C:
			b.changeState(HALFOPEN)
			tick.Stop()
			return
		}
	}
}

func (b *Breaker) setStateClosed() {
	b.state = CLOSED
	for i := 0; i < int(b.MetricsRollingCount); i++ {
		b.buckets[i].reset()
	}
}

func (b *Breaker) setStateOpen() {
	b.state = OPEN

	for i := 0; i < int(b.MetricsRollingCount); i++ {
		b.buckets[i].reset()
	}

	go b.recovery(time.NewTicker(b.RecoverInterval))
}

func (b *Breaker) setStateHalfOpen() {
	b.state = HALFOPEN
	b.recoveryBucket.reset()
}

func (b *Breaker) changeState(state uint) {
	b.Lock()
	defer b.Unlock()

	switch b.state {
	case CLOSED:
		if state == OPEN {
			b.setStateOpen()
		}
	case OPEN:
		if state == HALFOPEN {
			b.setStateHalfOpen()
		}
	case HALFOPEN:
		if state == OPEN {
			b.setStateOpen()
		}
		if state == CLOSED {
			b.setStateClosed()
		}
	}
}

func (b *Breaker) receiving(tick *time.Ticker) {
	var succeed, failed, timeout, reject uint32
	var rsucceed, rfailed, rtimeout, rreject uint32

	for {
		select {
		case <-b.ctx.Done():
			tick.Stop()
			return
		case <-tick.C:
			b.buckets[b.MetricsRollingCount-1].succeed += succeed
			b.buckets[b.MetricsRollingCount-1].failed += failed
			b.buckets[b.MetricsRollingCount-1].timeout += timeout
			b.buckets[b.MetricsRollingCount-1].reject += reject
			succeed, failed, timeout, reject = 0, 0, 0, 0
			b.recoveryBucket.succeed += rsucceed
			b.recoveryBucket.failed += rfailed
			b.recoveryBucket.timeout += rtimeout
			b.recoveryBucket.reject += rreject
			rsucceed, rfailed, rtimeout, rreject = 0, 0, 0, 0
		case state := <-b.c:
			switch b.state {
			case OPEN:
				continue
			case HALFOPEN:
				switch state {
				case SUCCEED:
					rsucceed++
				case FAILED:
					rfailed++
				case TIMEOUT:
					rtimeout++
				case REJECT:
					rreject++
				}
			case CLOSED:
				switch state {
				case SUCCEED:
					succeed++
				case FAILED:
					failed++
				case TIMEOUT:
					timeout++
				case REJECT:
					reject++
				}
			}
		}
	}
}

func (b *Breaker) updateState(tick *time.Ticker) {
	for {
		select {
		case <-b.ctx.Done():
			tick.Stop()
			return
		case <-tick.C:
			// fmt.Println(b.buckets, b.recoveryBucket, b.state)
			state := b.state
			switch state {
			case OPEN:
			case HALFOPEN:
				res := b.isClosed(b.recoveryBucket)
				switch res {
				case TRUE:
					b.changeState(CLOSED)
				case FALSE:
					b.changeState(OPEN)
				case HOLD:
				}
			case CLOSED:
				res := b.isOpen(b.buckets)
				if res == TRUE {
					b.changeState(OPEN)
				}
			}
		}
	}
}

func (b *Breaker) Chan() chan<- uint8 {
	return (chan<- uint8)(b.c)
}

func (b *Breaker) Active() bool {
	if b.state == OPEN {
		return true
	}

	return false
}

func (b *Breaker) Start() {
	go b.receiving(time.NewTicker(b.ReceiveInterval))
	go b.updateState(time.NewTicker(b.UpdateStateInterval))

	tick := time.NewTicker(b.MetricsInterval)

	for {
		select {
		case <-b.ctx.Done():
			tick.Stop()
			return
		case <-tick.C:
			for i := 0; i < int(b.MetricsRollingCount)-1; i++ {
				b.buckets[i] = b.buckets[i+1]
			}
			b.buckets[b.MetricsRollingCount-1].reset()
		}
	}
}

func (b *Breaker) Stop() {
	b.cancelFunc()
}
