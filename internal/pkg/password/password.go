//go:generate mockgen -source=password.go -destination=./mocks/password_hasher_mocks.go -package=passwordhashermocks

package password

import "golang.org/x/crypto/bcrypt"

type Hasher interface {
	Hash(password string) string
	Validate(hashed, password string) bool
}

type hasher struct {}

func (h hasher) Hash(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hashed)
}

func (h hasher) Validate(hashed, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}

func newPasswordHasher() Hasher {
	return hasher{}
}

var DefaultPasswordHasher = newPasswordHasher()