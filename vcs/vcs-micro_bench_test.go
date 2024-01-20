package vcs

import (
	"fmt"
	"testing"

	"github.com/alinush/go-mcl"
	"github.com/hyperproofs/gipa-go/batch"
)

var ell = []uint8{10, 27, 25, 23, 21} // Change the tree height here

// Microbenchmarks for UpdateAllProofs, Ver ,agg, veragg
// Block size = 1024 transcations.
func BenchmarkPrunedVCSMicro(b *testing.B) {

	mcl.InitFromString("bls12-381")
	fmt.Println("Curve order", mcl.GetCurveOrder())

	txns := []uint64{1024}

	for loop := range ell {
		for iTxn := range txns {
			benchmarkVCS(ell[loop], txns[iTxn], true, true, b)
		}
	}
}

func BenchmarkPrunedVCSMacro(b *testing.B) {

	mcl.InitFromString("bls12-381")
	fmt.Println("Curve order", mcl.GetCurveOrder())

	txns := []uint64{1024}

	for loop := range ell {
		for iTxn := range txns {
			benchmarkVCS(ell[loop], 2*txns[iTxn], false, false, b)
		}
	}
}

func benchmarkVCS(L uint8, txn uint64, DoAgg bool, Micro bool, b *testing.B) {

	var status bool
	var basecost int
	vcs := VCS{}

	vcs.KeyGenLoadFake(16, L, "../pkvk-30", txn)
	digest, indexVec, valueVec, upk_db, proofVec, proofTree := vcs.GenProofsTreeFake(txn)

	deltaVec := make([]mcl.Fr, len(indexVec))
	for i := range indexVec {
		deltaVec[i].Random()
	}

	b.Run(fmt.Sprintf("%d/UpdateComVec;%d", L, txn), func(b *testing.B) {
		b.ResetTimer()
		for bn := 0; bn < b.N; bn++ {
			digest = vcs.UpdateComVecDB(upk_db, digest, indexVec, deltaVec)
			b.StopTimer()
			// Since we are updating the digest, we need to keep track of the changes made as well.

			valueVec = SecondaryStateUpdate(indexVec, deltaVec, valueVec)
			b.StartTimer()
		}
	})

	// Save a copy of the ProofTree so that it can be used for benchmarking UpdateProofTreeBulk
	proofTree_b := make([]map[uint64]mcl.G1, vcs.L)
	for i := range proofTree_b {
		proofTree_b[i] = make(map[uint64]mcl.G1)
		for k, v := range proofTree[i] {
			proofTree_b[i][k] = v
		}
	}

	b.Run(fmt.Sprintf("%d/UpdateProofTreeBulk;%d", L, txn), func(b *testing.B) {
		b.ResetTimer()
		for bn := 0; bn < b.N; bn++ {
			proofTree_b, basecost = vcs.UpdateProofTreeBulkDB(proofTree_b, upk_db, indexVec, deltaVec)
		}
	})
	fmt.Println("UpdateProofTreeBulk (vcs-pruned.go): Unique Nodes:", basecost)

	if Micro {
		for i := range indexVec {
			proofVec[i] = vcs.GetProofPathDB(proofTree_b, indexVec[i])
		}

		b.Run(fmt.Sprintf("%d/VerifyNaive;%d", L, txn), func(b *testing.B) {
			status = true
			for bn := 0; bn < b.N; bn++ {
				for i := range indexVec {
					status = status && vcs.Verify(digest, indexVec[i], valueVec[i], proofVec[i])
					if !status {
						b.Errorf("Naive UpdateProofTree: Naive Verification Failed")
					}
				}
			}
		})

		b.Run(fmt.Sprintf("%d/UpdateProofTreeNaive;%d", L, txn), func(b *testing.B) {
			b.ResetTimer()
			for bn := 0; bn < b.N; bn++ {
				for i := range indexVec {
					proofTree, _ = vcs.UpdateProofTreeBulkDB(proofTree, upk_db, indexVec[i:i+1], deltaVec[i:i+1])
				}
			}
		})

		for i := range indexVec {
			proofVec[i] = vcs.GetProofPathDB(proofTree, indexVec[i])
		}

		b.Run(fmt.Sprintf("%d/VerifyMemoized;%d", L, txn), func(b *testing.B) {
			for bn := 0; bn < b.N; bn++ {
				status, basecost = vcs.VerifyMemoized(digest, indexVec, valueVec, proofVec)
				if !status {
					b.Errorf("Bulk UpdateProofTree: Fast Verification Failed")
				}
			}
		})
		fmt.Println("VerifyMemoized (vcs.go): Unique Nodes:", basecost)
	}

	if DoAgg {
		vcs.ResizeAgg(txn)
		vcs.LoadAggGipa()

		var aggProof batch.Proof
		var aggProofs []batch.Proof

		b.Run(fmt.Sprintf("%d/AggregateProve;%d", L, txn), func(b *testing.B) {
			b.ResetTimer()
			for bn := 0; bn < b.N; bn++ {
				aggProof = vcs.AggProve(indexVec[:txn], proofVec[:txn])
				b.StopTimer()
				aggProofs = append(aggProofs, aggProof)
				b.StartTimer()
			}
		})

		b.Run(fmt.Sprintf("%d/AggregateVerify;%d", L, txn), func(b *testing.B) {
			b.ResetTimer()
			for bn := 0; bn < b.N; bn++ {
				b.StopTimer()
				aggProof, aggProofs = aggProofs[0], aggProofs[1:]
				b.StartTimer()
				status = status && vcs.AggVerify(aggProof, digest, indexVec[:txn], valueVec[:txn])
				if status == false {
					b.Errorf("Aggregation failed")
				}
			}
		})
	}

}
