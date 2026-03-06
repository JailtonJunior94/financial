package entities

import (
"time"

sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
"github.com/jailtonjunior94/financial/internal/category/domain/vos"
)

type Subcategory struct {
ID         sharedVos.UUID
CategoryID sharedVos.UUID
UserID     sharedVos.UUID
Name       vos.CategoryName
Sequence   vos.CategorySequence
CreatedAt  sharedVos.NullableTime
UpdatedAt  sharedVos.NullableTime
DeletedAt  sharedVos.NullableTime
}

func NewSubcategory(userID, categoryID sharedVos.UUID, name vos.CategoryName, sequence vos.CategorySequence) (*Subcategory, error) {
return &Subcategory{
CategoryID: categoryID,
UserID:     userID,
Name:       name,
Sequence:   sequence,
CreatedAt:  sharedVos.NewNullableTime(time.Now()),
}, nil
}

func (s *Subcategory) Update(name string, sequence uint) error {
categoryName, err := vos.NewCategoryName(name)
if err != nil {
return err
}

categorySequence, err := vos.NewCategorySequence(sequence)
if err != nil {
return err
}

s.Name = categoryName
s.Sequence = categorySequence
s.UpdatedAt = sharedVos.NewNullableTime(time.Now())

return nil
}

func (s *Subcategory) Delete() {
s.DeletedAt = sharedVos.NewNullableTime(time.Now())
}
