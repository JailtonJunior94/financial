package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/messaging"
)

// mockSyncUseCase implements usecase.SyncBudgetSpentAmountUseCase for tests.
type mockSyncUseCase struct {
	mock.Mock
}

func (m *mockSyncUseCase) Execute(ctx context.Context, userID vos.UUID, referenceMonth pkgVos.ReferenceMonth, categoryID vos.UUID) error {
	args := m.Called(ctx, userID, referenceMonth, categoryID)
	return args.Error(0)
}

// mockProcessedEventsRepo implements outbox.ProcessedEventsRepository for tests.
type mockProcessedEventsRepo struct {
	mock.Mock
}

func (m *mockProcessedEventsRepo) IsProcessed(ctx context.Context, eventID uuid.UUID, consumerName string) (bool, error) {
	args := m.Called(ctx, eventID, consumerName)
	return args.Bool(0), args.Error(1)
}

func (m *mockProcessedEventsRepo) MarkAsProcessed(ctx context.Context, eventID uuid.UUID, consumerName string) error {
	args := m.Called(ctx, eventID, consumerName)
	return args.Error(0)
}

func (m *mockProcessedEventsRepo) TryClaimEvent(ctx context.Context, eventID uuid.UUID, consumerName string) (bool, error) {
	args := m.Called(ctx, eventID, consumerName)
	return args.Bool(0), args.Error(1)
}

func (m *mockProcessedEventsRepo) DeleteClaim(ctx context.Context, eventID uuid.UUID, consumerName string) error {
	args := m.Called(ctx, eventID, consumerName)
	return args.Error(0)
}

func (m *mockProcessedEventsRepo) DeleteOldProcessed(ctx context.Context, olderThan time.Duration) (int64, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

type BudgetEventConsumerSuite struct {
	suite.Suite
	ctx                 context.Context
	obs                 *fake.Provider
	syncUseCase         *mockSyncUseCase
	processedEventsRepo *mockProcessedEventsRepo
	consumer            *BudgetEventConsumer
}

func TestBudgetEventConsumerSuite(t *testing.T) {
	suite.Run(t, new(BudgetEventConsumerSuite))
}

func (s *BudgetEventConsumerSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.syncUseCase = new(mockSyncUseCase)
	s.processedEventsRepo = new(mockProcessedEventsRepo)
	s.consumer = NewBudgetEventConsumer(s.syncUseCase, s.processedEventsRepo, s.obs)
}

func (s *BudgetEventConsumerSuite) TestTopics_ShouldReturnTransactionCreated() {
	topics := s.consumer.Topics()
	s.Require().Len(topics, 1)
	s.Equal("transaction.created", topics[0])
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

	s.processedEventsRepo.On("TryClaimEvent", mock.Anything, eventID, "budget_event_consumer").Return(true, nil).Once()

	expectedUserID, _ := vos.NewUUIDFromString(userID.String())
	expectedCategoryID, _ := vos.NewUUIDFromString(categoryID.String())
	expectedMonth, _ := pkgVos.NewReferenceMonth("2026-03")
	s.syncUseCase.On("Execute", mock.Anything, expectedUserID, expectedMonth, expectedCategoryID).Return(nil).Once()

	err := s.consumer.Handle(s.ctx, msg)

	s.NoError(err)
	s.syncUseCase.AssertExpectations(s.T())
	s.processedEventsRepo.AssertExpectations(s.T())
}

func (s *BudgetEventConsumerSuite) TestHandle_InvalidJSON_ShouldReturnError() {
	eventID := uuid.New()

	msg := &messaging.Message{
		ID:      eventID.String(),
		Topic:   "transaction.created",
		Payload: []byte(`{invalid json`),
	}

	s.processedEventsRepo.On("TryClaimEvent", mock.Anything, eventID, "budget_event_consumer").Return(true, nil).Once()
	s.processedEventsRepo.On("DeleteClaim", mock.Anything, eventID, "budget_event_consumer").Return(nil).Maybe()

	err := s.consumer.Handle(s.ctx, msg)

	s.Error(err)
	s.syncUseCase.AssertNotCalled(s.T(), "Execute")
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

	// Second call: TryClaimEvent returns false (already processed)
	s.processedEventsRepo.On("TryClaimEvent", mock.Anything, eventID, "budget_event_consumer").Return(false, nil).Once()

	err := s.consumer.Handle(s.ctx, msg)

	s.NoError(err)
	s.syncUseCase.AssertNotCalled(s.T(), "Execute")
	s.processedEventsRepo.AssertExpectations(s.T())
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

	s.processedEventsRepo.On("TryClaimEvent", mock.Anything, eventID, "budget_event_consumer").Return(true, nil).Once()

	expectedUserID, _ := vos.NewUUIDFromString(userID.String())
	expectedCategoryID, _ := vos.NewUUIDFromString(categoryID.String())
	expectedMonth, _ := pkgVos.NewReferenceMonth("2026-03")
	s.syncUseCase.On("Execute", mock.Anything, expectedUserID, expectedMonth, expectedCategoryID).
		Return(errSyncFailed).Once()

	s.processedEventsRepo.On("DeleteClaim", mock.Anything, eventID, "budget_event_consumer").Return(nil).Once()

	err := s.consumer.Handle(s.ctx, msg)

	s.Error(err)
	s.syncUseCase.AssertExpectations(s.T())
	s.processedEventsRepo.AssertExpectations(s.T())
}

var errSyncFailed = fmt.Errorf("sync failed")
