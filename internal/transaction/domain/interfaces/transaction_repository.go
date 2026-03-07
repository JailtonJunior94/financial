package interfaces

import (
	"context"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
)

// ListParams represents the parameters for paginated transaction listing.
type ListParams struct {
	UserID        vos.UUID
	PaymentMethod string
	CategoryID    string
	StartDate     *time.Time
	EndDate       *time.Time
	Limit         int
	Cursor        string
}

// TransactionRepository defines the persistence contract for transactions.
type TransactionRepository interface {
	Save(ctx context.Context, tx database.DBTX, t *entities.Transaction) error
	SaveAll(ctx context.Context, tx database.DBTX, ts []*entities.Transaction) error
	FindByID(ctx context.Context, id vos.UUID) (*entities.Transaction, error)
	FindByInstallmentGroup(ctx context.Context, groupID vos.UUID) ([]*entities.Transaction, error)
	Update(ctx context.Context, tx database.DBTX, t *entities.Transaction) error
	UpdateAll(ctx context.Context, tx database.DBTX, ts []*entities.Transaction) error
	ListPaginated(ctx context.Context, params ListParams) ([]*entities.Transaction, string, error)
}
