package genesys

import (
	"fmt"
	"math"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
)

func GetMosTotals(config *platformclientv2.Configuration, startDate, endDate, timeZone string) (message, mosHTML, rFactorHTML string, err error) {
	jobID, err := getMosCreateJob(config, startDate, endDate, timeZone)
	if err != nil {
		return fmt.Sprintf("Error creating job: %v\n", err), "", "", err
	}

	jobStatus, err := getMosGetJobStatus(config, jobID)
	if err != nil {
		return fmt.Sprintf("Error getting job status: %v\n", err), "", "", err
	}

	if jobStatus == "FULFILLED" {
		results, err := getMosGetJobResults(config, jobID, "")
		if err != nil {
			return fmt.Sprintf("Error getting job results: %v\n", err), "", "", err
		}
		mosHTML, rFactorHTML := getMosFormatResults(results)
		return "Completed GetMosTotals", mosHTML, rFactorHTML, nil
	} else {
		return fmt.Sprintf("Unexpected job status: %s\n", jobStatus), "", "", fmt.Errorf("unexpected job status: %s", jobStatus)
	}
}

func getMosCreateJob(config *platformclientv2.Configuration, startDate, endDate, timeZone string) (string, error) {
	utcStartDate, err := convertLocalToUTCStr(startDate, timeZone)
	if err != nil {
		return "", fmt.Errorf("error converting start date to UTC: %v", err)
	}
	utcEndDate, err := convertLocalToUTCStr(endDate, timeZone)
	if err != nil {
		return "", fmt.Errorf("error converting end date to UTC: %v", err)
	}
	intervalStr := utcStartDate + "/" + utcEndDate
	apiInstance := platformclientv2.NewConversationsApiWithConfig(config)
	data, _, err := apiInstance.PostAnalyticsConversationsDetailsJobs(platformclientv2.Asyncconversationquery{
		Interval: &intervalStr,
		ConversationFilters: &[]platformclientv2.Conversationdetailqueryfilter{
			{
				VarType: new("and"),
				Predicates: &[]platformclientv2.Conversationdetailquerypredicate{{
					VarType:   new("dimension"),
					Operator:  new("exists"),
					Dimension: new("mediaStatsMinConversationMos"),
				}},
			},
		},
	})
	if err != nil {
		return "", err
	} else {
		return *data.JobId, nil
	}
}

func getMosGetJobStatus(config *platformclientv2.Configuration, jobID string) (string, error) {
	apiInstance := platformclientv2.NewConversationsApiWithConfig(config)
	for {
		data, _, err := apiInstance.GetAnalyticsConversationsDetailsJob(jobID)
		if err != nil {
			return "", err
		}
		state := *data.State

		switch state {
		case "FULFILLED":
			return "FULFILLED", nil
		case "FAILED":
			return "FAILED", fmt.Errorf("job failed on server")
		default:
			time.Sleep(2 * time.Second)
		}
	}
}

func getMosGetJobResults(config *platformclientv2.Configuration, jobID, cursor string) ([]platformclientv2.Analyticsconversation, error) {
	apiInstance := platformclientv2.NewConversationsApiWithConfig(config)
	var allConversations []platformclientv2.Analyticsconversation
	for {
		data, _, err := apiInstance.GetAnalyticsConversationsDetailsJobResults(jobID, cursor, 1000)
		if err != nil {
			return nil, err
		}
		if data.Conversations != nil {
			allConversations = append(allConversations, *data.Conversations...)
		}
		if data.Cursor == nil || *data.Cursor == "" {
			break
		}
		cursor = *data.Cursor
	}
	return allConversations, nil
}

func getMosFormatResults(results []platformclientv2.Analyticsconversation) (string, string) {

	totalMosInteractions := 0
	totalRFactorInteractions := 0
	maxMos := 0.00
	totalMos := 0.00
	rFactorMax := 0.00
	totalRFactor := 0.00
	minMos := 4.99
	rFactorMin := 99.99
	mosColor := "green"
	rFactorColor := "green"

	for _, conversation := range results {
		if conversation.MediaStatsMinConversationMos != nil {
			totalMosInteractions++
			mos := math.Round(*conversation.MediaStatsMinConversationMos*100) / 100
			if mos < minMos {
				minMos = mos
			}
			if mos > maxMos {
				maxMos = mos
			}
			totalMos += mos
		}
		if conversation.MediaStatsMinConversationRFactor != nil && *conversation.MediaStatsMinConversationRFactor > 0 {
			totalRFactorInteractions++
			rFactor := math.Round(*conversation.MediaStatsMinConversationRFactor*100) / 100
			if rFactor < rFactorMin {
				rFactorMin = rFactor
			}
			if rFactor > rFactorMax {
				rFactorMax = rFactor
			}
			totalRFactor += rFactor
		}
	}

	averageMos := math.Round((totalMos/float64(len(results)))*100) / 100
	averageRFactor := math.Round((totalRFactor/float64(len(results)))*100) / 100

	if minMos < 3.5 {
		mosColor = "red"
	}
	if rFactorMin < 50 {
		rFactorColor = "red"
	}

	mosHTML := fmt.Sprintf("'<div style=\"display: inline-flex; width: 100%%; justify-content: space-evenly;\"><h5>MOS Total Interactions: %d</h5><h5>Min: <strong style=\"color: %s\"> ▼ %.2f</strong></h5><h5>Max: <strong  style=\"color: green\"> ▲ %.2f</strong></h5><h5>Average: 📈 %.2f</h5></div>'", totalMosInteractions, mosColor, minMos, maxMos, averageMos)
	rFactorHTML := fmt.Sprintf("'<div style=\"display: inline-flex; width: 100%%; justify-content: space-evenly;\"><h5>R-Factor Total Interactions: %d</h5><h5>Min: <strong style=\"color: %s\"> ▼ %.2f</strong></h5><h5>Max: <strong  style=\"color: green\"> ▲ %.2f</strong></h5><h5>Average: 📈 %.2f</h5></div>'", totalRFactorInteractions, rFactorColor, rFactorMin, rFactorMax, averageRFactor)

	return mosHTML, rFactorHTML
}
