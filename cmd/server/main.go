package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/M4rSHaLll/weather-service/internal/client/http/geocoding"
	open_meteo "github.com/M4rSHaLll/weather-service/internal/client/http/open-meteo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"
)

const httpPort = "3000"

func main() {
	wg := sync.WaitGroup{}

	addr := ":" + httpPort
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	geocodingClient := geocoding.NewClient(httpClient)
	openMeteoClient := open_meteo.NewClient(httpClient)

	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")

		geoRes, err := geocodingClient.GetCoords(city)
		if err != nil {
			log.Println(err)
			return
		}

		openMeteoRes, err := openMeteoClient.GetTemperature(geoRes.Latitude, geoRes.Longitude)
		if err != nil {
			log.Println(err)
			return
		}

		raw, err := json.Marshal(openMeteoRes)
		if err != nil {
			log.Println(err)
		}

		_, err = w.Write(raw)
		if err != nil {
			log.Println(err)
		}
	})

	s, err := gocron.NewScheduler()
	if err != nil {
		log.Println(err)
	}

	jobs, err := initJobs(s)
	if err != nil {
		panic(err)
	}

	wg.Add(2)
	go func() {
		defer wg.Done()
		fmt.Println("starting server on port " + httpPort)
		err = http.ListenAndServe(addr, r)

		if err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		fmt.Printf("starting job: %v\n", jobs[0].ID())
		s.Start()
	}()

	wg.Wait()
}

func initJobs(scheduler gocron.Scheduler) ([]gocron.Job, error) {
	j, err := scheduler.NewJob(
		gocron.DurationJob(
			15*time.Second,
		),
		gocron.NewTask(
			func() {
				fmt.Println("hello")
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return []gocron.Job{j}, nil
}
