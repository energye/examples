package token

import "sync"

var currentTokensMu sync.RWMutex
var currentTokens = Derive(DefaultSeed(), ModeLight)

// DefaultLight returns the default light token set.
func DefaultLight() Tokens {
	return Derive(DefaultSeed(), ModeLight)
}

// DefaultDark returns the default dark token set.
func DefaultDark() Tokens {
	return Derive(DefaultSeed(), ModeDark)
}

// Current returns the active token set.
func Current() Tokens {
	currentTokensMu.RLock()
	defer currentTokensMu.RUnlock()
	return currentTokens
}

// SetCurrent sets the active token set.
func SetCurrent(tokens Tokens) {
	currentTokensMu.Lock()
	defer currentTokensMu.Unlock()
	currentTokens = tokens
}

// Reset restores the default light token set.
func Reset() {
	SetCurrent(DefaultLight())
}
