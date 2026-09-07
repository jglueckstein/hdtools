// package main provides a greeting utility.
package main

import "fmt"

// Hello returns a greeting string.
func Hello() string {
	return "Hello, world"
}

func main() {
	fmt.Println(Hello())
}
