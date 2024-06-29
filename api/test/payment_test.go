package test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/irsalhamdi/e-commerce-video/api/web"
	"github.com/irsalhamdi/e-commerce-video/core/course"
	"github.com/plutov/paypal/v4"
	mock "github.com/stripe/stripe-mock/param"
)

type mockPaypal struct {
	expectedCart []course.Course
}

func (m *mockPaypal) handle() http.Handler {
	checkout := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pu struct {
			Units []paypal.PurchaseUnitRequest `json:"purchase_units"`
		}
		if err := json.NewDecoder(r.Body).Decode(&pu); err != nil {
			web.Respond(context.Background(), w, nil, 400)
			return
		}

		if len(pu.Units) != 1 {
			web.Respond(context.Background(), w, nil, 400)
			return
		}

		if len(pu.Units[0].Items) != len(m.expectedCart) {
			web.Respond(context.Background(), w, nil, 400)
			return
		}

		var tot int
		for _, c := range m.expectedCart {
			tot += c.Price
		}

		if pu.Units[0].Amount.Value != strconv.Itoa(tot) {
			web.Respond(context.Background(), w, nil, 400)
			return
		}

		randID := fmt.Sprintf("paypal-%d", rand.Intn(300))
		ord := paypal.Order{ID: randID}
		web.Respond(context.Background(), w, ord, 200)
	})

	capture := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ord := paypal.Order{Status: "COMPLETED"}
		web.Respond(context.Background(), w, ord, 200)
	})

	r := mux.NewRouter()
	r.Handle("/v2/checkout/orders", checkout).Methods("POST")
	r.Handle("/v2/checkout/orders/{id}/capture", capture).Methods("POST")
	return r
}

type mockStripe struct {
	expectedCart []course.Course
}

func (m *mockStripe) handle() http.Handler {
	checkout := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params, _ := mock.ParseParams(r)
		lines := params["line_items"].(map[string]any)

		n := 0
		tot := 0
		for _, li := range lines {
			it := li.(map[string]any)

			if it["quantity"] != "1" {
				web.Respond(context.Background(), w, nil, 400)
				return
			}

			pd := it["price_data"].(map[string]any)
			s := pd["unit_amount"].(string)
			amount, err := strconv.ParseInt(s, 10, 0)
			if err != nil {
				web.Respond(context.Background(), w, err, 400)
				return
			}

			tot += int(amount / 100)
			n += 1
		}

		if n != len(m.expectedCart) {
			web.Respond(context.Background(), w, nil, 400)
			return
		}

		exp := 0
		for _, c := range m.expectedCart {
			exp += c.Price
		}

		if tot != exp {
			web.Respond(context.Background(), w, nil, 400)
			return
		}

		randID := fmt.Sprintf("stripe-%d", rand.Intn(300))
		ord := map[string]any{"ID": randID, "URL": randID}
		web.Respond(context.Background(), w, ord, 201)
	})

	r := mux.NewRouter()
	r.Handle("/v1/checkout/sessions", checkout).Methods("POST")
	return r
}
