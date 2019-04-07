package clarifai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"net/http"
)

var netClient = &http.Client{
	Timeout: time.Second * 2,
}

type app struct {
	apiKey   string
	modelURL string
}

// NewService - Creates a Clarifai service for using Clarifai APIs
func NewService(apiKey string, modelURL string) Service {
	return &app{apiKey, modelURL}
}

// ParseResponse parses response from clarifai predict
func ParseResponse(b64str string, pResp *PredictResponse) *FrameInfo {

	if (&PredictResponse{}) == pResp {
		fmt.Println("pResp is empty")
		return &FrameInfo{}
	}

	// TODO: revert to clarifai batching ?
	output := pResp.Outputs[0]

	frInfo := FrameInfo{}

	frInfo.InputID = output.Input.ID
	frInfo.URL = output.Input.Data.Image.URL
	frInfo.B64 = b64str

	objs := make([]object, 0)

	regions := output.Data.Regions
	for _, region := range regions {

		conceptName := region.RegionData.Concepts[0].Name
		conceptValue := region.RegionData.Concepts[0].Value
		bbox := region.RegionInfo.BoundingBox

		//TODO TBD use config file to filter? maybe filter later as we  may want to store all data?
		if conceptName == "person" && conceptValue > 0.90 {
			obj := object{bbox, conceptName, conceptValue}
			objs = append(objs, obj)
		}
	}
	frInfo.Objects = objs

	return &frInfo
}

func buildRequestByBytes(timestamp int64, bStr string) inputs {
	data := make([]dataByBytes, 0)

	d := dataByBytes{}
	d.ID = strconv.FormatInt(timestamp, 10)
	d.Data.Image.Base64 = bStr
	data = append(data, d)

	inputs := inputs{data}
	return inputs
}

func (a *app) PredictByBytes(timestamp int64, b64Str string) (*FrameInfo, error) {
	payload := buildRequestByBytes(timestamp, b64Str)

	reqStr, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}

	req, _ := http.NewRequest("POST", a.modelURL, bytes.NewBuffer(reqStr))
	req.Header.Set("Authorization", fmt.Sprintf("Key %s", a.apiKey))
	req.Header.Set("Content-Type", "application/json")

	client := netClient
	now := time.Now().UTC()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("request error:", err)
		return nil, err
	}
	fmt.Println("response elaspesd time: ", time.Since(now))
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}
	//fmt.Println(string(respBody))

	var predictResp PredictResponse
	json.Unmarshal(respBody, &predictResp)
	if err != nil {
		return nil, err
	}

	if predictResp.Status.Code != 10000 {
		fmt.Printf("%s\n", predictResp.Status.Description)
		return nil, errors.New("Clarifai status code something other than 10000")
	}

	/*
		//Debugging
		out, err := json.Marshal(predictResp)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(out))
	*/

	frameInfo := ParseResponse(b64Str, &predictResp)

	return frameInfo, nil
}
