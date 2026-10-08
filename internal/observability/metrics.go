package observability

import (
	"sync"
	"sync/atomic"
)

// Metrics compteurs process-local (bridge Prometheus possible ensuite).
type Metrics struct {
	mu sync.Mutex

	ValidationOK   atomic.Int64
	ValidationFail atomic.Int64
	ExportOK       atomic.Int64
	TransmitOK     atomic.Int64
	TransmitFail   atomic.Int64

	// par tenant (évite cardinalité explosive: on borne)
	byTenant map[string]*TenantCounters
}

type TenantCounters struct {
	ValidationOK   atomic.Int64
	ValidationFail atomic.Int64
	TransmitOK     atomic.Int64
	TransmitFail   atomic.Int64
}

func NewMetrics() *Metrics {
	return &Metrics{byTenant: make(map[string]*TenantCounters)}
}

func (m *Metrics) tenant(id string) *TenantCounters {
	m.mu.Lock()
	defer m.mu.Unlock()
	if id == "" {
		id = "_unknown"
	}
	if len(m.byTenant) > 10000 {
		// garde-fou cardinalité
		id = "_overflow"
	}
	tc, ok := m.byTenant[id]
	if !ok {
		tc = &TenantCounters{}
		m.byTenant[id] = tc
	}
	return tc
}

func (m *Metrics) IncValidation(tenantID string, ok bool) {
	if ok {
		m.ValidationOK.Add(1)
		m.tenant(tenantID).ValidationOK.Add(1)
	} else {
		m.ValidationFail.Add(1)
		m.tenant(tenantID).ValidationFail.Add(1)
	}
}

func (m *Metrics) IncExport(tenantID string) {
	m.ExportOK.Add(1)
}

func (m *Metrics) IncTransmit(tenantID string, ok bool) {
	if ok {
		m.TransmitOK.Add(1)
		m.tenant(tenantID).TransmitOK.Add(1)
	} else {
		m.TransmitFail.Add(1)
		m.tenant(tenantID).TransmitFail.Add(1)
	}
}

// Snapshot pour endpoint /metrics JSON simple.
type Snapshot struct {
	ValidationOK   int64 `json:"validation_ok"`
	ValidationFail int64 `json:"validation_fail"`
	ExportOK       int64 `json:"export_ok"`
	TransmitOK     int64 `json:"transmit_ok"`
	TransmitFail   int64 `json:"transmit_fail"`
	TenantsTracked int   `json:"tenants_tracked"`
}

func (m *Metrics) Snapshot() Snapshot {
	m.mu.Lock()
	n := len(m.byTenant)
	m.mu.Unlock()
	return Snapshot{
		ValidationOK:   m.ValidationOK.Load(),
		ValidationFail: m.ValidationFail.Load(),
		ExportOK:       m.ExportOK.Load(),
		TransmitOK:     m.TransmitOK.Load(),
		TransmitFail:   m.TransmitFail.Load(),
		TenantsTracked: n,
	}
}
