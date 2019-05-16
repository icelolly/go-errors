# Notes

## Output

Errors are only used in a few places:

* Programmatically, i.e. checking the kind of an error. No output.
* In logs, i.e. when they have bubbled up through the application.
* Output to the user, i.e. how a web application might respond, or perhaps more likely with Go, in 
an API response.

Programmatically, we don't really have any issues. We are only looking for an error kind. This is a 
simple comparison through the stack.

In logs, we are by far more likely to use a message, a stack trace, and also log the fields of all
errors as structured log fields. This part is fine really, we'll have all of the information we need
and it should work fine in Kibana too. The message is a little bit pointless, but might make it 
easier to filter for certain top-level errors. By providing fields as a generic map, it also means
we have a consistent format for submitting structured errors as logs, no matter what information is
in the error.

Finally, errors that are output to the user. So, most of the time we can probably get away with some
kind of generic "an internal error occurred, contact someone @ wherever" kind of messages. What
about those times where we might want a more structured error format? Really we're talking about two
distinct concepts; errors, and error responses. Given that we have fields on errors, we should be
able to use that to build error responses if we need to. We can base our expectations surrounding
the fields that should be present on an error on the kind of error that is encountered. In other 
words, we could make an error kind like `deals: submission validation`, and along with that kind of 
error, always make sure we set a certain set of fields. For structs that need to be attached as 
fields we can still make use of the fields.

This does introduce another issue; what if we don't want to log all fields? Well then in that case,
they could always be removed. If we can expect certain fields to be there, we can expect them and 
then remove them too. Simple really.
