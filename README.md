# go-errors [![Travis Build Badge]][Travis Build] [![Go Report Card Badge]][Go Report Card] [![GoDoc Badge]][GoDoc]

This package aims to provide flexible general-purpose error handling functionality, with the 
following specific features in mind:

* **Error wrapping**: Allowing an error's cause to be preserved along with additional information.
* **Stack tracing**: Allowing the path taken to return an error to be easily identified.
* **Structured fields**: Allowing errors to be logged with additional contextual information.

This library has built upon the mistakes we've made and lessons we've learnt with regards to error
handling at Icelolly whilst working on our internal APIs. This library was inspired by approaches 
found elsewhere in the community, most notably the approach found in [Upspin][1].

## Usage

### Creating and Wrapping Errors

Errors may have a few different pieces of information attached to them; an `errors.Kind`, a message,
and fields. All of these things are optional, but at least an `errors.Kind` _or_ message must be 
given if using `errors.New`. Along with this information, file and line information will be added 
automatically. If you're wrapping an error, the only thing you must pass is an error to wrap as the
first argument:

```go
const ErrInvalidName errors.Kind = "auth: user's name is invalid"

func persistUser(user *User) error {
    if user.Name == "" {
        // Error kinds like `ErrInvalidName` can be used to react to different 
        // errors to decide how to handle them at the caller.
        return errors.New(ErrInvalidName)
    }
    
    err := persist(user)
    if err != nil {
        // Wrapping errors let's you add contextual information which may be
        // extracted again later (e.g. for logging).
        return errors.Wrap(err, "auth: failed to persist user in DB").
            WithField("user", user)
    }
    
    return nil
}
```

### Handling Errors

Most error handling is done using `errors.Is`, which checks if the given error is of a given 
`errors.Kind`. If the error doesn't have it's own `errors.Kind`, then `errors.Is` will look through
the errors stack until it finds an `errors.Kind`:

```go
func registerUserHandler(w http.ResponseWriter, r *http.Request) {
    user := // ...

    err := persistUser(user)
    switch {
    case errors.Is(err, ErrInvalidName):
        http.Error(w, "user has invalid username", 400)
        return
    case err != nil:
        http.Error(w, http.StatusText(500), 500)
        return
    }
    
    // ...
}
```

A more thorough example of usage can be found in the `example/` directory. It showcases creating 
errors, wrapping them, handling different kinds of errors, and dealing with things like logging.

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
