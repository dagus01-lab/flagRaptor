package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type SubmitterFormat struct {
	SUB_ACCEPTED, SUB_INVALID, SUB_OLD, SUB_YOUR_OWN, SUB_STOLEN, SUB_NOP, SUB_NOT_AVAILABLE string
}

type ResponseItem struct {
	Flag    string
	Message string
}

var (
	DummySubmitterFormat = SubmitterFormat{
		SUB_ACCEPTED:      "accepted",
		SUB_INVALID:       "invalid",
		SUB_OLD:           "too old",
		SUB_YOUR_OWN:      "your own",
		SUB_STOLEN:        "already stolen",
		SUB_NOP:           "from NOP team",
		SUB_NOT_AVAILABLE: "is not available",
	}
	CCITSubmitterFormat = SubmitterFormat{
		SUB_ACCEPTED:      "accepted",
		SUB_INVALID:       "invalid",
		SUB_OLD:           "too old",
		SUB_YOUR_OWN:      "your own",
		SUB_STOLEN:        "already stolen",
		SUB_NOP:           "from NOP team",
		SUB_NOT_AVAILABLE: "is not available",
	}
)

func GetAppSubmitter() (SubmitterFormat, func(flags []string) []ResponseItem) {
	switch cfg.SubmissionConf.SubProtocol {
	case "ccit":
		return CCITSubmitterFormat, CCITSubmitter
	case "dummy":
		return DummySubmitterFormat, DummySubmitter
	default:
		return DummySubmitterFormat, nil
	}
}

func DummySubmitter(flags []string) []ResponseItem {

	var res []ResponseItem

	for _, flag := range flags {
		res = append(res, ResponseItem{Flag: flag, Message: DummySubmitterFormat.SUB_ACCEPTED})
	}
	return res
}

func CCITSubmitter(flags []string) []ResponseItem {

	var res []ResponseItem

	requestBody, err := json.Marshal(flags)
	if err != nil {
		fmt.Println("Error marshaling request payload: ", err)
		return nil
	}
	req, err := http.NewRequest("PUT", cfg.SubmissionConf.SubUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error creating PUT request: ", err)
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Team-Token", cfg.GameConf.TeamToken)
	//perform the PUT request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending PUT request: ", err)
		return nil
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error unmarshaling response payload: ", err)
		return nil
	}

	err = json.Unmarshal(responseBody, &res)
	if err != nil {
		fmt.Println("Error unmarshaling response payload: ", err)
		return nil
	}

	return res
}
