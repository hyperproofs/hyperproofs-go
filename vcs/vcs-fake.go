package vcs

import (
	"math/rand"

	"github.com/alinush/go-mcl"
)

// Generate the trapdoors
// Generate the keys for aggregation
// PRK and UPK is generated only during runtime using GenUpkFake
func (vcs *VCS) KeyGenFake(ncores uint8, L uint8, folder string, txnLimit uint64) {

	NCORES = ncores // Maximum number threads created. Set this to number of available cores.
	vcs.Init(L, folder, txnLimit)
	vcs.TrapdoorsGen()
	vcs.GenAggGipa()
}

func (vcs *VCS) KeyGenLoadFake(ncores uint8, L uint8, folder string, txnLimit uint64) {
	NCORES = ncores
	vcs.Init(L, folder, txnLimit)
	vcs.LoadTrapdoor(L)
	// vcs.LoadAggGipa() // No need to load this. We'll explicitly load this during every run.
}

// Given an index in the vector, get its upk. This goes from top of the tree to leaf.
func (vcs *VCS) GenUpkFake(index uint64) []mcl.G1 {
	L := vcs.L
	upk := make([]mcl.G1, L)
	for i := uint8(0); i < L; i++ {
		exp := vcs.SelectUPK(L-i, index)
		mcl.G1Mul(&upk[L-i-1], &vcs.G, &exp)
	}
	return upk
}

// Goal is to generate a proof tree using trapdoors
func (vcs *VCS) GenProofsTreeFake(count uint64) (mcl.G1, []uint64, []mcl.Fr, map[uint64][]mcl.G1, [][]mcl.G1, []map[uint64]mcl.G1) {
	N := int64(vcs.N)
	indexVec := make([]uint64, count)
	fakeQTree := make([]map[uint64]mcl.Fr, vcs.L) // Each slice index is level and each position contains DB of sub-indices.
	proofTree := make([]map[uint64]mcl.G1, vcs.L) // Each slice index is level and each position contains DB of sub-indices.
	proofs_db := make(map[uint64][]mcl.G1)        // Misnomer. It is a DB. Each vector index is a key and values are the proof (from 0th quotient).
	a_i := make([]mcl.Fr, count)
	var f_a mcl.Fr
	var digest mcl.G1

	// Genrate random indices
	for k := uint64(0); k < count; k++ {
		id := uint64(rand.Int63n(N)) // Pick an index from the saved list of vector positions. Could contain duplicates as well.
		indexVec[k] = id
	}

	for l := uint8(0); l < vcs.L; l++ {
		fakeQTree[l] = make(map[uint64]mcl.Fr)
		proofTree[l] = make(map[uint64]mcl.G1)
	}

	f_a.Random()
	mcl.G1Mul(&digest, &vcs.G, &f_a)

	// Set f_a. Populate random quotients and compute the vector element.
	for k := uint64(0); k < count; k++ {
		var id uint64
		var rhs mcl.Fr
		id = indexVec[k]
		binary := ToBinary(id, vcs.L)
		proofs := make([]mcl.G1, vcs.L)
		rhs.SetInt64(0)

		for l := uint8(0); l < vcs.L; l++ {
			id = id >> 1
			q, ok := fakeQTree[vcs.L-l-1][id]
			pi, _ := proofTree[vcs.L-l-1][id]
			if !ok {
				q.Random()
				fakeQTree[vcs.L-l-1][id] = q

				mcl.G1Mul(&pi, &vcs.G, &q)
				proofTree[vcs.L-l-1][id] = pi
			}

			proofs[l] = pi

			var tmp mcl.Fr
			if binary[l] {
				mcl.FrMul(&tmp, &q, &vcs.trapdoorsSubOneRev[l])
			} else {
				mcl.FrMul(&tmp, &q, &vcs.trapdoors[l])
			}
			mcl.FrAdd(&rhs, &rhs, &tmp)

		}
		mcl.FrSub(&a_i[k], &f_a, &rhs)

		_, ok := proofs_db[indexVec[k]]
		if !ok {
			proofs_db[indexVec[k]] = proofs
		}
	}

	// Build the upk db for the subvector. Key: Position in the vector. Value is the UPK vector for that index.
	upk_db := make(map[uint64][]mcl.G1)
	for k := uint64(0); k < count; k++ {
		_, ok := upk_db[indexVec[k]]
		if !ok {
			upk_i := vcs.GenUpkFake(indexVec[k])
			upk_db[indexVec[k]] = upk_i
		}
	}

	proofVec := GetProofVecFromDb(proofs_db, indexVec)
	return digest, indexVec, a_i, upk_db, proofVec, proofTree
}
