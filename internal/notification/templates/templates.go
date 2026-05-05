// Package templates holds tiny message templates. Replace with html/template
// + i18n when copywriting matures.
package templates

import "fmt"

// OrderPaid renders a confirmation message body.
func OrderPaid(orderID string, totalMinor int64, currency string) (subject, body string) {
	return "Your order " + orderID + " is confirmed",
		fmt.Sprintf("Order %s paid: %d %s. Tickets attached.", orderID, totalMinor, currency)
}

// PaymentFailed renders a failure-notice body.
func PaymentFailed(orderID string) (subject, body string) {
	return "Payment failed for order " + orderID,
		"We couldn't process your payment. Your hold has been released."
}
