package main

import (
	"fmt"
	"github.com/ofux/deluge/dsl/repl"
	"os"
	"os/user"
)

/*

TODO:
	+ add line + column in tokens so we can have better error messages
	+ support UTF8
	+ add stacktraces to runtime errors
	+ support character escaping in double-quoted strings
	+ support back-quoted strings like in Go
	+ 'if' should not be an expression but a statement
	+ support 'else if'
	+ support comments // and / * * /
	+ assign statement =
	+ 'for' loop
	+ operators <= and >=
	+ operators || and &&
	+ statements ++ -- += -= *= /=
	+ floats
	+ operator %
	+ handle scopes (environments) properly
	+ rename 'fn' to 'function'
	- while loop
	- async / async "group" / wait / wait "group"
	- add built-in functions:
		- exit
		- assert
		- http
		- mqtt
		- tcp
		- grpc
		- push (arrays)
		- split (arrays)
		- indexOf (arrays)
		- split (strings)
		- indexOf (strings)
		-
	- check variable declaration / assignment at compile time

*/

func main() {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is the Deluge programming language!\n",
		usr.Username)
	fmt.Printf("Feel free to type in commands\n")
	repl.Start(os.Stdin, os.Stdout)
}
