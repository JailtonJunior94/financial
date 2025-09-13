package repositories_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/category/domain/vos"
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/database"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type CategoryRepositorySuite struct {
	suite.Suite

	db                   *sql.DB
	ctx                  context.Context
	o11y                 o11y.Observability
	repository           interfaces.CategoryRepository
	cockroachDBContainer *database.CockroachDBContainer
}

func TestCategoryRepositorySuite(t *testing.T) {
	suite.Run(t, new(CategoryRepositorySuite))
}

func (s *CategoryRepositorySuite) SetupSuite() {
	s.ctx = context.Background()
	s.cockroachDBContainer = database.SetupCockroachDB(s.ctx, s.T())
	s.o11y = o11y.NewDevelopmentObservability("test", "1.0.0")
	s.db, _ = sql.Open("postgres", s.cockroachDBContainer.DSN(s.ctx, s.T()))
	s.cockroachDBContainer.Migrate(s.T(), s.db, "file://../../../../database/migrations", "financial")

	query := `insert into users values ('123e4567-e89b-12d3-a456-426614174000', 'Test User', 'test@example.com', 'password', NOW());`
	_, err := s.db.ExecContext(s.ctx, query)
	s.Require().NoError(err)

	s.repository = repositories.NewCategoryRepository(s.db, s.o11y)
}

func (s *CategoryRepositorySuite) TearDownSuite() {
	s.cockroachDBContainer.Teardown(s.ctx, s.T())
	if err := s.db.Close(); err != nil {
		s.T().Fatalf("failed to close database connection: %v", err)
	}
}

func (s *CategoryRepositorySuite) TestSave() {
	type args struct {
		entity *entities.Category
	}

	user, err := sharedVos.NewUUIDFromString("123e4567-e89b-12d3-a456-426614174000")
	s.Require().NoError(err)

	categoryName := vos.NewCategoryName("Test Category")
	sequence := vos.NewCategorySequence(1)

	category, err := entities.NewCategory(user, nil, categoryName, sequence)
	s.Require().NoError(err)

	scenarios := []struct {
		name     string
		args     args
		expected func(err error)
	}{
		{
			name: "should create a category successfully",
			args: args{entity: category},
			expected: func(err error) {
				s.Require().NoError(err)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			err := s.repository.Save(s.ctx, scenario.args.entity)
			scenario.expected(err)
		})
	}
}
