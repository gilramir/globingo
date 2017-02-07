package globingo

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

type tokenType int

const (
	kTokenPlainText tokenType = iota
	kTokenSingleChar
	kTokenRange
	kTokenMultiCharSingleDirectory
	kTokenMultiCharMultiDirectory
)

// Note: can lowercase these functions
type tokenInterface interface {
	Matches(string, int, rune) (bool, string)
	Type() tokenType
	IsWildcard() bool
	CanHaveMultipleAnswers() bool
	CanMatchZeroCharacters() bool
	String() string
	// Only for recursive tokens
	AllMatchedPatterns(ctx context.Context, haystack string, start int, directorySeparator rune) chan string
}

// ============================================================================
// Match a specific string
// ============================================================================
type tokenPlainText struct {
	text string
}

func (self *tokenPlainText) String() string {
	return fmt.Sprintf("PlainText: '%s'", self.text)
}

func (self *tokenPlainText) Type() tokenType {
	return kTokenPlainText
}

func (self *tokenPlainText) IsWildcard() bool {
	return false
}

func (self *tokenPlainText) CanHaveMultipleAnswers() bool {
	return false
}

func (self *tokenPlainText) CanMatchZeroCharacters() bool {
	return false
}

func (self *tokenPlainText) AllMatchedPatterns(ctx context.Context, haystack string, start int, directorySeparator rune) chan string {
	panic("should not be called")
	return nil
}

func (self *tokenPlainText) Matches(haystack string, start int, directorySeparator rune) (bool, string) {
	if len(haystack) < start+len(self.text) {
		return false, ""
	}

	if haystack[start:start+len(self.text)] == self.text {
		return true, self.text
	}
	return false, ""
}

// ============================================================================
// Match any single character
// ============================================================================
type tokenSingleChar struct{}

func (self *tokenSingleChar) Type() tokenType {
	return kTokenSingleChar
}

func (self *tokenSingleChar) String() string {
	return "?"
}

func (self *tokenSingleChar) IsWildcard() bool {
	return true
}

func (self *tokenSingleChar) CanHaveMultipleAnswers() bool {
	return false
}

func (self *tokenSingleChar) CanMatchZeroCharacters() bool {
	return false
}

func (self *tokenSingleChar) AllMatchedPatterns(ctx context.Context, haystack string, start int, directorySeparator rune) chan string {
	panic("should not be called")
	return nil
}

func (self *tokenSingleChar) Matches(haystack string, start int, directorySeparator rune) (bool, string) {
	if len(haystack) < start+1 {
		return false, ""
	}

	return true, haystack[start : start+1]
}

// ============================================================================
// Match any one character within a range of runes
// ============================================================================
type tokenRange struct {
	from     rune
	to       rune
	inverted bool
}

func (self *tokenRange) Type() tokenType {
	return kTokenRange
}

func (self *tokenRange) String() string {
	if self.inverted {
		return fmt.Sprintf("[^%s-%s]", self.from, self.to)
	} else {
		return fmt.Sprintf("[%s-%s]", self.from, self.to)
	}
}

func (self *tokenRange) IsWildcard() bool {
	return true
}

func (self *tokenRange) CanHaveMultipleAnswers() bool {
	return false
}

func (self *tokenRange) CanMatchZeroCharacters() bool {
	return false
}

func (self *tokenRange) AllMatchedPatterns(ctx context.Context, haystack string, start int, directorySeparator rune) chan string {
	panic("should not be called")
	return nil
}

func (self *tokenRange) Matches(haystack string, start int, directorySeparator rune) (bool, string) {
	if len(haystack) < start+1 {
		return false, ""
	}

	//r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	r, _ := utf8.DecodeRuneInString(haystack[start:])

	if self.inverted {
		if r >= self.from && r <= self.to {
			return false, ""
		} else {
			return true, string(r)
		}
	} else {
		if r >= self.from && r <= self.to {
			return true, string(r)
		} else {
			return false, ""
		}
	}
}

// ============================================================================
// Match multiple characters up to a single directory level
// ============================================================================

type tokenMultiCharSingleDirectory struct{}

func (self *tokenMultiCharSingleDirectory) Type() tokenType {
	return kTokenMultiCharSingleDirectory
}

func (self *tokenMultiCharSingleDirectory) String() string {
	return "*"
}

func (self *tokenMultiCharSingleDirectory) IsWildcard() bool {
	return true
}

func (self *tokenMultiCharSingleDirectory) CanHaveMultipleAnswers() bool {
	return true
}

func (self *tokenMultiCharSingleDirectory) CanMatchZeroCharacters() bool {
	return true
}

func (self *tokenMultiCharSingleDirectory) AllMatchedPatterns(ctx context.Context, haystack string, start int, directorySeparator rune) chan string {
	resultChan := make(chan string)

	go func() {
		defer close(resultChan)

		var result string

		if len(haystack) == start {
			resultChan <- ""
			return
		} else if len(haystack) < start {
			return
		}

		r, _ := utf8.DecodeRuneInString(haystack[start:])
		// The directory separator character is our delimiter
		if r == directorySeparator {
			return
		}

		var pos int
		for pos = start; ; {
			r, w := utf8.DecodeRuneInString(haystack[pos:])
			if w == 0 || r == directorySeparator {
				break
			}
			pos += w
			result += string(r)
			resultChan <- result
		}
	}()

	return resultChan
}

func (self *tokenMultiCharSingleDirectory) Matches(haystack string, start int, directorySeparator rune) (bool, string) {
	if len(haystack) == start {
		return true, ""

	} else if len(haystack) < start {
		return false, ""
	}

	r, _ := utf8.DecodeRuneInString(haystack[start:])
	// The directory separator character is our delimiter
	if r == directorySeparator {
		return true, ""
	}

	var pos int
	for pos = start; ; {
		r, w := utf8.DecodeRuneInString(haystack[pos:])
		if w == 0 || r == directorySeparator {
			break
		}
		pos += w
	}

	return true, haystack[start:pos]
}

// ============================================================================
// Match multiple characters up to a any directory level
// ============================================================================

type tokenMultiCharMultiDirectory struct {
	directoriesOnly bool
}

func (self *tokenMultiCharMultiDirectory) Type() tokenType {
	return kTokenMultiCharMultiDirectory
}

func (self *tokenMultiCharMultiDirectory) String() string {
	return "**"
}

func (self *tokenMultiCharMultiDirectory) IsWildcard() bool {
	return true
}

func (self *tokenMultiCharMultiDirectory) CanHaveMultipleAnswers() bool {
	return true
}

func (self *tokenMultiCharMultiDirectory) CanMatchZeroCharacters() bool {
	return false
}

func (self *tokenMultiCharMultiDirectory) AllMatchedPatterns(ctx context.Context, haystack string, start int, directorySeparator rune) chan string {
	resultChan := make(chan string)

	go func() {
		defer close(resultChan)

		if self.directoriesOnly {
			parts := strings.Split(haystack[start:], string(directorySeparator))

			// We need to discard the last item in the Split(), because it won't be
			// a directory (it will be a file). Even if haystack ends with directorySeparator,
			// the last string in the Split() will be "", which also needs to be discarded.
			// But, if there was no directorySeparator at all, then there's only one part.
			if len(parts) > 1 {
				parts = parts[0 : len(parts)-1]
			}

			var result string
			var prevResult string

			for _, part := range parts {
				if prevResult == "" {
					result = part
				} else {
					result = prevResult + string(directorySeparator) + part
				}
				resultChan <- result
				prevResult = result

				// Canceled?
				select {
				case <-ctx.Done():
					return
				default:
					// no-op
				}
			}
		} else {
			panic("Not yet implemented")
		}
	}()

	return resultChan
}

func (self *tokenMultiCharMultiDirectory) Matches(haystack string, start int, directorySeparator rune) (bool, string) {
	if len(haystack) == start {
		return true, ""

	} else if len(haystack) < start {
		return false, ""
	}

	//r, _ := utf8.DecodeRuneInString(haystack[start:])
	/*
		// The directory separator character is our delimiter
		if r == directorySeparator {
			return true, ""
		}*/

	var pos int
	for pos = start; ; {
		//r, w := utf8.DecodeRuneInString(haystack[pos:])
		_, w := utf8.DecodeRuneInString(haystack[pos:])
		if w == 0 /*|| r == directorySeparator */ {
			break
		}
		pos += w
	}

	return true, haystack[start:pos]
}
