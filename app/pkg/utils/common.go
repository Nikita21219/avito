package utils

import (
	"log"
	"time"
)

// DoWithTries attempts to execute a given function multiple times with retries.
// It takes the provided function, the number of attempts, and a delay duration between attempts.
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
