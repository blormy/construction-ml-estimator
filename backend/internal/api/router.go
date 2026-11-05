package httpapi

import (
	"constructmaterialai/internal/auth"
	"net/http"
)

// регистрация всех эндпоинтов API
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/default", DefaultHandler)
	mux.HandleFunc("/excel", ExcelHandler)

	mux.HandleFunc("/pump/estimate", PumpHandler)
	mux.HandleFunc("/drum/estimate", DrumHandler)
	mux.HandleFunc("/vessel/estimate", VesselHandler)
	mux.HandleFunc("/conveyor/estimate", ConveyorHandler)
	mux.HandleFunc("/utube/estimate", UTubeHandler)
	mux.HandleFunc("/tower/estimate", TowerHandler)
	mux.HandleFunc("/box_furnace/estimate", BoxFurnaceHandler)
	mux.HandleFunc("/centrifugal_compressor/estimate", CentrifugalCompressorHandler)

	// эндпоинты авторизации
	mux.HandleFunc("/api/auth/register", RegisterHandler)
	mux.HandleFunc("/api/auth/login", LoginHandler)

	// эндпоинты проектов (защищенные)
	mux.Handle("/api/projects", auth.AuthMiddleware(http.HandlerFunc(ProjectsHandler)))
	mux.Handle("/api/projects/", auth.AuthMiddleware(http.HandlerFunc(ProjectByIDHandler)))

	// эндпоинты команд (защищенные)
	mux.Handle("/api/teams", auth.AuthMiddleware(http.HandlerFunc(TeamsHandler)))
	mux.Handle("/api/teams/", auth.AuthMiddleware(http.HandlerFunc(TeamsSubHandler)))
}
