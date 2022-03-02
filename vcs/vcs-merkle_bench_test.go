package vcs

import (
	"fmt"
	"testing"

	"github.com/alinush/go-mcl"
	"github.com/hyperproofs/gipa-go/batch"
)

// Benchmark the aggregation for ell = 30.
// These results are used for baseline comparison with Merkle aggregation using SNARKs.
func BenchmarkVCSAgg(b *testing.B) {

	mcl.InitFromString("bls12-381")
	fmt.Println("Curve order", mcl.GetCurveOrder())
	var L uint8

	ell := []uint8{30}
	txnExpo := []uint8{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	txns := make([]uint64, len(txnExpo))
	for i := range txnExpo {
		txns[i] = uint64(1) << txnExpo[i]
	}

	for loop := range ell {
		L = ell[loop]
		// N := uint64(1) << L

		vcs := VCS{}
		vcs.KeyGenLoadFake(16, L, "../pkvk-30", txns[len(txns)-1])

		// digest, indexVec, valueVec, _, proofs_db := vcs.GenProofsFake(txns[0])

		var status bool
		fmt.Println("Num txns: ", txns[len(txns)-1])

		for iTxn := range txns {
			txn := txns[iTxn]
			digest, indexVec, valueVec, _, proofVec, _ := vcs.GenProofsTreeFake(txn)
			vcs.ResizeAgg(txn)
			vcs.LoadAggGipa()

			var aggProof batch.Proof
			var aggProofs []batch.Proof

			b.Run(fmt.Sprintf("%d/AggregateProve;%d", L, txn), func(b *testing.B) {
				for bn := 0; bn < b.N; bn++ {
					aggProof = vcs.AggProve(indexVec[:txn], proofVec[:txn])
					b.StopTimer()
					aggProofs = append(aggProofs, aggProof)
					b.StartTimer()
				}
			})

			status = true
			b.Run(fmt.Sprintf("%d/AggregateVerify;%d", L, txn), func(b *testing.B) {

				for bn := 0; bn < b.N; bn++ {
					aggProof, aggProofs = aggProofs[0], aggProofs[1:]
					status = status && vcs.AggVerify(aggProof, digest, indexVec[:txn], valueVec[:txn])
				}
				if status == false {
					b.Errorf("Aggregation failed")
				}
			})
		}
	}
}
