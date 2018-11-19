package dockerfile

import (
	"bufio"
	"bytes"
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/hashicorp/go-multierror"
	"github.com/moby/buildkit/frontend/dockerfile/command"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"reflect"
	"strings"
)

func init() {
	dockfmt.RegisterFormat(New())
}

// ensure Format is implemented
var _ dockfmt.Format = (*dockerfileFormat)(nil)

type dockerfileFormat struct {
	lines         []string
	result        *parser.Result
	parseFunction func(rwc io.Reader) (*parser.Result, error)
}

func (format *dockerfileFormat) Name() string {
	return "Dockerfile"
}

func New() dockfmt.Format {
	return newDockerfileFormat()
}

func newDockerfileFormat() *dockerfileFormat {
	format := new(dockerfileFormat)
	format.parseFunction = parser.Parse
	return format
}

func dockerfileFormatSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func (format *dockerfileFormat) ValidateInput(log logrus.FieldLogger, reader io.Reader, filename string) error {
	scanner := bufio.NewScanner(reader)
	var split bufio.SplitFunc = dockerfileFormatSplitFunc

	full := ""
	lines := make([]string, 0)
	scanner.Split(split)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		full = full + line
	}

	result, err := format.parseFunction(strings.NewReader(full))
	if err != nil {
		return err
	}

	// the parser will just ignore "unknown" commands in the Dockerfile but not report any error
	for _, cmd := range result.AST.Children {
		if _, ok := command.Commands[cmd.Value]; !ok {
			return errors.Errorf("Unknown command %s", cmd.Value)
		}
	}

	if result.AST.Children == nil {
		return errors.Errorf("No commands found")
	}

	// contains at least one FROM command
	containsFrom := false
	for _, cmd := range result.AST.Children {
		if cmd.Value == "from" {
			containsFrom = true
		}
	}

	if !containsFrom {
		return errors.Errorf("No FROM command found")
	}

	format.lines = lines
	format.result = result

	return nil
}

func saveFlush(log logrus.FieldLogger, writer *bufio.Writer) {
	err := writer.Flush()
	if err != nil {
		log.Errorf("Error flushing writer: %s", err.Error())
	}
}

func (format *dockerfileFormat) Process(log logrus.FieldLogger, reader io.Reader, w io.Writer, imageNameProcessor dockfmt.ImageNameProcessor) error {
	result := new(multierror.Error)
	writer := bufio.NewWriter(w)

	defer saveFlush(log, writer)

	root := format.result.AST
	lines := format.lines

	curLineNum := 0
	for _, cmd := range root.Children {
		curLineNum++
		for i := curLineNum; i < cmd.StartLine; i++ {
			_, err := writer.WriteString(lines[i-1])
			result = multierror.Append(result, err)
			curLineNum++
		}

		handled, err := format.processNode(log, cmd, writer, imageNameProcessor)
		if err != nil {
			return err
		}

		endLine := endLineOfNode(cmd)

		if !handled {
			for i := cmd.StartLine; i <= endLine; i++ {
				_, err := writer.WriteString(lines[i-1])
				result = multierror.Append(result, err)
			}
		}
		curLineNum = endLine
	}

	lastCommand := root.Children[len(root.Children)-1]
	endLine := endLineOfNode(lastCommand)

	for i := endLine; i < len(lines); i++ {
		_, err := writer.WriteString(lines[i])
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func endLineOfNode(command *parser.Node) int {
	v := reflect.ValueOf(*command)
	y := v.FieldByName("endLine")
	endLine := int(y.Int())
	return endLine
}

func (format *dockerfileFormat) processNode(log logrus.FieldLogger, node *parser.Node, writer *bufio.Writer, imageNameProcessor dockfmt.ImageNameProcessor) (bool, error) {
	result := new(multierror.Error)

	if node.Value == "from" {
		from := node.Next.Value
		log.Infof("Found image %s", from)

		ref, err := dockref.FromOriginal(from)
		if err != nil {
			return false, err
		}

		processed, err := imageNameProcessor(ref)
		if err != nil {
			return false, err
		}

		if processed != ref {
			log.Infof("Pinning '%s' as '%s'", from, processed)
		}

		//writer.WriteString(`FROM `)
		//writer.WriteString(processed)

		//next := node.Next.Next

		//for next != nil {
		//	writer.WriteString(` `)
		//	writer.WriteString(next.Value)
		//
		//	next = next.Next
		//}

		end := endLineOfNode(node)
		start := node.StartLine

		for i := start; i <= end; i++ {
			_, err := writer.WriteString(strings.Replace(format.lines[i-1], from, processed.String(), 1))
			result = multierror.Append(result, err)
		}
		return true, result.ErrorOrNil()
	}
	// pass-through
	return false, result.ErrorOrNil()
}
