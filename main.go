package main

import (
	"fmt"
lib "kmodules.xyz/schema-checker"
	"go.bytebuilders.dev/ui-wizards/apis/wizards/v1alpha1"
)

func main() {
	checker := lib.New(
		"/home/tamal/go/src/go.bytebuilders.dev/ui-wizards",
		[]interface{}{
		v1alpha1.IdentityServerSpec{},
	})

	result, err := checker.CheckAll()
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
