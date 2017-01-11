package utils

import (
	"testing"
)

func TestHumanizeString(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"API", "API"},
		{"OrderID", "Order ID"},
		{"OrderItem", "Order Item"},
		{"orderItem", "Order Item"},
		{"OrderIDItem", "Order ID Item"},
		{"OrderItemID", "Order Item ID"},
		{"VIEW SITE", "VIEW SITE"},
		{"Order Item", "Order Item"},
		{"Order ITEM", "Order ITEM"},
		{"ORDER Item", "ORDER Item"},
	}

	for _, c := range cases {
		if got := HumanizeString(c.input); got != c.want {
			t.Errorf("HumanizeString(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}
