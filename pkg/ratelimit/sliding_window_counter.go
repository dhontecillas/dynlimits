package ratelimit

import "fmt"

type SlidingCountersWindow struct {
	TimestampSec int64
	Window       [60]int64
	Sum          int64
	ReqPerMin    int64
}

// Print debug info for the window
func (scw *SlidingCountersWindow) Print() {
	fmt.Printf("ReqPerMin: %d\nSum: %d\nWindow: %v\n",
		scw.ReqPerMin, scw.Sum, scw.Window)
}

// NumEmptySlotsAtStarts gives the number of empty seconds
// at the start of the counting window that is the same
// than the minimum number of seconds that must pass
// before new requests are added to the minute window
func (scw *SlidingCountersWindow) NumEmptySlotsAtStart() int {
	for idx, v := range scw.Window {
		if v != 0 {
			return idx
		}
	}
	return 60
}
