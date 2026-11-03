package payments

type webhookEvent struct {
	Event string `json:"event"`
	Data  struct {
		ID        int64  `json:"id"`
		Reference string `json:"reference"`
		Amount    int64  `json:"amount"`
		Currency  string `json:"currency"`
		Channel   string `json:"channel"`
	} `json:"data"`
}

type ChargeSuccess struct {
	ProviderID int64
	Reference  string
	Amount     int64
	Currency   string
	Channel    string
}
