package handlers

import (
	"net/http"
	"net/url"

	"git.ecadlabs.com/ecad/auth/errors"
	"git.ecadlabs.com/ecad/auth/query"
	"git.ecadlabs.com/ecad/auth/storage"
	"git.ecadlabs.com/ecad/auth/utils"
	log "github.com/sirupsen/logrus"
)

func (u *Users) GetLogs(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	self := r.Context().Value(UserContextKey).(*storage.User)

	if err := self.Roles.Get().IsGranted(permissionLog, nil); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeForbidden)
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

	logSlice, count, nextQuery, err := u.Storage.GetLogs(u.context(r), q)
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
		nextUrl, err := url.Parse(u.LogURL())
		if err != nil {
			log.Error(err)
			utils.JSONErrorResponse(w, err)
			return
		}

		nextUrl.RawQuery = nextQuery.Values().Encode()
		res.Next = nextUrl.String()
	}

	utils.JSONResponse(w, http.StatusOK, &res)
}
