package token

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
	return currentTokens
}

// SetCurrent sets the active token set.
func SetCurrent(tokens Tokens) {
	currentTokens = tokens
}

// Reset restores the default light token set.
func Reset() {
	currentTokens = DefaultLight()
}
