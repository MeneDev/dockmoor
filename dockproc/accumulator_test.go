package dockproc

import (
	"bytes"
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"strconv"
	"testing"
)

var _ dockfmt.Format = (*delegatingFormatMock)(nil)

type delegatingFormatMock struct {
	NameDelegate          func() string
	ValidateInputDelegate func(log logrus.FieldLogger, reader io.Reader, filename string) error
	ProcessDelegate       func(log logrus.FieldLogger, reader io.Reader, writer io.Writer, imageNameProcessor dockfmt.ImageNameProcessor) error
}

func delegatingFormatMockNew() *delegatingFormatMock {
	return &delegatingFormatMock{
		NameDelegate: func() string {
			return "delegatingFormatMock"
		},
		ValidateInputDelegate: func(log logrus.FieldLogger, reader io.Reader, filename string) error {
			return nil
		},
		ProcessDelegate: func(log logrus.FieldLogger, reader io.Reader, writer io.Writer, imageNameProcessor dockfmt.ImageNameProcessor) error {
			return nil
		},
	}
}

func (m *delegatingFormatMock) Name() string {
	return m.NameDelegate()
}

func (m *delegatingFormatMock) ValidateInput(log logrus.FieldLogger, reader io.Reader, filename string) error {
	return m.ValidateInputDelegate(log, reader, filename)
}

func (m *delegatingFormatMock) Process(log logrus.FieldLogger, reader io.Reader, writer io.Writer, imageNameProcessor dockfmt.ImageNameProcessor) error {
	return m.ProcessDelegate(log, reader, writer, imageNameProcessor)
}

var _ Predicate = (*PredicateMock)(nil)

type PredicateMock struct {
	mock.Mock
}

func (m *PredicateMock) Matches(ref dockref.Reference) bool {
	called := m.Called(ref)
	return called.Bool(0)
}

func formatMockProcessing(images []string) *delegatingFormatMock {
	mockFormat := delegatingFormatMockNew()
	mockFormat.ProcessDelegate = func(log logrus.FieldLogger, reader io.Reader, writer io.Writer, imageNameProcessor dockfmt.ImageNameProcessor) error {
		for _, img := range images {
			ref, _ := dockref.FromOriginal(img)
			imageNameProcessor(ref)
		}
		return nil
	}
	return mockFormat
}

func TestContainsAccumulator_ReturnsErrorWhenParameterIsNil(t *testing.T) {
	logger, _ := logger()
	acc, err := MatchesAccumulatorNew(nil, logger, bytes.NewBuffer(nil))
	assert.Nil(t, acc)
	assert.Error(t, err)
}

func logger() (*logrus.Logger, *bytes.Buffer) {
	logger := logrus.New()
	buffer := bytes.NewBuffer(nil)
	logger.SetOutput(buffer)
	return logger, buffer
}

func TestContainsAccumulator(t *testing.T) {
	for _, num := range []int{0, 1, 2, 10} {
		imgs := []string{}
		for i := 0; i < num; i++ {
			imgs = append(imgs, "nginx")
		}

		mockFormat := formatMockProcessing(imgs)
		for _, matches := range []bool{true, false} {
			var desc string
			if matches {
				if num == 0 {
					continue
				}
				desc = "Matches when predicate matches " + strconv.Itoa(num) + " times"
			} else {
				desc = "Doesn't match when predicate doesn't match " + strconv.Itoa(num) + " times"
			}

			t.Run(desc, func(t *testing.T) {
				p := new(PredicateMock)
				p.On("Matches", mock.Anything).Return(matches)

				logger := logrus.New()
				logger.SetOutput(bytes.NewBuffer(nil))
				containsAccumulator, _ := MatchesAccumulatorNew(p, logger, bytes.NewBuffer(nil))

				formatProcessor := dockfmt.FormatProcessorNew(mockFormat, nil, nil)

				containsAccumulator.Accumulate(formatProcessor, NullHandler)
				result := containsAccumulator.Matches()

				if matches {
					assert.NotEmpty(t, result)
				} else {
					assert.Empty(t, result)
				}

				p.AssertNumberOfCalls(t, "Matches", num)
			})
		}
	}

	for _, num := range []int{1, 2, 3} {
		type matchInfo struct {
			name    string
			matches bool
		}

		mis := []matchInfo{}
		names := []string{}

		for i := 0; i < num; i++ {
			name := strconv.Itoa(i)
			mis = append(mis, matchInfo{name, i%2 == 0})
			names = append(names, name)
		}

		mockFormat := formatMockProcessing(names)

		desc := "Matches when alternating matches and non matches " + strconv.Itoa(num) + " times"

		t.Run(desc, func(t *testing.T) {
			p := new(PredicateMock)
			for _, mi := range mis {
				p.On("Matches", mock.Anything).Return(mi.matches).Once()
			}

			logger := logrus.New()
			logger.SetOutput(bytes.NewBuffer(nil))
			containsAccumulator, _ := MatchesAccumulatorNew(p, logger, bytes.NewBuffer(nil))

			formatProcessor := dockfmt.FormatProcessorNew(mockFormat, nil, nil)
			containsAccumulator.Accumulate(formatProcessor, NullHandler)
			result := containsAccumulator.Matches()
			assert.NotEmpty(t, result)
		})
	}
}

var _ dockref.Repository = (*MockRepository)(nil)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Resolve(reference dockref.Reference) (dockref.Reference, error) {
	called := m.Called(reference)
	refIf := called.Get(0)
	e := called.Error(1)

	ref := refIf.(dockref.Reference)

	return ref, e
}

func (m *MockRepository) OnResolve(reference dockref.Reference) *mock.Call {
	return m.On("Resolve", reference)
}

func TestPinAccumulator_Accumulate(t *testing.T) {
	log, _ := logger()

	output := bytes.NewBuffer(nil)

	ref, _ := dockref.FromOriginal("nginx")

	repository := new(MockRepository)
	resolved, _ := dockref.FromOriginal("nginx:latest@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf")
	repository.OnResolve(ref).Return(resolved, nil)

	accumulator, e := PinAccumulatorNew(AnyPredicateNew(), log, output, repository, dockref.FormatHasTag|dockref.FormatHasDigest)
	assert.Nil(t, e)
	assert.NotNil(t, accumulator)

	mockFormat := formatMockProcessing([]string{"nginx"})

	formatProcessor := dockfmt.FormatProcessorNew(mockFormat, nil, nil)

	accumulator.Accumulate(formatProcessor, NullHandler)

	assert.Len(t, accumulator.Results(), 1)
	assert.Equal(t, resolved.Original(), accumulator.Results()[0])
}
