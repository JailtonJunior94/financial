package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	repositoryMock "github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories/mocks"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/messaging"
)

type BudgetEventConsumerSuite struct {
	suite.Suite
	ctx                 context.Context
	obs                 *fake.Provider
	syncUseCase         *repositoryMock.SyncBudgetSpentAmountUseCase
	processedEventsRepo *repositoryMock.ProcessedEventsRepository
	consumer            *BudgetEventConsumer
}

func TestBudgetEventConsumerSuite(t *testing.T) {
	suite.Run(t, new(BudgetEventConsumerSuite))
}

func (s *BudgetEventConsumerSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.syncUseCase = repositoryMock.NewSyncBudgetSpentAmountUseCase(s.T())
	s.processedEventsRepo = repositoryMock.NewProcessedEventsRepository(s.T())
	s.consumer = NewBudgetEventConsumer(s.syncUseCase, s.processedEventsRepo, s.obs)
}

func (s *BudgetEventConsumerSuite) TestTopics_ShouldReturnBothTransactionTopics() {
	topics := s.consumer.Topics()
	s.Require().Len(topics, 2)
	s.Contains(topics, "transaction.created")
	s.Contains(topics, "transaction.reversed")
}

func (s *BudgetEventConsumerSuite) TestHandle_ValidPayload_ShouldSyncBudget() {
	eventID := uuid.New()
	userID := uuid.New()
	categoryID := uuid.New()

	payload := transactionCreatedPayload{
		TransactionID:  uuid.New().String(),
		UserID:         userID.String(),
		CategoryID:     categoryID.String(),
		ReferenceMonth: "2026-03",
	}
	body, _ := json.Marshal(payload)

	msg := &messaging.Message{
		ID:      eventID.String(),
		Topic:   "transaction.created",
		Payload: body,
	}

	expectedUserID, _ := vos.NewUUIDFromString(userID.String())
	expectedCategoryID, _ := vos.NewUUIDFromString(categoryID.String())
	expectedMonth, _ := pkgVos.NewReferenceMonth("2026-03")

	s.processedEventsRepo.EXPECT().
		TryClaimEvent(mock.Anything, eventID, "budget_event_consumer").
		Return(true, nil).
		Once()
	s.syncUseCase.EXPECT().
		Execute(mock.Anything, expectedUserID, expectedMonth, expectedCategoryID).
		Return(nil).
		Once()

	err := s.consumer.Handle(s.ctx, msg)

	s.NoError(err)
}

func (s *BudgetEventConsumerSuite) TestHandle_InvalidJSON_ShouldReturnError() {
	eventID := uuid.New()

	msg := &messaging.Message{
		ID:      eventID.String(),
		Topic:   "transaction.created",
		Payload: []byte(`{invalid json`),
	}

	s.processedEventsRepo.EXPECT().
		TryClaimEvent(mock.Anything, eventID, "budget_event_consumer").
		Return(true, nil).
		Once()

	err := s.consumer.Handle(s.ctx, msg)

	s.Error(err)
}

func (s *BudgetEventConsumerSuite) TestHandle_AlreadyProcessed_ShouldSkipSilently() {
	eventID := uuid.New()
	payload := transactionCreatedPayload{
		TransactionID:  uuid.New().String(),
		UserID:         uuid.New().String(),
		CategoryID:     uuid.New().String(),
		ReferenceMonth: "2026-03",
	}
	body, _ := json.Marshal(payload)

	msg := &messaging.Message{
		ID:      eventID.String(),
		Topic:   "transaction.created",
		Payload: body,
	}

	s.processedEventsRepo.EXPECT().
		TryClaimEvent(mock.Anything, eventID, "budget_event_consumer").
		Return(false, nil).
		Once()

	err := s.consumer.Handle(s.ctx, msg)

	s.NoError(err)
}

func (s *BudgetEventConsumerSuite) TestHandle_SyncUseCaseError_ShouldDeleteClaimAndReturnError() {
	eventID := uuid.New()
	userID := uuid.New()
	categoryID := uuid.New()

	payload := transactionCreatedPayload{
		TransactionID:  uuid.New().String(),
		UserID:         userID.String(),
		CategoryID:     categoryID.String(),
		ReferenceMonth: "2026-03",
	}
	body, _ := json.Marshal(payload)

	msg := &messaging.Message{
		ID:      eventID.String(),
		Topic:   "transaction.created",
		Payload: body,
	}

	expectedUserID, _ := vos.NewUUIDFromString(userID.String())
	expectedCategoryID, _ := vos.NewUUIDFromString(categoryID.String())
	expectedMonth, _ := pkgVos.NewReferenceMonth("2026-03")

	s.processedEventsRepo.EXPECT().
		TryClaimEvent(mock.Anything, eventID, "budget_event_consumer").
		Return(true, nil).
		Once()
	s.syncUseCase.EXPECT().
		Execute(mock.Anything, expectedUserID, expectedMonth, expectedCategoryID).
		Return(errSyncFailed).
		Once()
	s.processedEventsRepo.EXPECT().
		DeleteClaim(mock.Anything, eventID, "budget_event_consumer").
		Return(nil).
		Once()

	err := s.consumer.Handle(s.ctx, msg)

	s.Error(err)
}

func (s *BudgetEventConsumerSuite) TestHandle_TransactionReversed_ShouldSyncBudget() {
	eventID := uuid.New()
	userID := uuid.New()
	categoryID := uuid.New()

	payload := transactionCreatedPayload{
		TransactionID:  uuid.New().String(),
		UserID:         userID.String(),
		CategoryID:     categoryID.String(),
		ReferenceMonth: "2026-03",
	}
	body, _ := json.Marshal(payload)

	msg := &messaging.Message{
		ID:      eventID.String(),
		Topic:   "transaction.reversed",
		Payload: body,
	}

	expectedUserID, _ := vos.NewUUIDFromString(userID.String())
	expectedCategoryID, _ := vos.NewUUIDFromString(categoryID.String())
	expectedMonth, _ := pkgVos.NewReferenceMonth("2026-03")

	s.processedEventsRepo.EXPECT().
		TryClaimEvent(mock.Anything, eventID, "budget_event_consumer").
		Return(true, nil).
		Once()
	s.syncUseCase.EXPECT().
		Execute(mock.Anything, expectedUserID, expectedMonth, expectedCategoryID).
		Return(nil).
		Once()

	err := s.consumer.Handle(s.ctx, msg)

	s.NoError(err)
}

func (s *BudgetEventConsumerSuite) TestHandle_TransactionReversedDuplicated_ShouldSkipSilently() {
	eventID := uuid.New()
	payload := transactionCreatedPayload{
		TransactionID:  uuid.New().String(),
		UserID:         uuid.New().String(),
		CategoryID:     uuid.New().String(),
		ReferenceMonth: "2026-03",
	}
	body, _ := json.Marshal(payload)

	msg := &messaging.Message{
		ID:      eventID.String(),
		Topic:   "transaction.reversed",
		Payload: body,
	}

	s.processedEventsRepo.EXPECT().
		TryClaimEvent(mock.Anything, eventID, "budget_event_consumer").
		Return(false, nil).
		Once()

	err := s.consumer.Handle(s.ctx, msg)

	s.NoError(err)
}

var errSyncFailed = fmt.Errorf("sync failed")
