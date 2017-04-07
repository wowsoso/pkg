package limiter

import (
	"container/list"
	"strings"
	"sync"
	"time"

	"github.com/golang/time/rate"
)

const (
	PREFIXREQUEST = "req_"
	PREFIXBYTES   = "bytes_"
	MINSIZE       = 65535
	MAXCLIP       = 1024
)

type node struct {
	key   string
	value *rate.Limiter
}

type store struct {
	sync.Mutex

	size         int
	clip         int
	l            *list.List
	tokenBuckets map[string]*list.Element
}

func newStore(size int) *store {
	var s, c int

	s = size

	if size < MINSIZE {
		s = MINSIZE
	}

	c = s >> 8
	if c > MAXCLIP {
		c = MAXCLIP
	}

	return &store{
		size:         s,
		clip:         c,
		l:            list.New(),
		tokenBuckets: make(map[string]*list.Element),
	}
}

func (s *store) getSet(k string, ttl time.Duration, max int) *rate.Limiter {
	s.Lock()
	defer s.Unlock()

	res, ok := s.tokenBuckets[k]
	if ok == false {
		res = s.l.PushFront(&node{k, rate.NewLimiter(rate.Every(ttl), max)})
		s.tokenBuckets[k] = res

		if s.l.Len() > int(s.size) {
			for i := 0; i < s.clip; i++ {
				last := s.l.Back()
				delete(s.tokenBuckets, last.Value.(*node).key)
				s.l.Remove(last)
			}
		}
	} else {
		s.l.MoveToFront(res)
	}

	return res.Value.(*node).value
}

type Limiter struct {
	sync.Mutex

	config Config
	store  *store
}

func NewLimiter(config Config) *Limiter {
	return &Limiter{
		config: config,
		store:  newStore(config.Size),
	}
}

func (l *Limiter) LimitByRequests(remoteIP, path string) bool {
	return l.limitReached(strings.Join([]string{PREFIXREQUEST, remoteIP, path}, ""), 1)
}

func (l *Limiter) LimitByBytes(remoteIP string, length int) bool {
	return l.limitReached(strings.Join([]string{PREFIXBYTES, remoteIP}, ""), length)
}

func (l *Limiter) limitReached(key string, length int) bool {
	l.Lock()
	defer l.Unlock()

	res := l.store.getSet(key, l.config.TTL, int(l.config.Max))

	return !res.AllowN(time.Now(), length)
}
