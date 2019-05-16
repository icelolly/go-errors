package main

import (
	"flag"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/icelolly/go-errors"
)

// Error kinds may be defined as constants.
const (
	ErrUserNotFound errors.Kind = "user not found"
	ErrUserInactive errors.Kind = "user is inactive"
)

// Usage example
// =============

func main() {
	var username string

	flag.StringVar(&username, "username", "", "Pick a user to look up in the \"database\".")
	flag.Parse()

	if username == "" {
		log.Fatalln("No username provided, exiting...")
	}

	// The approach to error handling here is consistent for any kind of error, as long as you
	// follow the conventions from this library. You can always convert an existing error into an
	// error that follows these conventions. Here's how you can handle different kinds of errors:

	user, err := getUser(username)
	switch {
	case errors.Is(err, ErrUserNotFound):
		log.Println("User not found:")
		printErrorFields(err)
	case errors.Is(err, ErrUserInactive):
		log.Println("User is inactive:")
		printErrorFields(err)
	case err != nil:
		// Unexpected error, let's blow up instead. You can get a stack from any chain of wrapped
		// errors from this library. The stack contains all error information, and can often work
		// well as something passed to a structured logger.
		spew.Dump(errors.Stack(err))

		// Or if you just want to output something and exit:
		errors.Fatal(err)
	}

	spew.Dump(user)
}

// printErrorFields is a utility for printing our the fields on an error. We've tried to make it
// easy to integrate this library with structured loggers by providing the `errors.Fields` and
// `errors.FieldsSlice` functions that easily allow you to attach fields to loggers.
func printErrorFields(err error) {
	for key, val := range errors.Fields(err) {
		log.Printf("- %q: %v", key, val)
	}

	os.Exit(1)
}

// Domain example
// ==============

// users is our "database" of users.
var users = map[string]User{
	"laureen":  {Name: "Laureen I. Eason", Email: "dilaureen8@yopmail.com", Active: true},
	"harrison": {Name: "Harrison F. Perreault", Email: "ifperreault5@yopmail.com", Active: false},
	"stephen":  {Name: "Stephen I. Hasty", Email: "distephen8@yopmail.com", Active: true},
}

// User is an example type, representing some potential user data.
type User struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Active bool   `json:"active"`
}

// getUser attempts to get a user from our "database".
func getUser(username string) (User, error) {
	user, ok := users[username]
	if !ok {
		return user, errors.New(ErrUserNotFound).WithField("username", username)
	}

	if !user.Active {
		return user, errors.New(ErrUserInactive).WithField("username", username)
	}

	// Simulate "critical" error occurring when trying to look up this particular user. For example,
	// maybe our database server has just died.
	if username == "stephen" {
		// Pretend for a moment that this error was returned from some third-party library, etc. It
		// can be a regular error, we only expect the standard Go `error` interface when wrapping.
		err := errors.New("database went down, oh no")

		// Wrap is identical to New, but must always take a non-nil Go `error` as it's first
		// parameter. That means you could create kinds to handle built-in "sentinel" errors.
		return user, errors.Wrap(err, "errors without a 'Kind' should probably always be handled").
			WithField("username", username)
	}

	return user, nil
}
