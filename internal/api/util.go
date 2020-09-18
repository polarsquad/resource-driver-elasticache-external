package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

func AccountMapToAWSCredentials(accountMap interface{}) (AWSCredentials, error) {

	asMap, isMap := accountMap.(map[string]interface{})
	if !isMap {
		return AWSCredentials{}, fmt.Errorf("expected map[string]interface{}, got %T", accountMap)
	}
	accessKeyAsString, isAccessKeyString := asMap["aws_access_key_id"].(string)
	if !isAccessKeyString {
		return AWSCredentials{}, fmt.Errorf(`expected "aws_access_key_id" to be string, got %T`, asMap["aws_access_key_id"])
	}
	secretAccessKeyAsString, isSecretAccessKeyString := asMap["aws_secret_access_key"].(string)
	if !isSecretAccessKeyString {
		return AWSCredentials{}, fmt.Errorf(`expected "aws_secret_access_key" to be string, got %T`, asMap["aws_secret_access_key"])
	}
	return AWSCredentials{
		AccessKeyID:     accessKeyAsString,
		SecretAccessKey: secretAccessKeyAsString,
	}, nil
}

var validID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]+[a-z0-9]$`)

// writeAsJSON writes the supplied object to a response along with the status code.
func writeAsJSON(w http.ResponseWriter, statusCode int, obj interface{}) {
	jsonObj, err := json.Marshal(obj)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonObj)
}

func readAsJSON(w http.ResponseWriter, r *http.Request, obj interface{}) bool {
	if r.Body == nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return false
	}

	err := json.NewDecoder(r.Body).Decode(obj)
	if nil != err {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return false
	}
	return true
}

func isValidAsID(str string) bool {
	return validID.MatchString(str)
}

func (s *Server) isAlive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) isReady(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
