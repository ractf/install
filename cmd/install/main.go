package main

import (
	"errors"
	"fmt"
	. "github.com/logrusorgru/aurora"
	"github.com/markbates/pkger"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/template"
)

type options struct {
	EventName         string
	InstallComponents map[string]bool
	SecretKey         string
	FrontendURL       string
	APIDomain         string
	InternalName      string
	ComposePath       string
}

func main() {
	if runtime.GOOS == "windows" {
		fmt.Println("This script doesn't currently support windows.")
		fmt.Println("Maybe with your help, it could! Contributions to this script are welcome at https://github.com/ractf/install")
		return
	}

	fmt.Println(Cyan("Welcome to the"), Bold("RACTF"), Cyan("setup script"))

	installOptions := options{}

	var err error
	installOptions.ComposePath, err = exec.LookPath("docker-compose")
	if err != nil {
		fmt.Println(Red("docker-compose, a dependency of this script, doesn't appear to be installed."))
		fmt.Println(Red("If it is, ensure its executable is in the current user's PATH."))
		return
	}

	installOptions.InstallComponents, err = cumulativeSelect("Which services would you like to install?", []string{"Andromeda", "Core", "Shell"})
	if err != nil {
		fmt.Println(Red("There was an error displaying a prompt."))
		return
	}

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

	installOptions.EventName, err = promptString("What's the (short) name of your event (e.g. RACTF)?", stringValidator)
	if err != nil {
		fmt.Println(Red("There was an error displaying a prompt."))
		return
	}
	installOptions.InternalName = strings.Trim(strings.ReplaceAll(strings.ToLower(installOptions.EventName), " ", "_"), "./")

	if installOptions.InstallComponents["Shell"] {
		apiDomain, err := promptString("What's the public URL of your API? (e.g https://api.ractf.co.uk/)", stringValidator)
		if err != nil {
			fmt.Println(Red("There was an error displaying a prompt."))
			return
		}
		apiDomain = strings.TrimPrefix(apiDomain, "https://")
		apiDomain = strings.TrimPrefix(apiDomain, "http://")
		apiDomain = strings.TrimRight(apiDomain, "/")
		installOptions.APIDomain = apiDomain
	}

	if installOptions.InstallComponents["Core"] {
		frontendURL, err := promptString("What URL will visitors access your site through? (e.g. https://2020.ractf.co.uk/)", stringValidator)
		if err != nil {
			fmt.Println(Red("There was an error displaying a prompt."))
			return
		}
		if !strings.HasPrefix(frontendURL, "http") {
			frontendURL = "https://" + frontendURL
		}
		if !strings.HasSuffix(frontendURL, "/") {
			frontendURL += "/"
		}
		installOptions.FrontendURL = frontendURL
	}

	installOptions.SecretKey = GenerateRandomString(64)

	fmt.Println(Green("Proceeding with installation of"), Bold(installCount), Green("components."))

	err = generateAndWriteDockerFile(installOptions)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = generateAndWriteSystemdUnit(installOptions)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(Green("Selected services successfully installed. Run"), Yellow(fmt.Sprintf("`systemctl enable --now ractf_%s`", installOptions.InternalName)), Green("to start the service."))
}

func generateAndWriteSystemdUnit(options options) error {
	tf, err := pkger.Open("/assets/templates/systemd-unit.tmpl")
	if err != nil {
		return err
	}
	templ, err := ioutil.ReadAll(tf)
	if err != nil {
		return err
	}

	t, err := template.New("systemdUnit").Parse(string(templ))
	if err != nil {
		return err
	}

	f, err := os.Create(fmt.Sprintf("/etc/systemd/system/ractf_%s.service", options.InternalName))
	if err != nil {
		return err
	}

	err = t.Execute(f, options)
	if err != nil {
		return err
	}

	return nil
}

func generateAndWriteDockerFile(options options) error {
	tf, err := pkger.Open("/assets/templates/docker-compose.tmpl")
	if err != nil {
		return err
	}
	templ, err := ioutil.ReadAll(tf)
	if err != nil {
		return err
	}

	t, err := template.New("dockerCompose").Parse(string(templ))
	if err != nil {
		return err
	}

	if _, err := os.Stat(fmt.Sprintf("/opt/ractf/%s/", options.InternalName)); os.IsNotExist(err) {
		err = os.MkdirAll(fmt.Sprintf("/opt/ractf/%s/", options.InternalName), 0700)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(fmt.Sprintf("/opt/ractf/%s/docker-compose.yaml", options.InternalName))
	if err != nil {
		return err
	}

	err = t.Execute(f, options)
	if err != nil {
		return err
	}

	return nil
}

func stringValidator(input string) error {
	if len(input) == 0 {
		return errors.New("input must be longer than one char")
	}
	return nil
}
