package response

// UserInfo defines the basic user info returned after successful login
type UserInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

// LoginResponse defines the response body after successful login
type LoginResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}
