// Copyright (c) 2012, Jack Christopher Kastorff
// All rights reserved.
// BSD Licensed, see LICENSE for details.

// The alias package picks items from a discrete distribution
// efficiently using the alias method.
package alias

import (
	"errors"
	"math/rand"
)

type Alias struct {
	t []ipiece
	n int32
}

type fpiece struct {
	p float64
	a int32
}

type ipiece struct {
	p int32 // [0,2^31)
	a int32
}

// Create a new alias object.
// For example,
//   var v = alias.New([]float64{8,10,2})
// creates an alias that returns 0 40% of the time, 1 50% of the time, and
// 2 10% of the time.
func New(prob []float64) (*Alias, error) {

	// This implementation is based on
	// http://www.keithschwarz.com/darts-dice-coins/

	n := len(prob)

	if n < 1 {
		return nil, errors.New("Too few probabilities")
	}

	total := float64(0)
	for _, v := range prob {
		if v <= 0 {
			return nil, errors.New("A probability is non-positive")
		}
		total += v
	}

	var al Alias
	al.t = make([]ipiece, n)
	al.n = int32(n)

	// Michael Vose's algorithm

	// "small" stack grows from the bottom of this array
	// "large" stack from the top
	twins := make([]fpiece, n)

	smTop := -1
	lgBot := n

	// invariant: smTop < lgBot, that is, the twin stacks don't collide

	mult := float64(n) / total
	for i, p := range prob {
		p = p * mult

		// push large items (>=1 probability) into the large stack
		// others in the small stack
		if p >= 1 {
			lgBot--
			twins[lgBot] = fpiece{p, int32(i)}
		} else {
			smTop++
			twins[smTop] = fpiece{p, int32(i)}
		}
	}

	for smTop >= 0 && lgBot < n {
		// pair off a small and large block, taking the chunk from the large block that's wanted
		l := twins[smTop]
		smTop--

		g := twins[lgBot]
		lgBot++

		al.t[l.a].p = int32(l.p * (1<<31 - 1))
		al.t[l.a].a = g.a

		g.p = (g.p + l.p) - 1

		// put the rest of the large block back in a list
		if g.p < 1 {
			smTop++
			twins[smTop] = g
		} else {
			lgBot--
			twins[lgBot] = g
		}
	}

	// clear out any remaining blocks
	for i := n - 1; i >= lgBot; i-- {
		al.t[twins[i].a].p = 1<<31 - 1
	}

	// there shouldn't be anything here, but sometimes floating point
	// errors send a probability just under 1.
	for i := 0; i <= smTop; i++ {
		al.t[twins[i].a].p = 1<<31 - 1
	}

	return &al, nil
}

// Generates a random number according to the distribution using the rng passed.
func (al *Alias) Gen(rng *rand.Rand) int32 {
	ri := rng.Int31()
	w := ri % al.n
	if ri > al.t[w].p {
		return al.t[w].a
	}
	return w
}

// TableAlias returns a copy of the alias table used for generation.
func (al *Alias) TableAlias() []int32 {
	t := make([]int32, al.n)
	for i := 0; i < int(al.n); i++ {
		t[i] = al.t[i].a
	}
	return t
}

// TableProb returns a copy of the probability table used for generation
func (al *Alias) TableProb() []int32 {
	t := make([]int32, al.n)
	for i := 0; i < int(al.n); i++ {
		t[i] = al.t[i].p
	}
	return t
}
