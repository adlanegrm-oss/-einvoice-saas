package invoice

import (
"encoding/json"
"fmt"
"math"
)

type Currency string

const (
CurrencyEUR Currency = "EUR"
)

type Money struct {
Amount   int64    `json:"amount"`
Scale    int32    `json:"scale"`
Currency Currency `json:"currency"`
}

func NewMoney(amount int64, scale int32, cur Currency) Money {
if cur == "" {
cur = CurrencyEUR
}
return Money{
Amount:   amount,
Scale:    scale,
Currency: cur,
}
}

func NewMoneyFromCents(cents int64, cur Currency) Money {
return NewMoney(cents, 2, cur)
}

func NewMoneyFromFloat(val float64, scale int32, cur Currency) Money {
mult := math.Pow10(int(scale))
cents := int64(math.Round(val * mult))
return NewMoney(cents, scale, cur)
}

func (m Money) ToFloat() float64 {
div := math.Pow10(int(m.Scale))
return float64(m.Amount) / div
}

func (m Money) Add(other Money) (Money, error) {
if m.Currency != "" && other.Currency != "" && m.Currency != other.Currency {
return Money{}, fmt.Errorf("devises incompatibles: %s != %s", m.Currency, other.Currency)
}

maxScale := m.Scale
if other.Scale > maxScale {
maxScale = other.Scale
}

cur := m.Currency
if cur == "" {
cur = other.Currency
}

return Money{
Amount:   m.rescale(maxScale) + other.rescale(maxScale),
Scale:    maxScale,
Currency: cur,
}, nil
}

func (m Money) Sub(other Money) (Money, error) {
if m.Currency != "" && other.Currency != "" && m.Currency != other.Currency {
return Money{}, fmt.Errorf("devises incompatibles: %s != %s", m.Currency, other.Currency)
}

maxScale := m.Scale
if other.Scale > maxScale {
maxScale = other.Scale
}

cur := m.Currency
if cur == "" {
cur = other.Currency
}

return Money{
Amount:   m.rescale(maxScale) - other.rescale(maxScale),
Scale:    maxScale,
Currency: cur,
}, nil
}

func (m Money) RoundTo2() Money {
if m.Scale <= 2 {
return m.ToScale(2)
}

diff := m.Scale - 2
factor := int64(math.Pow10(int(diff)))
half := factor / 2

var rounded int64
if m.Amount >= 0 {
rounded = (m.Amount + half) / factor
} else {
rounded = (m.Amount - half) / factor
}

return Money{
Amount:   rounded,
Scale:    2,
Currency: m.Currency,
}
}

func (m Money) ToScale(targetScale int32) Money {
return Money{
Amount:   m.rescale(targetScale),
Scale:    targetScale,
Currency: m.Currency,
}
}

func (m Money) rescale(targetScale int32) int64 {
if m.Scale == targetScale {
return m.Amount
}
if targetScale > m.Scale {
return m.Amount * int64(math.Pow10(int(targetScale-m.Scale)))
}
return m.Amount / int64(math.Pow10(int(m.Scale-targetScale)))
}

func (m *Money) UnmarshalJSON(data []byte) error {
var val float64
if err := json.Unmarshal(data, &val); err == nil {
*m = NewMoneyFromFloat(val, 2, CurrencyEUR)
return nil
}

type alias Money
var obj alias
if err := json.Unmarshal(data, &obj); err != nil {
return err
}
*m = Money(obj)
return nil
}

func (m Money) MarshalJSON() ([]byte, error) {
return json.Marshal(m.ToFloat())
}
