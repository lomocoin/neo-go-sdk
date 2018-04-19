package neo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lomocoin/neo-go-sdk/neo/models/request"
	resp "github.com/lomocoin/neo-go-sdk/neo/models/response"
	"github.com/pkg/errors"
)

func executeRequest(method string, bodyParameters []interface{}, nodeURI string, model interface{}) error {
	var body []byte
	var err error

	if bodyParameters == nil {
		body, err = request.NewBody(method)
		if err != nil {
			return err
		}
	} else {
		body, err = request.NewBodyWithParameters(method, bodyParameters)
		if err != nil {
			return err
		}
	}

	ioBody := bytes.NewReader(body)

	request, err := http.NewRequest("POST", nodeURI, ioBody)
	if err != nil {
		return err
	}

	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf(
			"non-200 status code returned from NEO node, got: '%d'",
			response.StatusCode,
		)
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &model)
	if err != nil {
		return err
	}

	// handle error response info
	var errorResp resp.Error
	err = json.Unmarshal(bytes, &errorResp)
	if err != nil {
		return err
	} else if errorResp.Error.Message != "" {
		return errors.Errorf("error code: %v, error message: %v", errorResp.Error.Code, errorResp.Error.Message)
	}

	return nil
}
