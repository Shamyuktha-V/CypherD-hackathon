package handlers

type CreateWalletRequest struct {
	Address string  `json:"address"`
	Email   *string `json:"email,omitempty"`
}
