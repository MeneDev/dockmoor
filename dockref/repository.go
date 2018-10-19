package dockref

type Repository interface {
	Resolve(reference Reference) (Reference, error)
}
