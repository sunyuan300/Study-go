package main

import (
	"fmt"
	"strconv"
	"strings"
)

func main() {
	fmt.Println(restoreIpAddresses("000"))
}

func restoreIpAddresses(s string) []string {
	var res []string
	backtrack(s, 0, []string{}, &res)
	return res
}

func backtrack(s string, start int, track []string, res *[]string) {
	if len(track) == 4 {
		sum := 0
		for _, v := range track {
			sum += len(v)
		}
		if sum == len(s) {
			tmp := make([]string, len(track))
			copy(tmp, track)
			*res = append(*res, strings.Join(tmp, "."))
		}
		return
	}

	if len(track) > 4 {
		return
	}

	for i := start; i < len(s); i++ {
		if check(s, start, i) {
			track = append(track, s[start:i+1])
		} else {
			continue
		}
		backtrack(s, i+1, track, res)
		track = track[:len(track)-1]
	}
}

func check(s string, left, right int) bool {
	if right > left && s[left] == '0' {
		return false
	}
	i, _ := strconv.Atoi(s[left : right+1])
	if i > 255 {
		return false
	}
	return true
}
