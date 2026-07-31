package genesys

import (
	"encoding/json"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
	"strconv"
	"time"
)

func GetConversationTotals(config *platformclientv2.Configuration, startDate, endDate, timeZone string) (message string, inboundChart, outboundChart, inboundTable, outboundTable []byte, err error) {
	jobID, err := getConversationTotalsCreateJob(config, startDate, endDate, timeZone)
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
		formattedResultsInboundChart, formattedResultsOutboundChart, formattedResultsInboundTable, formattedResultsOutboundTable, err := getConversationTotalsFormatResults(results)
		if err != nil {
			return fmt.Sprintf("Error formatting results: %v\n", err), nil, nil, nil, nil, err
		}
		return "Completed GetConversationTotals", formattedResultsInboundChart, formattedResultsOutboundChart, formattedResultsInboundTable, formattedResultsOutboundTable, nil
	} else {
		return fmt.Sprintf("Unexpected job status: %s\n", jobStatus), nil, nil, nil, nil, fmt.Errorf("unexpected job status: %s", jobStatus)
	}
}

func getConversationTotalsCreateJob(config *platformclientv2.Configuration, startDate, endDate, timeZone string) (string, error) {
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
		GroupBy:  &[]string{"direction", "mediaType"},
		Metrics:  &[]string{"nConnected"},
	})
	if err != nil {
		return "", err
	} else {
		return *data.JobId, nil
	}
}

func getConversationTotalsGetJobStatus(config *platformclientv2.Configuration, jobID string) (string, error) {
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

func getConversationTotalsGetJobResults(config *platformclientv2.Configuration, jobID, cursor string) ([]platformclientv2.Conversationaggregatedatacontainer, error) {
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

func getConversationTotalsFormatResults(results []platformclientv2.Conversationaggregatedatacontainer) ([]byte, []byte, []byte, []byte, error) {
	inboundChart := []Bar{{Key: "voice", Count: 0}, {Key: "message", Count: 0}, {Key: "email", Count: 0}, {Key: "callback", Count: 0}}
	outboundChart := []Bar{{Key: "voice", Count: 0}, {Key: "message", Count: 0}, {Key: "email", Count: 0}, {Key: "callback", Count: 0}}
	inboundTable := [][]string{{"MediaType", "Interactions"}, {"voice", "0"}, {"message", "0"}, {"email", "0"}, {"callback", "0"}}
	outboundTable := [][]string{{"MediaType", "Interactions"}, {"voice", "0"}, {"message", "0"}, {"email", "0"}, {"callback", "0"}}

	for _, result := range results {
		mediaType := "Unknown"
		direction := "Unknown"

		if result.Group != nil {
			if val, ok := (*result.Group)["mediaType"]; ok {
				mediaType = val
			}
			if val, ok := (*result.Group)["direction"]; ok {
				direction = val
			}
		}
		count := *(*(*result.Data)[0].Metrics)[0].Stats.Count

		if direction == "inbound" {
			for i, m := range inboundChart {
				if m.Key == mediaType {
					inboundChart[i].Count = count
					inboundTable[i+1][1] = strconv.Itoa(count)
				}
			}
		}
		if direction == "outbound" {
			for i, m := range outboundChart {
				if m.Key == mediaType {
					outboundChart[i].Count = count
					outboundTable[i+1][1] = strconv.Itoa(count)
				}
			}
		}
	}

	jsonInboundChart, err := json.Marshal(inboundChart)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	jsonInboundTable, err := json.Marshal(inboundTable)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	jsonOutboundChart, err := json.Marshal(outboundChart)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	jsonOutboundTable, err := json.Marshal(outboundTable)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return jsonInboundChart, jsonOutboundChart, jsonInboundTable, jsonOutboundTable, nil
}
