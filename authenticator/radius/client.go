package radius

import (
	"context"
	"fmt"
	"git.ecadlabs.com/ecad/auth/authenticator"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"net"
	"time"
)

// Error represents Radius return code
type Error int

func (e Error) Error() string {
	return fmt.Sprintf("radius: %v (code = %d)", radius.Code(e), e)
}

// Rejected returns true if authentication request was rejected explicitly
func (e Error) Rejected() bool {
	return radius.Code(e) == radius.CodeAccessReject
}

// Config represents RADIUS client configuration
type Config struct {
	Client              *radius.Client
	SecretGetter        func() ([]byte, error)
	Address             string
	AttributesExtractor func(*radius.Packet) map[string]interface{}
	Logger              log.FieldLogger
}

// Authenticator is a Radius authentication backend
type Authenticator struct {
	conf    Config
	counter *prometheus.CounterVec
	hist    *prometheus.HistogramVec
}

func defaultExtractor(p *radius.Packet) map[string]interface{} {
	val, err := rfc2865.SessionTimeout_Lookup(p)
	if err != nil {
		return nil
	}

	now := time.Now()

	return map[string]interface{}{
		"exp": now.Add(time.Duration(val) * time.Second).Unix(),
		"iat": now.Unix(),
	}
}

func (a *Authenticator) log() log.FieldLogger {
	if a.conf.Logger != nil {
		return a.conf.Logger
	}
	return log.StandardLogger()
}

func (a *Authenticator) client() *radius.Client {
	if a.conf.Client != nil {
		return a.conf.Client
	}

	return radius.DefaultClient
}

func (a *Authenticator) extractAttributes(p *radius.Packet) map[string]interface{} {
	if a.conf.AttributesExtractor != nil {
		return a.conf.AttributesExtractor(p)
	}

	return defaultExtractor(p)
}

func NewAuthenticator(conf *Config) *Authenticator {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "radius_requests_total",
		Help:        "Total number of RADIUS requests",
		ConstLabels: prometheus.Labels{"address": conf.Address},
	}, []string{"status"})

	if err := prometheus.Register(counter); err != nil {
		// Reuse collector
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			counter = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			panic(err)
		}
	}

	hist := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "radius_request_duration_milliseconds",
		Help:        "RADIUS request duration",
		ConstLabels: prometheus.Labels{"address": conf.Address},
		// TODO buckets
	}, []string{"status"})

	if err := prometheus.Register(hist); err != nil {
		// Reuse collector
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			hist = are.ExistingCollector.(*prometheus.HistogramVec)
		} else {
			panic(err)
		}
	}

	return &Authenticator{
		conf:    *conf,
		counter: counter,
		hist:    hist,
	}
}

func (a *Authenticator) exchange(ctx context.Context, pkt *radius.Packet) (response *radius.Packet, err error) {
	timestamp := time.Now()
	response, err = a.client().Exchange(ctx, pkt, a.conf.Address)
	duration := time.Since(timestamp)

	var status string
	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			status = "Error-Timeout"
		} else {
			status = "Error-Unknown"
		}
	} else {
		status = response.Code.String()
	}

	a.counter.WithLabelValues(status).Inc()
	a.hist.WithLabelValues(status).Observe(float64(duration / time.Millisecond))

	return
}

// Authenticate performs Radius request
func (a *Authenticator) Authenticate(ctx context.Context, cred *authenticator.Credentials) (authenticator.Result, error) {
	secret, err := a.conf.SecretGetter()
	if err != nil {
		return nil, err
	}

	pkt := radius.New(radius.CodeAccessRequest, secret)
	rfc2865.UserName_SetString(pkt, cred.ID)
	rfc2865.UserPassword_AddString(pkt, string(cred.Secret))

	response, err := a.exchange(ctx, pkt)
	if err != nil {
		return nil, err
	}

	fields := make(log.Fields)
	for key, val := range response.Attributes {
		fields[fmt.Sprintf("attribute-%d", key)] = val
	}

	a.log().WithFields(fields).WithField("code", response.Code).Println("RADIUS reply")

	if response.Code != radius.CodeAccessAccept {
		return nil, Error(response.Code)
	}

	return &Result{
		Packet: response,
		a:      a,
	}, nil
}

// Ping sends dummy request
func (a *Authenticator) Ping(ctx context.Context) error {
	secret, err := a.conf.SecretGetter()
	if err != nil {
		return err
	}

	pkt := radius.New(radius.CodeAccessRequest, secret)
	if _, err := a.exchange(ctx, pkt); err != nil {
		return err
	}

	return nil
}

type Result struct {
	Packet *radius.Packet
	a      *Authenticator
}

func (r *Result) Claims() map[string]interface{} {
	// Extract additional attributes
	return r.a.extractAttributes(r.Packet)
}

// Build-time checks
var _ authenticator.Error = Error(0)
var _ authenticator.Authenticator = &Authenticator{}
