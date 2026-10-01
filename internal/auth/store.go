package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/mail"
	"strings"
	"sync"
	"time"
)

// ResetTokenTTL est la durée de validité d'un lien de réinitialisation.
const ResetTokenTTL = 15 * time.Minute

var (
	ErrInvalidCredentials = errors.New("identifiants invalides")
	ErrInvalidResetToken  = errors.New("jeton de réinitialisation invalide ou expiré")
)

// User est un compte applicatif.
type User struct {
	Email        string
	PasswordHash string
	Role         Role
	Tenant       string
}

type resetEntry struct {
	email   string
	expires time.Time
}

// Store garde les comptes et les jetons de réinitialisation en mémoire.
// Les comptes sont recréés au démarrage à partir de la configuration :
// une persistance en base est la prochaine étape (voir README).
type Store struct {
	mu     sync.Mutex
	users  map[string]*User
	resets map[string]resetEntry // clé : SHA-256(jeton)
	now    func() time.Time
	dummy  string // empreinte factice pour égaliser le temps de réponse
}

func NewStore() (*Store, error) {
	dummy, err := HashPassword("dummy-password-for-timing")
	if err != nil {
		return nil, err
	}
	return &Store{
		users:  make(map[string]*User),
		resets: make(map[string]resetEntry),
		now:    time.Now,
		dummy:  dummy,
	}, nil
}

// NormalizeEmail met l'adresse en minuscules et vérifie sa syntaxe.
func NormalizeEmail(email string) (string, error) {
	e := strings.ToLower(strings.TrimSpace(email))
	addr, err := mail.ParseAddress(e)
	if err != nil || addr.Address != e {
		return "", errors.New("adresse e-mail invalide")
	}
	return e, nil
}

// TenantID dérive un identifiant de dossier stable et sûr à partir de l'e-mail.
func TenantID(email string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
	return "t-" + hex.EncodeToString(sum[:8])
}

// AddUser crée un compte (mot de passe haché ici, politique appliquée).
func (s *Store) AddUser(email, password string, role Role) (*User, error) {
	e, err := NormalizeEmail(email)
	if err != nil {
		return nil, err
	}
	if err := ValidatePasswordPolicy(password); err != nil {
		return nil, err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &User{Email: e, PasswordHash: hash, Role: role, Tenant: TenantID(e)}
	s.mu.Lock()
	s.users[e] = u
	s.mu.Unlock()
	return u, nil
}

// Authenticate vérifie les identifiants. Le coût est identique que le compte existe ou non.
func (s *Store) Authenticate(email, password string) (*User, error) {
	e := strings.ToLower(strings.TrimSpace(email))
	s.mu.Lock()
	u := s.users[e]
	s.mu.Unlock()
	if u == nil {
		VerifyPassword(password, s.dummy)
		return nil, ErrInvalidCredentials
	}
	if !VerifyPassword(password, u.PasswordHash) {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}

// IssueResetToken crée un jeton à usage unique pour un compte existant.
// Le second retour est false si le compte n'existe pas : l'appelant ne doit
// jamais le révéler à l'utilisateur (réponse HTTP identique dans les deux cas).
func (s *Store) IssueResetToken(email string) (token string, ok bool, err error) {
	e := strings.ToLower(strings.TrimSpace(email))
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[e]; !exists {
		return "", false, nil
	}
	token, err = randomToken(32)
	if err != nil {
		return "", false, err
	}
	// un seul jeton actif par compte
	for k, v := range s.resets {
		if v.email == e {
			delete(s.resets, k)
		}
	}
	s.resets[hashToken(token)] = resetEntry{email: e, expires: s.now().Add(ResetTokenTTL)}
	return token, true, nil
}

// ConsumeResetToken change le mot de passe si le jeton est valide, puis l'invalide.
func (s *Store) ConsumeResetToken(token, newPassword string) error {
	if err := ValidatePasswordPolicy(newPassword); err != nil {
		return err
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	key := hashToken(token)
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, found := s.resets[key]
	if !found {
		return ErrInvalidResetToken
	}
	delete(s.resets, key) // usage unique, même en cas d'expiration
	if s.now().After(entry.expires) {
		return ErrInvalidResetToken
	}
	u := s.users[entry.email]
	if u == nil {
		return ErrInvalidResetToken
	}
	u.PasswordHash = hash
	return nil
}

func hashToken(t string) string {
	sum := sha256.Sum256([]byte(t))
	return hex.EncodeToString(sum[:])
}
