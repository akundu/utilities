package distribution

import (
	"math"
	"sort"
	"time"

	rng "github.com/leesper/go_rng"
)

type uniformGenerator struct {
	_urng        *rng.UniformGenerator
	_lower_bound int32
	_upper_bound int32
}

func (this uniformGenerator) GenerateNumber() int {
	return int(this._urng.Int32Range(this._lower_bound, this._upper_bound))
}
func NewuniformGenerator(lower_bound, upper_bound int) *uniformGenerator {
	return &uniformGenerator{
		_urng:        rng.NewUniformGenerator(time.Now().UnixNano()),
		_lower_bound: int32(lower_bound),
		_upper_bound: int32(upper_bound),
	}
}

type gaussianGenerator struct {
	_grng                *rng.GaussianGenerator
	_lower_bound         int
	_upper_bound         int
	_rounded_power_of_10 int
}

func NewgaussianGenerator(lower_bound, upper_bound int) *gaussianGenerator {
	return &gaussianGenerator{
		_grng:                rng.NewGaussianGenerator(time.Now().UnixNano()),
		_lower_bound:         lower_bound,
		_upper_bound:         upper_bound,
		_rounded_power_of_10: closestPowerOf10(upper_bound - 1),
	}
}
func (this gaussianGenerator) GenerateNumbers(num_to_generate int) []int {
	num_list := make([]int, num_to_generate)
	for i := 0; i < num_to_generate; i++ {
		num_list[i] = int(this._grng.Gaussian(0, 3) * float64(this._rounded_power_of_10))
	}

	result := make([]int, num_to_generate)
	sort.Ints(num_list)
	for _, num := range num_list {
		result = append(result, mappingValues(num, num_list[0], num_list[len(num_list)-1], this._lower_bound, this._upper_bound))
	}
	return result
}

func mappingValues(num, old_min, old_max, new_min, new_max int) int {
	old_range := old_max - old_min
	if old_range <= 0 {
		return 0
	}

	new_range := new_max - new_min
	if new_range <= 0 {
		return 0
	}

	return (((new_max - new_min) * (num - old_min)) / (old_max - old_min)) + new_min
}

func closestPowerOf10(num int) int {
	c_power_of_10 := float64(int(math.Log10(float64(num))))
	if math.Pow(10, c_power_of_10) < float64(num) {
		return int(math.Pow(10, c_power_of_10+1))
	}
	return int(math.Pow(10, c_power_of_10))
}
