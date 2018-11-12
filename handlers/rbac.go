package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/ecadlabs/auth/rbac"
	"github.com/ecadlabs/auth/utils"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type RolesHandler struct {
	DB      rbac.RoleDB
	Timeout time.Duration
}

func (r *RolesHandler) context(req *http.Request) context.Context {
	if r.Timeout != 0 {
		ctx, _ := context.WithTimeout(req.Context(), r.Timeout)
		return ctx
	}
	return req.Context()
}

func (r *RolesHandler) GetRoles(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	perm := req.Form["perm"]

	desc, err := r.DB.GetRolesDesc(r.context(req), perm...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if len(desc) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	utils.JSONResponse(w, http.StatusOK, desc)
}

func (r *RolesHandler) GetRole(w http.ResponseWriter, req *http.Request) {
	role := mux.Vars(req)["id"]

	desc, err := r.DB.GetRoleDesc(r.context(req), role)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	utils.JSONResponse(w, http.StatusOK, desc)
}

func (r *RolesHandler) GetPermissions(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	roles := req.Form["role"]

	desc, err := r.DB.GetPermissionsDesc(r.context(req), roles...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if len(desc) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	utils.JSONResponse(w, http.StatusOK, desc)
}

func (r *RolesHandler) GetPermission(w http.ResponseWriter, req *http.Request) {
	perm := mux.Vars(req)["id"]

	desc, err := r.DB.GetPermissionDesc(r.context(req), perm)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	utils.JSONResponse(w, http.StatusOK, desc)
}
