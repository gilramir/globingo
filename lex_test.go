package globingo

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestLexSingle(c *C) {
	var tokens []tokenInterface
	var err error

	// plain text
	tokens, err = tokenizePattern("foo", UnixStyle)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Check(tokens[0].Type(), Equals, kTokenPlainText)

	// single char
	tokens, err = tokenizePattern("?", UnixStyle)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Check(tokens[0].Type(), Equals, kTokenSingleChar)

	// range
	tokens, err = tokenizePattern("[a-b]", UnixStyle)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Check(tokens[0].Type(), Equals, kTokenRange)

	// inverted range
	tokens, err = tokenizePattern("[^a-b]", UnixStyle)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Check(tokens[0].Type(), Equals, kTokenRange)

	// escaped chracter
	tokens, err = tokenizePattern("[?]", UnixStyle)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Check(tokens[0].Type(), Equals, kTokenPlainText)

	// multi-char single directory
	tokens, err = tokenizePattern("*", UnixStyle)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Check(tokens[0].Type(), Equals, kTokenMultiCharSingleDirectory)

	// multi-char multi directory
	tokens, err = tokenizePattern("**", UnixStyle)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Check(tokens[0].Type(), Equals, kTokenMultiCharMultiDirectory)
}

func (s *MySuite) TestLexCombineTextPlainSequences(c *C) {
	var tokens []tokenInterface
	var err error

	tokens, err = tokenizePattern("foo[?][[][]][*]bar", UnixStyle)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Check(tokens[0].Type(), Equals, kTokenPlainText)
	c.Check(tokens[0].(*tokenPlainText).text, Equals, "foo?[]*bar")
}

func (s *MySuite) TestLexBracketRange(c *C) {
	var tokens []tokenInterface
	var err error

	tokens, err = tokenizePattern("[[-]]", UnixStyle)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Check(tokens[0].Type(), Equals, kTokenRange)
	c.Check(tokens[0].(*tokenRange).from, Equals, '[')
	c.Check(tokens[0].(*tokenRange).to, Equals, ']')
}

func (s *MySuite) TestLexBracketErrors(c *C) {
	var err error

	_, err = tokenizePattern("[a-bxyz", UnixStyle)
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "Expected right bracket at position 5")

	_, err = tokenizePattern("[a-a]", UnixStyle)
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "The start and end of the range at 1 are the same")

	_, err = tokenizePattern("[b-a]", UnixStyle)
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "The start of the range ('b') at 1 is greater than the end of the range ('a')")

	_, err = tokenizePattern("[x]", UnixStyle)
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "Use of [] escape sequence not required for 'x' at position 1")

	_, err = tokenizePattern("[abcdef", UnixStyle)
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "Bracket syntax error (neither range nor escape) at position 1")

	_, err = tokenizePattern("[", UnixStyle)
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "Opening bracket at position 1 not terminated")

	_, err = tokenizePattern("[^", UnixStyle)
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "Opening bracket at position 1 not terminated")

	_, err = tokenizePattern("[a-", UnixStyle)
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "Opening bracket at position 1 not terminated")

	_, err = tokenizePattern("[a-b", UnixStyle)
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "Opening bracket at position 1 not terminated")
}
