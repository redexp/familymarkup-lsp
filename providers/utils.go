package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	urlParser "net/url"
	"path/filepath"
)

func GetDoc(uri Uri) (doc *Doc) {
	root.UpdateLock.Lock()
	defer root.UpdateLock.Unlock()

	uri = NormalizeUri(uri)

	doc = root.Docs[uri]

	return
}

func NormalizeUri(uri Uri) Uri {
	url, err := urlParser.Parse(uri)

	if err != nil {
		panic(err)
	}

	return url.Scheme + "://" + url.Path
}

func EncUri(uri Uri) string {
	url, err := urlParser.Parse(uri)

	if err != nil {
		panic(err)
	}

	return url.String()
}

func RenameUri(uri Uri, name string) (Uri, error) {
	base, err := urlParser.Parse(uri)

	if err != nil {
		return "", err
	}

	base.Path = filepath.Join(base.Path, "..", name+filepath.Ext(base.Path))

	return base.String(), nil
}

func IsUriName(uri Uri, name string) bool {
	base := filepath.Base(uri)
	ext := filepath.Ext(uri)

	return name+ext == base
}

func P[T ~string | ~int32](src T) *T {
	return &src
}
