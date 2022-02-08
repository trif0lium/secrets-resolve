package main

import (
	"os"
	"strings"
)

func main() {
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		envKey := pair[0]
		envValue := pair[1]
	}
}
