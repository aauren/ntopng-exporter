package prometheus

func deepAppend(src []string, appends ...string) []string {
	newList := make([]string, len(src))
	_ = copy(newList, src)
	newList = append(newList, appends...)
	return newList
}
