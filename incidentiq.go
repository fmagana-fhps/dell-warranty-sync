package main

import (
	iiq "github.com/fmagana-fhps/incidentiq-api-go"
	"github.com/fmagana-fhps/incidentiq-api-go/models"
)

func getDellDevices(client *iiq.Client) []models.Asset {
	params := iiq.Parameters{
		PageSize: 20000,
		OrderBy:  "AssetTag DESC",
	}

	body := models.Search{
		OnlyShowDeleted: false,
		FilterByViewPermission: true,
		Filters: []models.Filter{
			{
				Facet:      "manufacturer",
				Name:       "Dell",
				ID:         "518000c0-4dff-e511-a789-005056bb000e",
				Value:      "",
				Negative:   false,
				SortOrder:  "",
				GroupIndex: 0,
			},
			{
				Facet:      "warrantyexpirationdate",
				Name:       "daterange:01/01/2000-12/31/2100",
				Value:      "daterange:01/01/2000-12/31/2100",
				Negative:   true,
				GroupIndex: 0,
			},
			{
				Facet: "AssetType",
				ID:    "2a1561e5-34ff-4fcf-87de-2a146f0e1c01",
			},
		},
	}

	if debug { logger.Printf("DEBUG %+v", params) }
	response, err := client.AllAssets(params, body)
	if debug { logger.Printf("DEBUG %d, %+v", response.StatusCode, err) }
	if err != nil { logger.Fatalln(err) }

	return response.Body.Items
}

func updateAssets(editedAssets []models.Asset) {
	for _, asset := range editedAssets {
		copied := asset	
		err := client.UpdateAsset("assets/"+copied.AssetID, copied)
		if err != nil { logger.Fatalln(err, copied, asset) }

		if debug { logger.Printf("DEBUG %s has been updated", copied.AssetTag) }
	}
}
