package practise

import (
	"english-study/internal/model/bean"
	"english-study/internal/types"
)

type Rule struct {
	ReviewTimes     int
	StrengthenTimes int
}
type transfer func(s *bean.WordStatus, rule *Rule) int

var statusFSM = map[int]map[int]transfer{
	types.WordStatusStudy: {
		1: func(s *bean.WordStatus, rule *Rule) int { return types.WordStatusReview },
		2: func(s *bean.WordStatus, rule *Rule) int { return types.WordStatusStudy }},
	types.WordStatusReview: {
		1: finishReview,
		2: func(s *bean.WordStatus, rule *Rule) int { return types.WordStatusStrengthen }},
	types.WordStatusStrengthen: {
		1: finishStrengthen,
		2: func(s *bean.WordStatus, rule *Rule) int { return types.WordStatusStrengthen }},
	types.WordStatusFinish: {
		1: func(s *bean.WordStatus, rule *Rule) int { return types.WordStatusFinish },
		2: func(s *bean.WordStatus, rule *Rule) int { return types.WordStatusStrengthen },
	},
}

func finishReview(ws *bean.WordStatus, rule *Rule) int {
	// SRS: finish when interval >= 30 days AND repetitions >= 4
	if ShouldFinish(ws.Interval, ws.Repetitions) {
		return types.WordStatusFinish
	}
	return types.WordStatusReview
}

func finishStrengthen(ws *bean.WordStatus, rule *Rule) int {
	if rule == nil {
		return types.WordStatusReview
	}
	if ws.Times >= rule.StrengthenTimes {
		return types.WordStatusReview
	}
	return types.WordStatusStrengthen
}

// statusTransferFSM 状态转移FSM
func statusTransferFSM(ws *bean.WordStatus, op int, rule *Rule) int {
	return statusFSM[ws.Status][op](ws, rule)
}
