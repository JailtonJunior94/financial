package mathutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClamp(t *testing.T) {
	t.Run("value within range", func(t *testing.T) {
		assert.Equal(t, 50, Clamp(50, 0, 100))
	})

	t.Run("value below minimum", func(t *testing.T) {
		assert.Equal(t, 0, Clamp(-5, 0, 100))
	})

	t.Run("value above maximum", func(t *testing.T) {
		assert.Equal(t, 100, Clamp(150, 0, 100))
	})

	t.Run("float values", func(t *testing.T) {
		assert.Equal(t, 5.5, Clamp(5.5, 0.0, 10.0))
		assert.Equal(t, 0.0, Clamp(-1.5, 0.0, 10.0))
		assert.Equal(t, 10.0, Clamp(15.5, 0.0, 10.0))
	})
}

func TestMaxOf(t *testing.T) {
	t.Run("find maximum", func(t *testing.T) {
		assert.Equal(t, 30, MaxOf(10, 25, 15, 30, 5))
	})

	t.Run("single value", func(t *testing.T) {
		assert.Equal(t, 42, MaxOf(42))
	})

	t.Run("empty returns zero", func(t *testing.T) {
		assert.Equal(t, 0, MaxOf[int]())
	})

	t.Run("negative numbers", func(t *testing.T) {
		assert.Equal(t, -1, MaxOf(-5, -3, -10, -1))
	})
}

func TestMinOf(t *testing.T) {
	t.Run("find minimum", func(t *testing.T) {
		assert.Equal(t, 5, MinOf(10, 25, 15, 30, 5))
	})

	t.Run("single value", func(t *testing.T) {
		assert.Equal(t, 42, MinOf(42))
	})

	t.Run("empty returns zero", func(t *testing.T) {
		assert.Equal(t, 0, MinOf[int]())
	})

	t.Run("negative numbers", func(t *testing.T) {
		assert.Equal(t, -10, MinOf(-5, -3, -10, -1))
	})
}

func TestSafeSlice(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	t.Run("normal range", func(t *testing.T) {
		assert.Equal(t, []int{2, 3}, SafeSlice(items, 1, 3))
	})

	t.Run("end exceeds length", func(t *testing.T) {
		assert.Equal(t, []int{1, 2, 3, 4, 5}, SafeSlice(items, 0, 100))
	})

	t.Run("start exceeds length", func(t *testing.T) {
		assert.Empty(t, SafeSlice(items, 10, 20))
	})

	t.Run("negative start", func(t *testing.T) {
		assert.Equal(t, []int{1, 2, 3}, SafeSlice(items, -1, 3))
	})

	t.Run("start >= end", func(t *testing.T) {
		assert.Empty(t, SafeSlice(items, 5, 3))
	})
}

func TestCapLimit(t *testing.T) {
	t.Run("value under limit", func(t *testing.T) {
		assert.Equal(t, 50, CapLimit(50, 100))
	})

	t.Run("value over limit", func(t *testing.T) {
		assert.Equal(t, 100, CapLimit(150, 100))
	})

	t.Run("value equals limit", func(t *testing.T) {
		assert.Equal(t, 100, CapLimit(100, 100))
	})
}

func TestAtLeast(t *testing.T) {
	t.Run("value above minimum", func(t *testing.T) {
		assert.Equal(t, 50, AtLeast(50, 10))
	})

	t.Run("value below minimum", func(t *testing.T) {
		assert.Equal(t, 10, AtLeast(5, 10))
	})

	t.Run("value equals minimum", func(t *testing.T) {
		assert.Equal(t, 10, AtLeast(10, 10))
	})

	t.Run("ensure positive", func(t *testing.T) {
		assert.Equal(t, 1, AtLeast(0, 1))
		assert.Equal(t, 1, AtLeast(-5, 1))
	})
}
