package dockreftst

import (
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/stretchr/testify/mock"
)

var _ dockref.Resolver = (*MockResolver)(nil)

type MockResolver struct {
	mock.Mock
}

func (m *MockResolver) Resolve(reference dockref.Reference) ([]dockref.Reference, error) {
	called := m.Called(reference)
	i := called.Get(0)
	refs := i.([]dockref.Reference)
	e := called.Error(1)
	return refs, e
}

func (m *MockResolver) OnResolve(reference dockref.Reference) *mock.Call {
	return m.On("Resolve", reference)
}

func MockResolverNew() *MockResolver {
	return &MockResolver{}
}
