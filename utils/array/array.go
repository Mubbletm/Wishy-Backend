package arrays

func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func Map[T any, R any](ss []T, transformer func(T) R) (ret []R) {
	for _, s := range ss {
		ret = append(ret, transformer(s))
	}
	return
}
