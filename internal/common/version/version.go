// Package version print
package version

import "fmt"

func PrintVersion(buildVersion, buildDate, buildCommit string) {

	fmt.Println("Build version: ", buildVersion)
	fmt.Println("Build date: ", buildDate)
	fmt.Println("Build commit: ", buildCommit)

}
