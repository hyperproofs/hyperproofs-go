package vcs

import (
	"fmt"
	"math"

	"github.com/alinush/go-mcl"
	"github.com/hyperproofs/gipa-go/batch"
	"github.com/hyperproofs/gipa-go/cm"
	"github.com/hyperproofs/gipa-go/utils"
)

func (vcs *VCS) GenAggGipa() {

	{
		mn := uint64(MAX_AGG_SIZE) // short circuiting things
		ck, kzg1, kzg2 := cm.IPPSetupKZG(mn, vcs.alpha, vcs.beta, vcs.G, vcs.H)
		cm.IPPSaveCmKzg(ck, kzg1, kzg2, vcs.folderPath)
	}
	vcs.LoadAggGipa()
}

func (self *VCS) LoadAggGipa() {

	L := uint64(self.L)
	limit := L * self.TxnLimit

	self.MN = utils.NextPowOf2(limit)

	self.ck, self.kzg1, self.kzg2 = cm.IPPCMLoadCmKzg(self.MN, self.folderPath)
	self.aggProver = batch.Prover{}
	self.aggVerifier = batch.Verifier{}

	self.nDiff = int64(uint64(math.Ceil(float64(self.MN)/float64(L))) - self.TxnLimit) // This is the size of padding for P and Q vector (gipa)
	self.mnDiff = int64(self.MN - (L * self.TxnLimit))                                 // This is the size of padding for A and B vector (gipa)

	fmt.Println("Size:", len(self.ck.V), len(self.ck.W), len(self.kzg1.PK), len(self.kzg1.VK), len(self.kzg2.PK), len(self.kzg2.VK))
	fmt.Println("padding:", self.nDiff, self.mnDiff)
}

// This resets the variable MN and txnLimit.
// Be sure to load the data from disk
func (self *VCS) ResizeAgg(txnLimit uint64) {
	L := uint64(self.L)
	self.TxnLimit = txnLimit
	limit := L * self.TxnLimit
	self.MN = utils.NextPowOf2(limit)
}

func (vcs *VCS) AggProve(indexVec []uint64, proofVec [][]mcl.G1) batch.Proof {

	var A []mcl.G1
	var B []mcl.G2
	txnLimit := int(vcs.TxnLimit)
	L := int(vcs.L)

	if len(indexVec) != txnLimit || len(proofVec) != txnLimit {
		panic("AggProof: Vectors are not of the expected size")
	}

	for t := range proofVec {
		if len(proofVec[t]) != L {
			panic(fmt.Sprintf("Bad proof: %d", t))
		}
		A = append(A, proofVec[t]...)
	}

	var binary []bool
	b := make([]mcl.G2, vcs.L)
	for t := range indexVec {
		binary = ToBinary(indexVec[t], vcs.L)
		for i := 0; i < L; i++ {
			if binary[i] == true {
				// mcl.G2Sub(&b[i], &vcs.VRK[i], &vcs.H)
				b[i] = vcs.VRKSubOneRev[i]
			} else {
				b[i] = vcs.VRK[i]
			}
		}
		B = append(B, b...)
	}

	aPad := make([]mcl.G1, vcs.mnDiff)
	bPad := make([]mcl.G2, vcs.mnDiff)
	A = append(A, aPad...)
	B = append(B, bPad...)

	vcs.aggProver.Init(uint32(vcs.L), uint32(vcs.TxnLimit+uint64(vcs.nDiff)), vcs.MN, &vcs.ck, &vcs.kzg1, &vcs.kzg2, A, B)

	proof := vcs.aggProver.Prove()
	return proof
}

func (vcs *VCS) AggVerify(proof batch.Proof, digest mcl.G1, indexVec []uint64, a_i []mcl.Fr) bool {

	txnLimit := int(vcs.TxnLimit)
	L := int(vcs.L)

	if len(indexVec) != txnLimit || len(a_i) != txnLimit {
		panic("AggProof: Vectors are not of the expected size")
	}

	P := make([]mcl.G1, txnLimit)
	Q := make([]mcl.G2, txnLimit)
	var p mcl.G1 // temp variables

	for t := range a_i {
		mcl.G1Mul(&p, &vcs.G, &a_i[t])
		mcl.G1Sub(&p, &digest, &p)
		P[t] = p
		Q[t] = vcs.H
	}

	pPad := make([]mcl.G1, vcs.nDiff)
	qPad := make([]mcl.G2, vcs.nDiff)
	P = append(P, pPad...)
	Q = append(Q, qPad...)

	var B []mcl.G2
	var binary []bool
	b := make([]mcl.G2, vcs.L)
	for t := range indexVec {
		binary = ToBinary(indexVec[t], vcs.L)
		for i := 0; i < L; i++ {
			if binary[i] == true {
				// mcl.G2Sub(&b[i], &vcs.VRK[i], &vcs.H)
				b[i] = vcs.VRKSubOneRev[i]
			} else {
				b[i] = vcs.VRK[i]
			}
		}
		B = append(B, b...)
	}
	bPad := make([]mcl.G2, vcs.mnDiff)
	B = append(B, bPad...)

	// fmt.Println("Agg Verifier", L, vcs.N, len(P))
	vcs.aggVerifier.Init(uint32(L), uint32(vcs.TxnLimit+uint64(vcs.nDiff)), vcs.MN, vcs.ck.W, &vcs.kzg1, &vcs.kzg2, P, Q, B)
	status := vcs.aggVerifier.VerifyEdrax(proof)
	// status := vcs.aggVerifier.Verify(proof, P, Q, B)
	return status
}
