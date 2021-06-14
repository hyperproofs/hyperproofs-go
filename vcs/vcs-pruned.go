// All the functionalities when only pruned proof and pruned UPK tree is available.
// UPKs are technically not sent around as a tree, as UPK is fixed prefetching all UPKs for all indices of interest saves the time to fetch UPK from its tree.
// Same trick does not work for proof trees as we need to update them often.
package vcs

import (
	"github.com/alinush/go-mcl"
)

// Proof 1...l
// Index 0 has one 0-variables, index L - 1 has L-1 variable
func (vcs *VCS) GetProofPathDB(proofTree []map[uint64]mcl.G1, index uint64) []mcl.G1 {

	proof := make([]mcl.G1, vcs.L)
	id := index
	for j := uint8(0); j < vcs.L; j++ {
		id = id >> 1 // Hacky. Need to write docs for this.
		proof[j] = proofTree[vcs.L-j-1][id]
	}
	return proof
}

func (vcs *VCS) UpdateComVecDB(upk_db map[uint64][]mcl.G1, digest mcl.G1, updateindex []uint64, delta []mcl.Fr) mcl.G1 {
	N := len(updateindex)
	// if N != len(delta) {
	// 	fmt.Print("UpdateComVec: Error")
	// }
	var result mcl.G1
	var temp mcl.G1
	upks := make([]mcl.G1, N)
	for i := 0; i < N; i++ {
		upks[i] = upk_db[updateindex[i]][vcs.L-1]
	}

	mcl.G1MulVec(&temp, upks, delta)
	mcl.G1Add(&result, &digest, &temp)
	return result
}

// This is analogous to the UpdateProofTree. UpdateProofTree uses the struct variable ProofTree to import all updates.
// Where as this code updates a pruned Proof tree.
// For ell = 30, it is not possible to store the entire proof tree in memory.
// Thus a pruned proof tree is only stored in-memory.
func (vcs *VCS) UpdateProofTreeBulkDB(proofTree []map[uint64]mcl.G1, upk_db map[uint64][]mcl.G1, updateindexVec []uint64, deltaVec []mcl.Fr) ([]map[uint64]mcl.G1, int) {

	var q_i mcl.G1

	// Temporary variables
	var x, upk_i uint8 // These will serve as GPS in the tree
	var y uint64       // These will serve as GPS in the tree

	g1Db := make(map[TreeGPS][]mcl.G1)
	frDb := make(map[TreeGPS][]mcl.Fr)

	for t := range updateindexVec {
		updateindex := updateindexVec[t]
		delta := deltaVec[t]

		updateindexBinary := ToBinary(updateindex, vcs.L)       // LSB first
		updateindexBinary = ReverseSliceBool(updateindexBinary) // MSB first

		upk := upk_db[updateindex]                         // Pop upk_{u,l} as it containts ell variables.
		upk = append([]mcl.G1{vcs.G}, upk[:len(upk)-1]...) // Since the root of the proof tree contains only ell - 1 variables, we need to pop the upk.

		L := vcs.L
		Y := FindTreeGPS(updateindex, int(L))
		// Start from the top of the prooftree (which implies start from the bottom of the UPK tree)
		for i := uint8(0); i < L; i++ {
			x = i
			y = Y[i]
			upk_i = L - i - 1
			loc := TreeGPS{x, y}

			q_i = upk[upk_i]

			if !updateindexBinary[x] {
				mcl.G1Neg(&q_i, &q_i)
			}
			g1Db[loc] = append(g1Db[loc], q_i)
			frDb[loc] = append(frDb[loc], delta)
		}
	}

	q_i.Clear() // Re-init the variable
	for key := range g1Db {
		A := g1Db[key]
		B := frDb[key]

		proof_i := proofTree[key.level][key.index]

		mcl.G1MulVec(&q_i, A, B)
		var tmp mcl.G1
		mcl.G1Add(&tmp, &proof_i, &q_i)
		proofTree[key.level][key.index] = tmp
	}
	basecost := len(g1Db)
	return proofTree, basecost
}
