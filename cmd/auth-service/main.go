package main

//import (
//	"context"
//	"crypto/rsa"
//	"encoding/json"
//	"flag"
//	"fmt"
//	"io/ioutil"
//	"net/http"
//	"os"
//	"time"
//
//	"github.com/go-chi/chi/v5"
//	"gopkg.in/square/go-jose.v2"
//	"gopkg.in/square/go-jose.v2/jwt"
//
//	"github.com/yourorg/pluto-auth/gen"
//	"gopkg.in/yaml.v3"
//)
//
//type Config struct {
//	Clients        map[string]string `yaml:"clients"`
//	PrivateKeyPath string            `yaml:"private_key_path"`
//	PublicKeyPath  string            `yaml:"public_key_path"`
//}
//
//type AuthServer struct {
//	cfg    Config
//	signer jose.Signer
//	jwks   jose.JSONWebKeySet
//}
//
//// PostOauthToken реализует выдачу JWT
//func (a *AuthServer) PostOauthToken(w http.ResponseWriter, r *http.Request) {
//	// 1) дебраузер формы
//	if err := r.ParseForm(); err != nil {
//		http.Error(w, "invalid_request", http.StatusBadRequest)
//		return
//	}
//	clientID := r.Form.Get("client_id")
//	clientSecret := r.Form.Get("client_secret")
//	grantType := r.Form.Get("grant_type")
//
//	// 2) проверяем grant_type
//	if grantType != string(gen.ClientCredentials) {
//		http.Error(w, "unsupported_grant_type", http.StatusBadRequest)
//		return
//	}
//	// 3) сверяем секрет
//	if a.cfg.Clients[clientID] != clientSecret {
//		http.Error(w, "invalid_client", http.StatusUnauthorized)
//		return
//	}
//
//	// 4) формируем JWT
//	now := time.Now()
//	claims := jwt.Claims{
//		Issuer:   "pluto-auth",
//		Subject:  clientID,
//		Audience: jwt.Audience{"pluto-manifests"},
//		IssuedAt: jwt.NewNumericDate(now),
//		Expiry:   jwt.NewNumericDate(now.Add(time.Hour)),
//	}
//	token, err := jwt.Signed(a.signer).Claims(claims).CompactSerialize()
//	if err != nil {
//		http.Error(w, "server_error", http.StatusInternalServerError)
//		return
//	}
//
//	// 5) отвечаем JSON
//	resp := gen.TokenResponse{
//		AccessToken: token,
//		TokenType:   "Bearer",
//		ExpiresIn:   3600,
//	}
//	w.Header().Set("Content-Type", "application/json")
//	json.NewEncoder(w).Encode(resp)
//}
//
//// GetPublicKey отдаёт JWKS
//func (a *AuthServer) GetPublicKey(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Content-Type", "application/json")
//	// jose.JSONWebKeySet.MarshalJSON даёт правильный формат Jwks
//	raw, err := json.Marshal(a.jwks)
//	if err != nil {
//		http.Error(w, "server_error", http.StatusInternalServerError)
//		return
//	}
//	w.Write(raw)
//}
//
//func main() {
//	var cfgPath string
//	flag.StringVar(&cfgPath, "config", "config.yaml", "path to config.yaml")
//	flag.Parse()
//
//	// 1) читаем конфиг
//	data, err := ioutil.ReadFile(cfgPath)
//	if err != nil {
//		fmt.Fprintf(os.Stderr, "failed to read config: %v\n", err)
//		os.Exit(1)
//	}
//	var cfg Config
//	if err := yaml.Unmarshal(data, &cfg); err != nil {
//		fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
//		os.Exit(1)
//	}
//
//	// 2) загружаем ключи
//	privPEM, err := ioutil.ReadFile(cfg.PrivateKeyPath)
//	if err != nil {
//		panic(err)
//	}
//	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privPEM)
//	if err != nil {
//		panic(err)
//	}
//	jwk := jose.JSONWebKey{
//		Key:       privKey.Public().(*rsa.PublicKey),
//		Use:       "sig",
//		Algorithm: string(jose.RS256),
//		KeyID:     "pluto-auth-key-1",
//	}
//	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}
//
//	signer, err := jose.NewSigner(
//		jose.SigningKey{Algorithm: jose.RS256, Key: privKey},
//		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", jwk.KeyID),
//	)
//	if err != nil {
//		panic(err)
//	}
//
//	// 3) собираем HTTP-сервер
//	auth := &AuthServer{cfg: cfg, signer: signer, jwks: jwks}
//	router := chi.NewRouter()
//	// (если нужен healthz)
//	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
//		w.WriteHeader(200)
//	})
//	// вешаем oapi-контракт
//	router.Mount("/", gen.Handler(auth))
//
//	srv := &http.Server{
//		Addr:         ":8080",
//		Handler:      router,
//		ReadTimeout:  5 * time.Second,
//		WriteTimeout: 5 * time.Second,
//		IdleTimeout:  60 * time.Second,
//	}
//
//	fmt.Println("🔑 pluto-auth listening on :8080")
//	if err := srv.ListenAndServe(); err != nil {
//		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
//		os.Exit(1)
//	}
//}
