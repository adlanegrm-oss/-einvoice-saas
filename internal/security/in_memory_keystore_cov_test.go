package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

func hashKeyForTest(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func TestInMemoryKeyStore_Lifecycle(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	s := NewInMemoryKeyStore()
	ctx := context.Background()

	s.AddKey(
		"cle-active",
		"tenant-1",
		true,
	)

	s.AddKey(
		"cle-inactive",
		"tenant-2",
		false,
	)

	s.AddKey(
		"cle-a-revoquer",
		"tenant-3",
		true,
	)

	s.RevokeKey("cle-a-revoquer")
	s.RevokeKey("cle-inexistante")

	rec, err := s.FindTenantByKeyHash(
		ctx,
		hashKeyForTest("cle-active"),
	)

	if err != nil ||
		rec == nil ||
		rec.ID != "tenant-1" ||
		!rec.Active {
		t.Fatalf(
			"cle active : rec=%+v err=%v",
			rec,
			err,
		)
	}

	if _, err := s.FindTenantByKeyHash(
		ctx,
		hashKeyForTest("cle-inactive"),
	); !errors.Is(err, ErrKeyInactive) {
		t.Fatalf(
			"ErrKeyInactive attendu, obtenu %v",
			err,
		)
	}

	if _, err := s.FindTenantByKeyHash(
		ctx,
		hashKeyForTest("cle-a-revoquer"),
	); !errors.Is(err, ErrKeyRevoked) {
		t.Fatalf(
			"ErrKeyRevoked attendu, obtenu %v",
			err,
		)
	}

	if _, err := s.FindTenantByKeyHash(
		ctx,
		hashKeyForTest("inconnue"),
	); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf(
			"ErrKeyNotFound attendu, obtenu %v",
			err,
		)
	}
}

func TestInMemoryKeyStore_PanicsInProduction(t *testing.T) {
	for _, env := range []string{
		"production",
		"prod",
	} {
		t.Run(env, func(t *testing.T) {
			t.Setenv("APP_ENV", env)

			defer func() {
				if recover() == nil {
					t.Fatal(
						"un panic etait attendu en production",
					)
				}
			}()

			NewInMemoryKeyStore()
		})
	}
}
