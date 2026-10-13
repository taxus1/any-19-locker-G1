package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"somepro/internal/common"
	"somepro/internal/service"
)

// LockerHandler 快递柜处理器
type LockerHandler struct {
	svc *service.LockerService
}

func NewLocker(svc *service.LockerService) *LockerHandler {
	return &LockerHandler{svc: svc}
}

// RegisterCell 登记格口
// POST /api/locker/cell
func (h *LockerHandler) RegisterCell(c *gin.Context) {
	cellNo := c.PostForm("cellNo")
	if cellNo == "" {
		cellNo = c.Query("cellNo")
	}
	size := c.PostForm("size")
	if size == "" {
		size = c.Query("size")
	}

	if cellNo == "" {
		c.JSON(http.StatusOK, common.Fail("cellNo 不能为空"))
		return
	}
	if size == "" {
		c.JSON(http.StatusOK, common.Fail("size 不能为空"))
		return
	}

	cell, err := h.svc.RegisterCell(c.Request.Context(), cellNo, size)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}

	c.JSON(http.StatusOK, common.OK(cell))
}

// DisableCell 停用格口
// POST /api/locker/cell/disable
func (h *LockerHandler) DisableCell(c *gin.Context) {
	cellNo := c.PostForm("cellNo")
	if cellNo == "" {
		cellNo = c.Query("cellNo")
	}

	if cellNo == "" {
		c.JSON(http.StatusOK, common.Fail("cellNo 不能为空"))
		return
	}

	if err := h.svc.DisableCell(c.Request.Context(), cellNo); err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}

	c.JSON(http.StatusOK, common.OK(map[string]interface{}{
		"cellNo": cellNo,
		"msg":    "格口已停用",
	}))
}

// StoreParcel 存件
// POST /api/locker/parcel
func (h *LockerHandler) StoreParcel(c *gin.Context) {
	parcelNo := c.PostForm("parcelNo")
	if parcelNo == "" {
		parcelNo = c.Query("parcelNo")
	}
	cellNo := c.PostForm("cellNo")
	if cellNo == "" {
		cellNo = c.Query("cellNo")
	}
	phone := c.PostForm("phone")
	if phone == "" {
		phone = c.Query("phone")
	}
	size := c.PostForm("size")
	if size == "" {
		size = c.Query("size")
	}

	if parcelNo == "" {
		c.JSON(http.StatusOK, common.Fail("parcelNo 不能为空"))
		return
	}
	if cellNo == "" {
		c.JSON(http.StatusOK, common.Fail("cellNo 不能为空"))
		return
	}
	if phone == "" {
		c.JSON(http.StatusOK, common.Fail("phone 不能为空"))
		return
	}
	if size == "" {
		c.JSON(http.StatusOK, common.Fail("size 不能为空"))
		return
	}

	parcel, err := h.svc.StoreParcel(c.Request.Context(), parcelNo, cellNo, phone, size)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}

	c.JSON(http.StatusOK, common.OK(parcel))
}

// PickupParcel 取件
// POST /api/locker/parcel/pickup
func (h *LockerHandler) PickupParcel(c *gin.Context) {
	parcelNo := c.PostForm("parcelNo")
	if parcelNo == "" {
		parcelNo = c.Query("parcelNo")
	}

	if parcelNo == "" {
		c.JSON(http.StatusOK, common.Fail("parcelNo 不能为空"))
		return
	}

	if err := h.svc.PickupParcel(c.Request.Context(), parcelNo); err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}

	c.JSON(http.StatusOK, common.OK(map[string]interface{}{
		"parcelNo": parcelNo,
		"msg":      "取件成功",
	}))
}

// ListParcelsByPhone 查询在柜包裹
// GET /api/locker/parcel/list
func (h *LockerHandler) ListParcelsByPhone(c *gin.Context) {
	phone := c.Query("phone")
	pageNumStr := c.DefaultQuery("pageNum", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "20")

	pageNum, _ := strconv.Atoi(pageNumStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	data, err := h.svc.ListParcelsByPhone(c.Request.Context(), phone, pageNum, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}

	c.JSON(http.StatusOK, common.OK(data))
}