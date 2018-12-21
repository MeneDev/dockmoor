package dockref

type Resolver interface {
	FindAllTags(reference Reference) ([]Reference, error)
}
