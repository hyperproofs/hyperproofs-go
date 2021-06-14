package vcs

import (
	"fmt"
	"log"
	"os"

	"github.com/alinush/go-mcl"
)

// Some global variables. Used in vcs.go
// LSB first
// 6 -> 0 1 1
func ToBinary(index uint64, L uint8) []bool {
	binary := make([]bool, L)
	for i := uint8(0); i < L; i++ {
		if index%2 == 0 {
			binary[i] = false
		} else {
			binary[i] = true
		}
		index = index / 2
	}
	return binary
}

// Converts 1D to 2D
// n = 7 the it corresponds to a value in [lg][start] in the tree.
// Say, when indexed from 0, [10] element in a 1D array is same as [3][3]
func IndexInTheLevel(n uint64) (uint8, uint64) {
	if n < 0 {
		panic("Has to be greater than 0")
	}
	nPrime := n
	n = n + 1
	start := uint64(1)
	lg := uint8(0)
	for n > 1 {
		n = n >> 1
		start = start << 1
		lg++
	}
	start = start - 1
	return lg, nPrime - start
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func GetFrByteSize() int {
	return 32
	// return mcl.GetFrByteSize()
}

func GetG1ByteSize() int {
	return 48
	// return mcl.GetG1ByteSize()
}

func GetG2ByteSize() int {
	return 96
	// return mcl.GetG1ByteSize()
}

func GetGTByteSize() int {
	return 576
}

func min(a uint8, b int64) int64 {
	A := int64(a)
	if A < b {
		return A
	}
	return b
}

func minUint64(a uint64, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func fileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	return fi.Size()
}

func BoundsPrint(start, stop uint64) string {
	return fmt.Sprintf("%10d %10d", start, stop)
}

// Index at each level of the proof tree.
// Ex: 33 will be [0:0, 1:0, 2:0, 3:1, 4:2, 5:4, 6:8, 7:16]
func FindTreeGPS(k uint64, L int) []uint64 {

	yCoordinate := make([]uint64, L)
	for i := 0; i < L; i++ {
		k = k >> 1
		yCoordinate[L-i-1] = k
	}
	return yCoordinate
}

func ReverseSliceBool(a []bool) []bool {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
	return a
}

func ReverseSliceUint64(a []uint64) []uint64 {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
	return a
}

func SliceIsEqual(a, b []mcl.G1) bool {
	var status bool
	if len(a) != len(b) {
		return false
	}
	status = true
	for i := 0; i < len(a); i++ {
		status = status && a[i].IsEqual(&b[i])
		if !status {
			fmt.Printf("Failed at %d index out of %d\n", i, len(a)-1)
			return status
		}
	}
	return true
}
