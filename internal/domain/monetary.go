package domain

import (
"errors"
"fmt"
)

type Amount int64

func NewAmount(units int64, cents int64) Amount {
return Amount(units*100 + cents)
}

func FromCents(cents int64) Amount {
return Amount(cents)
}

func (a Amount) Cents() int64 {
return int64(a)
}

func (a Amount) ToUnits() float64 {
return float64(a) / 100.0
}

func (a Amount) String() string {
units := a / 100
cents := a % 100
if cents < 0 {
cents = -cents
}
return fmt.Sprintf("%d.%02d", units, cents)
}

func (a Amount) Add(b Amount) Amount {
return a + b
}

func (a Amount) Sub(b Amount) Amount {
return a - b
}

func (a Amount) CalculateTax(taxBasisPoints int64) Amount {
product := int64(a) * taxBasisPoints
quotient := product / 10000
remainder := product % 10000

if remainder < 0 {
remainder = -remainder
}

if remainder > 5000 {
if product > 0 {
quotient++
} else {
quotient--
}
} else if remainder == 5000 {
if quotient%2 != 0 {
if product > 0 {
quotient++
} else {
quotient--
}
}
}

return Amount(quotient)
}

type MonetaryTotals struct {
NetTotal   Amount `json:"net_total_cents"`
TaxTotal   Amount `json:"tax_total_cents"`
GrossTotal Amount `json:"gross_total_cents"`
Prepaid    Amount `json:"prepaid_cents"`
Payable    Amount `json:"payable_cents"`
}

func (m MonetaryTotals) Validate() error {
expectedGross := m.NetTotal.Add(m.TaxTotal)
if m.GrossTotal != expectedGross {
return fmt.Errorf("incohérence monétaire : Net (%s) + Tax (%s) != Gross (%s)",
m.NetTotal, m.TaxTotal, m.GrossTotal)
}
expectedPayable := m.GrossTotal.Sub(m.Prepaid)
if m.Payable != expectedPayable {
return errors.New("le montant dû ne correspond pas au total TTC déduit des acomptes")
}
return nil
}
