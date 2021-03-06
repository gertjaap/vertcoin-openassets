package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gertjaap/vertcoin-extras/util"
	"github.com/gertjaap/vertcoin-extras/wallet"

	"github.com/mit-dci/lit/bech32"
)

type TransferAssetParameters struct {
	AssetID          string
	Amount           uint64
	RecipientAddress string
	UseStealth       bool
}

type TransferAssetResult struct {
	TxID string
}

func (h *HttpServer) TransferAsset(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var params TransferAssetParameters
	err := decoder.Decode(&params)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding json: %s", err.Error()))
		return
	}

	if params.AssetID == "VTC" {
		h.TransferVTC(w, r, params)
		return
	} else if params.AssetID == "SVTC" {
		h.TransferVTCStealth(w, r, params)
		return
	}

	assetID, err := hex.DecodeString(params.AssetID)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding asset ID: %s", err.Error()))
		return
	}

	var recipientPkh [20]byte
	decoded, err := bech32.SegWitAddressDecode(params.RecipientAddress)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding recipient address: %s", err.Error()))
		return
	}
	copy(recipientPkh[:], decoded[2:]) // skip version and pushdata byte returned by SegWitAddressDecode

	var tx wallet.OpenAssetTransaction
	tx.Transfers = append(tx.Transfers, wallet.OpenAssetTransferOutput{
		AssetID:      assetID,
		Value:        params.Amount,
		RecipientPkh: recipientPkh,
	})

	wireTx, err := h.wallet.GenerateOpenAssetTx(tx)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error generating transaction: %s", err.Error()))
		return
	}

	err = h.wallet.SignMyInputs(wireTx)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error signing: %s", err.Error()))
		return
	}

	txid, err := h.wallet.SendTransaction(wireTx)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error sending: %s", err.Error()))
		return
	}

	var res NewAssetResult
	res.TxID = txid.String()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}

func (h *HttpServer) TransferVTC(w http.ResponseWriter, r *http.Request, params TransferAssetParameters) {
	decoded, err := bech32.SegWitAddressDecode(params.RecipientAddress)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding recipient address: %s", err.Error()))
		return
	}

	var tx wallet.SendTransaction
	tx.Amount = params.Amount
	copy(tx.RecipientPkh[:], decoded[2:]) // skip version and pushdata byte returned by SegWitAddressDecode

	wireTx, err := h.wallet.GenerateNormalSendTx(tx, params.UseStealth)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error generating transaction: %s", err.Error()))
		return
	}

	err = h.wallet.SignMyInputs(wireTx)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error signing: %s", err.Error()))
		return
	}

	txid, err := h.wallet.SendTransaction(wireTx)
	if err != nil {
		util.PrintTx(wireTx)
		h.writeError(w, fmt.Errorf("Error sending: %s", err.Error()))
		return
	}

	var res NewAssetResult
	res.TxID = txid.String()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}

func (h *HttpServer) TransferVTCStealth(w http.ResponseWriter, r *http.Request, params TransferAssetParameters) {
	prefix, pubKey, err := bech32.Decode(params.RecipientAddress)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding recipient address: %s", err.Error()))
		return
	}
	if prefix != h.config.Network.StealthAddressPrefix {
		h.writeError(w, fmt.Errorf("Address has wrong prefix %s, expected %s", prefix, h.config.Network.StealthAddressPrefix))
		return
	}
	if len(pubKey) != 33 {
		h.writeError(w, fmt.Errorf("Address has incorrect byte length %d, expected 33", len(pubKey)))
		return
	}

	tx := wallet.StealthTransaction{
		Amount: params.Amount,
	}
	copy(tx.RecipientPubKey[:], pubKey)

	wireTx, err := h.wallet.GenerateStealthTx(tx, params.UseStealth)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error decoding recipient address: %s", err.Error()))
		return
	}

	err = h.wallet.SignMyInputs(wireTx)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error signing: %s", err.Error()))
		return
	}

	util.PrintTx(wireTx)

	txid, err := h.wallet.SendTransaction(wireTx)
	if err != nil {
		h.writeError(w, fmt.Errorf("Error sending: %s", err.Error()))
		return
	}

	var res NewAssetResult
	res.TxID = txid.String()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}
