package main

import (
	"blog/data"
	"context"
	"log"
	"net/http"
	"time"
)

type JsonPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *config) getBlog(w http.ResponseWriter, r *http.Request) {
	//reas json into var

	//check the cache
	ctx := context.Background()
	val, err := app.redclinet.Get(ctx, "foo").Result()
	log.Printf("sucessful connection with redis clinet", val)
	if err != nil {

		log.Printf("unable to find the record in cache", err)
		data, err := app.Models.BlogEntry.All()
		if err != nil {
			app.errorJson(w, err)
			return
		}

		var message string
		if len(data) > 0 {
			message = data[0].Data
		} else {
			message = "No blog entries found"
		}

		resp := jsonResponse{
			Error:   false,
			Message: message,
		}
		log.Printf("resp from blog aervice", resp)
		app.writeJson(w, http.StatusAccepted, resp)
		app.redclinet.Set(ctx, "foo", message, 5*time.Minute).Result()
		log.Printf("wrote the msg to redis")

	}
	resp := jsonResponse{
		Error:   false,
		Message: val + "coming from redis cache",
	}
	log.Printf("This came from redis cache not from mongodb", val)
	app.writeJson(w, http.StatusAccepted, resp)

}

func (app *config) pushBlog(w http.ResponseWriter, r *http.Request) {
	var requestPayload JsonPayload

	_ = app.readJson(w, r, &requestPayload)
	log.Printf("data", requestPayload.Data)

	// inserting the data

	event := data.BlogEntry{
		Name:      requestPayload.Name,
		Data:      requestPayload.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := app.Models.BlogEntry.Insert(event)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	resp := jsonResponse{
		Error:   false,
		Message: "we have pushed the blog" + event.Data,
	}

	app.writeJson(w, http.StatusAccepted, resp)
}

func (app *config) test(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the test button which is in blog service",
	}

	_ = app.writeJson(w, http.StatusOK, payload)
}
