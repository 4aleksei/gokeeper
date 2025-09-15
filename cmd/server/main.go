package main

import (
	"log"

	"github.com/4aleksei/gokeeper/internal/common/version"
	"github.com/4aleksei/gokeeper/internal/server/app"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	version.PrintVersion(buildVersion, buildDate, buildCommit)
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
