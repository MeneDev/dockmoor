package dockfmt

import (
	"testing"
	"github.com/sirupsen/logrus"
	"strings"
	"github.com/MeneDev/dockfix/dockref"
	"github.com/stretchr/testify/mock"
	"bytes"
)

func TestFormatProcessor_ProcessPassesLogAndReaderAndImageProcessor(t *testing.T) {
	log := logrus.New()

	formatMock := new(FormatMock)
	reader := strings.NewReader("input")
	processorFx := func(r dockref.Reference) (string, error) {
		return "", nil
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
	processorFx := func(r dockref.Reference) (string, error) {
		return "", nil
	}
	writer := bytes.NewBufferString("writer")
	formatMock.On("Process", log, reader, writer, mock.Anything).Return(nil)

	processor := FormatProcessorNew(formatMock, log, reader).WithWriter(writer)


	processor.Process(processorFx)

	formatMock.Mock.AssertNumberOfCalls(t, "Process", 1)
}
