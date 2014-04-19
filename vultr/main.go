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
        "vultr.config",
        pit.Requires{"api_key": "your api key. see https://my.vultr.com/settings/"},
    )
    if err != nil {
        errExit(err)
    }
    endpoint, err := url.Parse("https://api.vultr.com/")
    if err != nil {
        errExit(err)
    }
    apiKey := (*profile)["api_key"]

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
