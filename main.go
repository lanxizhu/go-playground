package main

import (
	"github.com/lanxizhu/go-playground/router"
)

func main() {
	r := router.SetupRouter()

	err := r.Run()
	if err != nil {
		return
	} // listens on 0.0.0.0:8080 by default
}
