package main

import (
	"fmt"
	"math/big"
)

func calculateF(n int) int {
	if n < 0 {
		return 0
	}

	// Jumlahkan menggunakan big.Int untuk menghindari overflow
	factorial := big.NewInt(1)
	for i := 1; i <= n; i++ {
		factorial.Mul(factorial, big.NewInt(int64(i)))
	}

	// Hitung 2^n menggunakan big.Int
	powerOfTwo := big.NewInt(1)
	powerOfTwo.Lsh(powerOfTwo, uint(n)) // ubah posisi ke kiri sebanyak n = 2^n

	// ubah ke big.Float untuk pembagian
	factorialFloat := new(big.Float).SetInt(factorial)
	powerOfTwoFloat := new(big.Float).SetInt(powerOfTwo)

	// Bagi n! dengan 2^n
	result := new(big.Float).Quo(factorialFloat, powerOfTwoFloat)

	// Pasangkan fungsi ceiling
	resultInt := new(big.Int)
	result.Int(resultInt)

	// cek apakah perlu dibulatkan ke atas (ceiling)
	resultFloat := new(big.Float).SetInt(resultInt)
	if result.Cmp(resultFloat) > 0 {
		resultInt.Add(resultInt, big.NewInt(1))
	}

	// ubah kembali ke int (asumsi hasil muat dalam int)
	return int(resultInt.Int64())
}

func main() {
	// Uji fungsi dengan beberapa contoh
	for n := 0; n <= 10; n++ {
		result := calculateF(n)
		fmt.Printf("f(%d) = %d\n", n, result)
	}
}
