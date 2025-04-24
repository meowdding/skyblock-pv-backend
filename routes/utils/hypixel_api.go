package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func GetFromHypixel(path string) (*string, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.hypixel.net%s", path),
		nil,
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("API-Key", os.Getenv("HYPIXEL_API_KEY"))

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
