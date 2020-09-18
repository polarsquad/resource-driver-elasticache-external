package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"log"

	"github.com/gorilla/mux"
	"humanitec.io/resources/driver-aws-external/internal/messages"
)

func DecodeSecretsHeader(secretsHeaderValue string) (map[string]interface{}, error) {
	decodedAccountHeader, err := base64.StdEncoding.DecodeString(secretsHeaderValue)
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("secrets header does not seem to be encoded in base64: %w", err)
	}
	var secrets map[string]interface{}
	err = json.Unmarshal(decodedAccountHeader, &secrets)
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("cannot parse decoded secrets header : %w", err)
	}
	return secrets, nil
}

// createOrUpdateAWSResource
func (s *Server) createOrUpdateAWSResource(w http.ResponseWriter, r *http.Request) {
	var drd messages.DriverResourceDefinition
	if !readAsJSON(w, r, &drd) {
		return
	}

	metadata, metadataExists, err := s.Model.SelectResourceMetadata(drd.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := messages.ValuesSecrets{
		Values:  map[string]interface{}{},
		Secrets: map[string]interface{}{},
	}
	if metadataExists {
		data.Values = metadata.Data
	} else {
		metadata.ID = drd.ID
		metadata.Type = drd.Type
		metadata.CreatedAt = time.Now().UTC()
		metadata.Params = drd.DriverParams

		if _, exists := drd.DriverSecrets["account"]; !exists {
			log.Println(`"account" property in driver_secrets is missing`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		awsCreds, err := AccountMapToAWSCredentials(drd.DriverSecrets["account"])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch drd.Type {
		case "s3":
			data, err = s.createS3Bucket(drd, awsCreds)
		case "redis":
			data, err = s.createRedis(drd, awsCreds)
		default:
			log.Printf(`Type "%s" not supported by this driver.`, metadata.Type)
			writeAsJSON(w, http.StatusBadRequest, fmt.Sprintf(`Type "%s" not supported by this driver.`, metadata.Type))
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		metadata.Data = data.Values
		err = s.Model.InsertOrUpdateResourceMetadata(metadata)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	awsCreds, err := AccountMapToAWSCredentials(drd.DriverSecrets["account"])
	if err != nil {
		log.Printf("Unable to read account as AWS credentials: %v", err)
		writeAsJSON(w, http.StatusBadRequest, "Unable to read account as AWS credentials")
		return
	}
	data.Secrets["aws_access_key_id"] = awsCreds.AccessKeyID
	data.Secrets["aws_secret_access_key"] = awsCreds.SecretAccessKey

	writeAsJSON(w, http.StatusOK, messages.ResourceData{
		Type:       metadata.Type,
		Data:       data,
		DriverType: "aws",
	})
	return
}

// deleteAWSResource
func (s *Server) deleteAWSResource(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if !isValidAsID(params["resourceId"]) {
		writeAsJSON(w, http.StatusNotFound, fmt.Sprintf("Resource not found: %s", params["resourceId"]))
		return
	}

	if r.Header.Get("Humanitec-Driver-Secrets") == "" {
		log.Print(`Missing HTTP header "Humanitec-Driver-Secrets"`)
		writeAsJSON(w, http.StatusBadRequest, `Missing HTTP header "Humanitec-Driver-Secrets"`)
		return
	}
	secrets, err := DecodeSecretsHeader(r.Header.Get("Humanitec-Driver-Secrets"))
	if err != nil {
		log.Printf(`Unable to decode "Humanitec-Driver-Secrets" header: %v`, err)
		writeAsJSON(w, http.StatusBadRequest, `Malformed HTTP header "Humanitec-Driver-Secrets"`)
		return
	}
	if _, exists := secrets["account"]; !exists {
		log.Print(`Decoded "Humanitec-Driver-Secrets" header is missing "account" key`)
		writeAsJSON(w, http.StatusBadRequest, `Decoded "Humanitec-Driver-Secrets" header is missing "account" key`)
		return
	}
	awsCreds, err := AccountMapToAWSCredentials(secrets["account"])
	metadata, metadataExists, err := s.Model.SelectResourceMetadata(params["resourceId"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !metadataExists {
		writeAsJSON(w, http.StatusNotFound, fmt.Sprintf("Resource not found: %s", params["resourceId"]))
		return
	}
	switch metadata.Type {
	case "s3":
		err = s.deleteS3Bucket(metadata.Data["bucket"].(string), metadata.Params["region"].(string), awsCreds)
		if err != nil {
			log.Printf(`Error deleting bucket "%s": %v`, metadata.Data["bucket"], err)
			writeAsJSON(w, http.StatusBadRequest, fmt.Sprintf(`Error deleting bucket "%s": %v`, metadata.Data["bucket"], err))
			return
		}
	case "redis":
		err = s.deleteS3Bucket(metadata.Data["bucket"].(string), metadata.Params["region"].(string), awsCreds)
		if err != nil {
			log.Printf(`Error deleting bucket "%s": %v`, metadata.Data["bucket"], err)
			writeAsJSON(w, http.StatusBadRequest, fmt.Sprintf(`Error deleting bucket "%s": %v`, metadata.Data["bucket"], err))
			return
		}
	default:
		log.Printf(`Type "%s" not supported by this driver.`, metadata.Type)
		writeAsJSON(w, http.StatusBadRequest, fmt.Sprintf(`Type "%s" not supported by this driver.`, metadata.Type))
		return
	}

	err = s.Model.DeleteResourceMetadata(params["resourceId"], time.Now().UTC())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNoContent)
}
