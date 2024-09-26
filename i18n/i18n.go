package i18n

import "fmt"

var Locale = "en"

func L(key string, args ...any) string {
	msg := translations[Locale][key]

	return fmt.Sprintf(msg, args...)
}

func SetLocale(locale string) error {
	_, exist := translations[locale]

	if !exist {
		return fmt.Errorf("unsupported locale %s", locale)
	}

	Locale = locale

	return nil
}
