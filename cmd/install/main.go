package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	. "github.com/logrusorgru/aurora"
	"github.com/markbates/pkger"
)

type options struct {
	EventName          string
	InstallComponents  map[string]bool
	SecretKey          string
	FrontendURL        string
	APIDomain          string
	InternalName       string
	ComposePath        string
	UserEmail          string
	AWSAccessKeyId     string
	AWSSecretAccessKey string
	UseWatchtower      bool
	AndromedaIp        string
	AndromedaKey       string
}

var installShellFlag = flag.Bool("shell", false, "Whether to install Shell")
var installCoreFlag = flag.Bool("core", false, "Whether to install Core")
var installAndromedaFlag = flag.Bool("andromeda", false, "Whether to install Andromeda")
var eventNameFlag = flag.String("eventname", "", "The name of the event")
var frontendURLFlag = flag.String("frontendurl", "", "The public URL of your shell instance")
var apiDomainFlag = flag.String("apidomain", "", "The public URL of your core instance")
var userEmailFlag = flag.String("email", "", "The email sent to LetsEncrypt for certificate provisioning")
var AWSAccessKeyIdFlag = flag.String("awsaccesskeyid", "", "AWS Acess Key ID (For mail)")
var AWSSecretAccessKeyFlag = flag.String("awsaccesskeysecret", "", "AWS Secret Access Key (For mail)")
var UseWatchtowerFlag = flag.Bool("usewatchtower", false, "Whether to use Watchtower to auto-update RACTF.")
var AndromedaIPFlag = flag.String("andromedaip", "", "IP users access challenges through")

func main() {
	flag.Parse()

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

	_, err = exec.LookPath("docker")
	if err != nil {
		fmt.Println(Red("docker, a dependency of this script, doesn't appear to be installed."))
		fmt.Println(Red("If it is, ensure its executable is in the current user's PATH."))
		return
	}

	if !(*installShellFlag || *installCoreFlag || *installAndromedaFlag) {
		installOptions.InstallComponents, err = cumulativeSelect("Which services would you like to install?", []string{"Andromeda", "Core", "Shell"})
		if err != nil {
			fmt.Println(Red("There was an error displaying a prompt."))
			return
		}
	} else {
		var install = make(map[string]bool)
		if *installCoreFlag {
			install["Core"] = true
		}
		if *installShellFlag {
			install["Shell"] = true
		}
		if *installAndromedaFlag {
			install["Andromeda"] = true
		}
		installOptions.InstallComponents = install
	}

	var installCount int
	for _, v := range installOptions.InstallComponents {
		if v {
			installCount++
		}
	}

	if installCount == 0 {
		fmt.Println(Red("You must select at least one service to continue."))
		return
	}

	installOptions.EventName, err = promptStringIfNotDefault("What's the (short) name of your event (e.g. RACTF)?", stringValidator, *eventNameFlag)
	if err != nil {
		fmt.Println(Red("There was an error displaying a prompt."))
		return
	}
	installOptions.InternalName = strings.Trim(strings.ReplaceAll(strings.ToLower(installOptions.EventName), " ", "_"), "./")

	installOptions.UserEmail, err = promptStringIfNotDefault("Which email should be sent to LetsEncrypt for certificate provisioning (Use one you control)?", stringValidator, *userEmailFlag)
	if err != nil {
		fmt.Println(Red("There was an error displaying a prompt."))
		return
	}

	apiDomain, err := promptStringIfNotDefault("What's the public URL of your API? (e.g https://api.ractf.co.uk/)", stringValidator, *apiDomainFlag)
	if err != nil {
		fmt.Println(Red("There was an error displaying a prompt."))
		return
	}
	apiDomain = strings.TrimPrefix(apiDomain, "https://")
	apiDomain = strings.TrimPrefix(apiDomain, "http://")
	apiDomain = strings.TrimRight(apiDomain, "/")
	installOptions.APIDomain = apiDomain

	frontendURL, err := promptStringIfNotDefault("What URL will visitors access your site through? (e.g. https://2020.ractf.co.uk/)", stringValidator, *frontendURLFlag)
	if err != nil {
		fmt.Println(Red("There was an error displaying a prompt."))
		return
	}
	frontendURL = strings.TrimPrefix(frontendURL, "https://")
	frontendURL = strings.TrimPrefix(frontendURL, "http://")
	frontendURL = strings.TrimRight(frontendURL, "/")
	installOptions.FrontendURL = frontendURL

	andromedaIP, err := promptStringIfNotDefault("What IP/hostname will users access challenges through? (e.g. 1.1.1.1)", stringValidator, *AndromedaIPFlag)
	if err != nil {
		fmt.Println(Red("There was an error displaying a prompt."))
		return
	}
	installOptions.AndromedaIp = andromedaIP

	installOptions.AWSAccessKeyId, err = promptStringIfNotDefault("AWS Access Key ID for mail?", awsKeyValidator, *AWSAccessKeyIdFlag)
	installOptions.AWSSecretAccessKey, err = promptStringIfNotDefault("AWS Secret Access Key ID for mail?", awsSecretValidator, *AWSSecretAccessKeyFlag)
	if err != nil {
		fmt.Println(Red("There was an error displaying a prompt."))
		return
	}

	installOptions.SecretKey = GenerateRandomString(64)
	installOptions.AndromedaKey = GenerateRandomString(64)
	installOptions.UseWatchtower = *UseWatchtowerFlag

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

	fmt.Println(Green("Selected services successfully installed."))
	fmt.Println(Blue(strings.Repeat("-", 30)))
	fmt.Println(Yellow("What you still need to do (if you haven't already!):"))
	fmt.Println(" - ", Green("Set your DNS so that the requisite domains point to this box"))
	fmt.Println(" - ", Green("Run"), Yellow(fmt.Sprintf("`systemctl enable --now ractf_%s`", installOptions.InternalName)), Green("to start the RACTF service on this box."), Red("(This might take a while on first run!)"))
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

func awsKeyValidator(input string) error {
	input = strings.TrimSpace(input)
	match, _ := regexp.MatchString("[A-Z0-9]{20}", input)
	if len(input) != 20 {
		return errors.New("AWS Access Key ID should be of length 20")
	}
	if !match {
		return errors.New("Invalid AWS Access key")
	}
	return nil
}

func awsSecretValidator(input string) error {
	input = strings.TrimSpace(input)
	match, _ := regexp.MatchString("[A-Za-z0-9/+=]{40}", input)
	if len(input) != 40 {
		return errors.New("AWS Secret Key should be of length 40")
	}
	if !match {
		return errors.New("Invalid AWS Secret Key")
	}
	return nil
}
