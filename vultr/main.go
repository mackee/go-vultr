package main

import (
	"fmt"
	"github.com/gonuts/commander"
	"github.com/typester/go-pit"
	"log"
	"net/url"
	"os"
)

var mainCmd = &commander.Command{
	UsageLine: os.Args[0],
}

var client *Client

func init() {
	mainCmd.Subcommands = []*commander.Command{
		listCmd,
		osCmd,
		regionsCmd,
		plansCmd,
		makeCreateCmd(),
		startCmd,
		haltCmd,
		rebootCmd,
		makeDestroyCmd(),
		sshCmd,
	}
}

func main() {
	profile, err := pit.Get(
		"vultr.config", pit.Requires{},
	)

	apiKey, exists := "", false
	if apiKey, exists = (*profile)["api_key"]; !exists {
		fmt.Print("your api key(https://my.vultr.com/settings API Information): ")
		_, err = fmt.Scanf("%s", &apiKey)
		if err != nil {
			errExit(err)
		}
		pit.Set("vultr.config", pit.Profile{"api_key": apiKey})
	}

	endpoint, err := url.Parse("https://api.vultr.com/")
	if err != nil {
		errExit(err)
	}

	log.Println("endpoint:", endpoint.String())

	client = NewClient(endpoint, apiKey)

	if err := mainCmd.Dispatch(os.Args[1:]); err != nil {
		errExit(err)
	}
}

func errExit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
