package relpath

type (
	Resolver interface {
		Resolve(oldPath, fromPath string) (string, error)
	}

	ResolverFactory interface {
		Create(fromDir, toDir string) Resolver
	}
)
