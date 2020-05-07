package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"

	hydra "github.com/ory/hydra-client-go/client"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
)

func main() {
	var (
		adminURL string
		bindAddr string
	)
	flag.StringVar(&adminURL, "a", "", "hydra admin URL address")
	flag.StringVar(&bindAddr, "b", ":9008", "Bind address")
	flag.Parse()

	if adminURL == "" {
		log.Fatal("hydra admin URL is required")
	}

	u, err := url.Parse(adminURL)
	if err != nil {
		log.Fatal(err)
	}

	h := hydra.NewHTTPClientWithConfig(
		nil,
		&hydra.TransportConfig{
			Schemes:  []string{u.Scheme},
			Host:     u.Host,
			BasePath: u.Path,
		},
	)

	http.Handle("/login", handleLogin(h.Admin))
	http.Handle("/consent", handleConsent(h.Admin))

	log.Printf("Starting hail-hydra on address %q. Hydra Admin URL is %q", bindAddr, adminURL)
	log.Fatal(http.ListenAndServe(bindAddr, nil))
}

func handleLogin(api admin.ClientService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		challengeID := r.URL.Query().Get("login_challenge")
		if challengeID == "" {
			log.Println("Missing login_challenge")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		loginChallenge, err := api.GetLoginRequest(
			&admin.GetLoginRequestParams{
				LoginChallenge: challengeID,
				Context:        r.Context(),
			},
		)
		if err != nil {
			log.Println("Unable to retrieve login challenge", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		subjectEmail := "foo@bar.com"
		loginAccept := admin.NewAcceptLoginRequestParamsWithContext(r.Context())
		loginAccept.LoginChallenge = loginChallenge.Payload.Challenge
		loginAccept.Body = &models.AcceptLoginRequest{Subject: &subjectEmail}

		acceptResp, err := api.AcceptLoginRequest(loginAccept)
		if err != nil {
			log.Println("Unable to accept consent", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, acceptResp.Payload.RedirectTo, http.StatusFound)

	})
}

func handleConsent(api admin.ClientService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		challenge := r.URL.Query().Get("consent_challenge")
		if challenge == "" {
			log.Println("Missing consent_challenge")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		consentChallenge, err := api.GetConsentRequest(
			&admin.GetConsentRequestParams{
				ConsentChallenge: challenge,
				Context:          r.Context(),
			},
		)
		if err != nil {
			log.Println("Unable to retrieve consent challenge", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		consentAccept := admin.NewAcceptConsentRequestParamsWithContext(r.Context())
		consentAccept.ConsentChallenge = consentChallenge.Payload.Challenge
		consentAccept.Body = &models.AcceptConsentRequest{
			GrantScope:               consentChallenge.Payload.RequestedScope,
			GrantAccessTokenAudience: consentChallenge.Payload.RequestedAccessTokenAudience,
		}

		acceptResp, err := api.AcceptConsentRequest(consentAccept)
		if err != nil {
			log.Println("Unable to accept consent", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, acceptResp.Payload.RedirectTo, http.StatusFound)
	})
}
