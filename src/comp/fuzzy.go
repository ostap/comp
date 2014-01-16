// Copyright (c) 2013 Julius Chrobak. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

type Fuzzy struct {
	m, n int
	d    [][]int
}

func min(a, b, c int) int {
	min := c
	if b < c {
		min = b
	}
	if a < min {
		min = a
	}
	return min
}

func max(a, b int) int {
	max := b
	if a > b {
		max = a
	}
	return max
}

func (f *Fuzzy) dist(left, right string) int {
	s := []rune(left)
	t := []rune(right)

	/*
	   LevenshteinDistance(char s[1..m], char t[1..n])
	   for all i and j, d[i,j] will hold the Levenshtein distance between
	   the first i characters of s and the first j characters of t
	   note that d has (m+1)x(n+1) values
	*/

	m := len(s) + 1
	n := len(t) + 1

	if f.m < m || f.n < n {
		f.m = m
		f.n = n

		f.d = make([][]int, m)
		for i := 0; i < m; i++ {
			f.d[i] = make([]int, n)
		}

		for i := 0; i < m; i++ {
			f.d[i][0] = i // the distance of first string to an empty second string
		}
		for j := 0; j < n; j++ {
			f.d[0][j] = j // the distance of second string to an empty first string
		}
	}

	for j := 1; j < n; j++ {
		for i := 1; i < m; i++ {
			if s[i-1] == t[j-1] {
				f.d[i][j] = f.d[i-1][j-1] // no operation required
			} else {
				f.d[i][j] = min(
					f.d[i-1][j]+1,   // a deletion
					f.d[i][j-1]+1,   // an insertion
					f.d[i-1][j-1]+1) // a substitution
			}
		}
	}

	return f.d[m-1][n-1]
}

func (f *Fuzzy) Compare(left, right string) float64 {
	d := float64(f.dist(left, right))
	if d == 0 {
		return 1
	}

	s := []rune(left)
	t := []rune(right)
	l := float64(max(len(s), len(t)))
	return (l - d) / l
}
