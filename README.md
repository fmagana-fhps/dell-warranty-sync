# dell-warranty-sync

Using Dell's provided API, we will take the warranty information and insert it into IncidentIQ

Using the current go.mod will not work without also having [iiq-request-scripts](https://github.com/fmagana-fhps/iiq-request-scripts) within the same parent directory.. for now

---

## TODO

- parse the date and time information from Dell to be use to return to IncidentIQ
- split the main.go file into multiple files for better readability
- figure out a sytem to save and load from .env
- save and use expiration time from Dell
- do a check on whether the asset already has warranty information in IncidentIQ