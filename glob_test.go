package globingo

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestNew(c *C) {

	_, err := New("foo", UnixStyle, false)
	c.Assert(err, IsNil)

	_, err = New("foo*", UnixStyle, false)
	c.Assert(err, IsNil)

	_, err = New("foo**", UnixStyle, false)
	c.Assert(err, NotNil)

	_, err = New("foo**", UnixStyle, true)
	c.Assert(err, IsNil)
}

func (s *MySuite) TestMatchGetWildcardText(c *C) {
	glob, err := New("foo*/ba?/[a-z]/bar.c", UnixStyle, false)
	c.Assert(err, IsNil)

	match := glob.Match("fooze/bat/x/bar.c")
	c.Assert(match, NotNil)

	var text string
	text, err = match.GetWildcardText(1)
	c.Assert(err, IsNil)
	c.Check(text, Equals, "ze")
	text, err = match.GetWildcardText(2)
	c.Assert(err, IsNil)
	c.Check(text, Equals, "t")
	text, err = match.GetWildcardText(3)
	c.Assert(err, IsNil)
	c.Check(text, Equals, "x")

	// Ensure that '*' can match 0 characters
	match = glob.Match("foo/bat/x/bar.c")
	c.Assert(match, NotNil)

	text, err = match.GetWildcardText(1)
	c.Assert(err, IsNil)
	c.Check(text, Equals, "")
	text, err = match.GetWildcardText(2)
	c.Assert(err, IsNil)
	c.Check(text, Equals, "t")
	text, err = match.GetWildcardText(3)
	c.Assert(err, IsNil)
	c.Check(text, Equals, "x")
}

func (s *MySuite) TestStartsWith(c *C) {
	glob, err := New("foo*", UnixStyle, false)
	c.Assert(err, IsNil)

	haystack := "fooze/bat/x/bar.c"
	match := glob.StartsWith(haystack)
	c.Assert(match, NotNil)

	var text string
	text, err = match.GetWildcardText(1)
	c.Assert(err, IsNil)
	c.Check(text, Equals, "ze")

	c.Check(match.Length(), Equals, len("fooze"))
	c.Check(haystack[match.Length():], Equals, "/bat/x/bar.c")
}
