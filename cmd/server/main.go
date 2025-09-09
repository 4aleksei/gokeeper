package main

import (
	"fmt"
	"log"

	"github.com/4aleksei/gokeeper/internal/server/app"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func printVersion() {
	fmt.Println("Build version: ", buildVersion)
	fmt.Println("Build date: ", buildDate)
	fmt.Println("Build commit: ", buildCommit)
}

func main() {
	printVersion()
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
