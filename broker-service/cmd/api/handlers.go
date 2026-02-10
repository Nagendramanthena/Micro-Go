package main

import (
	"broker/event"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Requestpayload struct {
	Action string      `json:"action"`
	Auth   Authpayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   Mailpayload `json:"mail,omitempty"`
}

type Mailpayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type Authpayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *config) broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJson(w, http.StatusOK, payload)
}

func (app *config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestpayload Requestpayload

	err := app.readJson(w, r, &requestpayload)

	if err != nil {
		app.errorJson(w, err)
		return
	}

	switch requestpayload.Action {
	case "auth":
		log.Printf("Handling authentication for %s\n", requestpayload)
		log.Printf("Handling authentication for %s\n", requestpayload.Auth)
		app.authenticate(w, r, requestpayload.Auth)
	case "log":
		log.Printf("Handling logging for %s\n", requestpayload.Log)
		app.logEventViaRabbitmq(w, requestpayload.Log)

	case "mail":
		app.sendMail(w, requestpayload.Mail)

	case "blog":
		log.Printf("Blogdata", requestpayload.Log.Data)
		log.Printf("Blogdata", requestpayload.Log)
		app.PushBlog(w, requestpayload.Log)

	default:
		app.errorJson(w, errors.New("unknown action"))
	}
}

func (app *config) logItem(w http.ResponseWriter, r *http.Request, entry LogPayload) {
	//create some json we'll send to the log microservice
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJson(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		log.Printf("log service responded with status: %d\n", response.StatusCode)
		app.errorJson(w, errors.New("error calling log service"))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.writeJson(w, http.StatusAccepted, payload)

}

func (app *config) authenticate(w http.ResponseWriter, r *http.Request, a Authpayload) {
	//create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	log.Printf("JSON to auth service: %s\n", jsonData)

	//call the service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))

	//
	if err != nil {
		app.errorJson(w, err)
		return
	}

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	defer response.Body.Close()
	//make sure we get back the right status code

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJson(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusOK {
		app.errorJson(w, errors.New("error calling auth service"))
		return
	}

	//Now we need to create a variable we'll read the response body into
	var jsonFromService jsonResponse

	//decode the json from the auth service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJson(w, errors.New(jsonFromService.Message))
		return
	}

	//if we get here then we have been authenticated and can send back to the user
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromService.Data

	app.writeJson(w, http.StatusOK, payload)
}

func (app *config) sendMail(w http.ResponseWriter, m Mailpayload) {
	//create some json we'll send to the mail microservice
	jsonData, _ := json.MarshalIndent(m, "", "\t")

	req, err := http.NewRequest("POST", "http://mail-service/send", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJson(w, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		app.errorJson(w, errors.New("error calling mail service"))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Message sent to " + m.To

	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *config) logEventViaRabbitmq(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)

	if err != nil {
		app.errorJson(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logges via rabbitmq"

	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *config) pushToQueue(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)

	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")

	err = emitter.Push(string(j), "log.INFO")

	if err != nil {
		return err
	}

	return nil

}

func (app *config) PushBlog(w http.ResponseWriter, entry LogPayload) {
	//create some json we'll send to the log microservice
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://blog-service/addBlog"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJson(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		log.Printf("blog service responded with status: %d\n", response.StatusCode)
		app.errorJson(w, errors.New("error calling blog service"))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Pushed the blog with data " + entry.Data

	app.writeJson(w, http.StatusAccepted, payload)

}

func (app *config) readBlog(w http.ResponseWriter, r *http.Request) {

	logServiceURL := "http://blog-service/getblog"
	log.Printf("request", r)

	request, err := http.NewRequest("GET", logServiceURL, nil)

	if err != nil {
		app.errorJson(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		log.Printf("blog service responded with status: %d\n", response.StatusCode)
		app.errorJson(w, errors.New("error calling read blog service"))
		return
	}

	var blogPayload jsonResponse
	err = json.NewDecoder(response.Body).Decode(&blogPayload)
	if err != nil {
		log.Println("Error decoding blog service response:", err)
		app.errorJson(w, err)
		return
	}

	log.Printf("blogpayload from borker service", blogPayload)

	app.writeJson(w, http.StatusAccepted, blogPayload)
}

func (app *config) test(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the test button",
	}

	_ = app.writeJson(w, http.StatusOK, payload)
}
