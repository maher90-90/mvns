package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

type Locale struct {
	messages map[string]string
}

func NewFromFS(fsys fs.ReadFileFS, dir, lang string) (*Locale, error) {
	var path string
	if dir == "." || dir == "" {
		path = lang + ".json"
	} else {
		path = dir + "/" + lang + ".json"
	}
	data, err := fsys.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load locale %q: %w", lang, err)
	}

	var messages map[string]string
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("parse locale %q: %w", lang, err)
	}

	return &Locale{messages: messages}, nil
}

func (l *Locale) T(key string) string {
	if msg, ok := l.messages[key]; ok {
		return msg
	}
	return key
}
