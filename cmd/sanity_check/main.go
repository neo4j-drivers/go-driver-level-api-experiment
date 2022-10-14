package main

import "fmt"

func main() {
	fmt.Println(hello("world"))
}

// use generics to make sure Go 1.18+ is installed
func hello[T string](value T) string {
	return fmt.Sprintf("Hello %s!", value)
}
