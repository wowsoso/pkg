package breaker

import (
	"testing"
)

func TestIsOpenShouldReturnFalseWhenAllFailedCountIsZero(t *testing.T) {
	if IsOpen(make([]bucket, 1)) != HOLD {
		t.Fatalf("isOpen should return HOLD")
	}
}

func TestIsOpenShouldReturnFalseWhenNoTriggerBreaker(t *testing.T) {
	buckets := make([]bucket, 1)
	buckets[0].succeed = 1
	buckets[0].failed = 1
	if IsOpen(buckets) != HOLD {
		t.Fatalf("isOpen should return HOLD")
	}
}

func TestIsOpenShouldReturnTrueWhenTriggerBreaker(t *testing.T) {
	buckets := make([]bucket, 1)
	buckets[0].succeed = 1
	buckets[0].failed = 2
	if IsOpen(buckets) != TRUE {
		t.Fatalf("isOpen should return TRUE")
	}
}

func TestIsClosedShouldReturnFalseWhenfailedsMoreThanTwo(t *testing.T) {
	var b bucket
	b.succeed = 1
	b.failed = 3

	if IsClosed(b) != FALSE {
		t.Fatalf("isClosed should return FALSE")
	}

}

func TestIsClosedShouldReturnHoldWhenallLessThanTen(t *testing.T) {
	var b bucket
	b.succeed = 1
	b.failed = 2

	if IsClosed(b) != HOLD {
		t.Fatalf("isClosed should return HOLD")
	}

}

func TestIsClosedShouldReturnTrue(t *testing.T) {
	var b bucket
	b.succeed = 10
	b.failed = 0

	if IsClosed(b) != TRUE {
		t.Fatalf("isClosed should return TRUE")
	}

}
