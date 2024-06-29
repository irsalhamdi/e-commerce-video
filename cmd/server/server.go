package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/ardanlabs/conf/v3"
	"github.com/irsalhamdi/e-commerce-video/api"
	"github.com/irsalhamdi/e-commerce-video/api/background"
	"github.com/irsalhamdi/e-commerce-video/config"
	"github.com/irsalhamdi/e-commerce-video/core/auth"
	"github.com/irsalhamdi/e-commerce-video/database"
	"github.com/irsalhamdi/e-commerce-video/email"
	"github.com/plutov/paypal/v4"
	"github.com/sirupsen/logrus"
	stripecl "github.com/stripe/stripe-go/v74/client"
)

func main() {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	if err := Run(log); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func Run(logger *logrus.Logger) error {
	logger.Infof("starting server")
	defer logger.Info("shutdown complete")

	const prefix = "GOVOD"
	var cfg config.Config
	if _, err := conf.Parse(prefix, &cfg); err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	lw := logger.Writer()
	defer lw.Close()
	errLog := log.New(lw, "", 0)

	db, err := database.Open(cfg.DB)
	if err != nil {
		return fmt.Errorf("failed to open db connection: %w", err)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	links := email.Links{
		ActivationURL: cfg.Email.ActivationURL,
		RecoveryURL:   cfg.Email.RecoveryURL,
	}
	mail := email.New(cfg.Email.Address, cfg.Email.Password, cfg.Email.Host, cfg.Email.Port, links)

	bg := background.New(logger)

	pp, err := paypal.NewClient(
		cfg.Paypal.ClientID,
		cfg.Paypal.Secret,
		cfg.Paypal.URL,
	)
	if err != nil {
		return fmt.Errorf("failed to build the paypal client: %w", err)
	}

	if _, err = pp.GetAccessToken(context.TODO()); err != nil {
		return fmt.Errorf("failed to get the first paypal access token: %w", err)
	}

	strp := &stripecl.API{}
	strp.Init(cfg.Stripe.APISecret, nil)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Oauth.DiscoveryTimeout)
	defer cancel()
	google := cfg.Oauth.Google
	oauthProvs, err := auth.MakeProviders(ctx, []auth.ProviderConfig{
		{Name: "google", Client: google.Client, Secret: google.Secret, URL: google.URL, RedirectURL: google.RedirectURL},
	})
	if err != nil {
		return fmt.Errorf("failed to discover oauth providers: %w", err)
	}

	mux := api.APIMux(api.APIConfig{
		CorsOrigin:         cfg.Cors.Origin,
		Log:                logger,
		DB:                 db,
		Session:            sessionManager,
		Mailer:             mail,
		TokenTimeout:       cfg.Email.TokenTimeout,
		Background:         bg,
		Paypal:             pp,
		Stripe:             strp,
		StripeCfg:          cfg.Stripe,
		Providers:          oauthProvs,
		LoginRedirectURL:   cfg.Oauth.LoginRedirectURL,
		ActivationRequired: cfg.Auth.ActivationRequired,
	})

	api := http.Server{
		Handler:      mux,
		Addr:         cfg.Web.Address,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     errLog,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Infof("starting api router at %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Infof("shutting down: signal %s", sig)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}

		if err := bg.Shutdown(ctx); err != nil {
			return fmt.Errorf("could not complete all background tasks: %w", err)
		}
	}
	return nil
}
