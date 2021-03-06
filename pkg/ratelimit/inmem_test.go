package ratelimit

import (
	"fmt"
	"testing"
)

func printIMRL(imrl *InMemRateLimit) {
	fmt.Printf("IMRL:\n")
	fmt.Printf("   Window: %v\n", imrl.Window)
	fmt.Printf("   Start: %d\n", imrl.StartIdx)
	fmt.Printf("   TmSec: %d\n", imrl.TimestampSec)
	fmt.Printf("   Sum: %d\n", imrl.Sum)
	fmt.Printf("\n")
}

func Test_InMemRateLimit(t *testing.T) {

	imrl := NewInMemRateLimit(120)
	if imrl.ReqPerMin != 120 {
		t.Errorf("reqPerMin, want: 120, got: %d", imrl.ReqPerMin)
		return
	}

	// basic increment at time 1
	imrl.Inc(1)

	if imrl.Sum != 1 {
		t.Errorf("want imrl.Sum = 1, got %d", imrl.Sum)
		return
	}

	printIMRL(imrl)
	scw := imrl.GetSlidingCountersWindow(1)
	if scw.Sum != 1 {
		t.Errorf("Sum, want: 1, got %d", scw.Sum)
		return
	}

	scw.Print()
	// fmt.Printf("Wnd: %v\n", scw.Window)

	// increment at the end of the buffer, should have both
	// values
	imrl.Inc(60)
	printIMRL(imrl)
	if imrl.Sum != 2 {
		t.Errorf("want imrl.Sum = 2, got %d (%v)", imrl.Sum,
			imrl.Window)
		return
	}

	scw = imrl.GetSlidingCountersWindow(60)
	if scw.Sum != 2 {
		t.Errorf("Sum, want: 2, got %d", scw.Sum)
		return
	}
	scw.Print()

	scw = imrl.GetSlidingCountersWindow(119)
	if scw.Sum != 1 {
		// we should have just one at the start
		t.Errorf("Sum, want: 1, got %d", scw.Sum)
		return
	}
	scw.Print()

	scw = imrl.GetSlidingCountersWindow(120)
	if scw.Sum != 0 {
		// we should have just one at the start
		t.Errorf("Sum, want: 0, got %d", scw.Sum)
		return
	}
	scw.Print()
}
