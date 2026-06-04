package htmlpath

type Resolver interface {
	Resolve(path string) (string, error)
}
