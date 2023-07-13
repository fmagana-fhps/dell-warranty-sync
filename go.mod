module github.com/fmagana-fhps/dell-warranty-sync

go 1.20

require (
	common v0.0.0-00010101000000-000000000000
	github.com/joho/godotenv v1.5.1
)

replace common => ../iiq-request-scripts/common
