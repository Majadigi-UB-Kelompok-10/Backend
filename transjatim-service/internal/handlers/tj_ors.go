package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/farildzaky/transjatim-service/internal/db"
)

const orsDirectionsURL = "https://api.openrouteservice.org/v2/directions/driving-car/geojson"

var orsHTTPClient = &http.Client{Timeout: 15 * time.Second}

type orsRequest struct {
	Coordinates [][2]float64 `json:"coordinates"`
}

type orsGeoJSONResponse struct {
	Features []struct {
		Geometry struct {
			Coordinates [][2]float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}


func fetchRouteGeometry(ctx context.Context, stops []db.GetRuteStopsWithCoordsRow, apiKey string) ([]byte, error) {
	coords := make([][2]float64, 0, len(stops))
	for _, s := range stops {
		if !s.Lat.Valid || !s.Lng.Valid {
			slog.Warn("tj.ors.skip_stop_no_coords", slog.Int("terminal_id", int(s.TerminalID)))
			continue
		}
		coords = append(coords, [2]float64{s.Lng.Float64, s.Lat.Float64})
	}

	if len(coords) < 2 {
		return nil, fmt.Errorf("kurang dari 2 titik koordinat valid untuk generate rute")
	}

	reqBody, err := json.Marshal(orsRequest{Coordinates: coords})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, orsDirectionsURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, application/geo+json")

	resp, err := orsHTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ors request gagal: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ors status %d: %s", resp.StatusCode, string(respBody))
	}

	var orsResp orsGeoJSONResponse
	if err := json.Unmarshal(respBody, &orsResp); err != nil {
		return nil, fmt.Errorf("parse ors response: %w", err)
	}
	if len(orsResp.Features) == 0 || len(orsResp.Features[0].Geometry.Coordinates) == 0 {
		return nil, fmt.Errorf("ors response tidak memiliki koordinat rute")
	}

	result, err := json.Marshal(orsResp.Features[0].Geometry.Coordinates)
	if err != nil {
		return nil, err
	}
	return result, nil
}
