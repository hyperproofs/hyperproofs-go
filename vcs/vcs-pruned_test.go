package vcs

import (
	"fmt"
	"testing"

	"github.com/alinush/go-mcl"
	"github.com/hyperproofs/gipa-go/batch"
)

// Generate proofs using trapdoor for ell = 30
// Check if the proofs verify.
// Also check if the aggregation works.
func TestVCSPruned(t *testing.T) {

	folderPath := fmt.Sprintf("../pkvk-30")
	L := uint8(30)
	txnscount := uint64(512)
	var status bool

	vcs := VCS{}
	// Besure to have the trapdoors and key generated.
	vcs.KeyGenLoadFake(16, L, folderPath, 1<<12)
	digest, indexVec, valueVec, upk_db, proofVec, proofTree := vcs.GenProofsTreeFake(txnscount)

	for i := range valueVec {
		if valueVec[i].IsZero() {
			fmt.Println("Index:", i, "is zero.")
		}
	}

	// Check indeed if the fake proofs verify.
	status = true
	t.Run(fmt.Sprintf("%d/NaiveVerify;%d", L, txnscount), func(t *testing.T) {
		status, _ = vcs.VerifyMemoized(digest, indexVec, valueVec, proofVec)
		if status == false {
			t.Errorf("Fast Verification Failed")
		}
	})
	fmt.Println("Done checking the proofs.", len(upk_db))

	t.Run(fmt.Sprintf("%d/GetProofPathDB;%d", L, txnscount), func(t *testing.T) {
		for i := range indexVec {
			proofVecChecker := vcs.GetProofPathDB(proofTree, indexVec[i])
			if len(proofVecChecker) != len(proofVec[i]) {
				t.Errorf("Length of proof extracted from DB is not same as the baseline.")
			}
			for j := range proofVecChecker {
				if !proofVecChecker[j].IsEqual(&proofVec[i][j]) {
					out := fmt.Sprintf("Proofs extracted from the DB is not same as the baseline: i: %d indexVec[i]: %d proof[j]: %d", i, indexVec[i], j)
					t.Errorf(out)
				}
			}
		}
	})

	deltaVec := make([]mcl.Fr, len(indexVec))
	for i := range indexVec {
		deltaVec[i].Random()
	}

	digest = vcs.UpdateComVecDB(upk_db, digest, indexVec, deltaVec)
	proofTree, _ = vcs.UpdateProofTreeBulkDB(proofTree, upk_db, indexVec, deltaVec)
	for i := range indexVec {
		proofVec[i] = vcs.GetProofPathDB(proofTree, indexVec[i])
	}

	for i := range valueVec {
		mcl.FrAdd(&valueVec[i], &valueVec[i], &deltaVec[i])
	}

	t.Run(fmt.Sprintf("%d/UpdateComAndTreeBulk;%d", L, txnscount), func(t *testing.T) {
		status, _ = vcs.VerifyMemoized(digest, indexVec, valueVec, proofVec)
		if status == false {
			t.Errorf("UpdateComAndTreeBulk: Fast Verification Failed")
		}
	})

	// Just to be doubly sure, we are making sure that aggregation code is also fine with fake proofs.
	vcs.ResizeAgg(txnscount)
	vcs.LoadAggGipa()

	var aggProof batch.Proof
	var aggProofs []batch.Proof

	aggProof = vcs.AggProve(indexVec[:txnscount], proofVec[:txnscount])
	aggProofs = append(aggProofs, aggProof)

	t.Run(fmt.Sprintf("%d/AggregateVerify;%d", L, txnscount), func(t *testing.T) {

		aggProof, aggProofs = aggProofs[0], aggProofs[1:]
		status = status && vcs.AggVerify(aggProof, digest, indexVec[:txnscount], valueVec[:txnscount])

		if status == false {
			t.Errorf("Aggregation failed")
		}
	})
}
