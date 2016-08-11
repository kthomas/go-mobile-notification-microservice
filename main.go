package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/anachronistic/apns"
)

func main() {
	router := httprouter.New()
	router.POST("/", MobileNotificationHandler)

	log.Fatal(http.ListenAndServe("0.0.0.0:8080", router))
}

func parse(obj interface{}, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&obj)
	if err != nil {

	}
}

func render(obj interface{}, status int, w http.ResponseWriter) {
	w.Header().Set("content-type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if &obj != nil && status != http.StatusNoContent {
		if err := json.NewEncoder(w).Encode(obj); err != nil {
			panic(err)
		}
	} else {
		w.Header().Set("content-length", "0")
	}
}

func renderError(message string, status int, w http.ResponseWriter) {
	err := map[string]*string{}
	err["message"] = &message
	render(err, status, w)
}

type ApnsNotification struct {
	Environment		string		`json:"environment,omitempty"`
	Certificate		string		`json:"certificate,omitempty"`
	PrivateKey		string		`json:"key,omitempty"`
	DeviceToken		string		`json:"device_token,omitempty"`
	Alert			ApnsAlert	`json:"alert,omitempty"`
	ContentAvailable	bool		`json:"content_available,omitempty"`
	Badge			int		`json:"badge,omitempty"`
	Sound			string		`json:"sound,omitempty"`
}

type ApnsAlert struct {
	Title        string   `json:"title,omitempty"`
	Body         string   `json:"body,omitempty"`
	TitleLocKey  string   `json:"title_loc_key,omitempty"`
	TitleLocArgs []string `json:"title_loc_args,omitempty"`
	ActionLocKey string   `json:"action_loc_key,omitempty"`
	LocKey       string   `json:"loc_key,omitempty"`
	LocArgs      []string `json:"loc_args,omitempty"`
	LaunchImage  string   `json:"launch_image,omitempty"`
}

func MobileNotificationHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	MobileNotification(w, r)
}

func MobileNotification(w http.ResponseWriter, r *http.Request) {
	var params ApnsNotification
	parse(&params, r)

	var endpoint string
	endpoint = "gateway.push.apple.com:2195"
	if &params.Environment != nil && strings.Compare(params.Environment, "development") == 0 {
		endpoint = "gateway.sandbox.push.apple.com:2195"
 	}

	certificate := params.Certificate
	privateKey := params.PrivateKey
	deviceToken := params.DeviceToken

	payload := apns.NewPayload()

	if &params.Alert != nil {
		payload.Alert = params.Alert
	}

	if &params.Badge != nil {
		payload.Badge = params.Badge
	}

	if &params.ContentAvailable != nil && params.ContentAvailable {
		payload.ContentAvailable = 1
	}

	if &params.Sound != nil {
		payload.Sound = params.Sound
	}

	pn := apns.NewPushNotification()
	pn.DeviceToken = deviceToken
	pn.AddPayload(payload)

	client := apns.BareClient(endpoint, certificate, privateKey)
	resp := client.Send(pn)

	if resp.Success {
		render(nil, http.StatusNoContent, w)
	} else {
		renderError(resp.Error.Error(), http.StatusInternalServerError, w)
	}
}
