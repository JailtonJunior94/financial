package vos_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financial/internal/category/domain/vos"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

type CategorySequenceSuite struct {
	suite.Suite
}

func TestCategorySequenceSuite(t *testing.T) {
	suite.Run(t, new(CategorySequenceSuite))
}

func (s *CategorySequenceSuite) TestNewCategorySequence() {
	scenarios := []struct {
		name   string
		input  uint
		expect func(sequence vos.CategorySequence, err error)
	}{
		{
			name:  "deve criar sequence com valor 1",
			input: 1,
			expect: func(sequence vos.CategorySequence, err error) {
				s.NoError(err)
				s.True(sequence.Valid)
				s.Equal(uint(1), sequence.Value())
			},
		},
		{
			name:  "deve criar sequence com valor 10",
			input: 10,
			expect: func(sequence vos.CategorySequence, err error) {
				s.NoError(err)
				s.True(sequence.Valid)
				s.Equal(uint(10), sequence.Value())
			},
		},
		{
			name:  "deve criar sequence com valor 500",
			input: 500,
			expect: func(sequence vos.CategorySequence, err error) {
				s.NoError(err)
				s.True(sequence.Valid)
				s.Equal(uint(500), sequence.Value())
			},
		},
		{
			name:  "deve criar sequence com valor máximo permitido (1000)",
			input: 1000,
			expect: func(sequence vos.CategorySequence, err error) {
				s.NoError(err)
				s.True(sequence.Valid)
				s.Equal(uint(1000), sequence.Value())
			},
		},
		{
			name:  "deve retornar erro para sequence 0",
			input: 0,
			expect: func(sequence vos.CategorySequence, err error) {
				s.Error(err)
				s.False(sequence.Valid)
				s.ErrorIs(err, customErrors.ErrSequenceIsRequired)
				s.Contains(err.Error(), "invalid category sequence")
			},
		},
		{
			name:  "deve retornar erro para sequence maior que 1000",
			input: 1001,
			expect: func(sequence vos.CategorySequence, err error) {
				s.Error(err)
				s.False(sequence.Valid)
				s.ErrorIs(err, customErrors.ErrSequenceTooLarge)
				s.Contains(err.Error(), "invalid category sequence")
			},
		},
		{
			name:  "deve retornar erro para sequence muito grande (10000)",
			input: 10000,
			expect: func(sequence vos.CategorySequence, err error) {
				s.Error(err)
				s.False(sequence.Valid)
				s.ErrorIs(err, customErrors.ErrSequenceTooLarge)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Act
			sequence, err := vos.NewCategorySequence(scenario.input)

			// Assert
			scenario.expect(sequence, err)
		})
	}
}

func (s *CategorySequenceSuite) TestCategorySequenceValue() {
	scenarios := []struct {
		name     string
		setup    func() vos.CategorySequence
		expected uint
	}{
		{
			name: "deve retornar valor quando válido",
			setup: func() vos.CategorySequence {
				seq, _ := vos.NewCategorySequence(42)
				return seq
			},
			expected: 42,
		},
		{
			name: "deve retornar 0 quando inválido",
			setup: func() vos.CategorySequence {
				return vos.CategorySequence{Sequence: nil, Valid: false}
			},
			expected: 0,
		},
		{
			name: "deve retornar 0 quando Sequence é nil",
			setup: func() vos.CategorySequence {
				return vos.CategorySequence{Sequence: nil, Valid: true}
			},
			expected: 0,
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Arrange
			sequence := scenario.setup()

			// Act
			result := sequence.Value()

			// Assert
			s.Equal(scenario.expected, result)
		})
	}
}
