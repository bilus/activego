package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/bilus/activego"
	"github.com/bilus/activego/examples/chat-gin/chat"
	"github.com/gin-gonic/gin"
	"github.com/go-webpack/webpack"
)

func init() {
	webpack.FsPath = "./public/webpack"
}

func viewHelpers() template.FuncMap {
	return template.FuncMap{
		"asset": webpack.AssetHelper,
	}
}

func main() {
	isDev := flag.Bool("dev", false, "development mode")
	flag.Parse()

	webpack.Init(*isDev)
	if *isDev {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Important part:
	router.SetFuncMap(viewHelpers())
	// End important part

	router.LoadHTMLFiles("./views/app.tmpl")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "app.tmpl", gin.H{})
	})

	server := activego.BuildServer(nil)
	server.Connected(chat.Connected)
	chatCh := server.Channel("ChatChannel")
	chatCh.Subscribed(chat.Subscribed).Received("message", chat.Message)
	embeddedAnycable := server.MakeEmbedded() // TODO: Passing nil to BuildServer doesn't make sense as DSL.
	router.GET("/cable", gin.WrapH(embeddedAnycable))

	if !*isDev {
		router.Static("/webpack", "../public/webpack")
	}

	log.Println("Listening on: 9000")
	log.Fatal(http.ListenAndServe(":9000", router))
}

func Connected(c activego.Connection) error {
	return c.IdentifiedBy("user", c.URL().Query().Get("user"))
}

func Subscribed(c activego.Connection, ch activego.Channel) error {
	return ch.StreamFrom("chat")
}

func Message(c activego.Connection, ch activego.Channel, data activego.ActionData) error {
	return ch.Broadcast("chat", data["text"])
}
