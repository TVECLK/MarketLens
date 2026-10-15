package api

import (
	"marketlens-go-backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type subMajorGroupsByParentResponse struct {
	MajorGroupID   uint                   `json:"major_group_id"`
	Count          int                    `json:"count"`
	SubMajorGroups []models.SubMajorGroup `json:"sub_major_groups"`
}

type minorGroupsByParentResponse struct {
	SubMajorGroupID uint                `json:"sub_major_group_id"`
	Count           int                 `json:"count"`
	MinorGroups     []models.MinorGroup `json:"minor_groups"`
}

type unitGroupsByParentResponse struct {
	MinorGroupID uint               `json:"minor_group_id"`
	Count        int                `json:"count"`
	UnitGroups   []models.UnitGroup `json:"unit_groups"`
}

type occupationGroupsByParentResponse struct {
	UnitGroupID      uint                     `json:"unit_group_id"`
	Count            int                      `json:"count"`
	OccupationGroups []models.OccupationGroup `json:"occupation_groups"`
}

// Methods needed for multi-level filtering in the crawler
// Occupation levels by parent ids
func (h *Handler) GetSubMajorGroupsByMajorGroupHandler(c *gin.Context) {
	id, ok := parseIDParamMsg(c, "Invalid major group id parameter")
	if !ok {
		return
	}
	fromDate, toDate, ok := parseOptionalDateRangeQuery(c)
	if !ok {
		return
	}

	items, err := h.SubMajorGroups.GetByParent(c.Request.Context(), id, fromDate, toDate)
	if err != nil {
		respondRepoError(c, "Failed to retrieve sub major groups", err)
		return
	}

	c.JSON(http.StatusOK, subMajorGroupsByParentResponse{MajorGroupID: id, Count: len(items), SubMajorGroups: items})
}

func (h *Handler) GetMinorGroupsBySubMajorGroupHandler(c *gin.Context) {
	id, ok := parseIDParamMsg(c, "Invalid sub major group id parameter")
	if !ok {
		return
	}
	fromDate, toDate, ok := parseOptionalDateRangeQuery(c)
	if !ok {
		return
	}

	items, err := h.MinorGroups.GetByParent(c.Request.Context(), id, fromDate, toDate)
	if err != nil {
		respondRepoError(c, "Failed to retrieve minor groups", err)
		return
	}

	c.JSON(http.StatusOK, minorGroupsByParentResponse{SubMajorGroupID: id, Count: len(items), MinorGroups: items})
}

func (h *Handler) GetUnitGroupsByMinorGroupHandler(c *gin.Context) {
	id, ok := parseIDParamMsg(c, "Invalid minor group id parameter")
	if !ok {
		return
	}
	fromDate, toDate, ok := parseOptionalDateRangeQuery(c)
	if !ok {
		return
	}

	items, err := h.UnitGroups.GetByParent(c.Request.Context(), id, fromDate, toDate)
	if err != nil {
		respondRepoError(c, "Failed to retrieve unit groups", err)
		return
	}

	c.JSON(http.StatusOK, unitGroupsByParentResponse{MinorGroupID: id, Count: len(items), UnitGroups: items})
}

func (h *Handler) GetOccupationGroupsByUnitGroupHandler(c *gin.Context) {
	id, ok := parseIDParamMsg(c, "Invalid unit group id parameter")
	if !ok {
		return
	}
	fromDate, toDate, ok := parseOptionalDateRangeQuery(c)
	if !ok {
		return
	}

	items, err := h.OccupationGroups.GetByParent(c.Request.Context(), id, fromDate, toDate)
	if err != nil {
		respondRepoError(c, "Failed to retrieve occupation groups", err)
		return
	}

	c.JSON(http.StatusOK, occupationGroupsByParentResponse{UnitGroupID: id, Count: len(items), OccupationGroups: items})
}

// Occupations CRUDS
// Major Group CRUD
type createMajorGroupInput struct {
	Name string `json:"name" binding:"required,max=255"`
	Code string `json:"code" binding:"required,max=50"`
}

// updateMajorGroupInput fields are pointers so a PUT can change just one
// of them without having to resend the other.
type updateMajorGroupInput struct {
	Name *string `json:"name" binding:"omitempty,max=255"`
	Code *string `json:"code" binding:"omitempty,max=50"`
}

type majorGroupListResponse struct {
	Count       int                `json:"count"`
	MajorGroups []models.MajorGroup `json:"major_groups"`
}

func (h *Handler) CreateMajorGroupHandler(c *gin.Context) {
	var in createMajorGroupInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.MajorGroup{Name: in.Name, Code: in.Code}
	if err := h.MajorGroups.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create major group", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetAllMajorGroupsHandler(c *gin.Context) {
	fromDate, toDate, ok := parseOptionalDateRangeQuery(c)
	if !ok {
		return
	}

	items, err := h.MajorGroups.GetAll(c.Request.Context(), 0, 0, fromDate, toDate)
	if err != nil {
		respondRepoError(c, "Failed to retrieve major groups", err)
		return
	}

	c.JSON(http.StatusOK, majorGroupListResponse{Count: len(items), MajorGroups: items})
}

func (h *Handler) GetMajorGroupByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.MajorGroups.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch major group", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateMajorGroupHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateMajorGroupInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	updates := map[string]interface{}{}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Code != nil {
		updates["code"] = *in.Code
	}
	item, err := h.MajorGroups.Update(c.Request.Context(), id, updates)
	if err != nil {
		respondRepoError(c, "Failed to update major group", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteMajorGroupHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.MajorGroups.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete major group", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Major group deleted successfully"})
}

// Sub Major Group CRUD
type createSubMajorGroupInput struct {
	MajorGroupID uint   `json:"major_group_id" binding:"required"`
	Name         string `json:"name" binding:"required,max=255"`
	Code         string `json:"code" binding:"required,max=50"`
}

type updateSubMajorGroupInput struct {
	MajorGroupID *uint   `json:"major_group_id"`
	Name         *string `json:"name" binding:"omitempty,max=255"`
	Code         *string `json:"code" binding:"omitempty,max=50"`
}

type subMajorGroupListResponse struct {
	Count          int                    `json:"count"`
	SubMajorGroups []models.SubMajorGroup `json:"sub_major_groups"`
}

func (h *Handler) CreateSubMajorGroupHandler(c *gin.Context) {
	var in createSubMajorGroupInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.SubMajorGroup{MajorGroupID: in.MajorGroupID, Name: in.Name, Code: in.Code}
	if err := h.SubMajorGroups.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create sub major group", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetAllSubMajorGroupsHandler(c *gin.Context) {
	items, err := h.SubMajorGroups.GetAll(c.Request.Context(), 0, 0, nil, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch sub major groups", err)
		return
	}
	c.JSON(http.StatusOK, subMajorGroupListResponse{Count: len(items), SubMajorGroups: items})
}

func (h *Handler) GetSubMajorGroupByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.SubMajorGroups.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch sub major group", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateSubMajorGroupHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateSubMajorGroupInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	updates := map[string]interface{}{}
	if in.MajorGroupID != nil {
		updates["major_group_id"] = *in.MajorGroupID
	}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Code != nil {
		updates["code"] = *in.Code
	}
	item, err := h.SubMajorGroups.Update(c.Request.Context(), id, updates)
	if err != nil {
		respondRepoError(c, "Failed to update sub major group", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteSubMajorGroupHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.SubMajorGroups.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete sub major group", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Sub major group deleted successfully"})
}

// Minor Group CRUD
type createMinorGroupInput struct {
	SubMajorGroupID uint   `json:"sub_major_group_id" binding:"required"`
	Name            string `json:"name" binding:"required,max=255"`
	Code            string `json:"code" binding:"required,max=50"`
}

type updateMinorGroupInput struct {
	SubMajorGroupID *uint   `json:"sub_major_group_id"`
	Name            *string `json:"name" binding:"omitempty,max=255"`
	Code            *string `json:"code" binding:"omitempty,max=50"`
}

type minorGroupListResponse struct {
	Count       int                 `json:"count"`
	MinorGroups []models.MinorGroup `json:"minor_groups"`
}

func (h *Handler) CreateMinorGroupHandler(c *gin.Context) {
	var in createMinorGroupInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.MinorGroup{SubMajorGroupID: in.SubMajorGroupID, Name: in.Name, Code: in.Code}
	if err := h.MinorGroups.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create minor group", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetAllMinorGroupsHandler(c *gin.Context) {
	items, err := h.MinorGroups.GetAll(c.Request.Context(), 0, 0, nil, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch minor groups", err)
		return
	}
	c.JSON(http.StatusOK, minorGroupListResponse{Count: len(items), MinorGroups: items})
}

func (h *Handler) GetMinorGroupByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.MinorGroups.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch minor group", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateMinorGroupHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateMinorGroupInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	updates := map[string]interface{}{}
	if in.SubMajorGroupID != nil {
		updates["sub_major_group_id"] = *in.SubMajorGroupID
	}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Code != nil {
		updates["code"] = *in.Code
	}
	item, err := h.MinorGroups.Update(c.Request.Context(), id, updates)
	if err != nil {
		respondRepoError(c, "Failed to update minor group", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteMinorGroupHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.MinorGroups.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete minor group", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Minor group deleted successfully"})
}

// Unit Group CRUD
type createUnitGroupInput struct {
	MinorGroupID uint   `json:"minor_group_id" binding:"required"`
	Name         string `json:"name" binding:"required,max=255"`
	Code         string `json:"code" binding:"required,max=50"`
}

type updateUnitGroupInput struct {
	MinorGroupID *uint   `json:"minor_group_id"`
	Name         *string `json:"name" binding:"omitempty,max=255"`
	Code         *string `json:"code" binding:"omitempty,max=50"`
}

type unitGroupListResponse struct {
	Count      int                `json:"count"`
	UnitGroups []models.UnitGroup `json:"unit_groups"`
}

func (h *Handler) CreateUnitGroupHandler(c *gin.Context) {
	var in createUnitGroupInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.UnitGroup{MinorGroupID: in.MinorGroupID, Name: in.Name, Code: in.Code}
	if err := h.UnitGroups.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create unit group", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetAllUnitGroupsHandler(c *gin.Context) {
	items, err := h.UnitGroups.GetAll(c.Request.Context(), 0, 0, nil, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch unit groups", err)
		return
	}
	c.JSON(http.StatusOK, unitGroupListResponse{Count: len(items), UnitGroups: items})
}

func (h *Handler) GetUnitGroupByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.UnitGroups.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch unit group", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateUnitGroupHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateUnitGroupInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	updates := map[string]interface{}{}
	if in.MinorGroupID != nil {
		updates["minor_group_id"] = *in.MinorGroupID
	}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Code != nil {
		updates["code"] = *in.Code
	}
	item, err := h.UnitGroups.Update(c.Request.Context(), id, updates)
	if err != nil {
		respondRepoError(c, "Failed to update unit group", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteUnitGroupHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.UnitGroups.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete unit group", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Unit group deleted successfully"})
}

// Occupation Group CRUD
type createOccupationGroupInput struct {
	UnitGroupID uint   `json:"unit_group_id" binding:"required"`
	Name        string `json:"name" binding:"required,max=255"`
	Code        string `json:"code" binding:"required,max=50"`
}

type updateOccupationGroupInput struct {
	UnitGroupID *uint   `json:"unit_group_id"`
	Name        *string `json:"name" binding:"omitempty,max=255"`
	Code        *string `json:"code" binding:"omitempty,max=50"`
}

type occupationGroupPaginatedResponse struct {
	Count            int                      `json:"count"`
	Limit            int                      `json:"limit"`
	Offset           int                      `json:"offset"`
	OccupationGroups []models.OccupationGroup `json:"occupation_groups"`
}

func (h *Handler) CreateOccupationGroupHandler(c *gin.Context) {
	var in createOccupationGroupInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.OccupationGroup{UnitGroupID: in.UnitGroupID, Name: in.Name, Code: in.Code}
	if err := h.OccupationGroups.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create occupation group", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetAllOccupationGroupsHandler(c *gin.Context) {
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			if v > 100 {
				v = 100
			}
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	items, err := h.OccupationGroups.GetAll(c.Request.Context(), limit, offset, nil, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch occupation groups", err)
		return
	}

	c.JSON(http.StatusOK, occupationGroupPaginatedResponse{Count: len(items), Limit: limit, Offset: offset, OccupationGroups: items})
}

func (h *Handler) GetOccupationGroupByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.OccupationGroups.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch occupation group", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateOccupationGroupHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateOccupationGroupInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	updates := map[string]interface{}{}
	if in.UnitGroupID != nil {
		updates["unit_group_id"] = *in.UnitGroupID
	}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Code != nil {
		updates["code"] = *in.Code
	}
	item, err := h.OccupationGroups.Update(c.Request.Context(), id, updates)
	if err != nil {
		respondRepoError(c, "Failed to update occupation group", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteOccupationGroupHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.OccupationGroups.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete occupation group", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Occupation group deleted successfully"})
}
