package main

import (
	"fmt"

	"github.com/akundu/utilities/statistics/distribution"
)

func main() {
	rt := distribution.NewgaussianGenerator(0, 100, 200000)
	for _, v := range rt.GenerateNumbers() {
		fmt.Println(v)
	}
}
