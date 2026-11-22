package api

// VerifyTokenForm is sent by hbbs for every PunchHoleRequest when
// MUST_LOGIN=api. Guid/Uuid/Ip are audit context only.
type VerifyTokenForm struct {
	Token string `json:"token"`
	Guid  string `json:"guid"`
	Uuid  string `json:"uuid"`
	Ip    string `json:"ip"`
}
