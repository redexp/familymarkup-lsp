package state

import (
	"io/fs"
	"math"
	"path/filepath"
	"slices"

	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
)

func addDuplicate(duplicates Duplicates, name string, dup *Duplicate) {
	_, exist := duplicates[name]

	if !exist {
		duplicates[name] = make([]*Duplicate, 0)
	}

	duplicates[name] = append(duplicates[name], dup)
}

func compareNames(a []rune, b []rune) uint {
	al := float64(len(a))
	bl := float64(len(b))
	max := uint(math.Max(al, bl))
	min := uint(math.Min(al, bl))
	diff := max - min

	if diff > 2 {
		return diff
	}

	for i := uint(0); i < min; i++ {
		if a[i] != b[i] {
			return max - 1 - i
		}
	}

	return diff
}

func IsEqNames(a string, b string) bool {
	return compareNames([]rune(a), []rune(b)) <= 2
}

func WalkFiles(uri Uri, extensions []string, cb func(Uri, string) error) (err error) {
	rootPath, err := UriToPath(uri)

	if err != nil {
		return
	}

	return filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := Ext(info.Name())

		if !slices.Contains(extensions, ext) {
			return nil
		}

		return cb(ToUri(path), ext)
	})
}

func createUniqYield[T Family | Member](yield func(*T) bool) func(*T) bool {
	list := make(map[*T]struct{})

	return func(item *T) bool {
		if item == nil {
			return false
		}

		_, exist := list[item]

		if exist {
			return false
		}

		list[item] = struct{}{}

		return !yield(item)
	}
}
