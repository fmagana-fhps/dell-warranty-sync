package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fmagana-fhps/incidentiq-api-go"
	"github.com/fmagana-fhps/incidentiq-api-go/models"
	"github.com/joho/godotenv"
)

var client *iiq.Client
var logger log.Logger
var debug bool

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile("C:\\Sched_Tasks\\dell_warranty.log", 
					os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	logger = *log.New(file, "[Dell_Warranty] ", log.Ldate|log.Ltime|log.Lshortfile)
	flag.BoolVar(&debug, "debug", false, "Enable Debug Logging")
	if debug { logger.Println("DEBUG Initialization Complete") }
}

func NewRequest(method, site, params, payload string) (*http.Request, error) {
	url := "https://" + site + params
	if debug { logger.Printf("DEBUG %s", url) }
	req, err := http.NewRequest(method, url, strings.NewReader(payload))
	if debug { logger.Printf("DEBUG %+v %+v", req, err) }

	if err != nil {
		return nil, err
	}

	return req, nil
}

func requestToResponse[T any](req *http.Request, model *T) T {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Fatalln(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Fatalln(err)
	}

	if err = json.Unmarshal(body, &model); err != nil {
		logger.Fatalln(err)
	}

	return *model
}

func run() {
	flag.Parse()

	updates := make([]models.Asset, 0)
	client, _ = iiq.NewClient(&iiq.Options{
		Domain: os.Getenv("DOMAIN"),
		SiteId: os.Getenv("SITEID"),
		Token:  os.Getenv("TOKEN"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	})
	iiqAssets := getDellDevices(client)

	dell := &Dell{
		Site:  "apigtwb2c.us.dell.com/auth/oauth/v2/token",
	}

	dell.getAccessToken(os.Getenv("DELL_CLIENT_ID"), os.Getenv("DELL_CLIENT_SECRET"))

	if debug { logger.Println("DEBUG Get warranty info in batches of 100") }
	for begin, size := 0, 100; begin < len(iiqAssets); begin += size {
		end := begin + size
		if end > len(iiqAssets) {
			end = len(iiqAssets)
		}

		batch := iiqAssets[begin:end]
		serials := make([]string, 0, size)

		for k := range batch {
			device := batch[k]
			if device.WarrantyExpirationDate == "" &&
				device.SerialNumber != "" && len(device.SerialNumber) < 8 {
				serials = append(serials, batch[k].SerialNumber)
			}
		}

		tags := strings.Join(serials, ",")

		if strings.Contains(tags, ",,") {
			if debug { logger.Printf("DEBUG %s\n", tags) }
			panic(errors.New("the provided string " + tags + " is invalid"))
		} else if tags == "" {
			continue
		}

		site := "apigtwb2c.us.dell.com/PROD/sbil/eapi/v5/asset-entitlements"
		req, err := NewRequest("GET", site, "?servicetags="+tags, "")
		if err != nil {
			logger.Fatalln("error getting entitlements; " + err.Error())
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", dell.Token))

		dellAssets := requestToResponse(req, &DellAssets{})

		updates = append(updates, addExpiration(dellAssets, batch)...)
	}
	if debug { logger.Printf("DEBUG %d devices to update in iiQ\n", len(updates)) }

	updateAssets(updates)
}

func main() {
	run()
}
