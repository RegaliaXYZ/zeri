package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/RegaliaXYZ/opgg-discord/bot"
	"github.com/RegaliaXYZ/opgg-discord/types"
	"github.com/RegaliaXYZ/opgg-discord/utils"
	"github.com/bwmarrin/discordgo"
)

var RiotCommand = types.Command{
	Name:        "get-matches",
	Description: "Fetches the last 5 matches for a Riot ID and tag",
	Handler:     riotHandler,
}

func riotHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Parse options for Riot ID and tag

	cfg, err := bot.ReadConfig()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error reading config: %v", err),
			},
		})
		return
	}
	options := i.ApplicationCommandData().Options
	if len(options) != 2 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide both Riot ID and tag. Usage: `/get-matches riot_id:username tag:tag`",
			},
		})
		return
	}

	var fullRiotID, riotID, tag, region string
	for _, opt := range options {
		switch opt.Name {
		case "riot_id":
			fullRiotID = opt.StringValue()
		case "region":
			region = opt.StringValue()
		}
	}

	parts := strings.Split(fullRiotID, "#")
	riotID = parts[0]
	tag = parts[1]
	// Get Patch version
	patch, err := utils.FetchCurrentPatch()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error fetching matches: %v", err),
			},
		})
		return
	}
	// Get PUUID from riot id
	puuid, err := getPUUID(riotID, tag, region, cfg.RiotAPIKey)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error fetching matches: %v", err),
			},
		})
		return
	}
	// Get Account Profile Details
	profileIconID, summonerLevel, err := fetchAccountProfileDetail(puuid, region, cfg.RiotAPIKey)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error fetching matches: %v", err),
			},
		})
		return
	}
	// Fetch matches using Riot API
	matches, err := fetchMatches(puuid, cfg.RiotAPIKey)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error fetching matches: %v", err),
			},
		})
		return
	}

	matchDetailsGlobals := []string{}
	for _, match := range matches {
		matchDetails, err := fetchMatchDetails(match, cfg.RiotAPIKey)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Error fetching match details: %v", err),
				},
			})
			return
		}
		matchDetailsGlobals = append(matchDetailsGlobals, matchDetails)

	}
	// Respond with the matches
	// matchList := strings.Join(matchDetailsGlobals, "\n")
	embed := &discordgo.MessageEmbed{
		Title:       "Regalia#xyz",
		URL:         "https://www.op.gg/summoners/euw/Regalia-xyz",
		Description: "LV." + summonerLevel,
		Color:       0x0000ff, // Green color for the embed
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Ranked Solo/Duo",
				Value:  "Master 1\n10LP\n46G 29W 17L 63%",
				Inline: false,
			},
			{
				Name:   "Most Champion 3",
				Value:  "Zeri 5.0 KDA 80%\nKalista 1.8 KDA 100%\nJinx 10.0KDA 100%",
				Inline: false,
			},
			{
				Name:   "Match History",
				Value:  "Recent 10 Games",
				Inline: false,
			},
			{
				Name:   "10G 10W 0L",
				Value:  ":regional_indicator_w: " + types.Zeri + " Ranked Solo/Duo 10/0/0 10.0KDA",
				Inline: false,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://ddragon.leagueoflegends.com/cdn/" + patch + "/img/profileicon/" + profileIconID + ".png",
		},
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "pong",                           // Optional text content
			Embeds:  []*discordgo.MessageEmbed{embed}, // Attach the embed here
		},
	})
}
func getPUUID(riotID, tag, region, api_key string) (string, error) {
	riotBaseURL := "https://europe.api.riotgames.com" // Replace <region> with the appropriate region
	url := fmt.Sprintf("%s/riot/account/v1/accounts/by-riot-id/%s/%s", riotBaseURL, riotID, tag)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Riot-Token", api_key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Riot API error: %s", string(body))
	}

	var summonerData struct {
		Puuid string `json:"puuid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&summonerData); err != nil {
		return "", err
	}
	return summonerData.Puuid, nil
}

func fetchAccountProfileDetail(puuid, region, api_key string) (string, string, error) {
	riotBaseURL := fmt.Sprintf("https://%s.api.riotgames.com", strings.ToLower(region))
	// Fetch match history using PUUID
	matchURL := fmt.Sprintf("%s/lol/summoner/v4/summoners/by-puuid/%s", riotBaseURL, puuid)
	req, _ := http.NewRequest("GET", matchURL, nil)
	req.Header.Set("X-Riot-Token", api_key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("Riot API error: %s", string(body))
	}

	var accountProfileDetailData struct {
		ProfileIconID int `json:"profileIconId"`
		SummonerLevel int `json:"summonerLevel"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&accountProfileDetailData); err != nil {
		return "", "", err
	}

	return strconv.Itoa(accountProfileDetailData.ProfileIconID), strconv.Itoa(accountProfileDetailData.SummonerLevel), nil
}

func fetchMatches(puuid, api_key string) ([]string, error) {
	// Convert Riot ID and tag to a PUUID (requires API call)
	riotBaseURL := "https://europe.api.riotgames.com"
	// Fetch match history using PUUID
	matchURL := fmt.Sprintf("%s/lol/match/v5/matches/by-puuid/%s/ids?count=5", riotBaseURL, puuid)
	req, _ := http.NewRequest("GET", matchURL, nil)
	req.Header.Set("X-Riot-Token", api_key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Riot API error: %s", string(body))
	}

	var matchIDs []string
	if err := json.NewDecoder(resp.Body).Decode(&matchIDs); err != nil {
		return nil, err
	}

	return matchIDs, nil
}

func fetchMatchDetails(matchID, api_key string) (string, error) {
	url := fmt.Sprintf("https://europe.api.riotgames.com/lol/match/v5/matches/%s", matchID) // Replace <region>
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Riot-Token", api_key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Riot API error: %s", string(body))
	}

	var matchData struct {
		Info struct {
			Participants []struct {
				RiotID       string `json:"riotIdGameName"`
				RiotTag      string `json:"riotIdTagline"`
				ChampionName string `json:"championName"`
				Kills        int    `json:"kills"`
				Deaths       int    `json:"deaths"`
				Assists      int    `json:"assists"`
				TeamID       int    `json:"teamId"`
			} `json:"participants"`
		} `json:"info"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&matchData); err != nil {
		return "", err
	}
	fmt.Println(matchData)

	team1 := "Team 1:\n"
	team2 := "Team 2:\n"
	for _, p := range matchData.Info.Participants {
		line := fmt.Sprintf("%s#%s - %s: %d/%d/%d\n", p.RiotID, p.RiotTag, p.ChampionName, p.Kills, p.Deaths, p.Assists)
		if p.TeamID == 100 {
			team1 += line
		} else {
			team2 += line
		}
	}

	return team1 + "\n" + team2, nil
}
