package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/ethermint/app"
	cmdcfg "github.com/tharsis/ethermint/cmd/config"
)

func extendSystemLimit() {
	var rLimit syscall.Rlimit
	rLimit.Max = 2_000_000
	rLimit.Cur = 2_000_000
	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Setting Rlimit ", err)
	}
}

func main() {

	extendSystemLimit()
	setupConfig()
	cmdcfg.RegisterDenoms()

	rootCmd, _ := NewRootCmd()

	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)

		default:
			os.Exit(1)
		}
	}
}

func setupConfig() {
	// set the address prefixes
	config := sdk.GetConfig()
	cmdcfg.SetBech32Prefixes(config)
	cmdcfg.SetBip44CoinType(config)
	config.Seal()
}
