package api

import (
	"encoding/json"
	"github.com/analogj/lodestone-processor/pkg/model"
	"io/ioutil"
	"net/http"
	"net/url"
)

func GetIncludeExcludeData(apiEndpoint *url.URL) (model.Filter, error) {

	//manipulate the path
	apiEndpoint.Path = "/api/v1/data/filetypes.json"

	resp, err := http.Get(apiEndpoint.String())
	if err != nil {
		return model.Filter{}, err
	}
	defer resp.Body.Close()

	bodyJson, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return model.Filter{}, err
	}

	var filter model.Filter
	err = json.Unmarshal(bodyJson, &filter)

	return filter, err
}
