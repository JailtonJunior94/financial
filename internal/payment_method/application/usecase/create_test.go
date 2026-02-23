package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/jailtonjunior94/financial/internal/payment_method/application/dtos"
	repositoryMock "github.com/jailtonjunior94/financial/internal/payment_method/infrastructure/repositories/mocks"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
)

type CreatePaymentMethodUseCaseSuite struct {
	suite.Suite

	ctx                     context.Context
	obs                     observability.Observability
	paymentMethodRepository *repositoryMock.PaymentMethodRepository
}

func TestCreatePaymentMethodUseCaseSuite(t *testing.T) {
	suite.Run(t, new(CreatePaymentMethodUseCaseSuite))
}

func (s *CreatePaymentMethodUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.paymentMethodRepository = repositoryMock.NewPaymentMethodRepository(s.T())
}

func (s *CreatePaymentMethodUseCaseSuite) TestExecute() {
	type args struct {
		input *dtos.PaymentMethodInput
	}

	type dependencies struct {
		paymentMethodRepository *repositoryMock.PaymentMethodRepository
	}

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       func(output *dtos.PaymentMethodOutput, err error)
	}{
		{
			name: "deve criar método de pagamento com sucesso",
			args: args{
				input: &dtos.PaymentMethodInput{
					Name:        "PIX",
					Code:        "PIX",
					Description: "Pagamento instantâneo",
				},
			},
			dependencies: dependencies{
				paymentMethodRepository: func() *repositoryMock.PaymentMethodRepository {
					s.paymentMethodRepository.
						EXPECT().
						Save(s.ctx, mock.AnythingOfType("*entities.PaymentMethod")).
						Return(nil).
						Once()
					return s.paymentMethodRepository
				}(),
			},
			expect: func(output *dtos.PaymentMethodOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("PIX", output.Name)
				s.Equal("PIX", output.Code)
				s.NotEmpty(output.ID)
			},
		},
		{
			name: "deve criar método de pagamento com code normalizado",
			args: args{
				input: &dtos.PaymentMethodInput{
					Name:        "Cartão de Crédito",
					Code:        "credit_card",
					Description: "Pagamento com cartão",
				},
			},
			dependencies: dependencies{
				paymentMethodRepository: func() *repositoryMock.PaymentMethodRepository {
					s.paymentMethodRepository.
						EXPECT().
						Save(s.ctx, mock.AnythingOfType("*entities.PaymentMethod")).
						Return(nil).
						Once()
					return s.paymentMethodRepository
				}(),
			},
			expect: func(output *dtos.PaymentMethodOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("Cartão de Crédito", output.Name)
				s.Equal("CREDIT_CARD", output.Code)
			},
		},
		{
			name: "deve retornar erro ao criar com nome vazio",
			args: args{
				input: &dtos.PaymentMethodInput{
					Name: "",
					Code: "PIX",
				},
			},
			dependencies: dependencies{
				paymentMethodRepository: s.paymentMethodRepository,
			},
			expect: func(output *dtos.PaymentMethodOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid payment method name")
			},
		},
		{
			name: "deve retornar erro ao criar com code vazio",
			args: args{
				input: &dtos.PaymentMethodInput{
					Name: "PIX",
					Code: "",
				},
			},
			dependencies: dependencies{
				paymentMethodRepository: s.paymentMethodRepository,
			},
			expect: func(output *dtos.PaymentMethodOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid payment method code")
			},
		},
		{
			name: "deve retornar erro ao falhar ao salvar no repositório",
			args: args{
				input: &dtos.PaymentMethodInput{
					Name: "PIX",
					Code: "PIX",
				},
			},
			dependencies: dependencies{
				paymentMethodRepository: func() *repositoryMock.PaymentMethodRepository {
					s.paymentMethodRepository.
						EXPECT().
						Save(s.ctx, mock.AnythingOfType("*entities.PaymentMethod")).
						Return(errors.New("database connection failed")).
						Once()
					return s.paymentMethodRepository
				}(),
			},
			expect: func(output *dtos.PaymentMethodOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "database connection failed")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Act
			uc := NewCreatePaymentMethodUseCase(s.obs, scenario.dependencies.paymentMethodRepository, metrics.NewFinancialMetrics(s.obs))
			output, err := uc.Execute(s.ctx, scenario.args.input)

			// Assert
			scenario.expect(output, err)
		})
	}
}
