package genesys

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
)

func GetRouting(config *platformclientv2.Configuration, allQueues []platformclientv2.Queue, startDate, endDate, timeZone string) (message string, routingTable [][]string, err error) {
	jobID, err := getRoutingCreateJob(config, startDate, endDate, timeZone)
	if err != nil {
		return fmt.Sprintf("Error creating job: %v\n", err), nil, err
	}

	jobStatus, err := getRoutingGetJobStatus(config, jobID)
	if err != nil {
		return fmt.Sprintf("Error getting job status: %v\n", err), nil, err
	}

	if jobStatus == "FULFILLED" {
		results, err := getRoutingGetJobResults(config, jobID, "")
		if err != nil {
			return fmt.Sprintf("Error getting job results: %v\n", err), nil, err
		}
		formattedResults, err := getRoutingFormatResults(results, allQueues)
		if err != nil {
			return fmt.Sprintf("Error formatting results: %v\n", err), nil, err
		}
		return "Completed GetRouting", formattedResults, nil
	} else {
		return fmt.Sprintf("Unexpected job status: %s\n", jobStatus), nil, fmt.Errorf("unexpected job status: %s", jobStatus)
	}
}

func getRoutingCreateJob(config *platformclientv2.Configuration, startDate, endDate, timeZone string) (string, error) {
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
		SegmentFilters: &[]platformclientv2.Segmentdetailqueryfilter{{
			VarType: new("and"),
			Predicates: &[]platformclientv2.Segmentdetailquerypredicate{{
				Dimension: new("requestedRouting"),
				Operator:  new("exists"),
			}},
		}},
		ConversationFilters: &[]platformclientv2.Conversationdetailqueryfilter{{
			VarType: new("and"),
			Predicates: &[]platformclientv2.Conversationdetailquerypredicate{{
				Metric:   new("nConnected"),
				Operator: new("exists"),
			}},
		}},
	})
	if err != nil {
		return "", err
	} else {
		return *data.JobId, nil
	}
}

func getRoutingGetJobStatus(config *platformclientv2.Configuration, jobID string) (string, error) {
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

func getRoutingGetJobResults(config *platformclientv2.Configuration, jobID, cursor string) ([]platformclientv2.Analyticsconversation, error) {
	apiInstance := platformclientv2.NewConversationsApiWithConfig(config)
	var allResults []platformclientv2.Analyticsconversation
	for {
		data, _, err := apiInstance.GetAnalyticsConversationsDetailsJobResults(jobID, cursor, 1000)
		if err != nil {
			return nil, err
		}
		if data.Conversations != nil {
			allResults = append(allResults, *data.Conversations...)
		}
		if data.Cursor == nil || *data.Cursor == "" {
			break
		}
		cursor = *data.Cursor
	}
	return allResults, nil
}

func getRoutingFormatResults(results []platformclientv2.Analyticsconversation, allQueues []platformclientv2.Queue) ([][]string, error) {
	tableData := [][]string{{"Object", "Total Built", "Total Used", "In Use"}}
	bullseye, predective, conditional, preferred, standard, direct, last, manual, vip := 0, 0, 0, 0, 0, 0, 0, 0, 0
	bullseyeBuilt, predectiveBuilt, conditionalBuilt, preferredBuilt, standardBuilt, directBuilt, lastBuilt, manualBuilt, vipBuilt := 0, 0, 0, 0, 0, 0, 0, 0, 0

	// Get used data
	for _, conversation := range results {
		if conversation.Participants == nil {
			continue
		}
		for _, participant := range *conversation.Participants {
			if participant.Purpose != nil && *participant.Purpose == "customer" {
				if participant.Sessions != nil && len(*participant.Sessions) > 0 {
					// get for each session in case routing changes during the conversation and count each routing used
					for _, session := range *participant.Sessions {
						if session.UsedRouting != nil {
							switch *session.UsedRouting {
							case "Bullseye":
								bullseye++
							case "Predictive":
								predective++
							case "Conditional":
								conditional++
							case "Preferred":
								preferred++
							case "Standard":
								standard++
							case "Direct":
								direct++
							case "Last":
								last++
							case "Manual":
								manual++
							case "Vip":
								vip++
							default:
								// Do nothing
							}
						}
					}
				}
			}
		}
	}

	// Get built data
	for _, queue := range allQueues {
		if queue.Bullseye != nil {
			bullseyeBuilt++
			continue
		}
		// This is not in the GO SDK this is on purpose
		//if queue.PredictiveRouting != nil {
		//	predectiveBuilt++
		//}
		if queue.ConditionalGroupRouting != nil || queue.ConditionalGroupActivation != nil {
			conditionalBuilt++
			continue
		}
		// This is due to a UI issue where the array was never set
		if queue.RoutingRules != nil {
			if len(*queue.RoutingRules) > 0 {
				preferredBuilt++
				continue
			}
		} else {
			standardBuilt++
		}
	}

	tableData = append(tableData, []string{"Bullseye", strconv.Itoa(bullseyeBuilt), strconv.Itoa(bullseye), "yes"})
	tableData = append(tableData, []string{"Predictive", strconv.Itoa(predectiveBuilt), strconv.Itoa(predective), "yes"})
	tableData = append(tableData, []string{"Conditional", strconv.Itoa(conditionalBuilt), strconv.Itoa(conditional), "yes"})
	tableData = append(tableData, []string{"Preferred", strconv.Itoa(preferredBuilt), strconv.Itoa(preferred), "yes"})
	tableData = append(tableData, []string{"Standard", strconv.Itoa(standardBuilt), strconv.Itoa(standard), "yes"})
	tableData = append(tableData, []string{"Direct", strconv.Itoa(directBuilt), strconv.Itoa(direct), "yes"})
	tableData = append(tableData, []string{"Last", strconv.Itoa(lastBuilt), strconv.Itoa(last), "yes"})
	tableData = append(tableData, []string{"Manual", strconv.Itoa(manualBuilt), strconv.Itoa(manual), "yes"})
	tableData = append(tableData, []string{"Vip", strconv.Itoa(vipBuilt), strconv.Itoa(vip), "yes"})

	return tableData, nil
}
