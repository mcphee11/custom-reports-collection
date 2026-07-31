package genesys

import (
	"fmt"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
)

func GetRecordingTotals(config *platformclientv2.Configuration, startDate, endDate, timeZone, mediaType string) (message string, tableData int, err error) {
	jobID, err := getRecordingTotalsCreateJob(config, startDate, endDate, timeZone, mediaType)
	if err != nil {
		return fmt.Sprintf("Error creating job: %v\n", err), 0, err
	}

	jobStatus, err := getRecordingTotalsGetJobStatus(config, jobID)
	if err != nil {
		return fmt.Sprintf("Error getting job status: %v\n", err), 0, err
	}

	if jobStatus == "FULFILLED" {
		results, err := getRecordingTotalsGetJobResults(config, jobID, "")
		if err != nil {
			return fmt.Sprintf("Error getting job results: %v\n", err), 0, err
		}
		formattedResults := getRecordingTotalsFormatResults(results)
		return "Completed GetRecordingTotals: " + mediaType, formattedResults, nil
	} else {
		return fmt.Sprintf("Unexpected job status: %s\n", jobStatus), 0, fmt.Errorf("unexpected job status: %s", jobStatus)
	}
}

func getRecordingTotalsCreateJob(config *platformclientv2.Configuration, startDate, endDate, timeZone, mediaType string) (string, error) {
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
		SegmentFilters: &[]platformclientv2.Segmentdetailqueryfilter{
			{
				VarType: new("and"),
				Predicates: &[]platformclientv2.Segmentdetailquerypredicate{{
					Dimension: new("recording"),
					Operator:  new("exists"),
				},
					{
						Dimension: new("mediaType"),
						Value:     new(mediaType),
					},
				},
			}},
	})
	if err != nil {
		return "", err
	} else {
		return *data.JobId, nil
	}
}

func getRecordingTotalsGetJobStatus(config *platformclientv2.Configuration, jobID string) (string, error) {
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

func getRecordingTotalsGetJobResults(config *platformclientv2.Configuration, jobID, cursor string) ([]platformclientv2.Analyticsconversation, error) {
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

func getRecordingTotalsFormatResults(results []platformclientv2.Analyticsconversation) int {
	count := len(results)
	return count
}
