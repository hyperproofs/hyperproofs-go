package vcs

import (
	"github.com/alinush/go-mcl"
)

var PRKNAME string
var VRKNAME string
var UPKNAME string
var TRAPDOORNAME string
var NFILES uint8
var NCORES uint8

const MAX_AGG_SIZE = 1 << 19
const SEP = "\n========================================================================================\n"

// Allocate space for UPK.
// This is not done with Init, as UPK and PRK may not fit in memory together.
func (vcs *VCS) MallocUpk() {
	vcs.UPK = make([][]mcl.G1, vcs.L+1)
	for i := uint64(0); i < uint64(vcs.L+1); i++ {
		vcs.UPK[i] = make([]mcl.G1, 1<<i)
	}
}

// Use selector function to compute the powers of g.
// 101 => (s_3 * s_1)
// It is a worker to compute g, g^{s_1}, g^{s_2}, g^{s_1 s_2}, g^{s_3}, , g^{s_3 s_1} ...
func (vcs *VCS) SelectPRK(index uint64) mcl.Fr {
	var prod mcl.Fr
	var out mcl.Fr
	prod.SetInt64(1)
	// fmt.Print(index, " ")
	for i := uint8(0); i < vcs.L; i++ {
		// fmt.Print(index&vcs.pow2[i], " ")
		if index&vcs.pow2[i] != 0 {
			out = vcs.trapdoors[i]
			mcl.FrMul(&prod, &prod, &out)
		}
	}
	// fmt.Println()
	// fmt.Println(prod)
	return prod
}

// Compute the UPK located in a specific coordinate of the tree
// Say 3, 3 => g^{}
func (vcs *VCS) SelectUPK(L uint8, index uint64) mcl.Fr {
	if L > vcs.L {
		panic("Select UPK error")
	}

	var prod mcl.Fr
	var out mcl.Fr
	prod.SetInt64(1)

	for i := uint8(0); i < L; i++ {
		if index&vcs.pow2[i] == 0 {
			out = vcs.trapdoorsSubOne[i]
		} else {
			out = vcs.trapdoors[i]
		}
		mcl.FrMul(&prod, &prod, &out)
	}
	return prod
}

func (vcs *VCS) VerifyUPK(index uint64, upkProof []mcl.G1) bool {

	binary := ToBinary(index, vcs.L)
	var temp1, temp2, result mcl.G1
	var r mcl.Fr
	r.SetInt64(1)
	if binary[0] == false {
		temp1 = vcs.UPK[1][0]
	} else {
		temp1 = vcs.UPK[1][1]
	}

	// Linear Combination is an option
	lhsP := make([]mcl.G1, vcs.L-1)
	lhsQ := make([]mcl.G2, vcs.L-1)
	// rhsP := make([]mcl.G1, vcs.L-1)
	// rhsQ := make([]mcl.G2, vcs.L-1)

	for i := uint8(1); i < vcs.L; i++ {
		r.Random()
		if i == 1 {
			mcl.G1Mul(&lhsP[i-1], &temp1, &r)
		} else {
			mcl.G1Mul(&lhsP[i-1], &upkProof[i-1], &r)
		}

		if binary[i] == false {
			lhsQ[i-1] = vcs.VRKSubOne[i]
		} else {
			lhsQ[i-1] = vcs.VRK[i]
		}

		mcl.G1Mul(&temp2, &upkProof[i], &r)
		mcl.G1Add(&result, &result, &temp2)
	}

	var lhs, rhs mcl.GT
	mcl.Pairing(&rhs, &result, &vcs.H)

	mcl.MillerLoopVec(&lhs, lhsP, lhsQ)
	mcl.FinalExp(&lhs, &lhs)
	return lhs.IsEqual(&rhs)
}

func IsEqual(a, b *VCS) bool {

	if a.N != b.N {
		return false
	}
	if a.L != b.L {
		return false
	}

	if !a.G.IsEqual(&b.G) {
		return false
	}

	if !a.H.IsEqual(&b.H) {
		return false
	}

	if len(a.trapdoors) != len(b.trapdoors) {
		return false
	}

	if len(a.trapdoorsSubOne) != len(b.trapdoorsSubOne) {
		return false
	}

	if len(a.trapdoorsSubOneRev) != len(b.trapdoorsSubOneRev) {
		return false
	}

	if len(a.VRK) != len(b.VRK) {
		return false
	}

	if len(a.VRKSubOne) != len(b.VRKSubOne) {
		return false
	}

	if len(a.VRKSubOneRev) != len(b.VRKSubOneRev) {
		return false
	}

	status := true
	for i := 0; i < int(a.L); i++ {
		status = status && a.trapdoors[i].IsEqual(&b.trapdoors[i])
		status = status && a.trapdoorsSubOne[i].IsEqual(&b.trapdoorsSubOne[i])
		status = status && a.trapdoorsSubOneRev[i].IsEqual(&b.trapdoorsSubOneRev[i])
	}

	for i := 0; i < int(a.L); i++ {
		status = status && a.VRK[i].IsEqual(&b.VRK[i])
		status = status && a.VRKSubOne[i].IsEqual(&b.VRKSubOne[i])
		status = status && a.VRKSubOneRev[i].IsEqual(&b.VRKSubOneRev[i])
		if !status {
			return status
		}
	}
	if !status {
		return status
	}

	if !a.alpha.IsEqual(&b.alpha) {
		return false
	}

	if !a.beta.IsEqual(&b.beta) {
		return false
	}

	if !SliceIsEqual(a.PRK, b.PRK) {
		return false
	}

	if len(a.UPK) != len(b.UPK) {
		return false
	}

	for i := 0; i < len(a.UPK); i++ {
		if !SliceIsEqual(a.UPK[i], b.UPK[i]) {
			return false
		}
	}
	return true
}
