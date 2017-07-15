package marvin

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/dgrijalva/jwt-go"
)

const JWTAlg = "AppEngine"

func init() {
	jwt.RegisterSigningMethod(JWTAlg, func() jwt.SigningMethod {
		return &signingMethodAppEngine{}
	})
}

// signingMethodAppEngine uses the built-in AppEngine private keys to sign
// JWT tokens.
type signingMethodAppEngine struct{}

func (s *signingMethodAppEngine) Alg() string {
	return JWTAlg
}

var (
	Timeout     = 250 * time.Millisecond
	initBackOff = 50 * time.Millisecond
	maxBackOff  = 2 * time.Second
	maxAttempts = 10
)

func exponential(d time.Duration) time.Duration {
	d *= 2
	if d > maxBackOff {
		d = maxBackOff
	}
	return d
}

func (s *signingMethodAppEngine) Sign(signingString string, key interface{}) (string, error) {
	var ctx context.Context

	switch k := key.(type) {
	case context.Context:
		ctx = k
	default:
		return "", jwt.ErrInvalidKey
	}
	var (
		err       error
		signature []byte
		backOff   = initBackOff
	)
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		// only sleep after an error
		if attempts > 1 {
			time.Sleep(backOff)
			// bump the backoff up
			backOff = exponential(backOff)
		}
		tctx, cancel := context.WithTimeout(ctx, Timeout)
		_, signature, err = appengine.SignBytes(tctx, []byte(signingString))
		cancel()
		if err != nil {
			log.Warningf(ctx, "unable to sign bytes on attempt %d: %s", attempts, err)
			continue
		}
		break
	}
	if err != nil {
		return "", err
	}
	return jwt.EncodeSegment(signature), nil
}

func (s *signingMethodAppEngine) Verify(signingString, signature string, key interface{}) error {
	var ctx context.Context
	switch k := key.(type) {
	case context.Context:
		ctx = k
	default:
		return jwt.ErrInvalidKey
	}

	sig, err := jwt.DecodeSegment(signature)
	if err != nil {
		return err
	}

	backOff := initBackOff
	var certs []appengine.Certificate
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		// only sleep after an error
		if attempts > 1 {
			time.Sleep(backOff)
			// bump the backoff up
			backOff = exponential(backOff)
		}
		tctx, cancel := context.WithTimeout(ctx, Timeout)
		certs, err = appengine.PublicCertificates(tctx)
		cancel()
		if err != nil {
			log.Warningf(ctx, "unable to get public certs on attempt %d: %s", attempts, err)
			continue
		}
		break
	}
	if err != nil {
		return err
	}

	hasher := sha256.New()
	hasher.Write([]byte(signingString))
	for _, cert := range certs {
		rsaKey, err := jwt.ParseRSAPublicKeyFromPEM(cert.Data)
		if err != nil {
			return err
		}

		err = rsa.VerifyPKCS1v15(rsaKey, crypto.SHA256, hasher.Sum(nil), sig)
		if err == nil {
			return nil
		}
	}

	return jwt.ErrSignatureInvalid
}
