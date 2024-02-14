package commonutils

import (
	"github.com/dlintw/goconf"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"
)

// Return current test testName
func GetTestName(t *testing.T) interface{} {
	return t.Name()
}

// GetFunctionName Get a function Name given a function
func GetFunctionName(function interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
}

// GetErrorMessage Get an error message given an error
func GetErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// Parse time duration field from conf
func GetTimeFromConfig(c *goconf.ConfigFile, section string, field string, defaultTime time.Duration) time.Duration {
	var currentTime = defaultTime
	if c == nil {
		return currentTime
	}
	timeStr, err := c.GetString(section, field)
	if err != nil {
		return currentTime
	}
	parsedTime, err := time.ParseDuration(timeStr)
	if err != nil {
		return currentTime
	}
	currentTime = parsedTime
	return currentTime
}

// Parse time duration field from conf
func GetIntFromConfig(c *goconf.ConfigFile, section string, field string, defaultValue int) int {
	var current = defaultValue
	if c == nil {
		return defaultValue
	}
	val, err := c.GetInt(section, field)
	if err != nil {
		return defaultValue
	}
	current = val
	return current
}

// RemovePaths paths
func RemovePaths(paths []string) error {
	for i := range paths {
		err := os.RemoveAll(paths[i])
		if err != nil {
			if os.IsNotExist(err) {
				err = nil
				continue
			}
			return err
		} else {
		}
	}
	return nil
}

// Function to generate a random string of given length using timestamp as seed
func RandomStringUsingTimestamp(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
