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

	var pinner Format
	var pinnerErrors error
	for _, p := range formats {
		validationErr := p.ValidateInput(log, reader, filename)
		if validationErr != nil {
			pinnerErrors = multierror.Append(pinnerErrors, validationErr)
			log.WithFields(logrus.Fields{
				"format": p,
				"error":  validationErr,
			}).Debug("Tried incompatible format")
		} else {
			if pinner != nil {
				return nil, AmbiguousFormatError{
					Formats: []Format{pinner, p},
				}
			}

			pinner = p
		}
	}

	if pinner == nil {
		log.Info("Unknown Format")
		return nil, UnknownFormatError{
			pinnerErrors,
		}
	}

	return pinner, pinnerErrors
}
