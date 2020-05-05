package main

import (
	"bytes"
	"errors"
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockproc"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/MeneDev/dockmoor/dockref/resolver"
	"github.com/MeneDev/dockmoor/docktst/dockreftst"
	"github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"os"
	"testing"
)

type pinOptionsTest struct {
	*pinOptions

	mainOptionsTest *mainOptionsTest
	mockResolver    *dockreftst.MockResolver
}

func (fo *pinOptionsTest) MainOptions() *mainOptionsTest {
	return fo.mainOptionsTest
}

func pinOptionsTestNew() *pinOptionsTest {
	mainOptions := mainOptionsTestNew()

	resolver := dockreftst.MockResolverNew()

	pinOptions := &pinOptionsTest{
		pinOptions:      pinOptionsNew(mainOptions.mainOptions),
		mainOptionsTest: mainOptions,
		mockResolver:    resolver,
	}
	pinOptions.resolverFactory = func(_name string) dockref.Resolver {
		return resolver
	}

	return pinOptions
}

func TestMainAsciiDocWithPin(t *testing.T) {
	os.Args = []string{"exe", "--asciidoc-usage"}

	mainOptions := mainOptionsACNew(addPinCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	helpText := buffer.String()
	assert.Contains(t, helpText, "pin command")

	assert.Equal(t, ExitSuccess, exitCode)
}

func TestMainMarkdownWithPin(t *testing.T) {
	os.Args = []string{"exe", "--markdown"}

	mainOptions := mainOptionsACNew(addPinCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	helpText := buffer.String()
	assert.Contains(t, helpText, "pin command")

	assert.Equal(t, ExitSuccess, exitCode)
}

func TestPinHelpIsNotAnError(t *testing.T) {
	os.Args = []string{"exe", "pin", "--help"}

	mainOptions := mainOptionsACNew(addPinCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	helpText := buffer.String()
	assert.Contains(t, helpText, "pin command")

	assert.Equal(t, ExitSuccess, exitCode)
}

func TestInvalidDockerfileWithPin(t *testing.T) {
	// given
	po := pinOptionsTestNew()
	mainOptions := po.MainOptions()

	formatProvider := mainOptions.FormatProvider()

	format := new(FormatMock)
	format.OnName().Return("mock")
	format.OnValidateInput(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("Not my department"))

	formatProvider.OnFormats().Return([]dockfmt.Format{format})

	mainOptions.formatProvider = formatProvider

	po.Positional.InputFile = flags.Filename(NotADockerfile)

	// when
	err := po.WithFormatProcessorDo(nil, func(processor dockfmt.FormatProcessor) error {
		return nil
	})

	// then
	assert.NotNil(t, err)

	_, ok := err.(dockfmt.UnknownFormatError)
	assert.True(t, ok)
}

func TestPinOptions_RefFormat(t *testing.T) {
	po := pinOptionsTestNew()
	po.ReferenceFormat.ForceDomain = false
	po.ReferenceFormat.NoName = false
	po.ReferenceFormat.NoTag = false
	po.ReferenceFormat.NoDigest = false
	t.Run("all unset", func(t *testing.T) {
		format, e := po.RefFormat()
		assert.Nil(t, e)
		assert.False(t, (format&dockref.FormatHasDomain) != 0)
		assert.True(t, (format&dockref.FormatHasName) != 0)
		assert.True(t, (format&dockref.FormatHasTag) != 0)
		assert.True(t, (format&dockref.FormatHasDigest) != 0)
	})

	po.ReferenceFormat.ForceDomain = true
	po.ReferenceFormat.NoName = true
	po.ReferenceFormat.NoTag = true
	po.ReferenceFormat.NoDigest = true
	t.Run("all set", func(t *testing.T) {
		_, e := po.RefFormat()
		assert.Error(t, e)
	})

	po.ReferenceFormat.ForceDomain = true
	po.ReferenceFormat.NoName = false
	po.ReferenceFormat.NoTag = true
	po.ReferenceFormat.NoDigest = true
	t.Run("all set but NoName", func(t *testing.T) {
		format, e := po.RefFormat()
		assert.Nil(t, e)
		assert.True(t, (format&dockref.FormatHasDomain) != 0)
		assert.True(t, (format&dockref.FormatHasName) != 0)
		assert.False(t, (format&dockref.FormatHasTag) != 0)
		assert.False(t, (format&dockref.FormatHasDigest) != 0)
	})

	po.ReferenceFormat.ForceDomain = true
	po.ReferenceFormat.NoName = false
	po.ReferenceFormat.NoTag = false
	po.ReferenceFormat.NoDigest = false
	t.Run("ForceDomain set", func(t *testing.T) {
		format, e := po.RefFormat()
		assert.Nil(t, e)
		assert.True(t, (format&dockref.FormatHasDomain) != 0)
		assert.True(t, (format&dockref.FormatHasName) != 0)
		assert.True(t, (format&dockref.FormatHasTag) != 0)
		assert.True(t, (format&dockref.FormatHasDigest) != 0)
	})

	po.ReferenceFormat.ForceDomain = false
	po.ReferenceFormat.NoName = true
	po.ReferenceFormat.NoTag = false
	po.ReferenceFormat.NoDigest = true
	t.Run("NoName and NoDigest set", func(t *testing.T) {
		_, e := po.RefFormat()
		assert.Error(t, e)
	})

	po.ReferenceFormat.ForceDomain = true
	po.ReferenceFormat.NoName = true
	po.ReferenceFormat.NoTag = false
	po.ReferenceFormat.NoDigest = false
	t.Run("ForceDomain and NoName set", func(t *testing.T) {
		_, e := po.RefFormat()
		assert.Error(t, e)
	})

	po.ReferenceFormat.ForceDomain = false
	po.ReferenceFormat.NoName = true
	po.ReferenceFormat.NoTag = false
	po.ReferenceFormat.NoDigest = true
	t.Run("NoName and NoDigest set", func(t *testing.T) {
		_, e := po.RefFormat()
		assert.Error(t, e)
	})

}

func TestPinCommand_applyFormatProcessor_FailsWithInvalidFormattingFlags(t *testing.T) {
	po := pinOptionsTestNew()
	po.ReferenceFormat.NoName = true
	po.ReferenceFormat.NoTag = true
	po.ReferenceFormat.NoDigest = true

	format := &FormatMock{}
	format.OnName().Return("Mock")
	format.OnProcess(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	processorMock := &FormatProcessorMock{}

	po.mockResolver.OnResolve(mock.Anything).
		Return(dockref.MustParse("nginx:tag@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"), nil)

	processorMock.process = func(imageNameProcessor dockfmt.ImageNameProcessor) error {
		_, e := imageNameProcessor(dockref.MustParse("nginx"))
		return e
	}

	predicate, e := dockproc.AnyPredicateNew()
	assert.Nil(t, e)
	err := po.applyFormatProcessor(predicate, processorMock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Reference Format")
}

func TestPinCommand_FailsWhenErrorInProcess(t *testing.T) {
	po := pinOptionsTestNew()
	po.ReferenceFormat.NoName = true
	po.ReferenceFormat.NoTag = true
	po.ReferenceFormat.NoDigest = true

	format := &FormatMock{}
	format.OnName().Return("Mock")
	expected := errors.New("A Process Error")

	format.OnProcess(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expected)

	format.OnValidateInput(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	formatProvider := &FormatProviderMock{}
	po.mainOptionsTest.formatProvider = formatProvider

	formatProvider.OnFormats().Return([]dockfmt.Format{format})

	exitCode, err := po.ExecuteWithExitCode(nil)
	assert.Error(t, err)
	assert.Equal(t, expected, err)
	assert.NotEqual(t, ExitSuccess, exitCode)
}

func TestPinCommandPins_unchanged(t *testing.T) {
	po := pinOptionsTestNew()
	po.mockResolver.OnResolve(dockref.MustParse("nginx")).
		Return(dockref.MustParse("nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"), nil)

	po.mockResolver.OnResolve(dockref.MustParse("nginx:tag")).
		Return(dockref.MustParse("nginx:tag@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"), nil)

	po.mockResolver.OnResolve(dockref.MustParse("nginx:latest")).
		Return(
			dockref.MustParse("nginx:latest@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"), nil)

	po.mockResolver.OnResolve(dockref.MustParse("menedev/testimagea")).
		Return(
			dockref.MustParse("menedev/testimagea:latest@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"), nil)

	processorMock := &FormatProcessorMock{}

	pin := func(refStr, expected string) {
		ran := false
		processorMock.process = func(imageNameProcessor dockfmt.ImageNameProcessor) error {
			ref, e := imageNameProcessor(dockref.MustParse(refStr))
			assert.Nil(t, e)
			str := ref.String()
			assert.Equal(t, expected, str)
			ran = true
			return nil
		}
		predicate, e := dockproc.AnyPredicateNew()
		assert.Nil(t, e)

		po.applyFormatProcessor(predicate, processorMock)
		assert.True(t, ran)
	}

	t.Run("tag and sha", func(t *testing.T) {
		po.ReferenceFormat.ForceDomain = false
		po.ReferenceFormat.NoName = false
		po.ReferenceFormat.NoTag = false
		po.ReferenceFormat.NoDigest = false

		pin("nginx:tag", "nginx:tag@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")
	})
	t.Run("tag only", func(t *testing.T) {
		po.ReferenceFormat.ForceDomain = false
		po.ReferenceFormat.NoName = false
		po.ReferenceFormat.NoTag = false
		po.ReferenceFormat.NoDigest = true

		pin("nginx:tag", "nginx:tag")
	})
	t.Run("sha and name", func(t *testing.T) {
		po.ReferenceFormat.ForceDomain = false
		po.ReferenceFormat.NoName = false
		po.ReferenceFormat.NoTag = true
		po.ReferenceFormat.NoDigest = false

		pin("nginx:tag", "nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")
	})
	t.Run("sha only", func(t *testing.T) {
		po.ReferenceFormat.ForceDomain = false
		po.ReferenceFormat.NoName = true
		po.ReferenceFormat.NoTag = true
		po.ReferenceFormat.NoDigest = false

		pin("nginx:tag", "d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")
	})

	t.Run("Does not add domain to well-known user image references when ForceDomain = false", func(t *testing.T) {
		po.ReferenceFormat.ForceDomain = false
		po.ReferenceFormat.NoName = false
		po.ReferenceFormat.NoTag = false
		po.ReferenceFormat.NoDigest = false

		pin("menedev/testimagea", "menedev/testimagea:latest@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991")
	})
}

func TestPinCommandPins_most_precise_version(t *testing.T) {
	t.SkipNow()

	//po := pinOptionsTestNew()
	//po.mockResolver.OnFindAllTags(dockref.MustParse("nginx")).
	//	Return([]dockref.Reference{dockref.MustParse("nginx:tag@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")}, nil)
	//
	//po.mockResolver.OnFindAllTags(dockref.MustParse("nginx:latest")).
	//	Return([]dockref.Reference{
	//		dockref.MustParse("nginx:1@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"),
	//		dockref.MustParse("nginx:1.15@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"),
	//		dockref.MustParse("nginx:1.15.6@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"),
	//		dockref.MustParse("nginx:latest@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"),
	//	}, nil)
	//
	//po.mockResolver.OnFindAllTags(dockref.MustParse("menedev/testimagea")).
	//	Return([]dockref.Reference{
	//		dockref.MustParse("menedev/testimagea:1.15.6@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"),
	//	}, nil)
	//
	//processorMock := &FormatProcessorMock{}
	//
	//pin := func(refStr, expected string) {
	//	ran := false
	//	processorMock.process = func(imageNameProcessor dockfmt.ImageNameProcessor) error {
	//		ref, e := imageNameProcessor(dockref.MustParse(refStr))
	//		assert.Nil(t, e)
	//		str := ref.String()
	//		assert.Equal(t, expected, str)
	//		ran = true
	//		return nil
	//	}
	//	predicate, e := dockproc.AnyPredicateNew()
	//	assert.Nil(t, e)
	//
	//	po.applyFormatProcessor(predicate, processorMock)
	//	assert.True(t, ran)
	//}
	//
	//pinNginx := func(expected string) {
	//	pin("nginx", expected)
	//}
	//
	//t.Run("tag and sha", func(t *testing.T) {
	//	po.ReferenceFormat.ForceDomain = false
	//	po.ReferenceFormat.NoName = false
	//	po.ReferenceFormat.NoTag = false
	//	po.ReferenceFormat.NoDigest = false
	//
	//	pinNginx("nginx:tag@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")
	//})
	//t.Run("tag only", func(t *testing.T) {
	//	po.ReferenceFormat.ForceDomain = false
	//	po.ReferenceFormat.NoName = false
	//	po.ReferenceFormat.NoTag = false
	//	po.ReferenceFormat.NoDigest = true
	//
	//	pinNginx("nginx:tag")
	//})
	//t.Run("sha and name", func(t *testing.T) {
	//	po.ReferenceFormat.ForceDomain = false
	//	po.ReferenceFormat.NoName = false
	//	po.ReferenceFormat.NoTag = true
	//	po.ReferenceFormat.NoDigest = false
	//
	//	pinNginx("nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")
	//})
	//t.Run("sha only", func(t *testing.T) {
	//	po.ReferenceFormat.ForceDomain = false
	//	po.ReferenceFormat.NoName = true
	//	po.ReferenceFormat.NoTag = true
	//	po.ReferenceFormat.NoDigest = false
	//
	//	pinNginx("d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")
	//})
	//
	//t.Run("FindAllTags to most precise tag", func(t *testing.T) {
	//	po.ReferenceFormat.ForceDomain = false
	//	po.ReferenceFormat.NoName = false
	//	po.ReferenceFormat.NoTag = false
	//	po.ReferenceFormat.NoDigest = false
	//
	//	pin("nginx:latest", "nginx:1.15.6@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991")
	//})
	//t.Run("Does not add domain to well-known user image references when ForceDomain = false", func(t *testing.T) {
	//	po.ReferenceFormat.ForceDomain = false
	//	po.ReferenceFormat.NoName = false
	//	po.ReferenceFormat.NoTag = false
	//	po.ReferenceFormat.NoDigest = false
	//
	//	pin("menedev/testimagea", "menedev/testimagea:1.15.6@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991")
	//})
}

func TestFilenameRequiredWithPin(t *testing.T) {
	_, _, exitCode, stdout := testMain([]string{"pin"}, addPinCommand)
	assert.NotEqual(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "level=error")
	assert.Contains(t, stdout.String(), "the required argument `InputFile`")
}

func TestPinCallsFindExecuteWithPin(t *testing.T) {
	cmd, _, _, stdout := testMain([]string{"pin", "fileName"}, addPinCommand)

	_, ok := cmd.(*pinOptions)
	assert.True(t, ok)
	assert.Empty(t, stdout)
}

func TestPinWritesToInputFile(t *testing.T) {
	df1 := dockerfile(`FROM img`)
	defer os.Remove(df1)

	os.Args = []string{"exe", "pin", df1}

	rslvr := dockreftst.MockResolverNew()

	rslvr.OnFindAllTags(dockref.MustParse("img")).Return([]dockref.Reference{
		dockref.MustParse("img:1.2.3@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"),
	}, nil)
	rslvr.OnResolve(dockref.MustParse("img")).Return(
		dockref.MustParse("img:1.2.3@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"), nil)

	rslvr.OnResolve(dockref.MustParse("img")).Return(
		dockref.MustParse("img:1.2.3@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"), nil)

	mainOptions := mainOptionsACNew(addPinCommandWith(func(mainOptions *mainOptions) *pinOptions {
		po := pinOptionsNew(mainOptions)

		po.resolverFactory = func(_name string) dockref.Resolver {
			return rslvr
		}

		return po
	}))

	exitCode := doMain(mainOptions)

	assert.Equal(t, ExitSuccess, exitCode)

	dfBytes, e := ioutil.ReadFile(df1)
	assert.Nil(t, e)

	s := string(dfBytes)

	assert.Equal(t, `FROM img:1.2.3@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf`, s)
}

func TestPinWritesToOutputFileAndNotToInputfile(t *testing.T) {
	df1 := dockerfile(`FROM img`)
	defer os.Remove(df1)

	df2 := tmpFile().Name()
	defer os.Remove(df2)

	os.Args = []string{"exe", "pin", "--output", df2, df1}

	resolver := dockreftst.MockResolverNew()

	resolver.OnFindAllTags(dockref.MustParse("img")).Return([]dockref.Reference{
		dockref.MustParse("img:1.2.3@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"),
	}, nil)
	resolver.OnResolve(dockref.MustParse("img")).Return(
		dockref.MustParse("img:1.2.3@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf"), nil)

	mainOptions := mainOptionsACNew(addPinCommandWith(func(mainOptions *mainOptions) *pinOptions {
		po := pinOptionsNew(mainOptions)

		po.resolverFactory = func(_name string) dockref.Resolver {
			return resolver
		}

		return po
	}))

	exitCode := doMain(mainOptions)

	assert.Equal(t, ExitSuccess, exitCode)

	fileBytes, e := ioutil.ReadFile(df1)
	assert.Nil(t, e)

	s := string(fileBytes)

	assert.Equal(t, `FROM img`, s)

	fileBytes, e = ioutil.ReadFile(df2)
	assert.Nil(t, e)

	s = string(fileBytes)

	assert.Equal(t, `FROM img:1.2.3@sha256:2c4269d573d9fc6e9e95d5e8f3de2dd0b07c19912551f25e848415b5dd783acf`, s)

}

func TestPinOptions_applyFormatProcessor_ReturnsError(t *testing.T) {
	po := pinOptionsTestNew()
	expected := errors.New("An error")

	processorMock := &FormatProcessorMock{}
	processorMock.process = func(imageNameProcessor dockfmt.ImageNameProcessor) error {
		return expected
	}

	predicate, e := dockproc.AnyPredicateNew()
	assert.Nil(t, e)

	err := po.applyFormatProcessor(predicate, processorMock)

	assert.Equal(t, expected, err)
}

func TestInvalidSolverIsNil(t *testing.T) {
	po := pinOptionsNew(nil)
	testMain([]string{"pin", "--resolver", "Invalid", "fileNameIn", "fileNameOut"}, addPinCommandWith(func(mainOptions *mainOptions) *pinOptions {
		return po
	}))

	solverFactory := po.resolverFactory("Invalid")
	assert.Nil(t, solverFactory)
}

func TestUsesDockerdResolver(t *testing.T) {
	po := pinOptionsNew(nil)
	testMain([]string{"pin", "--resolver", "dockerd", "fileNameIn", "fileNameOut"}, addPinCommandWith(func(mainOptions *mainOptions) *pinOptions {
		return po
	}))
	assert.Equal(t, "dockerd", po.PinOptions.Resolver)
	assert.IsType(t, resolver.DockerDaemonResolverNew(), po.resolverFactory(po.PinOptions.Resolver))
}

func TestUsesRegistryResolver(t *testing.T) {
	po := pinOptionsNew(nil)
	testMain([]string{"pin", "--resolver", "registry", "fileNameIn", "fileNameOut"}, addPinCommandWith(func(mainOptions *mainOptions) *pinOptions {
		return po
	}))
	assert.Equal(t, "registry", po.PinOptions.Resolver)
	assert.IsType(t, resolver.DockerRegistryResolverNew(), po.resolverFactory(po.PinOptions.Resolver))
}
