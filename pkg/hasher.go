package pkg

import "golang.org/x/crypto/bcrypt"

// Hasher struct provides methods for hashing and comparing passwords.
type Hasher struct{}

// HashPassword hashes a plain-text password using bcrypt with the minimum cost parameter.
// It returns the hashed password as a string and any error encountered during hashing.
func (h Hasher) HashPassword(plainPassword string) (hashedPassword string, err error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.MinCost)
	return string(bytes), err
}

// ComparePassword checks if a given plain-text password matches a hashed password.
// It returns `true` if the password matches, otherwise `false`.
func (h Hasher) ComparePassword(password, hashedPassword string) (passwordIsValid bool) {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
