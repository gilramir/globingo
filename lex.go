package globingo

import (
	"fmt"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// lexing ideas taken from:
// https://github.com/golang/go/blob/master/src/text/template/parse/lex.go

type lexerState struct {
	input  string // the string being scanned
	start  int    // the start position of this token
	pos    int    // current position within the input
	width  int    // width of the last rune read (ASCII/UTF)
	tokens []tokenInterface
	err    error
}

// returns the next rune in the input
func (l *lexerState) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

// returns but does not consume the next rune in the input
func (l *lexerState) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// step back one rune. Can only be called once per call of next()
func (l *lexerState) backup() {
	l.pos -= l.width
}

// returns the current text
func (l *lexerState) currentText() string {
	return l.input[l.start:l.pos]
}

// adds a new token and resets the counters
func (l *lexerState) addToken(token tokenInterface) {
	l.tokens = append(l.tokens, token)
	l.start = l.pos
}

// skips over the pending input before this point
func (l *lexerState) ignore() {
	l.start = l.pos
}

// rune position (starts at 1)
func (l *lexerState) currentPosition() int {
	return l.pos + 1
}

// sets the error and terminates the scan
func (l *lexerState) errorf(format string, args ...interface{}) stateFunc {
	l.err = errors.Errorf(format, args...)
	return nil
}

type stateFunc func(*lexerState) stateFunc

const eof = -1

func tokenizePattern(pattern string, style PathStyle) ([]tokenInterface, error) {

	var lexer lexerState

	lexer.input = pattern

	var state stateFunc
	for state = lexAnything; state != nil; {
		state = state(&lexer)
	}

	if lexer.err != nil {
		return nil, lexer.err
	}

	// The lexing process may have created multiple tokenPlainText sequences.
	// If so, combine them into single tokenPlainTexts
	optimizedTokens := make([]tokenInterface, 0, len(lexer.tokens))

	for _, token := range lexer.tokens {
		if len(optimizedTokens) == 0 {
			optimizedTokens = append(optimizedTokens, token)
		} else {
			if token.Type() == kTokenPlainText && optimizedTokens[len(optimizedTokens)-1].Type() == kTokenPlainText {
				optimizedTokens[len(optimizedTokens)-1].(*tokenPlainText).text +=
					token.(*tokenPlainText).text
			} else {
				optimizedTokens = append(optimizedTokens, token)
			}
		}
	}

	return optimizedTokens, nil
}

// is it a special character that is or starts a wildcard?
func isWildcardStart(r rune) bool {
	return r == '?' || r == '*' || r == '['
}

// is it any wildcarcd character at all (and thus, can be escape?)
func isAnyWildcard(r rune) bool {
	return r == '?' || r == '*' || r == '[' || r == ']'
}

func lexAnything(l *lexerState) stateFunc {
	var r rune

	r = l.peek()
	if isWildcardStart(r) {
		return lexWildcardStart
	} else if r == eof {
		return nil
	} else {
		return lexPlainText
	}
}

func lexPlainText(l *lexerState) stateFunc {
	consumed := false
	for {
		r := l.next()
		if isWildcardStart(r) || r == eof {
			l.backup()
			if consumed {
				l.addToken(&tokenPlainText{
					text: l.currentText(),
				})
			}
			if r == eof {
				return nil
			} else {
				return lexWildcardStart
			}
		}
		consumed = true
	}
}

func lexWildcardStart(l *lexerState) stateFunc {
	r := l.next()

	switch r {
	case '*':
		nextRune := l.next()
		if nextRune == '*' {
			l.addToken(&tokenMultiCharMultiDirectory{})
		} else {
			l.backup()
			l.addToken(&tokenMultiCharSingleDirectory{})
		}
		return lexAnything
	case '?':
		l.addToken(&tokenSingleChar{})
		return lexAnything
	case '[':
		return lexBracketStart
	default:
		panic(fmt.Sprintf("Unexpected rune: '%v'", r))
	}
}

// Range: [a-b] or []-}]
// Inverted range: [^a-b]
// Escape single character [?] or [[] or []]
func lexBracketStart(l *lexerState) stateFunc {
	var inverted bool

	// The position of '['
	startPos := l.currentPosition() - 1

	firstRune := l.next()
	if firstRune == '^' {
		inverted = true
		firstRune = l.next()

	} else if firstRune == eof {
		return l.errorf("Opening bracket at position %d not terminated", startPos)
	}

	separator := l.next()
	if separator == '-' {
		// we're in a range
		secondRune := l.next()
		rightBracket := l.next()
		if rightBracket != ']' {
			if rightBracket == eof {
				return l.errorf("Opening bracket at position %d not terminated", startPos)
			}
			return l.errorf("Expected right bracket at position %d", l.currentPosition()-1)
		}
		if firstRune == secondRune {
			return l.errorf("The start and end of the range at %d are the same", startPos)
		}
		if firstRune > secondRune {
			return l.errorf("The start of the range (%q) at %d is greater than the end of the range (%q)",
				firstRune, startPos, secondRune)
		}
		l.addToken(&tokenRange{
			from:     firstRune,
			to:       secondRune,
			inverted: inverted,
		})
		return lexAnything

	} else if separator == ']' {
		// we're in an escape sequence
		if isAnyWildcard(firstRune) {
			// if previous token is also PlainText, we're creating a sequence of plain text tokens.
			// After parsing, the calling routine will optimize these sequences into a single
			// tokenPlainText
			l.addToken(&tokenPlainText{
				text: string(firstRune),
			})
			return lexAnything
		} else {
			return l.errorf("Use of [] escape sequence not required for %q at position %d",
				firstRune, startPos)
		}

	} else if separator == eof {
		return l.errorf("Opening bracket at position %d not terminated", startPos)

	} else {
		// A syntax error
		return l.errorf("Bracket syntax error (neither range nor escape) at position %d", startPos)
	}
}
