package sliceutils

import "slices"

// Map transforma um slice de tipo T em um slice de tipo R usando a função de mapeamento.
// Usa generics do Go 1.18+ para type-safe transformations.
//
// Exemplo:
//
//	ids := Map(users, func(u User) string { return u.ID })
func Map[T, R any](slice []T, fn func(T) R) []R {
	if slice == nil {
		return nil
	}

	result := make([]R, len(slice))
	for i, item := range slice {
		result[i] = fn(item)
	}
	return result
}

// Filter retorna um novo slice contendo apenas os elementos que satisfazem o predicado.
// Wrapper moderno sobre slices.DeleteFunc com semântica de filtro positiva.
//
// Exemplo:
//
//	activeUsers := Filter(users, func(u User) bool { return u.Active })
func Filter[T any](slice []T, predicate func(T) bool) []T {
	if slice == nil {
		return nil
	}

	// Cria cópia para não modificar o original
	result := slices.Clone(slice)

	// Remove elementos que NÃO satisfazem o predicado (inverte a lógica)
	result = slices.DeleteFunc(result, func(item T) bool {
		return !predicate(item)
	})

	return result
}

// Contains verifica se um elemento existe no slice usando comparação customizada.
// Wrapper type-safe sobre slices.ContainsFunc do Go 1.21+.
//
// Exemplo:
//
//	hasAdmin := Contains(users, func(u User) bool { return u.Role == "admin" })
func Contains[T any](slice []T, predicate func(T) bool) bool {
	return slices.ContainsFunc(slice, predicate)
}

// Find retorna o primeiro elemento que satisfaz o predicado, ou nil se não encontrado.
// Útil para buscar um único item em uma coleção.
//
// Exemplo:
//
//	admin := Find(users, func(u User) bool { return u.Role == "admin" })
func Find[T any](slice []T, predicate func(T) bool) *T {
	idx := slices.IndexFunc(slice, predicate)
	if idx == -1 {
		return nil
	}
	return &slice[idx]
}

// Reduce aplica uma função acumuladora sobre todos os elementos do slice.
// Útil para agregações customizadas.
//
// Exemplo:
//
//	total := Reduce(prices, 0.0, func(acc float64, p float64) float64 { return acc + p })
func Reduce[T, R any](slice []T, initial R, fn func(R, T) R) R {
	result := initial
	for _, item := range slice {
		result = fn(result, item)
	}
	return result
}

// Chunk divide um slice em pedaços menores de tamanho especificado.
// O último chunk pode ter menos elementos se len(slice) não for divisível por size.
//
// Exemplo:
//
//	batches := Chunk(items, 100) // Processa em lotes de 100
func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return nil
	}

	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := min(i+size, len(slice))
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// Unique remove elementos duplicados do slice mantendo a ordem.
// Usa map interno para tracking eficiente com type constraint comparável.
//
// Exemplo:
//
//	uniqueIDs := Unique([]string{"a", "b", "a", "c"}) // ["a", "b", "c"]
func Unique[T comparable](slice []T) []T {
	if slice == nil {
		return nil
	}

	seen := make(map[T]struct{}, len(slice))
	result := make([]T, 0, len(slice))

	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// Partition divide o slice em dois: elementos que satisfazem o predicado e os que não satisfazem.
// Retorna (matches, nonMatches).
//
// Exemplo:
//
//	active, inactive := Partition(users, func(u User) bool { return u.Active })
func Partition[T any](slice []T, predicate func(T) bool) (matches []T, nonMatches []T) {
	for _, item := range slice {
		if predicate(item) {
			matches = append(matches, item)
		} else {
			nonMatches = append(nonMatches, item)
		}
	}
	return
}

// GroupBy agrupa elementos do slice por uma chave extraída pela função.
// Retorna um map onde cada chave mapeia para um slice de elementos.
//
// Exemplo:
//
//	byRole := GroupBy(users, func(u User) string { return u.Role })
func GroupBy[T any, K comparable](slice []T, keyFn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range slice {
		key := keyFn(item)
		result[key] = append(result[key], item)
	}
	return result
}

// Any retorna true se pelo menos um elemento satisfaz o predicado.
// Útil para validações existenciais.
//
// Exemplo:
//
//	hasErrors := Any(results, func(r Result) bool { return r.Error != nil })
func Any[T any](slice []T, predicate func(T) bool) bool {
	return slices.ContainsFunc(slice, predicate)
}

// All retorna true se todos os elementos satisfazem o predicado.
// Útil para validações universais.
//
// Exemplo:
//
//	allValid := All(inputs, func(i Input) bool { return i.Valid() })
func All[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// None retorna true se nenhum elemento satisfaz o predicado.
// Equivalente a !Any(slice, predicate).
//
// Exemplo:
//
//	noErrors := None(results, func(r Result) bool { return r.Error != nil })
func None[T any](slice []T, predicate func(T) bool) bool {
	return !Any(slice, predicate)
}
