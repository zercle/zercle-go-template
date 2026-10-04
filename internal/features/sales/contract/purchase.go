package contract

// PurchaseRequest is the payload for buying a product: the machine, the
// product, and the coins inserted. Coins are validated structurally (at least
// one, positive); the domain layer rejects values that are not accepted
// denominations because the accepted set is a domain fact, not a wire fact.
type PurchaseRequest struct {
	MachineID string  `json:"machine_id" validate:"required,uuid"`
	ProductID string  `json:"product_id" validate:"required,uuid"`
	Coins     []int32 `json:"coins" validate:"required,min=1,dive,gt=0"`
}

// PurchaseResponse is the JSON representation of a completed purchase.
type PurchaseResponse struct {
	ID                 string  `json:"id"`
	MachineID          string  `json:"machine_id"`
	ProductID          string  `json:"product_id"`
	PriceCents         int32   `json:"price_cents"`
	TotalInsertedCents int32   `json:"total_inserted_cents"`
	ChangeCents        int32   `json:"change_cents"`
	ChangeCoins        []int32 `json:"change_coins"`
	PurchasedAt        string  `json:"purchased_at"`
}
