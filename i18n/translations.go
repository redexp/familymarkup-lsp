package i18n

type Translations map[string]Messages // [locale][key]message
type Messages map[string]string

var translations = Translations{
	"en": EN,
	"uk": UK,
	"ru": RU,
}
