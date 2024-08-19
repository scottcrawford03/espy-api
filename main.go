package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/scottcrawford03/espn-api/internal/models"
)

const (
	BASE_URL  = "https://lm-api-reads.fantasy.espn.com/apis/v3/games/ffl"
	LEAGUE_ID = "1809227"
)

var (
	cookies = []*http.Cookie{
		// TODO
	}
)

func main() {
	client := &http.Client{}
	// ?view=mMatchupScore&view=mStatus&view=mSettings&view=mTeam&view=modular&view=mNav
	url := fmt.Sprintf("%s%s%s?view=mTeam", BASE_URL, "/seasons/2024/segments/0/leagues/", LEAGUE_ID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}

	espnResp := models.ESPNResponse{}
	err = json.Unmarshal(body, &espnResp)
	if err != nil {
		fmt.Print(err.Error())
	}

	for _, team := range espnResp.Teams {
		for _, member := range espnResp.Members {
			if member.Id == team.PrimaryOwner {
				team.Member = member
				break
			}
		}
		fmt.Printf("Team %s is owned by: %s %s\n", team.Name, team.Member.FristName, team.Member.LastName)
	}
}
