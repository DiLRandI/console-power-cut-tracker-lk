package model

type Response struct {
	ID              string `json:"id,omitempty"`
	StartTime       string `json:"startTime,omitempty"`
	EndTime         string `json:"endTime,omitempty"`
	LoadShedGroupID string `json:"loadShedGroupId,omitempty"`
	NoOfFeeders     string `json:"noOfFeeders,omitempty"`
	StatusID        string `json:"statusId,omitempty"`
	TimeStamp       string `json:"timeStamp,omitempty"`
	UserID          string `json:"userId,omitempty"`
}
