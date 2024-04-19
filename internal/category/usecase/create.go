package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/logger"
)

type (
	CreateCategoryUseCase interface {
		Execute(userID string, input *CreateCategoryInput) (*CreateCategoryOutput, error)
	}

	createCategoryUseCase struct {
		logger     logger.Logger
		repository interfaces.CategoryRepository
	}
)

func NewCreateCategoryUseCase(logger logger.Logger, repository interfaces.CategoryRepository) CreateCategoryUseCase {
	return &createCategoryUseCase{logger: logger, repository: repository}
}

func (u *createCategoryUseCase) Execute(userID string, input *CreateCategoryInput) (*CreateCategoryOutput, error) {
	newCategory, err := entities.NewCategory(userID, input.Name, input.Sequence)
	if err != nil {
		u.logger.Warn("error parsing category", logger.Field{Key: "warning", Value: err.Error()})
		return nil, err
	}

	category, err := u.repository.Insert(context.Background(), newCategory)
	if err != nil {
		u.logger.Error("error creating category",
			logger.Field{Key: "user_id", Value: category.UserID},
			logger.Field{Key: "error", Value: err.Error()},
		)
		return nil, err
	}

	return &CreateCategoryOutput{
		ID:        category.ID,
		Name:      category.Name,
		Sequence:  category.Sequence,
		CreatedAt: category.CreatedAt,
	}, nil
}
