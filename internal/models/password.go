// Package models provides data structures and functionality for password generation.
package models

import (
	"errors"
	"math/rand"
	"strings"
)

// character set contstants for password generation
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

// GeneratePassword creates a password based on the Password struct configuration.
// It ensures at least one character from each selected character set is included,
// fills the password to the desired length, and then shuffles it for randomness.
// Returns the generated password and any error encountered.
func (p *Password) GeneratePassword() (string, error) {
	p.ClampLength()

	var possible, pw strings.Builder

	// resize possible to accommodate character sets
	possible.Grow(len(Uppercase) + len(Lowercase) + len(Numbers) + len(Special))
	// resize pw to the length required for the password
	pw.Grow(p.Length)

	// ensure at least one character from each selection is used
	if p.Uppercase {
		possible.WriteString(Uppercase)
		pw.WriteByte(Uppercase[rand.Intn(len(Uppercase))])
	}
	if p.Lowercase {
		possible.WriteString(Lowercase)
		pw.WriteByte(Lowercase[rand.Intn(len(Lowercase))])
	}
	if p.Numbers {
		possible.WriteString(Numbers)
		pw.WriteByte(Numbers[rand.Intn(len(Numbers))])
	}
	if p.Special {
		possible.WriteString(Special)
		pw.WriteByte(Special[rand.Intn(len(Special))])
	}

	// check if at least one character set was selected, otherwise return error
	if possible.Len() == 0 {
		return "", errors.New("no character set selected")
	}

	// fill the password up to the selected length
	chars := possible.String()
	for i := pw.Len(); i < p.Length; i++ {
		pw.WriteByte(chars[rand.Intn(len(chars))])
	}

	// shuffle the password
	password := []byte(pw.String())
	rand.Shuffle(len(password), func(i, j int) {
		password[i], password[j] = password[j], password[i]
	})

	return string(password), nil
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
