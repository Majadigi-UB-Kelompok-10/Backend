package handlers

import "github.com/Majadigi-UB-Kelompok-10/api-gateway/internal/db"

type EndpointHandler struct {
	Queries *db.Queries
}

func NewEndpointHandler(q *db.Queries) *EndpointHandler {
	return &EndpointHandler{Queries: q}
}
