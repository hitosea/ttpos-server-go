package google

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"ttpos-server-go/config"
)

func PlacesSearchText(text string, latitude float64, longitude float64) (string, error) {
	if config.Google.PlacesApiKey == "" {
		return "", errors.New("谷歌地图搜索配置为空")
	}
	//
	payload := strings.NewReader(`{
		"textQuery" : "` + text + `",
		"locationBias": {
			"circle": {
				"center": {
					"latitude": ` + strconv.FormatFloat(latitude, 'f', -1, 64) + `,
					"longitude": ` + strconv.FormatFloat(longitude, 'f', -1, 64) + `
				},
				"radius": 500.0
			}
		}
	}`)
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://places.googleapis.com/v1/places:searchText", payload)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("X-Goog-Api-Key", config.Google.PlacesApiKey)
	req.Header.Add("X-Goog-FieldMask", "places.displayName,places.formattedAddress,places.location,places.id,places.name,places.plusCode,places.shortFormattedAddress,places.addressComponents")
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//
	result := map[string]interface{}{}
	json.Unmarshal(body, &result)
	//
	resultJson, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//
	return string(resultJson), nil
}
