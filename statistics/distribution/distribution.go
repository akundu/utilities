package distribution

import (
	"math"
	"sort"
	"time"

	rng "github.com/leesper/go_rng"
	"math/rand"
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
	_sorted_result       []int
}

func NewgaussianGenerator(lower_bound, upper_bound, num_to_generate int) *gaussianGenerator {
	result := &gaussianGenerator{
		_grng:                rng.NewGaussianGenerator(time.Now().UnixNano()),
		_lower_bound:         lower_bound,
		_upper_bound:         upper_bound,
		_rounded_power_of_10: closestPowerOf10(upper_bound),
	}
	result.buildNumbers(num_to_generate)
	return result
}
func (this *gaussianGenerator) buildNumbers(num_to_generate int) {
	num_list := make([]int, num_to_generate)
	for i := 0; i < num_to_generate; i++ {
		g_num := this._grng.Gaussian(0, 3)
		num_list[i] = int(g_num * float64(this._rounded_power_of_10))
	}

	sort.Ints(num_list)
	/* make everything positive
	if num_list[0] < 0 {
		amt_to_add_by := (-1 * num_list[0])
		for i, _ := range num_list {
			num_list[i] = num_list[i] + amt_to_add_by
		}
	}
	*/
	this._sorted_result = make([]int, num_to_generate)
	for i, num := range num_list {
		this._sorted_result[i] = mappingValues(num, num_list[0], num_list[len(num_list)-1], this._lower_bound, this._upper_bound)
	}
}
func (this gaussianGenerator) GenerateNumbers() []int {
	return this._sorted_result
}
func (this gaussianGenerator) GenerateNumber() int {
	return this._sorted_result[rand.Int63n(int64(len(this._sorted_result)))]
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
