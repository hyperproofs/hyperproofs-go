package vcs

import (
	"github.com/alinush/go-mcl"
	"github.com/hyperproofs/gipa-go/batch"
	"github.com/hyperproofs/gipa-go/cm"
	"github.com/hyperproofs/kzg-go/kzg"
)

type VCS struct {
	PRK          []mcl.G1   // PRK is technically not needed for our experiments
	UPK          [][]mcl.G1 // UPK forms a tree. The UPK API in code is different from the paper.
	VRK          []mcl.G2
	VRKSubOne    []mcl.G2 // h^(1-s_1)
	VRKSubOneRev []mcl.G2 // h^(s_1 - 1)

	N uint64
	L uint8
	G mcl.G1 //Generator
	H mcl.G2 //Generator

	trapdoors          []mcl.Fr
	trapdoorsSubOne    []mcl.Fr // (1-s_1)
	trapdoorsSubOneRev []mcl.Fr // (s_1-1)
	pow2               []uint64
	alpha              mcl.Fr // KZG + GIPA
	beta               mcl.Fr // KZG + GIPA

	ProofTree [][]mcl.G1 // Figure 2 from the paper is illustrative of how the ProofTree is saved.
	// Note that lowest level f(w) is not saved in this tree.
	// Proof serving node saves this proof tree all the time. Unable to fit beyond 2^26 in memory.

	// GIPA Stuff
	MN       uint64 // Power of 2 which is nearest to TxnLimit * L. In GIPA notation let M = L, n = TxnLimit = 1024
	ck       cm.Ck  // KZG + GIPA
	TxnLimit uint64 // Number of txns in a block that needs to be aggregated
	nDiff    int64  // These many variables are used to pad the GIPA instance if L * TxnLimit is not a power of 2.
	mnDiff   int64  // These many variables are used to pad the GIPA instance if L * TxnLimit is not a power of 2.

	// KZG Stuff
	kzg1 kzg.KZG1Settings // KZG + GIPA
	kzg2 kzg.KZG2Settings // KZG + GIPA

	folderPath string

	aggProver   batch.Prover
	aggVerifier batch.Verifier

	DISCARD_PRK bool // We do not use: g, g^{s_1}, g^{s_2}, g^{s_1}{s_2}, g^{s_3}.....
	// Thus, PRK is discarded by default
	// UPK tree is enough for the prover
	PARAM_TOO_LARGE bool
}

// Instantiate a new vector commitment instance
// Space for UPK and PRK will be created when keys are created and saved.
// This reduces the memory footprint.
func (vcs *VCS) Init(L uint8, folder string, txnLimit uint64) {

	NFILES = 16
	PRKNAME = "/prk-%02d.data"
	VRKNAME = "/vrk.data"
	TRAPDOORNAME = "/trapdoors.data"
	UPKNAME = "/upk-%02d.data"
	if L == 0 || L >= 32 {
		panic("KeyGen: Error. Either ell is 0 or >= 32")
	}

	vcs.folderPath = folder

	vcs.L = L
	vcs.N = uint64(1) << L

	vcs.pow2 = make([]uint64, L)
	for i := uint8(0); i < L; i++ {
		vcs.pow2[i] = 1 << i
	}

	// Allocate enough space for trapdoors
	vcs.trapdoors = make([]mcl.Fr, vcs.L)
	vcs.trapdoorsSubOne = make([]mcl.Fr, vcs.L)
	vcs.trapdoorsSubOneRev = make([]mcl.Fr, vcs.L)

	// Allocate enough space for VRK
	vcs.VRK = make([]mcl.G2, vcs.L)
	vcs.VRKSubOne = make([]mcl.G2, vcs.L)
	vcs.VRKSubOneRev = make([]mcl.G2, vcs.L)

	// Space for UPK is allocated during load

	vcs.TxnLimit = txnLimit

	if txnLimit*uint64(L) > MAX_AGG_SIZE {
		panic("Try with smaller block size")
	}

	vcs.DISCARD_PRK = true // It is assumed true by default.
	if L > 24 {
		vcs.PARAM_TOO_LARGE = true // When UPK and PRK is large, keys are just flushed to files without keeping it in memory.
	}
}

// Generate trapdoors for the VCS. Be sure to run this after ```Init```.
func (vcs *VCS) TrapdoorsGen() {

	// Need to find a source of randomness and generate trapdoors
	// Need to seed the randomness

	// Sample generators
	vcs.G.Random()
	vcs.H.Random()

	// Generate trapdoors
	var frOne mcl.Fr
	frOne.SetInt64(1)
	for i := range vcs.trapdoors {
		vcs.trapdoors[i].Random()
		mcl.FrSub(&vcs.trapdoorsSubOne[i], &frOne, &vcs.trapdoors[i])
		mcl.FrSub(&vcs.trapdoorsSubOneRev[i], &vcs.trapdoors[i], &frOne)
	}

	// Generate VRK: h^(s_1), h^(s_2), ....
	for i := range vcs.trapdoors {
		mcl.G2Mul(&vcs.VRK[i], &vcs.H, &vcs.trapdoors[i])
	}

	// Generate VRKSubOne: h^(1-s_1), h^(1-s_2), ....
	for i := range vcs.trapdoorsSubOne {
		mcl.G2Mul(&vcs.VRKSubOne[i], &vcs.H, &vcs.trapdoorsSubOne[i])
	}

	// Generate VRKSubOneRev: h^(s_1-1), h^(s_2-1), ....
	for i := range vcs.trapdoorsSubOneRev {
		mcl.G2Mul(&vcs.VRKSubOneRev[i], &vcs.H, &vcs.trapdoorsSubOneRev[i])
	}

	// Generate alpha and beta for KZG
	vcs.alpha.Random()
	vcs.beta.Random()

	vcs.SaveTrapdoor()
}

// Generates PRK VRK UPK etc
// Use this only once to generate the parameters.
func (vcs *VCS) KeyGen(ncores uint8, L uint8, folder string, txnLimit uint64) {

	NCORES = ncores               // Maximum number threads created. Set this to number of available cores.
	vcs.Init(L, folder, txnLimit) //
	vcs.TrapdoorsGen()
	vcs.PrkUpkGen()
	vcs.GenAggGipa()
}

// Defacto entry to VCS.
// Use this to load the files always
func (vcs *VCS) KeyGenLoad(ncores uint8, L uint8, folder string, txnLimit uint64) {
	NCORES = ncores
	vcs.Init(L, folder, txnLimit)
	vcs.LoadTrapdoor(L)
	vcs.PrkUpkLoad()
	vcs.LoadAggGipa()
}

// Do not remove L from the parameters. I am using it OpenAll
func (vcs *VCS) Commit(a []mcl.Fr, L uint64) mcl.G1 {
	var digest mcl.G1
	mcl.G1MulVec(&digest, vcs.UPK[L], a) // Not L - 1 as L = 0 has just vcs.G
	return digest
}

func (vcs *VCS) OpenAllRec(a []mcl.Fr, start uint64, end uint64, L uint8) {

	if end-start <= 1 {
		return
	}

	mid := (start + end) / 2
	bin := end - start
	index := start / bin

	aDiff := make([]mcl.Fr, mid-start)
	for i := uint64(0); i < mid-start; i++ {
		mcl.FrSub(&aDiff[i], &a[i+mid], &a[i+start])
	}

	result := vcs.Commit(aDiff, uint64(L-1))
	vcs.ProofTree[vcs.L-L][index] = result

	vcs.OpenAllRec(a, start, mid, L-1)
	vcs.OpenAllRec(a, mid, end, L-1)
}

func (vcs *VCS) OpenAll(a []mcl.Fr) {

	vcs.ProofTree = make([][]mcl.G1, vcs.L)
	for i := uint8(0); i < vcs.L; i++ {
		vcs.ProofTree[i] = make([]mcl.G1, 1<<i)
	}
	vcs.OpenAllRec(a, 0, vcs.N, vcs.L)
}

// Proof 1...l
// Index 0 has one 0-variables, index L - 1 has L-1 variable
func (vcs *VCS) GetProofPath(index uint64) []mcl.G1 {

	proof := make([]mcl.G1, vcs.L)
	id := index
	for j := uint8(0); j < vcs.L; j++ {
		id = id >> 1 // Hacky. Need to write docs for this.
		proof[j] = vcs.ProofTree[vcs.L-j-1][id]
		// fmt.Println(index, vcs.L-j-1, k)
	}
	return proof
}

type TreeGPS struct {
	level uint8 // Root is level 0
	index uint64
}

func (vcs *VCS) Verify(digest mcl.G1, index uint64, a_i mcl.Fr, proof []mcl.G1) bool {

	if len(proof) != int(vcs.L) {
		panic("Verify: Bad proof!")
	}

	// temp variables
	var rhs mcl.GT
	var p mcl.G1
	var ps []mcl.G1
	var qs []mcl.G2

	ps = make([]mcl.G1, vcs.L+1)
	qs = make([]mcl.G2, vcs.L+1)
	binary := ToBinary(index, vcs.L)
	for i := uint8(0); i < vcs.L; i++ {
		if binary[i] {
			qs[i] = vcs.VRKSubOneRev[i]
		} else {
			qs[i] = vcs.VRK[i]
		}
		ps[i] = proof[i]
	}

	// Move e(digest/g^{a_i}, h) to other side. Thus it will be e(g^{a_i}/digest, h)
	mcl.G1Mul(&p, &vcs.G, &a_i)
	mcl.G1Sub(&p, &p, &digest)
	ps[vcs.L] = p
	qs[vcs.L] = vcs.H

	mcl.MillerLoopVec(&rhs, ps, qs)
	mcl.FinalExp(&rhs, &rhs)
	return rhs.IsOne()
}

func (vcs *VCS) VerifyMemoized(digest mcl.G1, indexVec []uint64, a_i []mcl.Fr, proofVec [][]mcl.G1) (bool, int) {

	// fmt.Println(vcs.VRKSubOneRev[i].IsEqual(&qs[i]), vcs.VRKSubOneRev[i].IsZero())

	if len(proofVec) != len(indexVec) {
		panic("Verify: Bad proof!")
	}

	var p mcl.G1   // temp variables
	var lhs mcl.GT // temp variables
	var tempG2 mcl.G2
	var prod mcl.GT
	var result mcl.GT
	prod.SetInt64(1)

	db := make(map[TreeGPS]mcl.GT)
	status := true
	for t := range proofVec {
		proof := proofVec[t]
		index := indexVec[t]

		binary := ToBinary(index, vcs.L)
		prod.SetInt64(1)

		for i := vcs.L; i > 0; i-- {
			loc := TreeGPS{i, index}

			result1, keyStatus := db[loc]
			if keyStatus == false {
				if binary[vcs.L-i] == true {
					tempG2 = vcs.VRKSubOneRev[vcs.L-i]
				} else {
					tempG2 = vcs.VRK[vcs.L-i]
				}
				// mcl.Pairing(&result, &proof[vcs.L-i], &tempG2)
				mcl.MillerLoop(&result, &proof[vcs.L-i], &tempG2)
				db[loc] = result
				mcl.GTMul(&prod, &prod, &result)
			} else {
				mcl.GTMul(&prod, &prod, &result1)
			}
			index = index >> 1
		}

		mcl.G1Mul(&p, &vcs.G, &a_i[t])
		mcl.G1Sub(&p, &p, &digest)
		mcl.MillerLoop(&lhs, &p, &vcs.H)
		mcl.GTMul(&prod, &prod, &lhs)

		mcl.FinalExp(&prod, &prod)

		status = status && prod.IsOne()
	}
	return status, len(db)
}

func (vcs *VCS) UpdateCom(digest mcl.G1, updateindex uint64, delta mcl.Fr) mcl.G1 {
	var temp mcl.G1
	var result mcl.G1
	mcl.G1Mul(&temp, &vcs.UPK[vcs.L][updateindex], &delta)
	mcl.G1Add(&result, &digest, &temp)
	return result
}

func (vcs *VCS) UpdateComVec(digest mcl.G1, updateindex []uint64, delta []mcl.Fr) mcl.G1 {
	N := len(updateindex)
	// if N != len(delta) {
	// 	fmt.Print("UpdateComVec: Error")
	// }
	var result mcl.G1
	var temp mcl.G1
	upks := make([]mcl.G1, N)
	for i := 0; i < N; i++ {
		upks[i] = vcs.UPK[vcs.L][updateindex[i]]
	}

	mcl.G1MulVec(&temp, upks, delta)
	mcl.G1Add(&result, &digest, &temp)
	return result
}

func (vcs *VCS) UpdateProof(proof []mcl.G1, localindex uint64, updateindex uint64, delta mcl.Fr) []mcl.G1 {

	newProof := make([]mcl.G1, len(proof))
	copy(newProof, proof)
	var temp mcl.G1
	updateindexBinary := ToBinary(updateindex, vcs.L) // LSB first
	localindexBinary := ToBinary(localindex, vcs.L)   // LSB first
	// upk := vcs.UPK[updateindex]
	upk := vcs.GetUpk(updateindex)
	L := int(vcs.L)
	for i := L; i > 0; i-- {
		if i-1 > 0 {
			mcl.G1Mul(&temp, &upk[i-2], &delta)
		} else {
			mcl.G1Mul(&temp, &vcs.G, &delta)
		}
		if updateindexBinary[i-1] == false && localindexBinary[i-1] == true {
			mcl.G1Sub(&newProof[i-1], &proof[i-1], &temp)
			break
		} else if updateindexBinary[i-1] == true && localindexBinary[i-1] == false {
			mcl.G1Add(&newProof[i-1], &proof[i-1], &temp)
			break
		} else if updateindexBinary[i-1] == false && localindexBinary[i-1] == false {
			mcl.G1Sub(&newProof[i-1], &proof[i-1], &temp)
		} else {
			mcl.G1Add(&newProof[i-1], &proof[i-1], &temp)
		}
	}
	return newProof
}

func (vcs *VCS) UpdateProofTree(updateindex uint64, delta mcl.Fr) {

	var q_i mcl.G1
	updateindexBinary := ToBinary(updateindex, vcs.L)       // LSB first
	updateindexBinary = ReverseSliceBool(updateindexBinary) // MSB first

	upk := vcs.GetUpk(updateindex)                     // Pop upk_{u,l} as it containts ell variables.
	upk = append([]mcl.G1{vcs.G}, upk[:len(upk)-1]...) // Since the root of the proof tree contains only ell - 1 variables, we need to pop the upk.

	L := int(vcs.L)
	Y := FindTreeGPS(updateindex, L)

	var x, upk_i int // These will serve as GPS in the tree
	var y uint64     // These will serve as GPS in the tree

	// Start from the top of the prooftree (which implies start from the bottom of the UPK tree)
	for i := 0; i < L; i++ {
		x = i
		y = Y[i]
		upk_i = L - i - 1
		mcl.G1Mul(&q_i, &upk[upk_i], &delta)

		if updateindexBinary[x] {
			mcl.G1Add(&vcs.ProofTree[x][y], &vcs.ProofTree[x][y], &q_i)
		} else {
			mcl.G1Sub(&vcs.ProofTree[x][y], &vcs.ProofTree[x][y], &q_i)
		}
	}
}

func (vcs *VCS) UpdateProofTreeBulk(updateindexVec []uint64, deltaVec []mcl.Fr) int {

	// ProofTree[][]
	var q_i mcl.G1

	var x, upk_i uint8 // These will serve as GPS in the tree
	var y uint64       // These will serve as GPS in the tree

	g1Db := make(map[TreeGPS][]mcl.G1)
	frDb := make(map[TreeGPS][]mcl.Fr)

	for t := range updateindexVec {
		updateindex := updateindexVec[t]
		delta := deltaVec[t]

		updateindexBinary := ToBinary(updateindex, vcs.L)       // LSB first
		updateindexBinary = ReverseSliceBool(updateindexBinary) // MSB first

		upk := vcs.GetUpk(updateindex)                     // Pop upk_{u,l} as it containts ell variables.
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

	for key := range g1Db {
		A := g1Db[key]
		B := frDb[key]
		mcl.G1MulVec(&q_i, A, B)
		mcl.G1Add(&vcs.ProofTree[key.level][key.index], &vcs.ProofTree[key.level][key.index], &q_i)
	}
	return len(g1Db)
}

// g^{(1-s_1)}, g^{(1-s_2)(1-s_1)}, g^{(1-s_3)(1-s_2)(1-s_1)}
func (vcs *VCS) GetUpk(i uint64) []mcl.G1 {

	k := i
	upk := make([]mcl.G1, vcs.L)
	for j := uint8(vcs.L); j > 0; j-- {
		k = k & (^(1 << j)) // Clears the jth bit of k. Technically everything before jth and before has to be cleared.
		upk[j-1] = vcs.UPK[j][k]
	}
	return upk
}
