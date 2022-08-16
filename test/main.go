package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

const serverPort = 8080
const qpsKey = "limit:qps"
const queryKey = "limit:query"
const qpsLimit = 50
const queryLimit = 100
const resultPath = "results"

func main() {
	fmt.Println("Staring test with ::")
	fmt.Println("Overall allowed requests :", queryLimit)
	fmt.Println("Query per second limit: ", qpsLimit)
	log.SetFlags(0)

	logFile, err := setLogger(fmt.Sprintf("%dreqs.log", queryLimit))
	if err != nil {
		log.Panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	cfg := &redis.Options{
		Addr: "localhost:6379",
	}
	cmd := redis.NewClient(cfg)

	ctx := context.Background()

	counter := 0

	if err = cmd.Del(ctx, "limit:query").Err(); err != nil {
		log.Print(err)
	}
	if err = cmd.Del(ctx, "limit:qps").Err(); err != nil {
		log.Print(err)
	}

	requestURL := fmt.Sprintf("http://localhost:%d/v1/test-limiter", serverPort)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		log.Printf("client: could not create request: %s\n", err)
		os.Exit(1)
	}

	log.Println("Overall allowed requests :", queryLimit)
	log.Println("Query per second limit: ", qpsLimit)

	log.Println("==================================================")
	log.Println("------------------- Positive tests -----------------")

	// // Overall request limit per group
	for i := 0; i < queryLimit/qpsLimit; i++ {

		// QPS limit
		for j := 0; j < qpsLimit; j++ {
			counter++
			res, err := http.DefaultClient.Do(req)

			if err != nil {
				log.Printf("error making http request %d : %s\n", counter, err)
				os.Exit(1)
			}
			if res.StatusCode != 200 {
				log.Printf("Error: status code of request %d : \t\t\t%d\n", counter, res.StatusCode)
			} else {
				log.Printf("Status code of request %d : \t\t\t%d\n", counter, res.StatusCode)
			}
			rc, err := cmd.Get(ctx, "limit:query").Int()
			if err != nil {
				log.Println(err)
			}
			log.Printf("Overall request count in Redis : \t%d\n", rc)
			if rc != counter {
				log.Println("Error: Request count not represented in redis")
			}

			qps, err := cmd.Get(ctx, "limit:qps").Int()
			if err != nil {
				log.Println(err)
			}
			log.Printf("QPS count in Redis : \t\t\t\t%d\n", qps)
			if qps != j+1 {
				log.Println("Error: QPS count incorrect in redis")
			}
			log.Println(" ")

		}

		log.Println("============================\nWaiting 1 sec to avoid qps limit.....")
		time.Sleep(1 * time.Second)
	}

	log.Println("\n----------Negative test: overall req limit reached -------------")
	for k := 0; k < qpsLimit; k++ {
		counter++
		res, err := http.DefaultClient.Do(req)

		if err != nil {
			log.Printf("error making http request %d : %s\n", counter, err)
			os.Exit(1)
		}

		if res.StatusCode != 429 {
			log.Printf("Error: status code of request %d : \t\t\t%d\n", counter, res.StatusCode)
		} else {
			log.Printf("Status code of request %d : \t\t\t%d\n", counter, res.StatusCode)
		}
		rc, err := cmd.Get(ctx, "limit:query").Int()
		if err != nil {
			log.Println(err)
		}
		log.Printf("Overall request count in Redis : \t%d\n", rc)

		if rc != queryLimit {
			log.Println("Error: Request count not correct in redis")
		}

		qps, err := cmd.Get(ctx, "limit:qps").Int()
		if err != nil {
			log.Println(err)
		}
		log.Printf("QPS count in Redis : \t\t\t\t%d\n", qps)

		if qps != k+1 {
			log.Println("Error: QPS count incorrect in redis")
		}
		log.Println(" ")
	}

	log.Println("\n----------Negative test: QPS limit reached -------------")
	log.Println("Setting back overlimit req count.")
	if err = cmd.Del(ctx, "limit:query").Err(); err != nil {
		log.Print(err)
	}

	log.Printf("Sending %d req/sec.....\n", (qpsLimit + 5))

	for k := 0; k < (qpsLimit + 5); k++ {
		counter++
		res, err := http.DefaultClient.Do(req)

		if err != nil {
			log.Printf("error making http request %d : %s\n", counter, err)
			os.Exit(1)
		}

		if res.StatusCode != 429 {
			log.Printf("Error: status code of request %d : \t\t\t%d\n", counter, res.StatusCode)
		} else {
			log.Printf("Status code of request %d : \t\t\t%d\n", counter, res.StatusCode)
		}
		rc, err := cmd.Get(ctx, "limit:query").Int()
		if err != nil {
			log.Println(err)
		}
		log.Printf("Overall request count in Redis : \t%d\n", rc)

		qps, err := cmd.Get(ctx, "limit:qps").Int()
		if err != nil {
			log.Println(err)
		}
		log.Printf("QPS count in Redis : \t\t\t%d\n", qps)

		log.Println(" ")

	}
}

func setLogger(filename string) (*os.File, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	filepath := fmt.Sprintf("%s/%s/%s", path, resultPath, filename)
	if !fileExists(filepath) {
		createFile(filepath)
	} else {
		if err := os.Remove(filepath); err != nil {
			log.Printf("error removing file: %v", err)
			return nil, err
		}
	}
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening file: %v", err)
		return nil, err
	}
	return f, nil
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func createFile(name string) error {
	fo, err := os.Create(name)
	if err != nil {
		return err
	}
	defer func() {
		fo.Close()
	}()
	return nil
}
