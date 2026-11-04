// Command htmlmail2md converts HTML email content into Markdown.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ksonney/htmlmail2md/internal/converter"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "htmlmail2md: "+err.Error())
		os.Exit(1)
	}
}

// run wires the command line arguments into an HTML to Markdown conversion.
// It writes the resulting Markdown to stdout, or to the file given by -o.
func run(args []string, stdout, stderr io.Writer) error {
	var inputPath string
	var outPath string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-o" || arg == "--output":
			i++
			if i >= len(args) {
				return fmt.Errorf("flag %s requires a value", arg)
			}
			outPath = args[i]
		case strings.HasPrefix(arg, "--output="):
			outPath = strings.TrimPrefix(arg, "--output=")
		case len(arg) > 1 && strings.HasPrefix(arg, "-"):
			return fmt.Errorf("unknown flag: %s", arg)
		case inputPath == "":
			inputPath = arg
		default:
			return fmt.Errorf("unexpected argument: %s", arg)
		}
	}

	var input []byte
	if inputPath == "" || inputPath == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		input = data
	} else {
		data, err := os.ReadFile(inputPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", inputPath, err)
		}
		input = data
	}

	md, err := converter.Convert(input)
	if err != nil {
		return fmt.Errorf("convert: %w", err)
	}

	var out io.Writer = stdout
	if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
		defer f.Close()
		out = f
	}
	if _, err := io.WriteString(out, md); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
