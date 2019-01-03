package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	TokenContextKey = "token"
)

func (u *Users) writeUserToken(w http.ResponseWriter, user *storage.User, membership *storage.Membership) error {
	roles := make([]string, 0, len(membership.Roles))
	for r := range membership.Roles {
		roles = append(roles, r)
	}

	now := time.Now()

	claims := jwt.MapClaims{
		"sub":    user.ID,
		"exp":    now.Add(u.SessionMaxAge).Unix(),
		"iat":    now.Unix(),
		"iss":    u.BaseURL(),
		"aud":    u.BaseURL(),
		"tenant": membership.TenantID,
		utils.NSClaim(u.Namespace, "email"): user.Email,
		utils.NSClaim(u.Namespace, "name"):  user.Name,
		utils.NSClaim(u.Namespace, "roles"): roles,
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
		ID:         user.ID,
		RefreshURL: u.RefreshURL(),
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	utils.JSONResponse(w, http.StatusOK, &response)

	return nil
}

// Login is a login endpoint handler
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	if name, password, ok := r.BasicAuth(); ok {
		request.Name = name
		request.Password = password
	} else {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			utils.JSONError(w, err.Error(), errors.CodeBadRequest)
			return
		}
	}

	if request.Name == "" || request.Password == "" {
		utils.JSONError(w, "", errors.CodeUnauthorized)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	user, err := u.Storage.GetUserByEmail(ctx, request.Name)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, "", errors.CodeUnauthorized)
		return
	}

	if len(user.PasswordHash) == 0 {
		utils.JSONError(w, "", errors.CodeUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(request.Password)); err != nil {
		log.Error(err)
		utils.JSONError(w, "", errors.CodeUnauthorized)
		return
	}

	// Don't allow unverified users to log in
	if !user.EmailVerified {
		utils.JSONErrorResponse(w, errors.ErrEmailNotVerified)
		return
	}

	if err := u.Storage.UpdateLoginInfo(ctx, user.ID, getRemoteAddr(r)); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	var uid uuid.UUID
	tenantId := mux.Vars(r)["id"]

	if tenantId != "" {
		tenantUUID, err := uuid.FromString(tenantId)
		if err != nil {
			log.Error(err)
			utils.JSONError(w, err.Error(), errors.CodeBadRequest)
			return
		}
		uid = tenantUUID
	} else {
		uid = user.GetDefaultMembership()
	}

	membership, err := u.MembershipStorage.GetMembership(ctx, uid, user.ID)

	if membership == nil {
		utils.JSONErrorResponse(w, errors.ErrMembershipNotFound)
		return
	}

	// Don't allow login with invited membership
	if membership.Membership_status != storage.ActiveState {
		utils.JSONErrorResponse(w, errors.ErrMembershipNotActive)
		return
	}

	if err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	if err := u.writeUserToken(w, user, membership); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvLogin, membership.ID, membership.ID, r)).WithField("email", user.Email).Printf("User %v logged into tenant %v", user.ID, membership.TenantID)
	}
}

func (u *Users) Refresh(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)
	member := r.Context().Value(MembershipContextKey).(*storage.Membership)

	ctx, cancel := u.context(r)
	defer cancel()

	if err := u.Storage.UpdateRefreshInfo(ctx, self.ID, getRemoteAddr(r)); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	if err := u.writeUserToken(w, self, member); err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}
}
