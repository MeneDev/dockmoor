package dockreftst

import (
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/stretchr/testify/mock"
)

var _ dockref.Resolver = (*MockResolver)(nil)

type MockResolver struct {
	mock.Mock
}

func (m *MockResolver) FindAllTags(reference dockref.Reference) ([]dockref.Reference, error) {
	called := m.Called(reference)
	i := called.Get(0)
	refs := i.([]dockref.Reference)
	e := called.Error(1)
	return refs, e
}

func (m *MockResolver) OnFindAllTags(reference interface{}) *mock.Call {
	return m.On("FindAllTags", reference)
}

func MockResolverNew() *MockResolver {
	return &MockResolver{}
}
