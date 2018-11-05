# Error fields

Package contains `proto` representation for `fields` section inside [error response](../../errors).

[Proto file](error_fields.proto) defines `FieldsInfo message` which is a default representation of field details
that conforms REST API Syntax Specification. It has it's own specific marshal/unmarshal logic.
 
This functionality should not be directly used in your code. Use `WithFields` function from 
[errors package](../../errors) instead.
