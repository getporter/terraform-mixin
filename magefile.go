//go:build mage

package main

import (
	"get.porter.sh/magefiles/mixins"
	"get.porter.sh/magefiles/porter"
	"github.com/carolynvs/magex/shx"
)

const (
	mixinName    = "terraform"
	mixinPackage = "get.porter.sh/mixin/terraform"
	mixinBin     = "bin/mixins/" + mixinName
)

var (
	magefile = mixins.NewMagefile(mixinPackage, mixinName, mixinBin)
	must     = shx.CommandBuilder{StopOnError: true}
)

func ConfigureAgent() {
	magefile.ConfigureAgent()
}

// Build the mixin
func Build() {
	magefile.Build()
}

// Cross-compile the mixin before a release
func XBuildAll() {
	magefile.XBuildAll()
}

// Run unit tests
func TestUnit() {
	magefile.TestUnit()
}

func Test() {
	magefile.Test()
	Build()
	TestIntegration()
}

// Publish the mixin to github
func Publish() {
	magefile.Publish()
}

// TestPublish tries out publish locally, with your github forks
// Assumes that you forked and kept the repository name unchanged.
func TestPublish(username string) {
	magefile.TestPublish(username)
}

// Install the mixin
func Install() {
	magefile.Install()
}

// Remove generated build files
func Clean() {
	magefile.Clean()
}

// Install porter locally
func EnsureLocalPorter() {
	porter.UseBinForPorterHome()
	porter.EnsurePorter()
}

func TestIntegration() {
	EnsureLocalPorter()
	must.Command("./scripts/test/test-cli.sh").RunV()
}
