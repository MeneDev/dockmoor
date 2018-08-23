package dockfmt

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"bytes"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/pkg/errors"
	"github.com/hashicorp/go-multierror"
)


func TestIdentifyFormatWithEmptyFormatProvider(t *testing.T) {

	formatProviderMock := new(FormatProviderMock)
	formatProviderMock.On("Formats").Return([]Format{})

	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})

	format, e := IdentifyFormat(logger, formatProviderMock, bytes.NewBufferString("not a dockerfile"), "filename")
	_, ok := e.(UnknownFormatError)
	assert.True(t, ok)

	assert.Nil(t, format)
}

func TestIdentifyFormatWithSingleNonMatchingFormat(t *testing.T) {

	formatMock := new(FormatMock)
	formatMock.On("ValidateInput", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
	formatProviderMock := new(FormatProviderMock)
	formatProviderMock.On("Formats").Return([]Format{
		formatMock,
	})

	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})

	format, e := IdentifyFormat(logger, formatProviderMock, bytes.NewBufferString("not a dockerfile"), "filename")

	assert.Nil(t, format)

	_, ok := e.(UnknownFormatError)
	assert.True(t, ok)
}

func TestIdentifyFormatWithSingleMatchingFormat(t *testing.T) {

	formatMock := new(FormatMock)
	formatMock.On("ValidateInput", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	formatProviderMock := new(FormatProviderMock)
	formatProviderMock.On("Formats").Return([]Format{
		formatMock,
	})

	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})

	format, e := IdentifyFormat(logger, formatProviderMock, bytes.NewBufferString("not a dockerfile"), "filename")

	assert.Nil(t, e)
	assert.Equal(t, formatMock, format)
}

func TestIdentifyFormatWithSingleMatchingAndSeveralNonMatchingFormats(t *testing.T) {

	matchingFormatMock := new(FormatMock)
	matchingFormatMock.On("ValidateInput", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	nonMatchingFormatMock1 := new(FormatMock)
	nonMatchingFormatMock1.On("ValidateInput", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
	nonMatchingFormatMock2 := new(FormatMock)
	nonMatchingFormatMock2.On("ValidateInput", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
	formatProviderMock := new(FormatProviderMock)
	formatProviderMock.On("Formats").Return([]Format{
		nonMatchingFormatMock1,
		matchingFormatMock,
		nonMatchingFormatMock2,
	})

	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})

	format, e := IdentifyFormat(logger, formatProviderMock, bytes.NewBufferString("not a dockerfile"), "filename")

	assert.Equal(t, matchingFormatMock, format)
	multiError, ok := e.(*multierror.Error)
	assert.True(t, ok)

	assert.Len(t, multiError.Errors, 2)
}

func TestIdentifyFormatWithSeveralMatchingAndSeveralNonMatchingFormats(t *testing.T) {
	matchingFormatMock1 := new(FormatMock)
	matchingFormatMock1.On("ValidateInput", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	matchingFormatMock2 := new(FormatMock)
	matchingFormatMock2.On("ValidateInput", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	nonMatchingFormatMock1 := new(FormatMock)
	nonMatchingFormatMock1.On("ValidateInput", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
	nonMatchingFormatMock2 := new(FormatMock)
	nonMatchingFormatMock2.On("ValidateInput", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
	formatProviderMock := new(FormatProviderMock)
	formatProviderMock.On("Formats").Return([]Format{
		nonMatchingFormatMock1,
		matchingFormatMock1,
		matchingFormatMock2,
		nonMatchingFormatMock2,
	})

	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})

	format, e := IdentifyFormat(logger, formatProviderMock, bytes.NewBufferString("not a dockerfile"), "filename")

	assert.Nil(t, format)
	assert.NotNil(t, e)

	ambiguousFormatError, ok := e.(AmbiguousFormatError)
	assert.True(t, ok)

	assert.Contains(t, ambiguousFormatError.Formats, matchingFormatMock1)
	assert.Contains(t, ambiguousFormatError.Formats, matchingFormatMock2)
}

func TestDefaultFormatProviderExits(t *testing.T) {
	provider := DefaultFormatProvider()
	assert.NotNil(t, provider)
}

func TestDefaultFormatProviderProvidesRegisteredFormat(t *testing.T) {

	provider := DefaultFormatProvider()
	formatMock := new(FormatMock)

	formats := provider.Formats()
	assert.NotContains(t, formats, formatMock)

	RegisterFormat(formatMock)
	formats = provider.Formats()

	assert.Contains(t, formats, formatMock)
}

