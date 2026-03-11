package adapters_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financial/internal/category/infrastructure/adapters"
	pkginterfaces "github.com/jailtonjunior94/financial/pkg/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type CategoryProviderAdapterSuite struct {
	suite.Suite
	ctx context.Context
	obs *fake.Provider
	fm  *metrics.FinancialMetrics
}

func TestCategoryProviderAdapterSuite(t *testing.T) {
	suite.Run(t, new(CategoryProviderAdapterSuite))
}

func (s *CategoryProviderAdapterSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.fm = metrics.NewFinancialMetrics(s.obs)
}

func (s *CategoryProviderAdapterSuite) TestValidateCategories_EmptyList() {
	db, _, err := sqlmock.New()
	s.Require().NoError(err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			s.T().Logf("TestValidateCategories_EmptyList: failed to close db: %v", closeErr)
		}
	}()

	adapter := adapters.NewCategoryProviderAdapter(db, s.obs, s.fm)
	err = adapter.ValidateCategories(s.ctx, "user-1", []string{})
	s.NoError(err)
}

func (s *CategoryProviderAdapterSuite) TestValidateCategories_AllValid() {
	db, mock, err := sqlmock.New()
	s.Require().NoError(err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			s.T().Logf("TestValidateCategories_AllValid: failed to close db: %v", closeErr)
		}
	}()

	userID := "user-uuid-1"
	categoryID := "cat-uuid-1"

	rows := sqlmock.NewRows([]string{"id"}).AddRow(categoryID)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM categories WHERE id IN")).
		WillReturnRows(rows)

	adapter := adapters.NewCategoryProviderAdapter(db, s.obs, s.fm)
	err = adapter.ValidateCategories(s.ctx, userID, []string{categoryID})
	s.NoError(err)
	s.NoError(mock.ExpectationsWereMet())
}

func (s *CategoryProviderAdapterSuite) TestValidateCategories_CategoryNotFound() {
	db, mock, err := sqlmock.New()
	s.Require().NoError(err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			s.T().Logf("TestValidateCategories_CategoryNotFound: failed to close db: %v", closeErr)
		}
	}()

	userID := "user-uuid-1"
	categoryID := "cat-uuid-missing"

	rows := sqlmock.NewRows([]string{"id"})
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM categories WHERE id IN")).
		WillReturnRows(rows)

	adapter := adapters.NewCategoryProviderAdapter(db, s.obs, s.fm)
	err = adapter.ValidateCategories(s.ctx, userID, []string{categoryID})
	s.Error(err)
	s.True(errors.Is(err, pkginterfaces.ErrCategoryNotFound))
	s.NoError(mock.ExpectationsWereMet())
}

func (s *CategoryProviderAdapterSuite) TestValidateCategories_DBError() {
	db, mock, err := sqlmock.New()
	s.Require().NoError(err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			s.T().Logf("TestValidateCategories_DBError: failed to close db: %v", closeErr)
		}
	}()

	userID := "user-uuid-1"
	categoryID := "cat-uuid-1"
	dbErr := errors.New("connection refused")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM categories WHERE id IN")).
		WillReturnError(dbErr)

	adapter := adapters.NewCategoryProviderAdapter(db, s.obs, s.fm)
	err = adapter.ValidateCategories(s.ctx, userID, []string{categoryID})
	s.Error(err)
	s.False(errors.Is(err, pkginterfaces.ErrCategoryNotFound))
	s.NoError(mock.ExpectationsWereMet())
}

func (s *CategoryProviderAdapterSuite) TestValidateCategories_PartialFound() {
	db, mock, err := sqlmock.New()
	s.Require().NoError(err)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			s.T().Logf("TestValidateCategories_PartialFound: failed to close db: %v", closeErr)
		}
	}()

	userID := "user-uuid-1"
	catID1 := "cat-uuid-1"
	catID2 := "cat-uuid-2"

	rows := sqlmock.NewRows([]string{"id"}).AddRow(catID1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM categories WHERE id IN")).
		WillReturnRows(rows)

	adapter := adapters.NewCategoryProviderAdapter(db, s.obs, s.fm)
	err = adapter.ValidateCategories(s.ctx, userID, []string{catID1, catID2})
	s.Error(err)
	s.True(errors.Is(err, pkginterfaces.ErrCategoryNotFound))
	s.NoError(mock.ExpectationsWereMet())
}
