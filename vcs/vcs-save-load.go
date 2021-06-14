package vcs

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/alinush/go-mcl"
)

func (vcs *VCS) SaveTrapdoor() {

	fmt.Println(SEP, "Saving data to:", vcs.folderPath, SEP)

	os.MkdirAll(vcs.folderPath, os.ModePerm)
	f, err := os.Create(vcs.folderPath + TRAPDOORNAME)
	check(err)

	// Report the size.
	LBytes := make([]byte, 8) // Enough space for 64 bits of interger
	binary.LittleEndian.PutUint64(LBytes, uint64(vcs.L))
	_, err = f.Write(LBytes)
	check(err)

	// Write KZG stuff first, as it not related to VCS or size of the VCS.
	_, err = f.Write(vcs.alpha.Serialize())
	check(err)
	_, err = f.Write(vcs.beta.Serialize())
	check(err)

	// Write the Generator to the file
	_, err = f.Write(vcs.G.Serialize())
	check(err)
	_, err = f.Write(vcs.H.Serialize())
	check(err)

	// Write the trapdoors to the file.
	for i := range vcs.trapdoors {
		_, err = f.Write(vcs.trapdoors[i].Serialize())
		check(err)
		_, err = f.Write(vcs.trapdoorsSubOne[i].Serialize())
		check(err)
		_, err = f.Write(vcs.trapdoorsSubOneRev[i].Serialize())
		check(err)
	}

	f.Close()
	fmt.Println(SEP, "Saved trapdoors", SEP)

	// Create a new file for VRK and write it.
	f, err = os.Create(vcs.folderPath + VRKNAME)
	check(err)

	for i := range vcs.VRK {
		_, err = f.Write(vcs.VRK[i].Serialize())
		check(err)
		_, err = f.Write(vcs.VRKSubOne[i].Serialize())
		check(err)
		_, err = f.Write(vcs.VRKSubOneRev[i].Serialize())
		check(err)
	}
	f.Close()
	fmt.Println(SEP, "Saved VRK", SEP)

}

func (vcs *VCS) LoadTrapdoor(L uint8) {

	f, err := os.Open(vcs.folderPath + TRAPDOORNAME)
	check(err)

	// fileinfo, err := f.Stat()
	// check(err)
	// filesize := fileinfo.Size() // In Bytes
	// estimatedEll := (filesize - int64(GetG1ByteSize()+GetG2ByteSize())) / int64(GetFrByteSize()) / 2

	var data []byte

	data = make([]byte, 8)
	_, err = f.Read(data)
	check(err)
	reportedEll := uint8(binary.LittleEndian.Uint64(data))

	if reportedEll < L {
		// Assumes SaveTrapdoor is honest
		panic(fmt.Sprintf("There is not enough to read! Found: %d, Wants: %d", reportedEll, L))
	}

	// Load KZG related stuff
	data = make([]byte, GetFrByteSize())
	_, err = f.Read(data)
	check(err)
	vcs.alpha.Deserialize(data)

	data = make([]byte, GetFrByteSize())
	_, err = f.Read(data)
	check(err)
	vcs.beta.Deserialize(data)

	// Load the VCS related stuff
	data = make([]byte, GetG1ByteSize())
	_, err = f.Read(data)
	check(err)
	vcs.G.Deserialize(data)

	data = make([]byte, GetG2ByteSize())
	_, err = f.Read(data)
	check(err)
	vcs.H.Deserialize(data)

	vcs.L = uint8(L)

	fmt.Println("Loading trapdoors:", L)
	for i := uint8(0); i < L; i++ {

		data = make([]byte, GetFrByteSize())
		_, err = f.Read(data)
		check(err)
		vcs.trapdoors[i].Deserialize(data)

		data = make([]byte, GetFrByteSize())
		_, err = f.Read(data)
		check(err)
		vcs.trapdoorsSubOne[i].Deserialize(data)

		data = make([]byte, GetFrByteSize())
		_, err = f.Read(data)
		check(err)
		vcs.trapdoorsSubOneRev[i].Deserialize(data)
	}

	f.Close()

	// Load VRKs
	f, err = os.Open(vcs.folderPath + VRKNAME)
	check(err)

	data = make([]byte, GetG2ByteSize())
	for i := uint8(0); i < L; i++ {
		_, err = f.Read(data)
		check(err)
		vcs.VRK[i].Deserialize(data)

		_, err = f.Read(data)
		check(err)
		vcs.VRKSubOne[i].Deserialize(data)

		_, err = f.Read(data)
		check(err)
		vcs.VRKSubOneRev[i].Deserialize(data)
	}
	f.Close()
}

func (vcs *VCS) UpkLoad(fileName string, index uint8, start uint64, stop uint64, wg *sync.WaitGroup) {
	f, err := os.Open(fileName)
	check(err)

	data := make([]byte, GetG1ByteSize())

	var result mcl.G1
	for j := start; j < stop; j++ {
		i, k := IndexInTheLevel(j)
		_, err = f.Read(data)
		check(err)
		result.Deserialize(data)

		vcs.UPK[i][k] = result

	}
	fmt.Println("Read ", fileName, BoundsPrint(start, stop))
	defer f.Close()
	defer wg.Done()
}

func (vcs *VCS) UpkLoadDriver() {

	// Allocate space for UPK
	vcs.MallocUpk()

	var wg sync.WaitGroup
	var step, start, stop uint64
	var total, totalBytes int64
	var i uint8

	var files []string
	var err error

	files, err = filepath.Glob(vcs.folderPath + "/upk*")
	check(err)
	totalBytes = int64(0)
	for i := range files {
		totalBytes += fileSize(files[i])
	}
	total = totalBytes / int64(GetG1ByteSize())
	step = uint64(math.Ceil(float64(total) / float64(NFILES)))

	numUPK := (uint64(1) << (vcs.L + 1)) - 1
	start = uint64(0)
	stop = step
	stop = minUint64(stop, numUPK)

	i = uint8(0)
	for start < numUPK {
		wg.Add(1)
		fileName := vcs.folderPath + fmt.Sprintf(UPKNAME, i)
		go vcs.UpkLoad(fileName, i, start, stop, &wg)
		// fmt.Println(i)
		// fmt.Println("Reading chuck range:", i, BoundsPrint(start, stop))
		start += step
		stop += step
		stop = minUint64(stop, numUPK)
		i++
	}
	wg.Wait()
}

func (vcs *VCS) PrkLoad(fileName string, index uint8, start uint64, stop uint64, wg *sync.WaitGroup) {

	f, err := os.Open(fileName)
	check(err)

	var data []byte
	data = make([]byte, GetG1ByteSize())

	var result mcl.G1
	for i := start; i < stop; i++ {
		_, err = f.Read(data)
		check(err)
		result.Deserialize(data)
		vcs.PRK[i] = result
	}

	fmt.Println("Read ", fileName, BoundsPrint(start, stop))
	defer f.Close()
	defer wg.Done()
}

func (vcs *VCS) PrkLoadDriver() {
	// Allocate space for PRK
	vcs.PRK = make([]mcl.G1, vcs.N)
	var wg sync.WaitGroup
	var step, start, stop uint64
	var total, totalBytes int64
	var i uint8

	var files []string
	var err error

	files, err = filepath.Glob(vcs.folderPath + "/prk*")
	check(err)
	totalBytes = int64(0)
	for i := range files {
		totalBytes += fileSize(files[i])
	}
	total = totalBytes / int64(GetG1ByteSize())
	step = uint64(math.Ceil(float64(total) / float64(NFILES)))

	upperBound := (uint64(1) << vcs.L)

	i = uint8(0)
	start = uint64(0)
	stop = minUint64(step, upperBound)
	for start < upperBound {
		wg.Add(1)
		fileName := vcs.folderPath + fmt.Sprintf(PRKNAME, i)
		go vcs.PrkLoad(fileName, i, start, stop, &wg)

		// fmt.Println(i, fmt.Sprintf("%05d %05d", start, stop))
		start += step
		stop += step
		stop = minUint64(stop, upperBound)
		i++
	}
	wg.Wait()
}

func (vcs *VCS) PrkUpkLoad() {
	fmt.Println(SEP)
	vcs.UpkLoadDriver()
	fmt.Println(SEP)
	if !vcs.DISCARD_PRK {
		vcs.PrkLoadDriver()
		fmt.Println(SEP)
	}
}
