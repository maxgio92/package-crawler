package rpm

import (
	"github.com/pkg/errors"
)

var (
	ErrDBFormatNotSupported     = errors.New("the database file format is not supported")
	ErrDBMetadataResponseEmpty  = errors.New("response body is nil")
	ErrSearchPackagaNameMissing = errors.New("at least one package name must be specified")
)
