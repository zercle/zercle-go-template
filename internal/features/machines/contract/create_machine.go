package contract

// CreateMachineRequest is the payload for registering a new machine. Only
// structural constraints live here; accepted coin denominations are a domain
// fact enforced by the usecase layer, so they are not a wire constraint.
type CreateMachineRequest struct {
	Label        string  `json:"label" validate:"required"`
	InitialCoins []int32 `json:"initial_coins" validate:"omitempty,dive,gt=0"`
}

// MachineResponse is the JSON representation of a machine and its coin bank.
// CoinBank keys are denomination values in cents, rendered as JSON object keys.
type MachineResponse struct {
	ID        string          `json:"id"`
	Label     string          `json:"label"`
	CoinBank  map[int32]int32 `json:"coin_bank"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}
