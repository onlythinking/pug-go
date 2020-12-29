package main

import (
	"github.com/sethvargo/go-signalcontext"
	"log"
)

func main() {

	ctx, cancel := signalcontext.OnInterrupt()
	defer cancel()

	<-ctx.Done()

	log.Fatal("关闭")
}
