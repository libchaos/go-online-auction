package authn

const (
	RoleAdmin  = "admin"
	RoleSeller = "seller"
	RoleBidder = "bidder"
)

// Claims represents the authenticated user data extracted from an access token
type Claims struct {
	UserID uint64
	Role   string
	Email  string
}
