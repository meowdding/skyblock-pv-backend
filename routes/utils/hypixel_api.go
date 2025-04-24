package utils

import (
	"fmt"
	"io"
	"net/http"
)

func GetFromHypixel(ctx RouteContext, path string) (*string, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.hypixel.net%s", path),
		nil,
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("API-Key", ctx.Config.HypixelKey)

	client := http.Client{}

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data: %s", res.Status)
	}

	data, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	output := string(data)
	return &output, nil
}
