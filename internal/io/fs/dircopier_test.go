package fs

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// mockFilesFinder implémente l'interface FilesFinder avec le pattern de mock fonctionnel
type mockFilesFinder struct {
	FindFilesFunc func(dir string) ([]string, error)
}

func (m *mockFilesFinder) FindFiles(dir string) ([]string, error) {
	if m.FindFilesFunc != nil {
		return m.FindFilesFunc(dir)
	}
	return nil, nil
}

func TestDirCopier_CopyDir_ShouldCopyFilesSuccessfully(t *testing.T) {
	// given
	from := uuid.New().String()
	to := uuid.New().String()
	fileName := uuid.New().String() + ".txt"
	fileContent := []byte(uuid.New().String())

	// setup
	appFS := afero.NewMemMapFs()
	filePath := filepath.Join(from, fileName)
	_ = afero.WriteFile(appFS, filePath, fileContent, 0644)

	finder := &mockFilesFinder{
		FindFilesFunc: func(dir string) ([]string, error) {
			return []string{fileName}, nil
		},
	}
	copier := NewDirCopier(appFS)

	// test
	err := copier.CopyDir(finder, from, to)

	// expect
	assert.NoError(t, err)
	copiedContent, _ := afero.ReadFile(appFS, filepath.Join(to, fileName))
	assert.Equal(t, fileContent, copiedContent)
}

func TestDirCopier_CopyDir_ShouldReturnErrorOnFindFilesFailure(t *testing.T) {
	// given
	from := uuid.New().String()
	to := uuid.New().String()
	expectedErr := errors.New("mock finder error")

	// setup
	appFS := afero.NewMemMapFs()
	finder := &mockFilesFinder{
		FindFilesFunc: func(dir string) ([]string, error) {
			return nil, expectedErr
		},
	}
	copier := NewDirCopier(appFS)

	// test
	err := copier.CopyDir(finder, from, to)

	// expect
	assert.ErrorContains(t, err, "FindFiles err")
}

func TestDirCopier_CopyDir_ShouldReturnErrorOnReadFileFailure(t *testing.T) {
	// given
	from := uuid.New().String()
	to := uuid.New().String()
	nonExistentFileName := uuid.New().String() + ".txt"

	// setup
	appFS := afero.NewMemMapFs()
	finder := &mockFilesFinder{
		FindFilesFunc: func(dir string) ([]string, error) {
			// On retourne un fichier qui n'existe pas réellement dans notre afero.Fs
			return []string{nonExistentFileName}, nil
		},
	}
	copier := NewDirCopier(appFS)

	// test
	err := copier.CopyDir(finder, from, to)

	// expect
	assert.ErrorContains(t, err, "afero.ReadFile err")
}

func TestDirCopier_CopyDir_ShouldReturnErrorOnWriteFileFailure(t *testing.T) {
	// given
	from := uuid.New().String()
	to := uuid.New().String()
	fileName := uuid.New().String() + ".txt"
	fileContent := []byte(uuid.New().String())

	// setup
	baseFS := afero.NewMemMapFs()
	filePath := filepath.Join(from, fileName)
	_ = afero.WriteFile(baseFS, filePath, fileContent, 0644)

	// On wrap le file system en lecture seule pour forcer l'échec d'écriture
	readOnlyFS := afero.NewReadOnlyFs(baseFS)

	finder := &mockFilesFinder{
		FindFilesFunc: func(dir string) ([]string, error) {
			return []string{fileName}, nil
		},
	}
	copier := NewDirCopier(readOnlyFS)

	// test
	err := copier.CopyDir(finder, from, to)

	// expect
	assert.ErrorContains(t, err, "operation not permitted")
}
