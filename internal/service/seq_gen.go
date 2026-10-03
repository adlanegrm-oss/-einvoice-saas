package service

import (
	"fmt"
	"sync/atomic"
	"time"
)

var outboxSequence uint64

func generateOutboxID() string {
	seq := atomic.AddUint64(&outboxSequence, 1)
	return fmt.Sprintf("outbox-%d-%06d", time.Now().UnixNano(), seq)
}
