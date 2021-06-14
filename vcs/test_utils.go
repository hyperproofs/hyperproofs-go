package vcs

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"

	"github.com/alinush/go-mcl"
)

func getKeyValuesFr(db map[uint64]mcl.Fr) ([]uint64, []mcl.Fr) {

	keys := make([]uint64, 0, len(db))
	values := make([]mcl.Fr, 0, len(db))
	for k, v := range db {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}

func getKeyValuesG1(db map[uint64]mcl.G1) ([]uint64, []mcl.G1) {

	keys := make([]uint64, 0, len(db))
	values := make([]mcl.G1, 0, len(db))
	for k, v := range db {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}

func fillRange(aFr *[]mcl.Fr, start uint64, stop uint64, wg *sync.WaitGroup) {
	for i := start; i < stop; i++ {
		(*aFr)[i].Random()
	}
	wg.Done()
}

func GenerateVector(N uint64) []mcl.Fr {
	var aFr []mcl.Fr
	aFr = make([]mcl.Fr, N)

	step := N / 16
	if step < 1 {
		step = 1
	}

	start := uint64(0)
	stop := start + step
	stop = minUint64(stop, N)
	var wg sync.WaitGroup
	for start < N {
		wg.Add(1)
		fillRange(&aFr, start, stop, &wg)
		start += step
		stop += step
		stop = minUint64(stop, N)
	}
	wg.Wait()

	return aFr
}

func SaveVector(N uint64, aFr []mcl.Fr) {
	folderPath := "pkvk/"
	os.MkdirAll(folderPath, os.ModePerm)
	fileName := folderPath + "/Vec.data"

	f, err := os.Create(fileName)
	check(err)
	fmt.Println(fileName)

	intBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(intBytes, N)
	_, err = f.Write(intBytes)
	check(err)

	for i := uint64(0); i < N; i++ {
		_, err = f.Write(aFr[i].Serialize())
		check(err)
	}
	fmt.Println("Dumped ", fileName)
	defer f.Close()
}

func LoadVector(N uint64, folderPath string) []mcl.Fr {

	fileName := folderPath + "/Vec.data"

	f, err := os.Open(fileName)
	check(err)

	var n uint64
	data := make([]byte, 8)

	_, err = f.Read(data)

	n = binary.LittleEndian.Uint64(data)

	if N > n {
		panic("Vec Load Error: There is not enough to read")
	}

	dataFr := make([]byte, GetFrByteSize())
	aFr := make([]mcl.Fr, N)

	for i := uint64(0); i < N; i++ {
		_, err = f.Read(dataFr)
		check(err)
		aFr[i].Deserialize(dataFr)
	}

	defer f.Close()
	return aFr
}

// Export the proofs in the 2D format for the VCS API
func GetProofVecFromDb(proofs_db map[uint64][]mcl.G1, indexVec []uint64) [][]mcl.G1 {
	proofVec := make([][]mcl.G1, len(indexVec))

	for i := range indexVec {
		proofVec[i] = make([]mcl.G1, len(proofs_db[indexVec[i]]))
		copy(proofVec[i], proofs_db[indexVec[i]])
	}
	return proofVec
}
