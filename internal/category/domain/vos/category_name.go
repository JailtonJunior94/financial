package vos

type CategoryName struct {
	Value *string
	Valid bool
}

func NewCategoryName(name string) CategoryName {
	if len(name) == 0 || len(name) > 255 {
		return CategoryName{Value: nil, Valid: false}
	}
	return CategoryName{Value: &name, Valid: true}
}

func (v CategoryName) String() string {
	if v.Value != nil {
		return *v.Value
	}
	return ""
}
