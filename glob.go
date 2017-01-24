// A glob module that all
package globingo

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Are we dealing with Unix-style or Windows-style paths? This affects
// the choice of directory separator character. NativeStyle chooses
// based on what system the program is currently running on.
type PathStyle int

const (
	NativeStyle PathStyle = iota
	WindowsStyle
	UnixStyle
)

type Glob struct {
	pattern   string
	style     PathStyle
	recursive bool
	tokens    []tokenInterface

	// Maps Nth wildcard to Mth token (well, N-1, since the slice is 0-indexed)
	wildcardPositions []int
}

// Return a new Glob object. The pattern is the glob pattern to use.
// The style indicates which type of directory separator characters to expect:
// NativeStyle, WindowsStyle, or UnixStyle.
// When recursive is true, '**' is allowed as a wildcard which matches any directory
// or filename, to any depth. When false, '**' is not allowed in a glob pattern.
// An error is returned when the glob pattern contains a syntax error.
func New(pattern string, style PathStyle, recursive bool) (*Glob, error) {
        if recursive {
                return nil, errors.New("recursive wildcard not yet implemented")
        }
	if !recursive && strings.Contains(pattern, "**") {
		return nil, errors.Errorf("Non-recursive glob pattern '%s' cannot contain '**'", pattern)
	}

	tokens, err := tokenizePattern(pattern, style)
	if err != nil {
		return nil, err
	}

	var wildcardPositions []int

	for tokenIndex, token := range tokens {
		if token.IsWildcard() {
			wildcardPositions = append(wildcardPositions, tokenIndex)
		}
	}

	return &Glob{
		pattern:           pattern,
		style:             style,
		recursive:         recursive,
		tokens:            tokens,
		wildcardPositions: wildcardPositions,
	}, nil
}

// Match the glob against the entire string given as 'haystack'.
func (self *Glob) Match(haystack string) *Match {
	return self.match(haystack, true)
}

// Match the glob against the beginning of string given as 'haystack'.
// The glob does not have to match the entire string.
func (self *Glob) StartsWith(haystack string) *Match {
	return self.match(haystack, false)
}

func (self *Glob) match(haystack string, matchCompleteString bool) *Match {
	var directorySeparator rune

	switch self.style {
        case NativeStyle:
                directorySeparator = kNativeDirectorySeparator
	case UnixStyle:
		directorySeparator = '/'
	case WindowsStyle:
		directorySeparator = '\\'
	default:
		panic(fmt.Sprintf("Unexpected style %q", self.style))
	}

	m := &Match{}

	pos := 0
	for _, token := range self.tokens {
		matched, pattern := token.Matches(haystack, pos, directorySeparator)
		if matched {
			m.matchedStrings = append(m.matchedStrings, pattern)
		} else {
			// How to handle recursive here?
			return nil
		}
		pos += len(pattern)
	}

	m.lastPosition = pos
	if matchCompleteString {
		// any leftover characters that didn't match?
		if pos != len(haystack) {
			return nil
		}
	}

	m.wildcardPositions = self.wildcardPositions
	return m
}

