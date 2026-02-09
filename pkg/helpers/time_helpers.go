package helpers

import (
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

// ParseNullableTime converte *time.Time para vos.NullableTime
func ParseNullableTime(t *time.Time) vos.NullableTime {
	if t == nil {
		return vos.NullableTime{}
	}
	return vos.NewNullableTime(*t)
}
