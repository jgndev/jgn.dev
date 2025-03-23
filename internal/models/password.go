// Package models provides data structures and functionality for password generation.
package models

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"errors"
	"strings"
)

// character set constants for password generation
const (
	// Uppercase is the possible uppercase letters
	Uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Lowercase is the possible lowercase letters
	Lowercase = "abcdefghijklmnopqrstuvwxyz"
	// Numbers is the possible numbers
	Numbers = "0123456789"
	// Special is the possible special characters
	Special = "!@#$%^&*-_=+<>?;:[]{}(),./|"
	// MinLength is the shortest password that pwbot will generate
	MinLength = 6
	// MaxLength is the longest password that pwbot will generate
	MaxLength = 64
)

// PasswordGenerator interface defines the contract for password generation.
type PasswordGenerator interface {
	GeneratePassword() (string, error)
}

// Password struct holds the configuration options for password generation.
type Password struct {
	Uppercase bool
	Lowercase bool
	Numbers   bool
	Special   bool
	Length    int
}

// secureRandomInt generates a cryptographically secure random number between 0 and max
func secureRandomInt(max int) (int, error) {
	// Create an 8-byte buffer to store random bytes
	// We use 8 bytes (64 bits) to ensure we have enough random data to generate
	// numbers across the full range of uint64
	var b [8]byte

	// Read random bytes from the system's secure random number generator
	// crypto/rand.Read uses sources like /dev/urandom on Unix systems or
	// CryptoAPI on Windows to generate cryptographically secure random numbers
	// The '_' discards the number of bytes read since Read always fills the buffer
	_, err := cryptorand.Read(b[:])
	if err != nil {
		return 0, err
	}

	// Convert the 8 random bytes into a uint64 number
	// binary.LittleEndian.Uint64 interprets the bytes as a 64-bit unsigned integer
	// using little-endian byte order (least significant byte first)
	// This gives us a random number between 0 and 2^64-1
	n := binary.LittleEndian.Uint64(b[:])

	// Convert the uint64 to a number within our desired range [0, max)
	// We use modulo (%) to map the full uint64 range down to our target range
	// Note: This method can introduce a small bias if max doesn't divide 2^64 evenly
	// For password generation, this bias is negligible and doesn't affect security
	return int(n % uint64(max)), nil
}

// GeneratePassword creates a password based on the Password struct configuration.
// It ensures at least one character from each selected character set is included,
// fills the password to the desired length, and then shuffles it for randomness.
// Returns the generated password and any error encountered.
func (p *Password) GeneratePassword() (string, error) {
	p.ClampLength()

	var possible, pw strings.Builder
	requiredChars := make([]string, 0)

	// resize possible to accommodate character sets
	possible.Grow(len(Uppercase) + len(Lowercase) + len(Numbers) + len(Special))
	// resize pw to the length required for the password
	pw.Grow(p.Length)

	// ensure at least one character from each selection is used
	if p.Uppercase {
		possible.WriteString(Uppercase)
		requiredChars = append(requiredChars, Uppercase)
	}
	if p.Lowercase {
		possible.WriteString(Lowercase)
		requiredChars = append(requiredChars, Lowercase)
	}
	if p.Numbers {
		possible.WriteString(Numbers)
		requiredChars = append(requiredChars, Numbers)
	}
	if p.Special {
		possible.WriteString(Special)
		requiredChars = append(requiredChars, Special)
	}

	// check if at least one character set was selected, otherwise return error
	if possible.Len() == 0 {
		return "", errors.New("no character set selected")
	}

	// Ensure we have at least one character from each required set
	for _, charSet := range requiredChars {
		idx, err := secureRandomInt(len(charSet))
		if err != nil {
			return "", err
		}
		pw.WriteByte(charSet[idx])
	}

	// fill the password up to the selected length
	remainingLength := p.Length - len(requiredChars)
	possibleChars := possible.String()

	for i := 0; i < remainingLength; i++ {
		idx, err := secureRandomInt(len(possibleChars))
		if err != nil {
			return "", err
		}

		pw.WriteByte(possibleChars[idx])
	}

	// shuffle the password
	pwBytes := []byte(pw.String())
	for i := len(pwBytes) - 1; i > 0; i-- {
		j, err := secureRandomInt(i + 1)
		if err != nil {
			return "", err
		}
		pwBytes[i], pwBytes[j] = pwBytes[j], pwBytes[i]
	}

	return string(pwBytes), nil
}

// ClampLength ensures the password length is within the allowed range.
// It adjusts the Length field of the Password struct if necessary.
func (p *Password) ClampLength() {
	if p.Length < MinLength {
		p.Length = MinLength
	}
	if p.Length > MaxLength {
		p.Length = MaxLength
	}
}

// NewPassword creates and returns a new PasswordGenerator with the specified configuration.
// This function allows for easy creation of Password instances.
func NewPassword(uppercase, lowercase, numbers, special bool, length int) PasswordGenerator {
	return &Password{
		Uppercase: uppercase,
		Lowercase: lowercase,
		Numbers:   numbers,
		Special:   special,
		Length:    length,
	}
}
