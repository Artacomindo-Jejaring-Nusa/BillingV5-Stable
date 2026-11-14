package hsgqclient

import (
	"testing"
)

func TestParseONUInfoOutput(t *testing.T) {
	sampleOutput := `
----------------------------------------------------------------------------------------------------
 PON/ONU     Mac-Address    Status   Auth  Cfg     Reg-time            ONU-Name     ONU-Desc
----------------------------------------------------------------------------------------------------
 1/1     70:2e:22:0a:59:47 Online    TRUE  TRUE  2026/10/01 02:06:04      PRM-A2-26-Ar NO-DESCRIPTI
 1/2     70:2e:22:0a:b7:4d Offline   TRUE  TRUE  2026/08/10 18:06:02      PRM-A2-24-De NO-DESCRIPTI
 1/17    70:2e:22:0a:63:8f Initial   TRUE  FALSE  -                        PRM-A3-15-Ri NO-DESCRIPTI
 4/1     70:2e:22:0a:de:00 Online    TRUE  TRUE  2026/10/29 10:02:29      colocation   NO-DESCRIPTI
 4/2     70:2e:22:0a:46:0e Online    TRUE  TRUE  2026/11/12 18:49:05      ONU04/02     NO-DESCRIPTI
----------------------------------------------------------------------------------------------------
  Total: 5  Online:3  Offline:1
----------------------------------------------------------------------------------------------------
`

	onus := parseONUInfoOutput(sampleOutput)
	if len(onus) != 5 {
		t.Fatalf("expected 5 ONUs, got %d", len(onus))
	}

	first := onus[0]
	if first.PON != 1 || first.ONUID != 1 {
		t.Errorf("expected 1/1, got %d/%d", first.PON, first.ONUID)
	}
	if first.SerialNumber != "70:2e:22:0a:59:47" {
		t.Errorf("expected MAC 70:2e:22:0a:59:47, got %s", first.SerialNumber)
	}
	if first.Name != "PRM-A2-26-Ar" {
		t.Errorf("expected name PRM-A2-26-Ar, got %s", first.Name)
	}
	if first.Status != "Online" {
		t.Errorf("expected Online, got %s", first.Status)
	}

	last := onus[4]
	if last.PON != 4 || last.ONUID != 2 {
		t.Errorf("expected 4/2, got %d/%d", last.PON, last.ONUID)
	}
	if last.SerialNumber != "70:2e:22:0a:46:0e" {
		t.Errorf("expected MAC 70:2e:22:0a:46:0e, got %s", last.SerialNumber)
	}
	if last.Name != "ONU04/02" {
		t.Errorf("expected name ONU04/02, got %s", last.Name)
	}
}

func TestParseOpticalRSSIOutput(t *testing.T) {
	sampleOutput := `
  PON ONU RSSI Informations:
      Temperature      :43.875000 C
      Voltage          :3.288000 V
      Bias             :14.532000 mA
      Transmit Power   :7.663757 dBm
      Recieved Power   :-29.586073 dBm
`

	detail := parseOpticalRSSIOutput(sampleOutput, 1, 1)
	if detail.RxPower != "-29.59 dBm" {
		t.Errorf("expected RxPower -29.59 dBm, got %s", detail.RxPower)
	}
	if detail.TxPower != "7.66 dBm" {
		t.Errorf("expected TxPower 7.66 dBm, got %s", detail.TxPower)
	}
	if detail.Description == "" {
		t.Errorf("expected description to have temperature and voltage, got empty")
	}
}

