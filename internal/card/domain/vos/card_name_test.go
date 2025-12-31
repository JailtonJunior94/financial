//go:build integration
// +build integration

package vos_test

import (
	"testing"

	"github.com/jailtonjunior94/financial/internal/card/domain/vos"
	"github.com/stretchr/testify/suite"
)

type CardNameVOSuite struct {
	suite.Suite
}

func TestCardNameVOSuite(t *testing.T) {
	suite.Run(t, new(CardNameVOSuite))
}

func (s *CardNameVOSuite) TestNewCardName() {
	scenarios := []struct {
		name   string
		input  string
		expect func(cardName vos.CardName, err error)
	}{
		{
			name:  "deve criar card name válido",
			input: "Nubank",
			expect: func(cardName vos.CardName, err error) {
				s.NoError(err)
				s.True(cardName.Valid)
				s.Equal("Nubank", cardName.String())
			},
		},
		{
			name:  "deve criar card name com espaços nas extremidades",
			input: "  Inter  ",
			expect: func(cardName vos.CardName, err error) {
				s.NoError(err)
				s.True(cardName.Valid)
				s.Equal("Inter", cardName.String())
			},
		},
		{
			name:  "deve retornar erro para nome vazio",
			input: "",
			expect: func(cardName vos.CardName, err error) {
				s.Error(err)
				s.False(cardName.Valid)
				s.Contains(err.Error(), "invalid card name")
			},
		},
		{
			name:  "deve retornar erro para nome com apenas espaços",
			input: "   ",
			expect: func(cardName vos.CardName, err error) {
				s.Error(err)
				s.False(cardName.Valid)
				s.Contains(err.Error(), "invalid card name")
			},
		},
		{
			name:  "deve retornar erro para nome muito longo",
			input: string(make([]byte, 256)),
			expect: func(cardName vos.CardName, err error) {
				s.Error(err)
				s.False(cardName.Valid)
				s.Contains(err.Error(), "invalid card name")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Act
			cardName, err := vos.NewCardName(scenario.input)

			// Assert
			scenario.expect(cardName, err)
		})
	}
}
