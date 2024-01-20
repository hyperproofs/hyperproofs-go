package main

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/alinush/go-mcl"
	"github.com/hyperproofs/hyperproofs-go/vcs"
)

func main() {
	testing.Init()
	flag.Parse()

	fmt.Println("Hello, World!")
	mcl.InitFromString("bls12-381")

	dt := time.Now()
	fmt.Println("Specific date and time is: ", dt.Format(time.UnixDate))

	fmt.Println(vcs.SEP)

	args := os.Args

	if len(args) == 1 {
		var L uint8
		// L = uint8(26)
		// _ = hyperGenerateKeys(L, false)

		L = uint8(30)
		_ = hyperGenerateKeys(L, true)
	} else {
		if args[1] == "1" {
			snarks_verifier()
		} else {
			Benchmark() // Uncomment this benchmark Commit and OpenAll.
		}
	}
}

func hyperGenerateKeys(L uint8, fake bool) *vcs.VCS {

	N := uint64(1) << L
	vcs := vcs.VCS{}

	fmt.Println("L:", L, "N:", N)
	folderPath := fmt.Sprintf("pkvk-%02d", L)
	if fake {
		vcs.KeyGenFake(16, L, folderPath, 1<<12)
	} else {
		vcs.KeyGen(16, L, folderPath, 1<<12)
	}

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
