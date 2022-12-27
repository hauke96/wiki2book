package util

func Contains[T comparable](values []T, valueToCheck T) bool {
	for _, value := range values {
		if value == valueToCheck {
			return true
		}
	}

	return false
}
