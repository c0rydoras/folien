package code

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/c0rydoras/folien/pkg/parser"
)

// Block represents a code block.
type Block struct {
	Code     string
	Language string
}

// Result represents the output for an executed code block.
type Result struct {
	Out           string
	ExitCode      int
	ExecutionTime time.Duration
}

var (
	// ErrParse is the returned error when we cannot parse the code block (i.e.
	// there is no code block on the current slide) or the code block is
	// incorrectly written.
	ErrParse = errors.New("error: could not parse code block")
)

// Parse takes a block of markdown and returns an array of Block's with code
// and associated languages
func Parse(markdown string) ([]Block, error) {
	codeBlocks := parser.CollectCodeBlocks([]byte(markdown))

	var rv []Block

	for _, block := range codeBlocks {
		rv = append(rv, Block{
			Language: string(block.Language([]byte(markdown))),
			Code:     RemoveComments(string(block.Lines().Value([]byte(markdown)))),
		})
	}

	if len(rv) == 0 {
		return nil, ErrParse
	}

	return rv, nil
}

const (
	// ExitCodeInternalError represents the exit code in which the code
	// executing the code didn't work.
	ExitCodeInternalError = -1
)

// Execute takes a code.Block and returns the output of the executed code
func Execute(code Block) Result {
	// Check supported language
	language, ok := Languages[code.Language]
	if !ok {
		return Result{
			Out:      "Error: unsupported language",
			ExitCode: ExitCodeInternalError,
		}
	}

	// Write the code block to a temporary file
	f, err := os.CreateTemp(os.TempDir(), "folien-*."+Languages[code.Language].Extension)
	if err != nil {
		return Result{
			Out:      "Error: could not create file",
			ExitCode: ExitCodeInternalError,
		}
	}

	defer func() {
		if err := f.Close(); err != nil {
			_ = err // ignore error
		}
	}()
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			_ = err // ignore error
		}
	}()

	_, err = f.WriteString(TransformCode(code.Language, code.Code))
	if err != nil {
		return Result{
			Out:      "Error: could not write to file",
			ExitCode: ExitCodeInternalError,
		}
	}

	var (
		output   strings.Builder
		exitCode int
	)

	// replacer for commands
	repl := strings.NewReplacer(
		"<file>", f.Name(),
		// <name>: file name without extension and without path
		"<name>", filepath.Base(strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))),
		"<path>", filepath.Dir(f.Name()),
	)

	// For accuracy of program execution speed, we can't put anything after
	// recording the start time or before recording the end time.
	start := time.Now()

	for _, c := range language.Commands {
		var command []string
		// replace <file>, <name> and <path> in commands
		for _, v := range c {
			command = append(command, repl.Replace(v))
		}
		// execute and write output
		cmd := exec.Command(command[0], command[1:]...)

		out, err := cmd.CombinedOutput()
		if err != nil {
			if cmd.ProcessState.ExitCode() == 1 {
				output.Write(out)
			}
			output.Write([]byte(err.Error()))
		} else {
			output.Write(out)
		}

		// update status code
		if err != nil {
			if cmd.ProcessState != nil {
				exitCode = cmd.ProcessState.ExitCode()
			} else {
				exitCode = 1 // non-zero
			}
		}
	}

	end := time.Now()

	return Result{
		Out:           output.String(),
		ExitCode:      exitCode,
		ExecutionTime: end.Sub(start),
	}
}
