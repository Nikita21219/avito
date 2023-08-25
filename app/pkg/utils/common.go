package utils

import (
	"log"
	"time"
)

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
