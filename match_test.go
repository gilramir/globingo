package globingo

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestMatchPlainText(c *C) {
	glob, err := New("foo", UnixStyle, false)
	c.Assert(err, IsNil)

	match := glob.Match("foo")
	c.Assert(match, NotNil)
	c.Check(len(match.matchedStrings), Equals, 1)
	c.Check(match.matchedStrings[0], Equals, "foo")

	match = glob.Match("bar")
	c.Assert(match, IsNil)
}

func (s *MySuite) TestMatchPlainTextOptimized(c *C) {
	glob, err := New("foo[?][[][]]", UnixStyle, false)
	c.Assert(err, IsNil)

	match := glob.Match("foo?[]")
	c.Assert(match, NotNil)
	c.Check(len(match.matchedStrings), Equals, 1)
	c.Check(match.matchedStrings[0], Equals, "foo?[]")
}

func (s *MySuite) TestMatchSingleChar(c *C) {
	glob, err := New("foo?", UnixStyle, false)
	c.Assert(err, IsNil)

	match := glob.Match("foo")
	c.Assert(match, IsNil)

	match = glob.Match("fooz")
	c.Assert(match, NotNil)
	c.Check(len(match.matchedStrings), Equals, 2)
	c.Check(match.matchedStrings[0], Equals, "foo")
	c.Check(match.matchedStrings[1], Equals, "z")

	match = glob.Match("barz")
	c.Assert(match, IsNil)
}

func (s *MySuite) TestMatchRange(c *C) {
	glob, err := New("[a-z]?", UnixStyle, false)
	c.Assert(err, IsNil)

	match := glob.Match("a3")
	c.Assert(match, NotNil)
	c.Check(len(match.matchedStrings), Equals, 2)
	c.Check(match.matchedStrings[0], Equals, "a")
	c.Check(match.matchedStrings[1], Equals, "3")

	match = glob.Match("a")
	c.Assert(match, IsNil)

	match = glob.Match("A3")
	c.Assert(match, IsNil)

	match = glob.Match("A")
	c.Assert(match, IsNil)
}

func (s *MySuite) TestMatchInvertedRange(c *C) {
	glob, err := New("[^a-z]?", UnixStyle, false)
	c.Assert(err, IsNil)

	match := glob.Match("A3")
	c.Assert(match, NotNil)
	c.Check(len(match.matchedStrings), Equals, 2)
	c.Check(match.matchedStrings[0], Equals, "A")
	c.Check(match.matchedStrings[1], Equals, "3")

	match = glob.Match("A")
	c.Assert(match, IsNil)

	match = glob.Match("a3")
	c.Assert(match, IsNil)

	match = glob.Match("a")
	c.Assert(match, IsNil)
}

func (s *MySuite) TestMatchMultiChar(c *C) {
	glob, err := New("foo*", UnixStyle, false)
	c.Assert(err, IsNil)

	// '*' matches zero characters, too
	match := glob.Match("foo")
	c.Assert(match, NotNil)
	c.Check(len(match.matchedStrings), Equals, 2)
	c.Check(match.matchedStrings[0], Equals, "foo")
	c.Check(match.matchedStrings[1], Equals, "")

	// * only goes up to the directory separator
	match = glob.Match("foo/bar")
	c.Assert(match, IsNil)

	match = glob.Match("fooz")
	c.Assert(match, NotNil)
	c.Check(len(match.matchedStrings), Equals, 2)
	c.Check(match.matchedStrings[0], Equals, "foo")
	c.Check(match.matchedStrings[1], Equals, "z")

	match = glob.Match("fooxyz")
	c.Assert(match, NotNil)
	c.Check(len(match.matchedStrings), Equals, 2)
	c.Check(match.matchedStrings[0], Equals, "foo")
	c.Check(match.matchedStrings[1], Equals, "xyz")

	match = glob.Match("barz")
	c.Assert(match, IsNil)
}

func (s *MySuite) TestReplaceNoWildcard(c *C) {
	glob, err := New("foo", UnixStyle, false)
	c.Assert(err, IsNil)

	match := glob.Match("foo")
	c.Assert(match, NotNil)

	newString, err := match.Replace("nothing")
	c.Assert(err, IsNil)
	c.Check(newString, Equals, "nothing")

	newString, err = match.Replace("oops\\0")
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "Glob Match wildcard positions start at 1")

	newString, err = match.Replace("oops\\1")
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "This Glob Match has 0 wildcards; #1 was requested")

	newString, err = match.Replace("oops\\")
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "\\ should be followed by number or \\ at the end of the string")
}

func (s *MySuite) TestReplace(c *C) {
	glob, err := New("foo-?/*", UnixStyle, false)
	c.Assert(err, IsNil)

	match := glob.Match("foo-a/bar")
	c.Assert(match, NotNil)

	newString, err := match.Replace("\\2/\\1")
	c.Assert(err, IsNil)
	c.Check(newString, Equals, "bar/a")

	newString, err = match.Replace("\\1\\2\\1")
	c.Assert(err, IsNil)
	c.Check(newString, Equals, "abara")

}

func (s *MySuite) TestReplaceMoreThan9(c *C) {
	glob, err := New("????????????", UnixStyle, false)
	c.Assert(err, IsNil)

	match := glob.Match("ABCDEFGHIJKL")
	c.Assert(match, NotNil)

	var newString string
	newString, err = match.Replace("\\1")
	c.Assert(err, IsNil)
	c.Check(newString, Equals, "A")

	newString, err = match.Replace("\\2")
	c.Assert(err, IsNil)
	c.Check(newString, Equals, "B")

	newString, err = match.Replace("\\3")
	c.Assert(err, IsNil)
	c.Check(newString, Equals, "C")

	newString, err = match.Replace("\\10")
	c.Assert(err, IsNil)
	c.Check(newString, Equals, "J")

	newString, err = match.Replace("\\11")
	c.Assert(err, IsNil)
	c.Check(newString, Equals, "K")

	newString, err = match.Replace("\\12")
	c.Assert(err, IsNil)
	c.Check(newString, Equals, "L")
}
