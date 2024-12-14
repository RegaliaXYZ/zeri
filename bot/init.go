package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func FetchCurrentPatch() (string, error) {
	url := "https://ddragon.leagueoflegends.com/api/versions.json"
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data := []string{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data[0], nil
}

type Champion struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ChampionFull struct {
	ID       string
	Name     string
	ImageURL string
}

func FetchChampionData(version string) (map[string]ChampionFull, error) {
	// URL for the champions' data
	url := fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/data/en_US/champion.json", version)

	// Perform HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var ChampData struct {
		Data map[string]Champion `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ChampData); err != nil {
		return nil, err
	}

	// Prepare the map to return
	champions := make(map[string]ChampionFull)

	// Populate the map with champion data
	for id, champ := range ChampData.Data {
		champions[id] = struct {
			ID       string
			Name     string
			ImageURL string
		}{
			champ.ID,
			champ.Name,
			fmt.Sprintf("https://ddragon.leagueoflegends.com/cdn/%s/img/champion/%s.png", version, champ.ID),
		}
	}

	return champions, nil
}
