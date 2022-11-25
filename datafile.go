package main

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	alexaDir  = "https://s3.amazonaws.com/alexa-static"
	alexaFile = "top-1m.csv.zip"
)

func getTop1mFileFromAWS(dataDir string) ([]byte, error) {
	fileName := filepath.Join(dataDir, alexaFile)
	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		resp, err := http.Get(alexaDir + "/" + alexaFile)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile(fileName, b, 0644)
		if err != nil {
			return nil, err
		}
	}

	return os.ReadFile(fileName)
}

func parseTop1mFile(data []byte) ([]string, error) {
	var domains []string

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	fn := func(f *zip.File) error {
		fr, err := f.Open()
		if err != nil {
			return err
		}
		defer fr.Close()

		cr := csv.NewReader(fr)

		records, err := cr.ReadAll()
		if err != nil {
			return err
		}

		for _, record := range records {
			domain := record[1]
			domains = append(domains, domain)
		}
		return nil
	}

	for _, f := range r.File {
		if f.Name != strings.TrimSuffix(alexaFile, ".zip") {
			continue
		}
		err = fn(f)
		if err != nil {
			return nil, err
		}
	}

	return domains, nil
}

func GetTop1mDomains(dataDir string) (domains []string, err error) {
	buf, err := getTop1mFileFromAWS(dataDir)
	if err != nil {
		return nil, err
	}

	domains, err = parseTop1mFile(buf)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

func GetTestDomains(filepath string) (domains []string, err error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	return lines, nil
}
