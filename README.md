Golang Discord Bot for the [Riot Games API](https://developer.riotgames.com/) to provide a less error prone OP.GG replacement in discord.

Zeri's goals are _reliability_, and _maintainability_.

### `log`

Zeri produces logs using Uber Zap logger

# TODO

- [x] Functioning Bot with extensible commands
- [x] Region selection
- [x] Get Riot PUUID from Riot ID & Tag
- [x] Get Summoner Level & Icon from Account Details
- [x] Functioning champion icons as emojis
- [x] Last 5 matches queried properly
- [ ] Fetch & display current LP & winrate
- [ ] Fetch & display top 3 most played champions & their winrates
- [ ] Display last 10 matches played with W/L - Champion played - KDA - timestamp
- [ ] Select at the bottom to choose which match details to see
- [ ] Matches Details

# Development

Needs config.json at the root with this format

```

{
  "token": "bot token",
  "appID": "application id",
  "BotPrefix": "/",
  "RiotApiKey": "riot api key"
}

```

Needs to run `utils/download_champs.go` each time a champion icon is updated or added, upload the new or updated version as emoji and update the `types/type.go` accordingly
