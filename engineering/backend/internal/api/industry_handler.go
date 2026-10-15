package api

import (
	"marketlens-go-backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type industryDivisionsByParentResponse struct {
	IndustrySectorID  uint                      `json:"industry_sector_id"`
	Count             int                       `json:"count"`
	IndustryDivisions []models.IndustryDivision `json:"industry_divisions"`
}

type industryGroupsByParentResponse struct {
	IndustryDivisionID uint                   `json:"industry_division_id"`
	Count              int                    `json:"count"`
	IndustryGroups     []models.IndustryGroup `json:"industry_groups"`
}

type industryClassesByParentResponse struct {
	IndustryGroupID uint                   `json:"industry_group_id"`
	Count           int                    `json:"count"`
	IndustryClasses []models.IndustryClass `json:"industry_classes"`
}

type industrySubclassesByParentResponse struct {
	IndustryClassID    uint                      `json:"industry_class_id"`
	Count              int                       `json:"count"`
	IndustrySubclasses []models.IndustrySubclass `json:"industry_subclasses"`
}

// Industry levels by parent ids
//This function returns industry divisions under an industry sector, currently-active by default,
//or active during a given date range if from-date and to-date query params are provided
func (h *Handler) GetIndustryDivisionsByIndustrySectorHandler(c *gin.Context) {
	id, ok := parseIDParamMsg(c, "Invalid industry sector id parameter")
	if !ok {
		return
	}
	fromDate, toDate, ok := parseOptionalDateRangeQuery(c)
	if !ok {
		return
	}

	items, err := h.IndustryDivisions.GetByParent(c.Request.Context(), id, fromDate, toDate)
	if err != nil {
		respondRepoError(c, "Failed to retrieve industry divisions", err)
		return
	}

	c.JSON(http.StatusOK, industryDivisionsByParentResponse{IndustrySectorID: id, Count: len(items), IndustryDivisions: items})
}

//This function returns industry groups under an industry division, currently-active by default,
//or active during a given date range if from-date and to-date query params are provided
func (h *Handler) GetIndustryGroupsByIndustryDivisionHandler(c *gin.Context) {
	id, ok := parseIDParamMsg(c, "Invalid industry division id parameter")
	if !ok {
		return
	}
	fromDate, toDate, ok := parseOptionalDateRangeQuery(c)
	if !ok {
		return
	}

	items, err := h.IndustryGroups.GetByParent(c.Request.Context(), id, fromDate, toDate)
	if err != nil {
		respondRepoError(c, "Failed to retrieve industry groups", err)
		return
	}

	c.JSON(http.StatusOK, industryGroupsByParentResponse{IndustryDivisionID: id, Count: len(items), IndustryGroups: items})
}

//This function returns industry classes under an industry group, currently-active by default,
//or active during a given date range if from-date and to-date query params are provided
func (h *Handler) GetIndustryClassesByIndustryGroupHandler(c *gin.Context) {
	id, ok := parseIDParamMsg(c, "Invalid industry group id parameter")
	if !ok {
		return
	}
	fromDate, toDate, ok := parseOptionalDateRangeQuery(c)
	if !ok {
		return
	}

	items, err := h.IndustryClasses.GetByParent(c.Request.Context(), id, fromDate, toDate)
	if err != nil {
		respondRepoError(c, "Failed to retrieve industry classes", err)
		return
	}

	c.JSON(http.StatusOK, industryClassesByParentResponse{IndustryGroupID: id, Count: len(items), IndustryClasses: items})
}

//This function returns industry subclasses under an industry class, currently-active by default,
//or active during a given date range if from-date and to-date query params are provided
func (h *Handler) GetIndustrySubclassesByIndustryClassHandler(c *gin.Context) {
	id, ok := parseIDParamMsg(c, "Invalid industry class id parameter")
	if !ok {
		return
	}
	fromDate, toDate, ok := parseOptionalDateRangeQuery(c)
	if !ok {
		return
	}

	items, err := h.IndustrySubclasses.GetByParent(c.Request.Context(), id, fromDate, toDate)
	if err != nil {
		respondRepoError(c, "Failed to retrieve industry subclasses", err)
		return
	}

	c.JSON(http.StatusOK, industrySubclassesByParentResponse{IndustryClassID: id, Count: len(items), IndustrySubclasses: items})
}

// Industry Sectors CRUDS
// Industry Sector CRUD
type createIndustrySectorInput struct {
	Name string `json:"name" binding:"required,max=255"`
	Code string `json:"code" binding:"required,max=50"`
}

type updateIndustrySectorInput struct {
	Name *string `json:"name" binding:"omitempty,max=255"`
	Code *string `json:"code" binding:"omitempty,max=50"`
}

type industrySectorListResponse struct {
	Count           int                     `json:"count"`
	IndustrySectors []models.IndustrySector `json:"industry_sectors"`
}

func (h *Handler) CreateIndustrySectorHandler(c *gin.Context) {
	var in createIndustrySectorInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.IndustrySector{Name: in.Name, Code: in.Code}
	if err := h.IndustrySectors.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create industry sector", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetAllIndustrySectorsHandler(c *gin.Context) {
	fromDate, toDate, ok := parseOptionalDateRangeQuery(c)
	if !ok {
		return
	}

	items, err := h.IndustrySectors.GetAll(c.Request.Context(), 0, 0, fromDate, toDate)
	if err != nil {
		respondRepoError(c, "Failed to retrieve industry sectors", err)
		return
	}

	c.JSON(http.StatusOK, industrySectorListResponse{Count: len(items), IndustrySectors: items})
}

func (h *Handler) GetIndustrySectorByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.IndustrySectors.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch industry sector", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateIndustrySectorHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateIndustrySectorInput
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
	item, err := h.IndustrySectors.Update(c.Request.Context(), id, updates)
	if err != nil {
		respondRepoError(c, "Failed to update industry sector", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteIndustrySectorHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.IndustrySectors.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete industry sector", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Industry sector deleted successfully"})
}

// Industry Division CRUD
type createIndustryDivisionInput struct {
	IndustrySectorID uint   `json:"industry_sector_id" binding:"required"`
	Name             string `json:"name" binding:"required,max=255"`
	Code             string `json:"code" binding:"required,max=50"`
}

type updateIndustryDivisionInput struct {
	IndustrySectorID *uint   `json:"industry_sector_id"`
	Name             *string `json:"name" binding:"omitempty,max=255"`
	Code             *string `json:"code" binding:"omitempty,max=50"`
}

type industryDivisionListResponse struct {
	Count             int                       `json:"count"`
	IndustryDivisions []models.IndustryDivision `json:"industry_divisions"`
}

func (h *Handler) CreateIndustryDivisionHandler(c *gin.Context) {
	var in createIndustryDivisionInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.IndustryDivision{IndustrySectorID: in.IndustrySectorID, Name: in.Name, Code: in.Code}
	if err := h.IndustryDivisions.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create industry division", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetAllIndustryDivisionsHandler(c *gin.Context) {
	items, err := h.IndustryDivisions.GetAll(c.Request.Context(), 0, 0, nil, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch industry divisions", err)
		return
	}
	c.JSON(http.StatusOK, industryDivisionListResponse{Count: len(items), IndustryDivisions: items})
}

func (h *Handler) GetIndustryDivisionByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.IndustryDivisions.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch industry division", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateIndustryDivisionHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateIndustryDivisionInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	updates := map[string]interface{}{}
	if in.IndustrySectorID != nil {
		updates["industry_sector_id"] = *in.IndustrySectorID
	}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Code != nil {
		updates["code"] = *in.Code
	}
	item, err := h.IndustryDivisions.Update(c.Request.Context(), id, updates)
	if err != nil {
		respondRepoError(c, "Failed to update industry division", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteIndustryDivisionHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.IndustryDivisions.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete industry division", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Industry division deleted successfully"})
}

// Industry Group CRUD
type createIndustryGroupInput struct {
	IndustryDivisionID uint   `json:"industry_division_id" binding:"required"`
	Name               string `json:"name" binding:"required,max=255"`
	Code               string `json:"code" binding:"required,max=50"`
}

type updateIndustryGroupInput struct {
	IndustryDivisionID *uint   `json:"industry_division_id"`
	Name               *string `json:"name" binding:"omitempty,max=255"`
	Code               *string `json:"code" binding:"omitempty,max=50"`
}

type industryGroupListResponse struct {
	Count          int                    `json:"count"`
	IndustryGroups []models.IndustryGroup `json:"industry_groups"`
}

func (h *Handler) CreateIndustryGroupHandler(c *gin.Context) {
	var in createIndustryGroupInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.IndustryGroup{IndustryDivisionID: in.IndustryDivisionID, Name: in.Name, Code: in.Code}
	if err := h.IndustryGroups.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create industry group", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetAllIndustryGroupsHandler(c *gin.Context) {
	items, err := h.IndustryGroups.GetAll(c.Request.Context(), 0, 0, nil, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch industry groups", err)
		return
	}
	c.JSON(http.StatusOK, industryGroupListResponse{Count: len(items), IndustryGroups: items})
}

func (h *Handler) GetIndustryGroupByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.IndustryGroups.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch industry group", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateIndustryGroupHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateIndustryGroupInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	updates := map[string]interface{}{}
	if in.IndustryDivisionID != nil {
		updates["industry_division_id"] = *in.IndustryDivisionID
	}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Code != nil {
		updates["code"] = *in.Code
	}
	item, err := h.IndustryGroups.Update(c.Request.Context(), id, updates)
	if err != nil {
		respondRepoError(c, "Failed to update industry group", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteIndustryGroupHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.IndustryGroups.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete industry group", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Industry group deleted successfully"})
}

// Industry Class CRUD
type createIndustryClassInput struct {
	IndustryGroupID uint   `json:"industry_group_id" binding:"required"`
	Name            string `json:"name" binding:"required,max=255"`
	Code            string `json:"code" binding:"required,max=50"`
}

type updateIndustryClassInput struct {
	IndustryGroupID *uint   `json:"industry_group_id"`
	Name            *string `json:"name" binding:"omitempty,max=255"`
	Code            *string `json:"code" binding:"omitempty,max=50"`
}

type industryClassListResponse struct {
	Count           int                    `json:"count"`
	IndustryClasses []models.IndustryClass `json:"industry_classes"`
}

func (h *Handler) CreateIndustryClassHandler(c *gin.Context) {
	var in createIndustryClassInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.IndustryClass{IndustryGroupID: in.IndustryGroupID, Name: in.Name, Code: in.Code}
	if err := h.IndustryClasses.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create industry class", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetAllIndustryClassesHandler(c *gin.Context) {
	items, err := h.IndustryClasses.GetAll(c.Request.Context(), 0, 0, nil, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch industry classes", err)
		return
	}
	c.JSON(http.StatusOK, industryClassListResponse{Count: len(items), IndustryClasses: items})
}

func (h *Handler) GetIndustryClassByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.IndustryClasses.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch industry class", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateIndustryClassHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateIndustryClassInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	updates := map[string]interface{}{}
	if in.IndustryGroupID != nil {
		updates["industry_group_id"] = *in.IndustryGroupID
	}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Code != nil {
		updates["code"] = *in.Code
	}
	item, err := h.IndustryClasses.Update(c.Request.Context(), id, updates)
	if err != nil {
		respondRepoError(c, "Failed to update industry class", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteIndustryClassHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.IndustryClasses.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete industry class", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Industry class deleted successfully"})
}

// Industry Sub Class CRUD
type createIndustrySubclassInput struct {
	IndustryClassID uint   `json:"industry_class_id" binding:"required"`
	Name            string `json:"name" binding:"required,max=255"`
	Code            string `json:"code" binding:"required,max=50"`
}

type updateIndustrySubclassInput struct {
	IndustryClassID *uint   `json:"industry_class_id"`
	Name            *string `json:"name" binding:"omitempty,max=255"`
	Code            *string `json:"code" binding:"omitempty,max=50"`
}

type industrySubclassPaginatedResponse struct {
	Count              int                       `json:"count"`
	Limit              int                       `json:"limit"`
	Offset             int                       `json:"offset"`
	IndustrySubclasses []models.IndustrySubclass `json:"industry_subclasses"`
}

func (h *Handler) CreateIndustrySubclassHandler(c *gin.Context) {
	var in createIndustrySubclassInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.IndustrySubclass{IndustryClassID: in.IndustryClassID, Name: in.Name, Code: in.Code}
	if err := h.IndustrySubclasses.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create industry subclass", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetAllIndustrySubclassesHandler(c *gin.Context) {
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

	items, err := h.IndustrySubclasses.GetAll(c.Request.Context(), limit, offset, nil, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch industry subclasses", err)
		return
	}

	c.JSON(http.StatusOK, industrySubclassPaginatedResponse{Count: len(items), Limit: limit, Offset: offset, IndustrySubclasses: items})
}

func (h *Handler) GetIndustrySubclassByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.IndustrySubclasses.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch industry subclass", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateIndustrySubclassHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateIndustrySubclassInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	updates := map[string]interface{}{}
	if in.IndustryClassID != nil {
		updates["industry_class_id"] = *in.IndustryClassID
	}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Code != nil {
		updates["code"] = *in.Code
	}
	item, err := h.IndustrySubclasses.Update(c.Request.Context(), id, updates)
	if err != nil {
		respondRepoError(c, "Failed to update industry subclass", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteIndustrySubclassHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.IndustrySubclasses.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete industry subclass", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Industry subclass deleted successfully"})
}
