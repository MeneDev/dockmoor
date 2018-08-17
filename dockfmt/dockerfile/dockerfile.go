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
}

func (pinner *dockerfileFormat) Name() string {
	return "Dockerfile"
}

func DockerfileFormatNew() *dockerfileFormat {
	pinner := new(dockerfileFormat)
	return pinner
}

func (pinner *dockerfileFormat) ValidateInput(log logrus.FieldLogger, reader io.Reader, filename string) error {
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

	result, err := parser.Parse(strings.NewReader(full))
	if err != nil {
		return err
	}

	// the parser will just ignore "unknown" commands in the Dockerfile but not report any error
	for _, cmd := range result.AST.Children {
		if _, ok := command.Commands[cmd.Value]; !ok {
			return errors.Errorf("Unknown command %s", cmd)
		}
	}
	pinner.lines = lines
	pinner.result = result

	return nil
}

func (pinner *dockerfileFormat) Process(log logrus.FieldLogger, reader io.Reader, w io.Writer, imageNameProcessor dockfmt.ImageNameProcessor) error {
	writer := bufio.NewWriter(w)
	defer writer.Flush()

	root := pinner.result.AST
	lines := pinner.lines

	curLineNum := 0
	for _, cmd := range root.Children {
		curLineNum++
		for i := curLineNum; i < cmd.StartLine; i++ {
			writer.WriteString(lines[i - 1])
			curLineNum++
		}

		handled, err := pinner.processNode(log, cmd, writer, imageNameProcessor)
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

func (pinner *dockerfileFormat) processChildren(log logrus.FieldLogger, node *parser.Node, writer *bufio.Writer, imageNameProcessor dockfmt.ImageNameProcessor) (bool, error) {
	currentLine := 1
	handled := false
	for _, n := range node.Children {
		for currentLine < n.StartLine  {
			writer.WriteString("\n")
			currentLine++
		}
		nodeHandled, err := pinner.processNode(log, n, writer, imageNameProcessor)
		if err != nil {
			return false, err
		}
		handled = handled || nodeHandled
	}

	return handled, nil
}

func (pinner *dockerfileFormat) processNode(log logrus.FieldLogger, node *parser.Node, writer *bufio.Writer, imageNameProcessor dockfmt.ImageNameProcessor) (bool, error) {
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
			writer.WriteString(strings.Replace(pinner.lines[i - 1], from, canonicalString, 1))
		}
		return true, nil
	} else {
		// pass-through
		return false, nil
	}
}
