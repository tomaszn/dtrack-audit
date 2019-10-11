package dtrack

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	BOM_UPLOAD_URL       = "/api/v1/bom"
	PROJECT_FINDINGS_URL = "/api/v1/finding/project"
	BOM_TOKEN_URL        = "/api/v1/bom/token"
	API_POLLING_STEP     = 5 * time.Second
)

func checkError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

type Payload struct {
	Project string `json:"project"`
	Bom     string `json:"bom"`
}

type UploadResult struct {
	Token string `json:"token"`
}

type ProcessState struct {
	Processing bool `json:"processing"`
}

type ApiClient struct {
	ApiKey string
	ApiUrl string
}

func (apiClient ApiClient) Upload(inputFileName, projectId string) (uploadResult UploadResult, err error) {
	bomDataXml, err := ioutil.ReadFile(inputFileName)

	if err != nil {
		return
	}

	bomDataB64 := b64.StdEncoding.EncodeToString(bomDataXml)
	payload := Payload{Project: projectId, Bom: bomDataB64}
	payloadJson, err := json.Marshal(payload)

	if err != nil {
		return
	}

	client := apiClient.getHttpClient()
	req, err := http.NewRequest(http.MethodPut, apiClient.ApiUrl+BOM_UPLOAD_URL, bytes.NewBuffer(payloadJson))
	req.Header.Add("X-API-Key", apiClient.ApiKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()
	checkError(apiClient.checkRespStatusCode(resp.StatusCode))

	err = json.NewDecoder(resp.Body).Decode(&uploadResult)

	if err != nil {
		return
	}

	return uploadResult, nil
}

func (apiClient ApiClient) checkRespStatusCode(statusCode int) error {
	errorStatusCodes := map[int]string{
		404: "The project could not be found or invalid API URL",
		401: "Authorization error",
	}

	if errorMsg, ok := errorStatusCodes[statusCode]; ok {
		return fmt.Errorf(errorMsg)
	}
	return nil
}

func (apiClient ApiClient) isTokenBeingProcessed(token string) (result bool, err error) {
	processState := ProcessState{}
	client := apiClient.getHttpClient()
	req, err := http.NewRequest(http.MethodGet, apiClient.ApiUrl+BOM_TOKEN_URL+"/"+token, nil)
	req.Header.Add("X-API-Key", apiClient.ApiKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()
	checkError(apiClient.checkRespStatusCode(resp.StatusCode))
	err = json.NewDecoder(resp.Body).Decode(&processState)

	if err != nil {
		return
	}

	if processState.Processing {
		return true, nil
	}

	return false, nil

}

func (apiClient ApiClient) getHttpClient() *http.Client {
	// Workaround for empty response body
	// See https://github.com/DependencyTrack/dependency-track/issues/474
	tr := &http.Transport{
		DisableCompression: true,
	}
	return &http.Client{Transport: tr}
}

type Component struct {
	Uuid    string `json:"uuid"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Vulnerability struct {
	Uuid           string `json:"uuid"`
	VulnId         string `json:"vulnId"`
	Source         string `json:"source"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Severity       string `json:"severity"`
	Recommendation string `json:"recommendation"`
}

type Analysis struct {
	AnalysisState string `json:"analysisState"`
}

type Finding struct {
	Comp   Component     `json:"component"`
	Vuln   Vulnerability `json:"vulnerability"`
	An     Analysis      `json:"analysis"`
	Matrix string        `json:"matrix"`
}

func (apiClient ApiClient) GetFindings(projectId string) (result []Finding, err error) {
	client := apiClient.getHttpClient()
	req, err := http.NewRequest(http.MethodGet, apiClient.ApiUrl+PROJECT_FINDINGS_URL+"/"+projectId, nil)
	req.Header.Add("X-API-Key", apiClient.ApiKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()
	checkError(apiClient.checkRespStatusCode(resp.StatusCode))
	err = json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		return
	}
	return
}

func (apiClient ApiClient) PollTokenBeingProcessed(token string, timeout <-chan time.Time) error {
	time.Sleep(API_POLLING_STEP)
	for {
		select {
		case <-timeout:
			return nil
		default:
			state, err := apiClient.isTokenBeingProcessed(token)
			if err != nil {
				return err
			}
			if state == false {
				return nil
			}
			time.Sleep(API_POLLING_STEP)
		}
	}
	return nil
}

func (c ApiClient) GetVulnViewUrl(v Vulnerability) string {
	// TODO use url builder
	return c.ApiUrl + "/vulnerability/?source=" + v.Source + "&vulnId=" + v.VulnId
}