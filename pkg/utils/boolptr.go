package utils

// IsSetToTrue returns true if and only if the bool pointer is non-nil and set to true.
func IsSetToTrue(b *bool) bool {
	return b != nil && *b == true
}

// IsSetToFalse returns true if and only if the bool pointer is non-nil and set to false.
func IsSetToFalse(b *bool) bool {
	return b != nil && *b == false
}

// True returns a *bool whose underlying value is true.
func True() *bool {
	t := true
	return &t
}

// False returns a *bool whose underlying value is false.
func False() *bool {
	t := false
	return &t
}