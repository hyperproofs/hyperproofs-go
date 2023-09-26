package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/QED-it/go-jubjub/pkg/pedersenhash"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"golang.org/x/crypto/blake2b"
)

func benchmarkHashing(b *testing.B, kind string, logarraysize int) {

	datasize := (1 << (logarraysize)) * 20
	b.Run(fmt.Sprintf("%s;%d", kind, logarraysize), func(b *testing.B) {
		bytes := make([]byte, datasize)
		b.ResetTimer()
		for bn := 0; bn < b.N; bn++ {
			b.StopTimer()
			_, err := rand.Read(bytes)
			if err != nil {
				panic(fmt.Sprintf("Error randomly generating data: %v\n", err))
			}
			b.StartTimer()
			if kind == "Blake2b256" {
				_ = blake2b.Sum256(bytes)
			} else if kind == "Poseidon" {
				poseidon.HashBytes(bytes)
			}
		}
	})
}

func benchmarkPoseidon(b *testing.B) {

	datasize := 32
	b.Run(fmt.Sprintf("Poseidon;"), func(b *testing.B) {
		bytes := make([]byte, datasize)
		b.ResetTimer()
		for bn := 0; bn < b.N; bn++ {
			b.StopTimer()
			_, _ = rand.Read(bytes)
			A, _ := poseidon.HashBytes(bytes)
			_, _ = rand.Read(bytes)
			B, _ := poseidon.HashBytes(bytes)
			array := []*big.Int{A, B}
			b.StartTimer()

			poseidon.Hash(array)
		}
	})
}

func benchmarkBlake2b256(b *testing.B) {

	datasize := 32
	b.Run(fmt.Sprintf("Blake2b256;"), func(b *testing.B) {
		bytes := make([]byte, datasize)
		b.ResetTimer()
		for bn := 0; bn < b.N; bn++ {
			b.StopTimer()
			_, _ = rand.Read(bytes)
			A := blake2b.Sum256(bytes)
			_, _ = rand.Read(bytes)
			B := blake2b.Sum256(bytes)
			array := make([]byte, 0, 2*len(A))
			array = append(array, A[:]...)
			array = append(array, B[:]...)
			b.StartTimer()

			_ = blake2b.Sum256(bytes)
		}
	})
}

func convert_bytes_to_bool(byteArray []byte) []bool {
	X := make([]bool, len(byteArray)*8)
	for _, in := range byteArray {
		for i := 7; i > 0; i-- {
			if in&(1<<i) > 0 {
				X[i] = true
			}
		}
	}
	return X
}

func benchmarkPedersenHash(b *testing.B) {

	datasize := 32
	bytesX := make([]byte, datasize/2)
	bytesY := make([]byte, datasize/2)
	_, _ = rand.Read(bytesX)
	_, _ = rand.Read(bytesY)
	X := convert_bytes_to_bool(bytesX)
	Y := convert_bytes_to_bool(bytesY)

	hasher, _ := pedersenhash.NewPedersenHasher()
	p, _ := hasher.PedersenHashForBits(X, Y)
	bytesX = p.X().Bytes()
	bytesY = p.Y().Bytes()

	b.Run(fmt.Sprintf("Pedersen128;"), func(b *testing.B) {
		b.ResetTimer()
		for bn := 0; bn < b.N; bn++ {
			b.StopTimer()
			bytesX = p.X().Bytes()
			bytesY = p.Y().Bytes()
			X = convert_bytes_to_bool(bytesX)
			Y = convert_bytes_to_bool(bytesY)
			b.StartTimer()
			p, _ = hasher.PedersenHashForBits(X, Y)
		}
	})
}

func BenchmarkHashing(b *testing.B) {
	logarraysize := []int{6, 7}
	for i := range logarraysize {
		benchmarkHashing(b, "Poseidon", logarraysize[i])
		benchmarkHashing(b, "Blake2b256", logarraysize[i])
	}
	benchmarkPoseidon(b)
	benchmarkBlake2b256(b)
	// benchmarkPedersenHash(b)
}
