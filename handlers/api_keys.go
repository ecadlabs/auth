package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/rbac"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type APIKeys struct {
	Storage  storage.APIKeyStorage
	Timeout  time.Duration
	Enforcer rbac.Enforcer

	AuxLogger *log.Logger
}

func (a *APIKeys) context(r *http.Request) (context.Context, context.CancelFunc) {
	if a.Timeout != 0 {
		return context.WithTimeout(r.Context(), a.Timeout)
	}
	return r.Context(), func() {}
}

func (a *APIKeys) checkWritePermissions(role rbac.Role, self bool) error {
	perm := []string{permissionServiceFull, permissionServiceWrite}
	if self {
		perm = append(perm, permissionWriteSelf)
	}

	granted, err := role.IsAnyGranted(perm...)
	if err != nil {
		return err
	}

	if !granted {
		return errors.ErrForbidden
	}

	return nil
}

func (a *APIKeys) checkReadPermissions(role rbac.Role, self bool) error {
	perm := []string{permissionServiceFull, permissionServiceRead}
	if self {
		perm = append(perm, permissionReadSelf)
	}

	granted, err := role.IsAnyGranted(perm...)
	if err != nil {
		return err
	}

	if !granted {
		return errors.ErrForbidden
	}

	return nil
}

type newKey struct {
	TenantID uuid.UUID `json:"tenant_id"`
}

func (a *APIKeys) NewAPIKey(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey{}).(*storage.User)
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)

	uid, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	var req newKey
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ctx, cancel := a.context(r)
	defer cancel()

	role, err := a.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if err = a.checkWritePermissions(role, self.ID == uid); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	key, err := a.Storage.NewKey(ctx, uid, req.TenantID)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	a.AuxLogger.WithFields(logFields(EvDeleteAPIKey, self.ID, uid, r)).WithFields(log.Fields{"key_id": key.ID, "tenant_id": key.TenantID}).Printf("User %v issued API key for service account %v in tenant in tenant %v", self.ID, uid, key.TenantID)

	utils.JSONResponse(w, http.StatusCreated, key)
}

func (a *APIKeys) GetAPIKey(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey{}).(*storage.User)
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)

	uid, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	kid, err := uuid.FromString(mux.Vars(r)["keyId"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ctx, cancel := a.context(r)
	defer cancel()

	role, err := a.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if err = a.checkReadPermissions(role, self.ID == uid); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	key, err := a.Storage.GetKey(ctx, kid, uid)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	utils.JSONResponse(w, http.StatusOK, key)
}

func (a *APIKeys) GetAPIKeys(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey{}).(*storage.User)
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)

	uid, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ctx, cancel := a.context(r)
	defer cancel()

	role, err := a.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if err = a.checkReadPermissions(role, self.ID == uid); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	keys, err := a.Storage.GetKeys(ctx, uid)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if len(keys) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	utils.JSONResponse(w, http.StatusOK, keys)
}

func (a *APIKeys) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey{}).(*storage.User)
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)

	uid, err := uuid.FromString(mux.Vars(r)["userId"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	kid, err := uuid.FromString(mux.Vars(r)["keyId"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ctx, cancel := a.context(r)
	defer cancel()

	role, err := a.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if err = a.checkWritePermissions(role, self.ID == uid); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	if err = a.Storage.DeleteKey(ctx, kid, uid); err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	a.AuxLogger.WithFields(logFields(EvDeleteAPIKey, self.ID, uid, r)).WithField("key_id", kid).Printf("User %v removed API key for service account %v", self.ID, uid)

	w.WriteHeader(http.StatusNoContent)
}
