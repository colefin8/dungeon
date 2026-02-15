package shared

const (
	RequestTypeLogin byte = iota
	RequestTypeSay
)

const (
	ResponseTypeLogin byte = iota
	ResponseTypeLogout
	ResponseTypeLoggedInUsers
	ResponseTypeSay
)
