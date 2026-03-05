package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/domain/entities"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/user/domain/vos"
	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/encrypt"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	devkitVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	UpdateUserUseCase interface {
		Execute(ctx context.Context, id string, input *dtos.UpdateUserInput) (*dtos.UserOutput, error)
	}

	updateUserUseCase struct {
		o11y       observability.Observability
		fm         *metrics.FinancialMetrics
		hash       encrypt.HashAdapter
		repository interfaces.UserRepository
	}
)

func NewUpdateUserUseCase(
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	hash encrypt.HashAdapter,
	repository interfaces.UserRepository,
) UpdateUserUseCase {
	return &updateUserUseCase{
		o11y:       o11y,
		fm:         fm,
		hash:       hash,
		repository: repository,
	}
}

func (u *updateUserUseCase) Execute(ctx context.Context, id string, input *dtos.UpdateUserInput) (*dtos.UserOutput, error) {
	start := time.Now()
	ctx, span := u.o11y.Tracer().Start(ctx, "update_user_usecase.execute")
	defer span.End()
	span.AddEvent("updating user", observability.Field{Key: "user.id", Value: id})
	user, err := u.repository.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "update_user", "user", "infra", time.Since(start))
		return nil, err
	}
	if user == nil {
		u.fm.RecordUsecaseFailure(ctx, "update_user", "user", "not_found", time.Since(start))
		return nil, customerrors.ErrUserNotFound
	}
	if err := u.applyFieldUpdates(ctx, span, start, id, input, user); err != nil {
		return nil, err
	}
	now := time.Now()
	user.UpdatedAt = devkitVos.NewNullableTime(now)
	updated, err := u.repository.Update(ctx, user)
	if err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "update_user", "user", "infra", time.Since(start))
		return nil, err
	}
	u.fm.RecordUsecaseOperation(ctx, "update_user", "user", time.Since(start))
	return &dtos.UserOutput{
		ID:        updated.ID.String(),
		Name:      updated.Name.String(),
		Email:     updated.Email.String(),
		CreatedAt: updated.CreatedAt,
		UpdatedAt: updated.UpdatedAt.Ptr(),
	}, nil
}

func (u *updateUserUseCase) applyFieldUpdates(ctx context.Context, span observability.Span, start time.Time, id string, input *dtos.UpdateUserInput, user *entities.User) error {
	if input.Name != nil {
		name, err := vos.NewUserName(*input.Name)
		if err != nil {
			span.RecordError(err)
			u.fm.RecordUsecaseFailure(ctx, "update_user", "user", "validation", time.Since(start))
			return err
		}
		user.UpdateName(name)
	}
	if input.Email != nil {
		if err := u.updateEmail(ctx, span, start, id, *input.Email, user); err != nil {
			return err
		}
	}
	if input.Password != nil {
		if err := u.updatePassword(ctx, span, start, *input.Password, user); err != nil {
			return err
		}
	}
	return nil
}

func (u *updateUserUseCase) updateEmail(ctx context.Context, span observability.Span, start time.Time, id, rawEmail string, user *entities.User) error {
	email, err := vos.NewEmail(rawEmail)
	if err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "update_user", "user", "validation", time.Since(start))
		return err
	}
	existing, err := u.repository.FindByEmail(ctx, rawEmail)
	if err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "update_user", "user", "infra", time.Since(start))
		return err
	}
	if existing != nil && existing.ID.String() != id {
		u.fm.RecordUsecaseFailure(ctx, "update_user", "user", "conflict", time.Since(start))
		return customerrors.ErrEmailAlreadyExists
	}
	user.UpdateEmail(email)
	return nil
}

func (u *updateUserUseCase) updatePassword(ctx context.Context, span observability.Span, start time.Time, rawPassword string, user *entities.User) error {
	hash, err := u.hash.GenerateHash(rawPassword)
	if err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "update_user", "user", "infra", time.Since(start))
		return fmt.Errorf("generating password hash: %w", err)
	}
	if err := user.SetPassword(hash); err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "update_user", "user", "validation", time.Since(start))
		return err
	}
	return nil
}
