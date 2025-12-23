package vos_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financial/internal/category/domain/vos"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

type CategoryNameSuite struct {
	suite.Suite
}

func TestCategoryNameSuite(t *testing.T) {
	suite.Run(t, new(CategoryNameSuite))
}

func (s *CategoryNameSuite) TestNewCategoryName() {
	scenarios := []struct {
		name   string
		input  string
		expect func(categoryName vos.CategoryName, err error)
	}{
		{
			name:  "deve criar category name com sucesso",
			input: "Transport",
			expect: func(categoryName vos.CategoryName, err error) {
				s.NoError(err)
				s.True(categoryName.Valid)
				s.Equal("Transport", categoryName.String())
			},
		},
		{
			name:  "deve criar category name com nome de 1 caractere",
			input: "A",
			expect: func(categoryName vos.CategoryName, err error) {
				s.NoError(err)
				s.True(categoryName.Valid)
				s.Equal("A", categoryName.String())
			},
		},
		{
			name:  "deve criar category name com 255 caracteres",
			input: strings.Repeat("a", 255),
			expect: func(categoryName vos.CategoryName, err error) {
				s.NoError(err)
				s.True(categoryName.Valid)
				s.Equal(255, len(categoryName.String()))
			},
		},
		{
			name:  "deve criar category name com caracteres unicode",
			input: "Transporte 🚗 São Paulo",
			expect: func(categoryName vos.CategoryName, err error) {
				s.NoError(err)
				s.True(categoryName.Valid)
				s.Equal("Transporte 🚗 São Paulo", categoryName.String())
			},
		},
		{
			name:  "deve fazer trim de espaços em branco",
			input: "  Transport  ",
			expect: func(categoryName vos.CategoryName, err error) {
				s.NoError(err)
				s.True(categoryName.Valid)
				s.Equal("Transport", categoryName.String())
			},
		},
		{
			name:  "deve retornar erro para nome vazio",
			input: "",
			expect: func(categoryName vos.CategoryName, err error) {
				s.Error(err)
				s.False(categoryName.Valid)
				s.ErrorIs(err, customErrors.ErrNameIsRequired)
				s.Contains(err.Error(), "invalid category name")
			},
		},
		{
			name:  "deve retornar erro para nome com apenas espaços",
			input: "   ",
			expect: func(categoryName vos.CategoryName, err error) {
				s.Error(err)
				s.False(categoryName.Valid)
				s.ErrorIs(err, customErrors.ErrNameIsRequired)
			},
		},
		{
			name:  "deve retornar erro para nome com 256 caracteres",
			input: strings.Repeat("a", 256),
			expect: func(categoryName vos.CategoryName, err error) {
				s.Error(err)
				s.False(categoryName.Valid)
				s.ErrorIs(err, customErrors.ErrTooLong)
				s.Contains(err.Error(), "invalid category name")
			},
		},
		{
			name:  "deve retornar erro para nome muito longo (1000 caracteres)",
			input: strings.Repeat("x", 1000),
			expect: func(categoryName vos.CategoryName, err error) {
				s.Error(err)
				s.False(categoryName.Valid)
				s.ErrorIs(err, customErrors.ErrTooLong)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Act
			categoryName, err := vos.NewCategoryName(scenario.input)

			// Assert
			scenario.expect(categoryName, err)
		})
	}
}

func (s *CategoryNameSuite) TestCategoryNameString() {
	scenarios := []struct {
		name     string
		setup    func() vos.CategoryName
		expected string
	}{
		{
			name: "deve retornar string quando válido",
			setup: func() vos.CategoryName {
				name, _ := vos.NewCategoryName("Test")
				return name
			},
			expected: "Test",
		},
		{
			name: "deve retornar string vazia quando inválido",
			setup: func() vos.CategoryName {
				return vos.CategoryName{Value: nil, Valid: false}
			},
			expected: "",
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Arrange
			categoryName := scenario.setup()

			// Act
			result := categoryName.String()

			// Assert
			s.Equal(scenario.expected, result)
		})
	}
}
