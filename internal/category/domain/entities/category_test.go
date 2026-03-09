package entities_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/vos"
)

type CategoryEntitySuite struct {
	suite.Suite
}

func TestCategoryEntitySuite(t *testing.T) {
	suite.Run(t, new(CategoryEntitySuite))
}

func (s *CategoryEntitySuite) TestNewCategory() {
	scenarios := []struct {
		name string
		args struct {
			userID   sharedVos.UUID
			name     vos.CategoryName
			sequence vos.CategorySequence
		}
		expect func(category *entities.Category, err error)
	}{
		{
			name: "deve criar categoria com sucesso",
			args: struct {
				userID   sharedVos.UUID
				name     vos.CategoryName
				sequence vos.CategorySequence
			}{
				userID:   s.createUUID(),
				name:     s.createCategoryName("Transport"),
				sequence: s.createSequence(1),
			},
			expect: func(category *entities.Category, err error) {
				s.NoError(err)
				s.NotNil(category)
				s.Equal("Transport", category.Name.String())
				s.Equal(uint(1), category.Sequence.Value())
				s.False(category.CreatedAt.ValueOr(time.Time{}).IsZero())
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			category, err := entities.NewCategory(
				scenario.args.userID,
				scenario.args.name,
				scenario.args.sequence,
			)
			scenario.expect(category, err)
		})
	}
}

func (s *CategoryEntitySuite) TestUpdate() {
	scenarios := []struct {
		name string
		args struct {
			name     string
			sequence uint
		}
		expect func(category *entities.Category, err error)
	}{
		{
			name: "deve atualizar categoria com sucesso",
			args: struct {
				name     string
				sequence uint
			}{
				name:     "Updated Name",
				sequence: 10,
			},
			expect: func(category *entities.Category, err error) {
				s.NoError(err)
				s.NotNil(category)
				s.Equal("Updated Name", category.Name.String())
				s.Equal(uint(10), category.Sequence.Value())
				s.False(category.UpdatedAt.ValueOr(time.Time{}).IsZero())
			},
		},
		{
			name: "deve retornar erro ao atualizar com nome vazio",
			args: struct {
				name     string
				sequence uint
			}{
				name:     "",
				sequence: 1,
			},
			expect: func(category *entities.Category, err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid category name")
			},
		},
		{
			name: "deve retornar erro ao atualizar com sequence 0",
			args: struct {
				name     string
				sequence uint
			}{
				name:     "Valid Name",
				sequence: 0,
			},
			expect: func(category *entities.Category, err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid category sequence")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			userID := s.createUUID()
			name := s.createCategoryName("Original")
			sequence := s.createSequence(1)
			category, _ := entities.NewCategory(userID, name, sequence)

			err := category.Update(scenario.args.name, scenario.args.sequence)
			scenario.expect(category, err)
		})
	}
}

func (s *CategoryEntitySuite) TestDelete() {
	s.Run("deve deletar categoria (soft delete)", func() {
		userID := s.createUUID()
		name := s.createCategoryName("To Delete")
		sequence := s.createSequence(1)
		category, _ := entities.NewCategory(userID, name, sequence)

		category.Delete()

		s.False(category.DeletedAt.ValueOr(time.Time{}).IsZero())
	})
}

// Helper methods.
func (s *CategoryEntitySuite) createUUID() sharedVos.UUID {
	uuid, _ := sharedVos.NewUUID()
	return uuid
}

func (s *CategoryEntitySuite) createCategoryName(name string) vos.CategoryName {
	categoryName, _ := vos.NewCategoryName(name)
	return categoryName
}

func (s *CategoryEntitySuite) createSequence(seq uint) vos.CategorySequence {
	sequence, _ := vos.NewCategorySequence(seq)
	return sequence
}
