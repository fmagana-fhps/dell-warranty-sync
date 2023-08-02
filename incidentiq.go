package main

import (
	"common/models"
	"common/requests"
	"encoding/json"
	"log"
)

type IncidentIQ struct {
}

func getDellDevices() models.MultipleAssets {
	response, err := requests.Post("assets/?$s=17000&$o=AssetTag&$d=Ascending&$filter=(ManufacturerId%20eq%20%27%5B518000c0-4dff-e511-a789-005056bb000e%5D%27)", "")
	if err != nil {
		log.Fatalln(err)
	}

	dellAssets := models.MultipleAssets{}
	err = json.Unmarshal(response, &dellAssets)
	if err != nil {
		log.Fatalln(err)
	}

	return dellAssets
}

func updateAssets(editedAssets []models.Asset) {
	for _, asset := range editedAssets {

		updated, err := json.Marshal(asset)
		if err != nil {
			log.Fatalln(err)
		}
		payload := string(updated)

		_, err = requests.Post("assets/"+asset.AssetId, payload)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
