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

const (
	httpPort = "3000"
	city     = "moscow"
)

type Reading struct {
	Timestamp   time.Time `json:"timestamp"`
	Temperature float64   `json:"temperature"`
}

type Storage struct {
	date map[string][]Reading
	mu   sync.RWMutex
}

func main() {
	wg := sync.WaitGroup{}

	storage := &Storage{
		date: make(map[string][]Reading),
	}
	addr := ":" + httpPort
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		cityName := chi.URLParam(r, "city")

		storage.mu.RLock()
		defer storage.mu.RUnlock()

		reading, ok := storage.date[cityName]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
			return
		}

		raw, err := json.Marshal(reading)
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

	jobs, err := initJobs(s, storage)
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

func initJobs(scheduler gocron.Scheduler, storage *Storage) ([]gocron.Job, error) {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	geocodingClient := geocoding.NewClient(httpClient)
	openMeteoClient := open_meteo.NewClient(httpClient)

	j, err := scheduler.NewJob(
		gocron.DurationJob(
			15*time.Second,
		),
		gocron.NewTask(
			func() {
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

				storage.mu.Lock()
				defer storage.mu.Unlock()

				timestamp, err := time.Parse("2006-01-02T15:04", openMeteoRes.Current.Time)
				if err != nil {
					log.Println(err)
					return
				}

				storage.date[city] = append(storage.date[city], Reading{
					Timestamp:   timestamp,
					Temperature: openMeteoRes.Current.Temperature2m,
				})

				fmt.Printf("update data for city: %s\n", city)
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return []gocron.Job{j}, nil
}
