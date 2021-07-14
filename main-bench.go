package main

import (
	"fmt"
	"time"

	vc "github.com/hyperproofs/hyperproofs-go/vcs"
)

const FOLDER = "./pkvk-30"

func BenchmarkVCSCommit(L uint8, txnLimit uint64) string {
	N := uint64(1) << L
	K := txnLimit
	vcs := vc.VCS{}
	vcs.KeyGenLoad(16, L, FOLDER, K)

	aFr := vc.GenerateVector(N)
	dt := time.Now()
	vcs.Commit(aFr, uint64(L))
	duration := time.Since(dt)
	out := fmt.Sprintf("BenchmarkVCS/%d/Commit;%d%40d ns/op", L, txnLimit, duration.Nanoseconds())
	fmt.Println(vc.SEP)
	fmt.Println(out)
	fmt.Println(vc.SEP)
	return out
}

func BenchmarkVCSOpenAll(L uint8, txnLimit uint64) string {
	N := uint64(1) << L
	K := txnLimit
	vcs := vc.VCS{}
	vcs.KeyGenLoad(16, L, FOLDER, K)

	aFr := vc.GenerateVector(N)
	dt := time.Now()
	vcs.OpenAll(aFr)
	duration := time.Since(dt)
	out := fmt.Sprintf("BenchmarkVCS/%d/OpenAll;%d%40d ns/op", L, txnLimit, duration.Nanoseconds())
	fmt.Println(vc.SEP)
	fmt.Println(out)
	fmt.Println(vc.SEP)
	return out
}

func Benchmark() {
	var ell []uint8
	var txns []uint64
	var logs []string

	txns = []uint64{1024}

	for itxn := range txns {
		txnLimit := txns[itxn]
		ell = []uint8{10, 20, 22, 24, 26}[:1]
		for i := range ell {
			l := BenchmarkVCSCommit(ell[i], txnLimit)
			logs = append(logs, l)
		}

		ell = []uint8{10, 20, 22, 24}[:1]
		for i := range ell {
			l := BenchmarkVCSOpenAll(ell[i], txnLimit)
			logs = append(logs, l)
		}
	}
	fmt.Println(vc.SEP)
	for iLog := range logs {
		fmt.Println(logs[iLog])
	}
	fmt.Println(vc.SEP)
}
