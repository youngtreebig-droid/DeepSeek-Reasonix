package cli

import (
	"fmt"
	"strings"


)

func parseCLIPricingCurrency(value string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "", "AUTO":
		return "", nil
	case "CNY":
		return "CNY", nil
	case "USD":
		return "USD", nil
	default:
		return "", fmt.Errorf("pricing currency %q: must be auto|CNY|USD", value)
	}
}

func pricingCurrencyDisplay(currency string) string {
	if strings.TrimSpace(currency) == "" {
		return "auto"
	}
	return strings.ToUpper(strings.TrimSpace(currency))
}

func describePricingCurrencies(current, resolved string) string {
	items := []string{"auto", "CNY", "USD"}
	var b strings.Builder
	for _, item := range items {
		marker := "  "
		if item == current {
			marker = "• "
		}
		hint := ""
		if item == "auto" {
			hint = " (" + resolved + ")"
		}
		fmt.Fprintf(&b, "%s%s%s\n", marker, item, hint)
	}
	return strings.TrimRight(b.String(), "\n")
}
