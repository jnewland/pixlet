package browser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	TidbytAPIPush = "https://api.tidbyt.com/v0/devices/%s/push"
)

type TidbytPushJSON struct {
	DeviceID       string `json:"deviceID"`
	Image          string `json:"image"`
	InstallationID string `json:"installationID"`
	Background     bool   `json:"background"`
}

func (b *Browser) pushHandler(w http.ResponseWriter, r *http.Request) {
	var (
		deviceID       string
		apiToken       string
		installationID string
		background     bool
	)

	// Parse the request form so we can use it as config values.
	if err := r.ParseMultipartForm(100); err != nil {
		log.Printf("form parsing failed: %+v", err)
		http.Error(w, "bad form data", http.StatusBadRequest)
		return
	}
	config := make(map[string]string)
	for k, val := range r.Form {
		switch k {
		case "deviceID":
			deviceID = val[0]
		case "apiToken":
			apiToken = val[0]
		case "installationID":
			installationID = val[0]
		case "background":
			background = val[0] == "true"
		default:
			config[k] = val[0]
		}
	}

	webp, err := b.loader.LoadApplet(config)

	payload, err := json.Marshal(
		TidbytPushJSON{
			DeviceID:       deviceID,
			Image:          webp,
			InstallationID: installationID,
			Background:     background,
		},
	)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(TidbytAPIPush, deviceID),
		bytes.NewReader(payload),
	)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
		return
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Tidbyt API returned status %s\n", resp.Status)
		w.WriteHeader(resp.StatusCode)
		fmt.Fprintln(w, err)

		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
}
