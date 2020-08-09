package main

import (
	"fmt"
	. "github.com/logrusorgru/aurora"
	"github.com/manifoldco/promptui"
	"strings"
)

func main() {
	fmt.Println(Cyan("Welcome to the"), Bold("RACTF"), Cyan("setup script"))

	selectedComponents := cumulativeSelect("Which services would you like to install?", []string{"Andromeda", "Core", "Shell"})

	var installCount int
	for _, v := range selectedComponents {
		if v {
			installCount += 1
		}
	}

	if installCount == 0 {
		fmt.Println(Red("You must select at least one service to continue."))
		return
	}
	fmt.Println(Green("Proceeding with installation of"), Bold(installCount), Green("components."))
}

func cumulativeSelect(prompt string, items []string) map[string]bool {
	selected := make(map[string]bool)
	for _, v := range items {
		selected[v] = false
	}

	items = append(items, "Confirm")

	for {
		var enabledList []string
		for i, v := range selected {
			if v {
				enabledList = append(enabledList, i)
			}
		}
		if len(enabledList) == 0 {
			enabledList = []string{"[None]"}
		}

		prompt := promptui.Select{
			Label:        fmt.Sprintf("%s (Currently selected: %s)", Yellow(prompt), strings.Join(enabledList, ", ")),
			Items:        items,
			HideSelected: true,
		}

		index, choice, err := prompt.Run()

		if err != nil {
			fmt.Println("Prompt failed to display.")
			break
		}

		if index == len(items)-1 {
			break
		}

		selected[choice] = !selected[choice]
	}

	return selected
}
