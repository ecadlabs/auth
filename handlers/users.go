package handlers

import (
	"context"
	"encoding/json"
	"git.ecadlabs.com/ecad/auth/users"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/schema"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	UserContextKey = "user"
	DefaultLimit   = 20
)

var schemaDecoder = schema.NewDecoder()

type Users struct {
	BaseURL string
	Storage *users.Storage
	Timeout time.Duration
}

func (u *Users) context(parent context.Context) context.Context {
	if u.Timeout != 0 {
		ctx, _ := context.WithTimeout(parent, u.Timeout)
		return ctx
	}
	return parent
}

// Get current user from the DB and attach to the context
// Is this needed?
func (u *Users) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(TokenContextKey).(*jwt.Token).Claims.(jwt.MapClaims)
		uid, err := uuid.FromString(claims["sub"].(string))
		if err != nil {
			JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		user, err := u.Storage.GetUserByID(u.context(r.Context()), uid)
		if err != nil {
			var status int
			if err == users.ErrNotFound {
				status = http.StatusNotFound
			} else {
				status = http.StatusInternalServerError
			}

			log.Error(err)
			JSONError(w, err.Error(), status)
			return
		}

		newRequest := r.WithContext(context.WithValue(r.Context(), UserContextKey, user))

		next.ServeHTTP(w, newRequest)
	})
}

func (u *Users) GetUsers(w http.ResponseWriter, r *http.Request) {
	// TODO Access control
	var opt users.GetOptions

	if s := r.FormValue("sort"); s != "" {
		opt.SortBy = s
	} else {
		opt.SortBy = users.DefaultSortColumn
	}

	if s := r.FormValue("start"); s != "" {
		opt.Start = s
	}

	if i, err := strconv.ParseInt(r.FormValue("limit"), 10, 64); err == nil {
		opt.Limit = int(i)
	} else {
		opt.Limit = DefaultLimit
	}

	if r.FormValue("order") == "desc" {
		opt.Order = users.SortDesc
	}

	userSlice, err := u.Storage.GetUsers(u.context(r.Context()), &opt)
	if err != nil {
		log.Error(err)
		JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(userSlice) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	nextUrl, err := url.Parse(u.BaseURL)
	if err != nil {
		log.Error(err)
		JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lastItem := userSlice[len(userSlice)-1]
	var start string

	switch opt.SortBy {
	case users.ColumnID:
		start = lastItem.ID.String()
	case users.ColumnEmail:
		start = lastItem.Email
	case users.ColumnName:
		start = lastItem.Name
	case users.ColumnAdded:
		v, _ := lastItem.Added.MarshalText()
		start = string(v)
	case users.ColumnModified:
		v, _ := lastItem.Modified.MarshalText()
		start = string(v)
	}

	// Preserve original
	q := make(url.Values)
	for k, v := range r.Form {
		q[k] = v
	}

	q.Set("start", start)

	nextUrl.RawQuery = q.Encode()

	res := Paginated{
		Value: userSlice,
		Next:  nextUrl.String(),
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

	ret, err := u.Storage.NewUser(u.context(r.Context()), &user)
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

	w.Header().Set("Location", u.BaseURL+"/"+ret.ID.String())
	JSONResponse(w, http.StatusCreated, ret)
}
