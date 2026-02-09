package vos

import (
	sharedVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

// ReferenceMonth é um alias para a versão compartilhada.
// DEPRECATED: Use pkg/domain/vos.ReferenceMonth diretamente.
type ReferenceMonth = sharedVos.ReferenceMonth

// NewReferenceMonth é um alias para o construtor compartilhado.
// DEPRECATED: Use pkg/domain/vos.NewReferenceMonth diretamente.
var NewReferenceMonth = sharedVos.NewReferenceMonth

// NewReferenceMonthFromDate é um alias para o construtor compartilhado.
// DEPRECATED: Use pkg/domain/vos.NewReferenceMonthFromDate diretamente.
var NewReferenceMonthFromDate = sharedVos.NewReferenceMonthFromDate

// ErrInvalidReferenceMonth é um alias para o erro compartilhado.
// DEPRECATED: Use pkg/domain/vos.ErrInvalidReferenceMonth diretamente.
var ErrInvalidReferenceMonth = sharedVos.ErrInvalidReferenceMonth
