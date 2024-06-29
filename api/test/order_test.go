package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"testing"
	"time"

	"github.com/irsalhamdi/e-commerce-video/core/course"
	"github.com/plutov/paypal/v4"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
)

type orderTest struct {
	*TestEnv
}

func TestOrder(t *testing.T) {
	env, err := NewTestEnv(t, "order_test")
	if err != nil {
		t.Fatalf("initializing test env: %v", err)
	}

	ot := &orderTest{env}
	ct := &courseTest{env}
	rt := &cartTest{env}

	c1 := ct.createCourseOK(t)
	c2 := ct.createCourseOK(t)
	_ = ct.createCourseOK(t)
	_ = ct.createCourseOK(t)
	c3 := ct.createCourseOK(t)
	c4 := ct.createCourseOK(t)

	ct.listCoursesOwnedOK(t, []course.Course{})

	rt.createItemOK(t, c1.ID)
	rt.createItemOK(t, c2.ID)

	ot.Paypal.expectedCart = []course.Course{c1, c2}
	ot.testPaypal(t)

	ct.listCoursesOwnedOK(t, []course.Course{c1, c2})

	rt.createItemOK(t, c3.ID)
	rt.createItemOK(t, c4.ID)

	ot.Stripe.expectedCart = []course.Course{c3, c4}
	ot.testStripe(t)

	ct.listCoursesOwnedOK(t, []course.Course{c1, c2, c3, c4})
}

func (ot *orderTest) testPaypal(t *testing.T) {
	if err := Login(ot.Server, ot.UserEmail, ot.UserPass); err != nil {
		t.Fatal(err)
	}
	defer Logout(ot.Server)

	r, err := http.NewRequest(http.MethodPost, ot.URL+"/orders/paypal", nil)
	if err != nil {
		t.Fatal(err)
	}

	w, err := ot.Client().Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Body.Close()

	if w.StatusCode != http.StatusOK {
		t.Fatalf("can't create paypal order: status code %s", w.Status)
	}

	var ord paypal.Order
	if err := json.NewDecoder(w.Body).Decode(&ord); err != nil {
		t.Fatalf("cannot unmarshal paypal order: %v", err)
	}

	r, err = http.NewRequest(http.MethodPost, ot.URL+"/orders/paypal/"+ord.ID+"/capture", nil)
	if err != nil {
		t.Fatal(err)
	}

	w, err = ot.Client().Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Body.Close()

	if w.StatusCode != http.StatusNoContent {
		t.Fatalf("can't capture paypal order: status code %s", w.Status)
	}
}

func (ot *orderTest) testStripe(t *testing.T) {
	if err := Login(ot.Server, ot.UserEmail, ot.UserPass); err != nil {
		t.Fatal(err)
	}
	defer Logout(ot.Server)

	r, err := http.NewRequest(http.MethodPost, ot.URL+"/orders/stripe", nil)
	if err != nil {
		t.Fatal(err)
	}

	w, err := ot.Client().Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Body.Close()

	if w.StatusCode != http.StatusOK {
		t.Fatalf("can't create stripe order: status code %s", w.Status)
	}

	urlBytes, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}

	var url string
	if err := json.Unmarshal(urlBytes, &url); err != nil {
		t.Fatal(err)
	}

	obj := map[string]any{
		"id":   path.Base(url),
		"mode": stripe.CheckoutSessionModePayment,
	}

	raw, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}

	evt := stripe.Event{
		APIVersion: "2022-11-15",
		Type:       "checkout.session.completed",
		Data: &stripe.EventData{
			Raw: json.RawMessage(raw),
		},
	}

	b, err := json.Marshal(evt)
	if err != nil {
		t.Fatal(err)
	}

	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   b,
		Secret:    ot.WebhookSecret,
		Timestamp: time.Now(),
	})

	r, err = http.NewRequest(http.MethodPost, ot.URL+"/orders/stripe/capture", bytes.NewBuffer(b))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Stripe-Signature", signed.Header)

	w, err = ot.Client().Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Body.Close()

	if w.StatusCode != http.StatusNoContent {
		t.Fatalf("can't trigger stripe webhook: status code %s", w.Status)
	}
}
