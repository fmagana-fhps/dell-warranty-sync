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

	"common/models"
	"common/requests"

	"github.com/joho/godotenv"
)

func loadEnvVariables() map[string]string {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}
	return map[string]string{
		"DELL_CLIENT_ID":     os.Getenv("DELL_CLIENT_ID"),
		"DELL_CLIENT_SECRET": os.Getenv("DELL_CLIENT_SECRET"),
		"DELL_ACCESS":        os.Getenv("DELL_ACCESS"),
		"DELL_EXPIRES":       os.Getenv("DELL_EXPIRES"),
	}
}

func requestToResponse[T any](req *http.Request, model *T) T {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if err = json.Unmarshal(body, &model); err != nil {
		log.Fatalln(err)
	}

	return *model
}

func run() {
	updates := make([]models.Asset, 0)
	env := loadEnvVariables()
	iiqAssets := getDellDevices()

	dell := Dell{
		Site:  "apigtwb2c.us.dell.com/auth/oauth/v2/token",
		Token: env["DELL_ACCESS"],
	}

	if dell.Token == "" {
		dell.getAccessToken(env["DELL_CLIENT_ID"], env["DELL_CLIENT_SECRET"])

		env["DELL_ACCESS"] = dell.Token
		godotenv.Write(env, ".env")
	}

	fmt.Println("get dell warranties, by batches of 100")
	for begin, size := 0, 100; begin < len(iiqAssets.Assets); begin += size {
		end := begin + size
		if end > len(iiqAssets.Assets) {
			end = len(iiqAssets.Assets)
		}

		batch := iiqAssets.Assets[begin:end]
		serials := make([]string, 0, size)

		for k := range batch {
			if batch[k].WarrantyExpirationDate == "" &&
				batch[k].SerialNumber != "" {
				serials = append(serials, batch[k].SerialNumber)
			}
		}

		tags := strings.Join(serials, ",")

		if strings.Contains(tags, ",,") {
			fmt.Println(tags)
			panic(errors.New("the provided string " + tags + " is invalid"))
		}

		site := "apigtwb2c.us.dell.com/PROD/sbil/eapi/v5/asset-entitlements"
		req, err := requests.NewRequest("GET", site, "?servicetags="+tags, "")
		if err != nil {
			log.Fatalln(err)
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", dell.Token))

		dellAssets := requestToResponse(req, &DellAsset{})

		updates = append(updates, addExpiration(dellAssets, batch)...)
	}

	fmt.Println(len(updates), "devices to update in IIQ")

	updateAssets(updates)
}

func main() {
	run()
}
