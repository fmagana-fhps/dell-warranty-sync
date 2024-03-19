package main

import (
	"github.com/fmagana-fhps/incidentiq-api-go/models"
	"fmt"
	"time"
)

type DellAssets []struct {
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

type Dell struct {
	Site  string
	Token string `json:"access_token"`
}

func (d *Dell) getAccessToken(id, secret string) {
	payload := "client_id=" + id +
		"&client_secret=" + secret +
		"&grant_type=client_credentials"

	request, err := NewRequest("POST", d.Site, "", payload)
	if err != nil {
		panic(err)
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	requestToResponse(request, d)
}

func addExpiration(dell DellAssets, batch []models.Asset) []models.Asset {
	var list []models.Asset

	for i := range dell {
		if dell[i].Invalid {
			continue
		}

		entitle := len(dell[i].Entitlements)
		if entitle > 0 {
			warranty := dell[i].Entitlements[entitle-1]
			year, month, day := warranty.EndDate.Date()
			date := fmt.Sprintf("%v-%v-%v", year, int(month), day)

			for j := range batch {
				if dell[i].ServiceTag == batch[j].SerialNumber {
					batch[j].WarrantyExpirationDate = date

					list = append(list, batch[j])
					continue
				}
			}
		}
	}
	return list
}
