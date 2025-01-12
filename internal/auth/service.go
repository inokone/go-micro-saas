package auth

import (
	"fmt"
	"math"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/inokone/go-micro-saas/internal/auth/account"
	"github.com/inokone/go-micro-saas/internal/auth/user"
)

// Service is a worker for authentication and authorization.
type Service struct {
	users    user.Storer
	accounts account.Storer
	jwt      *JWTHandler
}

// NewService creates a new `Service`, based on the user and account persistence.
func NewService(users user.Storer, auths account.Storer, jwt *JWTHandler) *Service {
	return &Service{
		users:    users,
		accounts: auths,
		jwt:      jwt,
	}
}

// InvalidCredentials is an error for a bad email address
type InvalidCredentials string

// Error is the string representation of an `CredentialsError`
func (e InvalidCredentials) Error() string { return string(e) }

// LockedUser is an error for a user locked out of the system
type LockedUser struct {
	seconds int64
}

// Error is the string representation of an `LockedUserError`
func (e LockedUser) Error() string { return fmt.Sprintf("%v", e.seconds) }

// ValidateCredentials validates the user credentials sets and clears retry timeout for failed creds
func (s *Service) ValidateCredentials(usr *user.User, password string) error {
	secs, err := s.checkTimeout(usr)
	if err != nil {
		log.WithError(err).WithField("UserID", usr.ID.String()).Error("Failed to collect login timeout.")
		return InvalidCredentials("")
	}
	if secs > 0 {
		return LockedUser{
			seconds: secs,
		}
	}

	verified := usr.VerifyPassword(password)
	if !verified {
		err = s.increaseTimeout(usr)
		if err != nil {
			log.WithField("user", usr.ID.String()).Error("Failed to increase timeout for user")
		}
		return InvalidCredentials("")
	}
	return s.clearTimeout(usr)
}

func (s *Service) checkTimeout(usr *user.User) (int64, error) {
	acc, err := s.accounts.ByUser(usr.ID)
	if err != nil {
		return 0, err
	}
	if !acc.FailedLoginLock.IsZero() && acc.FailedLoginLock.After(time.Now()) {
		return acc.FailedLoginLock.Unix() - time.Now().Unix(), nil
	}
	return 0, nil
}

func (s *Service) increaseTimeout(usr *user.User) error {
	acc, err := s.accounts.ByUser(usr.ID)
	if err != nil {
		return err
	}
	acc.FailedLoginCounter++
	acc.LastFailedLogin = time.Now()
	if acc.FailedLoginCounter > 2 {
		timeout := int(math.Pow(10, float64(acc.FailedLoginCounter-2))) // exponential backoff - 10 sec, 10 sec, 1000 sec, ...
		acc.FailedLoginLock = time.Now().Add(time.Second * time.Duration(timeout))
	}
	return s.accounts.Update(acc)
}

func (s *Service) clearTimeout(usr *user.User) error {
	acc, err := s.accounts.ByUser(usr.ID)
	if err != nil {
		return err
	}
	acc.FailedLoginCounter = 0
	acc.FailedLoginLock = time.Time{} // zero time
	return s.accounts.Update(acc)
}
