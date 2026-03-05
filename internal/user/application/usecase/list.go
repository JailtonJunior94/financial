package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	ListUsersInput struct {
		Limit  int
		Cursor string
	}

	ListUsersOutput struct {
		Users      []*dtos.UserOutput
		NextCursor *string
	}

	ListUsersUseCase interface {
		Execute(ctx context.Context, input ListUsersInput) (*ListUsersOutput, error)
	}

	listUsersUseCase struct {
		o11y       observability.Observability
		fm         *metrics.FinancialMetrics
		repository interfaces.UserRepository
	}
)

func NewListUsersUseCase(
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	repository interfaces.UserRepository,
) ListUsersUseCase {
	return &listUsersUseCase{
		o11y:       o11y,
		fm:         fm,
		repository: repository,
	}
}

func (u *listUsersUseCase) Execute(ctx context.Context, input ListUsersInput) (*ListUsersOutput, error) {
	start := time.Now()
	ctx, span := u.o11y.Tracer().Start(ctx, "list_users_usecase.execute")
	defer span.End()

	span.AddEvent("listing users",
		observability.Field{Key: "pagination.limit", Value: input.Limit},
		observability.Field{Key: "pagination.has_cursor", Value: input.Cursor != ""},
	)

	users, nextCursor, err := u.repository.FindAll(ctx, input.Limit, input.Cursor)
	if err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "list_users", "user", "infra", time.Since(start))
		return nil, err
	}

	output := make([]*dtos.UserOutput, len(users))
	for i, user := range users {
		output[i] = &dtos.UserOutput{
			ID:        user.ID.String(),
			Name:      user.Name.String(),
			Email:     user.Email.String(),
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt.Ptr(),
		}
	}

	u.fm.RecordUsecaseOperation(ctx, "list_users", "user", time.Since(start))

	return &ListUsersOutput{
		Users:      output,
		NextCursor: nextCursor,
	}, nil
}
