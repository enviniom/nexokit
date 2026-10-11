package refresh_token

import "github.com/enviniom/nexokit/internal/platform/validator"

// RefreshRequest is the DTO for refresh and logout requests.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r *RefreshRequest) Validate() validator.ValidationErrors {
	errs := make(validator.ValidationErrors)
	validator.Field(errs, "refresh_token", r.RefreshToken).Required()
	return errs
}

// TokenPairResponse returns a rotated token pair, never stored token hashes.
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
