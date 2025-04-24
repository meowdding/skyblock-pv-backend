package routes

import (
	"fmt"
	"io"
	"net/http"
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

	req.Header.Set("API-Key", "77578c67-0fad-40e5-8ae7-d95461019308")

	client := http.Client{}

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data: %s", res.Status)
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	output := string(data)
	return &output, nil
}
