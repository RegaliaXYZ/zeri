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
	"github.com/bwmarrin/discordgo"
)

var RiotCommand = types.Command{
	Name:        "rank",
	Description: "Fetches the last 5 matches for a Riot ID#Tag and Region",
	Handler:     riotHandler,
}

var Routing = ""

func riotHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Parse options for Riot ID and tag

	cfg, err := bot.ReadConfig()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error reading internal config.",
			},
		})
		return
	}
	options := i.ApplicationCommandData().Options

	var fullRiotID, riotID, tag, region string
	for _, opt := range options {
		switch opt.Name {
		case "riot_id":
			fullRiotID = opt.StringValue()
		case "region":
			region = opt.StringValue()
		}
	}
	switch region {
	case "EUW1", "EUN1", "ME1", "TR1", "RU":
		Routing = "europe"
	case "NA1", "LA1", "LA2", "BR1":
		Routing = "americas"
	case "KR", "JP1":
		Routing = "asia"
	case "OC1", "PH2", "SG2", "TH2", "TW2", "VN2":
		Routing = "sea"
	default:
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error with the region routing.",
			},
		})
		return
	}
	parts := strings.Split(fullRiotID, "#")
	riotID = parts[0]
	tag = parts[1]
	// Get Patch version
	patch, err := fetchCurrentPatch()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error fetching current patch.",
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
				Content: "Error fetching puuid from Riot API",
			},
		})
		return
	}
	// Get Account Profile Details
	profileIconID, summonerLevel, encryptedId, err := fetchAccountProfileDetail(puuid, region, cfg.RiotAPIKey)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error fetching account details",
			},
		})
		return
	}
	// Rank, LP, Wins & Losses
	queuetype, tier, rank, winrate, leaguePoints, wins, losses, err := fetchRankedDetails(encryptedId, region, cfg.RiotAPIKey)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error fetching ranked details.",
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
				Content: "Error fetching matches.",
			},
		})
		return
	}

	matchDetailsGlobals := []string{}
	total_wins := 0
	total_losses := 0
	for _, match := range matches {
		matchDetails, won, err := fetchMatchDetails(match, puuid, cfg.RiotAPIKey)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error fetching match details.",
				},
			})
			return
		}
		if won {
			total_wins++
		} else {
			total_losses++
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
				Name:   queuetype,
				Value:  tier + " " + rank + "\n" + strconv.Itoa(leaguePoints) + "LP\n" + strconv.Itoa(wins+losses) + "G " + strconv.Itoa(wins) + "W " + strconv.Itoa(losses) + "L " + strconv.FormatFloat(winrate, 'f', 2, 64) + "%",
				Inline: false,
			},
			{
				Name:   "Match History",
				Value:  "Recent 10 Games",
				Inline: false,
			},
			{
				Name:   fmt.Sprintf("10G %sW %sL", strconv.Itoa(total_wins), strconv.Itoa(total_losses)),
				Value:  strings.Join(matchDetailsGlobals, "\n"),
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
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func fetchCurrentPatch() (string, error) {
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

func fetchRankedDetails(encryptedId, region, api_key string) (string, string, string, float64, int, int, int, error) {
	riotBaseURL := fmt.Sprintf("https://%s.api.riotgames.com", strings.ToLower(region))
	// Fetch match history using PUUID
	matchURL := fmt.Sprintf("%s/lol/league/v4/entries/by-summoner/%s", riotBaseURL, encryptedId)
	req, _ := http.NewRequest("GET", matchURL, nil)
	req.Header.Set("X-Riot-Token", api_key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", 0.0, 0, 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", "", 0.0, 0, 0, 0, fmt.Errorf("riot api error: %s", string(body))
	}

	var accountProfileDetailData []struct {
		QueueType    string `json:"queueType"`
		Tier         string `json:"tier"`
		Rank         string `json:"rank"`
		LeaguePoints int    `json:"leaguePoints"`
		Wins         int    `json:"wins"`
		Losses       int    `json:"losses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&accountProfileDetailData); err != nil {
		return "", "", "", 0.0, 0, 0, 0, err
	}
	if len(accountProfileDetailData) == 0 {
		return "Unranked", "", "", 0.0, 0, 0, 0, err
	}
	detailData := accountProfileDetailData[0]

	winrate := (float64(detailData.Wins) / (float64(detailData.Wins) + float64(detailData.Losses))) * 100

	return detailData.QueueType, detailData.Tier, detailData.Rank, winrate, detailData.LeaguePoints, detailData.Wins, detailData.Losses, nil
}
func getPUUID(riotID, tag, region, api_key string) (string, error) {
	riotBaseURL := fmt.Sprintf("https://%s.api.riotgames.com", Routing)
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
		return "", fmt.Errorf("riot api error: %s", string(body))
	}

	var summonerData struct {
		Puuid string `json:"puuid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&summonerData); err != nil {
		return "", err
	}
	return summonerData.Puuid, nil
}

func fetchAccountProfileDetail(puuid, region, api_key string) (string, string, string, error) {
	riotBaseURL := fmt.Sprintf("https://%s.api.riotgames.com", strings.ToLower(region))
	// Fetch match history using PUUID
	matchURL := fmt.Sprintf("%s/lol/summoner/v4/summoners/by-puuid/%s", riotBaseURL, puuid)
	req, _ := http.NewRequest("GET", matchURL, nil)
	req.Header.Set("X-Riot-Token", api_key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", "", fmt.Errorf("riot api error: %s", string(body))
	}

	var accountProfileDetailData struct {
		ProfileIconID int    `json:"profileIconId"`
		SummonerLevel int    `json:"summonerLevel"`
		EncryptedID   string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&accountProfileDetailData); err != nil {
		return "", "", "", err
	}

	return strconv.Itoa(accountProfileDetailData.ProfileIconID), strconv.Itoa(accountProfileDetailData.SummonerLevel), accountProfileDetailData.EncryptedID, nil
}

func fetchMatches(puuid, api_key string) ([]string, error) {
	// Convert Riot ID and tag to a PUUID (requires API call)
	riotBaseURL := fmt.Sprintf("https://%s.api.riotgames.com", Routing)
	// Fetch match history using PUUID
	matchURL := fmt.Sprintf("%s/lol/match/v5/matches/by-puuid/%s/ids?count=10", riotBaseURL, puuid)
	req, _ := http.NewRequest("GET", matchURL, nil)
	req.Header.Set("X-Riot-Token", api_key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("riot api error: %s", string(body))
	}

	var matchIDs []string
	if err := json.NewDecoder(resp.Body).Decode(&matchIDs); err != nil {
		return nil, err
	}

	return matchIDs, nil
}

func fetchMatchDetails(matchID, puuid, api_key string) (string, bool, error) {
	url := fmt.Sprintf("https://europe.api.riotgames.com/lol/match/v5/matches/%s", matchID) // Replace <region>
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Riot-Token", api_key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", false, fmt.Errorf("riot api error: %s", string(body))
	}

	var matchData struct {
		Info struct {
			Participants []struct {
				RiotID       string `json:"riotIdGameName"`
				RiotTag      string `json:"riotIdTagline"`
				Puuid        string `json:"puuid"`
				ChampionName string `json:"championName"`
				Kills        int    `json:"kills"`
				Deaths       int    `json:"deaths"`
				Assists      int    `json:"assists"`
				TeamID       int    `json:"teamId"`
				Win          bool   `json:"win"`
			} `json:"participants"`
			QueueId int `json:"queueId"`
		} `json:"info"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&matchData); err != nil {
		return "", false, err
	}
	//fmt.Println(matchData)

	matchLine := ""
	won := false
	for _, p := range matchData.Info.Participants {
		if p.Puuid == puuid {
			if p.Win {
				matchLine += ":regional_indicator_w: "
				won = true
			} else {
				matchLine += types.Loss + " "
			}
			kda := "Perfect"
			if p.Deaths != 0 {
				calculated_kda := float64(p.Kills+p.Assists) / float64(p.Deaths)
				kda = fmt.Sprintf("%.1f", calculated_kda)
			}
			matchLine += fmt.Sprintf("%s %s %d/%d/%d %s KDA", types.GameModes[matchData.Info.QueueId], types.Champions[p.ChampionName], p.Kills, p.Deaths, p.Assists, kda)
		}
	}

	return matchLine, won, nil
}
