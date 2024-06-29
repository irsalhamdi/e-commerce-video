package api

import (
	"context"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/mux"
	"github.com/irsalhamdi/e-commerce-video/api/background"
	"github.com/irsalhamdi/e-commerce-video/api/middleware"
	"github.com/irsalhamdi/e-commerce-video/api/web"
	"github.com/irsalhamdi/e-commerce-video/config"
	"github.com/irsalhamdi/e-commerce-video/core/auth"
	"github.com/irsalhamdi/e-commerce-video/core/cart"
	"github.com/irsalhamdi/e-commerce-video/core/course"
	"github.com/irsalhamdi/e-commerce-video/core/order"
	"github.com/irsalhamdi/e-commerce-video/core/token"
	"github.com/irsalhamdi/e-commerce-video/core/user"
	"github.com/irsalhamdi/e-commerce-video/core/video"
	"github.com/jmoiron/sqlx"
	"github.com/plutov/paypal/v4"
	"github.com/sirupsen/logrus"
	stripecl "github.com/stripe/stripe-go/v74/client"
)

type APIConfig struct {
	CorsOrigin         string
	Log                logrus.FieldLogger
	DB                 *sqlx.DB
	Session            *scs.SessionManager
	Mailer             token.Mailer
	TokenTimeout       time.Duration
	Background         *background.Background
	Paypal             *paypal.Client
	Stripe             *stripecl.API
	StripeCfg          config.Stripe
	Providers          map[string]auth.Provider
	LoginRedirectURL   string
	ActivationRequired bool
}

type api struct {
	*mux.Router
	mw  []web.Middleware
	log logrus.FieldLogger
}

func APIMux(cfg APIConfig) http.Handler {
	a := &api{
		Router: mux.NewRouter(),
		log:    cfg.Log,
	}

	a.mw = append(a.mw, auth.LoadAndSave(cfg.Session))
	a.mw = append(a.mw, middleware.RequestID())
	a.mw = append(a.mw, middleware.Logger(cfg.Log))
	a.mw = append(a.mw, middleware.Errors(cfg.Log))
	a.mw = append(a.mw, middleware.Panics())

	if cfg.CorsOrigin != "" {
		a.mw = append(a.mw, middleware.Cors(cfg.CorsOrigin))

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusNoContent)
			return nil
		}

		a.Handle(http.MethodOptions, "/{path:.*}", h)
	}

	authen := auth.Authenticate(cfg.Session)
	admin := auth.Admin(cfg.Session)

	a.Handle(http.MethodPost, "/auth/signup", auth.HandleSignup(cfg.DB, cfg.Session, cfg.ActivationRequired))
	a.Handle(http.MethodPost, "/auth/login", auth.HandleLogin(cfg.DB, cfg.Session))
	a.Handle(http.MethodPost, "/auth/logout", auth.HandleLogout(cfg.Session))
	a.Handle(http.MethodGet, "/auth/oauth-login/{provider}", auth.HandleOauthLogin(cfg.Session, cfg.Providers))
	a.Handle(http.MethodGet, "/auth/oauth-callback/{provider}", auth.HandleOauthCallback(cfg.DB, cfg.Session, cfg.Providers, cfg.LoginRedirectURL))

	a.Handle(http.MethodPost, "/tokens", token.HandleToken(cfg.DB, cfg.Mailer, cfg.TokenTimeout, cfg.Background))
	a.Handle(http.MethodPost, "/tokens/activate", token.HandleActivation(cfg.DB, cfg.Session))
	a.Handle(http.MethodPost, "/tokens/recover", token.HandleRecovery(cfg.DB))

	a.Handle(http.MethodGet, "/users/current", user.HandleShowCurrent(cfg.DB), authen)
	a.Handle(http.MethodGet, "/users/{id}", user.HandleShow(cfg.DB), authen)
	a.Handle(http.MethodPost, "/users", user.HandleCreate(cfg.DB), authen)

	a.Handle(http.MethodGet, "/courses/owned", course.HandleListOwned(cfg.DB), authen)
	a.Handle(http.MethodGet, "/courses/{course_id}/videos", video.HandleListByCourse(cfg.DB))
	a.Handle(http.MethodGet, "/courses/{course_id}/progress", video.HandleListProgressByCourse(cfg.DB), authen)
	a.Handle(http.MethodGet, "/courses/{id}", course.HandleShow(cfg.DB))
	a.Handle(http.MethodGet, "/courses", course.HandleList(cfg.DB))
	a.Handle(http.MethodPost, "/courses", course.HandleCreate(cfg.DB), admin)
	a.Handle(http.MethodPut, "/courses/{id}", course.HandleUpdate(cfg.DB), admin)

	a.Handle(http.MethodGet, "/videos/{id}/full", video.HandleShowFull(cfg.DB), authen)
	a.Handle(http.MethodGet, "/videos/{id}/free", video.HandleShowFree(cfg.DB))
	a.Handle(http.MethodGet, "/videos/{id}", video.HandleShow(cfg.DB))
	a.Handle(http.MethodGet, "/videos", video.HandleList(cfg.DB))
	a.Handle(http.MethodPost, "/videos", video.HandleCreate(cfg.DB), admin)
	a.Handle(http.MethodPut, "/videos/{id}/progress", video.HandleUpdateProgress(cfg.DB), authen)
	a.Handle(http.MethodPut, "/videos/{id}", video.HandleUpdate(cfg.DB), admin)

	a.Handle(http.MethodGet, "/cart", cart.HandleShow(cfg.DB), authen)
	a.Handle(http.MethodDelete, "/cart", cart.HandleDelete(cfg.DB), authen)
	a.Handle(http.MethodPut, "/cart/items", cart.HandleCreateItem(cfg.DB), authen)
	a.Handle(http.MethodDelete, "/cart/items/{course_id}", cart.HandleDeleteItem(cfg.DB), authen)

	a.Handle(http.MethodPost, "/orders/paypal", order.HandlePaypalCheckout(cfg.DB, cfg.Paypal), authen)
	a.Handle(http.MethodPost, "/orders/paypal/{id}/capture", order.HandlePaypalCapture(cfg.DB, cfg.Paypal), authen)
	a.Handle(http.MethodPost, "/orders/stripe", order.HandleStripeCheckout(cfg.DB, cfg.Stripe, cfg.StripeCfg), authen)
	a.Handle(http.MethodPost, "/orders/stripe/capture", order.HandleStripeCapture(cfg.DB, cfg.StripeCfg))

	return a.Router
}

func (a *api) Handle(method string, path string, handler web.Handler, mw ...web.Middleware) {

	handler = web.WrapMiddleware(mw, handler)

	handler = web.WrapMiddleware(a.mw, handler)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		if err := handler(ctx, w, r); err != nil {

			a.log.WithFields(logrus.Fields{
				"req_id":  middleware.ContextRequestID(ctx),
				"message": err,
			}).Error("ERROR")
		}
	})

	a.Router.Handle(path, h).Methods(method)
}
