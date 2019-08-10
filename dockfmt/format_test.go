package dockfmt

import (
	"bytes"
	"strings"
	"testing"

	"github.com/MeneDev/dockmoor/dockref"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFormatProcessor_ProcessPassesLogAndReaderAndImageProcessor(t *testing.T) {
	log := logrus.New()

	formatMock := new(FormatMock)
	reader := strings.NewReader("input")
	processorFx := func(r dockref.Reference) (dockref.Reference, error) {
		return r, nil
	}
	formatMock.On("Process", log, reader, bytes.NewBuffer(nil), mock.Anything).Return(nil)

	processor := FormatProcessorNew(formatMock, log, reader)

	processor.Process(processorFx)

	formatMock.Mock.AssertNumberOfCalls(t, "Process", 1)
}

func TestFormatProcessor_ProcessPassesLogAndReaderAndImageProcessorAndWriter(t *testing.T) {
	log := logrus.New()

	formatMock := new(FormatMock)
	reader := strings.NewReader("input")
	processorFx := func(r dockref.Reference) (dockref.Reference, error) {
		return r, nil
	}
	writer := bytes.NewBufferString("writer")
	formatMock.On("Process", log, reader, writer, mock.Anything).Return(nil)

	processor := FormatProcessorNew(formatMock, log, reader).WithWriter(writer)

	processor.Process(processorFx)

	formatMock.Mock.AssertNumberOfCalls(t, "Process", 1)
}

func TestFormatError(t *testing.T) {
	err := errors.New("test")
	formatError := FormatErrorNew(err)

	assert.Error(t, formatError)
	assert.Error(t, formatError.reason)

	assert.Equal(t, "FormatError: test", formatError.Error())
}
