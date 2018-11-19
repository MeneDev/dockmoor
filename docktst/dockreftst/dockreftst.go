package dockreftst

import (
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/stretchr/testify/mock"
)

var _ dockref.Repository = (*MockRepository)(nil)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Resolve(reference dockref.Reference) ([]dockref.Reference, error) {
	called := m.Called(reference)
	i := called.Get(0)
	ref := i.(dockref.Reference)
	e := called.Error(1)
	return []dockref.Reference{ref}, e
}

func (m *MockRepository) OnResolve(reference dockref.Reference) *mock.Call {
	return m.On("Resolve", reference)
}

func MockRepositoryNew() *MockRepository {
	return &MockRepository{}
}
