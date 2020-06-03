# eris ![minecraft golden apple](https://cdn.emojidex.com/emoji/hdpi/minecraft_golden_apple.png?1511637499 "minecraft golden apple")

Package eris provides a better way to handle, trace, and log errors in Go. This package is inspired by a few existing packages: [xerrors](https://github.com/golang/xerrors), [pkg/errors](https://github.com/pkg/errors), and [Go 1.13 errors](https://golang.org/pkg/errors/).

`go get github.com/rotisserie/eris`

Check out the [package docs](https://godoc.org/github.com/rotisserie/eris) for more detailed information or connect with us on our [Slack channel](https://gorotisserie.slack.com/archives/CS13EC3T6) if you want to discuss anything in depth.

## How is eris different?

Named after the Greek goddess of strife and discord, this package is intended to give you more control over error handling via error wrapping, stack tracing, and output formatting. Basic error wrapping was added in Go 1.13, but it omitted user-friendly `Wrap` methods and built-in stack tracing. Other error packages provide some of the features found in `eris` but without flexible control over error output formatting. This package provides default string and JSON formatters with options to control things like separators and stack trace output. However, it also provides an option to write custom formatters via [`eris.Unpack`](https://godoc.org/github.com/morningvera/eris#Unpack).

Error wrapping behaves somewhat differently than existing packages. It relies on root errors that contain a full stack trace and wrap errors that contain a single stack frame. When errors from other packages are wrapped, a root error is automatically created before wrapping it with the new context. This allows `eris` to work with other error packages transparently and elimates the need to manage stack traces manually. Unlike other packages, `eris` also works well with global error types by automatically updating stack traces during error wrapping.

## Types of errors

`eris` is concerned with only three different types of errors: root errors, wrap errors, and external errors. Root and wrap errors are defined types in this package and all other error types are external or third-party errors.

Root errors are created via [`eris.New`](https://godoc.org/github.com/rotisserie/eris#New) and [`eris.Errorf`](https://godoc.org/github.com/rotisserie/eris#Errorf). Generally, it's a good idea to maintain a set of root errors that are then wrapped with additional context whenever an error of that type occurs. Wrap errors represent a stack of errors that have been wrapped with additional context. Unwrapping these errors via [`eris.Unwrap`](https://godoc.org/github.com/rotisserie/eris#Unwrap) will return the next error in the stack until a root error is reached. [`eris.Cause`](https://godoc.org/github.com/rotisserie/eris#Cause) will also retrieve the root error.

When external error types are wrapped with additional context, a root error is first created from the original error. This creates a stack trace for the error and allows it to function with the rest of the `eris` package.

## Wrapping errors with additional context

[`eris.Wrap`](https://godoc.org/github.com/rotisserie/eris#Wrap) adds context to an error while preserving the type of the original error. This method behaves differently for each error type. For root errors, the stack trace is reset to the current callers which ensures traces are correct when using global/sentinel error values. Wrapped error types are simply wrapped with the new context. For external types (i.e. something other than root or wrap errors), a new root error is created for the original error and then it's wrapped with the additional context.

```golang
_, err := db.Get(id)
if err != nil {
  // return the error with some useful context
  return eris.Wrapf(err, "error getting resource '%v'", id)
}
```

## Inspecting error types

The `eris` package provides a few ways to inspect and compare error types. [`eris.Is`](https://godoc.org/github.com/rotisserie/eris#Is) returns true if a particular error appears anywhere in the error chain, and `eris.Cause` returns the root cause of the error. Currently, `eris.Is` works simply by comparing error messages with each other. If an error contains a particular message anywhere in its chain (e.g. "not found"), it's defined to be that error type (i.e. `eris.Is` will return `true`).

```golang
NotFound := eris.New("not found")
_, err := db.Get(id)
// check if the resource was not found
if eris.Is(err, NotFound) || eris.Cause(err) == NotFound {
  // return the error with some useful context
  return eris.Wrapf(err, "error getting resource '%v'", id)
}
```

## Stack traces

Errors created with this package contain stack traces that are managed automatically even when wrapping global errors or errors from other libraries. Stack traces are currently mandatory when creating and wrapping errors but optional when printing or logging errors. Printing an error with or without the stack trace is simple:

```golang
_, err := db.Get(id)
if err != nil {
  return eris.Wrapf(err, "error getting resource '%v'", id)
}
fmt.Printf("%v", err) // print without the stack trace
fmt.Printf("%+v", err) // print with the stack trace
```

For an error that has been wrapped once, the output will look something like this:

```
# output without the stack trace
error getting resource 'example-id': not found

# output with the stack trace
error getting resource 'example-id'
  api.GetResource: /path/to/file/api.go: 30
not found
  api.GetResource: /path/to/file/api.go: 30
  db.Get: /path/to/file/db.go: 99
  runtime.goexit: /path/to/go/src/libexec/src/runtime/asm_amd64.s: 1337
```

The first layer of the full error output shows a message ("error getting resource 'example-id'") and a single stack frame. The next layer shows the root error ("not found") and the full stack trace.

## Formatted error printing

The default format in `eris` is returned by the method [`NewDefaultFormat()`](https://godoc.org/github.com/morningvera/eris#NewDefaultFormat). Below you can see what a default formatted error in `eris` might look like.

Errors printed without trace using `fmt.Printf("%v\n", err)`

```
even more context: additional context: root error
```

Errors printed with trace using `fmt.Printf("%+v\n", err)`

```
even more context
        eris_test.setupTestCase: ../eris/eris_test.go: 17
additional context
        eris_test.setupTestCase: ../eris/eris_test.go: 17
root error
        eris_test.setupTestCase: ../eris/eris_test.go: 17
        eris_test.TestErrorFormatting: ../eris/eris_test.go: 226
        testing.tRunner: ../go1.11.4/src/testing/testing.go: 827
        runtime.goexit: ../go1.11.4/src/runtime/asm_amd64.s: 1333
```

'eris' also provides developers a way to define a custom format to print the errors. The [`Format`](https://godoc.org/github.com/morningvera/eris#Format) object defines separators for various components of the error/trace and can be passed to utility methods for printing string and JSON formats.

## Error object

The [`UnpackedError`](https://godoc.org/github.com/morningvera/eris#UnpackedError) object provides a convenient and developer friendly way to store and access existing error traces. The `ErrChain` and `ErrRoot` fields correspond to `wrapError` and `rootError` types, respectively. If any other error type is unpacked, it will appear in the ExternalErr field.

The [`Unpack()`](https://godoc.org/github.com/morningvera/eris#Unpack) method returns the corresponding `UnpackedError` object for a given error. This object can also be converted to string and JSON for logging and printing error traces. This can be done by using the methods [`ToString()`](https://godoc.org/github.com/morningvera/eris#UnpackedError.ToString) and [`ToJSON()`](https://godoc.org/github.com/morningvera/eris#UnpackedError.ToJSON). Note the `ToJSON()` method returns a `map[string]interface{}` type which can be marshalled to JSON using the `encoding/json` package.

## Logging errors with more control

While `eris` supports logging errors with Go's `fmt` package, it's often advantageous to use the provided string and JSON formatters instead. These methods provide much more control over the error output and should work seamlessly with whatever logging package you choose. The example below shows how to integrate `eris` with (logrus)[https://github.com/sirupsen/logrus].

```golang
var fields log.Fields
unpackedErr := eris.Unpack(err)
fields["method"] = "api.GetResource"
fields["error"] = unpackedErr.ToJSON(eris.NewDefaultFormat(true))
logger.WithFields(fields).Errorf("method completed with error (%v)", err)
```

When using a JSON logger, the output should look something like this:

```json
{
  "method":"api.GetResource",
  "error":{
    "error chain":[
      {
        "message":"error getting resource 'example-id'",
        "stack":"api.GetResource: /path/to/file/api.go: 30"
      }
    ],
    "error root":{
      "message":"not found",
      "stack":[
        "api.GetResource: /path/to/file/api.go: 30",
        "db.Get: /path/to/file/db.go: 99",
        "runtime.goexit: /path/to/go/src/runtime/asm_amd64.s: 1337"
      ]
    }
  }
}
```

## Migrating to eris

Migrating to `eris` should be a very simple process. If it doesn't offer something that you currently use from existing error packages, feel free to submit an issue to us. If you don't want to refactor all of your error handling yet, `eris` should work relatively seamlessly with your existing error types. Please submit an issue if this isn't the case for some reason.

## Contributing

If you'd like to contribute to `eris`, we'd love your input! Please submit an issue first so we can discuss your proposal. We're also available to discuss potential issues and features on our [Slack channel](https://gorotisserie.slack.com/archives/CS13EC3T6).
