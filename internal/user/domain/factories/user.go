package factories

import (
	"time"

	sharedVos "github.com/jailtonjunior94/financial/internal/shared/vos"

	"github.com/jailtonjunior94/financial/internal/user/domain/entities"
	"github.com/jailtonjunior94/financial/internal/user/domain/vos"
)

func CreateUser(name string, email string) (*entities.User, error) {
	id, err := sharedVos.NewUUID()
	if err != nil {
		return nil, err
	}

	nameValid, err := vos.NewUserName(name)
	if err != nil {
		return nil, err
	}

	emailValid, err := vos.NewEmail(email)
	if err != nil {
		return nil, err
	}

	user, err := entities.NewUser(nameValid, emailValid)
	if err != nil {
		return nil, err
	}
	user.ID = id
	user.CreatedAt = time.Now()

	return user, nil
}
