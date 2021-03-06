package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type AddressesResponse struct {
	VertcoinAddress string
	AssetAddress    string
	StealthAddress  string
}

func (h *HttpServer) Addresses(w http.ResponseWriter, r *http.Request) {
	vtcAdr, err := h.wallet.VertcoinAddress()
	if err != nil {
		h.writeError(w, fmt.Errorf("Error fetching VTC Address: %s", err.Error()))
		return
	}

	assAdr, err := h.wallet.AssetsAddress()
	if err != nil {
		h.writeError(w, fmt.Errorf("Error fetching Asset Address: %s", err.Error()))
		return
	}

	stealthAdr, err := h.wallet.StealthAddress()
	if err != nil {
		h.writeError(w, fmt.Errorf("Error fetching Stealth Address: %s", err.Error()))
		return
	}

	resp := AddressesResponse{
		VertcoinAddress: vtcAdr,
		AssetAddress:    assAdr,
		StealthAddress:  stealthAdr,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(resp)
}
