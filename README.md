# structfilter

![](https://github.com/TheCount/go-structfilter/workflows/CI/badge.svg)
[![Documentation](https://godoc.org/github.com/TheCount/go-structfilter/structfilter?status.svg)](https://godoc.org/github.com/TheCount/go-structfilter/structfilter)

structfilter is a Go package for filtering out structure fields and/or changing structure tags on the fly.

This package is useful for censoring sensitive data before publishing it (e. g., to a logging framework), or altering the behaviour of marshallers by injecting new key/value pairs into struct tags. The package creates fresh types from scratch at runtime, to hold the censored data, so not only sensitive values but also their field names are gone.

## Install

```sh
go get github.com/TheCount/go-structfilter/structfilter
```

## Usage

For the detailed API, see the [Documentation](https://godoc.org/github.com/TheCount/go-structfilter/structfilter).

In the following example, you will learn how to use the structfilter package to filter out sensitive password information from a user database and prepare it for marshalling to JSON.

Suppose you have a simple user database like this:

```golang
type User struct {
	Name          string
	Password      string
	PasswordAdmin string
	LoginTime     int64
}

var userDB = []User{
	User{
		Name:          "Alice",
		Password:      "$6$sensitive",
		PasswordAdmin: "$6$verysensitive",
		LoginTime:     1234567890,
	},
	User{
		Name:      "Bob",
		Password:  "$6$private",
		LoginTime: 1357924680,
	},
}
```

Now suppose you want to convert this user database to JSON. However, you don't want the sensitive password hash fields `Password` and `PasswordAdmin` in the JSON data. Furthermore, you want the JSON marshaller to print the JSON object keys in all lowercase. With structfilter, you can accomplish this with a filter consisting of two *filter functions*, one to strip out the `Password` and `PasswordAdmin` fields, and another one to inject `json:"fieldname"` tags to force lowercase fields:

```golang
filter := structfilter.New(
	structfilter.RemoveFieldFilter(regexp.MustCompile("^Password.*$")),
	func(f *structfilter.Field) error {
		f.Tag = reflect.StructTag(fmt.Sprintf(`json:"%s"`,
			strings.ToLower(f.Name())))
		return nil
	},
)
converted, err := filter.Convert(userDB)
if err != nil {
	log.Fatal(err)
}
jsonData, err := json.MarshalIndent(converted, "", "    ")
if err != nil {
	log.Fatal(err)
}
fmt.Println(string(jsonData))
```

This will print the following output to the console:

```json
[
    {
        "name": "Alice",
        "logintime": 1234567890
    },
    {
        "name": "Bob",
        "logintime": 1357924680
    }
]
```

Check out the [complete example here](https://github.com/TheCount/go-structfilter/blob/master/structfilter/examples/userdbjson/userdbjson.go)!

### Unfiltered types

Suppose in the above example, you had a slightly different definition of the `User` type:

```golang
type User struct {
	Name          string
	Password      string
	PasswordAdmin string
	LoginTime     time.Time
}
```

The `LoginTime` field now has the structure type `time.Time` instead of `inte64`. With this definition, your output becomes:

```json
[
    {
        "name": "Alice",
        "logintime": {}
    },
    {
        "name": "Bob",
        "logintime": {}
    }
]
```

Obviously, this is not what you want. What happened here? The `time.Time` type has a `MarshalJSON` method which is recognised by the `json` package, so you would get a nice time representation. However, the `time.Time` structure type gets filtered into a new structure type without any methods and without any unexported fields (in this case no fields at all, since `time.Time` does not have any exported fields, leading to a plain empty `struct{}`). See also [Restrictions](#Restrictions) below.

In order to avoid this problem, you can mark a type not to be filtered using the `UnfilteredType` method, e. g.,

```golang
filter.UnfilteredType(time.Time{})
```

Now the generated structure will keep the `time.Time` field instead of creating an empty structure type:

```json
[
    {
        "name": "Alice",
        "logintime": "2020-05-07T22:38:48.569442624+02:00"
    },
    {
        "name": "Bob",
        "logintime": "2020-05-07T23:37:48.569443375+02:00"
    }
]
```

The [complete updated example](https://github.com/TheCount/go-structfilter/blob/master/structfilter/examples/userdbjson_unfiltered/userdbjson_unfiltered.go) is available in the repository.

## Restrictions

structfilter uses Go's [reflect package](https://golang.org/pkg/reflect/) internally. Unfortunately, the reflect package comes with certain restrictions.

### Recursive types

The reflect package does not allow the creation of recursive types at runtime. A recursive type is a type which, directly or indirectly, refers to itself. A standard example of a recursive type would be:

```golang
type TreeNode struct {
  Value        interface{}
  LeftSibling  *TreeNode
  RightSibling *TreeNode
}
```

Here, the `TreeNode` type refers directly to itself via its `LeftSibling` and `RightSibling` fields. Whenever structfilter encounters a recursive type, it maps it to a plain `interface{}`. This generally works well, as `interface{}` takes any value, and third-party packages usually don't care whether a value is wrapped in an interface or not.

### Methods and unexported fields

The reflect package does not allow creation of named types with methods (cf. [golang/go#16522](https://github.com/golang/go/issues/16522)) or structures with unexported fields (cf. [golang/go#25401](https://github.com/golang/go/issues/25401)). As a result, the types generated by structfilter will have no methods at all and no unexported fields.
As a consequence, the generated types also have no fields with a static interface type other than plain `interface{}`. The loss of methods can have unintended consequences, e. g., when a `MarshalJSON()` or similar method is lost, or when a marshaller or logger expects a specific interface. The loss of unexported fields is generally not a problem, except in the case where the filtered value is somehow passed back to the package which defined those fields in the first place.

The `UnfilteredType` method can be used to work around some of these restrictions.

### Unsafe pointers

If a third-party package hides information behind `unsafe.Pointer` fields, structfilter has no way of knowing what this data may be and merely copies the pointer without trying to dereference it. This usually does not cause *additional* problems as other packages will have the same conundrum as well.
