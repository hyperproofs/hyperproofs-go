package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"testing"

	"github.com/alinush/go-mcl"
	"github.com/dustin/go-humanize"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func snarks_verifier() {
	var db map[string]float64
	db = make(map[string]float64)
	BenchmarkSnarkVerifierBinaryFieldElements(&db)
	BenchmarkSnarkVerifierRandomFieldElements(&db)
	// keys, _ := getKeyValues(db)

	json, err := json.Marshal(db)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("./plots/benchmarking-snarks-verifier.json", json, 0666)
	if err != nil {
		panic(err)
	}
	fmt.Println("Data saved to: ./plots/benchmarking-snarks-verifier.json")
	macro_merkle_snark_driver() // TODO merge all the micro and macro results to a single JSON file.
}

func Summary(size uint64, op string, aux string, r *testing.BenchmarkResult) {

	// a := time.Duration(r.NsPerOp() / int64(size))
	// out := fmt.Sprintf("Time per %s (%d iters%s):", op, r.N, aux)
	// fmt.Printf("%-60s %20v\n", out, a)

	p := message.NewPrinter(language.English)
	a := float64(r.NsPerOp()) / float64(size) / float64(1000) // Convert ns to us
	out := fmt.Sprintf("Time per %s (%s%d iters):", op, aux, r.N)
	p.Printf("%-60s %20.3f us\n", out, a)
}

// Merkle proof aggregation of N leaves:
// During verification the public inputs are:
// 1. N leaf values
// 2. Each vector index requires log N bits. Thus, N log N inputs are required for checking indices. Not that exponents are binary.
// 3. Old index requires 1 input
// 4. New index requires 1 input
func BenchmarkSnarkVerifierBinaryFieldElements(db *map[string]float64) {

	var size []uint64

	size = []uint64{8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384} // Number of Merkle proofs

	for i := 0; i < len(size); i++ {
		N := size[i]
		logN := uint64(math.Log2(float64(N)))

		M := N + N*logN + 1
		baseG1_elements := generateG1(N + 1)
		expoFr_elements := generateFr(N + 1)

		baseG1_indices := generateG1(N * logN)
		expoFr_indices := generateBinaryFr(N * logN)
		// baseG1_indices := generateG1(N)
		// expoFr_indices := generateFr(N)

		baseG1 := append(baseG1_elements, baseG1_indices...)
		expoFr := append(expoFr_elements, expoFr_indices...)

		P, Q := generate_pairing_data(4) // 1 out 4 pairing can be precomputed, I think, for Groth16.

		fmt.Println("Done generating the data. N =", N, "M =", M)

		var results testing.BenchmarkResult

		// =============================================
		results = testing.Benchmark(func(t *testing.B) {
			var result mcl.G1
			var out mcl.GT
			t.ResetTimer()
			for i := 0; i < t.N; i++ {
				mcl.G1MulVec(&result, baseG1, expoFr)
				mcl.G1Neg(&P[0], &P[0])
				mcl.MillerLoopVec(&out, P, Q)
				mcl.FinalExp(&out, &out)
			}
		})

		Summary(1, "G1MulVecBinary", fmt.Sprintf("size %s; ", humanize.Comma(int64(M))), &results)
		Summary(M, "G1MulVecBinary", fmt.Sprintf("per exp; "), &results)
		(*db)[fmt.Sprintf("%d;%d;G1MulVecBinary", N, M)] = float64(results.NsPerOp())
		(*db)[fmt.Sprintf("%d;%d;G1MulVecBinaryAvg", N, M)] = float64(results.NsPerOp()) / float64(M)
		fmt.Println(sep_string(""))
		// =============================================
	}
}

func BenchmarkSnarkVerifierRandomFieldElements(db *map[string]float64) {

	var size []uint64

	size = []uint64{8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384} // Number of Merkle proofs

	for i := 0; i < len(size); i++ {
		N := size[i]
		// logN := uint64(math.Log2(float64(N)))

		M := N + N + 1
		baseG1_elements := generateG1(N + 1)
		expoFr_elements := generateFr(N + 1)

		baseG1_indices := generateG1(N)
		expoFr_indices := generateFr(N)

		baseG1 := append(baseG1_elements, baseG1_indices...)
		expoFr := append(expoFr_elements, expoFr_indices...)

		P, Q := generate_pairing_data(4) // 1 out 4 pairing can be precomputed, I think, for Groth16.

		fmt.Println("Done generating the data. N =", N, "M =", M)

		var results testing.BenchmarkResult

		// =============================================
		results = testing.Benchmark(func(t *testing.B) {
			var result mcl.G1
			var out mcl.GT
			t.ResetTimer()
			for i := 0; i < t.N; i++ {
				mcl.G1MulVec(&result, baseG1, expoFr)
				mcl.G1Neg(&P[0], &P[0])
				mcl.MillerLoopVec(&out, P, Q)
				mcl.FinalExp(&out, &out)
			}
		})

		Summary(1, "G1MulVecRandom", fmt.Sprintf("size %s; ", humanize.Comma(int64(M))), &results)
		Summary(M, "G1MulVecRandom", fmt.Sprintf("per exp; "), &results)
		(*db)[fmt.Sprintf("%d;%d;G1MulVecRandom", N, M)] = float64(results.NsPerOp())
		(*db)[fmt.Sprintf("%d;%d;G1MulVecRandomAvg", N, M)] = float64(results.NsPerOp()) / float64(M)
		fmt.Println(sep_string(""))
		// =============================================
	}
}

func generateG1(count uint64) []mcl.G1 {
	base := make([]mcl.G1, count)
	for i := uint64(0); i < count; i++ {
		base[i].Random()
	}
	return base
}

func generateFr(count uint64) []mcl.Fr {
	base := make([]mcl.Fr, count)
	for i := uint64(0); i < count; i++ {
		base[i].Random()
	}
	return base
}

func generateBinaryFr(count uint64) []mcl.Fr {
	base := make([]mcl.Fr, count)
	for i := uint64(0); i < count; i++ {
		b := rand.Uint64()
		if b%2 == 0 {
			base[i].SetInt64(0)
		} else {
			base[i].SetInt64(1)
		}
	}
	return base
}

func sep_string(in string) string {
	return fmt.Sprintf("%s=============================================", in)
}

func generate_pairing_data(N int) ([]mcl.G1, []mcl.G2) {

	P := make([]mcl.G1, N)
	Q := make([]mcl.G2, N)
	for i := range P {
		P[i].Random()
		Q[i].Random()
	}
	return P, Q
}

func macro_merkle_snark_driver() {
	db := make(map[string]float64)
	BenchmarkStatelessSnarkVerifierBinaryFieldElements(&db)

	json, err := json.Marshal(db)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("./plots/benchmarking-snarks-verifier-macro.json", json, 0666)
	if err != nil {
		panic(err)
	}
	fmt.Println("Data saved to: ./plots/benchmarking-snarks-verifier-macro.json")
}

// Stateless setting(macro benchmarking): (different from the above microbenchmarking settings) Merkle proof aggregation of N leaves:
// During verification the public inputs are:
// 1. old digest
// 2. New digest
// 3. (sender's index, receiver's index, delta) Note that indices will be in binary form
func BenchmarkStatelessSnarkVerifierBinaryFieldElements(db *map[string]float64) {

	var size []uint64

	size = []uint64{8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384} // Number of Merkle proofs

	for i := 0; i < len(size); i++ {
		N := size[i]
		logN := uint64(math.Log2(float64(N)))

		M := N + 2*N*logN + 2
		baseG1_elements := generateG1(N + 2) // N deltas and 2 digest
		expoFr_elements := generateFr(N + 2) // N deltas and 2 digest

		baseG1_indices := generateG1(2 * N * logN)       // alice's indices in binary and bob's indices in binary
		expoFr_indices := generateBinaryFr(2 * N * logN) // alice's indices in binary and bob's indices in binary
		// baseG1_indices := generateG1(N)
		// expoFr_indices := generateFr(N)

		baseG1 := append(baseG1_elements, baseG1_indices...)
		expoFr := append(expoFr_elements, expoFr_indices...)

		P, Q := generate_pairing_data(4) // 1 out 4 pairing can be precomputed, I think, for Groth16.

		fmt.Println("Done generating the data. N =", N, "M =", M)

		var results testing.BenchmarkResult

		// =============================================
		results = testing.Benchmark(func(t *testing.B) {
			var result mcl.G1
			var out mcl.GT
			t.ResetTimer()
			for i := 0; i < t.N; i++ {
				mcl.G1MulVec(&result, baseG1, expoFr)
				mcl.G1Neg(&P[0], &P[0])
				mcl.MillerLoopVec(&out, P, Q)
				mcl.FinalExp(&out, &out)
			}
		})

		Summary(1, "G1MulVecBinaryMacro", fmt.Sprintf("size %s; ", humanize.Comma(int64(M))), &results)
		Summary(M, "G1MulVecBinaryMacro", fmt.Sprintf("per exp; "), &results)
		(*db)[fmt.Sprintf("%d;%d;G1MulVecBinaryMacro", N, M)] = float64(results.NsPerOp())
		(*db)[fmt.Sprintf("%d;%d;G1MulVecBinaryMacroAvg", N, M)] = float64(results.NsPerOp()) / float64(M)
		fmt.Println(sep_string(""))
		// =============================================
	}
}
