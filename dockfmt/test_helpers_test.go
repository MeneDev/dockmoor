package dockfmt

import (
	"github.com/stretchr/testify/mock"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
)

var _ FormatProvider = (*FormatProviderMock)(nil)

type FormatProviderMock struct {
	mock.Mock
}

func (m *FormatProviderMock) Formats() []Format {
	called := m.Called()
	return getFormats(called, 0)
}
func (m *FormatProviderMock) OnFormats() *mock.Call {
	return m.On("Formats")
}

func getFormats(args mock.Arguments, index int) []Format {
	obj := args.Get(index)
	var v []Format
	var ok bool
	if obj == nil {
		return nil
	}
	if v, ok = obj.([]Format); !ok {
		panic(fmt.Sprintf("assert: arguments: Error(%d) failed because object wasn't correct type: %v", index, args.Get(index)))
	}
	return v

}

var _ Format = (*FormatMock)(nil)

type FormatMock struct {
	mock.Mock
}

func (m *FormatMock) Name() string {
	called := m.Called()
	return called.String(0)
}
func (m *FormatMock) OnName() *mock.Call {
	return m.On("Name")
}

func (m *FormatMock) ValidateInput(log logrus.FieldLogger, reader io.Reader, filename string) error {
	called := m.Called(log, reader, filename)
	return called.Error(0)
}
func (m *FormatMock) OnValidateInput(log interface{}, reader interface{}, filename interface{}) *mock.Call {
	return m.On("ValidateInput", log, reader, filename)
}

func (m *FormatMock) Process(log logrus.FieldLogger, reader io.Reader, writer io.Writer, imageNameProcessor ImageNameProcessor) error {
	called := m.Called(log, reader, writer, imageNameProcessor)
	return called.Error(0)
}
func (m *FormatMock) OnProcess(log interface{}, reader interface{}, writer interface{}, imageNameProcessor interface{}) *mock.Call {
	return m.On("Process", log, reader, writer, imageNameProcessor)
}
