package ratelimit

// InMemRateLimit implements a circular
// buffer with the requests performed in
// each second of the last minute
type InMemRateLimit struct {
	TimestampSec int64
	StartIdx     int64
	Window       [60]int64
	Sum          int64
	ReqPerMin    int64
}

// NewInMemRateLimit creates a new InMemRateLimit structure
// to keep track of a rate limit
func NewInMemRateLimit(reqPerMin int64) *InMemRateLimit {
	return &InMemRateLimit{
		ReqPerMin: reqPerMin,
	}
}

// UpdateSlidingCountersWindow gets an SlidingCountersWindows from the
// InMemRateLimit at a given current timestamp
func (imrl *InMemRateLimit) UpdateSlidingCountersWindow(toUpdate *SlidingCountersWindow,
	timestampSec int64) {

	// reset the struct
	toUpdate.TimestampSec = timestampSec
	for idx := range toUpdate.Window {
		toUpdate.Window[idx] = 0
	}
	toUpdate.Sum = 0
	toUpdate.ReqPerMin = imrl.ReqPerMin

	if imrl.Sum == 0 {
		return
	}

	from := imrl.TimestampSec - toUpdate.TimestampSec
	to := from + 60
	if from < 0 {
		if to <= 0 {
			return // no overlapping segment
		}
		from = 0
	}
	if to > 60 {
		if from >= 60 {
			return // no overlapping segment
		}
		to = 60
	}

	for i := from; i < to; i++ {
		toUpdate.Window[i] = imrl.Window[(imrl.StartIdx+i)%60]
		toUpdate.Sum += toUpdate.Window[i]
	}
}

// GetSlinginCountersWindows creates a new SlidingCounterWindow for
// a given timestamp in seconds.
func (imrl *InMemRateLimit) GetSlidingCountersWindow(
	atTimestampSec int64) *SlidingCountersWindow {

	scw := SlidingCountersWindow{}
	imrl.UpdateSlidingCountersWindow(&scw, atTimestampSec)
	return &scw
}

// Inc increments the given
func (imrl *InMemRateLimit) Inc(timestampSec int64) {
	advance := timestampSec - imrl.TimestampSec

	// in case we increment a value from the past (that could
	// happen if different threads with the timestampSec from
	// a different time is used, but should be an edge case):
	backSeconds := int64(0)
	if advance < -59 {
		// the increment is before the last minute
		return
	}

	// clean up the new cell between last seen second
	// and the current second
	if advance > 0 {
		imrl.TimestampSec = timestampSec
		if advance > 59 {
			// is a full reset of the circular buffer
			advance = 60
			imrl.StartIdx = 0
		}
		for i := int64(0); i < advance; i++ {
			cellIdx := (i + imrl.StartIdx) % 60
			imrl.Sum -= imrl.Window[cellIdx]
			imrl.Window[cellIdx] = 0
		}
		imrl.StartIdx = (imrl.StartIdx + advance) % 60
	} else {
		backSeconds = advance
	}

	// intial 60 to compensate for negative advance (we only do
	// modulo on positive numbers), and the current time is at
	// the endo of the buffer, so at StartIdx + 59. And we 'advance'
	// from the last timestamp
	incIdx := (60 + imrl.StartIdx + 59 + backSeconds) % 60
	imrl.Window[incIdx] += 1
	imrl.Sum += 1
}

// NumEmptySlotsAtStarts gives the number of empty seconds
// at the start of the counting window that is the same
// than the minimum number of seconds that must pass
// before new requests are added to the minute window
func (imrl *InMemRateLimit) NumEmptySlotsAtStart() int {
	for idx := int64(0); idx < 60; idx++ {
		i := (imrl.StartIdx + idx) % 60
		if imrl.Window[i] != 0 {
			return int(idx)
		}
	}
	return 60
}
