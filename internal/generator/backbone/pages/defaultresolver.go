package pages

import "github.com/spf13/afero"

type defaultResolver struct {
	fs         afero.Fs
	contentDir string
}

func NewDefaultResolver(fs afero.Fs, contentDir string) *defaultResolver {
	return &defaultResolver{
		fs:         fs,
		contentDir: contentDir,
	}
}
func (d *defaultResolver) Resolve() ([]string, error) {
	return make([]string, 0), nil
}
