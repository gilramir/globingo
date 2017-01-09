package globingo

import (
	"github.com/pkg/errors"
)

type Match struct {

	// The items in this slice correlate exactly with the
	// items in the tokens slice in the Glob
	matchedStrings []string

	// Copied from the parent glob object
	wildcardPositions []int

	// Last position (used with StartsWWith)
	lastPosition int
}

func (self *Match) LastPosition() int {
	return self.lastPosition
}

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
