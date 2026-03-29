package main

import "fmt"

func g() (int, int) {
	return 1, 2
}

func f(a int, b int, c int, d int, e int) {
	fmt.Println(a, b, c, d, e)
}

func main() {
	// f(0, g(), 3, 4) // Does this work?
}
