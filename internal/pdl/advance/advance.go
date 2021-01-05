package advance

import (
	"bytes"
	"encoding/json"
	"github.com/onlythinking/pug-go/internal/config"
	"github.com/onlythinking/pug-go/pkg/help"

	log "github.com/onlythinking/pug-go/pkg/logging"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	// SUCCESS

	SUCCESS = "SUCCESS"

	// free Server error
	ERROR = "ERROR"

	//free Parameter should not be empty
	EMPTY_PARAMETER_ERROR = "EMPTY_PARAMETER_ERROR"

	//free Insufficient balance in your account,please recharge your account as soon as possible
	INSUFFICIENT_BALANCE = "INSUFFICIENT_BALANCE"

	//free The service is busy, please query later
	SERVICE_BUSY = "SERVICE_BUSY"

	//free Identity or access management certification failed
	IAM_FAILED = "IAM_FAILED"

	//free Free query limit is reached, please query later
	OVER_QUERY_LIMIT = "OVER_QUERY_LIMIT"

	// free parameter error, details refer to the samples on the right
	PARAMETER_ERROR = "PARAMETER_ERROR"

	// pay The card type is not match
	CARD_TYPE_NOT_MATCH = "CARD_TYPE_NOT_MATCH"

	//	pay No supported card detected from the image
	NO_SUPPORTED_CARD = "NO_SUPPORTED_CARD"

	// pay Too many cards detected from the image
	TOO_MANY_CARDS = "TOO_MANY_CARDS"

	// pay OCR check failed, unable to find any available fields on the uploaded picture
	OCR_NO_RESULT = "OCR_NO_RESULT"
)

type AdvResp struct {
	Code            string      `json:"code"`
	Message         string      `json:"message"`
	Data            AdvRespData `json:"data"`
	Extra           string      `json:"extra"`
	TransactionId   string      `json:"transactionId"`
	PricingStrategy string      `json:"pricingStrategy"`
}

type AdvRespData struct {
	CardType string           `json:"cardType"`
	Values   AdvRespDataValue `json:"values"`
}

type AdvRespDataValue struct {
	IdNumber   string `json:"idNumber"`
	Name       string `json:"name"`
	Birthday   string `json:"birthday"`
	FatherName string `json:"fatherName"`
}

type AdvClient struct {
	advUrl  string
	headers map[string]string
	params  map[string]string
}

// PAN_FRONT
func NewAdvOcrClient(cardType string) *AdvClient {
	cfg := config.App()
	advClient := AdvClient{}
	advClient.advUrl = cfg.Pdl.AdvanceAI.IdCardOcrUrl
	advClient.headers = map[string]string{"X-ADVAI-KEY": cfg.Pdl.AdvanceAI.AdvanceAiKey}
	advClient.params = map[string]string{"cardType": cardType}
	return &advClient
}

func (ths *AdvClient) ReqIdCardOcr(filename string) ([]byte, error) {
	count := 0

	exist, err := help.PathExists(filename)
	if !exist || err != nil {
		return nil, err
	}

	data, err := ths.DoReqIdCardOcr(filename, count)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (ths *AdvClient) DoReqIdCardOcr(filename string, reqCount int) ([]byte, error) {
	reqCount++
	request, err := newFileUploadRequest(ths.advUrl, ths.headers, ths.params, "image", filename)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		log.Errorf("ADV_ HTTP status：%s %s", resp.Status, string(respBody))
		return nil, err
	}

	adResp := AdvResp{}
	err = json.Unmarshal(respBody, &adResp)

	if err != nil {
		return nil, err
	}

	if SUCCESS != adResp.Code {
		if SERVICE_BUSY == adResp.Code && reqCount < 3 {
			return ths.DoReqIdCardOcr(filename, reqCount)
		}
	}

	return respBody, nil
}

func newFileUploadRequest(uri string, headers map[string]string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, fi.Name())
	if err != nil {
		return nil, err
	}
	part.Write(fileContents)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, _ := http.NewRequest("POST", uri, body)

	for key, val := range headers {
		request.Header.Add(key, val)
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	return request, err
}
