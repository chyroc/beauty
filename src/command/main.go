package main

import (
	"github.com/chyroc/beauty/src/image_adapter"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	weiboCli := image_adapter.NewWeibo()
	images,err:=weiboCli.GetImage("3942679334")
	if err != nil {
		panic(err)
	}
	spew.Dump(images)
}
