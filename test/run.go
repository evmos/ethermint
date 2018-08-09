package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime/pprof"
	"syscall"

	"github.com/cosmos/ethermint/state"
	"github.com/cosmos/ethermint/test/importer"

	dbm "github.com/tendermint/tendermint/libs/db"
)

var (
	cpuprofile = flag.String("cpu-profile", "", "write cpu profile `file`")
	blockchain = flag.String("blockchain", "data/blockchain", "file containing blocks to load")
	datadir    = flag.String("datadir", path.Join(os.Getenv("HOME"), ".ethermint"), "directory for ethermint data")
	cachesize  = flag.Int("cachesize", 1024*1024, "number of key-value pairs for the state stored in memory")
)

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Printf("could not create CPU profile: %v\n", err)
			return
		}

		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Printf("could not start CPU profile: %v\n", err)
			return
		}

		defer pprof.StopCPUProfile()
	}

	sigs := make(chan os.Signal, 1)
	interruptCh := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		interruptCh <- true
	}()

	stateDB := dbm.NewDB("state", dbm.LevelDBBackend, *datadir)
	codeDB := dbm.NewDB("code", dbm.LevelDBBackend, *datadir)

	ethermintDB, err := state.NewDatabase(stateDB, codeDB, *cachesize)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize geth Database: %v", err))
	}

	importer := importer.Importer{
		EthermintDB:    ethermintDB,
		BlockchainFile: *blockchain,
		Datadir:        *datadir,
		InterruptCh:    interruptCh,
	}

	importer.Import()
}
