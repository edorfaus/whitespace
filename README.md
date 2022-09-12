# Whitespace

This repository holds some code for the [Whitespace][ws] esolang.

[ws]: https://en.wikipedia.org/wiki/Whitespace_(programming_language)

In addition to two interpreters written in Go, there's a test suite
(in the tests directory) that should work with most interpreters.

The top-level directory contains one of the interpreters, which uses a
fairly standard model of parsing the entire program up-front, then
executing it.

The "direct" directory contains the other interpreter, which does not
parse up-front, instead executing the code as it goes, and not keeping
it in memory (except the labels it has seen). It is intended for
handling exceptionally large programs (that wouldn't fit in memory),
and is best suited to programs that don't use a lot of loops since it
will re-read the instructions from the file on every iteration.
