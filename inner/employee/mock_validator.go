package employee

import "github.com/stretchr/testify/mock"

type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) Validate(request any) error {
	args := m.Called(request)
	return args.Error(0)
}

func (m *MockValidator) ExpectValidate(request any, err error) {
	m.On("Validate", request).Return(err)
}
