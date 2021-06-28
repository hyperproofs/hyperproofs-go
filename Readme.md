# Hyperproofs

Hyperproofs, the first vector commitment (VC) scheme that is efficiently maintainable and aggregatable.
This repo contains the implementation of Hyperproofs in go.

This repo depends on:
- [go-mcl](https://github.com/alinush/go-mcl/) for elliptic curve operations.
- [kzg-go](https://github.com/hyperproofs/kzg-go) for KZG commitments.
- [gipa-go](https://github.com/hyperproofs/gipa-go) for proof aggregation.

[hyperproofs]: https://ia.cr/2021/599
## Instructions

0. Run ```time bash scripts/hyper-go.sh``` to setup PRK, VRK, UPK, etc.
1. Run ```time bash scripts/hyper-test.sh``` to run the test cases.
2. Run ```time bash scripts/hyper-bench.sh``` to replicate the benchmarks reported in the [paper][hyperproofs].

## Reference

[_Hyperproofs: Aggregating and Maintaining Proofs in Vector Commitments_][hyperproofs]\
[Shravan Srinivasan](https://github.com/sshravan), Alex Chepurnoy, Charalampos Papamanthou, [Alin Tomescu](https://github.com/alinush), and Yupeng Zhang\
ePrint, 2021
