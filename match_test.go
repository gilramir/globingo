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

