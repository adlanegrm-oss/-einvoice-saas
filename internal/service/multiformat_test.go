package service

import (
	"context"
	"strings"
	"testing"

	"github.com/adlanegrm-oss/einvoice-saas/internal/validator"
)

func TestCrossBorder_FatturaPA_To_KSeF(t *testing.T) {
	engine := NewMultiFormatEngine()

	fatturaPAXml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<p:FatturaElettronica versione="FPR12" xmlns:p="http://ivaservizi.agenziaentrate.gov.it/docs/xsd/fatture/v1.2">
  <FatturaElettronicaHeader>
    <CedentePrestatore>
      <DatiAnagrafici>
        <IdFiscaleIVA><IdPaese>IT</IdPaese><IdCodice>12345678901</IdCodice></IdFiscaleIVA>
        <Anagrafica><Denominazione>Milano Supplies SRL</Denominazione></Anagrafica>
      </DatiAnagrafici>
      <Sede><Indirizzo>Via Roma 1</Indirizzo><CAP>20121</CAP><Comune>Milano</Comune><Nazione>IT</Nazione></Sede>
    </CedentePrestatore>
    <CessionarioCommittente>
      <DatiAnagrafici>
        <IdFiscaleIVA><IdPaese>PL</IdPaese><IdCodice>9876543210</IdCodice></IdFiscaleIVA>
        <Anagrafica><Denominazione>Warszawa Client Sp. z o.o.</Denominazione></Anagrafica>
      </DatiAnagrafici>
      <Sede><Indirizzo>ul. Marszalkowska 10</Indirizzo><CAP>00-001</CAP><Comune>Warszawa</Comune><Nazione>PL</Nazione></Sede>
    </CessionarioCommittente>
  </FatturaElettronicaHeader>
  <FatturaElettronicaBody>
    <DatiGenerali>
      <DatiGeneraliDocumento>
        <Divisa>EUR</Divisa>
        <Data>2026-10-02</Data>
        <Numero>IT-2026-999</Numero>
        <ImportoTotaleDocumento>1200.00</ImportoTotaleDocumento>
      </DatiGeneraliDocumento>
    </DatiGenerali>
    <DatiBeniServizi>
      <DettaglioLinee>
        <NumeroLinea>1</NumeroLinea>
        <Descrizione>Equipement Serveur</Descrizione>
        <Quantita>1.00</Quantita>
        <PrezzoUnitario>1000.00</PrezzoUnitario>
        <PrezzoTotale>1000.00</PrezzoTotale>
        <AliquotaIVA>20.00</AliquotaIVA>
      </DettaglioLinee>
      <DatiRiepilogo>
        <AliquotaIVA>20.00</AliquotaIVA>
        <ImponibileImporto>1000.00</ImponibileImporto>
        <Imposta>200.00</Imposta>
      </DatiRiepilogo>
    </DatiBeniServizi>
  </FatturaElettronicaBody>
</p:FatturaElettronica>`)

	inv, fmtDetected, err := engine.IngestPayload(context.Background(), fatturaPAXml)
	if err != nil {
		t.Fatalf("Erreur ingestion: %v", err)
	}

	if fmtDetected != validator.FormatFatturaPA {
		t.Fatalf("Format attendu FATTURAPA, obtenu %s", fmtDetected)
	}

	if inv.TotalHT.ToFloat() != 1000.00 || inv.TotalVAT.ToFloat() != 200.00 || inv.TotalTTC.ToFloat() != 1200.00 {
		t.Fatalf("Montants incohérents: HT=%.2f TTC=%.2f", inv.TotalHT.ToFloat(), inv.TotalTTC.ToFloat())
	}

	ksefOutput, err := engine.ConvertToTarget(*inv, validator.FormatKSeF)
	if err != nil {
		t.Fatalf("Erreur export KSeF: %v", err)
	}

	ksefStr := string(ksefOutput)
	if !strings.Contains(ksefStr, "<P_2>IT-2026-999</P_2>") {
		t.Errorf("Numéro de facture manquant dans le flux KSeF")
	}
	if !strings.Contains(ksefStr, "<P_13_1>1000.00</P_13_1>") {
		t.Errorf("Montant net P_13_1 manquant dans KSeF")
	}
	if !strings.Contains(ksefStr, "<P_15>1200.00</P_15>") {
		t.Errorf("Montant TTC P_15 manquant dans KSeF")
	}
}
