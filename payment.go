package main

import (
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

// InitStripe initialise la clé secrète Stripe.
// Appelle cette fonction dans ton `main` au démarrage.
func InitStripe() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY") // Ta clé privée ici
}

// HandleCheckout crée une session Stripe Checkout et redirige l'utilisateur.
func HandleCheckout(w http.ResponseWriter, r *http.Request) {
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String("price_1RPHv4B3bzBJncSGAN0FG0Kf"),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String("payment"),
		SuccessURL: stripe.String("http://localhost:4242/success"),
		CancelURL:  stripe.String("http://localhost:4242/cancel"),
	}

	sess, err := session.New(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	

	http.Redirect(w, r, sess.URL, http.StatusSeeOther)
}

// RegisterRoutes enregistre les routes de paiement.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/create-checkout-session", HandleCheckout)

	mux.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Paiement réussi !"))
	})

	mux.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Paiement annulé."))
	})
}
