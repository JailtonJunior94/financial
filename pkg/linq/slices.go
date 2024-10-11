package linq

type FilterFunc[T any] func(T) bool

func Filter[T any](items []T, fn FilterFunc[T]) []T {
	newSlice := make([]T, len(items))
	for index, item := range items {
		if fn(item) {
			newSlice[index] = item
		}
	}
	return newSlice
}

type FindFunc[T any] func(T) bool

func Find[T any](items []T, fn RemoveFunc[T]) T {
	var empty T
	for _, item := range items {
		if fn(item) {
			return item
		}
	}
	return empty
}

type RemoveFunc[T any] func(T) bool

func Remove[T any](items []T, fn RemoveFunc[T]) []T {
	newSlice := make([]T, len(items))
	for index, item := range items {
		if fn(item) {
			newSlice = append(items[:index], items[index+1:]...)
		}
	}
	return newSlice
}

type MapFunc[I, O any] func(I) O

func Map[I, O any](items []I, fn MapFunc[I, O]) []O {
	newSlice := make([]O, len(items))
	for index, item := range items {
		newSlice[index] = fn(item)
	}
	return newSlice
}
