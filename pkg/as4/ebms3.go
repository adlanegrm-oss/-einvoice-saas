package as4

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrReplayDetected = errors.New("as4: duplicate message detected (replay protection)")
)

type AS4Message struct {
	TenantID      string    `json:"tenant_id"`
	MessageID     string    `json:"message_id"`
	SenderID      string    `json:"sender_id"`
	Payload       []byte    `json:"payload"`
	PayloadDigest string    `json:"payload_digest"`
	Timestamp     time.Time `json:"timestamp"`
}

type Receipt struct {
	MessageID  string    `json:"message_id"`
	RefToMsgID string    `json:"ref_to_msg_id"`
	Status     string    `json:"status"`
	Timestamp  time.Time `json:"timestamp"`
}

type InMemoryReplayStore struct {
	mu       sync.Mutex
	messages map[string]bool
}

func NewReplayStore() *InMemoryReplayStore {
	return &InMemoryReplayStore{
		messages: make(map[string]bool),
	}
}

func (s *InMemoryReplayStore) Has(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.messages[key]
}

func (s *InMemoryReplayStore) Set(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[key] = true
}

func ProcessInboundAS4(msg AS4Message, store *InMemoryReplayStore) (*Receipt, error) {
	key := msg.TenantID + ":" + msg.MessageID
	if store.Has(key) {
		return nil, ErrReplayDetected
	}
	store.Set(key)

	return &Receipt{
		MessageID:  "receipt-" + msg.MessageID,
		RefToMsgID: msg.MessageID,
		Status:     "RECEIVED",
		Timestamp:  time.Now().UTC(),
	}, nil
}
