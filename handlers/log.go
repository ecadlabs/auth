package handlers

import (
	"net/http"
	"net/url"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/middleware"
	"github.com/ecadlabs/auth/query"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
	log "github.com/sirupsen/logrus"
)

//GetLogs Endpoint handler to get list of logs
func (u *Users) GetLogs(w http.ResponseWriter, r *http.Request) {
	site := r.Context().Value(middleware.DomainConfigContextKey).(*middleware.DomainConfigData)
	r.ParseForm()
	member := r.Context().Value(middleware.MembershipContextKey).(*storage.Membership)

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	granted, err := role.IsAnyGranted(permissionLogs, permissionFull)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if !granted {
		utils.JSONErrorResponse(w, errors.ErrForbidden)
		return
	}

	q, err := query.FromValues(r.Form)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeQuerySyntax)
		return
	}

	if q.Limit <= 0 {
		q.Limit = DefaultLimit
	}

	logSlice, count, nextQuery, err := u.Storage.GetLogs(ctx, q)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if len(logSlice) == 0 && !q.TotalCount {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Pagination
	res := utils.Paginated{
		Value: logSlice,
	}

	if q.TotalCount {
		res.TotalCount = &count
	}

	if nextQuery != nil {
		nextURL, err := url.Parse(u.LogURL(site))
		if err != nil {
			log.Error(err)
			utils.JSONErrorResponse(w, err)
			return
		}

		nextURL.RawQuery = nextQuery.Values().Encode()
		res.Next = nextURL.String()
	}

	utils.JSONResponse(w, http.StatusOK, &res)
}
