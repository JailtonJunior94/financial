package dtos

type UpdateUserInput struct {
	Name     *string `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"` //nolint:gosec
}
