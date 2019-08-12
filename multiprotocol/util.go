package multiprotocol

import (
	"errors"
	"math"
	"sort"
)

func searchLimit(CurrLimit int64, limitArray []int64, flowArray []int64, TotalFlow int64) (limit int64, err error) {
	limit = 0
	if len(limitArray) == 0 {
		err = errors.New("limitArray size is 0")
		return
	}
	if len(flowArray) == 0 {
		err = errors.New("flowArray size is 0")
		return
	}
	if len(limitArray) != len(flowArray) {
		err = errors.New("limitArray != flowArray")
		return
	}
	index := sort.Search(len(flowArray), func(i int) bool {
		return flowArray[i] > TotalFlow
	})
	if index != 0 {
		limit = limitArray[index-1]
	}
	//
	if limit > 0 && CurrLimit > 0 {
		return int64(math.Min(float64(limit), float64(CurrLimit))), nil
	} else if limit > 0 && CurrLimit == 0 {
		return limit, nil
	} else if limit == 0 && CurrLimit > 0 {
		return CurrLimit, nil
	} else if limit == 0 && CurrLimit == 0 {
		return 0, nil
	}
	return
}
