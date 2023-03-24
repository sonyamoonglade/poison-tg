package functools

func Map[A any, B any](fun func(A) B, input []A) []B {
	result := make([]B, 0, len(input))

	for i := range input {
		result = append(result, fun(input[i]))
	}

	return result
}

func Reduce[A, B any](f func(B, A) B, s []A, initValue B) B {
	acc := initValue
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}
