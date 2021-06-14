package main

import (
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/alinush/go-mcl"
	"github.com/sshravan/hyperproofs-go/vcs"
)

func main() {
	testing.Init()
	flag.Parse()

	fmt.Println("Hello, World!")
	mcl.InitFromString("bls12-381")

	dt := time.Now()
	fmt.Println("Specific date and time is: ", dt.Format(time.UnixDate))

	fmt.Println(vcs.SEP)
	// L := uint8(17)
	// _ = hyperGenerateKeys(L) // Uncomment this to generate UPK for a specific ell.
	// // _ = hyperLoadKeys(L)     // Check if save and load works.
	// // fmt.Println("Did load and save work?", vcs.IsEqual(aVcs, bVcs))
	// Benchmark()
}

func hyperGenerateKeys(L uint8) *vcs.VCS {

	N := uint64(1) << L
	vcs := vcs.VCS{}

	fmt.Println("L:", L, "N:", N)
	folderPath := fmt.Sprintf("pkvk-%02d", L)
	vcs.KeyGen(16, L, folderPath, 1<<12)

	fmt.Println("KeyGen ... Done")
	return &vcs
}

func hyperLoadKeys(L uint8) *vcs.VCS {

	folderPath := fmt.Sprintf("pkvk-%02d", L)
	vcs := vcs.VCS{}

	vcs.KeyGenLoad(16, L, folderPath, 1<<12)

	fmt.Println("KeyGenLoad ... Done")
	return &vcs
}
