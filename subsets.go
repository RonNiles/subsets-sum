package main

// how many subsets of {1,...,2000} are there,
// such that the sum of their elements is divisible by 5

import (
	"fmt"
	"math/big"
)

// Think of the problem in terms of columns of elements whose value is the same modulo 5
//     1,    2,    3,    4,    5
//     6,    7,    8,    9,   10
//...
//  1996, 1997, 1998, 1999, 2000

// Each column has 400 entries, each of which can be part of the set or not
// Break each column down by binomial coefficients
// In column 1, each element contributes 1 mod 5 to the sum
//   There are COMBIN(0, 400) ways to choose zero elements. In total these contribute 0 mod 5 to the sum
//   There are COMBIN(1, 400) ways to choose 1 element. In total these contribute 1 mod 5 to the sum
//   There are COMBIN(2, 400) ways to choose 2 elements. In total these contribute 2 mod 5 to the sum
//   There are COMBIN(3, 400) ways to choose 3 elements. In total these contribute 3 mod 5 to the sum
//   There are COMBIN(4, 400) ways to choose 4 elements. In total these contribute 4 mod 5 to the sum
//   There are COMBIN(5, 400) ways to choose 5 elements. In total these contribute 0 mod 5 to the sum
//   There are COMBIN(6, 400) ways to choose 6 elements. In total these contribute 1 mod 5 to the sum
//   ...
// In column 2, each element contributes 2 mod 5 to the sum
//   There are COMBIN(0, 400) ways to choose zero elements. In total these contribute 0 mod 5 to the sum
//   There are COMBIN(1, 400) ways to choose 1 element. In total these contribute 2 mod 5 to the sum
//   There are COMBIN(2, 400) ways to choose 2 elements. In total these contribute 4 mod 5 to the sum
//   There are COMBIN(3, 400) ways to choose 3 elements. In total these contribute 1 mod 5 to the sum
//   There are COMBIN(4, 400) ways to choose 4 elements. In total these contribute 3 mod 5 to the sum
//   There are COMBIN(5, 400) ways to choose 5 elements. In total these contribute 0 mod 5 to the sum
//   There are COMBIN(6, 400) ways to choose 6 elements. In total these contribute 2 mod 5 to the sum
//   ...

// We can sum all these up in a 5x5 array using computeColumnModuloTotals()

// Once we have that 5x5 array we choose one item from each row in all possible ways, adding the column
// numbers mod 5 to find the subsset being contributed to and multiplying the matrix values together
// to find the number of subsets being contributed by the current selection.
//  This is done using computeTotalsRecursively()

const columns = 5
const rows = 400

type binomial struct {
	vals [rows + 1]*big.Int
	sum  *big.Int
}

func (b *binomial) populate() {
	// compute each binomial in sequence
	accum := big.NewInt(1)
	num := big.NewInt(int64(len(b.vals) - 1))
	denom := big.NewInt(1)
	for i, _ := range b.vals {
		// allocate and set the next coefficient to the accumulator value
		val := new(big.Int)
		val.Set(accum)
		b.vals[i] = val
		// Multiply previous coefficient by numerator and divide by denominator
		accum.Mul(accum, num)
		accum.Div(accum, denom)
		// numerator decreases, denominator increase
		denom.Add(denom, big.NewInt(1))
		num.Sub(num, big.NewInt(1))
	}

	// Check that the sum of the binomials equals 2^N
	b.sum = new(big.Int)
	for _, val := range b.vals {
		b.sum.Add(b.sum, val)
	}
	power := big.NewInt(1)
	for n := 1; n <= len(b.vals)-1; n++ {
		power.Mul(power, big.NewInt(2))
	}
	if b.sum.Cmp(power) != 0 {
		panic(fmt.Errorf("bad binomial sum"))
	}
}

// structure that has everything we need to pass during recursion
type recurse struct {
	binom  binomial
	sums   [columns][columns]*big.Int
	mod    int
	accum  *big.Int
	totals [columns]*big.Int
}

func (r *recurse) initialize() {
	// compute binomial coefficients
	r.binom.populate()

	// populate the totals, each entry initializes to zero
	for i, row := range r.sums {
		r.totals[i] = new(big.Int)
		for j, _ := range row {
			r.sums[i][j] = new(big.Int)
		}
	}
}

// Compute the 2x2 modulo totals array
func (r *recurse) computeColumnModuloTotals() {
	// Each column contains values with a constant modulo from zero to columns -1
	for mod := 0; mod < columns; mod++ {
		for k, b := range r.binom.vals {
			// The binomial coefficient 'b' represents how many ways to select 'k' items from this column
			// Each item in this column is 'mod' modulo 'columns' and therefore these k items comtribute k * mod % columns to the sum
			contribution := (k * mod) % columns
			// this next statement is actually just r.sums[mod][contribution] += b (in big.Int semantics)
			r.sums[mod][contribution].Add(r.sums[mod][contribution], b)
		}
	}
}

func (r *recurse) doNextLevel(level int) {
	// If the accumulator is zero, it won't contribute
	if r.accum.Cmp(big.NewInt(0)) == 0 {
		return
	}

	// recursion ends when the level equals the number of columns
	if level == columns {
		// the accumulator has the total ways
		r.totals[r.mod].Add(r.totals[r.mod], r.accum)
		r.accum = nil
		return
	}

	// Go through each column at this level.
	for n := 0; n < columns; n++ {
		// save old
		oldMod := r.mod
		oldAccum := r.accum

		// multiply accumulator by number of subsets of items from column 'level' that have sum 'n' modulo 'columns'
		r.accum = new(big.Int)
		r.accum.Mul(oldAccum, r.sums[level][n])

		// Since these subsets have sum 'n' modulo 'columns' they increase the overall sum by 'n'
		r.mod += n
		if r.mod >= columns {
			r.mod -= columns
		}

		r.doNextLevel(level + 1)

		// restore old
		r.mod = oldMod
		r.accum = oldAccum
	}
}

func (r *recurse) computeTotalsRecursively() {
	// initialize level zero
	r.mod = 0
	r.accum = big.NewInt(1)

	// perform the recursion
	r.doNextLevel(0)

	// Check result: first add the total columns
	sum := big.NewInt(0)
	for _, t := range r.totals {
		sum.Add(sum, t)
	}

	// Total should be 2^rows*columns
	power := big.NewInt(1)
	for n := 1; n <= rows*columns; n++ {
		power.Mul(power, big.NewInt(2))
	}
	if sum.Cmp(power) != 0 {
		panic(fmt.Errorf("bad total sum"))
	}
}

func main() {
	r := &recurse{}
	r.initialize()

	r.computeColumnModuloTotals()

	r.computeTotalsRecursively()

	fmt.Println("Number of subsets whose sum is divisible by", columns, ":")
	fmt.Println(r.totals[0])
}
