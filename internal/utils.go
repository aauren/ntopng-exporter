package internal

func IsItemInArray(haystack []string, needle string) bool {
	for _, hay := range haystack {
		if needle == hay {
			return true
		}
	}
	return false
}
