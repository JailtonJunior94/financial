package domain

import (
	pkginterfaces "github.com/jailtonjunior94/financial/pkg/domain/interfaces"
)

var (
	// ErrCategoryNotFound delegates to the shared cross-module sentinel so that
	// errors.Is works correctly across the budget→category boundary.
	ErrCategoryNotFound = pkginterfaces.ErrCategoryNotFound
)
