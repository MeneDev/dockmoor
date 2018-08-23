package dockerfile

import (
	"io"
	"github.com/sirupsen/logrus"
	"bufio"
	"bytes"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/command"
	"strings"
	"reflect"
	"github.com/pkg/errors"
	"github.com/MeneDev/dockfix/dockfmt"
	"github.com/MeneDev/dockfix/dockref"
)

func init()  {
	dockfmt.RegisterFormat(DockerfileFormatNew())
}

// ensure Format is implemented
var _ dockfmt.Format = (*dockerfileFormat)(nil)
type dockerfileFormat struct {
	lines []string
	result *parser.Result
	parseFunction func(rwc io.Reader) (*parser.Result, error)
}

func (format *dockerfileFormat) Name() string {
	return "Dockerfile"
}

func DockerfileFormatNew() *dockerfileFormat {
	format := new(dockerfileFormat)
	format.parseFunction = parser.Parse
	return format
}

func (format *dockerfileFormat) ValidateInput(log logrus.FieldLogger, reader io.Reader, filename string) error {
	scanner := bufio.NewScanner(reader)
	var split bufio.SplitFunc = func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			// We have a full newline-terminated line.
			return i + 1, data[0:i + 1], nil
		}
		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}

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
			return errors.Errorf("Unknown command %s", cmd)
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

func (format *dockerfileFormat) Process(log logrus.FieldLogger, reader io.Reader, w io.Writer, imageNameProcessor dockfmt.ImageNameProcessor) error {
	writer := bufio.NewWriter(w)
	defer writer.Flush()

	root := format.result.AST
	lines := format.lines

	curLineNum := 0
	for _, cmd := range root.Children {
		curLineNum++
		for i := curLineNum; i < cmd.StartLine; i++ {
			writer.WriteString(lines[i - 1])
			curLineNum++
		}

		handled, err := format.processNode(log, cmd, writer, imageNameProcessor)
		if err != nil {
			return err
		}

		endLine := endLineOfNode(cmd)

		if !handled {
			for i := cmd.StartLine; i <= endLine; i++ {
				writer.WriteString(lines[i-1])
			}
		}
		curLineNum = endLine
	}

	lastCommand := root.Children[len(root.Children)-1]
	endLine := endLineOfNode(lastCommand)

	for i := endLine; i < len(lines); i++ {
		writer.WriteString(lines[i])
	}

	return nil
}

func endLineOfNode(command *parser.Node) int {
	v := reflect.ValueOf(*command)
	y := v.FieldByName("endLine")
	endLine := int(y.Int())
	return endLine
}

func (format *dockerfileFormat) processNode(log logrus.FieldLogger, node *parser.Node, writer *bufio.Writer, imageNameProcessor dockfmt.ImageNameProcessor) (bool, error) {
	if node.Value == "from" {
		from := node.Next.Value
		log.Infof("Found image %s", from)

		ref, err := dockref.FromOriginal(from)
		if err != nil {
			return false, err
		}

		canonicalString, err := imageNameProcessor(ref)
		if err != nil {
			return false, err
		}

		log.Infof("Pinning %s as %s", from, canonicalString)

		//writer.WriteString(`FROM `)
		//writer.WriteString(canonicalString)

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
			writer.WriteString(strings.Replace(format.lines[i - 1], from, canonicalString, 1))
		}
		return true, nil
	} else {
		// pass-through
		return false, nil
	}
}
