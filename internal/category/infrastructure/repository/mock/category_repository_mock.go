// Code generated by mockery v2.42.2. DO NOT EDIT.

package repositoryMock

import (
	context "context"

	entities "github.com/jailtonjunior94/financial/internal/category/domain/entities"

	mock "github.com/stretchr/testify/mock"
)

// CategoryRepository is an autogenerated mock type for the CategoryRepository type
type CategoryRepository struct {
	mock.Mock
}

// Find provides a mock function with given fields: ctx, userID
func (_m *CategoryRepository) Find(ctx context.Context, userID string) ([]*entities.Category, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for Find")
	}

	var r0 []*entities.Category
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*entities.Category, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []*entities.Category); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*entities.Category)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindByID provides a mock function with given fields: ctx, userID, id
func (_m *CategoryRepository) FindByID(ctx context.Context, userID string, id string) (*entities.Category, error) {
	ret := _m.Called(ctx, userID, id)

	if len(ret) == 0 {
		panic("no return value specified for FindByID")
	}

	var r0 *entities.Category
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*entities.Category, error)); ok {
		return rf(ctx, userID, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *entities.Category); ok {
		r0 = rf(ctx, userID, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entities.Category)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, userID, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Insert provides a mock function with given fields: ctx, category
func (_m *CategoryRepository) Insert(ctx context.Context, category *entities.Category) (*entities.Category, error) {
	ret := _m.Called(ctx, category)

	if len(ret) == 0 {
		panic("no return value specified for Insert")
	}

	var r0 *entities.Category
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *entities.Category) (*entities.Category, error)); ok {
		return rf(ctx, category)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *entities.Category) *entities.Category); ok {
		r0 = rf(ctx, category)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entities.Category)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *entities.Category) error); ok {
		r1 = rf(ctx, category)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, category
func (_m *CategoryRepository) Update(ctx context.Context, category *entities.Category) (*entities.Category, error) {
	ret := _m.Called(ctx, category)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 *entities.Category
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *entities.Category) (*entities.Category, error)); ok {
		return rf(ctx, category)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *entities.Category) *entities.Category); ok {
		r0 = rf(ctx, category)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entities.Category)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *entities.Category) error); ok {
		r1 = rf(ctx, category)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewCategoryRepository creates a new instance of CategoryRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCategoryRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *CategoryRepository {
	mock := &CategoryRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
