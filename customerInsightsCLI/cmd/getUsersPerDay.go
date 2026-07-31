package genesys

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
)

func GetUsersPerDay(config *platformclientv2.Configuration, startDate, endDate, timeZone string) (string, []byte, error) {
	chunk := new([]platformclientv2.Useraggregatequerypredicate)
	allBar := []Bar{}
	allUsers, err := getAllUsers(config, 1)
	if err != nil {
		return fmt.Sprintf("Error getting users: %v\n", err), nil, err
	}

	for i, userID := range allUsers {
		*chunk = append(*chunk, platformclientv2.Useraggregatequerypredicate{
			Dimension: new("userId"),
			Value:     new(userID),
		})
		// need to group by 100 users to avoid "Dimension value exceeds limit of 100 for dimension userId" error
		if (i+1)%100 == 0 {
			_, usersChart, err := getUsersPerDayCall(config, chunk, startDate, endDate, timeZone)
			if err != nil {
				return fmt.Sprintf("Error processing chunk: %v\n", err), nil, err
			}
			allBar = mergeBars(allBar, usersChart)
			chunk = new([]platformclientv2.Useraggregatequerypredicate)
		}
	}

	// left over users from chunks
	if len(*chunk) > 0 {
		_, usersChart, err := getUsersPerDayCall(config, chunk, startDate, endDate, timeZone)
		if err != nil {
			return fmt.Sprintf("Error processing final chunk: %v\n", err), nil, err
		}
		allBar = mergeBars(allBar, usersChart)
	}

	jsonUsersChart, _ := json.Marshal(allBar)

	return "Completed GetUsersPerDay", jsonUsersChart, nil
}

func mergeBars(allBar, usersChart []Bar) []Bar {
	totals := make(map[string]int)
	for _, b := range allBar {
		totals[b.Key] += b.Count
	}
	for _, b := range usersChart {
		totals[b.Key] += b.Count
	}

	merged := make([]Bar, 0, len(totals))
	dates := make([]string, 0, len(totals))
	for d := range totals {
		dates = append(dates, d)
	}
	sort.Strings(dates)
	for _, d := range dates {
		merged = append(merged, Bar{Key: d, Count: totals[d]})
	}
	return merged
}

func getUsersPerDayCall(config *platformclientv2.Configuration, chunk *[]platformclientv2.Useraggregatequerypredicate, startDate, endDate, timeZone string) (message string, uesrsChart []Bar, err error) {
	jobID, err := getUsersPerDayCreateJob(config, chunk, startDate, endDate, timeZone)
	if err != nil {
		return fmt.Sprintf("Error creating job: %v\n", err), nil, err
	}

	jobStatus, err := getUsersPerDayGetJobStatus(config, jobID)
	if err != nil {
		return fmt.Sprintf("Error getting job status: %v\n", err), nil, err
	}

	if jobStatus == "FULFILLED" {
		results, err := getUsersPerDayGetJobResults(config, jobID, "")
		if err != nil {
			return fmt.Sprintf("Error getting job results: %v\n", err), nil, err
		}
		usersChart, err := getUsersPerDayFormatResults(results, startDate, endDate, timeZone)
		if err != nil {
			return fmt.Sprintf("Error formatting results: %v\n", err), nil, err
		}
		return "Completed GetUsersPerDay", usersChart, nil
	} else {
		return fmt.Sprintf("Unexpected job status: %s\n", jobStatus), nil, fmt.Errorf("unexpected job status: %s", jobStatus)
	}
}

func getUsersPerDayCreateJob(config *platformclientv2.Configuration, userChunk *[]platformclientv2.Useraggregatequerypredicate, startDate, endDate, timeZone string) (string, error) {
	utcStartDate, err := convertLocalToUTCStr(startDate, timeZone)
	if err != nil {
		return "", fmt.Errorf("error converting start date to UTC: %v", err)
	}
	utcEndDate, err := convertLocalToUTCStr(endDate, timeZone)
	if err != nil {
		return "", fmt.Errorf("error converting end date to UTC: %v", err)
	}
	intervalStr := utcStartDate + "/" + utcEndDate
	apiInstance := platformclientv2.NewUsersApiWithConfig(config)
	data, _, err := apiInstance.PostAnalyticsUsersAggregatesJobs(platformclientv2.Userasyncaggregationquery{
		Interval:    &intervalStr,
		Granularity: new("P1D"),
		TimeZone:    new(timeZone),
		GroupBy:     &[]string{"userId"},
		Metrics:     &[]string{"tSystemPresence"},
		Filter: &platformclientv2.Useraggregatequeryfilter{
			VarType:    new("and"),
			Predicates: userChunk,
		},
	})
	if err != nil {
		return "", err
	} else {
		return *data.JobId, nil
	}
}

func getUsersPerDayGetJobStatus(config *platformclientv2.Configuration, jobID string) (string, error) {
	apiInstance := platformclientv2.NewUsersApiWithConfig(config)
	for {
		data, _, err := apiInstance.GetAnalyticsUsersAggregatesJob(jobID)
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

func getUsersPerDayGetJobResults(config *platformclientv2.Configuration, jobID, cursor string) ([]platformclientv2.Useraggregatedatacontainer, error) {
	apiInstance := platformclientv2.NewUsersApiWithConfig(config)
	var allResults []platformclientv2.Useraggregatedatacontainer
	for {
		data, _, err := apiInstance.GetAnalyticsUsersAggregatesJobResults(jobID, cursor)
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

func getUsersPerDayFormatResults(results []platformclientv2.Useraggregatedatacontainer, startDate, endDate, timeZone string) ([]Bar, error) {
	usersChart := []Bar{}
	layout := "2006-01-02T15:04:05"

	dateToIndex := make(map[string]int)
	idx := 0

	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, err
	}

	start, err := time.ParseInLocation(layout, startDate, loc)
	if err != nil {
		return nil, err
	}

	end, err := time.ParseInLocation(layout, endDate, loc)
	if err != nil {
		return nil, err
	}

	// build dates in arrays
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		dateKey := day.Format("2006-01-02")
		dateToIndex[dateKey] = idx

		usersChart = append(usersChart, Bar{Key: day.Format("2006-01-02"), Count: 0})
		idx++
	}

	for _, result := range results {

		for _, interval := range *result.Data {
			rawDate := strings.Split(*interval.Interval, "/")[0]
			t, _ := time.Parse(time.RFC3339, rawDate)
			dayKey := t.In(loc).Format("2006-01-02")

			if i, ok := dateToIndex[dayKey]; ok {
				for _, metric := range *interval.Metrics {
					if metric.Metric != nil && *metric.Metric == "tSystemPresence" {
						if metric.Qualifier != nil && *metric.Qualifier != "OFFLINE" {
							usersChart[i].Count++
							break
						}
					}
				}
			}
		}
	}

	return usersChart, nil
}

func getAllUsers(config *platformclientv2.Configuration, pageNumber int) ([]string, error) {
	var allUsers []string
	pageSize := 200
	apiInstance := platformclientv2.NewUsersApiWithConfig(config)

	for {
		data, _, err := apiInstance.GetUsers(pageSize, pageNumber, nil, nil, "", nil, "", nil, "active")
		if err != nil {
			return nil, err
		}
		if data.Entities == nil || len(*data.Entities) == 0 {
			break
		}
		for _, user := range *data.Entities {
			if user.Id != nil {
				allUsers = append(allUsers, *user.Id)
			}
		}
		if len(*data.Entities) < pageSize {
			break
		}
		pageNumber++
	}
	return allUsers, nil
}
