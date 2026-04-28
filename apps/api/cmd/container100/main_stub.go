//go:build !linux

package main

import "fmt"

func main() {
	fmt.Println("container100 está disponível apenas em hosts Linux.")
}
