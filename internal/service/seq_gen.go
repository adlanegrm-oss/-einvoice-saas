package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

var (
	outboxSequence uint64
	nodeIdentifier string
)

func init() {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = fmt.Sprintf("pid-%d", os.Getpid())
	}
	h := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", hostname, os.Getpid())))
	nodeIdentifier = hex.EncodeToString(h[:])[:8]
}

func generateOutboxID() string {
	seq := atomic.AddUint64(&outboxSequence, 1)
	return fmt.Sprintf("outbox-%d-%s-%06d", time.Now().UnixNano(), nodeIdentifier, seq)
}
