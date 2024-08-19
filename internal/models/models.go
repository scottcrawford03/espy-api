package models

type ESPNResponse struct {
	Teams   []Team   `json:"teams"`
	Members []Member `json:"members"`
}

type Member struct {
	DisplyName           string      `json:"displayName"`
	FristName            string      `json:"firstName"`
	Id                   string      `json:"id"`
	LastName             string      `json:"lastName"`
	NotificationSettings interface{} `json:"notificationSettings"`
}

type Team struct {
	Abbrev                string      `json:"abbrev"`
	CurrentProjectedRank  int         `json:"currentProjectedRank"`
	DivisionId            int         `json:"divisionId"`
	DraftDayProjectedRank int         `json:"draftDayProjectedRank"`
	DraftStrategy         interface{} `json:"draftStrategy"`
	Id                    int         `json:"id"`
	IsActive              bool        `json:"isActive"`
	Logo                  string      `json:"logo"`
	LogoType              string      `json:"logoType"`
	Name                  string      `json:"name"`
	Owners                interface{} `json:"owners"`
	PlayoffSeed           int         `json:"playoffSeed"`
	Points                float32     `json:"points"`
	PointsAdjusted        float32     `json:"pointsAdjusted"`
	PointsDelta           float32     `json:"pointsDelta"`
	PrimaryOwner          string      `json:"primaryOwner"`
	RankCalculatedFinal   int         `json:"rankCalculatedFinal"`
	RankFinal             int         `json:"rankFinal"`
	Record                interface{} `json:"record"`
	TradeBlock            interface{} `json:"tradeBlock"`
	TransactionCounter    interface{} `json:"transactionCounter"`
	WaiverRank            int         `json:"waiverRank"`

	Member Member
}
