package breaker

func IsOpen(buckets []bucket) uint8 {
	var succeed, failed, timeout, reject uint32

	for i := 0; i < len(buckets); i++ {
		succeed = succeed + buckets[i].succeed
		failed += failed + buckets[i].failed
		timeout += timeout + buckets[i].timeout
		reject += reject + buckets[i].reject
	}

	allFailed := failed + timeout + reject

	switch {
	case allFailed == 0:
		return HOLD
	case float64(allFailed)/float64(succeed+allFailed) > 0.5:
		return TRUE
	}

	return HOLD
}

func IsClosed(recoveryBucket bucket) uint8 {
	if recoveryBucket.faileds() > 2 {
		return FALSE
	}

	if recoveryBucket.all() < 10 {
		return HOLD
	}

	return TRUE
}
