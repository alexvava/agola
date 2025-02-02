// Copyright 2019 Sorint.lab
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"encoding/json"
	"net/http"

	csapi "agola.io/agola/internal/services/configstore/api"
	"agola.io/agola/internal/services/gateway/action"
	"agola.io/agola/internal/services/types"
	"agola.io/agola/internal/util"
	"go.uber.org/zap"

	"github.com/gorilla/mux"
)

type SecretResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ParentPath string `json:"parent_path"`
}

func createSecretResponse(s *csapi.Secret) *SecretResponse {
	return &SecretResponse{
		ID:         s.ID,
		Name:       s.Name,
		ParentPath: s.ParentPath,
	}
}

type SecretHandler struct {
	log *zap.SugaredLogger
	ah  *action.ActionHandler
}

func NewSecretHandler(logger *zap.Logger, ah *action.ActionHandler) *SecretHandler {
	return &SecretHandler{log: logger.Sugar(), ah: ah}
}

func (h *SecretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	_, tree := query["tree"]

	parentType, parentRef, err := GetConfigTypeRef(r)
	if httpError(w, err) {
		h.log.Errorf("err: %+v", err)
		return
	}

	areq := &action.GetSecretsRequest{
		ParentType: parentType,
		ParentRef:  parentRef,
		Tree:       tree,
	}
	cssecrets, err := h.ah.GetSecrets(ctx, areq)
	if httpError(w, err) {
		h.log.Errorf("err: %+v", err)
		return
	}

	secrets := make([]*SecretResponse, len(cssecrets))
	for i, s := range cssecrets {
		secrets[i] = createSecretResponse(s)
	}

	if err := httpResponse(w, http.StatusOK, secrets); err != nil {
		h.log.Errorf("err: %+v", err)
	}
}

type CreateSecretRequest struct {
	Name string `json:"name,omitempty"`

	Type types.SecretType `json:"type,omitempty"`

	// internal secret
	Data map[string]string `json:"data,omitempty"`

	// external secret
	SecretProviderID string `json:"secret_provider_id,omitempty"`
	Path             string `json:"path,omitempty"`
}

type CreateSecretHandler struct {
	log *zap.SugaredLogger
	ah  *action.ActionHandler
}

func NewCreateSecretHandler(logger *zap.Logger, ah *action.ActionHandler) *CreateSecretHandler {
	return &CreateSecretHandler{log: logger.Sugar(), ah: ah}
}

func (h *CreateSecretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parentType, parentRef, err := GetConfigTypeRef(r)
	if httpError(w, err) {
		return
	}

	var req CreateSecretRequest
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&req); err != nil {
		httpError(w, util.NewErrBadRequest(err))
		return
	}

	areq := &action.CreateSecretRequest{
		Name:             req.Name,
		ParentType:       parentType,
		ParentRef:        parentRef,
		Type:             req.Type,
		Data:             req.Data,
		SecretProviderID: req.SecretProviderID,
		Path:             req.Path,
	}
	cssecret, err := h.ah.CreateSecret(ctx, areq)
	if httpError(w, err) {
		h.log.Errorf("err: %+v", err)
		return
	}

	res := createSecretResponse(cssecret)
	if err := httpResponse(w, http.StatusCreated, res); err != nil {
		h.log.Errorf("err: %+v", err)
	}
}

type UpdateSecretRequest struct {
	Name string `json:"name,omitempty"`

	Type types.SecretType `json:"type,omitempty"`

	// internal secret
	Data map[string]string `json:"data,omitempty"`

	// external secret
	SecretProviderID string `json:"secret_provider_id,omitempty"`
	Path             string `json:"path,omitempty"`
}

type UpdateSecretHandler struct {
	log *zap.SugaredLogger
	ah  *action.ActionHandler
}

func NewUpdateSecretHandler(logger *zap.Logger, ah *action.ActionHandler) *UpdateSecretHandler {
	return &UpdateSecretHandler{log: logger.Sugar(), ah: ah}
}

func (h *UpdateSecretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	secretName := vars["secretname"]

	parentType, parentRef, err := GetConfigTypeRef(r)
	if httpError(w, err) {
		h.log.Errorf("err: %+v", err)
		return
	}

	var req UpdateSecretRequest
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&req); err != nil {
		httpError(w, util.NewErrBadRequest(err))
		return
	}
	areq := &action.UpdateSecretRequest{
		SecretName: secretName,

		Name:             req.Name,
		ParentType:       parentType,
		ParentRef:        parentRef,
		Type:             req.Type,
		Data:             req.Data,
		SecretProviderID: req.SecretProviderID,
		Path:             req.Path,
	}
	cssecret, err := h.ah.UpdateSecret(ctx, areq)
	if httpError(w, err) {
		h.log.Errorf("err: %+v", err)
		return
	}

	res := createSecretResponse(cssecret)
	if err := httpResponse(w, http.StatusOK, res); err != nil {
		h.log.Errorf("err: %+v", err)
	}
}

type DeleteSecretHandler struct {
	log *zap.SugaredLogger
	ah  *action.ActionHandler
}

func NewDeleteSecretHandler(logger *zap.Logger, ah *action.ActionHandler) *DeleteSecretHandler {
	return &DeleteSecretHandler{log: logger.Sugar(), ah: ah}
}

func (h *DeleteSecretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	secretName := vars["secretname"]

	parentType, parentRef, err := GetConfigTypeRef(r)
	if httpError(w, err) {
		return
	}

	err = h.ah.DeleteSecret(ctx, parentType, parentRef, secretName)
	if httpError(w, err) {
		h.log.Errorf("err: %+v", err)
		return
	}

	if err := httpResponse(w, http.StatusNoContent, nil); err != nil {
		h.log.Errorf("err: %+v", err)
	}
}
