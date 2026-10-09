package tests_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"marketlens-go-backend/controllers"
	"marketlens-go-backend/models"
	"marketlens-go-backend/repositories"
)

// ---------- Test DB + router setup ----------

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_foreign_keys=on", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get generic sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(
		&models.EducationLevel{}, &models.Formality{}, &models.Gender{},
		&models.EmploymentSector{}, &models.VocationalEducation{}, &models.Experience{},
		&models.MajorGroup{}, &models.SubMajorGroup{}, &models.MinorGroup{}, &models.UnitGroup{}, &models.OccupationGroup{},
		&models.IndustrySector{}, &models.IndustryDivision{}, &models.IndustryGroup{}, &models.IndustryClass{}, &models.IndustrySubclass{},
	); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	t.Cleanup(func() {
		sqlDB.Close()
	})

	return db
}

// setupRouter wires only the education-level routes needed for these tests.
func setupRouter(ctrl *controllers.JobController) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/education-levels", ctrl.CreateEducationLevelHandler)
	r.GET("/education-levels", ctrl.GetAllEducationLevelsHandler)
	r.GET("/education-levels/:id", ctrl.GetEducationLevelByIDHandler)
	r.PUT("/education-levels/:id", ctrl.UpdateEducationLevelHandler)
	r.DELETE("/education-levels/:id", ctrl.DeleteEducationLevelHandler)
	r.POST("/formalities", ctrl.CreateFormalityHandler)
	r.GET("/formalities", ctrl.GetAllFormalitiesHandler)
	r.GET("/formalities/:id", ctrl.GetFormalityByIDHandler)
	r.PUT("/formalities/:id", ctrl.UpdateFormalityHandler)
	r.DELETE("/formalities/:id", ctrl.DeleteFormalityHandler)

	r.POST("/genders", ctrl.CreateGenderHandler)
	r.GET("/genders", ctrl.GetAllGendersHandler)
	r.GET("/genders/:id", ctrl.GetGenderByIDHandler)
	r.PUT("/genders/:id", ctrl.UpdateGenderHandler)
	r.DELETE("/genders/:id", ctrl.DeleteGenderHandler)

	r.POST("/employment-sectors", ctrl.CreateEmploymentSectorHandler)
	r.GET("/employment-sectors", ctrl.GetAllEmploymentSectorsHandler)
	r.GET("/employment-sectors/:id", ctrl.GetEmploymentSectorByIDHandler)
	r.PUT("/employment-sectors/:id", ctrl.UpdateEmploymentSectorHandler)
	r.DELETE("/employment-sectors/:id", ctrl.DeleteEmploymentSectorHandler)

	r.POST("/vocational-educations", ctrl.CreateVocationalEducationHandler)
	r.GET("/vocational-educations", ctrl.GetAllVocationalEducationsHandler)
	r.GET("/vocational-educations/:id", ctrl.GetVocationalEducationByIDHandler)
	r.PUT("/vocational-educations/:id", ctrl.UpdateVocationalEducationHandler)
	r.DELETE("/vocational-educations/:id", ctrl.DeleteVocationalEducationHandler)

	r.POST("/experiences", ctrl.CreateExperienceHandler)
	r.GET("/experiences/:id", ctrl.GetExperienceByIDHandler)
	r.PUT("/experiences/:id", ctrl.UpdateExperienceHandler)
	r.DELETE("/experiences/:id", ctrl.DeleteExperienceHandler)

	r.POST("/major-groups", ctrl.CreateMajorGroupHandler)
	r.GET("/major-groups", ctrl.GetAllMajorGroupsHandler)
	r.GET("/major-groups/:id", ctrl.GetMajorGroupByIDHandler)
	r.PUT("/major-groups/:id", ctrl.UpdateMajorGroupHandler)
	r.DELETE("/major-groups/:id", ctrl.DeleteMajorGroupHandler)

	r.POST("/sub-major-groups", ctrl.CreateSubMajorGroupHandler)
	r.GET("/sub-major-groups", ctrl.GetAllSubMajorGroupsHandler)
	r.GET("/sub-major-groups/:id", ctrl.GetSubMajorGroupByIDHandler)
	r.PUT("/sub-major-groups/:id", ctrl.UpdateSubMajorGroupHandler)
	r.DELETE("/sub-major-groups/:id", ctrl.DeleteSubMajorGroupHandler)

	r.POST("/minor-groups", ctrl.CreateMinorGroupHandler)
	r.GET("/minor-groups", ctrl.GetAllMinorGroupsHandler)
	r.GET("/minor-groups/:id", ctrl.GetMinorGroupByIDHandler)
	r.PUT("/minor-groups/:id", ctrl.UpdateMinorGroupHandler)
	r.DELETE("/minor-groups/:id", ctrl.DeleteMinorGroupHandler)

	r.POST("/unit-groups", ctrl.CreateUnitGroupHandler)
	r.GET("/unit-groups", ctrl.GetAllUnitGroupsHandler)
	r.GET("/unit-groups/:id", ctrl.GetUnitGroupByIDHandler)
	r.PUT("/unit-groups/:id", ctrl.UpdateUnitGroupHandler)
	r.DELETE("/unit-groups/:id", ctrl.DeleteUnitGroupHandler)

	r.POST("/occupation-groups", ctrl.CreateOccupationGroupHandler)
	r.GET("/occupation-groups", ctrl.GetAllOccupationGroupsHandler)
	r.GET("/occupation-groups/:id", ctrl.GetOccupationGroupByIDHandler)
	r.PUT("/occupation-groups/:id", ctrl.UpdateOccupationGroupHandler)
	r.DELETE("/occupation-groups/:id", ctrl.DeleteOccupationGroupHandler)

	r.POST("/industry-sectors", ctrl.CreateIndustrySectorHandler)
	r.GET("/industry-sectors", ctrl.GetAllIndustrySectorsHandler)
	r.GET("/industry-sectors/:id", ctrl.GetIndustrySectorByIDHandler)
	r.PUT("/industry-sectors/:id", ctrl.UpdateIndustrySectorHandler)
	r.DELETE("/industry-sectors/:id", ctrl.DeleteIndustrySectorHandler)

	r.POST("/industry-divisions", ctrl.CreateIndustryDivisionHandler)
	r.GET("/industry-divisions", ctrl.GetAllIndustryDivisionsHandler)
	r.GET("/industry-divisions/:id", ctrl.GetIndustryDivisionByIDHandler)
	r.PUT("/industry-divisions/:id", ctrl.UpdateIndustryDivisionHandler)
	r.DELETE("/industry-divisions/:id", ctrl.DeleteIndustryDivisionHandler)

	r.POST("/industry-groups", ctrl.CreateIndustryGroupHandler)
	r.GET("/industry-groups", ctrl.GetAllIndustryGroupsHandler)
	r.GET("/industry-groups/:id", ctrl.GetIndustryGroupByIDHandler)
	r.PUT("/industry-groups/:id", ctrl.UpdateIndustryGroupHandler)
	r.DELETE("/industry-groups/:id", ctrl.DeleteIndustryGroupHandler)

	r.POST("/industry-classes", ctrl.CreateIndustryClassHandler)
	r.GET("/industry-classes", ctrl.GetAllIndustryClassesHandler)
	r.GET("/industry-classes/:id", ctrl.GetIndustryClassByIDHandler)
	r.PUT("/industry-classes/:id", ctrl.UpdateIndustryClassHandler)
	r.DELETE("/industry-classes/:id", ctrl.DeleteIndustryClassHandler)

	r.POST("/industry-subclasses", ctrl.CreateIndustrySubclassHandler)
	r.GET("/industry-subclasses", ctrl.GetAllIndustrySubclassesHandler)
	r.GET("/industry-subclasses/:id", ctrl.GetIndustrySubclassByIDHandler)
	r.PUT("/industry-subclasses/:id", ctrl.UpdateIndustrySubclassHandler)
	r.DELETE("/industry-subclasses/:id", ctrl.DeleteIndustrySubclassHandler)

	return r
}

func doRequest(t *testing.T, router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestIntegration_EducationLevel_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	// Create
	w := doRequest(t, router, http.MethodPost, "/education-levels", map[string]string{"level": "Bachelor's Degree"})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.EducationLevel
	json.Unmarshal(w.Body.Bytes(), &created)
	if created.ID == 0 {
		t.Fatalf("create: expected ID to be populated, got 0")
	}

	// Read back
	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/education-levels/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
	var fetched models.EducationLevel
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Level != "Bachelor's Degree" {
		t.Fatalf("get: expected Level %q, got %q", "Bachelor's Degree", fetched.Level)
	}

	// Update
	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/education-levels/%d", created.ID), map[string]string{"level": "Master's Degree"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Read back again to confirm the update actually persisted, not just returned
	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/education-levels/%d", created.ID), nil)
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Level != "Master's Degree" {
		t.Fatalf("get after update: expected Level %q, got %q", "Master's Degree", fetched.Level)
	}

	// Delete
	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/education-levels/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Confirm it's actually gone
	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/education-levels/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= Formality =================

func TestIntegration_Formality_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/formalities", map[string]string{"formality_type": "Formal"})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.Formality
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/formalities/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/formalities/%d", created.ID), map[string]string{"formality_type": "Informal"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
	var updated models.Formality
	json.Unmarshal(w.Body.Bytes(), &updated)
	if updated.FormalityType != "Informal" {
		t.Fatalf("update: expected FormalityType %q, got %q", "Informal", updated.FormalityType)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/formalities/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/formalities/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= Gender =================

func TestIntegration_Gender_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/genders", map[string]string{"gender_type": "Male"})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.Gender
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/genders/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/genders/%d", created.ID), map[string]string{"gender_type": "Female"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/genders/%d", created.ID), nil)
	var fetched models.Gender
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.GenderType != "Female" {
		t.Fatalf("get after update: expected GenderType %q, got %q", "Female", fetched.GenderType)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/genders/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/genders/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= EmploymentSector =================

func TestIntegration_EmploymentSector_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/employment-sectors", map[string]string{"sector": "Private"})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.EmploymentSector
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/employment-sectors/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/employment-sectors/%d", created.ID), map[string]string{"sector": "Government"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/employment-sectors/%d", created.ID), nil)
	var fetched models.EmploymentSector
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Sector != "Government" {
		t.Fatalf("get after update: expected Sector %q, got %q", "Government", fetched.Sector)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/employment-sectors/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/employment-sectors/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= VocationalEducation =================

func TestIntegration_VocationalEducation_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/vocational-educations", map[string]string{"level": "NVQ 3"})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.VocationalEducation
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/vocational-educations/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/vocational-educations/%d", created.ID), map[string]string{"level": "NVQ 5"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/vocational-educations/%d", created.ID), nil)
	var fetched models.VocationalEducation
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Level != "NVQ 5" {
		t.Fatalf("get after update: expected Level %q, got %q", "NVQ 5", fetched.Level)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/vocational-educations/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/vocational-educations/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= Experience =================
// No "get all" endpoint exists for this resource, but the lifecycle
// (create -> get -> update -> delete) is still fully exercised via GetByID.

func TestIntegration_Experience_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/experiences", map[string]string{"name": "Entry Level"})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.Experience
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/experiences/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/experiences/%d", created.ID), map[string]string{"name": "Senior Level"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/experiences/%d", created.ID), nil)
	var fetched models.Experience
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Name != "Senior Level" {
		t.Fatalf("get after update: expected Name %q, got %q", "Senior Level", fetched.Name)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/experiences/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/experiences/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= MajorGroup =================

func TestIntegration_MajorGroup_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/major-groups", map[string]string{"name": "Managers", "code": "1"})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.MajorGroup
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/major-groups/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/major-groups/%d", created.ID), map[string]string{"name": "Senior Managers"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/major-groups/%d", created.ID), nil)
	var fetched models.MajorGroup
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Name != "Senior Managers" {
		t.Fatalf("get after update: expected Name %q, got %q", "Senior Managers", fetched.Name)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/major-groups/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/major-groups/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= SubMajorGroup (requires a parent MajorGroup) =================

func TestIntegration_SubMajorGroup_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	parent := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&parent)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/sub-major-groups", map[string]interface{}{
		"major_group_id": parent.ID, "name": "Chief Executives", "code": "11",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.SubMajorGroup
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/sub-major-groups/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
	var fetched models.SubMajorGroup
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.MajorGroup == nil || fetched.MajorGroup.Name != "Managers" {
		t.Fatalf("get: expected preloaded MajorGroup.Name %q, got %+v", "Managers", fetched.MajorGroup)
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/sub-major-groups/%d", created.ID), map[string]string{"name": "Senior Executives"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/sub-major-groups/%d", created.ID), nil)
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Name != "Senior Executives" {
		t.Fatalf("get after update: expected Name %q, got %q", "Senior Executives", fetched.Name)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/sub-major-groups/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/sub-major-groups/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= MinorGroup (requires MajorGroup -> SubMajorGroup) =================

func TestIntegration_MinorGroup_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	major := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&major)
	sub := models.SubMajorGroup{MajorGroupID: major.ID, Name: "Chief Executives", Code: "11"}
	db.Create(&sub)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/minor-groups", map[string]interface{}{
		"sub_major_group_id": sub.ID, "name": "Legislators", "code": "111",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.MinorGroup
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/minor-groups/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/minor-groups/%d", created.ID), map[string]string{"name": "Senior Legislators"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/minor-groups/%d", created.ID), nil)
	var fetched models.MinorGroup
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Name != "Senior Legislators" {
		t.Fatalf("get after update: expected Name %q, got %q", "Senior Legislators", fetched.Name)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/minor-groups/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/minor-groups/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= UnitGroup (requires MajorGroup -> SubMajorGroup -> MinorGroup) =================

func TestIntegration_UnitGroup_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	major := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&major)
	sub := models.SubMajorGroup{MajorGroupID: major.ID, Name: "Chief Executives", Code: "11"}
	db.Create(&sub)
	minor := models.MinorGroup{SubMajorGroupID: sub.ID, Name: "Legislators", Code: "111"}
	db.Create(&minor)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/unit-groups", map[string]interface{}{
		"minor_group_id": minor.ID, "name": "Senior Officials", "code": "1111",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.UnitGroup
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/unit-groups/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/unit-groups/%d", created.ID), map[string]string{"name": "Senior Govt Officials"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/unit-groups/%d", created.ID), nil)
	var fetched models.UnitGroup
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Name != "Senior Govt Officials" {
		t.Fatalf("get after update: expected Name %q, got %q", "Senior Govt Officials", fetched.Name)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/unit-groups/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/unit-groups/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= OccupationGroup (full MajorGroup -> ... -> UnitGroup chain) =================

func TestIntegration_OccupationGroup_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	major := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&major)
	sub := models.SubMajorGroup{MajorGroupID: major.ID, Name: "Chief Executives", Code: "11"}
	db.Create(&sub)
	minor := models.MinorGroup{SubMajorGroupID: sub.ID, Name: "Legislators", Code: "111"}
	db.Create(&minor)
	unit := models.UnitGroup{MinorGroupID: minor.ID, Name: "Senior Officials", Code: "1111"}
	db.Create(&unit)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/occupation-groups", map[string]interface{}{
		"unit_group_id": unit.ID, "name": "Legislator", "code": "11111",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.OccupationGroup
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/occupation-groups/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/occupation-groups/%d", created.ID), map[string]string{"name": "Senior Legislator"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/occupation-groups/%d", created.ID), nil)
	var fetched models.OccupationGroup
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Name != "Senior Legislator" {
		t.Fatalf("get after update: expected Name %q, got %q", "Senior Legislator", fetched.Name)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/occupation-groups/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/occupation-groups/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= IndustrySector =================

func TestIntegration_IndustrySector_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/industry-sectors", map[string]string{"name": "Agriculture", "code": "A"})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.IndustrySector
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-sectors/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/industry-sectors/%d", created.ID), map[string]string{"name": "Agri & Fisheries"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-sectors/%d", created.ID), nil)
	var fetched models.IndustrySector
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Name != "Agri & Fisheries" {
		t.Fatalf("get after update: expected Name %q, got %q", "Agri & Fisheries", fetched.Name)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/industry-sectors/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-sectors/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= IndustryDivision (requires a parent IndustrySector) =================

func TestIntegration_IndustryDivision_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&sector)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/industry-divisions", map[string]interface{}{
		"industry_sector_id": sector.ID, "name": "Crop Farming", "code": "01",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.IndustryDivision
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-divisions/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
	var fetched models.IndustryDivision
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.IndustrySector == nil || fetched.IndustrySector.Name != "Agriculture" {
		t.Fatalf("get: expected preloaded IndustrySector.Name %q, got %+v", "Agriculture", fetched.IndustrySector)
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/industry-divisions/%d", created.ID), map[string]string{"name": "Arable Farming"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-divisions/%d", created.ID), nil)
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Name != "Arable Farming" {
		t.Fatalf("get after update: expected Name %q, got %q", "Arable Farming", fetched.Name)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/industry-divisions/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-divisions/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= IndustryGroup (requires IndustrySector -> IndustryDivision) =================

func TestIntegration_IndustryGroup_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&sector)
	division := models.IndustryDivision{IndustrySectorID: sector.ID, Name: "Crop Farming", Code: "01"}
	db.Create(&division)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/industry-groups", map[string]interface{}{
		"industry_division_id": division.ID, "name": "Cereal Growing", "code": "011",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.IndustryGroup
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-groups/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/industry-groups/%d", created.ID), map[string]string{"name": "Grain Growing"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-groups/%d", created.ID), nil)
	var fetched models.IndustryGroup
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Name != "Grain Growing" {
		t.Fatalf("get after update: expected Name %q, got %q", "Grain Growing", fetched.Name)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/industry-groups/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-groups/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= IndustryClass (requires IndustrySector -> Division -> Group) =================

func TestIntegration_IndustryClass_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&sector)
	division := models.IndustryDivision{IndustrySectorID: sector.ID, Name: "Crop Farming", Code: "01"}
	db.Create(&division)
	group := models.IndustryGroup{IndustryDivisionID: division.ID, Name: "Cereal Growing", Code: "011"}
	db.Create(&group)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/industry-classes", map[string]interface{}{
		"industry_group_id": group.ID, "name": "Rice Growing", "code": "0111",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.IndustryClass
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-classes/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/industry-classes/%d", created.ID), map[string]string{"name": "Paddy Growing"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-classes/%d", created.ID), nil)
	var fetched models.IndustryClass
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Name != "Paddy Growing" {
		t.Fatalf("get after update: expected Name %q, got %q", "Paddy Growing", fetched.Name)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/industry-classes/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-classes/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

// ================= IndustrySubclass (full IndustrySector -> ... -> IndustryClass chain) =================

func TestIntegration_IndustrySubclass_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&sector)
	division := models.IndustryDivision{IndustrySectorID: sector.ID, Name: "Crop Farming", Code: "01"}
	db.Create(&division)
	group := models.IndustryGroup{IndustryDivisionID: division.ID, Name: "Cereal Growing", Code: "011"}
	db.Create(&group)
	class := models.IndustryClass{IndustryGroupID: group.ID, Name: "Rice Growing", Code: "0111"}
	db.Create(&class)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/industry-subclasses", map[string]interface{}{
		"industry_class_id": class.ID, "name": "Rice Milling", "code": "01111",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var created models.IndustrySubclass
	json.Unmarshal(w.Body.Bytes(), &created)

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-subclasses/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodPut, fmt.Sprintf("/industry-subclasses/%d", created.ID), map[string]string{"name": "Paddy Milling"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-subclasses/%d", created.ID), nil)
	var fetched models.IndustrySubclass
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.Name != "Paddy Milling" {
		t.Fatalf("get after update: expected Name %q, got %q", "Paddy Milling", fetched.Name)
	}

	w = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/industry-subclasses/%d", created.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	w = doRequest(t, router, http.MethodGet, fmt.Sprintf("/industry-subclasses/%d", created.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}
}