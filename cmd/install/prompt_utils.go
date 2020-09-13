package main

import (
	"fmt"
	. "github.com/logrusorgru/aurora"
	"github.com/manifoldco/promptui"
	"strings"
)

func promptString(promptMessage string, validate promptui.ValidateFunc) (string, error) {
	prompt := promptui.Prompt{
		Label:    fmt.Sprintf("%s", Yellow(promptMessage)),
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(result), nil
}

func cumulativeSelect(prompt string, items []string) (map[string]bool, error) {
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
			return nil, err
		}

		if index == len(items)-1 {
			break
		}

		selected[choice] = !selected[choice]
	}

	return selected, nil
}
