package stripeCheckoutController

import (
	"encoding/json"
	"net/http"
	"server/common"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/context"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

type experienceCheckoutPayload struct {
	SuccessUrl string `json:"successUrl"`
	CancelUrl  string `json:"cancelUrl"`
}

type response struct {
	Url string `json:"url"`
}

// Buy 1000 experience points
func HandleExperienceCheckout(w http.ResponseWriter, r *http.Request) {
	payload := experienceCheckoutPayload{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if payload.SuccessUrl == "" || payload.CancelUrl == "" {
		http.Error(w, "Missing successUrl or cancelUrl", http.StatusBadRequest)
		return
	}

	userID, err := context.Get(r, "user").(jwt.MapClaims).GetSubject()

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(common.Config.StripePrice1KXP),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String("payment"),
		SuccessURL: stripe.String(payload.SuccessUrl),
		CancelURL:  stripe.String(payload.CancelUrl),
		Metadata: map[string]string{
			"userId": userID,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := response{Url: sess.URL}
	rawResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(rawResponse)
}
