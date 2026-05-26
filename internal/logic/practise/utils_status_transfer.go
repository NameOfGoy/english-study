package practise

import (
	"english-study/internal/model/bean"
	"english-study/internal/types"
	"fmt"
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

// statusTransferFSM 状态转移FSM。
// 如果 ws.Status 或 op 不在 FSM 表里，返回 error 而不是 panic。
func statusTransferFSM(ws *bean.WordStatus, op int, rule *Rule) (int, error) {
	inner, ok := statusFSM[ws.Status]
	if !ok {
		return ws.Status, fmt.Errorf("statusTransferFSM: unknown status %d", ws.Status)
	}
	fn, ok := inner[op]
	if !ok {
		return ws.Status, fmt.Errorf("statusTransferFSM: unknown operation %d for status %d", op, ws.Status)
	}
	return fn(ws, rule), nil
}
