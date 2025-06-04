//go:build !solution

package hotelbusiness

import (
	"sort"
)

type Guest struct {
	CheckInDate  int
	CheckOutDate int
}

type Load struct {
	StartDate  int
	GuestCount int
}

func ComputeLoad(guests []Guest) []Load {
	loadChng := make(map[int]int)
	for _, g := range guests {
		loadChng[g.CheckInDate]++
		loadChng[g.CheckOutDate]--
	}

	dates := make([]int, 0)
	for k := range loadChng {
		dates = append(dates, k)
	}
	sort.Ints(dates)

	res := make([]Load, 0)
	curLoad := 0
	for _, d := range dates {
		if loadChng[d] == 0 {
			continue
		}
		curLoad += loadChng[d]
		res = append(res, Load{
			StartDate:  d,
			GuestCount: curLoad,
		})
	}

	return res
}
