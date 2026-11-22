package api

type MyPeerUpdateForm struct {
	RowId uint   `json:"row_id" validate:"required"`
	Alias string `json:"alias"`
}