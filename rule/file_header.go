package rule

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/mgechev/revive/lint"
)

// FileHeaderRule lints the header that each file should have.
type FileHeaderRule struct {
	header string

	configureOnce sync.Once
}

var (
	multiRegexp  = regexp.MustCompile(`^/\*`)
	singleRegexp = regexp.MustCompile("^//")
)

func (r *FileHeaderRule) configure(arguments lint.Arguments) error {
	if len(arguments) < 1 {
		return nil
	}

	var ok bool
	r.header, ok = arguments[0].(string)
	if !ok {
		return fmt.Errorf(`invalid argument for "file-header" rule: argument should be a string, got %T`, arguments[0])
	}
	return nil
}

// Apply applies the rule to given file.
func (r *FileHeaderRule) Apply(file *lint.File, arguments lint.Arguments) []lint.Failure {
	var configureErr error
	r.configureOnce.Do(func() { configureErr = r.configure(arguments) })

	if configureErr != nil {
		return newInternalFailureError(configureErr)
	}

	if r.header == "" {
		return nil
	}

	failure := []lint.Failure{
		{
			Node:       file.AST,
			Confidence: 1,
			Failure:    "the file doesn't have an appropriate header",
		},
	}

	if len(file.AST.Comments) == 0 {
		return failure
	}

	g := file.AST.Comments[0]
	if g == nil {
		return failure
	}
	comment := ""
	for _, c := range g.List {
		text := c.Text
		if multiRegexp.MatchString(text) {
			text = text[2 : len(text)-2]
		} else if singleRegexp.MatchString(text) {
			text = text[2:]
		}
		comment += text
	}

	regex, err := regexp.Compile(r.header)
	if err != nil {
		return newInternalFailureError(err)
	}

	if !regex.MatchString(comment) {
		return failure
	}
	return nil
}

// Name returns the rule name.
func (*FileHeaderRule) Name() string {
	return "file-header"
}
