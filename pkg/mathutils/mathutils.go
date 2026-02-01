package mathutils

// Clamp limita um valor entre um mínimo e um máximo usando os built-ins min/max do Go 1.21+.
//
// Exemplo:
//
//	value := Clamp(150, 0, 100) // retorna 100
//	value := Clamp(-5, 0, 100)  // retorna 0
//	value := Clamp(50, 0, 100)  // retorna 50
func Clamp[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](value, minVal, maxVal T) T {
	return min(max(value, minVal), maxVal)
}

// MaxOf retorna o maior valor entre múltiplos argumentos usando o built-in max do Go 1.21+.
// Aceita número variável de argumentos.
//
// Exemplo:
//
//	largest := MaxOf(10, 25, 15, 30, 5) // retorna 30
func MaxOf[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](values ...T) T {
	if len(values) == 0 {
		var zero T
		return zero
	}
	result := values[0]
	for _, v := range values[1:] {
		result = max(result, v)
	}
	return result
}

// MinOf retorna o menor valor entre múltiplos argumentos usando o built-in min do Go 1.21+.
// Aceita número variável de argumentos.
//
// Exemplo:
//
//	smallest := MinOf(10, 25, 15, 30, 5) // retorna 5
func MinOf[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](values ...T) T {
	if len(values) == 0 {
		var zero T
		return zero
	}
	result := values[0]
	for _, v := range values[1:] {
		result = min(result, v)
	}
	return result
}

// SafeSlice retorna um sub-slice seguro usando min para evitar index out of bounds.
// Útil para paginação e limitação de resultados.
//
// Exemplo:
//
//	items := []int{1, 2, 3, 4, 5}
//	page := SafeSlice(items, 0, 10) // retorna [1, 2, 3, 4, 5] ao invés de panic
func SafeSlice[T any](slice []T, start, end int) []T {
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}

	length := len(slice)
	start = min(start, length)
	end = min(end, length)

	if start >= end {
		return []T{}
	}

	return slice[start:end]
}

// CapLimit retorna o valor limitado ao máximo especificado.
// Útil para limitar tamanhos de página, batch sizes, etc.
//
// Exemplo:
//
//	pageSize := CapLimit(userRequestedSize, 100) // Nunca excede 100
func CapLimit[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](value, maxLimit T) T {
	return min(value, maxLimit)
}

// AtLeast garante que o valor seja pelo menos o mínimo especificado.
// Útil para garantir valores positivos ou mínimos.
//
// Exemplo:
//
//	pageSize := AtLeast(userValue, 1) // Garante pelo menos 1
func AtLeast[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](value, minValue T) T {
	return max(value, minValue)
}
