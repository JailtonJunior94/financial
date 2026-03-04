package entities_test

import (
	"testing"

	"github.com/jailtonjunior94/financial/internal/user/domain/entities"
	"github.com/jailtonjunior94/financial/internal/user/domain/vos"
	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestUser(t *testing.T) *entities.User {
	t.Helper()
	name, err := vos.NewUserName("Test User")
	require.NoError(t, err)
	email, err := vos.NewEmail("test@email.com")
	require.NoError(t, err)
	user, err := entities.NewUser(name, email)
	require.NoError(t, err)
	return user
}

func TestUser_UpdateName(t *testing.T) {
	t.Run("should update name successfully", func(t *testing.T) {
		user := newTestUser(t)
		newName, err := vos.NewUserName("Updated Name")
		require.NoError(t, err)

		user.UpdateName(newName)

		assert.Equal(t, "Updated Name", user.Name.String())
	})
}

func TestUser_UpdateEmail(t *testing.T) {
	t.Run("should update email successfully", func(t *testing.T) {
		user := newTestUser(t)
		newEmail, err := vos.NewEmail("updated@email.com")
		require.NoError(t, err)

		user.UpdateEmail(newEmail)

		assert.Equal(t, "updated@email.com", user.Email.String())
	})
}

func TestUser_MarkAsDeleted(t *testing.T) {
	t.Run("should set deleted_at timestamp", func(t *testing.T) {
		user := newTestUser(t)
		assert.False(t, user.DeletedAt.IsValid())

		user.MarkAsDeleted()

		assert.True(t, user.DeletedAt.IsValid())
		deletedAt, ok := user.DeletedAt.Get()
		assert.True(t, ok)
		assert.False(t, deletedAt.IsZero())
	})
}

func TestUser_SetPassword(t *testing.T) {
	t.Run("should set password successfully", func(t *testing.T) {
		user := newTestUser(t)
		hash := "validhash123456789012345"

		err := user.SetPassword(hash)

		require.NoError(t, err)
		assert.Equal(t, hash, user.Password)
	})

	t.Run("should return ErrPasswordIsRequired for empty password", func(t *testing.T) {
		user := newTestUser(t)

		err := user.SetPassword("")

		assert.ErrorIs(t, err, customerrors.ErrPasswordIsRequired)
	})

	t.Run("should return error for short password hash", func(t *testing.T) {
		user := newTestUser(t)

		err := user.SetPassword("short")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid password hash format")
	})
}
