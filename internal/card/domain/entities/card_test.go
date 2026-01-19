//go:build integration
// +build integration

package entities_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/card/domain/entities"
	"github.com/jailtonjunior94/financial/internal/card/domain/vos"
)

type CardEntitySuite struct {
	suite.Suite
}

func TestCardEntitySuite(t *testing.T) {
	suite.Run(t, new(CardEntitySuite))
}

func (s *CardEntitySuite) TestNewCard() {
	scenarios := []struct {
		name string
		args struct {
			userID sharedVos.UUID
			name   vos.CardName
			dueDay vos.DueDay
		}
		expect func(card *entities.Card, err error)
	}{
		{
			name: "deve criar cartão com sucesso",
			args: struct {
				userID sharedVos.UUID
				name   vos.CardName
				dueDay vos.DueDay
			}{
				userID: s.createUUID(),
				name:   s.createCardName("Nubank"),
				dueDay: s.createDueDay(15),
			},
			expect: func(card *entities.Card, err error) {
				s.NoError(err)
				s.NotNil(card)
				s.Equal("Nubank", card.Name.String())
				s.Equal(15, card.DueDay.Int())
				s.False(card.CreatedAt.ValueOr(time.Time{}).IsZero())
			},
		},
		{
			name: "deve criar cartão com dia de vencimento 1",
			args: struct {
				userID sharedVos.UUID
				name   vos.CardName
				dueDay vos.DueDay
			}{
				userID: s.createUUID(),
				name:   s.createCardName("Banco do Brasil"),
				dueDay: s.createDueDay(1),
			},
			expect: func(card *entities.Card, err error) {
				s.NoError(err)
				s.NotNil(card)
				s.Equal("Banco do Brasil", card.Name.String())
				s.Equal(1, card.DueDay.Int())
			},
		},
		{
			name: "deve criar cartão com dia de vencimento 31",
			args: struct {
				userID sharedVos.UUID
				name   vos.CardName
				dueDay vos.DueDay
			}{
				userID: s.createUUID(),
				name:   s.createCardName("Inter"),
				dueDay: s.createDueDay(31),
			},
			expect: func(card *entities.Card, err error) {
				s.NoError(err)
				s.NotNil(card)
				s.Equal("Inter", card.Name.String())
				s.Equal(31, card.DueDay.Int())
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Act
			card, err := entities.NewCard(
				scenario.args.userID,
				scenario.args.name,
				scenario.args.dueDay,
			)

			// Assert
			scenario.expect(card, err)
		})
	}
}

func (s *CardEntitySuite) TestUpdate() {
	scenarios := []struct {
		name string
		args struct {
			name   string
			dueDay int
		}
		expect func(card *entities.Card, err error)
	}{
		{
			name: "deve atualizar cartão com sucesso",
			args: struct {
				name   string
				dueDay int
			}{
				name:   "Updated Card",
				dueDay: 20,
			},
			expect: func(card *entities.Card, err error) {
				s.NoError(err)
				s.NotNil(card)
				s.Equal("Updated Card", card.Name.String())
				s.Equal(20, card.DueDay.Int())
				s.False(card.UpdatedAt.ValueOr(time.Time{}).IsZero())
			},
		},
		{
			name: "deve retornar erro ao atualizar com nome vazio",
			args: struct {
				name   string
				dueDay int
			}{
				name:   "",
				dueDay: 15,
			},
			expect: func(card *entities.Card, err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid card name")
			},
		},
		{
			name: "deve retornar erro ao atualizar com nome muito longo",
			args: struct {
				name   string
				dueDay int
			}{
				name:   string(make([]byte, 256)),
				dueDay: 15,
			},
			expect: func(card *entities.Card, err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid card name")
			},
		},
		{
			name: "deve retornar erro ao atualizar com due_day 0",
			args: struct {
				name   string
				dueDay int
			}{
				name:   "Valid Name",
				dueDay: 0,
			},
			expect: func(card *entities.Card, err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid due day")
			},
		},
		{
			name: "deve retornar erro ao atualizar com due_day maior que 31",
			args: struct {
				name   string
				dueDay int
			}{
				name:   "Valid Name",
				dueDay: 32,
			},
			expect: func(card *entities.Card, err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid due day")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Arrange
			userID := s.createUUID()
			name := s.createCardName("Original Card")
			dueDay := s.createDueDay(10)
			card, _ := entities.NewCard(userID, name, dueDay)

			// Act
			err := card.Update(scenario.args.name, scenario.args.dueDay, 7) // Default closing offset days

			// Assert
			scenario.expect(card, err)
		})
	}
}

func (s *CardEntitySuite) TestDelete() {
	s.Run("deve deletar cartão (soft delete)", func() {
		// Arrange
		userID := s.createUUID()
		name := s.createCardName("Card to Delete")
		dueDay := s.createDueDay(15)
		card, _ := entities.NewCard(userID, name, dueDay)

		// Act
		result := card.Delete()

		// Assert
		s.NotNil(result)
		s.False(result.DeletedAt.ValueOr(time.Time{}).IsZero())
		s.Equal(card, result)
	})
}

// Helper methods
func (s *CardEntitySuite) createUUID() sharedVos.UUID {
	uuid, _ := sharedVos.NewUUID()
	return uuid
}

func (s *CardEntitySuite) createCardName(name string) vos.CardName {
	cardName, _ := vos.NewCardName(name)
	return cardName
}

func (s *CardEntitySuite) createDueDay(day int) vos.DueDay {
	dueDay, _ := vos.NewDueDay(day)
	return dueDay
}
