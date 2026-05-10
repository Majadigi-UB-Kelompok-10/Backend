package routes

import (
	"context"
	"fmt"
	"sync"

	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"
	catHandler "github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers/category_list"
	endHandler "github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers/endpoint_list"
	imgHandler "github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers/image_list"
	intHandler "github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers/integration_list"
	opHandler "github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers/operational_list"
	polHandler "github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers/policy_list"
	shcHandler "github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers/services_has_categories"
	svcHandler "github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/handlers/services_list"
	"github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/proxy"
)

// ---------------------------------------------------------------------------
// Gateway Registry
// ---------------------------------------------------------------------------

type RouteRegistry struct {
	sync.RWMutex
	Routes map[string]string
}

var GatewayRegistry = &RouteRegistry{
	Routes: make(map[string]string),
}

func RefreshEndpoints(ctx context.Context, queries *db.Queries) error {
	dbRoutes, err := queries.ListEndpoints(ctx)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	newRoutes := make(map[string]string)
	for _, r := range dbRoutes {
		newRoutes[r.SlugName] = r.PageUrl
	}

	GatewayRegistry.Lock()
	GatewayRegistry.Routes = newRoutes
	GatewayRegistry.Unlock()

	fmt.Printf("RAM Synced: %d rute dinamis siap digunakan\n", len(newRoutes))
	return nil
}

// ---------------------------------------------------------------------------
// SetupDynamicEndpoint — slug-based reverse-proxy catch-all.
// Must be registered LAST so that the specific public and admin routes
// declared above take priority over the /:slug/* wildcard.
// ---------------------------------------------------------------------------

func SetupDynamicEndpoint(app *fiber.App) {
	api := app.Group("/api/v1")

	api.All("/:slug/*", func(c fiber.Ctx) error {
		slug := c.Params("slug")

		GatewayRegistry.RLock()
		targetUrl, exists := GatewayRegistry.Routes[slug]
		GatewayRegistry.RUnlock()

		if !exists {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Endpoint layanan tidak ditemukan.",
			})
		}

		extraPath := c.Params("*")
        if extraPath != "" {
            targetUrl = targetUrl + "/" + extraPath
        }

        queryString := string(c.Request().URI().QueryString())
        if queryString != "" {
            targetUrl = targetUrl + "?" + queryString
        }

        if err := proxy.Do(c, targetUrl); err != nil {
            fmt.Printf("Proxy error untuk slug %s ke %s: %v\n", slug, targetUrl, err)
            return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
                "error":   "Bad Gateway",
                "message": "Layanan sedang mengalami gangguan.",
            })
        }

		return nil
	})
}

// ---------------------------------------------------------------------------
// SetupPublicRoutes — all GET (read-only) routes at /api/v1.
// Must be registered BEFORE SetupDynamicEndpoint so the specific static
// segments take precedence over the /:slug/* catch-all.
// ---------------------------------------------------------------------------

func SetupPublicRoutes(app *fiber.App, q *db.Queries) {
	pub := app.Group("/api/v1")

	setupPublicCategoryRoutes(pub, q)
	setupPublicEndpointRoutes(pub, q)
	setupPublicServiceRoutes(pub, q)
	setupPublicImageRoutes(pub, q)
	setupPublicIntegrationRoutes(pub, q)
	setupPublicOperationalRoutes(pub, q)
	setupPublicPolicyRoutes(pub, q)
	setupPublicServicesCategoriesRoutes(pub, q)
}

// ---------------------------------------------------------------------------
// SetupAdminRoutes — all write (POST / PUT / DELETE) routes at /api/v1/admin.
// ---------------------------------------------------------------------------

func SetupAdminRoutes(app *fiber.App, q *db.Queries) {
	admin := app.Group("/api/v1/admin", middleware.RequireAuth(), middleware.RequireAdmin)

	setupAdminCategoryRoutes(admin, q)
	setupAdminEndpointRoutes(admin, q)
	setupAdminServiceRoutes(admin, q)
	setupAdminImageRoutes(admin, q)
	setupAdminIntegrationRoutes(admin, q)
	setupAdminOperationalRoutes(admin, q)
	setupAdminPolicyRoutes(admin, q)
}

// ===========================================================================
// PUBLIC route helpers  (GET only)
// ===========================================================================

// ---------------------------------------------------------------------------
// GET  /api/v1/categories
// ---------------------------------------------------------------------------

func setupPublicCategoryRoutes(pub fiber.Router, q *db.Queries) {
	h := catHandler.NewCategoryHandler(q)
	shc := shcHandler.NewServicesCategoriesHandler(q)

	cats := pub.Group("/categories")

	cats.Get("/", h.ListCategories)

	byID := cats.Group("/:id")
	byID.Get("/", h.GetCategoryById)
	byID.Get("/services", shc.ListServicesByCategoryId)
}

// ---------------------------------------------------------------------------
// GET  /api/v1/endpoints
// ---------------------------------------------------------------------------

func setupPublicEndpointRoutes(pub fiber.Router, q *db.Queries) {
	h := endHandler.NewEndpointHandler(q)

	eps := pub.Group("/endpoints")

	eps.Get("/", h.GetEndpointList)
	eps.Get("/all", h.GetAllEndpoints)
	// Static segment before /:id
	eps.Get("/slug/:slug", h.GetEndpointBySlug)
	eps.Get("/:id", h.GetEndpointById)
}

// ---------------------------------------------------------------------------
// GET  /api/v1/services
// ---------------------------------------------------------------------------

func setupPublicServiceRoutes(pub fiber.Router, q *db.Queries) {
	h := svcHandler.NewServiceHandler(q)
	shc := shcHandler.NewServicesCategoriesHandler(q)

	svcs := pub.Group("/services")

	svcs.Get("/", h.ListServices)
	svcs.Get("/personalized", shc.GetPersonalizedServices)

	byID := svcs.Group("/:id")
	byID.Get("/", h.GetServiceById)
	byID.Get("/categories", shc.ListCategoriesByServiceId)
}

// ---------------------------------------------------------------------------
// GET  /api/v1/images
// ---------------------------------------------------------------------------

func setupPublicImageRoutes(pub fiber.Router, q *db.Queries) {
	h := imgHandler.NewImageHandler(q)

	imgs := pub.Group("/images")

	imgs.Get("/", h.GetAllImages)
	// Static segment before /:id
	imgs.Get("/service/:service_id", h.ListImagesByServiceId)
	imgs.Get("/:id", h.GetImageById)
}

// ---------------------------------------------------------------------------
// GET  /api/v1/integrations
// ---------------------------------------------------------------------------

func setupPublicIntegrationRoutes(pub fiber.Router, q *db.Queries) {
	h := intHandler.NewIntegrationHandler(q)

	ints := pub.Group("/integrations")

	ints.Get("/", h.GetAllIntegrations)
	// Static segment before /:id
	ints.Get("/service/:service_id", h.ListIntegrationsByServiceId)
	ints.Get("/:id", h.GetIntegrationById)
}

// ---------------------------------------------------------------------------
// GET  /api/v1/operational
// ---------------------------------------------------------------------------

func setupPublicOperationalRoutes(pub fiber.Router, q *db.Queries) {
	h := opHandler.NewOperationalHandler(q)

	op := pub.Group("/operational")
	op.Get("/", h.GetAllOperationals)
	op.Get("/service/:service_id", h.GetOperationalByServiceId)
}

// ---------------------------------------------------------------------------
// GET  /api/v1/policies
// ---------------------------------------------------------------------------

func setupPublicPolicyRoutes(pub fiber.Router, q *db.Queries) {
	h := polHandler.NewPolicyHandler(q)

	pol := pub.Group("/policies")
	pol.Get("/", h.GetAllPolicies)
	pol.Get("/service/:service_id", h.GetPolicyByServiceId)
}

// ---------------------------------------------------------------------------
// GET  /api/v1/services-has-categories
// ---------------------------------------------------------------------------

func setupPublicServicesCategoriesRoutes(pub fiber.Router, q *db.Queries) {
	shc := shcHandler.NewServicesCategoriesHandler(q)

	sc := pub.Group("/services-has-categories")
	sc.Get("/", shc.GetAllServicesHasCategories)
	sc.Get("/normalized", shc.GetAllNormalizedServiceCategory)
	sc.Get("/normalized/:id", shc.GetNormalizedServiceCategoryById)
}

// ===========================================================================
// ADMIN route helpers  (POST / PUT / DELETE only)
// ===========================================================================

// ---------------------------------------------------------------------------
// POST/PUT/DELETE  /api/v1/admin/categories
// ---------------------------------------------------------------------------

func setupAdminCategoryRoutes(admin fiber.Router, q *db.Queries) {
	h := catHandler.NewCategoryHandler(q)

	cats := admin.Group("/categories")
	cats.Post("/", h.CreateCategory)

	byID := cats.Group("/:id")
	byID.Put("/", h.UpdateCategory)
	byID.Delete("/", h.DeleteCategory)
}

// ---------------------------------------------------------------------------
// POST/PUT/DELETE  /api/v1/admin/endpoints
// ---------------------------------------------------------------------------

func setupAdminEndpointRoutes(admin fiber.Router, q *db.Queries) {
	h := endHandler.NewEndpointHandler(q)

	eps := admin.Group("/endpoints")
	eps.Post("/", h.CreateEndpoint)

	byID := eps.Group("/:id")
	byID.Put("/", h.UpdateEndpoint)
	byID.Delete("/", h.DeleteEndpoint)
}

// ---------------------------------------------------------------------------
// POST/PUT/DELETE  /api/v1/admin/services
// ---------------------------------------------------------------------------

func setupAdminServiceRoutes(admin fiber.Router, q *db.Queries) {
	h := svcHandler.NewServiceHandler(q)
	shc := shcHandler.NewServicesCategoriesHandler(q)

	svcs := admin.Group("/services")
	svcs.Post("/", h.CreateService)

	byID := svcs.Group("/:id")
	byID.Put("/", h.UpdateService)
	byID.Delete("/", h.DeleteService)
	byID.Post("/categories", shc.AddCategoryToService)
	byID.Delete("/categories/:category_id", shc.RemoveCategoryFromService)
}

// ---------------------------------------------------------------------------
// POST/PUT/DELETE  /api/v1/admin/images
// ---------------------------------------------------------------------------

func setupAdminImageRoutes(admin fiber.Router, q *db.Queries) {
	h := imgHandler.NewImageHandler(q)

	imgs := admin.Group("/images")
	imgs.Post("/", h.CreateImage)

	byID := imgs.Group("/:id")
	byID.Put("/", h.UpdateImage)
	byID.Delete("/", h.DeleteImage)
}

// ---------------------------------------------------------------------------
// POST/PUT/DELETE  /api/v1/admin/integrations
// ---------------------------------------------------------------------------

func setupAdminIntegrationRoutes(admin fiber.Router, q *db.Queries) {
	h := intHandler.NewIntegrationHandler(q)

	ints := admin.Group("/integrations")
	ints.Post("/", h.CreateIntegration)

	byID := ints.Group("/:id")
	byID.Put("/", h.UpdateIntegration)
	byID.Delete("/", h.DeleteIntegration)
}

// ---------------------------------------------------------------------------
// POST/PUT/DELETE  /api/v1/admin/operational
// ---------------------------------------------------------------------------

func setupAdminOperationalRoutes(admin fiber.Router, q *db.Queries) {
	h := opHandler.NewOperationalHandler(q)

	op := admin.Group("/operational")
	op.Post("/", h.CreateOperational)
	op.Put("/service/:service_id", h.UpdateOperational)
	op.Delete("/service/:service_id", h.DeleteOperational)
}

// ---------------------------------------------------------------------------
// POST/PUT/DELETE  /api/v1/admin/policies
// ---------------------------------------------------------------------------

func setupAdminPolicyRoutes(admin fiber.Router, q *db.Queries) {
	h := polHandler.NewPolicyHandler(q)

	pol := admin.Group("/policies")
	pol.Post("/", h.CreatePolicy)
	pol.Put("/service/:service_id", h.UpdatePolicy)
	pol.Delete("/service/:service_id", h.DeletePolicy)
}
