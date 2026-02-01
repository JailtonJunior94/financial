package sliceutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	t.Run("transform integers to strings", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5}
		strings := Map(numbers, func(n int) string {
			return string(rune('A' + n - 1))
		})

		assert.Equal(t, []string{"A", "B", "C", "D", "E"}, strings)
	})

	t.Run("nil slice returns nil", func(t *testing.T) {
		var numbers []int
		result := Map(numbers, func(n int) string { return "" })

		assert.Nil(t, result)
	})

	t.Run("extract field from struct", func(t *testing.T) {
		type User struct{ ID string }
		users := []User{{"1"}, {"2"}, {"3"}}

		ids := Map(users, func(u User) string { return u.ID })

		assert.Equal(t, []string{"1", "2", "3"}, ids)
	})
}

func TestFilter(t *testing.T) {
	t.Run("filter even numbers", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5, 6}
		evens := Filter(numbers, func(n int) bool { return n%2 == 0 })

		assert.Equal(t, []int{2, 4, 6}, evens)
	})

	t.Run("nil slice returns nil", func(t *testing.T) {
		var numbers []int
		result := Filter(numbers, func(n int) bool { return true })

		assert.Nil(t, result)
	})

	t.Run("no matches returns empty slice", func(t *testing.T) {
		numbers := []int{1, 3, 5}
		evens := Filter(numbers, func(n int) bool { return n%2 == 0 })

		assert.Empty(t, evens)
	})
}

func TestContains(t *testing.T) {
	t.Run("element exists", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5}
		exists := Contains(numbers, func(n int) bool { return n == 3 })

		assert.True(t, exists)
	})

	t.Run("element does not exist", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5}
		exists := Contains(numbers, func(n int) bool { return n == 10 })

		assert.False(t, exists)
	})
}

func TestFind(t *testing.T) {
	t.Run("find existing element", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5}
		result := Find(numbers, func(n int) bool { return n == 3 })

		assert.NotNil(t, result)
		assert.Equal(t, 3, *result)
	})

	t.Run("element not found returns nil", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5}
		result := Find(numbers, func(n int) bool { return n == 10 })

		assert.Nil(t, result)
	})
}

func TestReduce(t *testing.T) {
	t.Run("sum numbers", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5}
		sum := Reduce(numbers, 0, func(acc int, n int) int { return acc + n })

		assert.Equal(t, 15, sum)
	})

	t.Run("concatenate strings", func(t *testing.T) {
		words := []string{"Hello", " ", "World"}
		result := Reduce(words, "", func(acc string, s string) string { return acc + s })

		assert.Equal(t, "Hello World", result)
	})
}

func TestChunk(t *testing.T) {
	t.Run("divide evenly", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5, 6}
		chunks := Chunk(numbers, 2)

		assert.Len(t, chunks, 3)
		assert.Equal(t, []int{1, 2}, chunks[0])
		assert.Equal(t, []int{3, 4}, chunks[1])
		assert.Equal(t, []int{5, 6}, chunks[2])
	})

	t.Run("last chunk smaller", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5}
		chunks := Chunk(numbers, 2)

		assert.Len(t, chunks, 3)
		assert.Equal(t, []int{5}, chunks[2])
	})

	t.Run("size zero returns nil", func(t *testing.T) {
		numbers := []int{1, 2, 3}
		chunks := Chunk(numbers, 0)

		assert.Nil(t, chunks)
	})
}

func TestUnique(t *testing.T) {
	t.Run("remove duplicates", func(t *testing.T) {
		numbers := []int{1, 2, 2, 3, 3, 3, 4, 5, 5}
		unique := Unique(numbers)

		assert.Equal(t, []int{1, 2, 3, 4, 5}, unique)
	})

	t.Run("maintains order", func(t *testing.T) {
		words := []string{"b", "a", "b", "c", "a"}
		unique := Unique(words)

		assert.Equal(t, []string{"b", "a", "c"}, unique)
	})

	t.Run("nil slice returns nil", func(t *testing.T) {
		var numbers []int
		result := Unique(numbers)

		assert.Nil(t, result)
	})
}

func TestPartition(t *testing.T) {
	t.Run("partition by predicate", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5, 6}
		evens, odds := Partition(numbers, func(n int) bool { return n%2 == 0 })

		assert.Equal(t, []int{2, 4, 6}, evens)
		assert.Equal(t, []int{1, 3, 5}, odds)
	})

	t.Run("all match", func(t *testing.T) {
		numbers := []int{2, 4, 6}
		evens, odds := Partition(numbers, func(n int) bool { return n%2 == 0 })

		assert.Equal(t, []int{2, 4, 6}, evens)
		assert.Empty(t, odds)
	})
}

func TestGroupBy(t *testing.T) {
	t.Run("group by modulo", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5, 6}
		grouped := GroupBy(numbers, func(n int) int { return n % 3 })

		assert.Len(t, grouped, 3)
		assert.Equal(t, []int{3, 6}, grouped[0])
		assert.Equal(t, []int{1, 4}, grouped[1])
		assert.Equal(t, []int{2, 5}, grouped[2])
	})

	t.Run("group structs by field", func(t *testing.T) {
		type User struct {
			Name string
			Role string
		}
		users := []User{
			{"Alice", "admin"},
			{"Bob", "user"},
			{"Charlie", "admin"},
		}

		byRole := GroupBy(users, func(u User) string { return u.Role })

		assert.Len(t, byRole["admin"], 2)
		assert.Len(t, byRole["user"], 1)
	})
}

func TestAny(t *testing.T) {
	t.Run("at least one matches", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5}
		hasEven := Any(numbers, func(n int) bool { return n%2 == 0 })

		assert.True(t, hasEven)
	})

	t.Run("none match", func(t *testing.T) {
		numbers := []int{1, 3, 5}
		hasEven := Any(numbers, func(n int) bool { return n%2 == 0 })

		assert.False(t, hasEven)
	})
}

func TestAll(t *testing.T) {
	t.Run("all match", func(t *testing.T) {
		numbers := []int{2, 4, 6}
		allEven := All(numbers, func(n int) bool { return n%2 == 0 })

		assert.True(t, allEven)
	})

	t.Run("not all match", func(t *testing.T) {
		numbers := []int{2, 3, 4}
		allEven := All(numbers, func(n int) bool { return n%2 == 0 })

		assert.False(t, allEven)
	})

	t.Run("empty slice returns true", func(t *testing.T) {
		numbers := []int{}
		allEven := All(numbers, func(n int) bool { return n%2 == 0 })

		assert.True(t, allEven)
	})
}

func TestNone(t *testing.T) {
	t.Run("none match", func(t *testing.T) {
		numbers := []int{1, 3, 5}
		noEvens := None(numbers, func(n int) bool { return n%2 == 0 })

		assert.True(t, noEvens)
	})

	t.Run("at least one matches", func(t *testing.T) {
		numbers := []int{1, 2, 3}
		noEvens := None(numbers, func(n int) bool { return n%2 == 0 })

		assert.False(t, noEvens)
	})
}
