package main

import (
	"os"

	fyneapp "fyne.io/fyne/v2/app"
	powertoys "github.com/jaypaulb/MT-Canvus-Tools/go/tools/powertoys/internal/organisms/app"
)

func main() {
	insecureTLS := os.Getenv("CANVUS_INSECURE_TLS") == "1" || os.Getenv("CANVUS_INSECURE_TLS") == "true"
	for _, arg := range os.Args[1:] {
		if arg == "--insecure-tls" {
			insecureTLS = true
		}
	}

	application := fyneapp.New()

	mainWindow := powertoys.NewMainWindow(application, insecureTLS)
	mainWindow.ShowAndRun()
}
