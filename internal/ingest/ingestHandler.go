package ingest

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	p "log-analyzer/internal/parser"
)

const (
	outputFile = "access.log"
)

func IngestHandler(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	les, err := p.ParseJson(bodyBytes)
	if err != nil {
		// errString := fmt.Sprintf("Unable to parse log as a JSON string: %s", err)
		// http.Error(w, errString, http.StatusBadRequest)
		http.Error(w, "!!", http.StatusBadRequest)
		return
	}

	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	for _, le := range les {
		pString := fmt.Sprintln("Date:", le.Date, "Pod:", le.K8sMetadata.PodName, "Log:", le.Log)
		if _, err := f.Write([]byte(pString + "\n")); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("End of handle ingest")
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
