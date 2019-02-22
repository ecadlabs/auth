package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/jsonpatch"
	"github.com/ecadlabs/auth/notification"
	"github.com/ecadlabs/auth/query"
	"github.com/ecadlabs/auth/rbac"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	UserContextKey       = "user"
	MembershipContextKey = "membership"
	DefaultLimit         = 20
)

type Users struct {
	Storage           *storage.Storage
	TenantStorage     *storage.TenantStorage
	MembershipStorage *storage.MembershipStorage
	Timeout           time.Duration

	SessionMaxAge    time.Duration
	JWTSecretGetter  func() ([]byte, error)
	JWTSigningMethod jwt.SigningMethod

	BaseURL         func() string
	UsersPath       string
	RefreshPath     string
	ResetPath       string
	LogPath         string
	EmailUpdatePath string
	Namespace       string

	Notifier               notification.Notifier
	ResetTokenMaxAge       time.Duration
	EmailUpdateTokenMaxAge time.Duration

	Enforcer rbac.Enforcer

	AuxLogger *log.Logger
}

func (u *Users) UsersURL() string {
	return u.BaseURL() + u.UsersPath
}

func (u *Users) RefreshURL() string {
	return u.BaseURL() + u.RefreshPath
}

func (u *Users) ResetURL() string {
	return u.BaseURL() + u.ResetPath
}

func (u *Users) LogURL() string {
	return u.BaseURL() + u.LogPath
}

func (u *Users) EmailUpdateURL() string {
	return u.BaseURL() + u.EmailUpdatePath
}

func (u *Users) context(r *http.Request) (context.Context, context.CancelFunc) {
	if u.Timeout != 0 {
		return context.WithTimeout(r.Context(), u.Timeout)
	}
	return r.Context(), func() {}
}

func (u *Users) GetUser(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)
	member := r.Context().Value(MembershipContextKey).(*storage.Membership)

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	perm := []string{permissionFull, permissionRead}
	if self.ID == uid {
		perm = append(perm, permissionReadSelf)
	}

	granted, err := role.IsAnyGranted(perm...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if !granted {
		utils.JSONErrorResponse(w, errors.ErrForbidden)
		return
	}

	user, err := u.Storage.GetUserByID(ctx, uid)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	utils.JSONResponse(w, http.StatusOK, user)
}

func (u *Users) GetUsers(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	member := r.Context().Value(MembershipContextKey).(*storage.Membership)

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	granted, err := role.IsAnyGranted(permissionFull, permissionRead)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if !granted {
		utils.JSONErrorResponse(w, errors.ErrForbidden)
		return
	}

	q, err := query.FromValues(r.Form, nil)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeQuerySyntax)
		return
	}

	if q.Limit <= 0 {
		q.Limit = DefaultLimit
	}

	userSlice, count, nextQuery, err := u.Storage.GetUsers(ctx, q)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if len(userSlice) == 0 && !q.TotalCount {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Pagination
	res := utils.Paginated{
		Value: userSlice,
	}

	if q.TotalCount {
		res.TotalCount = &count
	}

	if nextQuery != nil {
		nextUrl, err := url.Parse(u.UsersURL())
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

func (u *Users) resetToken(user *storage.User) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": now.Add(u.ResetTokenMaxAge).Unix(),
		"iat": now.Unix(),
		"iss": u.BaseURL(),
		"aud": u.ResetURL(),
		utils.NSClaim(u.Namespace, "gen"): user.PasswordGen,
	}

	token := jwt.NewWithClaims(u.JWTSigningMethod, claims)

	secret, err := u.JWTSecretGetter()
	if err != nil {
		return "", err
	}

	return token.SignedString(secret)
}

func (u *Users) NewUser(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)
	member := r.Context().Value(MembershipContextKey).(*storage.Membership)

	var user storage.CreateUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	if !utils.ValidEmail(user.Email) {
		utils.JSONErrorResponse(w, errors.ErrEmailFmt)
		return
	}

	if len(user.Roles) == 0 {
		utils.JSONErrorResponse(w, errors.ErrRolesEmpty)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Check write access
	fullGranted, err := role.IsAnyGranted(permissionFull)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	writeGranted, err := role.IsAnyGranted(permissionWrite)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if !fullGranted && !writeGranted {
		utils.JSONErrorResponse(w, errors.ErrForbidden)
		return
	}

	// Check delegating access
	if !fullGranted {
		granted, err := member.CanDelegate(role, user.Roles, permissionDelegatePrefix)

		if err != nil {
			log.Error(err)
			utils.JSONErrorResponse(w, err)
			return
		}
		if !granted {
			utils.JSONErrorResponse(w, errors.ErrForbidden)
			return
		}
	}

	user.EmailVerified = false
	user.PasswordHash = nil

	ret, err := u.Storage.NewUser(ctx, &user)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Create reset token
	token, err := u.resetToken(ret)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if err = u.Notifier.Notify(ctx, notification.NotificationInvite, &notification.NotificationData{
		Addr:        getRemoteAddr(r),
		CurrentUser: self,
		TargetUser:  ret,
		Token:       token,
		TokenMaxAge: u.ResetTokenMaxAge,
	}); err != nil {
		log.Error(err)
	}

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvCreate, member.ID, ret.ID, r)).WithFields(log.Fields{
			"email":          ret.Email,
			"name":           ret.Name,
			"added":          ret.Added,
			"email_verified": ret.EmailVerified,
		}).Printf("User %v created account %v from tenant %v", self.ID, ret.ID, member.TenantID)
	}

	w.Header().Set("Location", u.UsersURL()+ret.ID.String())
	utils.JSONResponse(w, http.StatusCreated, ret)
}

func (u *Users) PatchUser(w http.ResponseWriter, r *http.Request) {
	// TODO Email verification
	self := r.Context().Value(UserContextKey).(*storage.User)
	member := r.Context().Value(MembershipContextKey).(*storage.Membership)

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	var p jsonpatch.Patch
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ops, err := storage.OpsFromPatch(p)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if _, ok := ops.Update["password_hash"]; ok {
		utils.JSONError(w, "Invalid property", errors.CodeBadRequest)
		return
	}

	if v, ok := ops.Update["password"]; ok {
		delete(ops.Update, "password")

		if p, ok := v.(string); ok {
			if p == "" {
				utils.JSONError(w, "Password is empty", errors.CodeBadRequest)
				return
			}

			hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
			if err != nil {
				log.Error(err)
				utils.JSONErrorResponse(w, err)
				return
			}

			ops.Update["password_hash"] = hash
		}
	}

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	perm := []string{permissionFull, permissionWrite}
	if self.ID == uid {
		perm = append(perm, permissionWriteSelf)
	}

	granted, err := role.IsAnyGranted(perm...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if !granted {
		utils.JSONErrorResponse(w, errors.ErrForbidden)
		return
	}

	user, err := u.Storage.UpdateUser(ctx, uid, ops)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Log
	if u.AuxLogger != nil {
		if len(ops.Update) != 0 {
			u.AuxLogger.WithFields(logFields(EvUpdate, member.ID, uid, r)).WithFields(log.Fields(ops.Update)).Printf("User %v updated account %v from tenant %v", self.ID, uid, member.TenantID)

		}
	}

	utils.JSONResponse(w, http.StatusOK, user)
}

func (u *Users) DeleteUser(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)
	member := r.Context().Value(MembershipContextKey).(*storage.Membership)

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	perm := []string{permissionFull, permissionWrite}
	if self.ID == uid {
		perm = append(perm, permissionWriteSelf)
	}

	granted, err := role.IsAnyGranted(perm...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if !granted {
		utils.JSONErrorResponse(w, errors.ErrForbidden)
		return
	}

	tenants, err := u.TenantStorage.GetTenantsSoleMember(ctx, uid)

	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// TODO: archive user
	if err = u.Storage.DeleteUser(ctx, uid); err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvDelete, member.ID, uid, r)).Printf("User %v deleted account %v from tenant %v", self.ID, uid, member.TenantID)
	}

	individualTenants := []*storage.TenantModel{}

	for _, tenant := range tenants {
		if tenant.TenantType == storage.TenantTypeIndividual {
			individualTenants = append(individualTenants, tenant)
		}
	}

	// Archive any individual tenant made orphan by this deletion
	if len(individualTenants) > 0 {
		for _, tenant := range individualTenants {
			deleteErr := u.TenantStorage.DeleteTenant(ctx, tenant.ID)
			// Log
			if deleteErr == nil && u.AuxLogger != nil {
				u.AuxLogger.WithFields(logFields(EvArchiveTenant, member.ID, tenant.ID, r)).Printf("User %v from tenant %v archived tenant %v", self.ID, member.TenantID, tenant.ID)
			}

			if deleteErr != nil {
				err = deleteErr
			}
		}

		if err != nil {
			log.Error(err)
			utils.JSONErrorResponse(w, err)
			return
		}
	}

	// Archive any tenant made orphan by this deletion
	if len(tenants) > 0 {
		for _, tenant := range tenants {
			deleteErr := u.TenantStorage.DeleteTenant(ctx, tenant.ID)
			// Log
			if deleteErr == nil && u.AuxLogger != nil {
				u.AuxLogger.WithFields(logFields(EvArchiveTenant, self.ID, tenant.ID, r)).Printf("User %v archived tenant %v", self.ID, tenant.ID)
			}

			if deleteErr != nil {
				err = deleteErr
			}
		}

		if err != nil {
			log.Error(err)
			utils.JSONErrorResponse(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (u *Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	// Allow authorized requests too
	if token, err := jwtmiddleware.FromAuthHeader(r); err != nil {
		utils.JSONError(w, err.Error(), errors.CodeUnauthorized)
		return
	} else if token != "" {
		request.Token = token
	}

	if request.Token == "" {
		utils.JSONErrorResponse(w, errors.ErrTokenEmpty)
		return
	}

	if request.Password == "" {
		utils.JSONErrorResponse(w, errors.ErrPasswordEmpty)
		return
	}

	// Verify token
	token, err := jwt.Parse(request.Token, func(token *jwt.Token) (interface{}, error) { return u.JWTSecretGetter() })
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeTokenFmt)
		return
	}

	if u.JWTSigningMethod.Alg() != token.Header["alg"] {
		log.Errorf("Expected %s signing method but token specified %s", u.JWTSigningMethod.Alg(), token.Header["alg"])
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	if !token.Valid {
		log.Errorln("Invalid token")
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	// Verify audience
	claims := token.Claims.(jwt.MapClaims)
	if !claims.VerifyAudience(u.ResetURL(), true) {
		utils.JSONErrorResponse(w, errors.ErrAudience)
		return
	}

	// Get password generation
	gen, ok := claims[utils.NSClaim(u.Namespace, "gen")].(float64)
	if !ok {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	// Get user id
	sub, ok := claims["sub"].(string)
	if !ok {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	id, err := uuid.FromString(sub)
	if err != nil {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	err = u.Storage.UpdatePasswordWithGen(ctx, id, hash, int(gen))
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvReset, id, id, r)).Printf("Password for account %v reset", id)
	}
}

func (u *Users) SendResetRequest(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email string `json:"email"`
	}

	if r.Method == "POST" {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Error(err)
			utils.JSONError(w, err.Error(), errors.CodeBadRequest)
			return
		}
	} else {
		// Allow GET requests for testing purposes
		request.Email = r.FormValue("email")
	}

	if request.Email == "" {
		utils.JSONErrorResponse(w, errors.ErrEmailEmpty)
		return
	}

	// Don't return anything
	defer w.WriteHeader(http.StatusNoContent)

	ctx, cancel := u.context(r)
	defer cancel()

	user, err := u.Storage.GetUserByEmail(ctx, request.Email)
	if err != nil {
		log.Error(err)
		return
	}

	// Create reset token
	token, err := u.resetToken(user)
	if err != nil {
		log.Error(err)
		return
	}

	if err = u.Notifier.Notify(r.Context(), notification.NotificationReset, &notification.NotificationData{
		Addr:        getRemoteAddr(r),
		CurrentUser: user,
		TargetUser:  user,
		Token:       token,
		TokenMaxAge: u.ResetTokenMaxAge,
	}); err != nil {
		log.Error(err)
		return
	}

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvResetRequest, user.ID, user.ID, r)).WithField("email", user.Email).Printf("User %v requested password reset", user.ID)
	}
}

func (u *Users) SendUpdateEmailRequest(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)
	member := r.Context().Value(MembershipContextKey).(*storage.Membership)

	var request struct {
		Email string    `json:"email"`
		ID    uuid.UUID `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	if request.Email == "" {
		utils.JSONErrorResponse(w, errors.ErrEmailEmpty)
		return
	}

	if !utils.ValidEmail(request.Email) {
		utils.JSONErrorResponse(w, errors.ErrEmailFmt)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	perm := []string{permissionFull, permissionWrite}
	if self.ID == request.ID {
		perm = append(perm, permissionWriteSelf)
	}

	granted, err := role.IsAnyGranted(perm...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if !granted {
		utils.JSONErrorResponse(w, errors.ErrForbidden)
		return
	}

	user, err := u.Storage.GetUserByID(ctx, request.ID)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Create update token
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": now.Add(u.EmailUpdateTokenMaxAge).Unix(),
		"iat": now.Unix(),
		"iss": u.BaseURL(),
		"aud": u.EmailUpdateURL(),
		utils.NSClaim(u.Namespace, "email"): request.Email,
		utils.NSClaim(u.Namespace, "gen"):   user.EmailGen,
	}

	token := jwt.NewWithClaims(u.JWTSigningMethod, claims)

	secret, err := u.JWTSecretGetter()
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	tokStr, err := token.SignedString(secret)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if err = u.Notifier.Notify(ctx, notification.NotificationEmailUpdateRequest, &notification.NotificationData{
		Addr:        getRemoteAddr(r),
		To:          []string{request.Email},
		Email:       request.Email,
		CurrentUser: user,
		TargetUser:  user,
		Token:       tokStr,
		TokenMaxAge: u.EmailUpdateTokenMaxAge,
	}); err != nil {
		log.Error(err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvEmailUpdateRequest, self.ID, user.ID, r)).WithField("email", request.Email).Printf("User %v requested email update for user %v", self.ID, user.ID)
	}
}

func (u *Users) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	if request.Token == "" {
		utils.JSONErrorResponse(w, errors.ErrTokenEmpty)
		return
	}

	// Verify token
	token, err := jwt.Parse(request.Token, func(token *jwt.Token) (interface{}, error) { return u.JWTSecretGetter() })
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeTokenFmt)
		return
	}

	if u.JWTSigningMethod.Alg() != token.Header["alg"] {
		log.Errorf("Expected %s signing method but token specified %s", u.JWTSigningMethod.Alg(), token.Header["alg"])
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	if !token.Valid {
		log.Errorln("Invalid token")
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	// Verify audience
	claims := token.Claims.(jwt.MapClaims)
	if !claims.VerifyAudience(u.EmailUpdateURL(), true) {
		utils.JSONErrorResponse(w, errors.ErrAudience)
		return
	}

	// Get email
	email, ok := claims[utils.NSClaim(u.Namespace, "email")].(string)
	if !ok || !utils.ValidEmail(email) {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	// Get email generation
	gen, ok := claims[utils.NSClaim(u.Namespace, "gen")].(float64)
	if !ok {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	// Get user id
	sub, ok := claims["sub"].(string)
	if !ok {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	id, err := uuid.FromString(sub)
	if err != nil {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	user, prevEmail, err := u.Storage.UpdateEmailWithGen(ctx, id, email, int(gen))
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if err = u.Notifier.Notify(ctx, notification.NotificationEmailUpdate, &notification.NotificationData{
		Addr:        getRemoteAddr(r),
		Email:       prevEmail,
		To:          []string{prevEmail, user.Email},
		CurrentUser: user,
		TargetUser:  user,
	}); err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvEmailUpdate, id, id, r)).WithField("email", email).Printf("Email for account %v updated", id)
	}
}
