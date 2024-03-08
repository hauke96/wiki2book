package util

func Contains[T comparable](values []T, valueToCheck T) bool {
	for _, value := range values {
		if value == valueToCheck {
			return true
		}
	}

	return false
}

func RemoveDuplicates[T comparable](originallist []T) []T {
	handledEntries := make(map[T]bool)
	var listWithoutDuplicates []T
	for _, item := range originallist {
		if _, value := handledEntries[item]; !value {
			handledEntries[item] = true
			listWithoutDuplicates = append(listWithoutDuplicates, item)
		}
	}
	return listWithoutDuplicates
}
