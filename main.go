package main

import (
	"errors"
	"fmt"
	. "github.com/logrusorgru/aurora"
	"github.com/markbates/pkger"
	"io/ioutil"
	"os"
	"strings"
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
		installOptions.EventName = promptString("What's the (short) name of your event (e.g. RACTF)?", stringValidator)
		installOptions.APIDomain = promptString("What's the public URL of your API? (Don't include http(s) or a trailing slash, include a port if necessary)", partialDomainValidator)
	}

	if installOptions.InstallComponents["Core"] {
		installOptions.FrontendURL = promptString("What URL will visitors access your site through? (Include http(s) and a trailing /)", fullDomainValidator)
	}

	installOptions.SecretKey = GenerateRandomString(64)

	fmt.Println(Green("Proceeding with installation of"), Bold(installCount), Green("components."))
	generateAndWriteDockerFile(installOptions)
}

func generateAndWriteDockerFile(options options) {
	tf, err := pkger.Open("/docker-compose.tmpl")
	if err != nil {
		fmt.Println(err)
	}
	templ, err := ioutil.ReadAll(tf)
	if err != nil {
		fmt.Println(err)
	}

	t, err := template.New("dockerCompose").Parse(string(templ))
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

func stringValidator(input string) error {
	if len(input) == 0 {
		return errors.New("input must be longer than one char")
	}
	return nil
}

func partialDomainValidator(input string) error {
	if strings.HasPrefix(input, "http") {
		return errors.New("string should not start with http")
	}
	if strings.HasSuffix(input, "/") {
		return errors.New("string should not end with /")
	}
	return nil
}

func fullDomainValidator(input string) error {
	if !strings.HasPrefix(input, "http") {
		return errors.New("string should start with http")
	}
	if !strings.HasSuffix(input, "/") {
		return errors.New("string should end with /")
	}
	return nil
}
