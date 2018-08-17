package dockproc

import (
	"strconv"
	"testing"
	"github.com/stretchr/testify/mock"
	"github.com/MeneDev/dockfix/dockfmt"
	"github.com/stretchr/testify/assert"
	"io"
	"github.com/sirupsen/logrus"
	"github.com/MeneDev/dockfix/dockref"
)

var _ dockfmt.Format = (*delegatingFormatMock)(nil)

type delegatingFormatMock struct {
	NameDelegate          func() string
	ValidateInputDelegate func(log logrus.FieldLogger, reader io.Reader, filename string) error
	ProcessDelegate       func(log logrus.FieldLogger, reader io.Reader, writer io.Writer, imageNameProcessor dockfmt.ImageNameProcessor) error
}

func DelegatingFormatMockNew() *delegatingFormatMock {
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
	mockFormat := DelegatingFormatMockNew()
	mockFormat.ProcessDelegate = func(log logrus.FieldLogger, reader io.Reader, writer io.Writer, imageNameProcessor dockfmt.ImageNameProcessor) error {
		for _, img := range images {
			ref, _ := dockref.FromOriginal(img)
			imageNameProcessor(ref)
		}
		return nil
	}
	return mockFormat
}

func TestFindAccumulator(t *testing.T) {
	for _, num := range []int{0, 1, 2, 10} {
		imgs := []string{}
		for i := 0; i < num; i++ {
			imgs = append(imgs, "nginx")
		}

		mockFormat := formatMockProcessing(imgs)
		for _, matches := range []bool{true, false} {
			var desc string
			if (matches) {
				if num == 0 {
					continue
				}
				desc = "Finds when predicate matches " + strconv.Itoa(num) + " times"
			} else {
				desc = "Doesn't find when predicate doesn't match " + strconv.Itoa(num) + " times"
			}

			t.Run(desc, func(t *testing.T) {
				p := new(PredicateMock)
				p.On("Matches", mock.Anything).Return(matches)

				findAccumulator, _ := FindAccumulatorNew(p)

				formatProcessor := dockfmt.FormatProcessorNew(mockFormat, nil, nil)

				result := findAccumulator.Accumulate(formatProcessor)
				assert.Equal(t, matches, result)
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

		desc := "Finds when alternating matches and non matches " + strconv.Itoa(num) + " times"

		t.Run(desc, func(t *testing.T) {
			p := new(PredicateMock)
			for _, mi := range mis {
				p.On("Matches", mock.Anything).Return(mi.matches).Once()
			}

			findAccumulator, _ := FindAccumulatorNew(p)

			formatProcessor := dockfmt.FormatProcessorNew(mockFormat, nil, nil)

			result := findAccumulator.Accumulate(formatProcessor)
			assert.True(t, result)
		})
	}
}
