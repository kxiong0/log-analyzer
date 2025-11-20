package parser

import (
	"encoding/json"
	"fmt"
	common "log-analyzer/internal/common"
)

func ParseJson(b []byte) ([]common.LogEvent, error) {
	var logs []common.LogEvent
	err := json.Unmarshal(b, &logs)
	if err != nil {
		return logs, err
	}

	fmt.Println("Num of parsed les", len(logs))
	return logs, nil
}
