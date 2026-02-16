package shared

const (
	RequestTypeLogin byte = iota
	RequestTypeSay
	RequestTypeWho
	RequestTypeLook
)

const (
	ResponseTypeLogin byte = iota
	ResponseTypeLogout
	ResponseTypeLoggedInUsers
	ResponseTypeSay
	ResponseTypeLook
)
