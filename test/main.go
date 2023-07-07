package main

import (
	"fmt"
	"strings"
)

const url = "https://stablediffusionapi.com/api/v4/dreambooth"

func main() {
	str := "key:dadasdadasdadasdasd"
	fmt.Println(strings.Cut(str, ":"))
}
