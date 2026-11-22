package model

// AuditLogin records every token verification performed by hbbs through
// /api/server/verify-token, so a rejected or accepted connection attempt is
// traceable even when the peer never completes a rendezvous handshake.
type AuditLogin struct {
	IdModel
	Guid     string `json:"guid" gorm:"default:'';not null;index"`
	Uuid     string `json:"uuid" gorm:"default:'';not null;index"`
	Ip       string `json:"ip" gorm:"default:'';not null;"`
	UserId   uint   `json:"user_id" gorm:"default:0;not null;index"`
	Username string `json:"username" gorm:"default:'';not null;"`
	Result   string `json:"result" gorm:"default:'';not null;index"`
	Reason   string `json:"reason" gorm:"default:'';not null;"`
	TimeModel
}

type AuditLoginList struct {
	AuditLogins []*AuditLogin `json:"list"`
	Pagination
}
