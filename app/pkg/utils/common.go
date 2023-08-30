package utils

import (
	"gopkg.in/yaml.v3"
	"log"
	"main/internal/config"
	"os"
	"time"
)

// DoWithTries attempts to execute a given function multiple times with retries.
// If the provided function returns an error, it will retry according to the given attempts and delay.
// If all attempts are exhausted and the function still returns an error, the last error will be returned.
func DoWithTries(fn func() error, attempts int, delay time.Duration) (err error) {
	for attempts > 0 {
		if err = fn(); err != nil {
			log.Println("Error to connect. Try again...")
			time.Sleep(delay)
			attempts--
			continue
		}
		return nil
	}
	return err
}

func LoadConfig(path string) *config.Config {
	confStream, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln("Error to open read config file:", err)
	}

	conf := config.NewConfig()
	err = yaml.Unmarshal(confStream, conf)
	if err != nil {
		log.Fatalln("Error to unmarshal data from config file:", err)
	}
	return conf
}
