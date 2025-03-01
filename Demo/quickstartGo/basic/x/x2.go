package main

import (
	"fmt"
	"math/big"
	"time"
)

func main() {

	allInputs := [5]int64{10000, 50000, 100000, 500000, 1000000}

	var i int = 1
	for i = 0; i < len(allInputs); i++ {

		var before = GetTimeStamp()

		var factorial *big.Int = big.NewInt(1)
		var j int64 = 1
		for j = 1; j <= allInputs[i]; j++ {
			factorial = factorial.Mul(factorial, big.NewInt(j))
		}
		var after = GetTimeStamp()
		var elapsedTime = after - before
		fmt.Println("Elapsed Time for %s is %s", allInputs[i], elapsedTime)
	}
}

func GetTimeStamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
