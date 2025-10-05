package handlers

import "time"

type InitiateTransferRequest struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type InitiateTransferResponse struct {
	MessageToSign string `json:"messageToSign"`
	TransactionID uint   `json:"transactionId"`
}

type PendingTransaction struct {
	From      string
	To        string
	AmountETH string
	Message   string
	CreatedAt time.Time
}

type SkipRoute struct {
	AmountOut string `json:"amount_out"`
}

type SkipAPIResponse struct {
	Route SkipRoute `json:"route"`
}

type ExecuteTransferRequest struct {
	TransactionID string `json:"transactionId"`
	Signature     string `json:"signature"`
}
