package usecase_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/application/usecase"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/database"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type CreateSuite struct {
	suite.Suite

	db                   *sql.DB
	ctx                  context.Context
	o11y                 o11y.Observability
	orderRepository      interfaces.CategoryRepository
	cockroachDBContainer *database.CockroachDBContainer
	usecase              usecase.CreateCategoryUseCase
}

func TestCreateSuite(t *testing.T) {
	suite.Run(t, new(CreateSuite))
}

func (s *CreateSuite) SetupSuite() {
	s.ctx = context.Background()
	s.cockroachDBContainer = database.SetupCockroachDB(s.ctx, s.T())
	s.o11y = o11y.NewDevelopmentObservability("test", "1.0.0")
	s.db, _ = sql.Open("postgres", s.cockroachDBContainer.DSN(s.ctx, s.T()))
	s.cockroachDBContainer.Migrate(s.T(), s.db, "file://../../../../database/migrations", "financial")

	query := `
	insert into users values ('123e4567-e89b-12d3-a456-426614174000', 'Test User', 'test@example.com', 'password', NOW());

	insert into categories (id, user_id, parent_id, name, sequence, created_at) values
	('223e4567-e89b-12d3-a456-426614174000', '123e4567-e89b-12d3-a456-426614174000', NULL, 'Custos Fixos', 1, NOW());
	`

	_, err := s.db.ExecContext(s.ctx, query)
	s.Require().NoError(err)

	s.orderRepository = repositories.NewCategoryRepository(s.db, s.o11y)
	s.usecase = usecase.NewCreateCategoryUseCase(s.o11y, s.orderRepository)
}

func (s *CreateSuite) TearDownSuite() {
	s.cockroachDBContainer.Teardown(s.ctx, s.T())
	if err := s.db.Close(); err != nil {
		s.T().Fatalf("failed to close database connection: %v", err)
	}
}

func (s *CreateSuite) TestExecute() {
	type args struct {
		userID string
		input  *dtos.CategoryInput
	}

	userID := "123e4567-e89b-12d3-a456-426614174000"
	newCategory := &dtos.CategoryInput{Name: "Test Category", Sequence: 1}
	newSubcategory := &dtos.CategoryInput{ParentID: "223e4567-e89b-12d3-a456-426614174000", Name: "Subcategory", Sequence: 2}

	scenarios := []struct {
		name     string
		args     args
		expected func(output *dtos.CategoryOutput, err error)
	}{
		{
			name: "should create a category successfully",
			args: args{userID: userID, input: newCategory},
			expected: func(output *dtos.CategoryOutput, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(output)
				s.Require().Equal("Test Category", output.Name)
				s.Require().Equal(uint(1), output.Sequence)
				s.Require().NotEmpty(output.ID)
				s.Require().NotEmpty(output.CreatedAt)
			},
		},
		{
			name: "should create a subcategory successfully",
			args: args{userID: userID, input: newSubcategory},
			expected: func(output *dtos.CategoryOutput, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(output)
				s.Require().Equal("Subcategory", output.Name)
				s.Require().Equal(uint(2), output.Sequence)
				s.Require().NotEmpty(output.ID)
				s.Require().NotEmpty(output.CreatedAt)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			output, err := s.usecase.Execute(s.ctx, scenario.args.userID, scenario.args.input)
			scenario.expected(output, err)
		})
	}
}
