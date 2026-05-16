package templates

import "testing"

func TestNormalizePackage(t *testing.T) {
	if got := NormalizePackage("Products"); got != "products" {
		t.Errorf("NormalizePackage(Products) = %q, want products", got)
	}
	if got := NormalizePackage(" order_items "); got != "order_items" {
		t.Errorf("NormalizePackage( order_items ) = %q, want order_items", got)
	}
}

func TestStructName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"products", "Products"},
		{"order_items", "OrderItems"},
		{"user", "User"},
		{"status", "Status"},
		{"address", "Address"},
		{"process", "Process"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := StructName(tt.input); got != tt.want {
				t.Errorf("StructName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPluralStructName(t *testing.T) {
	if got := PluralStructName("products"); got != "Products" {
		t.Errorf("PluralStructName(products) = %q, want Products", got)
	}
}

func TestTableName(t *testing.T) {
	if got := TableName("OrderItems"); got != "orderitems" {
		t.Errorf("TableName(OrderItems) = %q, want orderitems", got)
	}
}
