package util

func Contains[T comparable](values []T, valueToCheck T) bool {
	for _, value := range values {
		if value == valueToCheck {
			return true
		}
	}

	return false
}

func EqualsInAnyOrder[T comparable](arrayA []T, arrayB []T) bool {
	mapA := make(map[T]T)
	for _, value := range arrayA {
		mapA[value] = value
	}

	mapB := make(map[T]T)
	for _, value := range arrayB {
		mapB[value] = value
	}

	for _, value := range arrayA {
		_, okA := mapA[value]
		_, okB := mapB[value]
		if !okA || !okB {
			return false
		}
	}

	return true
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
