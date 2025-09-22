package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/scottcrawford03/espn-api/internal/models"
)

const (
	BASE_URL  = "https://lm-api-reads.fantasy.espn.com/apis/v3/games/ffl"
	LEAGUE_ID = "1809227"
)

var (
	cookies = []*http.Cookie{
		// {
		// 	Name:  "espn_s2",
		// 	Value: "AECHqz9YhlmhYFP43%2FVuGxSUgPoXZ8ngxm%2BWaMkt4PrIIUDLg6CPt6z34WXQQRJsYO5nPGBSxvgWFAHGvNspIvqoy8G%2BiZJgF3dVc5zsjGf4BPMk%2BRMfELd3g1XwnvdtWn41nkRTc9r9j3GEWbiS5MOyLjNzR722esBcVFBo560gvfU05fv%2BE1Z%2FPsYJ%2F2aBMyR1S9uPA1w2MLg%2F4%2BA1%2FIe%2FTi%2FQdAk9kPFC9d7Xrn6eOCbMNd2nFGfhIsOtUmsnFjo2Qy1%2FzOx58MzGDUTjHBYNlfDRxZLV6y57WtHqQTnVlA%3D%3D",
		// },
		{
			Name: "SWID", Value: "{A442F34F-73BA-4D75-8996-1C2DF9261046}",
		},
	}
)

type MatchupResponse struct {
	CurrentMatchups []CurrentMatchup `json:"currentMatchups"`
	Teams           []TeamInfo       `json:"teams"`
	Schedule        []ScheduleItem   `json:"schedule"`
}

type CurrentMatchup struct {
	Home TeamInfo `json:"home"`
	Away TeamInfo `json:"away"`
}

type TeamInfo struct {
	Id     int     `json:"id"`
	Name   string  `json:"name"`
	Owner  string  `json:"owner"`
	Points float32 `json:"points"`
}

type ScheduleItem struct {
	Id              int                `json:"id"`
	MatchupPeriodId int                `json:"matchupPeriodId"`
	PlayoffTierType string             `json:"playoffTierType"`
	Winner          string             `json:"winner"`
	Home            ScheduleTeamInfo   `json:"home"`
	Away            ScheduleTeamInfo   `json:"away"`
}

type ScheduleTeamInfo struct {
	TeamId          int     `json:"teamId"`
	TotalPoints     float32 `json:"totalPoints"`
	TotalPointsLive float32 `json:"totalPointsLive"`
}

func main() {
	http.HandleFunc("/v1/matchups", matchupsHandler)

	log.Println("Server starting on :8001")
	log.Fatal(http.ListenAndServe(":8001", nil))
}

func matchupsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Matchups handler called")
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("getting espn_s2")
	espn_s2 := r.URL.Query().Get("espn_s2")
	if espn_s2 == "" {
		http.Error(w, "espn_s2 is required", http.StatusBadRequest)
		return
	}

	espns2Cookie := &http.Cookie{
		Name:  "espn_s2",
		Value: espn_s2,
	}

	client := &http.Client{}
	url := fmt.Sprintf("%s%s%s?view=mMatchupScore&view=mStatus&view=mSettings&view=mTeam&view=modular&view=mNav", BASE_URL, "/seasons/2025/segments/0/leagues/", LEAGUE_ID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, fmt.Sprintf("Error creating request: %v", err), http.StatusInternalServerError)
		return
	}

	for _, c := range cookies {
		req.AddCookie(c)
	}
	req.AddCookie(espns2Cookie)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		http.Error(w, fmt.Sprintf("Error making request: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading response: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Println("Body: ", string(body))

	espnResp := models.ESPNResponse{}
	err = json.Unmarshal(body, &espnResp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error unmarshaling response: %v", err), http.StatusInternalServerError)
		return
	}

	teamsToId := map[int]models.Team{}

	// Process teams and members
	for _, team := range espnResp.Teams {
		teamsToId[team.Id] = team

		for _, member := range espnResp.Members {
			if member.Id == team.PrimaryOwner {
				team.Member = member
				break
			}
		}
	}

	// Find current matchups
	var currentMatchups []CurrentMatchup
	for _, s := range espnResp.Schedule {
		if s.MatchupPeriodId == espnResp.Status.CurrentMatchupPeriod {
			home := teamsToId[s.Home.TeamId]
			away := teamsToId[s.Away.TeamId]

			currentMatchups = append(currentMatchups, CurrentMatchup{
				Home: TeamInfo{
					Id:     home.Id,
					Name:   home.Name,
					Owner:  fmt.Sprintf("%s %s", home.Member.FristName, home.Member.LastName),
					Points: s.Home.TotalPointsLive, // Use live scores for current matchups
				},
				Away: TeamInfo{
					Id:     away.Id,
					Name:   away.Name,
					Owner:  fmt.Sprintf("%s %s", away.Member.FristName, away.Member.LastName),
					Points: s.Away.TotalPointsLive, // Use live scores for current matchups
				},
			})
		}
	}

	// Convert teams to response format
	var teams []TeamInfo
	for _, team := range teamsToId {
		teams = append(teams, TeamInfo{
			Id:     team.Id,
			Name:   team.Name,
			Owner:  fmt.Sprintf("%s %s", team.Member.FristName, team.Member.LastName),
			Points: team.Points,
		})
	}

	// Convert schedule to response format
	var scheduleItems []ScheduleItem
	for _, s := range espnResp.Schedule {
		scheduleItems = append(scheduleItems, ScheduleItem{
			Id:              s.Id,
			MatchupPeriodId: s.MatchupPeriodId,
			PlayoffTierType: s.PlayoffTierType,
			Winner:          s.Winner,
			Home: ScheduleTeamInfo{
				TeamId:          s.Home.TeamId,
				TotalPoints:     s.Home.TotalPoints,
				TotalPointsLive: s.Home.TotalPointsLive,
			},
			Away: ScheduleTeamInfo{
				TeamId:          s.Away.TeamId,
				TotalPoints:     s.Away.TotalPoints,
				TotalPointsLive: s.Away.TotalPointsLive,
			},
		})
	}

	// Create response
	response := MatchupResponse{
		CurrentMatchups: currentMatchups,
		Teams:           teams,
		Schedule:        scheduleItems,
	}

	// Print current scores to console
	printCurrentScores(currentMatchups)

	// Send JSON response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func printCurrentScores(matchups []CurrentMatchup) {
	fmt.Println("\n=== CURRENT FANTASY FOOTBALL SCORES ===")
	if len(matchups) == 0 {
		fmt.Println("No current matchups found.")
		return
	}
	
	for i, matchup := range matchups {
		fmt.Printf("\nMatchup %d:\n", i+1)
		fmt.Printf("  %s (%s): %.2f points\n", matchup.Home.Name, matchup.Home.Owner, matchup.Home.Points)
		fmt.Printf("  %s (%s): %.2f points\n", matchup.Away.Name, matchup.Away.Owner, matchup.Away.Points)
		
		if matchup.Home.Points > matchup.Away.Points {
			fmt.Printf("  🏆 %s is winning!\n", matchup.Home.Name)
		} else if matchup.Away.Points > matchup.Home.Points {
			fmt.Printf("  🏆 %s is winning!\n", matchup.Away.Name)
		} else {
			fmt.Printf("  🤝 It's a tie!\n")
		}
	}
	fmt.Println("=====================================\n")
}
