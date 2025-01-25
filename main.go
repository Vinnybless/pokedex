package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
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
	caughtPokemon       map[string]pokeapi.Pokemon
}

func commandHelp(cfg *config, args ...string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	cmds := getCommands()
	for k, v := range cmds {
		fmt.Printf("%s: %s\n", k, v.description)
	}

	return nil
}

func commandExit(cfg *config, args ...string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandMap(cfg *config, args ...string) error {
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

func commandMapb(cfg *config, args ...string) error {
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

func commandExplore(cfg *config, args ...string) error {
	if len(args) != 1 {
		return errors.New("you must provide a location name")
	}

	name := args[0]
	location, err := cfg.pokeapiClient.GetLocation(name)
	if err != nil {
		return err
	}
	fmt.Printf("Exploring %s...\n", location.Name)
	fmt.Println("Found Pokemon: ")
	for _, enc := range location.PokemonEncounters {
		fmt.Printf(" - %s\n", enc.Pokemon.Name)
	}
	return nil
}

func commandCatch(cfg *config, args ...string) error {
	if len(args) != 1 {
		return errors.New("you must provide a pokemon name")
	}

	name := args[0]
	pokemon, err := cfg.pokeapiClient.GetPokemon(name)
	if err != nil {
		return err
	}

	res := rand.Intn(pokemon.BaseExperience)

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon.Name)
	if res > 40 {
		fmt.Printf("%s escaped!\n", pokemon.Name)
		return nil
	}

	fmt.Printf("%s was caught!\n", pokemon.Name)
	fmt.Println("You may now inspect it with the inspect command.")

	cfg.caughtPokemon[pokemon.Name] = pokemon
	return nil
}

func commandInspect(cfg *config, args ...string) error {
	if len(args) != 1 {
		return errors.New("you must provide a pokemon name")
	}

	name := args[0]
	pokemon, ok := cfg.caughtPokemon[name]
	if !ok {
		return errors.New("you have not caught that pokemon")
	}

	fmt.Println("Name:", pokemon.Name)
	fmt.Println("Height:", pokemon.Height)
	fmt.Println("Weight:", pokemon.Weight)
	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  -%s: %v\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Println("Types:")
	for _, typeInfo := range pokemon.Types {
		fmt.Println("  -", typeInfo.Type.Name)
	}
	return nil
}

func commandPokedex(cfg *config, args ...string) error {
	fmt.Println("Your Pokedex:")
	for _, p := range cfg.caughtPokemon {
		fmt.Printf(" - %s\n", p.Name)
	}
	return nil
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config, ...string) error
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
		"explore": {
			name:        "explore <location_name>",
			description: "Explore at location",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch <pokemon_name>",
			description: "Attempt to catch a pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect <pokemon_name>",
			description: "View details about a caught Pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "See all the pokemon you've caught",
			callback:    commandPokedex,
		},
	}
}

func startREPL(cfg *config) {
	reader := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		reader.Scan()

		words := cleanInput(reader.Text())
		if len(words) == 0 {
			continue
		}

		commandName := words[0]
		args := []string{}
		if len(words) > 1 {
			args = words[1:]
		}

		command, exists := getCommands()[commandName]
		if exists {
			err := command.callback(cfg, args...)
			if err != nil {
				fmt.Println(err)
			}
			continue
		} else {
			fmt.Println("Unknown command")
			continue
		}
	}
}

func main() {
	pokeClient := pokeapi.NewClient(5*time.Second, 5*time.Minute)
	cfg := &config{
		caughtPokemon: map[string]pokeapi.Pokemon{},
		pokeapiClient: pokeClient,
	}

	startREPL(cfg)
}
