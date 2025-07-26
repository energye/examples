package main

import (
	"fmt"
	"github.com/energye/cefapi"
)

func main() {
	s := cefapi.CreateRefCounted(128, cefapi.OwnedByCEF)
	fmt.Println(s, cefapi.BaseRefCountedSize())
	s.Release()
}
