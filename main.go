package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Vinnybless/pokedex/internal/pokeapi"
)

func cleanInput(text string) []string {
	loweredText := strings.ToLower(text)
	words := strings.Fields(loweredText)
	return words
}

type config struct {
	pokeapiClient       pokeapi.Client
	nextLocationURL     *string
	previousLocationURL *string
}

func commandHelp(cfg *config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	cmds := getCommands()
	for k, v := range cmds {
		fmt.Printf("%s: %s\n", k, v.description)
	}

	return nil
}

func commandExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandMap(cfg *config) error {
	locationsResp, err := cfg.pokeapiClient.ListLocations(cfg.nextLocationURL)
	if err != nil {
		return err
	}

	cfg.nextLocationURL = locationsResp.Next
	cfg.previousLocationURL = locationsResp.Previous

	for _, loc := range locationsResp.Results {
		fmt.Println(loc.Name)
	}
	return nil
}

func commandMapb(cfg *config) error {
	if cfg.previousLocationURL == nil {
		return errors.New("you're on the first page")
	}

	locationResp, err := cfg.pokeapiClient.ListLocations(cfg.previousLocationURL)
	if err != nil {
		return err
	}

	cfg.nextLocationURL = locationResp.Next
	cfg.previousLocationURL = locationResp.Previous

	for _, loc := range locationResp.Results {
		fmt.Println(loc.Name)
	}
	return nil
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Lists next 20 locations",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Lists previous 20 locations",
			callback:    commandMapb,
		},
	}
}

func startREPL(cfg *config) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex > ")
		var userInput string

		for scanner.Scan() {
			word := scanner.Text()
			userInput = cleanInput(word)[0]
			break
		}

		cmd, ok := getCommands()[userInput]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		if err := cmd.callback(cfg); err != nil {
			fmt.Println(err)
		}
		continue
	}
}

func main() {
	pokeClient := pokeapi.NewClient(5 * time.Second)
	cfg := &config{
		pokeapiClient: pokeClient,
	}

	startREPL(cfg)
}
