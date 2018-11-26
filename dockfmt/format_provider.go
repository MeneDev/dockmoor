package dockfmt

import (
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"io"
)

type FormatProvider interface {
	Formats() []Format
}

func DefaultFormatProvider() FormatProvider {
	provider := new(defaultFormatProvider)

	return provider
}

var _ FormatProvider = (*defaultFormatProvider)(nil)

type defaultFormatProvider struct {
}

var registeredFormats []Format

func RegisterFormat(format Format) {
	registeredFormats = append(registeredFormats, format)
}

func (defaultFormatProvider) Formats() []Format {

	return registeredFormats
}

type UnknownFormatError struct {
	error
}

type AmbiguousFormatError struct {
	error
	Formats []Format
}

func IdentifyFormat(log logrus.FieldLogger, formatProvider FormatProvider, reader io.Reader, filename string) (Format, error) {
	formats := formatProvider.Formats()

	log = log.WithFields(logrus.Fields{
		"filename":     filename,
		"knownFormats": formats,
	})

	var format Format
	var formatErrors error
	for _, p := range formats {
		validationErr := p.ValidateInput(log, reader, filename)
		if validationErr != nil {
			formatErrors = multierror.Append(formatErrors, validationErr)
			log.WithFields(logrus.Fields{
				"format": p.Name(),
				"error":  validationErr,
			}).Debug("Tried incompatible format")
		} else {
			if format != nil {
				return nil, AmbiguousFormatError{
					Formats: []Format{format, p},
				}
			}

			format = p
		}
	}

	if format == nil {
		log.Info("Unknown Format")
		return nil, UnknownFormatError{
			formatErrors,
		}
	}

	return format, formatErrors
}
