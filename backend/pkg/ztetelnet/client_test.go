package ztetelnet

import (
	"testing"
)

func TestParseZTECardsOutput(t *testing.T) {
	sample := `
Rack Shelf Slot CfgType RealType Port  HardVer  SoftVer         Status
-------------------------------------------------------------------------------
1    1     1    GTGH    GTGHG    16    120400   V1.2.5P3        INSERVICE
1    1     2    GTGH    GTGHG    16    120400   V1.2.5P3        INSERVICE
1    1     3    SMXA    SMXA     3     120501   V1.2.5P3        INSERVICE
1    1     4    SMXA    SMXA     3     120501   V1.2.5P3        STANDBY
`
	cards := parseZTECardsOutput(sample)
	if len(cards) != 4 {
		t.Fatalf("expected 4 cards, got %d", len(cards))
	}

	if cards[0].Slot != 1 || cards[0].Type != "GTGH" || cards[0].Role != "Line Card" {
		t.Errorf("unexpected card 0: %+v", cards[0])
	}
	if cards[2].Slot != 3 || cards[2].Type != "SMXA" || cards[2].Role != "Control & Switch Card" {
		t.Errorf("unexpected card 2: %+v", cards[2])
	}
}

func TestParseZTEPortsOutput(t *testing.T) {
	sample := `
Port             Admin-Status   Oper-Status   Speed(Mbps)
---------------------------------------------------------
gei_1/3/1        up             up            1000
xgei_1/4/1       up             up            10000
`
	ports := parseZTEPortsOutput(sample)
	if len(ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(ports))
	}
	if ports[0].Name != "gei_1/3/1" || ports[0].Kind != "GE" || ports[0].SpeedMbps != 1000 {
		t.Errorf("unexpected port 0: %+v", ports[0])
	}
	if ports[1].Name != "xgei_1/4/1" || ports[1].Kind != "10GE" || ports[1].SpeedMbps != 10000 {
		t.Errorf("unexpected port 1: %+v", ports[1])
	}
}

func TestParseZTEONUs(t *testing.T) {
	outState := `
OnuIndex              Admin State    OMCC State    Phase State    Channel 
-------------------------------------------------------------------------
gpon-onu_1/1/1:1      enable         enable        working        1
gpon-onu_1/1/1:2      enable         enable        los            1
gpon-onu_1/1/1:3      enable         enable        dyinggasp      1
`
	outConfig := `
interface gpon-olt_1/1/1
  onu 1 type ZTE-F660 sn ZTEGC70A5947
  onu 1 name Customer-A
  onu 2 type ZTE-F670 sn ZTEGC70AB74D
  onu 2 name Customer-B
`
	outBase := `
OnuIndex              Type           SN                  State
--------------------------------------------------------------------
gpon-onu_1/1/1:1      ZTE-F660       ZTEGC70A5947        ready
gpon-onu_1/1/1:2      ZTE-F670       ZTEGC70AB74D        ready
gpon-onu_1/1/1:3      ZTE-F609       ZTEGC70A8165        ready
`
	onus := parseZTEONUs(1, 1, outState, outConfig, outBase)
	if len(onus) != 3 {
		t.Fatalf("expected 3 ONUs, got %d", len(onus))
	}

	// ONU 1: Online
	if onus[0].ONUID != 1 || onus[0].Name != "Customer-A" || onus[0].Status != "Online" || onus[0].SerialNumber != "ZTEGC70A5947" {
		t.Errorf("unexpected ONU 1: %+v", onus[0])
	}

	// ONU 2: Offline (los)
	if onus[1].ONUID != 2 || onus[1].Name != "Customer-B" || onus[1].Status != "Offline" || onus[1].SerialNumber != "ZTEGC70AB74D" {
		t.Errorf("unexpected ONU 2: %+v", onus[1])
	}

	// ONU 3: Offline (dyinggasp)
	if onus[2].ONUID != 3 || onus[2].Status != "Offline" || onus[2].SerialNumber != "ZTEGC70A8165" {
		t.Errorf("unexpected ONU 3: %+v", onus[2])
	}
}

func TestParseZTEONUDetailOutput(t *testing.T) {
	outPower := `
              OLT Rx(dBm)        ONU Rx(dBm)
-------------------------------------------------------------
up            -21.432            -
down          -                  -22.150
`
	outDist := `
Optical Distance: 1420m
`
	outDetail := `
Name: Customer-A
SN: ZTEGC70A5947
Type: ZTE-F660
Phase State: working
`
	detail := parseZTEONUDetailOutput(1, 1, 1, outPower, outDist, outDetail)
	if detail.RxPower != "-22.15 dBm" {
		t.Errorf("expected RxPower -22.15 dBm, got %s", detail.RxPower)
	}
	if detail.TxPower != "-21.43 dBm" {
		t.Errorf("expected TxPower -21.43 dBm, got %s", detail.TxPower)
	}
	if detail.GponOpticalDistance != "1420" {
		t.Errorf("expected GponOpticalDistance 1420, got %s", detail.GponOpticalDistance)
	}
	if detail.Status != "Online" {
		t.Errorf("expected Status Online, got %s", detail.Status)
	}
	if detail.Name != "Customer-A" {
		t.Errorf("expected Name Customer-A, got %s", detail.Name)
	}
}

func TestParseZTEONUs_RealTiparCLI(t *testing.T) {
	outState := `
OnuIndex   Admin State  OMCC State  Phase State  Channel    
--------------------------------------------------------------
1/1/1:1     enable       enable      working      1(GPON)
1/1/1:2     enable       enable      working      1(GPON)
1/1/1:3     enable       disable     LOS          1(GPON)
1/1/1:4     enable       enable      working      1(GPON)
1/1/2:1     enable       enable      working      1(GPON)
1/1/10:1    enable       enable      working      1(GPON)
1/1/10:2    enable       enable      working      1(GPON)
1/1/10:3    enable       disable     LOS          1(GPON)
1/1/10:4    enable       enable      working      1(GPON)
1/1/10:5    enable       enable      working      1(GPON)
`
	outConfigPon1 := `
interface gpon-olt_1/1/1
  onu 1 type ZTE-F660 sn RTEGC69E3F18
  onu 1 name CKG-TPR-AKS-30-107-0224-Junus
  onu 2 type ALL sn RTEGC60AD1A8
  onu 3 type ZTE-F660 sn ZTEGC041E7C6
  onu 3 name CKG-TPR-AKS-20-0411-0623-Eva
  onu 4 type ALL sn RTEGC60AD556
  onu 4 name CKG-TPR-AKS-112-Rika
`
	onusPon1 := parseZTEONUs(1, 1, outState, outConfigPon1, "")
	if len(onusPon1) != 4 {
		t.Fatalf("expected 4 ONUs on PON 1, got %d", len(onusPon1))
	}
	if onusPon1[0].Status != "Online" {
		t.Errorf("expected ONU 1 to be Online, got %s", onusPon1[0].Status)
	}
	if onusPon1[1].Status != "Online" {
		t.Errorf("expected ONU 2 to be Online, got %s", onusPon1[1].Status)
	}
	if onusPon1[2].Status != "Offline" {
		t.Errorf("expected ONU 3 to be Offline, got %s", onusPon1[2].Status)
	}
	if onusPon1[3].Status != "Online" {
		t.Errorf("expected ONU 4 to be Online, got %s", onusPon1[3].Status)
	}

	outConfigPon10 := `
interface gpon-olt_1/1/10
  onu 1 type ALL sn MHEGCDEEA884
  onu 2 type ALL sn MHEGCDEA97D8
  onu 3 type ALL sn MHEGCDE9ADF9
  onu 4 type ALL sn RTEGC60A1E7E
  onu 4 name CKG-BRT-F-308-Deasy
  onu 5 type ZTE-F660 sn ZTEGC2D40FF7
  onu 5 name CKG-BRT-F-105-20-0723-Alfin
`
	onusPon10 := parseZTEONUs(1, 10, outState, outConfigPon10, "")
	if len(onusPon10) != 5 {
		t.Fatalf("expected 5 ONUs on PON 10, got %d", len(onusPon10))
	}
	if onusPon10[0].Status != "Online" {
		t.Errorf("expected PON 10 ONU 1 to be Online, got %s", onusPon10[0].Status)
	}
	if onusPon10[1].Status != "Online" {
		t.Errorf("expected PON 10 ONU 2 to be Online, got %s", onusPon10[1].Status)
	}
	if onusPon10[2].Status != "Offline" {
		t.Errorf("expected PON 10 ONU 3 to be Offline (LOS), got %s", onusPon10[2].Status)
	}
	if onusPon10[3].Status != "Online" {
		t.Errorf("expected PON 10 ONU 4 to be Online, got %s", onusPon10[3].Status)
	}
	if onusPon10[4].Status != "Online" {
		t.Errorf("expected PON 10 ONU 5 to be Online, got %s", onusPon10[4].Status)
	}
}

