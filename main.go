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

type TeamResponse struct {
	Teams []TeamInfo `json:"teams"`
}

type TeamInfo struct {
	Id           int     `json:"id"`
	Name         string  `json:"name"`
	Owner        string  `json:"owner"`
	CurrentScore float32 `json:"currentScore"`
	PlayoffSeed  int     `json:"playoffSeed"`
}

func main() {
	http.HandleFunc("/v1/teams", teamsHandler)

	log.Println("Server starting on :8001")
	log.Fatal(http.ListenAndServe(":8001", nil))
}

func teamsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Teams handler called")
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

	// Convert teams to response format with current scores
	var teams []TeamInfo
	for _, team := range teamsToId {
		currentScore := getCurrentScore(espnResp, team.Id)
		teams = append(teams, TeamInfo{
			Id:           team.Id,
			Name:         team.Name,
			Owner:        fmt.Sprintf("%s %s", team.Member.FristName, team.Member.LastName),
			CurrentScore: currentScore,
			PlayoffSeed:  team.PlayoffSeed,
		})
	}

	// Create simplified response
	response := TeamResponse{
		Teams: teams,
	}

	// Print team information to console
	printTeamInfo(teams)

	// Send JSON response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func getCurrentScore(espnResp models.ESPNResponse, teamId int) float32 {
	// Look for the team's current live score in the schedule
	for _, s := range espnResp.Schedule {
		if s.MatchupPeriodId == espnResp.Status.CurrentMatchupPeriod {
			if s.Home.TeamId == teamId {
				return s.Home.TotalPointsLive
			}
			if s.Away.TeamId == teamId {
				return s.Away.TotalPointsLive
			}
		}
	}
	return 0.0 // Return 0 if no current score found
}

func printTeamInfo(teams []TeamInfo) {
	fmt.Println("\n=== FANTASY FOOTBALL TEAMS ===")
	if len(teams) == 0 {
		fmt.Println("No teams found.")
		return
	}

	for _, team := range teams {
		fmt.Printf("Seed %d: %s (%s) - %.2f points\n",
			team.PlayoffSeed, team.Name, team.Owner, team.CurrentScore)
	}
	fmt.Println("=============================\n")
}
