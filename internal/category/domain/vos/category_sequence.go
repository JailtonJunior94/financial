package vos

type CategorySequence struct {
	Sequence *uint
	Valid    bool
}

func NewCategorySequence(i uint) CategorySequence {
	if i == 0 {
		return CategorySequence{Sequence: nil, Valid: false}
	}
	return CategorySequence{Sequence: &i, Valid: true}
}

func (n CategorySequence) Value() uint {
	if n.Sequence != nil {
		return *n.Sequence
	}

	if n.Valid {
		return *n.Sequence
	}
	return 0
}
