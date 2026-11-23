package apiv1

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

// ListMachinesRequest carries pagination parameters for listing machines. It
// is bound from the GET query string (the query tags); the upper page-size
// limit is deployment-configurable and enforced in the usecase layer, so a
// hardcoded max= tag would drift.
type ListMachinesRequest struct {
	Limit  int32 `json:"limit" query:"limit" validate:"min=0"`
	Offset int32 `json:"offset" query:"offset" validate:"min=0"`
}

// ListMachinesResponse wraps a page of machines.
type ListMachinesResponse struct {
	Machines []MachineResponse `json:"machines"`
}

// RestockBankRequest is the payload for adding coins to a machine's bank.
// Coins are validated structurally (at least one, positive); the domain layer
// rejects values that are not accepted denominations because the accepted set
// is a domain fact, not a wire fact.
type RestockBankRequest struct {
	Coins []int32 `json:"coins" validate:"required,min=1,dive,gt=0"`
}
