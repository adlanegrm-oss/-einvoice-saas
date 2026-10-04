package as4

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type PersistentReplayStore struct {
	db *sql.DB
}

func NewPersistentReplayStore(db *sql.DB) *PersistentReplayStore {
	return &PersistentReplayStore{db: db}
}

func (s *PersistentReplayStore) RegisterAndVerify(msg AS4Message) (*Receipt, error) {
	if s.db == nil {
		return nil, errors.New("db connection is required for persistent anti-replay")
	}

	query := `
		INSERT INTO inbound_messages (tenant_id, message_id, protocol, payload_digest, sender_id, received_at)
		VALUES ($1, $2, 'AS4', $3, $4, $5)
	`
	_, err := s.db.Exec(query, msg.TenantID, msg.MessageID, msg.PayloadDigest, msg.SenderID, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReplayDetected, err)
	}

	return &Receipt{
		MessageID:  fmt.Sprintf("receipt-%s", msg.MessageID),
		RefToMsgID: msg.MessageID,
		Status:     "RECEIVED",
		Timestamp:  time.Now().UTC(),
	}, nil
}
