package auth

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func TestPasswordHashAndVerify(t *testing.T) {
	h, err := HashPassword("un-mot-de-passe-solide")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(h, "un-mot-de-passe-solide") {
		t.Fatal("l'empreinte ne doit pas contenir le mot de passe")
	}
	if !VerifyPassword("un-mot-de-passe-solide", h) {
		t.Error("le bon mot de passe doit être accepté")
	}
	if VerifyPassword("autre-mot-de-passe", h) {
		t.Error("un mauvais mot de passe doit être refusé")
	}
	if VerifyPassword("x", "n'importe-quoi") {
		t.Error("une empreinte mal formée doit être refusée")
	}
}

func TestPasswordPolicy(t *testing.T) {
	if ValidatePasswordPolicy("court") == nil {
		t.Error("un mot de passe trop court doit être refusé")
	}
	if err := ValidatePasswordPolicy("assez-long-10"); err != nil {
		t.Errorf("mot de passe valide refusé : %v", err)
	}
}

func newTM(t *testing.T) *TokenManager {
	t.Helper()
	tm, err := NewTokenManager([]byte(strings.Repeat("s", 32)), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return tm
}

func TestTokenIssueVerify(t *testing.T) {
	tm := newTM(t)
	tok, _, err := tm.Issue("a@b.com", RoleClient, "t-1")
	if err != nil {
		t.Fatal(err)
	}
	c, err := tm.Verify(tok)
	if err != nil {
		t.Fatalf("jeton valide refusé : %v", err)
	}
	if c.Subject != "a@b.com" || c.Role != RoleClient || c.Tenant != "t-1" {
		t.Errorf("claims inattendus : %+v", c)
	}
}

func TestTokenRejections(t *testing.T) {
	tm := newTM(t)
	tok, _, _ := tm.Issue("a@b.com", RoleClient, "t-1")
	parts := strings.Split(tok, ".")

	// signature altérée
	if _, err := tm.Verify(parts[0] + "." + parts[1] + "." + parts[2][:len(parts[2])-2] + "AA"); err == nil {
		t.Error("signature altérée acceptée")
	}

	// charge utile modifiée (élévation de rôle) avec l'ancienne signature
	evil := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"a@b.com","role":"ADMIN","tenant":"t-1","iat":1,"exp":9999999999}`))
	if _, err := tm.Verify(parts[0] + "." + evil + "." + parts[2]); err == nil {
		t.Error("charge utile modifiée acceptée")
	}

	// alg:none
	none := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	if _, err := tm.Verify(none + "." + parts[1] + "."); err == nil {
		t.Error("jeton alg:none accepté")
	}

	// secret différent
	other, _ := NewTokenManager([]byte(strings.Repeat("x", 32)), time.Hour)
	if _, err := other.Verify(tok); err == nil {
		t.Error("jeton signé avec un autre secret accepté")
	}

	// expiration
	tm.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }
	old, _, _ := tm.Issue("a@b.com", RoleClient, "t-1")
	tm.now = time.Now
	if _, err := tm.Verify(old); err != ErrExpiredToken {
		t.Errorf("jeton expiré : erreur attendue %v, reçue %v", ErrExpiredToken, err)
	}
}

func TestSecretTooShort(t *testing.T) {
	if _, err := NewTokenManager([]byte("court"), time.Hour); err == nil {
		t.Error("un secret trop court doit être refusé")
	}
}

func TestStoreAuthenticate(t *testing.T) {
	s, _ := NewStore()
	if _, err := s.AddUser("Admin@Example.com", "mot-de-passe-admin", RoleAdmin); err != nil {
		t.Fatal(err)
	}
	u, err := s.Authenticate(" admin@example.com ", "mot-de-passe-admin")
	if err != nil || u.Role != RoleAdmin {
		t.Fatalf("authentification valide refusée : %v", err)
	}
	if _, err := s.Authenticate("admin@example.com", "faux"); err != ErrInvalidCredentials {
		t.Error("mauvais mot de passe accepté")
	}
	if _, err := s.Authenticate("inconnu@example.com", "peu-importe-123"); err != ErrInvalidCredentials {
		t.Error("compte inconnu accepté")
	}
	if _, err := s.AddUser("pas-un-email", "mot-de-passe-ok-1", RoleClient); err == nil {
		t.Error("e-mail invalide accepté")
	}
}

func TestResetFlow(t *testing.T) {
	s, _ := NewStore()
	s.AddUser("u@example.com", "ancien-mot-de-passe", RoleClient)

	if _, ok, _ := s.IssueResetToken("inconnu@example.com"); ok {
		t.Error("un jeton ne doit pas être émis pour un compte inexistant")
	}

	tok, ok, err := s.IssueResetToken("u@example.com")
	if err != nil || !ok || tok == "" {
		t.Fatalf("émission du jeton : ok=%v err=%v", ok, err)
	}
	if err := s.ConsumeResetToken(tok, "court"); err == nil {
		t.Error("mot de passe faible accepté")
	}
	// le jeton a été consommé par la tentative précédente ? Non : la politique est vérifiée avant.
	if err := s.ConsumeResetToken(tok, "nouveau-mot-de-passe"); err != nil {
		t.Fatalf("réinitialisation valide refusée : %v", err)
	}
	if _, err := s.Authenticate("u@example.com", "nouveau-mot-de-passe"); err != nil {
		t.Error("le nouveau mot de passe ne fonctionne pas")
	}
	if err := s.ConsumeResetToken(tok, "encore-un-autre-mdp"); err != ErrInvalidResetToken {
		t.Error("un jeton doit être à usage unique")
	}
}

func TestResetTokenExpiryAndReplacement(t *testing.T) {
	s, _ := NewStore()
	s.AddUser("u@example.com", "ancien-mot-de-passe", RoleClient)

	tok1, _, _ := s.IssueResetToken("u@example.com")
	tok2, _, _ := s.IssueResetToken("u@example.com")
	if err := s.ConsumeResetToken(tok1, "nouveau-mot-de-passe"); err != ErrInvalidResetToken {
		t.Error("l'ancien jeton doit être invalidé par l'émission d'un nouveau")
	}

	s.now = func() time.Time { return time.Now().Add(ResetTokenTTL + time.Minute) }
	if err := s.ConsumeResetToken(tok2, "nouveau-mot-de-passe"); err != ErrInvalidResetToken {
		t.Error("un jeton expiré doit être refusé")
	}
}

func TestTenantIDStable(t *testing.T) {
	if TenantID("A@b.com") != TenantID(" a@b.com ") {
		t.Error("le tenant doit être insensible à la casse et aux espaces")
	}
	if TenantID("a@b.com") == TenantID("c@d.com") {
		t.Error("deux e-mails distincts ne doivent pas partager un tenant")
	}
}

func TestLimiter(t *testing.T) {
	l := NewLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.Allow("ip") {
			t.Fatalf("tentative %d refusée trop tôt", i+1)
		}
	}
	if l.Allow("ip") {
		t.Error("la 4e tentative aurait dû être refusée")
	}
	if !l.Allow("autre-ip") {
		t.Error("une autre clé ne doit pas être bloquée")
	}
	l.now = func() time.Time { return time.Now().Add(2 * time.Minute) }
	if !l.Allow("ip") {
		t.Error("la fenêtre écoulée doit réautoriser")
	}
}
