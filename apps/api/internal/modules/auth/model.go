package auth

type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type CompanyContext struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Timezone string `json:"timezone"`
	Currency string `json:"currency"`
	Role     string `json:"role"`
}

type Principal struct {
	User    User            `json:"user"`
	Company *CompanyContext `json:"company,omitempty"`
}
