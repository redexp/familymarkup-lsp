package state

import (
	"path/filepath"
	"slices"
	s "strings"

	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
)

type File struct {
	Uri  Uri
	Path []string
}

type FilesTree map[string]*FileTree

type FileTree struct {
	Name     string
	File     *File
	Family   *Family
	Member   *Member
	Children FilesTree
}

func CreateFile(uri Uri, folder Uri) (file *File, err error) {
	path, err := UriToPath(uri)

	if err != nil {
		return
	}

	file = &File{
		Uri: uri,
	}

	path = s.TrimPrefix(path, folder)
	path = s.TrimLeft(path, "/")

	file.Path = s.Split(path, "/")

	count := len(file.Path)

	if count > 0 {
		last := file.Path[count-1]
		file.Path[count-1] = s.TrimSuffix(last, filepath.Ext(last))
	}

	return
}

func (file *File) PathIncludes(parts ...string) bool {
	index := -1

	for _, part := range parts {
		i := slices.Index(file.Path, part)

		if i == -1 || i < index {
			return false
		}
	}

	return true
}
