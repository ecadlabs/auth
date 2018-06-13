package handlers

import (
	"context"
	"encoding/json"
	"git.ecadlabs.com/ecad/auth/jsonpatch"
	"git.ecadlabs.com/ecad/auth/query"
	"git.ecadlabs.com/ecad/auth/roles"
	"git.ecadlabs.com/ecad/auth/users"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	UserContextKey = "user"
	DefaultLimit   = 20
)

var schemaDecoder = schema.NewDecoder()

type Users struct {
	BaseURL   func() string
	Namespace string
	Storage   *users.Storage
	Timeout   time.Duration
}

func (u *Users) context(r *http.Request) context.Context {
	if u.Timeout != 0 {
		ctx, _ := context.WithTimeout(r.Context(), u.Timeout)
		return ctx
	}
	return r.Context()
}

func errorHTTPStatus(err error) int {
	if e, ok := err.(*users.Error); ok {
		return e.HTTPStatus
	}

	return http.StatusInternalServerError
}

func (u *Users) getTokenData(r *http.Request) (uid uuid.UUID, ret roles.Roles) {
	if token, ok := r.Context().Value(TokenContextKey).(*jwt.Token); ok {
		claims := token.Claims.(jwt.MapClaims)
		ns := u.Namespace
		if ns == "" {
			ns = DefaultNamespace
		}

		if n, ok := claims[nsClaim(ns, "roles")].([]interface{}); ok {
			names := make([]string, 0, len(n))
			for _, name := range n {
				if s, ok := name.(string); ok {
					names = append(names, s)
				}
			}

			if len(names) != 0 {
				ret = roles.GetKnownRoles(names)
			}
		}

		if sub, ok := claims["sub"].(string); ok {
			if id, err := uuid.FromString(sub); err == nil {
				uid = id
			}
		}
	}

	if len(ret) == 0 {
		ret = roles.Roles{roles.GetRole(RoleAnonymous)}
	}

	return
}

func (u *Users) GetUser(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	self, userRoles := u.getTokenData(r)
	if err = userRoles.IsGranted(permissionGet, map[string]interface{}{
		"self": self,
		"id":   uid,
	}); err != nil {
		log.Error(err)
		JSONError(w, err.Error(), http.StatusForbidden)
		return
	}

	user, err := u.Storage.GetUserByID(u.context(r), uid)
	if err != nil {
		log.Error(err)
		JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	JSONResponse(w, http.StatusOK, user)
}

func (u *Users) GetUsers(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	_, userRoles := u.getTokenData(r)

	if err := userRoles.IsGranted(permissionList, nil); err != nil {
		log.Error(err)
		JSONError(w, err.Error(), http.StatusForbidden)
		return
	}

	q, err := query.FromValues(r.Form)
	if err != nil {
		log.Error(err)
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if q.Limit <= 0 {
		q.Limit = DefaultLimit
	}

	userSlice, count, nextQuery, err := u.Storage.GetUsers(u.context(r), q)
	if err != nil {
		log.Error(err)
		JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	if len(userSlice) == 0 && !q.TotalCount {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Pagination
	res := Paginated{
		Value: userSlice,
	}

	if q.TotalCount {
		res.TotalCount = &count
	}

	if nextQuery != nil {
		nextUrl, err := url.Parse(u.BaseURL())
		if err != nil {
			log.Error(err)
			JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nextUrl.RawQuery = nextQuery.Values().Encode()
		res.Next = nextUrl.String()
	}

	JSONResponse(w, http.StatusOK, &res)
}

func (u *Users) NewUser(w http.ResponseWriter, r *http.Request) {
	// TODO Email confirmation
	// TODO Access control

	var user users.User

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			log.Error(err)
			JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		r.ParseForm()

		if err := schemaDecoder.Decode(&user, r.PostForm); err != nil {
			log.Error(err)
			JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if user.Email == "" || user.Password == "" {
		JSONError(w, "Email and password must not be empty", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error(err)
		JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user.PasswordHash = hash

	if !user.Roles.HasPrefix(RolePrefix) {
		user.Roles.Add(RoleRegular)
	}

	_, userRoles := u.getTokenData(r)
	if err = userRoles.IsGranted(permissionCreate, map[string]interface{}{"user": &user}); err != nil {
		log.Error(err)
		JSONError(w, err.Error(), http.StatusForbidden)
		return
	}

	ret, err := u.Storage.NewUser(u.context(r), &user)
	if err != nil {
		log.Error(err)

		var code int
		if err == users.ErrEmail {
			code = http.StatusConflict
		} else {
			code = http.StatusInternalServerError
		}
		JSONError(w, err.Error(), code)
		return
	}

	w.Header().Set("Location", u.BaseURL()+ret.ID.String())
	JSONResponse(w, http.StatusCreated, ret)
}

func (u *Users) PatchUser(w http.ResponseWriter, r *http.Request) {
	// TODO Email verification

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	self, userRoles := u.getTokenData(r)
	if err = userRoles.IsGranted(permissionModify, map[string]interface{}{
		"self": self,
		"id":   uid,
	}); err != nil {
		log.Error(err)
		JSONError(w, err.Error(), http.StatusForbidden)
		return
	}

	var p jsonpatch.Patch

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Error(err)
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, op := range p {
		if strings.HasPrefix(op.Path, "/roles/") {
			role := strings.TrimPrefix(op.Path, "/roles/")

			switch op.Op {
			case "add":
				err = userRoles.IsGranted(permissionAddRole, map[string]interface{}{"role": role})
			case "remove":
				err = userRoles.IsGranted(permissionDeleteRole, map[string]interface{}{"role": role})
			}

			if err != nil {
				log.Error(err)
				JSONError(w, err.Error(), http.StatusForbidden)
				return
			}
		}
	}

	user, err := u.Storage.PatchUser(u.context(r), uid, p)
	if err != nil {
		log.Error(err)
		JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	JSONResponse(w, http.StatusOK, user)
}

func (u *Users) DeleteUser(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	self, userRoles := u.getTokenData(r)
	if err = userRoles.IsGranted(permissionDelete, map[string]interface{}{
		"self": self,
		"id":   uid,
	}); err != nil {
		log.Error(err)
		JSONError(w, err.Error(), http.StatusForbidden)
		return
	}

	if err := u.Storage.DeleteUser(u.context(r), uid); err != nil {
		log.Error(err)
		JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
