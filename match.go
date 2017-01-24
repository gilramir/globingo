package globingo

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

type Match struct {

	// The items in this slice correlate exactly with the
	// items in the tokens slice in the Glob
	matchedStrings []string

	// Copied from the parent glob object
	wildcardPositions []int

	// Last position (used with StartsWith)
	lastPosition int
}

// Returns the last position that was matched.
func (self *Match) LastPosition() int {
	return self.lastPosition
}

// Return the string that the Nth wildcard matched.
func (self *Match) GetWildcardText(n int) (string, error) {
	if n == 0 {
		return "", errors.New("Glob Match wildcard positions start at 1")
	}
	if n > len(self.wildcardPositions) {
		return "", errors.Errorf("This Glob Match has %d wildcards; #%d was requested",
			len(self.wildcardPositions), n)
	}

	return self.matchedStrings[self.wildcardPositions[n-1]], nil
}

const (
	kAnything      = 1
	kFirstEscape   = 2
	kNumericEscape = 3
)

// Using a match, return a sting with tokens like \1, \2, replaced
// with the patterns matched from the glob.
func (self *Match) Replace(pattern string) (string, error) {

	var result string
	var state int = kAnything
	var numericText string

	for i, runeValue := range pattern {
		switch state {
		case kAnything:
			if runeValue == '\\' {
				state = kFirstEscape
				numericText = ""
			} else {
				result += string(runeValue)
			}

		case kFirstEscape:
			if '0' <= runeValue && runeValue <= '9' {
				numericText += string(runeValue)
				state = kNumericEscape
			} else if runeValue == '\\' {
				result += "\\"
				state = kAnything
			} else {
				return "", errors.Errorf("\\ should be followed by number or \\, not %s at position %d",
					string(runeValue), i+1)
			}

		case kNumericEscape:
			if '0' <= runeValue && runeValue <= '9' {
				numericText += string(runeValue)
				state = kNumericEscape
			} else {
				// Convert the number and find that Nth pattern
				n, err := strconv.Atoi(numericText)
				if err != nil {
					panic(fmt.Sprintf("Escaped value %s should be a number but it's not", numericText))
				}
				matchedText, err := self.GetWildcardText(n)
				if err != nil {
					return "", err
				}
				result += matchedText
				numericText = ""
				if runeValue == '\\' {
					state = kFirstEscape
				} else {
					result += string(runeValue)
					state = kAnything
				}
			}

		default:
			panic(fmt.Sprintf("state = %d", state))
		}
	}

	// Reached the end of the string. Was the \\ pattern at the end of the string?
	switch state {
	case kAnything:
		// nothing to do

	case kFirstEscape:
		return "", errors.Errorf("\\ should be followed by number or \\ at the end of the string")

	case kNumericEscape:
		// Convert the number and find that Nth pattern
		n, err := strconv.Atoi(numericText)
		if err != nil {
			panic(fmt.Sprintf("Escaped value %s should be a number but it's not", numericText))
		}
		matchedText, err := self.GetWildcardText(n)
		if err != nil {
			return "", err
		}
		result += matchedText

	default:
		panic(fmt.Sprintf("state = %d", state))
	}

	return result, nil
}
