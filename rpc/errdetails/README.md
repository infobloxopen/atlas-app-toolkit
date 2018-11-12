# Error details

Package contains `proto` representation for `details` section inside [error response](../../errors).

[Proto file](error_details.proto) defines `TargetInfo message` which is a default representation of error details
that conforms REST API Syntax Specification. It has it's own specific marshal/unmarshal logic.
 
This functionality should not be directly used in your code. Use `WithDetails` function from 
[errors package](../../errors) instead.
