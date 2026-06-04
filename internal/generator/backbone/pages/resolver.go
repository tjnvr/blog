package pages

type Resolver interface {
	ResolveAll() ([]string, error)
}
