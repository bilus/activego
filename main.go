package main

import (
	"log"
	"sync"

	"activego/actions"
	"activego/anycable"
	"activego/anycable/adapters"
	"activego/anycable/test"
)

// main is the starting point for your Buffalo application.
// You can feel free and add to this `main` method, change
// what it does, etc...
// All we ask is that, at some point, you make sure to
// call `app.Serve()`, unless you don't want to start your
// application that is. :)
func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	// go func() {
	// log.Println("Starting AnyCable backend listening on 50051")
	// TODO: Pass adapter instead of broadcaster, keep the latter internal..
	broadcaster := anycable.NewBroadcaster(adapters.NewHTTPBroadcastAdapter("http://localhost:8090/_broadcast")) // TODO: Make configurable.
	server := anycable.BuildServer(broadcaster)

	// server.Connected(chat.Connected)
	// chatCh := server.Channel("ChatChannel")
	// chatCh.Subscribed(chat.Subscribed).Received("message", chat.Message)
	test.Setup(server)
	// embeddedAnycable := server.MakeEmbedded() // TODO: Passing nil to BuildServer doesn't make sense as DSL.

	if err := server.Serve(50051); err != nil {
		log.Fatal(err)
	}

	// }()

	wg.Add(1)
	go func() {
		// app := actions.App(&embeddedAnycable)
		app := actions.App(nil)
		if err := app.Serve(); err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("Servers started")
	wg.Wait()
}

/*
# Notes about `main.go`

## SSL Support

We recommend placing your application behind a proxy, such as
Apache or Nginx and letting them do the SSL heavy lifting
for you. https://gobuffalo.io/en/docs/proxy

## Buffalo Build

When `buffalo build` is run to compile your binary, this `main`
function will be at the heart of that binary. It is expected
that your `main` function will start your application using
the `app.Serve()` method.

*/
