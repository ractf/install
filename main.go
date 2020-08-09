package main

import (
	"fmt"

	. "github.com/logrusorgru/aurora"
	"github.com/manifoldco/promptui"
)

func main() {
	fmt.Println(Cyan("Welcome to the"), Bold("RACTF"), Cyan("setup script"))

	selectedComponents := cumulativeSelect("Which services would you like to install?", []string{"Andromeda", "Core", "Shell"})

	if len(selectedComponents) == 0 {
		fmt.Println(Red("You must select at least one component to continue."))
		return
	}
}

func cumulativeSelect(prompt string, items []string) []string {
	var selected []string

	items = append(items, "Confirm")

	for index := 0; ; {
		prompt := promptui.Select{
			Label:        fmt.Sprintf("%s (Currently selected: %s)", Yellow(prompt), selected),
			Items:        items,
			HideSelected: true,
			CursorPos:    index,
		}

		index, choice, err := prompt.Run()

		if err != nil {
			fmt.Println("Prompt failed to display.")
			break
		}

		if index == len(items) - 1 {
			break
		}

		var removedFromSelected bool

		for i, val := range selected {
			if val == choice {
				selected = append(selected[:i], selected[i+1:]...)
				removedFromSelected = true
				break
			}
		}
		if !removedFromSelected {
			selected = append(selected, choice)
		}
	}

	return selected
}