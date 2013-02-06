package main

import (
	. "math"
)

func min(a, b, c int) int {
	m := Min(float64(a), float64(b))
	return int(Min(m, float64(c)))
}

func dist(s, t string) int {
	/*
	   LevenshteinDistance(char s[1..m], char t[1..n])
	   for all i and j, d[i,j] will hold the Levenshtein distance between
	   the first i characters of s and the first j characters of t
	   note that d has (m+1)x(n+1) values
	*/

	m := len(s) + 1
	n := len(t) + 1
	var d [][]int

	d = make([][]int, m)
	for i := 0; i < m; i++ {
		d[i] = make([]int, n)
	}

	for i := 0; i < m; i++ {
		d[i][0] = i // the distance of any first string to an empty second string
	}
	for j := 0; j < n; j++ {
		d[0][j] = j // the distance of any second string to an empty first string
	}

	for j := 1; j < n; j++ {
		for i := 1; i < m; i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1] // no operation required
			} else {
				d[i][j] = min(
					d[i-1][j]+1,   // a deletion
					d[i][j-1]+1,   // an insertion
					d[i-1][j-1]+1) // a substitution
			}
		}
	}

	return d[m-1][n-1]
}

func Fuzzy(s, t string) float64 {
	d := float64(dist(s, t))
	if d == 0 {
		return 1
	}

	l := Max(float64(len(s)), float64(len(t)))
	return (l - d) / l
}
