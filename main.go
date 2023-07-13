package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"common/models"
	"common/requests"

	"github.com/joho/godotenv"
)

const ()

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type DellResponse struct {
	AccessToken string
}

type DellAsset []struct {
	ID                     int       `json:"id"`
	ServiceTag             string    `json:"serviceTag"`
	OrderBuid              int       `json:"orderBuid"`
	ShipDate               time.Time `json:"shipDate"`
	ProductCode            string    `json:"productCode"`
	LocalChannel           string    `json:"localChannel"`
	ProductID              any       `json:"productId"`
	ProductLineDescription string    `json:"productLineDescription"`
	ProductFamily          any       `json:"productFamily"`
	SystemDescription      any       `json:"systemDescription"`
	ProductLobDescription  string    `json:"productLobDescription"`
	CountryCode            string    `json:"countryCode"`
	Duplicated             bool      `json:"duplicated"`
	Invalid                bool      `json:"invalid"`
	Entitlements           []struct {
		ItemNumber              string    `json:"itemNumber"`
		StartDate               time.Time `json:"startDate"`
		EndDate                 time.Time `json:"endDate"`
		EntitlementType         string    `json:"entitlementType"`
		ServiceLevelCode        string    `json:"serviceLevelCode"`
		ServiceLevelDescription string    `json:"serviceLevelDescription"`
		ServiceLevelGroup       int       `json:"serviceLevelGroup"`
	} `json:"entitlements"`
}

//	I HAVE ALREADY RAN THIS!! CHECK INFO.TXT FOR TIME AND DATE INFORMATION TO PARSE

func main() {

	err := godotenv.Load()
	check(err)
	dellClientId := os.Getenv("DELL_CLIENT_ID")
	dellClientSecret := os.Getenv("DELL_CLIENT_SECRET")

	response, err := requests.Post("assets/?$s=16000&$o=AssetTag&$d=Ascending&$filter=(ManufacturerId%20eq%20'%5B518000c0-4dff-e511-a789-005056bb000e%5D')", strings.NewReader(``))
	check(err)

	schema := models.MultipleAssets{}
	err = json.Unmarshal(response, &schema)
	check(err)

	access := DellResponse{}
	if os.Getenv("DELL_ACCESS") == "" {
		payload := "client_id=" + dellClientId + "&client_secret=" + dellClientSecret + "&grant_type=client_credentials"

		request, err := requests.NewRequest("POST", "apigtwb2c.us.dell.com/auth/oauth/v2/token", "", strings.NewReader(payload))
		check(err)

		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		res, err := http.DefaultClient.Do(request)
		check(err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		check(err)

		fmt.Println(string(body))

		err = json.Unmarshal(body, &access)
		check(err)
	} else {
		access.AccessToken = os.Getenv("DELL_ACCESS")
	}

	file, err := os.Create("info.txt")
	check(err)

	defer file.Close()

	batch := 100

	for i := 0; i < len(schema.Assets); i += batch {
		j := i + batch
		if j > len(schema.Assets) {
			j = len(schema.Assets)
		}

		batchAssets := schema.Assets[i:j]
		assets := make([]string, 1, 100)

		for idx, asset := range batchAssets {
			if idx == 0 {
				assets[idx] = asset.SerialNumber
			}

			if asset.SerialNumber != "" {
				assets = append(assets, asset.SerialNumber)
			}
		}

		serials := strings.Join(assets, ",")

		if strings.Contains(serials, ",,") {
			fmt.Println(serials)
			check(errors.New("the provided string " + serials + " is invalid"))
		}

		req, err := requests.NewRequest("GET", "apigtwb2c.us.dell.com/PROD/sbil/eapi/v5/asset-entitlements", "?servicetags="+serials, strings.NewReader(``))
		check(err)

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Authorization", "Bearer "+access.AccessToken)

		res, err := http.DefaultClient.Do(req)
		check(err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		check(err)

		dell := DellAsset{}
		err = json.Unmarshal(body, &dell)
		check(err)

		for _, device := range dell {
			if device.Invalid {
				continue
			}

			entitle := len(device.Entitlements)
			if entitle > 0 {
				file.WriteString(device.ServiceTag + ": " + device.Entitlements[entitle-1].EndDate.String() + "\n")
			}
		}
	}
}
