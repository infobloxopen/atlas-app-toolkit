# Object Data Transfer Tool

The `pb/converter.go` contains a tool to transfer data from service/API (pb generated)
objects to ORM compatible objects, because PB generated objects do not follow
golint case requirements (i.e. `Ip` or `Id` instead of `IP` and `ID`), and have special
handling for nullable values (Well-Known-Types wrappers).

The generated nature also prevents any custom ORM or SQL level tags from being added to the
Protocol Buffer objects.

This tool uses reflection to transfer between fields of the same name,
case insensitively, and for the WKTs from `wrappers.StringValue` to `*string`, and
`wrappers.UInt32Value` to `*uint32`.

More complicated PB behavior, such as oneof and maps is **not currently convertible**.

Note that this requires the definition of an ORM compatible object from
the Protocol Buffer generated objects with the proper types and foreign keys defined
for child objects.

Additional fields desired in the DB, but not exposed in the API,
potentially such as create/update times, and any necessary SQL or ORM tags
can be defined in this ORMified object.

```
package pb // import "github.com/infobloxopen/atlas-app-toolkit/pb"

func Convert(source interface{}, dest interface{}) error
    Convert Copies data between fields at ORM and service levels. Works under
    the assumption that any WKT fields in proto map to * fields at ORM.
```
