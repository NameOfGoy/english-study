package dictionary

import "errors"

var (
	ErrWordExist    = errors.New("word exist")
	ErrWordNotExist = errors.New("word not exist")
	ErrWordAdding   = errors.New("word has been adding")
)
