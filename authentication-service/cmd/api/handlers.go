package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

func (app *config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJson(w, r, &requestPayload)

	//debugging info

	log.Printf("request pay load body: %v\n", requestPayload)
	log.Printf("Got email: %s and password: %s\n", requestPayload.Email, requestPayload.Password)
	log.Printf("Error: %v\n", err)

	if err != nil {
		app.errorJson(w, err, http.StatusBadRequest)
		return
	}

	// validating the user from the credentials that we got from user

	user, err := app.Models.User.GetByEmail(requestPayload.Email)

	// debugging info
	log.Printf("User from DB: %v\n", user)
	log.Printf("users password %v\n", user.Password)

	if err != nil {
		app.errorJson(w, errors.New("invalid credentials passed"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	log.Printf("Password match: %v\n", valid)

	if err != nil || !valid {
		app.errorJson(w, errors.New("wrong password"), http.StatusBadRequest)
		return
	}

	// log Authentication success

	err = app.logRequest("authentication", fmt.Sprintf("User %s logged in", user.Email))
	if err != nil {
		app.errorJson(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.writeJson(w, http.StatusOK, payload)
}

func (app *config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	log.Println("Logging to remote service")

	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServideUrl := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServideUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}
