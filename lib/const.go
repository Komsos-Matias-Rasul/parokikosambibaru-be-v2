package lib

import "errors"

var ErrArticleNotFound error = errors.New("article not found")
var ErrInvalidArticle error = errors.New("invalid article id")
var ErrInvalidCategory error = errors.New("invalid category id")
var ErrInvalidEdition error = errors.New("invalid edition id")
var ErrInvalidYear error = errors.New("invalid year")
var ErrInvalidBody error = errors.New("invalid request body")

var ErrDatabase error = errors.New("mysql: database error")
var ErrTimeout error = errors.New("mysql: connection timeout")

var ErrStorage error = errors.New("storage: storage connection error")
var ErrNoObject error = errors.New("storage: object not found")
var ErrReadFailure error = errors.New("storage: failed to read object")
