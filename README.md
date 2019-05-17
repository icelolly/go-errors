# go-errors [![Travis Build Badge]][Travis Build] [![Go Report Card Badge]][Go Report Card] [![GoDoc Badge]][GoDoc]

This package aims to provide flexible general-purpose error handling functionality, with the 
following specific features in mind:

* **Error wrapping**: Allowing an error's cause to be preserved along with additional information.
* **Stack tracing**: Allowing the path taken to return an error to be easily identified.
* **Structured fields**: Allowing errors to be logged with additional contextual information.

This library has built upon the mistakes we've made and lessons we've learnt with regards to error
handling at Icelolly whilst working on our internal APIs. This library was inspired by approaches 
found elsewhere in the community, most notably the approach found in [Upspin][1].

## Example

An example of usage can be found in the `example/` directory. It showcases creating errors, wrapping
them, handling different kinds of errors, and dealing with things like logging.

## See More

* https://github.com/upspin/upspin/tree/master/errors
* https://middlemost.com/failure-is-your-domain/


[1]: https://github.com/upspin/upspin/blob/master/errors/errors.go#L23

[GoDoc]: https://godoc.org/github.com/icelolly/go-errors
[GoDoc Badge]: https://godoc.org/github.com/icelolly/go-errors?status.svg

[Go Report Card]: https://goreportcard.com/report/github.com/icelolly/go-errors
[Go Report Card Badge]: https://goreportcard.com/badge/github.com/icelolly/go-errors

[Travis Build]: https://travis-ci.org/icelolly/go-errors
[Travis Build Badge]: https://api.travis-ci.org/icelolly/go-errors.svg?branch=master
