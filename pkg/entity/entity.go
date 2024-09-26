package entity

import (
	"time"

	"github.com/jailtonjunior94/financial/pkg/vos"
)

type Base struct {
	ID        vos.UUID
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}
