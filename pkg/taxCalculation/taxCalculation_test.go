package taxCalculation

import (
	"testing"
)

func TestIncomeTax(t *testing.T) {
	tests := []struct {
		income      float64
		expectedTax float64
	}{
		{500000, 126728},
		{300000, 74728},
	}

	for _, test := range tests {
		tax := IncomeTax(test.income)
		if tax != test.expectedTax {
			t.Errorf("IncomeTax(%f) = %f; expected %f", test.income, tax, test.expectedTax)
		}
	}
}
