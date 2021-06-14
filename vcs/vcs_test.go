package vcs

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/alinush/go-mcl"
	"github.com/hyperproofs/gipa-go/batch"
)

// Basic unit test cases for testing VCS functionality.
func TestVCS(t *testing.T) {

	// This method assumes that VRK, UPK are already computed and saved on disk.
	// Be sure to run the run vcs.KeyGen if PRK, VRK, UPK keys are not computed.

	mcl.InitFromString("bls12-381")
	fmt.Println("Curve order", mcl.GetCurveOrder())
	// Get K random positions in the tree
	var L uint8
	K := 21 // Number of transactions
	txnLimit := uint64(K)
	ell := []uint8{16}

	for loop := range ell {
		L = ell[loop]
		N := uint64(1) << L

		vcs := VCS{}
		vcs.KeyGenLoad(16, L, "../pkvk-17", txnLimit)

		indexVec := make([]uint64, K)   // List of indices that chanaged (there can be duplicates.)
		proofVec := make([][]mcl.G1, K) // Proofs of the changed indices.
		deltaVec := make([]mcl.Fr, K)   // Magnitude of the changes.
		valueVec := make([]mcl.Fr, K)   // Current value in that position.

		var digest mcl.G1
		var status bool

		{
			aFr := GenerateVector(N)
			digest = vcs.Commit(aFr, uint64(L))
			vcs.OpenAll(aFr)

			for k := 0; k < K; k++ {
				indexVec[k] = uint64(rand.Intn(int(N))) // Can contain duplicates
				proofVec[k] = vcs.GetProofPath(indexVec[k])
				deltaVec[k].Random()
				valueVec[k] = aFr[indexVec[k]]
			}
		}

		t.Run(fmt.Sprintf("%d/VerifyNaive;", L), func(t *testing.T) {
			status = true
			var loc uint64

			for k := 0; k < K; k++ {
				loc = indexVec[k]
				// status = status && vcs.Verify(digest, loc, valueMap[loc], proofVec[k])
				status = status && vcs.Verify(digest, loc, valueVec[k], proofVec[k])
				if status == false {
					t.Errorf("Verification Failed")
				}
			}
		})

		t.Run(fmt.Sprintf("%d/VerifyMemoized;", L), func(t *testing.T) {
			status = true
			status, _ = vcs.VerifyMemoized(digest, indexVec, valueVec, proofVec)
			if status == false {
				t.Errorf("Fast Verification Failed")
			}
		})

		// Make some changes to the vector positions.
		for k := 0; k < K; k++ {
			loc := indexVec[k]
			delta := deltaVec[k]
			vcs.UpdateProofTree(loc, delta)
		}

		// Update the value vector
		valueVec = SecondaryStateUpdate(indexVec, deltaVec, valueVec)

		// Get latest proofs
		for k := 0; k < K; k++ {
			proofVec[k] = vcs.GetProofPath(indexVec[k])
		}

		digest = vcs.UpdateComVec(digest, indexVec, deltaVec)

		t.Run(fmt.Sprintf("%d/UpdateProofTree;", L), func(t *testing.T) {
			status = true
			status, _ = vcs.VerifyMemoized(digest, indexVec, valueVec, proofVec)
			if status == false {
				t.Errorf("UpdateProofTree Failed")
			}
		})

		vcs.UpdateProofTreeBulk(indexVec, deltaVec)

		// Update the value vector
		valueVec = SecondaryStateUpdate(indexVec, deltaVec, valueVec)

		// Get latest proofs
		for k := 0; k < K; k++ {
			proofVec[k] = vcs.GetProofPath(indexVec[k])
		}
		digest = vcs.UpdateComVec(digest, indexVec, deltaVec)

		t.Run(fmt.Sprintf("%d/UpdateProofTreeBulk;", L), func(t *testing.T) {
			status = true
			status, _ = vcs.VerifyMemoized(digest, indexVec, valueVec, proofVec)
			if status == false {
				t.Errorf("UpdateProofTreeBulk Failed")
			}
		})

		var aggProof batch.Proof
		aggProof = vcs.AggProve(indexVec, proofVec)

		t.Run(fmt.Sprintf("%d/AggregateVerify;%d", L, txnLimit), func(t *testing.T) {

			status = status && vcs.AggVerify(aggProof, digest, indexVec, valueVec)
			if status == false {
				t.Errorf("Aggregation failed")
			}
		})

		// Simple do another round of updates to check if aggregated succeeded
		vcs.UpdateProofTreeBulk(indexVec, deltaVec)
		valueVec = SecondaryStateUpdate(indexVec, deltaVec, valueVec)
		for k := 0; k < K; k++ {
			proofVec[k] = vcs.GetProofPath(indexVec[k])
		}
		digest = vcs.UpdateComVec(digest, indexVec, deltaVec)

		var aggIndex []uint64
		var aggProofIndv [][]mcl.G1
		var aggValue []mcl.Fr

		aggIndex = make([]uint64, txnLimit)
		aggProofIndv = make([][]mcl.G1, txnLimit)
		aggValue = make([]mcl.Fr, txnLimit)

		for j := uint64(0); j < txnLimit; j++ {
			id := uint64(rand.Intn(int(K))) // Pick an index from the saved list of vector positions
			aggIndex[j] = indexVec[id]
			aggProofIndv[j] = proofVec[id]
			aggValue[j] = valueVec[id]
		}

		aggProof = vcs.AggProve(aggIndex, aggProofIndv)
		t.Run(fmt.Sprintf("%d/AggregateVerify2;%d", L, txnLimit), func(t *testing.T) {

			status = status && vcs.AggVerify(aggProof, digest, aggIndex, aggValue)
			if status == false {
				t.Errorf("Aggregation#2 failed")
			}
		})

	}
}

func SecondaryStateUpdate(indexVec []uint64, deltaVec []mcl.Fr, valueVec []mcl.Fr) []mcl.Fr {

	K := len(indexVec)
	valueMap := make(map[uint64]mcl.Fr)  // loc: Current value in that position.
	updateMap := make(map[uint64]mcl.Fr) // loc: Magnitude of the changes.

	for k := 0; k < K; k++ {
		valueMap[indexVec[k]] = valueVec[k]
	}

	// Make some changes to the vector positions.
	for k := 0; k < K; k++ {
		loc := indexVec[k]
		delta := deltaVec[k]
		temp := updateMap[loc]
		mcl.FrAdd(&temp, &temp, &delta)
		updateMap[loc] = temp
	}

	// Import the bunch of changes made to local slice of aFr
	for key, value := range updateMap {
		temp := valueMap[key]
		mcl.FrAdd(&temp, &temp, &value)
		valueMap[key] = temp
	}

	// Update the value vector
	for k := 0; k < K; k++ {
		valueVec[k] = valueMap[indexVec[k]]
	}

	return valueVec
}
