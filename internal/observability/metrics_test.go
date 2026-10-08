package observability

import "testing"

func TestMetrics_Counters(t *testing.T) {
m := NewMetrics()
m.IncValidation("tenant-a", true)
m.IncValidation("tenant-a", false)
m.IncTransmit("tenant-a", true)
m.IncExport("tenant-a")
snap := m.Snapshot()
if snap.ValidationOK != 1 || snap.ValidationFail != 1 {
t.Fatalf("validation counters: %+v", snap)
}
if snap.TransmitOK != 1 || snap.ExportOK != 1 {
t.Fatalf("transmit/export: %+v", snap)
}
if snap.TenantsTracked < 1 {
t.Fatal("expected tenant tracked")
}
}
