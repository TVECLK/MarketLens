package api

import (
	"marketlens-go-backend/internal/auth"

	"github.com/gin-gonic/gin"
)

type crudHandlers struct {
	Create gin.HandlerFunc
	List   gin.HandlerFunc
	Get    gin.HandlerFunc
	Update gin.HandlerFunc
	Delete gin.HandlerFunc
}

// registerCRUD wires the standard five routes for a flat-lookup or
// hierarchy resource: reads are public, writes need auth + a
// "<scope>:<action>" scope.
func registerCRUD(rg *gin.RouterGroup, path, scope string, h crudHandlers) {
	g := rg.Group(path)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.POST("", auth.AuthRequired(), auth.RequireScope(scope+":create"), h.Create)
	g.PUT("/:id", auth.AuthRequired(), auth.RequireScope(scope+":update"), h.Update)
	g.DELETE("/:id", auth.AuthRequired(), auth.RequireScope(scope+":delete"), h.Delete)
}

// RegisterRoutes attaches every API route and its middleware to r.
func RegisterRoutes(r *gin.Engine, h *Handler) {
	// Kubernetes probes: unversioned, outside /api/v1
	r.GET("/healthz", h.HealthzHandler)
	r.GET("/readyz", h.ReadyzHandler)

	v1 := r.Group("/api/v1")
	{
		// Dashboard endpoints
		v1.GET("/vacancy-trend", h.GetVacancyTrendHandler)
		v1.GET("/vacancy-total", h.GetTotalVacancyCountHandler)
		v1.GET("/occupations/by-date-range", h.GetOccupationJobCountByDateRangeHandler)
		v1.GET("/industries/by-date-range", h.GetIndustryJobCountByDateRangeHandler)
		v1.GET("/:standard/:level/:id/total-job-count", h.GetTotalJobCountByLevelHandler)
		v1.GET("/:standard/:level/:id/children", h.GetLevelChildrenHandler)
		v1.GET("/:standard/:level/:id/employment-sector", h.GetEmploymentSectorByLevelHandler)
		v1.GET("/:standard/:level/:id/experience", h.GetExperienceByLevelHandler)
		v1.GET("/:standard/:level/:id/province", h.GetProvinceByLevelHandler)
		v1.GET("/:standard/:level/:id/education", h.GetEducationLevelByLevelHandler)
		v1.GET("/:standard/:level/:id/formality", h.GetFormalityByLevelHandler)
		v1.GET("/:standard/:level/:id/gender", h.GetGenderByLevelHandler)
		v1.GET("/:standard/:level/:id/vocational-education", h.GetVocationalEducationByLevelHandler)
		v1.GET("/:standard/:level/:id/remote-onsite-hybrid", h.GetRemoteOnSiteHybridByLevelHandler)
		v1.GET("/:standard/:level/:id/job-type", h.GetJobTypeByLevelHandler)
		v1.GET("/occupation/:level/:id/top-15-skills", h.GetTop15SkillsByOccupationLevelHandler)
		v1.GET("/occupation/:level/:id/all-skills", h.GetAllSkillsByOccupationLevelHandler)
		v1.GET("/occupation/:level/:id/top-hiring-employers", h.GetTopHiringEmployersByOccupationLevelHandler)

		// Crawler/ingestion pipeline (authenticated)
		ingestionRoutes := v1.Group("/crawler")
		ingestionRoutes.Use(auth.AuthRequired())
		{
			ingestionRoutes.POST("/runs", auth.RequireScope("crawler:runs"), h.StartCrawlerRunHandler)
			ingestionRoutes.POST("/runs/:id/complete", auth.RequireScope("crawler:complete"), h.CompleteCrawlerRunHandler)
			ingestionRoutes.POST("/jobs/batch-save", auth.RequireScope("crawler:batch-save"), h.BatchSaveJobsHandler)
			ingestionRoutes.POST("/jobs/reconcile", auth.RequireScope("crawler:reconcile"), h.ReconcileStaleVacanciesHandler)
		}

		// Flat lookup tables
		registerCRUD(v1, "/education-levels", "education-levels", crudHandlers{
			Create: h.CreateEducationLevelHandler, List: h.GetAllEducationLevelsHandler, Get: h.GetEducationLevelByIDHandler,
			Update: h.UpdateEducationLevelHandler, Delete: h.DeleteEducationLevelHandler,
		})

		registerCRUD(v1, "/formalities", "formalities", crudHandlers{
			Create: h.CreateFormalityHandler, List: h.GetAllFormalitiesHandler, Get: h.GetFormalityByIDHandler,
			Update: h.UpdateFormalityHandler, Delete: h.DeleteFormalityHandler,
		})

		registerCRUD(v1, "/genders", "genders", crudHandlers{
			Create: h.CreateGenderHandler, List: h.GetAllGendersHandler, Get: h.GetGenderByIDHandler,
			Update: h.UpdateGenderHandler, Delete: h.DeleteGenderHandler,
		})

		registerCRUD(v1, "/employment-sectors", "employment-sectors", crudHandlers{
			Create: h.CreateEmploymentSectorHandler, List: h.GetAllEmploymentSectorsHandler, Get: h.GetEmploymentSectorByIDHandler,
			Update: h.UpdateEmploymentSectorHandler, Delete: h.DeleteEmploymentSectorHandler,
		})

		registerCRUD(v1, "/vocational-educations", "vocational-educations", crudHandlers{
			Create: h.CreateVocationalEducationHandler, List: h.GetAllVocationalEducationsHandler, Get: h.GetVocationalEducationByIDHandler,
			Update: h.UpdateVocationalEducationHandler, Delete: h.DeleteVocationalEducationHandler,
		})

		registerCRUD(v1, "/experiences", "experiences", crudHandlers{
			Create: h.CreateExperienceHandler, List: h.GetAllExperiencesHandler, Get: h.GetExperienceByIDHandler,
			Update: h.UpdateExperienceHandler, Delete: h.DeleteExperienceHandler,
		})

		// Occupation hierarchy
		registerCRUD(v1, "/major-groups", "major-groups", crudHandlers{
			Create: h.CreateMajorGroupHandler, List: h.GetAllMajorGroupsHandler, Get: h.GetMajorGroupByIDHandler,
			Update: h.UpdateMajorGroupHandler, Delete: h.DeleteMajorGroupHandler,
		})
		v1.GET("/major-groups/:id/sub-major-groups", h.GetSubMajorGroupsByMajorGroupHandler)

		registerCRUD(v1, "/sub-major-groups", "sub-major-groups", crudHandlers{
			Create: h.CreateSubMajorGroupHandler, List: h.GetAllSubMajorGroupsHandler, Get: h.GetSubMajorGroupByIDHandler,
			Update: h.UpdateSubMajorGroupHandler, Delete: h.DeleteSubMajorGroupHandler,
		})
		v1.GET("/sub-major-groups/:id/minor-groups", h.GetMinorGroupsBySubMajorGroupHandler)

		registerCRUD(v1, "/minor-groups", "minor-groups", crudHandlers{
			Create: h.CreateMinorGroupHandler, List: h.GetAllMinorGroupsHandler, Get: h.GetMinorGroupByIDHandler,
			Update: h.UpdateMinorGroupHandler, Delete: h.DeleteMinorGroupHandler,
		})
		v1.GET("/minor-groups/:id/unit-groups", h.GetUnitGroupsByMinorGroupHandler)

		registerCRUD(v1, "/unit-groups", "unit-groups", crudHandlers{
			Create: h.CreateUnitGroupHandler, List: h.GetAllUnitGroupsHandler, Get: h.GetUnitGroupByIDHandler,
			Update: h.UpdateUnitGroupHandler, Delete: h.DeleteUnitGroupHandler,
		})
		v1.GET("/unit-groups/:id/occupation-groups", h.GetOccupationGroupsByUnitGroupHandler)

		registerCRUD(v1, "/occupation-groups", "occupation-groups", crudHandlers{
			Create: h.CreateOccupationGroupHandler, List: h.GetAllOccupationGroupsHandler, Get: h.GetOccupationGroupByIDHandler,
			Update: h.UpdateOccupationGroupHandler, Delete: h.DeleteOccupationGroupHandler,
		})

		// Industry hierarchy
		registerCRUD(v1, "/industry-sectors", "industry-sectors", crudHandlers{
			Create: h.CreateIndustrySectorHandler, List: h.GetAllIndustrySectorsHandler, Get: h.GetIndustrySectorByIDHandler,
			Update: h.UpdateIndustrySectorHandler, Delete: h.DeleteIndustrySectorHandler,
		})
		v1.GET("/industry-sectors/:id/industry-divisions", h.GetIndustryDivisionsByIndustrySectorHandler)

		registerCRUD(v1, "/industry-divisions", "industry-divisions", crudHandlers{
			Create: h.CreateIndustryDivisionHandler, List: h.GetAllIndustryDivisionsHandler, Get: h.GetIndustryDivisionByIDHandler,
			Update: h.UpdateIndustryDivisionHandler, Delete: h.DeleteIndustryDivisionHandler,
		})
		v1.GET("/industry-divisions/:id/industry-groups", h.GetIndustryGroupsByIndustryDivisionHandler)

		registerCRUD(v1, "/industry-groups", "industry-groups", crudHandlers{
			Create: h.CreateIndustryGroupHandler, List: h.GetAllIndustryGroupsHandler, Get: h.GetIndustryGroupByIDHandler,
			Update: h.UpdateIndustryGroupHandler, Delete: h.DeleteIndustryGroupHandler,
		})
		v1.GET("/industry-groups/:id/industry-classes", h.GetIndustryClassesByIndustryGroupHandler)

		registerCRUD(v1, "/industry-classes", "industry-classes", crudHandlers{
			Create: h.CreateIndustryClassHandler, List: h.GetAllIndustryClassesHandler, Get: h.GetIndustryClassByIDHandler,
			Update: h.UpdateIndustryClassHandler, Delete: h.DeleteIndustryClassHandler,
		})
		v1.GET("/industry-classes/:id/industry-subclasses", h.GetIndustrySubclassesByIndustryClassHandler)

		registerCRUD(v1, "/industry-subclasses", "industry-sub-classes", crudHandlers{
			Create: h.CreateIndustrySubclassHandler, List: h.GetAllIndustrySubclassesHandler, Get: h.GetIndustrySubclassByIDHandler,
			Update: h.UpdateIndustrySubclassHandler, Delete: h.DeleteIndustrySubclassHandler,
		})
	}
}
