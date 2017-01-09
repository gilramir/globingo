package globingo

import (
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

type tokenInterface interface {
	Matches(string, int, rune) (bool, string)
	Type() tokenType
	IsWildcard() bool
}

// ============================================================================
// Match a specific string
// ============================================================================
type tokenPlainText struct {
	text string
}

func (self *tokenPlainText) Type() tokenType {
	return kTokenPlainText
}

func (self *tokenPlainText) IsWildcard() bool {
	return false
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

func (self *tokenSingleChar) IsWildcard() bool {
	return true
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

func (self *tokenRange) IsWildcard() bool {
	return true
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

func (self *tokenMultiCharSingleDirectory) IsWildcard() bool {
	return true
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
// TBD - Match multiple characters up to a any directory level
// ============================================================================

type tokenMultiCharMultiDirectory struct {
	text string
}

func (self *tokenMultiCharMultiDirectory) Type() tokenType {
	return kTokenMultiCharMultiDirectory
}

func (self *tokenMultiCharMultiDirectory) IsWildcard() bool {
	return true
}

func (self *tokenMultiCharMultiDirectory) Matches(haystack string, start int, directorySeparator rune) (bool, string) {
	if len(haystack) < start+len(self.text) {
		return false, ""
	}

	if haystack[start:start+len(self.text)] == self.text {
		return true, self.text
	}
	return false, ""
}
