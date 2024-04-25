package runner

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type KvtoolRunnerConfig struct {
	ZgChainConfigTemplate string

	ImageTag   string
	IncludeIBC bool

	EnableAutomatedUpgrade     bool
	ZgChainUpgradeName         string
	ZgChainUpgradeHeight       int64
	ZgChainUpgradeBaseImageTag string

	SkipShutdown bool
}

// KvtoolRunner implements a NodeRunner that spins up local chains with kvtool.
// It has support for the following:
// - running a ZgChain node
// - optionally, running an IBC node with a channel opened to the ZgChain node
// - optionally, start the ZgChain node on one version and upgrade to another
type KvtoolRunner struct {
	config KvtoolRunnerConfig
}

var _ NodeRunner = &KvtoolRunner{}

// NewKvtoolRunner creates a new KvtoolRunner.
func NewKvtoolRunner(config KvtoolRunnerConfig) *KvtoolRunner {
	return &KvtoolRunner{
		config: config,
	}
}

// StartChains implements NodeRunner.
// For KvtoolRunner, it sets up, runs, and connects to a local chain via kvtool.
func (k *KvtoolRunner) StartChains() Chains {
	// install kvtool if not already installed
	installKvtoolCmd := exec.Command("./scripts/install-kvtool.sh")
	installKvtoolCmd.Stdout = os.Stdout
	installKvtoolCmd.Stderr = os.Stderr
	if err := installKvtoolCmd.Run(); err != nil {
		panic(fmt.Sprintf("failed to install kvtool: %s", err.Error()))
	}

	// start local test network with kvtool
	log.Println("starting 0gchain node")
	kvtoolArgs := []string{"testnet", "bootstrap", "--0gchain.configTemplate", k.config.ZgChainConfigTemplate}
	// include an ibc chain if desired
	if k.config.IncludeIBC {
		kvtoolArgs = append(kvtoolArgs, "--ibc")
	}
	// handle automated upgrade functionality, if defined
	if k.config.EnableAutomatedUpgrade {
		kvtoolArgs = append(kvtoolArgs,
			"--upgrade-name", k.config.ZgChainUpgradeName,
			"--upgrade-height", fmt.Sprint(k.config.ZgChainUpgradeHeight),
			"--upgrade-base-image-tag", k.config.ZgChainUpgradeBaseImageTag,
		)
	}
	// start the chain
	startZgChainCmd := exec.Command("kvtool", kvtoolArgs...)
	startZgChainCmd.Env = os.Environ()
	startZgChainCmd.Env = append(startZgChainCmd.Env, fmt.Sprintf("0GCHAIN_TAG=%s", k.config.ImageTag))
	startZgChainCmd.Stdout = os.Stdout
	startZgChainCmd.Stderr = os.Stderr
	log.Println(startZgChainCmd.String())
	if err := startZgChainCmd.Run(); err != nil {
		panic(fmt.Sprintf("failed to start 0gchain: %s", err.Error()))
	}

	// wait for chain to be live.
	// if an upgrade is defined, this waits for the upgrade to be completed.
	if err := waitForChainStart(kvtoolZgChainChain); err != nil {
		k.Shutdown()
		panic(err)
	}
	log.Println("0gchain is started!")

	chains := NewChains()
	chains.Register("0gchain", &kvtoolZgChainChain)
	if k.config.IncludeIBC {
		chains.Register("ibc", &kvtoolIbcChain)
	}
	return chains
}

// Shutdown implements NodeRunner.
// For KvtoolRunner, it shuts down the local kvtool network.
// To prevent shutting down the chain (eg. to preserve logs or examine post-test state)
// use the `SkipShutdown` option on the config.
func (k *KvtoolRunner) Shutdown() {
	if k.config.SkipShutdown {
		log.Printf("would shut down but SkipShutdown is true")
		return
	}
	log.Println("shutting down 0gchain node")
	shutdownZgChainCmd := exec.Command("kvtool", "testnet", "down")
	shutdownZgChainCmd.Stdout = os.Stdout
	shutdownZgChainCmd.Stderr = os.Stderr
	if err := shutdownZgChainCmd.Run(); err != nil {
		panic(fmt.Sprintf("failed to shutdown kvtool: %s", err.Error()))
	}
}
