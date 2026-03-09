package entities_test

import (
	"testing"
	"time"

	"github.com/jailtonjunior94/financial/internal/category/domain/factories"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/stretchr/testify/require"
)

func TestCreateSubcategory(t *testing.T) {
	validUserID := "550e8400-e29b-41d4-a716-446655440000"
	validCategoryID := "660e8400-e29b-41d4-a716-446655440001"

	t.Run("T6 valid subcategory creation", func(t *testing.T) {
		sub, err := factories.CreateSubcategory(validUserID, validCategoryID, "Uber", 1)
		require.NoError(t, err)
		require.NotNil(t, sub)
		require.Equal(t, "Uber", sub.Name.String())
		require.Equal(t, uint(1), sub.Sequence.Value())
		require.NotEmpty(t, sub.ID.String())
		require.False(t, sub.CreatedAt.ValueOr(time.Time{}).IsZero())
	})

	t.Run("T7 empty name returns ErrNameIsRequired", func(t *testing.T) {
		sub, err := factories.CreateSubcategory(validUserID, validCategoryID, "", 1)
		require.Error(t, err)
		require.Nil(t, sub)
		require.ErrorIs(t, err, customErrors.ErrNameIsRequired)
	})

	t.Run("T8 name >255 chars returns error", func(t *testing.T) {
		longName := string(make([]byte, 256))
		sub, err := factories.CreateSubcategory(validUserID, validCategoryID, longName, 1)
		require.Error(t, err)
		require.Nil(t, sub)
	})

	t.Run("T9 sequence 0 returns ErrSequenceIsRequired", func(t *testing.T) {
		sub, err := factories.CreateSubcategory(validUserID, validCategoryID, "Uber", 0)
		require.Error(t, err)
		require.Nil(t, sub)
		require.ErrorIs(t, err, customErrors.ErrSequenceIsRequired)
	})

	t.Run("T10 sequence >1000 returns error", func(t *testing.T) {
		sub, err := factories.CreateSubcategory(validUserID, validCategoryID, "Uber", 1001)
		require.Error(t, err)
		require.Nil(t, sub)
	})

	t.Run("T11 invalid userID returns error", func(t *testing.T) {
		sub, err := factories.CreateSubcategory("invalid-uuid", validCategoryID, "Uber", 1)
		require.Error(t, err)
		require.Nil(t, sub)
	})

	t.Run("T12 invalid categoryID returns error", func(t *testing.T) {
		sub, err := factories.CreateSubcategory(validUserID, "invalid-uuid", "Uber", 1)
		require.Error(t, err)
		require.Nil(t, sub)
	})

	t.Run("T13 Update valid sets UpdatedAt", func(t *testing.T) {
		sub, err := factories.CreateSubcategory(validUserID, validCategoryID, "Uber", 1)
		require.NoError(t, err)

		err = sub.Update("Taxi", 2)
		require.NoError(t, err)
		require.Equal(t, "Taxi", sub.Name.String())
		require.Equal(t, uint(2), sub.Sequence.Value())
		require.False(t, sub.UpdatedAt.ValueOr(time.Time{}).IsZero())
	})

	t.Run("T14 Update with empty name returns error", func(t *testing.T) {
		sub, err := factories.CreateSubcategory(validUserID, validCategoryID, "Uber", 1)
		require.NoError(t, err)

		err = sub.Update("", 1)
		require.Error(t, err)
	})

	t.Run("T15 Delete sets DeletedAt", func(t *testing.T) {
		sub, err := factories.CreateSubcategory(validUserID, validCategoryID, "Uber", 1)
		require.NoError(t, err)

		sub.Delete()
		require.False(t, sub.DeletedAt.ValueOr(time.Time{}).IsZero())
	})
}
