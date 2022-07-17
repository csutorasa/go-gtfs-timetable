package main

import (
	"github.com/artonge/go-gtfs"
)

func main() {
	h := &handler{}
	closeCh := make(chan error, 1)
	go func() {
		err := startWeb(h)
		closeCh <- err
	}()
	gs, err := gtfs.LoadSplitted("data", nil)
	if err != nil {
		panic(err)
	}
	h.Loaded(gs[0])
	err = <-closeCh
	if err != nil {
		panic(err)
	}
}
