package goldenmod

// CreateGoldenmodRequest is the payload for creating a Goldenmod.
type CreateGoldenmodRequest struct {
	Name string `json:"name" binding:"required"`
	CompanyID uint `json:"company_id" binding:"required"`
}

// UpdateGoldenmodRequest is the payload for updating a Goldenmod.
type UpdateGoldenmodRequest struct {
	Name string `json:"name" binding:"required"`
	CompanyID uint `json:"company_id" binding:"required"`
}

// GoldenmodResponse is the JSON representation returned by the API.
type GoldenmodResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CompanyID uint   `json:"company_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
