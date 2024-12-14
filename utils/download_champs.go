package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
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

func downloadImage(url string, fileName string) error {
	// Make a GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download image from %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Create a new file

	out, err := os.Create("icons/" + strings.ToLower(fileName) + ".png")
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", fileName, err)
	}
	defer out.Close()

	// Copy the response body to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save image to %s: %v", fileName, err)
	}

	return nil
}

func main() {
	file, err := os.OpenFile("champs.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error creating/opening champs.txt: ", err)
		return
	}
	defer file.Close()
	patch, err := FetchCurrentPatch()
	if err != nil {
		fmt.Println("Error getting the current patch: ", err)
		return
	}
	champions, err := FetchChampionData(patch)
	if err != nil {
		fmt.Println("Error getting the champions: ", err)
		return
	}
	keys := make([]string, 0, len(champions))
	for key := range champions {
		keys = append(keys, key)
	}

	// Step 2: Sort the keys slice alphabetically
	sort.Strings(keys)

	for _, key := range keys {
		champ := champions[key]
		_, err := fmt.Fprintf(file, "%s = \"\"\n", champ.ID)
		if err != nil {
			fmt.Println("Error writing to champs.txt:", err)
		}
		err = downloadImage(champ.ImageURL, strings.ToLower(champ.ID))
		if err != nil {
			fmt.Println("Error downloading:", err)
		} else {
			fmt.Printf("Successfully downloaded: %s\n", champ.ID)
		}
	}
}
