package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/TheCount/go-structfilter/structfilter"
)

// User represents a user entry in a user database.
type User struct {
	Name          string
	Password      string
	PasswordAdmin string
	LoginTime     time.Time
}

var userDB = []User{
	User{
		Name:          "Alice",
		Password:      "$6$sensitive",
		PasswordAdmin: "$6$verysensitive",
		LoginTime:     time.Now().Add(-time.Hour),
	},
	User{
		Name:      "Bob",
		Password:  "$6$private",
		LoginTime: time.Now().Add(-time.Minute),
	},
}

func main() {
	filter := structfilter.New(
		structfilter.RemoveFieldFilter(regexp.MustCompile("^Password.*$")),
		func(f *structfilter.Field) error {
			f.Tag = reflect.StructTag(fmt.Sprintf(`json:"%s"`,
				strings.ToLower(f.Name())))
			return nil
		},
	)
	filter.UnfilteredType(time.Time{})
	converted, err := filter.Convert(userDB)
	if err != nil {
		log.Fatal(err)
	}
	jsonData, err := json.MarshalIndent(converted, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jsonData))
}
