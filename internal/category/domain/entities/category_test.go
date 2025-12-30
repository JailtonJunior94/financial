//go:build integration
// +build integration

package entities_test

import (
	"testing"

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
		name   string
		args   struct {
			userID   sharedVos.UUID
			parentID *sharedVos.UUID
			name     vos.CategoryName
			sequence vos.CategorySequence
		}
		expect func(category *entities.Category, err error)
	}{
		{
			name: "deve criar categoria com sucesso",
			args: struct {
				userID   sharedVos.UUID
				parentID *sharedVos.UUID
				name     vos.CategoryName
				sequence vos.CategorySequence
			}{
				userID:   s.createUUID(),
				parentID: nil,
				name:     s.createCategoryName("Transport"),
				sequence: s.createSequence(1),
			},
			expect: func(category *entities.Category, err error) {
				s.NoError(err)
				s.NotNil(category)
				s.Equal("Transport", category.Name.String())
				s.Equal(uint(1), category.Sequence.Value())
				s.Nil(category.ParentID)
				s.NotNil(category.CreatedAt.Time)
			},
		},
		{
			name: "deve criar categoria com parent_id",
			args: struct {
				userID   sharedVos.UUID
				parentID *sharedVos.UUID
				name     vos.CategoryName
				sequence vos.CategorySequence
			}{
				userID:   s.createUUID(),
				parentID: s.createUUIDPtr(),
				name:     s.createCategoryName("Uber"),
				sequence: s.createSequence(2),
			},
			expect: func(category *entities.Category, err error) {
				s.NoError(err)
				s.NotNil(category)
				s.Equal("Uber", category.Name.String())
				s.NotNil(category.ParentID)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Act
			category, err := entities.NewCategory(
				scenario.args.userID,
				scenario.args.parentID,
				scenario.args.name,
				scenario.args.sequence,
			)

			// Assert
			scenario.expect(category, err)
		})
	}
}

func (s *CategoryEntitySuite) TestUpdate() {
	scenarios := []struct {
		name   string
		args   struct {
			name     string
			sequence uint
			parentID *sharedVos.UUID
		}
		expect func(category *entities.Category, err error)
	}{
		{
			name: "deve atualizar categoria com sucesso",
			args: struct {
				name     string
				sequence uint
				parentID *sharedVos.UUID
			}{
				name:     "Updated Name",
				sequence: 10,
				parentID: nil,
			},
			expect: func(category *entities.Category, err error) {
				s.NoError(err)
				s.NotNil(category)
				s.Equal("Updated Name", category.Name.String())
				s.Equal(uint(10), category.Sequence.Value())
				s.NotNil(category.UpdatedAt.Time)
			},
		},
		{
			name: "deve atualizar categoria com parent_id",
			args: struct {
				name     string
				sequence uint
				parentID *sharedVos.UUID
			}{
				name:     "New Name",
				sequence: 5,
				parentID: s.createUUIDPtr(),
			},
			expect: func(category *entities.Category, err error) {
				s.NoError(err)
				s.NotNil(category)
				s.Equal("New Name", category.Name.String())
				s.NotNil(category.ParentID)
			},
		},
		{
			name: "deve retornar erro ao atualizar com nome vazio",
			args: struct {
				name     string
				sequence uint
				parentID *sharedVos.UUID
			}{
				name:     "",
				sequence: 1,
				parentID: nil,
			},
			expect: func(category *entities.Category, err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid category name")
			},
		},
		{
			name: "deve retornar erro ao atualizar com nome muito longo",
			args: struct {
				name     string
				sequence uint
				parentID *sharedVos.UUID
			}{
				name:     string(make([]byte, 256)),
				sequence: 1,
				parentID: nil,
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
				parentID *sharedVos.UUID
			}{
				name:     "Valid Name",
				sequence: 0,
				parentID: nil,
			},
			expect: func(category *entities.Category, err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid category sequence")
			},
		},
		{
			name: "deve retornar erro ao atualizar com sequence maior que 1000",
			args: struct {
				name     string
				sequence uint
				parentID *sharedVos.UUID
			}{
				name:     "Valid Name",
				sequence: 1001,
				parentID: nil,
			},
			expect: func(category *entities.Category, err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid category sequence")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Arrange
			userID := s.createUUID()
			name := s.createCategoryName("Original")
			sequence := s.createSequence(1)
			category, _ := entities.NewCategory(userID, nil, name, sequence)

			// Act
			err := category.Update(scenario.args.name, scenario.args.sequence, scenario.args.parentID)

			// Assert
			scenario.expect(category, err)
		})
	}
}

func (s *CategoryEntitySuite) TestDelete() {
	s.Run("deve deletar categoria (soft delete)", func() {
		// Arrange
		userID := s.createUUID()
		name := s.createCategoryName("To Delete")
		sequence := s.createSequence(1)
		category, _ := entities.NewCategory(userID, nil, name, sequence)

		// Act
		result := category.Delete()

		// Assert
		s.NotNil(result)
		s.NotNil(result.DeletedAt.Time)
		s.Equal(category, result)
	})
}

func (s *CategoryEntitySuite) TestAddChildrens() {
	scenarios := []struct {
		name     string
		children []entities.Category
		expect   func(category *entities.Category)
	}{
		{
			name:     "deve adicionar lista vazia de children",
			children: []entities.Category{},
			expect: func(category *entities.Category) {
				s.NotNil(category.Children)
				s.Len(category.Children, 0)
			},
		},
		{
			name: "deve adicionar um child",
			children: []entities.Category{
				s.createCategory("Child 1", 1),
			},
			expect: func(category *entities.Category) {
				s.Len(category.Children, 1)
				s.Equal("Child 1", category.Children[0].Name.String())
			},
		},
		{
			name: "deve adicionar múltiplos children",
			children: []entities.Category{
				s.createCategory("Child 1", 1),
				s.createCategory("Child 2", 2),
				s.createCategory("Child 3", 3),
			},
			expect: func(category *entities.Category) {
				s.Len(category.Children, 3)
				s.Equal("Child 1", category.Children[0].Name.String())
				s.Equal("Child 2", category.Children[1].Name.String())
				s.Equal("Child 3", category.Children[2].Name.String())
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Arrange
			userID := s.createUUID()
			name := s.createCategoryName("Parent")
			sequence := s.createSequence(1)
			category, _ := entities.NewCategory(userID, nil, name, sequence)

			// Act
			category.AddChildrens(scenario.children)

			// Assert
			scenario.expect(category)
		})
	}
}

// Helper methods
func (s *CategoryEntitySuite) createUUID() sharedVos.UUID {
	uuid, _ := sharedVos.NewUUID()
	return uuid
}

func (s *CategoryEntitySuite) createUUIDPtr() *sharedVos.UUID {
	uuid, _ := sharedVos.NewUUID()
	return &uuid
}

func (s *CategoryEntitySuite) createCategoryName(name string) vos.CategoryName {
	categoryName, _ := vos.NewCategoryName(name)
	return categoryName
}

func (s *CategoryEntitySuite) createSequence(seq uint) vos.CategorySequence {
	sequence, _ := vos.NewCategorySequence(seq)
	return sequence
}

func (s *CategoryEntitySuite) createCategory(name string, seq uint) entities.Category {
	userID := s.createUUID()
	categoryName := s.createCategoryName(name)
	sequence := s.createSequence(seq)
	category, _ := entities.NewCategory(userID, nil, categoryName, sequence)
	return *category
}
