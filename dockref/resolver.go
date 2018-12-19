package dockref

type Resolver interface {
	Resolve(reference Reference) ([]Reference, error)
}
