package auth

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, u *User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint64) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, u *User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id uint64, hashedPassword string) error {
	args := m.Called(ctx, id, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockIdentityRepository struct {
	mock.Mock
}

func (m *MockIdentityRepository) CreateIdentity(ctx context.Context, i *Identity) error {
	args := m.Called(ctx, i)
	return args.Error(0)
}

func (m *MockIdentityRepository) GetIdentityByID(ctx context.Context, id uint64) (*Identity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Identity), args.Error(1)
}

func (m *MockIdentityRepository) ListIdentitiesByUser(ctx context.Context, userID uint64) ([]Identity, error) {
	args := m.Called(ctx, userID)
	var ids []Identity
	if args.Get(0) != nil {
		ids = args.Get(0).([]Identity)
	}
	return ids, args.Error(1)
}

func (m *MockIdentityRepository) CountIdentitiesByUser(ctx context.Context, userID uint64) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockIdentityRepository) FindIdentityByProviderIdentifier(ctx context.Context, provider, identifier string) (*Identity, error) {
	args := m.Called(ctx, provider, identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Identity), args.Error(1)
}

func (m *MockIdentityRepository) DeleteIdentity(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockIdentityRepository) DeleteIdentitiesByUser(ctx context.Context, userID uint64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
