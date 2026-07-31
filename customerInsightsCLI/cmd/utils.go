// Package genesys provides access to the platform API
package genesys

import "time"

type Bar struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type UsageTable struct {
	Type       string `json:"type"`
	CX1        string `json:"cx1"`
	CX2        string `json:"cx2"`
	CX3        string `json:"cx3"`
	BuiltCount int    `json:"builtCount"`
	UsedCount  int    `json:"usedCount"`
	InUse      string `json:"inUse"`
}

func convertLocalToUTCStr(localStr, localZoneStr string) (string, error) {
	layout := "2006-01-02T15:04:05"
	loc, _ := time.LoadLocation(localZoneStr)
	t, err := time.ParseInLocation(layout, localStr, loc)
	if err != nil {
		return "", err
	}
	return t.UTC().Format(time.RFC3339), nil
}
