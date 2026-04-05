package main

import (
	"os"

	"aily-skills-auth-authcli/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:]))
}
