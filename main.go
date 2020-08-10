package main

import (
	"fmt"
	. "github.com/logrusorgru/aurora"
	"os"
	"text/template"
)

type options struct {
	EventName         string
	InstallComponents map[string]bool
	SecretKey         string
	FrontendURL       string
	APIDomain         string
}

func main() {
	fmt.Println(Cyan("Welcome to the"), Bold("RACTF"), Cyan("setup script"))

	installOptions := options{}

	installOptions.InstallComponents = cumulativeSelect("Which services would you like to install?", []string{"Andromeda", "Core", "Shell"})

	var installCount int
	for _, v := range installOptions.InstallComponents {
		if v {
			installCount += 1
		}
	}

	if installCount == 0 {
		fmt.Println(Red("You must select at least one service to continue."))
		return
	}
	if installOptions.InstallComponents["Andromeda"] {
		fmt.Println(Red("Andromeda install is not currently supported by this script."))
		return
	}

	if installOptions.InstallComponents["Shell"] {
		installOptions.EventName = promptString("What's the (short) name of your event (e.g. RACTF)?")
		installOptions.APIDomain = promptString("What's the public URL of your API? (Don't include http(s) or a trailing slash, include a port if necessary)")
	}

	if installOptions.InstallComponents["Core"] {
		installOptions.FrontendURL = promptString("What URL will visitors access your site through? (Include http(s) and a trailing /)")
	}

	installOptions.SecretKey = GenerateRandomString(64)

	fmt.Println(Green("Proceeding with installation of"), Bold(installCount), Green("components."))
	generateAndWriteDockerFile(installOptions)
}

func generateAndWriteDockerFile(options options) {
	t, err := template.ParseFiles("docker-compose.tmpl")
	if err != nil {
		fmt.Println(err)
		return
	}

	f, err := os.Create("docker-compose.yaml")
	if err != nil {
		fmt.Println("create file: ", err)
		return
	}

	err = t.Execute(f, options)
	if err != nil {
		fmt.Println(err)
		return
	}
}
