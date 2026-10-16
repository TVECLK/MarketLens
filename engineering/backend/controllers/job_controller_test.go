package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"marketlens-go-backend/controllers"
	"marketlens-go-backend/models"
	"marketlens-go-backend/repositories"
)

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

// ---------- CreateEducationLevelHandler ----------

func TestCreateEducationLevelHandler(t *testing.T) {
	t.Run("valid payload returns 201 with created item", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodPost, "/education-levels", map[string]string{
			"level": "Bachelor's Degree",
		})

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var got models.EducationLevel
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if got.ID == 0 {
			t.Fatalf("expected ID to be populated in response, got 0")
		}
		if got.Level != "Bachelor's Degree" {
			t.Fatalf("expected Level %q, got %q", "Bachelor's Degree", got.Level)
		}
	})

	t.Run("malformed JSON body returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		req := httptest.NewRequest(http.MethodPost, "/education-levels", bytes.NewReader([]byte("{not valid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})
}

// ---------- GetAllEducationLevelsHandler ----------

func TestGetAllEducationLevelsHandler(t *testing.T) {
	t.Run("empty table returns 200 with count 0", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/education-levels", nil)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var body map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if int(body["count"].(float64)) != 0 {
			t.Fatalf("expected count 0, got %v", body["count"])
		}
	})

	t.Run("seeded rows are returned with correct count", func(t *testing.T) {
		db := setupTestDB(t)
		db.Create(&models.EducationLevel{Level: "Diploma"})
		db.Create(&models.EducationLevel{Level: "Master's Degree"})
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/education-levels", nil)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var body map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if int(body["count"].(float64)) != 2 {
			t.Fatalf("expected count 2, got %v", body["count"])
		}
	})
}

// ---------- GetEducationLevelByIDHandler ----------

func TestGetEducationLevelByIDHandler(t *testing.T) {
	t.Run("existing id returns 200 with the item", func(t *testing.T) {
		db := setupTestDB(t)
		seeded := models.EducationLevel{Level: "PhD"}
		db.Create(&seeded)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, fmt.Sprintf("/education-levels/%d", seeded.ID), nil)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var got models.EducationLevel
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if got.Level != "PhD" {
			t.Fatalf("expected Level %q, got %q", "PhD", got.Level)
		}
	})

	t.Run("non-numeric id returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/education-levels/abc", nil)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/education-levels/999999", nil)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

// ---------- UpdateEducationLevelHandler ----------

func TestUpdateEducationLevelHandler(t *testing.T) {
	t.Run("valid update returns 200 with updated item", func(t *testing.T) {
		db := setupTestDB(t)
		seeded := models.EducationLevel{Level: "Diploma"}
		db.Create(&seeded)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/education-levels/%d", seeded.ID), map[string]string{
			"level": "Advanced Diploma",
		})

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var got models.EducationLevel
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if got.Level != "Advanced Diploma" {
			t.Fatalf("expected Level %q, got %q", "Advanced Diploma", got.Level)
		}
	})

	t.Run("non-numeric id returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodPut, "/education-levels/abc", map[string]string{"level": "X"})

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("malformed JSON body returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		seeded := models.EducationLevel{Level: "Diploma"}
		db.Create(&seeded)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/education-levels/%d", seeded.ID), bytes.NewReader([]byte("{bad json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	// NOTE: this documents the handler's CURRENT behavior, not necessarily
	// the desired behavior. UpdateEducationLevelHandler returns 500 for a
	// non-existent id (unlike GetEducationLevelByIDHandler, which maps
	// gorm.ErrRecordNotFound to 404). If you'd rather it return 404, the
	// handler needs an errors.Is(err, gorm.ErrRecordNotFound) check same as
	// the GET handler has — flag this to your team.
	t.Run("non-existent id currently returns 500 (not 404)", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodPut, "/education-levels/999999", map[string]string{
			"level": "Should Not Apply",
		})

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected current status %d, got %d: %s", http.StatusInternalServerError, w.Code, w.Body.String())
		}
	})
}

// ---------- DeleteEducationLevelHandler ----------

func TestDeleteEducationLevelHandler(t *testing.T) {
	t.Run("existing id returns 200 and row is actually removed", func(t *testing.T) {
		db := setupTestDB(t)
		seeded := models.EducationLevel{Level: "Diploma"}
		db.Create(&seeded)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/education-levels/%d", seeded.ID), nil)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var count int64
		db.Model(&models.EducationLevel{}).Where("id = ?", seeded.ID).Count(&count)
		if count != 0 {
			t.Fatalf("expected row to be deleted, but %d rows still match id %d", count, seeded.ID)
		}
	})

	t.Run("non-numeric id returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodDelete, "/education-levels/abc", nil)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})
}

// ---------- CreateFormalityHandler ----------

func TestCreateFormalityHandler(t *testing.T) {
	t.Run("valid payload returns 201 with created item", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodPost, "/formalities", map[string]string{
			"formality_type": "Formal",
		})

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var got models.Formality
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if got.ID == 0 {
			t.Fatalf("expected ID to be populated in response, got 0")
		}
		if got.FormalityType != "Formal" {
			t.Fatalf("expected FormalityType %q, got %q", "Formal", got.FormalityType)
		}
	})

	t.Run("malformed JSON body returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		req := httptest.NewRequest(http.MethodPost, "/formalities", bytes.NewReader([]byte("{not valid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})
}

// ---------- GetAllFormalitiesHandler ----------

func TestGetAllFormalitiesHandler(t *testing.T) {
	t.Run("empty table returns 200 with count 0", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/formalities", nil)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var body map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if int(body["count"].(float64)) != 0 {
			t.Fatalf("expected count 0, got %v", body["count"])
		}
	})

	t.Run("seeded rows are returned with correct count", func(t *testing.T) {
		db := setupTestDB(t)
		db.Create(&models.Formality{FormalityType: "Formal"})
		db.Create(&models.Formality{FormalityType: "Informal"})
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/formalities", nil)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var body map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if int(body["count"].(float64)) != 2 {
			t.Fatalf("expected count 2, got %v", body["count"])
		}
	})
}

// ---------- GetFormalityByIDHandler ----------

func TestGetFormalityByIDHandler(t *testing.T) {
	t.Run("existing id returns 200 with the item", func(t *testing.T) {
		db := setupTestDB(t)
		seeded := models.Formality{FormalityType: "Formal"}
		db.Create(&seeded)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, fmt.Sprintf("/formalities/%d", seeded.ID), nil)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var got models.Formality
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if got.FormalityType != "Formal" {
			t.Fatalf("expected FormalityType %q, got %q", "Formal", got.FormalityType)
		}
	})

	t.Run("non-numeric id returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/formalities/abc", nil)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/formalities/999999", nil)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

// ---------- UpdateFormalityHandler ----------

func TestUpdateFormalityHandler(t *testing.T) {
	t.Run("valid update returns 200 with updated item", func(t *testing.T) {
		db := setupTestDB(t)
		seeded := models.Formality{FormalityType: "Formal"}
		db.Create(&seeded)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/formalities/%d", seeded.ID), map[string]string{
			"formality_type": "Semi-Formal",
		})

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var got models.Formality
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if got.FormalityType != "Semi-Formal" {
			t.Fatalf("expected FormalityType %q, got %q", "Semi-Formal", got.FormalityType)
		}
	})

	t.Run("non-numeric id returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodPut, "/formalities/abc", map[string]string{"formality_type": "X"})

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("malformed JSON body returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		seeded := models.Formality{FormalityType: "Formal"}
		db.Create(&seeded)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/formalities/%d", seeded.ID), bytes.NewReader([]byte("{bad json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	// Same note as UpdateEducationLevelHandler: no gorm.ErrRecordNotFound
	// check in this handler, so a non-existent id currently falls through
	// to 500 rather than 404.
	t.Run("non-existent id currently returns 500 (not 404)", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodPut, "/formalities/999999", map[string]string{
			"formality_type": "Should Not Apply",
		})

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected current status %d, got %d: %s", http.StatusInternalServerError, w.Code, w.Body.String())
		}
	})
}

// ---------- DeleteFormalityHandler ----------

func TestDeleteFormalityHandler(t *testing.T) {
	t.Run("existing id returns 200 and row is actually removed", func(t *testing.T) {
		db := setupTestDB(t)
		seeded := models.Formality{FormalityType: "Formal"}
		db.Create(&seeded)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/formalities/%d", seeded.ID), nil)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var count int64
		db.Model(&models.Formality{}).Where("id = ?", seeded.ID).Count(&count)
		if count != 0 {
			t.Fatalf("expected row to be deleted, but %d rows still match id %d", count, seeded.ID)
		}
	})

	t.Run("non-numeric id returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodDelete, "/formalities/abc", nil)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})
}

// ================= Gender =================

func TestCreateGenderHandler(t *testing.T) {
	t.Run("valid payload returns 201", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodPost, "/genders", map[string]string{"gender_type": "Male"})
		if w.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
		}
		var got models.Gender
		json.Unmarshal(w.Body.Bytes(), &got)
		if got.GenderType != "Male" {
			t.Fatalf("expected GenderType %q, got %q", "Male", got.GenderType)
		}
	})

	t.Run("malformed JSON returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		req := httptest.NewRequest(http.MethodPost, "/genders", bytes.NewReader([]byte("{bad")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})
}

func TestGetAllGendersHandler(t *testing.T) {
	db := setupTestDB(t)
	db.Create(&models.Gender{GenderType: "Male"})
	db.Create(&models.Gender{GenderType: "Female"})
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodGet, "/genders", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if int(body["count"].(float64)) != 2 {
		t.Fatalf("expected count 2, got %v", body["count"])
	}
}

func TestGetGenderByIDHandler(t *testing.T) {
	t.Run("existing id returns 200", func(t *testing.T) {
		db := setupTestDB(t)
		seeded := models.Gender{GenderType: "Male"}
		db.Create(&seeded)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, fmt.Sprintf("/genders/%d", seeded.ID), nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})

	t.Run("non-numeric id returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/genders/abc", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/genders/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

func TestUpdateGenderHandler(t *testing.T) {
	db := setupTestDB(t)
	seeded := models.Gender{GenderType: "Male"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/genders/%d", seeded.ID), map[string]string{"gender_type": "Female"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteGenderHandler(t *testing.T) {
	db := setupTestDB(t)
	seeded := models.Gender{GenderType: "Male"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/genders/%d", seeded.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// ================= EmploymentSector =================

func TestCreateEmploymentSectorHandler(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/employment-sectors", map[string]string{"sector": "Private"})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestGetAllEmploymentSectorsHandler(t *testing.T) {
	db := setupTestDB(t)
	db.Create(&models.EmploymentSector{Sector: "Private"})
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodGet, "/employment-sectors", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetEmploymentSectorByIDHandler(t *testing.T) {
	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/employment-sectors/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

func TestUpdateEmploymentSectorHandler(t *testing.T) {
	db := setupTestDB(t)
	seeded := models.EmploymentSector{Sector: "Private"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/employment-sectors/%d", seeded.ID), map[string]string{"sector": "Government"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteEmploymentSectorHandler(t *testing.T) {
	db := setupTestDB(t)
	seeded := models.EmploymentSector{Sector: "Private"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/employment-sectors/%d", seeded.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// ================= VocationalEducation =================

func TestCreateVocationalEducationHandler(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/vocational-educations", map[string]string{"level": "NVQ 3"})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestGetAllVocationalEducationsHandler(t *testing.T) {
	db := setupTestDB(t)
	db.Create(&models.VocationalEducation{Level: "NVQ 3"})
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodGet, "/vocational-educations", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetVocationalEducationByIDHandler(t *testing.T) {
	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/vocational-educations/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

func TestUpdateVocationalEducationHandler(t *testing.T) {
	db := setupTestDB(t)
	seeded := models.VocationalEducation{Level: "NVQ 3"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/vocational-educations/%d", seeded.ID), map[string]string{"level": "NVQ 5"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteVocationalEducationHandler(t *testing.T) {
	db := setupTestDB(t)
	seeded := models.VocationalEducation{Level: "NVQ 3"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/vocational-educations/%d", seeded.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// ================= Experience =================
// NOTE: no GetAllExperiencesHandler was provided, so no "get all" test here.

func TestCreateExperienceHandler(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/experiences", map[string]string{"name": "Entry Level"})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestGetExperienceByIDHandler(t *testing.T) {
	t.Run("existing id returns 200", func(t *testing.T) {
		db := setupTestDB(t)
		seeded := models.Experience{Name: "Entry Level"}
		db.Create(&seeded)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, fmt.Sprintf("/experiences/%d", seeded.ID), nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})

	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/experiences/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

func TestUpdateExperienceHandler(t *testing.T) {
	db := setupTestDB(t)
	seeded := models.Experience{Name: "Entry Level"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/experiences/%d", seeded.ID), map[string]string{"name": "Senior"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteExperienceHandler(t *testing.T) {
	db := setupTestDB(t)
	seeded := models.Experience{Name: "Entry Level"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/experiences/%d", seeded.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// ================= MajorGroup (has from-date/to-date filtering) =================

func TestCreateMajorGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/major-groups", map[string]string{"name": "Managers", "code": "1"})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestGetAllMajorGroupsHandler(t *testing.T) {
	t.Run("no date params returns all rows", func(t *testing.T) {
		db := setupTestDB(t)
		db.Create(&models.MajorGroup{Name: "Managers", Code: "1"})
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/major-groups", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}
		var body map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &body)
		if int(body["count"].(float64)) != 1 {
			t.Fatalf("expected count 1, got %v", body["count"])
		}
	})

	t.Run("valid from-date/to-date returns filtered rows", func(t *testing.T) {
		db := setupTestDB(t)
		db.Create(&models.MajorGroup{Name: "Managers", Code: "1"})
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		today := time.Now().Format("2006-01-02")
		w := doRequest(t, router, http.MethodGet, fmt.Sprintf("/major-groups?from-date=%s&to-date=%s", today, today), nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})

	t.Run("invalid from-date format returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/major-groups?from-date=not-a-date&to-date=2024-01-01", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("invalid to-date format returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/major-groups?from-date=2024-01-01&to-date=not-a-date", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("to-date before from-date returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/major-groups?from-date=2024-06-01&to-date=2024-01-01", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})
}

func TestGetMajorGroupByIDHandler(t *testing.T) {
	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/major-groups/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

func TestUpdateMajorGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	seeded := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/major-groups/%d", seeded.ID), map[string]string{"name": "Senior Managers"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteMajorGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	seeded := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/major-groups/%d", seeded.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// ================= SubMajorGroup =================

func TestCreateSubMajorGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	parent := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&parent)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/sub-major-groups", map[string]interface{}{
		"major_group_id": parent.ID, "name": "Chief Executives", "code": "11",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestGetAllSubMajorGroupsHandler(t *testing.T) {
	db := setupTestDB(t)
	parent := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&parent)
	db.Create(&models.SubMajorGroup{MajorGroupID: parent.ID, Name: "Chief Executives", Code: "11"})
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodGet, "/sub-major-groups", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetSubMajorGroupByIDHandler(t *testing.T) {
	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/sub-major-groups/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

func TestUpdateSubMajorGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	parent := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&parent)
	seeded := models.SubMajorGroup{MajorGroupID: parent.ID, Name: "Chief Executives", Code: "11"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/sub-major-groups/%d", seeded.ID), map[string]string{"name": "Senior Executives"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteSubMajorGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	parent := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&parent)
	seeded := models.SubMajorGroup{MajorGroupID: parent.ID, Name: "Chief Executives", Code: "11"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/sub-major-groups/%d", seeded.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// ================= MinorGroup =================

func seedMinorGroupParent(t *testing.T, db *gorm.DB) uint {
	t.Helper()
	major := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&major)
	sub := models.SubMajorGroup{MajorGroupID: major.ID, Name: "Chief Executives", Code: "11"}
	db.Create(&sub)
	return sub.ID
}

func TestCreateMinorGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	subID := seedMinorGroupParent(t, db)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/minor-groups", map[string]interface{}{
		"sub_major_group_id": subID, "name": "Legislators", "code": "111",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestGetAllMinorGroupsHandler(t *testing.T) {
	db := setupTestDB(t)
	subID := seedMinorGroupParent(t, db)
	db.Create(&models.MinorGroup{SubMajorGroupID: subID, Name: "Legislators", Code: "111"})
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodGet, "/minor-groups", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetMinorGroupByIDHandler(t *testing.T) {
	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/minor-groups/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

func TestUpdateMinorGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	subID := seedMinorGroupParent(t, db)
	seeded := models.MinorGroup{SubMajorGroupID: subID, Name: "Legislators", Code: "111"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/minor-groups/%d", seeded.ID), map[string]string{"name": "Senior Legislators"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteMinorGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	subID := seedMinorGroupParent(t, db)
	seeded := models.MinorGroup{SubMajorGroupID: subID, Name: "Legislators", Code: "111"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/minor-groups/%d", seeded.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// ================= UnitGroup =================

func seedUnitGroupParent(t *testing.T, db *gorm.DB) uint {
	t.Helper()
	minorID := seedMinorGroupParent(t, db)
	minor := models.MinorGroup{SubMajorGroupID: minorID, Name: "Legislators", Code: "111"}
	db.Create(&minor)
	return minor.ID
}

func TestCreateUnitGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	minorID := seedUnitGroupParent(t, db)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/unit-groups", map[string]interface{}{
		"minor_group_id": minorID, "name": "Senior Officials", "code": "1111",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestGetAllUnitGroupsHandler(t *testing.T) {
	db := setupTestDB(t)
	minorID := seedUnitGroupParent(t, db)
	db.Create(&models.UnitGroup{MinorGroupID: minorID, Name: "Senior Officials", Code: "1111"})
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodGet, "/unit-groups", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetUnitGroupByIDHandler(t *testing.T) {
	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/unit-groups/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

func TestUpdateUnitGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	minorID := seedUnitGroupParent(t, db)
	seeded := models.UnitGroup{MinorGroupID: minorID, Name: "Senior Officials", Code: "1111"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/unit-groups/%d", seeded.ID), map[string]string{"name": "Senior Govt Officials"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteUnitGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	minorID := seedUnitGroupParent(t, db)
	seeded := models.UnitGroup{MinorGroupID: minorID, Name: "Senior Officials", Code: "1111"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/unit-groups/%d", seeded.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// ================= OccupationGroup (has limit/offset/total pagination) =================

func seedOccupationGroupParent(t *testing.T, db *gorm.DB) uint {
	t.Helper()
	unitID := seedUnitGroupParent(t, db)
	unit := models.UnitGroup{MinorGroupID: unitID, Name: "Senior Officials", Code: "1111"}
	db.Create(&unit)
	return unit.ID
}

func TestCreateOccupationGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	unitID := seedOccupationGroupParent(t, db)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/occupation-groups", map[string]interface{}{
		"unit_group_id": unitID, "name": "Legislator", "code": "11111",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestGetAllOccupationGroupsHandler(t *testing.T) {
	t.Run("default limit is 20, total reflects all rows", func(t *testing.T) {
		db := setupTestDB(t)
		unitID := seedOccupationGroupParent(t, db)
		for i := 0; i < 5; i++ {
			db.Create(&models.OccupationGroup{UnitGroupID: unitID, Name: "Group", Code: "X"})
		}
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/occupation-groups", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}
		var body map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &body)
		if int(body["limit"].(float64)) != 20 {
			t.Fatalf("expected default limit 20, got %v", body["limit"])
		}
		if int(body["total"].(float64)) != 5 {
			t.Fatalf("expected total 5, got %v", body["total"])
		}
	})

	t.Run("limit query param is respected and capped at 100", func(t *testing.T) {
		db := setupTestDB(t)
		unitID := seedOccupationGroupParent(t, db)
		db.Create(&models.OccupationGroup{UnitGroupID: unitID, Name: "Group", Code: "X"})
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/occupation-groups?limit=500", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}
		var body map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &body)
		if int(body["limit"].(float64)) != 100 {
			t.Fatalf("expected limit capped at 100, got %v", body["limit"])
		}
	})

	t.Run("offset query param is applied", func(t *testing.T) {
		db := setupTestDB(t)
		unitID := seedOccupationGroupParent(t, db)
		for i := 0; i < 3; i++ {
			db.Create(&models.OccupationGroup{UnitGroupID: unitID, Name: "Group", Code: "X"})
		}
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/occupation-groups?offset=2", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}
		var body map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &body)
		if int(body["count"].(float64)) != 1 {
			t.Fatalf("expected 1 item after offset=2 of 3, got %v", body["count"])
		}
	})
}

func TestGetOccupationGroupByIDHandler(t *testing.T) {
	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/occupation-groups/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

func TestUpdateOccupationGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	unitID := seedOccupationGroupParent(t, db)
	seeded := models.OccupationGroup{UnitGroupID: unitID, Name: "Legislator", Code: "11111"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/occupation-groups/%d", seeded.ID), map[string]string{"name": "Senior Legislator"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteOccupationGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	unitID := seedOccupationGroupParent(t, db)
	seeded := models.OccupationGroup{UnitGroupID: unitID, Name: "Legislator", Code: "11111"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/occupation-groups/%d", seeded.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// ================= IndustrySector (has from-date/to-date filtering) =================

func TestCreateIndustrySectorHandler(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/industry-sectors", map[string]string{"name": "Agriculture", "code": "A"})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestGetAllIndustrySectorsHandler(t *testing.T) {
	t.Run("no date params returns all rows", func(t *testing.T) {
		db := setupTestDB(t)
		db.Create(&models.IndustrySector{Name: "Agriculture", Code: "A"})
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/industry-sectors", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})

	t.Run("invalid from-date format returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/industry-sectors?from-date=bad&to-date=2024-01-01", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})

	t.Run("to-date before from-date returns 400", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/industry-sectors?from-date=2024-06-01&to-date=2024-01-01", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})
}

func TestGetIndustrySectorByIDHandler(t *testing.T) {
	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/industry-sectors/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

func TestUpdateIndustrySectorHandler(t *testing.T) {
	db := setupTestDB(t)
	seeded := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/industry-sectors/%d", seeded.ID), map[string]string{"name": "Agri & Fisheries"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteIndustrySectorHandler(t *testing.T) {
	db := setupTestDB(t)
	seeded := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/industry-sectors/%d", seeded.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// ================= IndustryDivision =================

func TestCreateIndustryDivisionHandler(t *testing.T) {
	db := setupTestDB(t)
	sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&sector)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/industry-divisions", map[string]interface{}{
		"industry_sector_id": sector.ID, "name": "Crop Farming", "code": "01",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestGetAllIndustryDivisionsHandler(t *testing.T) {
	db := setupTestDB(t)
	sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&sector)
	db.Create(&models.IndustryDivision{IndustrySectorID: sector.ID, Name: "Crop Farming", Code: "01"})
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodGet, "/industry-divisions", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetIndustryDivisionByIDHandler(t *testing.T) {
	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/industry-divisions/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

func TestUpdateIndustryDivisionHandler(t *testing.T) {
	db := setupTestDB(t)
	sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&sector)
	seeded := models.IndustryDivision{IndustrySectorID: sector.ID, Name: "Crop Farming", Code: "01"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/industry-divisions/%d", seeded.ID), map[string]string{"name": "Arable Farming"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteIndustryDivisionHandler(t *testing.T) {
	db := setupTestDB(t)
	sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&sector)
	seeded := models.IndustryDivision{IndustrySectorID: sector.ID, Name: "Crop Farming", Code: "01"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/industry-divisions/%d", seeded.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// ================= IndustryGroup =================

func seedIndustryGroupParent(t *testing.T, db *gorm.DB) uint {
	t.Helper()
	sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&sector)
	division := models.IndustryDivision{IndustrySectorID: sector.ID, Name: "Crop Farming", Code: "01"}
	db.Create(&division)
	return division.ID
}

func TestCreateIndustryGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	divisionID := seedIndustryGroupParent(t, db)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/industry-groups", map[string]interface{}{
		"industry_division_id": divisionID, "name": "Cereal Growing", "code": "011",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestGetAllIndustryGroupsHandler(t *testing.T) {
	db := setupTestDB(t)
	divisionID := seedIndustryGroupParent(t, db)
	db.Create(&models.IndustryGroup{IndustryDivisionID: divisionID, Name: "Cereal Growing", Code: "011"})
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodGet, "/industry-groups", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetIndustryGroupByIDHandler(t *testing.T) {
	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/industry-groups/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

func TestUpdateIndustryGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	divisionID := seedIndustryGroupParent(t, db)
	seeded := models.IndustryGroup{IndustryDivisionID: divisionID, Name: "Cereal Growing", Code: "011"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/industry-groups/%d", seeded.ID), map[string]string{"name": "Grain Growing"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteIndustryGroupHandler(t *testing.T) {
	db := setupTestDB(t)
	divisionID := seedIndustryGroupParent(t, db)
	seeded := models.IndustryGroup{IndustryDivisionID: divisionID, Name: "Cereal Growing", Code: "011"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/industry-groups/%d", seeded.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// ================= IndustryClass =================

func seedIndustryClassParent(t *testing.T, db *gorm.DB) uint {
	t.Helper()
	divisionID := seedIndustryGroupParent(t, db)
	group := models.IndustryGroup{IndustryDivisionID: divisionID, Name: "Cereal Growing", Code: "011"}
	db.Create(&group)
	return group.ID
}

func TestCreateIndustryClassHandler(t *testing.T) {
	db := setupTestDB(t)
	groupID := seedIndustryClassParent(t, db)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/industry-classes", map[string]interface{}{
		"industry_group_id": groupID, "name": "Rice Growing", "code": "0111",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestGetAllIndustryClassesHandler(t *testing.T) {
	db := setupTestDB(t)
	groupID := seedIndustryClassParent(t, db)
	db.Create(&models.IndustryClass{IndustryGroupID: groupID, Name: "Rice Growing", Code: "0111"})
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodGet, "/industry-classes", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetIndustryClassByIDHandler(t *testing.T) {
	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/industry-classes/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

func TestUpdateIndustryClassHandler(t *testing.T) {
	db := setupTestDB(t)
	groupID := seedIndustryClassParent(t, db)
	seeded := models.IndustryClass{IndustryGroupID: groupID, Name: "Rice Growing", Code: "0111"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/industry-classes/%d", seeded.ID), map[string]string{"name": "Paddy Growing"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteIndustryClassHandler(t *testing.T) {
	db := setupTestDB(t)
	groupID := seedIndustryClassParent(t, db)
	seeded := models.IndustryClass{IndustryGroupID: groupID, Name: "Rice Growing", Code: "0111"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/industry-classes/%d", seeded.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// ================= IndustrySubclass (has limit/offset/total pagination) =================

func seedIndustrySubclassParent(t *testing.T, db *gorm.DB) uint {
	t.Helper()
	groupID := seedIndustryClassParent(t, db)
	class := models.IndustryClass{IndustryGroupID: groupID, Name: "Rice Growing", Code: "0111"}
	db.Create(&class)
	return class.ID
}

func TestCreateIndustrySubclassHandler(t *testing.T) {
	db := setupTestDB(t)
	classID := seedIndustrySubclassParent(t, db)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPost, "/industry-subclasses", map[string]interface{}{
		"industry_class_id": classID, "name": "Rice Milling", "code": "01111",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestGetAllIndustrySubclassesHandler(t *testing.T) {
	t.Run("default limit is 20, total reflects all rows", func(t *testing.T) {
		db := setupTestDB(t)
		classID := seedIndustrySubclassParent(t, db)
		for i := 0; i < 5; i++ {
			db.Create(&models.IndustrySubclass{IndustryClassID: classID, Name: "Subclass", Code: "X"})
		}
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/industry-subclasses", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}
		var body map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &body)
		if int(body["limit"].(float64)) != 20 {
			t.Fatalf("expected default limit 20, got %v", body["limit"])
		}
		if int(body["total"].(float64)) != 5 {
			t.Fatalf("expected total 5, got %v", body["total"])
		}
	})

	t.Run("limit query param is respected and capped at 100", func(t *testing.T) {
		db := setupTestDB(t)
		classID := seedIndustrySubclassParent(t, db)
		db.Create(&models.IndustrySubclass{IndustryClassID: classID, Name: "Subclass", Code: "X"})
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/industry-subclasses?limit=500", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}
		var body map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &body)
		if int(body["limit"].(float64)) != 100 {
			t.Fatalf("expected limit capped at 100, got %v", body["limit"])
		}
	})
}

func TestGetIndustrySubclassByIDHandler(t *testing.T) {
	t.Run("non-existent id returns 404", func(t *testing.T) {
		db := setupTestDB(t)
		router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

		w := doRequest(t, router, http.MethodGet, "/industry-subclasses/999999", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
		}
	})
}

func TestUpdateIndustrySubclassHandler(t *testing.T) {
	db := setupTestDB(t)
	classID := seedIndustrySubclassParent(t, db)
	seeded := models.IndustrySubclass{IndustryClassID: classID, Name: "Rice Milling", Code: "01111"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodPut, fmt.Sprintf("/industry-subclasses/%d", seeded.ID), map[string]string{"name": "Paddy Milling"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteIndustrySubclassHandler(t *testing.T) {
	db := setupTestDB(t)
	classID := seedIndustrySubclassParent(t, db)
	seeded := models.IndustrySubclass{IndustryClassID: classID, Name: "Rice Milling", Code: "01111"}
	db.Create(&seeded)
	router := setupRouter(controllers.NewJobController(repositories.NewJobRepository(db), nil))

	w := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/industry-subclasses/%d", seeded.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}