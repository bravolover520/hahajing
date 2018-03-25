package door

const (
	douBanCode   = 0
	mtimeCode    = 1
	guardCodeNbr = 2
)

const codeRound = 6 // one in @codeRound will select DouBan

// limits of request per time interval for each site code
var doorReqLimits = []int{40, 200}
var doorReqLimitTimes = []int64{60, 60} // second

// Guard x
type Guard struct {
	reqs  [guardCodeNbr][]int64 // [time]
	round int
}

func (g *Guard) canPass(t int64, code int) (int, bool) {
	times := g.reqs[code]
	limit := doorReqLimits[code]
	limitTime := doorReqLimitTimes[code]
	count := 1 // assume this one is added.
	i := len(times) - 1
	for ; i >= 0; i-- {
		if t-times[i] > limitTime {
			break
		}

		count++
		if count > limit {
			return 0, false
		}
	}

	return i, true // -1: empty
}

func (g *Guard) add(t int64) (int, bool) {
	// get round
	rounds := []int{douBanCode, mtimeCode}
	g.round = g.round % codeRound
	if g.round != 0 {
		rounds = []int{mtimeCode, douBanCode}
	}

	g.round++

	// try one by one
	for _, code := range rounds {
		i, pass := g.canPass(t, code)
		if !pass {
			continue
		}

		times := append(g.reqs[code], t)

		// cut, we don't need more
		if i < 0 {
			i = 0
		}
		g.reqs[code] = times[i:]

		return code, true
	}

	return -1, false
}
