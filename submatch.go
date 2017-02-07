package globingo

// One subMatch per token, maintaining a linked list.
type subMatch struct {
	startPos       int
	matchedPattern string
	next           *subMatch
}

// Takes a chain of subMatches and returns a single Match object
func (self *subMatch) combineIntoMatch() *Match {

	// How many in the linked list?
	numLinks := 1
	for next := self.next; next != nil; next = next.next {
		numLinks++
	}

	var lastPosition int
	matchedStrings := make([]string, numLinks)
	i := 0
	for this := self; this != nil; this = this.next {
		matchedStrings[i] = this.matchedPattern
		i++
		// Last in the chain?
		if i == numLinks {
			lastPosition = this.startPos + len(this.matchedPattern)
		}
	}

	return &Match{
		matchedStrings: matchedStrings,
		lastPosition:   lastPosition,
	}
}
