package genesys

import (
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
	"strings"
	"time"
)

func GetInteractionsPerDay(config *platformclientv2.Configuration, startDate, endDate, timeZone string) (message string, voiceChart, messageChart, emailChart, callbackChart []Bar, err error) {
	jobID, err := getInteractionsPerDayCreateJob(config, startDate, endDate, timeZone)
	if err != nil {
		return fmt.Sprintf("Error creating job: %v\n", err), nil, nil, nil, nil, err
	}

	jobStatus, err := getConversationTotalsGetJobStatus(config, jobID)
	if err != nil {
		return fmt.Sprintf("Error getting job status: %v\n", err), nil, nil, nil, nil, err
	}

	if jobStatus == "FULFILLED" {
		results, err := getConversationTotalsGetJobResults(config, jobID, "")
		if err != nil {
			return fmt.Sprintf("Error getting job results: %v\n", err), nil, nil, nil, nil, err
		}
		voiceChart, messageChart, emailChart, callbackChart, err := getInteractionsPerDayFormatResults(results, startDate, endDate, timeZone)
		if err != nil {
			return fmt.Sprintf("Error formatting results: %v\n", err), nil, nil, nil, nil, err
		}
		return "Completed GetInteractionsPerDay", voiceChart, messageChart, emailChart, callbackChart, nil
	} else {
		return fmt.Sprintf("Unexpected job status: %s\n", jobStatus), nil, nil, nil, nil, fmt.Errorf("unexpected job status: %s", jobStatus)
	}
}

func getInteractionsPerDayCreateJob(config *platformclientv2.Configuration, startDate, endDate, timeZone string) (string, error) {
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
	data, _, err := apiInstance.PostAnalyticsConversationsAggregatesJobs(platformclientv2.Conversationasyncaggregationquery{
		Interval:    &intervalStr,
		Granularity: new("P1D"),
		TimeZone:    new(timeZone),
		GroupBy:     &[]string{"mediaType"},
		Metrics:     &[]string{"nConnected"},
		Filter: &platformclientv2.Conversationaggregatequeryfilter{
			VarType: new("and"),
			Predicates: &[]platformclientv2.Conversationaggregatequerypredicate{
				{
					Dimension: new("conversationId"),
					Operator:  new("exists"),
				},
			},
		},
	})
	if err != nil {
		return "", err
	} else {
		return *data.JobId, nil
	}
}

func getInteractionsPerDayGetJobStatus(config *platformclientv2.Configuration, jobID string) (string, error) {
	apiInstance := platformclientv2.NewConversationsApiWithConfig(config)
	for {
		data, _, err := apiInstance.GetAnalyticsConversationsAggregatesJob(jobID)
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

func getInteractionsPerDayGetJobResults(config *platformclientv2.Configuration, jobID, cursor string) ([]platformclientv2.Conversationaggregatedatacontainer, error) {
	apiInstance := platformclientv2.NewConversationsApiWithConfig(config)
	var allResults []platformclientv2.Conversationaggregatedatacontainer
	for {
		data, _, err := apiInstance.GetAnalyticsConversationsAggregatesJobResults(jobID, cursor)
		if err != nil {
			return nil, err
		}
		if data.Results != nil {
			allResults = append(allResults, *data.Results...)
		}
		if data.Cursor == nil || *data.Cursor == "" {
			break
		}
		cursor = *data.Cursor
	}
	return allResults, nil
}

func getInteractionsPerDayFormatResults(results []platformclientv2.Conversationaggregatedatacontainer, startDate, endDate, timeZone string) ([]Bar, []Bar, []Bar, []Bar, error) {
	voiceChart := []Bar{}
	messageChart := []Bar{}
	emailChart := []Bar{}
	callbackChart := []Bar{}
	layout := "2006-01-02T15:04:05"

	dateToIndex := make(map[string]int)
	idx := 0

	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	start, err := time.ParseInLocation(layout, startDate, loc)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	end, err := time.ParseInLocation(layout, endDate, loc)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// build dates in arrays
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		dateKey := day.Format("2006-01-02")
		dateToIndex[dateKey] = idx

		voiceChart = append(voiceChart, Bar{Key: day.Format("2006-01-02"), Count: 0})
		messageChart = append(messageChart, Bar{Key: day.Format("2006-01-02"), Count: 0})
		emailChart = append(emailChart, Bar{Key: day.Format("2006-01-02"), Count: 0})
		callbackChart = append(callbackChart, Bar{Key: day.Format("2006-01-02"), Count: 0})
		idx++
	}

	for _, result := range results {
		mediaType := (*result.Group)["mediaType"]

		for _, interval := range *result.Data {
			rawDate := strings.Split(*interval.Interval, "/")[0]
			t, _ := time.Parse(time.RFC3339, rawDate)
			dayKey := t.In(loc).Format("2006-01-02")

			if i, ok := dateToIndex[dayKey]; ok {
				for _, metric := range *interval.Metrics {
					if *metric.Metric == "nConnected" {
						val := int(*metric.Stats.Count)
						switch mediaType {
						case "voice":
							voiceChart[i].Count += val
						case "message":
							messageChart[i].Count += val
						case "email":
							emailChart[i].Count += val
						case "callback":
							callbackChart[i].Count += val
						}
					}
				}
			}
		}
	}

	return voiceChart, messageChart, emailChart, callbackChart, nil
}
