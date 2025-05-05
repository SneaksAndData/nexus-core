package models

// CancellationRequest hold information on who and why cancelled the submission
type CancellationRequest struct {
	Initiator string `json:"initiator" binding:"required"`
	Reason    string `json:"reason"  binding:"required"`
	RequestId string `json:"requestId" binding:"required"`
}
