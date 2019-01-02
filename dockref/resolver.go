package dockref

type Resolver interface {
	FindAllTags(reference Reference) ([]Reference, error)
	Resolve(reference Reference) (Reference, error)
}

type ResolveMode int

const (
	ResolveModeUnchanged          = iota
	ResolveModeMostPreciseVersion = iota
)

type ResolverOptions struct {
	Mode ResolveMode
}
