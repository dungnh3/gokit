package main

import (
	"encoding/json"
	"fmt"
	"github.com/dungnh3/gokit/library/nap"
)

type Resp struct {
	Data []struct {
		Ref  string `json:"ref"`
		Ts   uint64 `json:"ts"`
		Data struct {
			StoryId  string `json:"storyId"`
			MemberId string `json:"memberId"`
			Step     string `json:"step"`
		}
	}
}

func main() {
	res, err := doAction()
	if err != nil {
		fmt.Println("error =>> ")
		fmt.Println(err.Error())
		return
	}
	marshal,_ := json.Marshal(*res)
	fmt.Println(string(marshal))
}

func doAction() (*Resp, error) {
	var (
		response Resp
		errorRaw nap.Raw
	)
	client := nap.New().Base( /* host */ "http://localhost:3000").
		Get("/api/v1/tracking?storyId=X87aZRAAACEAPOFC&level=beginner&memberId=dung")

	_, err := client.Receive(&response, &errorRaw)
	if err != nil {
		return nil, err
	}
	if errorRaw != nil {
		return nil, fmt.Errorf("error call sellable_products Response %v", string(errorRaw))
	}
	return &response, nil
}
