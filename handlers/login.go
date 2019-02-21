package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/rbac"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	//TokenContextKey Context value key for request token
	TokenContextKey = "token"
)

type userTokenOptions struct {
	addr          string
	user          *storage.User
	membership    *storage.Membership
	key           *storage.APIKey
	role          rbac.Role
	sessionMaxAge time.Duration
	refresh       string
}

func (u *Users) writeUserToken(w http.ResponseWriter, opt *userTokenOptions) error {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": opt.user.ID,
		"iat": now.Unix(),
		"iss": u.BaseURL(),
		"aud": u.BaseURL(),
	}

	if opt.sessionMaxAge != 0 {
		claims["exp"] = now.Add(opt.sessionMaxAge).Unix()
	}

	if opt.key != nil {
		claims[utils.NSClaim(u.Namespace, "api_key")] = opt.key.ID
	}

	if opt.addr != "" {
		claims[utils.NSClaim(u.Namespace, "address")] = opt.addr
	}

	if opt.user.Email != "" {
		claims[utils.NSClaim(u.Namespace, "email")] = opt.user.Email
	}

	if opt.user.Name != "" {
		claims[utils.NSClaim(u.Namespace, "name")] = opt.user.Name
	}

	if opt.membership != nil {
		claims[utils.NSClaim(u.Namespace, "tenant")] = opt.membership.TenantID
		claims[utils.NSClaim(u.Namespace, "roles")] = opt.membership.Roles.Get()
	}

	if opt.role != nil {
		claims[utils.NSClaim(u.Namespace, "permissions")] = opt.role.Permissions()
	}

	token := jwt.NewWithClaims(u.JWTSigningMethod, claims)
	secret, err := u.JWTSecretGetter()
	if err != nil {
		return err
	}

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return err
	}

	response := struct {
		Token      string    `json:"token"`
		ID         uuid.UUID `json:"id,omitempty"`
		RefreshURL string    `json:"refresh,omitempty"`
	}{
		Token:      tokenString,
		ID:         opt.user.ID,
		RefreshURL: opt.refresh,
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	utils.JSONResponse(w, http.StatusOK, &response)

	return nil
}

func (u *Users) getTenantFromRequest(r *http.Request, user *storage.User) (uuid.UUID, error) {
	var uid uuid.UUID
	tenantID := mux.Vars(r)["id"]

	if tenantID != "" {
		tenantUUID, err := uuid.FromString(tenantID)
		if err != nil {
			return uuid.Nil, err
		}
		uid = tenantUUID
	} else {
		uid = user.GetDefaultMembership()
	}

	return uid, nil
}

func (u *Users) getMembershipLogin(ctx context.Context, tenantID, userID uuid.UUID) (*storage.Membership, error) {
	membership, err := u.Storage.GetMembership(ctx, tenantID, userID)

	if membership == nil {
		return nil, errors.ErrMembershipNotFound
	}

	// Don't allow login with invited membership
	if membership.MembershipStatus != storage.ActiveState {
		return nil, errors.ErrMembershipNotActive
	}

	if err != nil {
		return nil, err
	}

	return membership, nil
}

// Login is a login endpoint handler
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var writePerm bool
	if v := r.FormValue("permissions"); v != "" {
		writePerm, _ = strconv.ParseBool(v)
	}

	type loginRequest struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	var request *loginRequest
	if name, password, ok := r.BasicAuth(); ok {
		request = &loginRequest{
			Name:     name,
			Password: password,
		}
	} else if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			utils.JSONError(w, err.Error(), errors.CodeBadRequest)
			return
		}
	}

	ctx, cancel := u.context(r)
	defer cancel()

	var (
		user       *storage.User
		remoteAddr string
	)

	if request == nil {
		// Try login by IP
		var err error
		remoteAddr = utils.GetRemoteAddr(r)
		log.WithField("address", remoteAddr).Println("Empty login request")

		user, err = u.Storage.GetServiceAccountByAddress(ctx, remoteAddr)
		if err != nil {
			if err != errors.ErrUserNotFound {
				log.Error(err)
			}

			utils.JSONError(w, "", errors.CodeUnauthorized)
			return
		}
	} else {
		// Normal login
		if request.Name == "" || request.Password == "" {
			utils.JSONError(w, "", errors.CodeUnauthorized)
			return
		}

		var err error
		user, err = u.Storage.GetUserByEmail(ctx, storage.AccountRegular, request.Name)
		if err != nil {
			log.Error(err)
			utils.JSONError(w, "", errors.CodeUnauthorized)
			return
		}

		if len(user.PasswordHash) == 0 {
			utils.JSONError(w, "", errors.CodeUnauthorized)
			return
		}

		if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(request.Password)); err != nil {
			log.Error(err)
			utils.JSONError(w, "", errors.CodeUnauthorized)
			return
		}

		// Don't allow unverified users to log in
		if !user.EmailVerified {
			utils.JSONErrorResponse(w, errors.ErrEmailNotVerified)
			return
		}
	}

	tid, err := u.getTenantFromRequest(r, user)

	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	membership, err := u.getMembershipLogin(ctx, tid, user.ID)
	if err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	var role rbac.Role
	if writePerm {
		role, err = u.Enforcer.GetRole(ctx, membership.Roles.Get()...)
		if err != nil {
			utils.JSONErrorResponse(w, err)
			return
		}
	}

	opt := userTokenOptions{
		addr:          remoteAddr,
		user:          user,
		membership:    membership,
		role:          role,
		sessionMaxAge: u.SessionMaxAge,
		refresh:       u.RefreshURL(),
	}

	if err := u.writeUserToken(w, &opt); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	if err := u.Storage.UpdateLoginInfo(ctx, user.ID, utils.GetRemoteAddr(r)); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvLogin, user.ID, user.ID, r)).WithField("email", user.Email).Printf("User %v logged in", user.ID)
	}
}

// Refresh is a refresh endpoint handler
func (u *Users) Refresh(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey{}).(*storage.User)
	member := r.Context().Value(MembershipContextKey{}).(*storage.Membership)
	token := r.Context().Value(TokenContextKey).(*jwt.Token)

	ctx, cancel := u.context(r)
	defer cancel()

	if err := u.Storage.UpdateRefreshInfo(ctx, self.ID, utils.GetRemoteAddr(r)); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	claims := token.Claims.(jwt.MapClaims)

	var (
		role rbac.Role
		err  error
	)
	if _, ok := claims[utils.NSClaim(u.Namespace, "permissions")]; ok {
		role, err = u.Enforcer.GetRole(ctx, member.Roles.Get()...)
		if err != nil {
			utils.JSONErrorResponse(w, err)
			return
		}
	}

	opt := userTokenOptions{
		user:          self,
		membership:    member,
		role:          role,
		sessionMaxAge: u.SessionMaxAge,
		refresh:       u.RefreshURL(),
	}

	opt.addr, _ = claims[utils.NSClaim(u.Namespace, "address")].(string)

	if err := u.writeUserToken(w, &opt); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}
}
