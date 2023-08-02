package main

import (
	"fmt"
	"os"

	"gosearch/application"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please enter a subCommand [index, search, serve]")
		os.Exit(1)
	}
	app := &application.Application{DirPath: "content/craftinginterpreters/book/", IndexPath: "index.json", StaticContent: "./static"}
	subCommand := os.Args[1]

	switch subCommand {
	case "index":
		app.Index()
	case "search":
		if len(os.Args) < 3 {
			fmt.Println("Please enter a search term")
			os.Exit(1)
		}

		query := os.Args[2]

    result, err := app.Search(query)

    if err != nil {
      fmt.Println(err.Error())
    }

    for k, v := range result {
      fmt.Println(k)
      for _, v := range v {
        fmt.Println(v)
      }
    }

	case "serve":
		app.Serve()
	default:
		fmt.Println("Sub-Command not supported")
		os.Exit(1)
	}
}
