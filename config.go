package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	MapSize       uint32
	NflogGroup    uint16
	VPNIP         uint32
	InternalNet   uint32
	InternalMask  uint32
	InputMarkVPN  uint32
	OutputMarkVPN uint32
}

type HexUint32 uint32

type HexConfig struct {
	XMLName       xml.Name  `xml:"config"`
	MapSize       uint32    `xml:"map_size"`
	NflogGroup    uint16    `xml:"nflog_group"`
	VPNIP         HexUint32 `xml:"vpn_ip"`
	InternalNet   HexUint32 `xml:"internal_net"`
	InternalMask  HexUint32 `xml:"internal_mask"`
	InputMarkVPN  HexUint32 `xml:"input_mark"`
	OutputMarkVPN HexUint32 `xml:"output_mark"`
}

func (h *HexUint32) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}

	s = strings.TrimPrefix(s, "0x")

	val, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return err
	}

	*h = HexUint32(val)
	return nil
}

func LoadConfig(path string) (cfg Config, err error) {
	content, freadErr := os.ReadFile(path)
	if freadErr != nil {
		err = fmt.Errorf("failed to read config file: %v", freadErr)
		return
	}

	hexCfg := HexConfig{}

	unmrshlErr := xml.Unmarshal(content, &hexCfg)
	if unmrshlErr != nil {
		err = fmt.Errorf("failed to parse config file: %v", unmrshlErr)
		return
	}

	cfg = Config{
		MapSize:       hexCfg.MapSize,
		NflogGroup:    hexCfg.NflogGroup,
		VPNIP:         uint32(hexCfg.VPNIP),
		InternalNet:   uint32(hexCfg.InternalNet),
		InternalMask:  uint32(hexCfg.InternalMask),
		InputMarkVPN:  uint32(hexCfg.InputMarkVPN),
		OutputMarkVPN: uint32(hexCfg.OutputMarkVPN),
	}

	return
}
