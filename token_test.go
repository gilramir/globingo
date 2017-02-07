package globingo

import (
	"context"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestTokenPlainText(c *C) {

	token := &tokenPlainText{
		text: "foo",
	}

	var m bool
	var t string

	// exactly the same
	m, t = token.Matches("foo", 0, '/')
	c.Check(m, Equals, true)
	c.Check(t, Equals, "foo")

	// not enough characters to match
	m, t = token.Matches("fo", 0, '/')
	c.Check(m, Equals, false)

	// substring at pos 0
	m, t = token.Matches("foodle", 0, '/')
	c.Check(m, Equals, true)
	c.Check(t, Equals, "foo")

	// substring in the middle
	m, t = token.Matches("trefooblah", 0, '/')
	c.Check(m, Equals, false)

	m, t = token.Matches("trefooblah", 3, '/')
	c.Check(m, Equals, true)
	c.Check(t, Equals, "foo")

	// substring at the end
	m, t = token.Matches("trefoo", 0, '/')
	c.Check(m, Equals, false)

	m, t = token.Matches("trefoo", 3, '/')
	c.Check(m, Equals, true)
	c.Check(t, Equals, "foo")
}

func (s *MySuite) TestTokenSingleChar(c *C) {

	token := &tokenSingleChar{}

	var m bool
	var t string

	// first character
	m, t = token.Matches("bar", 0, '/')
	c.Check(m, Equals, true)
	c.Check(t, Equals, "b")

	// middle character
	m, t = token.Matches("bar", 1, '/')
	c.Check(m, Equals, true)
	c.Check(t, Equals, "a")

	// last character
	m, t = token.Matches("bar", 2, '/')
	c.Check(m, Equals, true)
	c.Check(t, Equals, "r")

	// empty string
	m, t = token.Matches("", 0, '/')
	c.Check(m, Equals, false)
}

func (s *MySuite) TestTokenRangeNormal(c *C) {
	token := &tokenRange{
		from: 'A',
		to:   'C',
	}

	var m bool
	var t string

	// first character
	m, t = token.Matches("bar", 0, '/')
	c.Check(m, Equals, false)

	m, t = token.Matches("Bar", 0, '/')
	c.Check(m, Equals, true)
	c.Check(t, Equals, "B")

	// middle character
	m, t = token.Matches("bar", 1, '/')
	c.Check(m, Equals, false)

	m, t = token.Matches("bAr", 1, '/')
	c.Check(m, Equals, true)
	c.Check(t, Equals, "A")

	// last character
	m, t = token.Matches("bac", 2, '/')
	c.Check(m, Equals, false)

	m, t = token.Matches("baC", 2, '/')
	c.Check(m, Equals, true)
	c.Check(t, Equals, "C")

	// empty string
	m, t = token.Matches("", 0, '/')
	c.Check(m, Equals, false)
}

func (s *MySuite) TestTokenRangeInverted(c *C) {
	token := &tokenRange{
		from:     'A',
		to:       'C',
		inverted: true,
	}

	var m bool
	var t string

	// first character
	m, t = token.Matches("bar", 0, '/')
	c.Check(m, Equals, true)
	c.Check(t, Equals, "b")

	m, t = token.Matches("Bar", 0, '/')
	c.Check(m, Equals, false)

	// middle character
	m, t = token.Matches("bar", 1, '/')
	c.Check(m, Equals, true)
	c.Check(t, Equals, "a")

	m, t = token.Matches("bAr", 1, '/')
	c.Check(m, Equals, false)

	// last character
	m, t = token.Matches("bac", 2, '/')
	c.Check(m, Equals, true)
	c.Check(t, Equals, "c")

	m, t = token.Matches("baC", 2, '/')
	c.Check(m, Equals, false)

	// empty string
	m, t = token.Matches("", 0, '/')
	c.Check(m, Equals, false)
}

func (s *MySuite) TestTokenSingleDirAllMatchedPatterns1(c *C) {
	// We pretend the glob is *.c
	token := &tokenMultiCharSingleDirectory{}

	haystack := "foo.c"

	ctx, _ := context.WithCancel(context.Background())
	var patterns []string
	for pattern := range token.AllMatchedPatterns(ctx, haystack, 0, '/') {
		patterns = append(patterns, pattern)
	}

	c.Assert(len(patterns), Equals, 5)
	c.Check(patterns[0], Equals, "f")
	c.Check(patterns[1], Equals, "fo")
	c.Check(patterns[2], Equals, "foo")
	c.Check(patterns[3], Equals, "foo.")
	c.Check(patterns[4], Equals, "foo.c")
}

func (s *MySuite) TestTokenSingleDirAllMatchedPatterns2(c *C) {
	// We pretend the glob is */foo.c
	token := &tokenMultiCharSingleDirectory{}

	haystack := "bar/foo.c"

	ctx, _ := context.WithCancel(context.Background())
	var patterns []string
	for pattern := range token.AllMatchedPatterns(ctx, haystack, 0, '/') {
		patterns = append(patterns, pattern)
	}

	c.Assert(len(patterns), Equals, 3)
	c.Check(patterns[0], Equals, "b")
	c.Check(patterns[1], Equals, "ba")
	c.Check(patterns[2], Equals, "bar")
}

func (s *MySuite) TestTokenMultiDirAllMatchedPatterns1(c *C) {
	// We pretend the glob is a/**/foo.c
	token := &tokenMultiCharMultiDirectory{
		directoriesOnly: true,
	}

	haystack := "a/bb/ccc/ddd/eee/foo.c"

	ctx, _ := context.WithCancel(context.Background())
	var patterns []string
	for pattern := range token.AllMatchedPatterns(ctx, haystack, 2, '/') {
		patterns = append(patterns, pattern)
	}

	c.Assert(len(patterns), Equals, 4)
	c.Check(patterns[0], Equals, "bb")
	c.Check(patterns[1], Equals, "bb/ccc")
	c.Check(patterns[2], Equals, "bb/ccc/ddd")
	c.Check(patterns[3], Equals, "bb/ccc/ddd/eee")
}

func (s *MySuite) TestTokenMultiDirAllMatchedPatterns2(c *C) {
	// We pretend the glob is a/**/foo.c
	token := &tokenMultiCharMultiDirectory{
		directoriesOnly: true,
	}

	haystack := "a/bb/ccc/ddd/eee/"

	ctx, _ := context.WithCancel(context.Background())
	var patterns []string
	for pattern := range token.AllMatchedPatterns(ctx, haystack, 2, '/') {
		patterns = append(patterns, pattern)
	}

	c.Assert(len(patterns), Equals, 4)
	c.Check(patterns[0], Equals, "bb")
	c.Check(patterns[1], Equals, "bb/ccc")
	c.Check(patterns[2], Equals, "bb/ccc/ddd")
	c.Check(patterns[3], Equals, "bb/ccc/ddd/eee")
}
