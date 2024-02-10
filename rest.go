package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type (
	Result struct {
		Result string `json:"result"`
		Emoji  string `json:"emoji"`
		IsNew  bool   `json:"isNew"`
	}
)

const (
	nothing = "Nothing"
)

func checkForPair(ctx context.Context, httpClient *http.Client, first, second string) (Result, error) {
	url := fmt.Sprintf("https://neal.fun/api/infinite-craft/pair?first=%s&second=%s", first, second)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Referer", "https://neal.fun/infinite-craft/")
	req.Header.Set("Accept", "application/json")
	if ua := os.Getenv("HTTP_USER_AGENT"); ua != "" {
		req.Header.Set("User-Agent", ua)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	var result Result
	if strings.EqualFold(os.Getenv("DEBUG_REST"), "true") {
		// "invalid character '<' looking for beginning of value" is a common error
		// this is here for when I was debugging that error.
		// as best I can tell, the api just occasionally returns a webpage instead of json

		rdr := io.LimitReader(resp.Body, 1024)
		buf := make([]byte, 1024)
		var n int
		n, err = rdr.Read(buf)
		if err != nil {
			return result, err
		}

		err = json.Unmarshal(buf[:n], &result)
		if err != nil {
			err = fmt.Errorf("%w\n%s", err, buf[:n])
		}
	} else {
		err = json.NewDecoder(io.LimitReader(resp.Body, 1024)).Decode(&result)
	}

	return result, err
}
