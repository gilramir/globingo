# globingo
Glob library for Go (golang)

This handles globbing for strings that look like file paths.

Patterns:

    * - match any sequence of non-separator characters

    ? - match any single non-separator character

    [ - ] - a range of characters

    [^ - ] - negated range of characters

    To match any of the special characters, enclose in brackets, like: [?] or [[] or []]

Not yet implemented:
    ** - if recursive is set, match any files and zero or more directories and subdirectories.
        If ** is followed by a separator character, only directories and subdirectories match.
        If recursive is not set, this is an illegal character combination.


To use:

1. Create a new Glob object with the glob string:


```
import "github.com/gilramir/globingo"

glob, err := globingo.New("*.tar.gz", UnixStyle, false)
```

2. Use that Glob object to match a pattern, either with Match(), which matches the entire pattern,
or StartsWith(), which checks if the pattern starts with the glob.

```
match := glob.Match("foo.tar.gz")
```

3. The returned match object is nil if the match was not successful. A non-nil value means
the match was successful.

4. You can also use the Match object to replace the matched wildcards into a new string:

```
match := glob.Match("dirname/*")
newString, err := match.Replace("Filename: \\1")
```

