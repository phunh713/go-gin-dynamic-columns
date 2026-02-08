package utils

func StringSlicesIntersect(a, b []string) []string {
	set := make(map[string]struct{})
	for _, item := range a {
		set[item] = struct{}{}
	}
	res := make([]string, 0)
	for _, item := range b {
		if _, exists := set[item]; exists {
			res = append(res, item)
		}
	}
	return res
}

// AppendUnique appends only unique items from 'newItems' to 'existing' slice
// Works with any comparable type (string, int, etc.)
func AppendUnique[T comparable](existing []T, newItems ...T) []T {
	// Create a set of existing items for fast lookup
	existingSet := make(map[T]bool, len(existing))
	for _, item := range existing {
		existingSet[item] = true
	}

	// Only append items that don't exist
	for _, item := range newItems {
		if !existingSet[item] {
			existing = append(existing, item)
			existingSet[item] = true
		}
	}

	return existing
}

func SliceContains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
