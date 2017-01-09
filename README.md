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
