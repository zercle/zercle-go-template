package contract

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
