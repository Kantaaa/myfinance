package taxCalculation

import "fmt"

const (
	BaseTaxRate                 = 0.22  // 110 000
	NICRate                     = 0.078 //National Insurance Contributions
	MaxNICRate                  = 69650 //max 69 650
	MiniumStandardDeductionRate = 104450
	HiddenBracketRate           = 0.10  // 4200			{123 928}
	FirstBracketRate            = 0.017 // 1442
	SecondBracketRate           = 0.04  // 8 286
	ThirdBracketRate            = 0.136 // 51292,4
	FourthBracketRate           = 0.166 // 10 308,6
	FifthBracketRate            = 0.176
)

//Personfradrag

const (
	HiddenBracketThreshold = 70000
	FirstBracketThreshold  = 208050
	SecondBracketThreshold = 292850
	ThirdBracketThreshold  = 670000
	FourthBracketThreshold = 937900
	FifthBracketThreshold  = 1350000
)

type TaxBracket struct {
	threshold float64
	rate      float64
}

var taxBrackets = []TaxBracket{
	{FirstBracketThreshold, FirstBracketRate},
	{SecondBracketThreshold, SecondBracketRate},
	{ThirdBracketThreshold, ThirdBracketRate},
	{FourthBracketThreshold, FourthBracketRate},
	{FifthBracketThreshold, FifthBracketRate},
}

func CalcHiddenTax(income float64) float64 {
	if !(income > HiddenBracketThreshold) {
		return income * HiddenBracketRate
	}
	return HiddenBracketThreshold * HiddenBracketRate

}

func CalcMinimumStandardDeduction(income float64) float64 {

	MSDBasedOnIncome := income * 0.46

	if (MSDBasedOnIncome) < MiniumStandardDeductionRate {
		return MSDBasedOnIncome
	} else {
		return MiniumStandardDeductionRate
	}

}

func CalcBaseTax(income float64) float64 {
	//Bruttolønn - Personfradrag
	TaxFoundation := income - MiniumStandardDeductionRate
	if income >= FirstBracketThreshold {
		return (TaxFoundation * BaseTaxRate)
	}
	return 0
}

func CalcBracketTax(income float64) float64 {
	tax := 0.0

	for i := 0; i < len(taxBrackets); i++ {
		curBracket := taxBrackets[i]

		if i < len(taxBrackets)-1 {
			nextBracket := taxBrackets[i+1]

			if income > curBracket.threshold && income >= nextBracket.threshold {
				nextBracket := taxBrackets[i+1]
				tax += (nextBracket.threshold - curBracket.threshold) * curBracket.rate

				fmt.Printf("Processing bracket: %+v\n", curBracket)
				fmt.Printf("Current tax: %f\n", tax)
			} else {
				tax += (income - curBracket.threshold) * curBracket.rate
				fmt.Printf("Processing bracket: %+v\n", curBracket)
				fmt.Printf("Current tax: %f\n", tax)
				break
			}
		}

	}

	return tax
}

func CalcNationInsurance(income float64) float64 {

	if income > MaxNICRate {
		return NICRate * income
	}
	return 0
}

func IncomeTax(income float64) float64 {
	tax := 0.0

	//tax += CalcHiddenTax(income)

	tax += CalcBracketTax(income)

	tax += CalcBaseTax(income)

	tax += CalcNationInsurance(income)

	return tax
}

func EffectiveTax(income float64) float64 {
	taxedAmount := IncomeTax(income)
	effectiveTax := (taxedAmount / income) * 100
	return effectiveTax
}
