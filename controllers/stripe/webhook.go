package stripeController

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"server/common"
	"server/models"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

func HandleWebhook(w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event := stripe.Event{}

	// Replace this endpoint secret with your endpoint's unique secret
	// If you are testing with the CLI, find the secret by running 'stripe listen'
	// If you are using an endpoint defined with the API or dashboard, look in your webhook settings
	// at https://dashboard.stripe.com/webhooks
	signatureHeader := req.Header.Get("Stripe-Signature")
	event, err = webhook.ConstructEvent(payload, signatureHeader, common.Config.StripeWebhookSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook signature verification failed. %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}

	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "checkout.session.completed":
		tx, err := common.Db.Begin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting transaction: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var paymentIntent stripe.PaymentIntent
		err = json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("user %s purchased 1000 experience points", paymentIntent.Metadata["userId"])
		user, err := models.FetchOneUserByCloudIamSub(tx, paymentIntent.Metadata["userId"])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching user: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		user.Rank += 1.0
		err = models.Update(tx, user)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating user: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = tx.Commit()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error committing transaction: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
