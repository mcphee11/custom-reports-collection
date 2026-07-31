package genesys

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
)

func getAllQueues(config *platformclientv2.Configuration, pageNumber int) ([]platformclientv2.Queue, error) {
	var allQueues []platformclientv2.Queue
	pageSize := 200
	apiInstance := platformclientv2.NewRoutingApiWithConfig(config)

	for {
		data, _, err := apiInstance.GetRoutingQueues(pageNumber, pageSize, "", "", []string{}, []string{}, []string{}, "", false, []string{})
		if err != nil {
			return nil, err
		}
		if data.Entities == nil || len(*data.Entities) == 0 {
			break
		}
		for _, flow := range *data.Entities {
			if flow.Id != nil {
				allQueues = append(allQueues, flow)
			}
		}
		if len(*data.Entities) < pageSize {
			break
		}
		pageNumber++
	}
	return allQueues, nil
}

func GetObjectQueues(config *platformclientv2.Configuration, startDate, endDate, timeZone string) (message string, allQueue []platformclientv2.Queue, flowsTable []string, err error) {

	allQueues, err := getAllQueues(config, 1)
	if err != nil {
		return fmt.Sprintf("Error getting queues: %v\n", err), nil, nil, err
	}

	jobID, err := getObjectQueuesCreateJob(config, startDate, endDate, timeZone)
	if err != nil {
		return fmt.Sprintf("Error creating job: %v\n", err), nil, nil, err
	}

	jobStatus, err := getObjectQueuesGetJobStatus(config, jobID)
	if err != nil {
		return fmt.Sprintf("Error getting job status: %v\n", err), nil, nil, err
	}

	if jobStatus == "FULFILLED" {
		usedFlowsResults, err := getObjectQueuesGetJobResults(config, jobID, "")
		if err != nil {
			return fmt.Sprintf("Error getting job results: %v\n", err), nil, nil, err
		}
		flowsTable, err := formatObjectQueuesResults(usedFlowsResults, allQueues)
		if err != nil {
			return fmt.Sprintf("Error formatting results: %v\n", err), nil, nil, err
		}
		return "Completed GetObjectQueues", allQueues, flowsTable, nil
	} else {
		return fmt.Sprintf("Unexpected job status: %s\n", jobStatus), nil, nil, fmt.Errorf("unexpected job status: %s", jobStatus)
	}
}

func getObjectQueuesCreateJob(config *platformclientv2.Configuration, startDate, endDate, timeZone string) (string, error) {
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
		Interval: &intervalStr,
		TimeZone: new(timeZone),
		GroupBy:  &[]string{"queueId"},
		Metrics:  &[]string{"nConnected"},
	})
	if err != nil {
		return "", err
	} else {
		return *data.JobId, nil
	}
}

func getObjectQueuesGetJobStatus(config *platformclientv2.Configuration, jobID string) (string, error) {
	apiInstance := platformclientv2.NewAnalyticsApiWithConfig(config)
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

func getObjectQueuesGetJobResults(config *platformclientv2.Configuration, jobID, cursor string) ([]platformclientv2.Conversationaggregatedatacontainer, error) {
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

func formatObjectQueuesResults(results []platformclientv2.Conversationaggregatedatacontainer, allQueues []platformclientv2.Queue) ([]string, error) {
	inUse := "⛔"
	if len(allQueues) > 0 {
		inUse = "✅"
	}

	tableData := []string{"Queues", strconv.Itoa(len(allQueues)), strconv.Itoa(len(results)), inUse}

	return tableData, nil
}
