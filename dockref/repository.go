package dockref

type Repository interface {
	Resolve(reference Reference) (Reference, error)
}

type dockerDaemonRegistry struct {
}

var _ Repository = (*dockerDaemonRegistry)(nil)

func (dockerDaemonRegistry) Resolve(reference Reference) (Reference, error) {
	panic("implement me")
}
