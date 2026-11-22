package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	apiReq "github.com/lejianwen/rustdesk-api/v2/http/request/api"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
)

// My is the self-service API used by the user portal.
// It only ever exposes data owned by the logged-in user.
type My struct {
}

// Peers lists the devices bound to the current user.
// @Tags My
// @Summary my devices
// @Success 200 {object} response.DataResponse{data=model.PeerList}
// @Router /my/peers [get]
// @Security BearerAuth
func (m *My) Peers(c *gin.Context) {
	u := service.AllService.UserService.CurUser(c)
	if u == nil || u.Id == 0 {
		response.Error(c, "unauthorized")
		return
	}
	q := &apiReq.PeerListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		response.Error(c, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	res := service.AllService.PeerService.ListByUserIds([]uint{u.Id}, q.Page, q.PageSize)
	response.Success(c, res)
}

// PeerUpdate renames one of the current user's devices.
// @Tags My
// @Summary rename my device
// @Router /my/peer/update [post]
// @Security BearerAuth
func (m *My) PeerUpdate(c *gin.Context) {
	u := service.AllService.UserService.CurUser(c)
	f := &apiReq.MyPeerUpdateForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Error(c, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	if f.RowId == 0 {
		response.Error(c, response.TranslateMsg(c, "ParamsError"))
		return
	}
	peer := service.AllService.PeerService.InfoByRowId(f.RowId)
	if peer.RowId == 0 || peer.UserId != u.Id {
		response.Fail(c, 101, response.TranslateMsg(c, "NoAccess"))
		return
	}
	if err := global.DB.Model(&model.Peer{}).Where("row_id = ? and user_id = ?", f.RowId, u.Id).
		Update("alias", f.Alias).Error; err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}

// PeerUnbind removes the device from the current user without touching hbbs data.
// @Tags My
// @Summary unbind my device
// @Router /my/peer/unbind [post]
// @Security BearerAuth
func (m *My) PeerUnbind(c *gin.Context) {
	u := service.AllService.UserService.CurUser(c)
	f := &apiReq.MyPeerUpdateForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Error(c, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	peer := service.AllService.PeerService.InfoByRowId(f.RowId)
	if peer.RowId == 0 || peer.UserId != u.Id {
		response.Fail(c, 101, response.TranslateMsg(c, "NoAccess"))
		return
	}
	if err := global.DB.Model(&model.Peer{}).Where("row_id = ? and user_id = ?", f.RowId, u.Id).
		Updates(map[string]interface{}{"user_id": 0, "alias": ""}).Error; err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}

// AddressBook lists the address book entries stored for the current user.
// @Tags My
// @Summary my address book
// @Router /my/address_book [get]
// @Security BearerAuth
func (m *My) AddressBook(c *gin.Context) {
	u := service.AllService.UserService.CurUser(c)
	if u == nil || u.Id == 0 {
		response.Error(c, "unauthorized")
		return
	}
	page, pageSize := 1, 50
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			pageSize = n
		}
	}
	res := &model.AddressBookList{}
	tx := global.DB.Model(&model.AddressBook{}).Where("user_id = ?", u.Id)
	tx.Count(&res.Total)
	tx.Order("row_id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&res.AddressBooks)
	res.Page = int64(page)
	res.PageSize = int64(pageSize)
	response.Success(c, res)
}
