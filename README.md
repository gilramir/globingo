# globingo
Glob library for Go (golang)

This handles globbing for strings that look like file paths.

Patterns:

    * - match any sequence of non-separator characters, including the empty-string

    ? - match any single non-separator character

    [ - ] - a range of characters

    [^ - ] - negated range of characters

    To match any of the special characters, enclose in brackets, like: [?] or [[] or []]

    ** - if recursive is true when New is called, match any files and zero or more directories
        and subdirectories.  If ** is followed by a separator character, only directories and
        subdirectories match.  If recursive is not set, this is an illegal character combination.

API docs at: [godoc.org](https://godoc.org/github.com/gilramir/globingo "GoDoc")

To use:

Create a new Glob object with the glob string:
```
import "github.com/gilramir/globingo"

glob, err := globingo.New("*.tar.gz", NativeStyle, false)
```

Use that Glob object to match a pattern, either with Match(), which matches the entire pattern,
or StartsWith(), which checks if the pattern starts with the glob.
```
match := glob.Match("foo.tar.gz")
```

The returned match object is nil if the match was not successful. A non-nil value means
the match was successful.


Alternatively, you the Glob object provides a StartsWith() method which returns a
Match object which only has to match the beginning of the string.
```
glob, err := globingo.New("foo/*", UnixStyle, false)

shouldBeNil := glob.StartsWith("foo")
positiveMatch := glob.StartsWith("foo/bar/baz")
```

In the above example, in the case of the positive match, the glob will match "foo/bar", since "\*"
only matches up to the directory separator character. To find out how much of the
string was matched, the Match object provides a Length() method.

With the Match object, you can also replace the matched wildcards into a new string.

```
glob, err := globingo.New("foo/*", NativeStyle, false)
match := glob.Match("foo/bar")

// Get the "*" text with GetWildcardText
subDir, err := match.GetWildcardText(1)

// Or create a new string by replacing wildcards via "\\n", where n is the Nth wildcard,
// starting a 1
subDir, err := match.Replace("bar/\\1")
```


