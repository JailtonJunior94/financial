//go:build integration
// +build integration

package vos_test

import (
	"testing"

	"github.com/jailtonjunior94/financial/internal/card/domain/vos"
	"github.com/stretchr/testify/suite"
)

type DueDayVOSuite struct {
	suite.Suite
}

func TestDueDayVOSuite(t *testing.T) {
	suite.Run(t, new(DueDayVOSuite))
}

func (s *DueDayVOSuite) TestNewDueDay() {
	scenarios := []struct {
		name   string
		input  int
		expect func(dueDay vos.DueDay, err error)
	}{
		{
			name:  "deve criar due day válido com dia 1",
			input: 1,
			expect: func(dueDay vos.DueDay, err error) {
				s.NoError(err)
				s.True(dueDay.Valid)
				s.Equal(1, dueDay.Int())
			},
		},
		{
			name:  "deve criar due day válido com dia 15",
			input: 15,
			expect: func(dueDay vos.DueDay, err error) {
				s.NoError(err)
				s.True(dueDay.Valid)
				s.Equal(15, dueDay.Int())
			},
		},
		{
			name:  "deve criar due day válido com dia 31",
			input: 31,
			expect: func(dueDay vos.DueDay, err error) {
				s.NoError(err)
				s.True(dueDay.Valid)
				s.Equal(31, dueDay.Int())
			},
		},
		{
			name:  "deve retornar erro para dia 0",
			input: 0,
			expect: func(dueDay vos.DueDay, err error) {
				s.Error(err)
				s.False(dueDay.Valid)
				s.Contains(err.Error(), "invalid due day")
			},
		},
		{
			name:  "deve retornar erro para dia negativo",
			input: -1,
			expect: func(dueDay vos.DueDay, err error) {
				s.Error(err)
				s.False(dueDay.Valid)
				s.Contains(err.Error(), "invalid due day")
			},
		},
		{
			name:  "deve retornar erro para dia 32",
			input: 32,
			expect: func(dueDay vos.DueDay, err error) {
				s.Error(err)
				s.False(dueDay.Valid)
				s.Contains(err.Error(), "invalid due day")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Act
			dueDay, err := vos.NewDueDay(scenario.input)

			// Assert
			scenario.expect(dueDay, err)
		})
	}
}
