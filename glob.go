// A glob module that also provides access to which strings matched
// each wildcard.
package globingo

import (
	"context"
	"fmt"
	//        "log"
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

const (
	kWindowsStyle = '\\'
	kUnixStyle    = '/'
)

type Glob struct {
	pattern                     string
	directorySeparator          rune
	recursiveAllowed            bool
	tokens                      []tokenInterface
	hasTokenWithMultipleAnswers bool

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
	if !recursive && strings.Contains(pattern, "**") {
		return nil, errors.Errorf("Non-recursive glob pattern '%s' cannot contain '**'", pattern)
	}

	var directorySeparator rune
	switch style {
	case NativeStyle:
		directorySeparator = kNativeDirectorySeparator
	case UnixStyle:
		directorySeparator = '/'
	case WindowsStyle:
		directorySeparator = '\\'
	default:
		panic(fmt.Sprintf("Unexpected style %q", style))
	}

	tokens, err := tokenizePattern(pattern, directorySeparator)
	if err != nil {
		return nil, err
	}

	var wildcardPositions []int

	hasTokenWithMultipleAnswers := false
	for tokenIndex, token := range tokens {
		if token.IsWildcard() {
			wildcardPositions = append(wildcardPositions, tokenIndex)
		}
		if token.CanHaveMultipleAnswers() {
			hasTokenWithMultipleAnswers = true
		}
	}

	return &Glob{
		pattern:                     pattern,
		directorySeparator:          directorySeparator,
		recursiveAllowed:            recursive,
		hasTokenWithMultipleAnswers: hasTokenWithMultipleAnswers,
		tokens:            tokens,
		wildcardPositions: wildcardPositions,
	}, nil
}

// Returns the number of wildcard patterns. Useful for Match.GetWildcardText()
func (self *Glob) NumWildcards() int {
	return len(self.wildcardPositions)
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
	var m *Match

	if self.hasTokenWithMultipleAnswers {
		subMatch := self._matchRecursive(0, 0, haystack, matchCompleteString)
		if subMatch != nil {
			m = subMatch.combineIntoMatch()
		} else {
			return nil
		}
	} else {
		m = self._matchNoRecursive(haystack, matchCompleteString)
	}

	if m != nil {
		if matchCompleteString {
			// any leftover characters that didn't match?
			if m.lastPosition != len(haystack) {
				return nil
			}
		}
		m.wildcardPositions = self.wildcardPositions
	}
	return m
}

func (self *Glob) _matchRecursive(tokenIndex int, pos int, haystack string, matchCompleteString bool) *subMatch {
	token := self.tokens[tokenIndex]
	//        log.Printf("_matchRecursive %s tokenIndex=%d pos=%d haystack=%s", token.String(), tokenIndex, pos, haystack)

	// Is this the last token? There's no more multiple recursion paths, so just Match.
	if tokenIndex == len(self.tokens)-1 {
		matched, pattern := token.Matches(haystack, pos, self.directorySeparator)
		//            log.Printf("Last token; matched=%v pattern=%s", matched, pattern)
		if matched {
			return &subMatch{
				startPos:       pos,
				matchedPattern: pattern,
			}
		} else {
			return nil
		}
	}

	// Not the last token, so we need to check for multiple possibilities
	// given a recursive token
	if token.CanHaveMultipleAnswers() {
		//            log.Printf("AllMatchedPatterns of '%s' pos=%d", haystack, pos)
		if token.CanMatchZeroCharacters() {
			// Try it as if the pattern matched ""
			nextSubMatch := self._matchRecursive(tokenIndex+1, pos, haystack, matchCompleteString)
			if nextSubMatch != nil {
				return &subMatch{
					startPos:       pos,
					matchedPattern: "",
					next:           nextSubMatch,
				}
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		for pattern := range token.AllMatchedPatterns(ctx, haystack, pos, self.directorySeparator) {
			//                log.Printf("Got pattern: %s", pattern)
			nextSubMatch := self._matchRecursive(tokenIndex+1, pos+len(pattern), haystack, matchCompleteString)
			if nextSubMatch != nil {
				// Cancel the recursion
				cancel()
				return &subMatch{
					startPos:       pos,
					matchedPattern: pattern,
					next:           nextSubMatch,
				}
			}
		}
		// No match.
		return nil
	} else {
		// It wasn't a recursive token, so just Match
		matched, pattern := token.Matches(haystack, pos, self.directorySeparator)
		//            log.Printf("Not-last token; matched=%v pattern=%s", matched, pattern)
		if matched {
			nextSubMatch := self._matchRecursive(tokenIndex+1, pos+len(pattern), haystack, matchCompleteString)
			return &subMatch{
				startPos:       pos,
				matchedPattern: pattern,
				next:           nextSubMatch,
			}
		} else {
			return nil
		}
	}
	// Cannot reach here
}

func (self *Glob) _matchNoRecursive(haystack string, matchCompleteString bool) *Match {
	m := &Match{}
	pos := 0
	for _, token := range self.tokens {
		matched, pattern := token.Matches(haystack, pos, self.directorySeparator)
		if matched {
			m.matchedStrings = append(m.matchedStrings, pattern)
		} else {
			// No recursive at all in this function, so we can return now.
			return nil
		}
		pos += len(pattern)
	}

	m.lastPosition = pos
	return m
}
