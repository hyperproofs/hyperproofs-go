package vcs

import (
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/alinush/go-mcl"
)

// [start, stop)
func (vcs *VCS) PrkGen(index uint8, start uint64, stop uint64, wg *sync.WaitGroup) {

	os.MkdirAll(vcs.folderPath, os.ModePerm)
	fileName := vcs.folderPath + fmt.Sprintf(PRKNAME, index)
	f, err := os.Create(fileName)
	check(err)

	var result mcl.G1
	for i := start; i < stop; i++ {
		exponent := vcs.SelectPRK(i)
		mcl.G1Mul(&result, &vcs.G, &exponent)
		_, err = f.Write(result.Serialize())
		check(err)
		if !vcs.PARAM_TOO_LARGE {
			vcs.PRK[i] = result
		}
		// fmt.Println(i)
	}
	fmt.Println("Dumped ", fileName, BoundsPrint(start, stop))
	defer f.Close()
	defer wg.Done()
}

func (vcs *VCS) PrkGenDriver() {
	fmt.Println(SEP, "Generating the PRK", SEP)
	if !vcs.PARAM_TOO_LARGE {
		// Actually we can avoid during Save
		vcs.PRK = make([]mcl.G1, vcs.N) // Allocate space for PRK
	}
	var wg sync.WaitGroup
	step := uint64(math.Ceil(float64(vcs.N) / float64(NFILES))) // Maximum size of each file.

	start := uint64(0)
	stop := step
	for i := uint8(0); i < NFILES; i++ {
		wg.Add(1)
		go vcs.PrkGen(i, start, stop, &wg)

		start += step
		stop += step
		if (i+1)%NCORES == 0 {
			wg.Wait()
		}
	}
	wg.Wait()
}

func (vcs *VCS) UpkGen(index uint8, start uint64, stop uint64, wg *sync.WaitGroup) {

	os.MkdirAll(vcs.folderPath, os.ModePerm)
	fileName := vcs.folderPath + fmt.Sprintf(UPKNAME, index)
	f, err := os.Create(fileName)
	check(err)
	// fmt.Println(fileName)
	var result mcl.G1
	for j := start; j < stop; j++ {
		i, k := IndexInTheLevel(j)
		exponent := vcs.SelectUPK(i, k)
		mcl.G1Mul(&result, &vcs.G, &exponent)
		_, err = f.Write(result.Serialize())
		check(err)
		if !vcs.PARAM_TOO_LARGE {
			vcs.UPK[i][k] = result
		}

		// fmt.Println(i, k, exponent.IsZero(), result.IsZero(), vcs.PRK[i][k].IsZero())
	}

	fmt.Println("Dumped ", fileName, BoundsPrint(start, stop))
	defer f.Close()
	defer wg.Done()
}

func (vcs *VCS) UpkGenDriver() {

	var wg sync.WaitGroup
	if !vcs.PARAM_TOO_LARGE {
		// Allocate space for UPK
		vcs.MallocUpk()
	}
	fmt.Println(SEP, "Generating the UPK", SEP)

	numUPK := (uint64(1) << (vcs.L + 1)) - 1 // Number of nodes in the UPK tree
	step := uint64(math.Ceil(float64(numUPK) / float64(NFILES)))
	start := uint64(0)
	stop := step

	for i := uint8(0); i < NFILES; i++ {
		// fmt.Println(i, start, stop)
		wg.Add(1)
		go vcs.UpkGen(i, start, stop, &wg)

		start += step
		stop += step
		stop = minUint64(stop, numUPK)

		if (i+1)%NCORES == 0 {
			wg.Wait()
		}
	}
	wg.Wait()
}

func (vcs *VCS) PrkUpkGen() {

	vcs.UpkGenDriver()
	if !vcs.DISCARD_PRK && !vcs.PARAM_TOO_LARGE {
		fmt.Println(SEP)
		vcs.PrkGenDriver() // This also allocates memory for PRK
	}
	fmt.Println(SEP)
}
