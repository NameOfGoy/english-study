package types

import (
	"fmt"
	"strconv"
)

func (p PathIDReq) GetUintId() (uint, error) {
	id := p.PathID
	if id == "" {
		return 0, fmt.Errorf("id is empty")
	}
	uid, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("id to uint failed: %w", err)
	}
	return uint(uid), nil
}
