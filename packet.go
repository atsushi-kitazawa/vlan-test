package main

import (
	"fmt"
	"strings"
)

// EtherType 定義
const (
	EtherTypeIPv4 = 0x0800
	EtherTypeARP  = 0x0806
	EtherType8021Q = 0x8100 // IEEE 802.1Q VLAN Tag
)

// IEEE 802.1Q VLANタグ構造
// イーサネットフレームの送信元MACとEtherTypeの間に4バイト挿入される
type VLANHeader struct {
	TPID uint16 // Tag Protocol Identifier: 常に 0x8100
	PCP  uint8  // Priority Code Point (3bit): QoSの優先度 (0-7)
	DEI  bool   // Drop Eligible Indicator (1bit): 輻輳時の破棄可否
	VID  uint16 // VLAN Identifier (12bit): VLAN ID (1〜4094)
}

// イーサネットフレーム（Layer 2）
type EthernetFrame struct {
	DstMAC    string      // 宛先MACアドレス (例: "AA:AA:AA:AA:AA:01", ブロードキャスト: "FF:FF:FF:FF:FF:FF")
	SrcMAC    string      // 送信元MACアドレス
	VLANTag   *VLANHeader // nil の場合は Untagged (タグなし)、存在する場合は Tagged (802.1Qタグ付き)
	EtherType uint16      // ペイロードの種類 (IPv4: 0x0800, ARP: 0x0806 など)
	Payload   string      // 運ぶデータ（簡易的に文字列で表現）
}

// IsBroadcast 宛先がブロードキャストアドレスかどうか
func (f *EthernetFrame) IsBroadcast() bool {
	return strings.ToUpper(f.DstMAC) == "FF:FF:FF:FF:FF:FF"
}

// Clone フレームの複製を作成
func (f *EthernetFrame) Clone() *EthernetFrame {
	var tagCopy *VLANHeader
	if f.VLANTag != nil {
		tagCopy = &VLANHeader{
			TPID: f.VLANTag.TPID,
			PCP:  f.VLANTag.PCP,
			DEI:  f.VLANTag.DEI,
			VID:  f.VLANTag.VID,
		}
	}
	return &EthernetFrame{
		DstMAC:    f.DstMAC,
		SrcMAC:    f.SrcMAC,
		VLANTag:   tagCopy,
		EtherType: f.EtherType,
		Payload:   f.Payload,
	}
}

// VisualString フレームの構造を視覚的に分かりやすく文字列化
func (f *EthernetFrame) VisualString() string {
	var sb strings.Builder

	etherTypeStr := fmt.Sprintf("0x%04X", f.EtherType)
	switch f.EtherType {
	case EtherTypeIPv4:
		etherTypeStr += "(IPv4)"
	case EtherTypeARP:
		etherTypeStr += "(ARP)"
	}

	if f.VLANTag == nil {
		// Untagged Frame
		sb.WriteString("┌──────────────┬──────────────┬──────────────┬────────────────────────────┐\n")
		sb.WriteString(fmt.Sprintf("│ Dst: %-7s │ Src: %-7s │ Type: %-6s │ Payload: %-18s │  [Untagged / 標準フレーム]\n",
			f.DstMAC, f.SrcMAC, etherTypeStr, truncate(f.Payload, 18)))
		sb.WriteString("└──────────────┴──────────────┴──────────────┴────────────────────────────┘")
	} else {
		// Tagged Frame (IEEE 802.1Q)
		vlanStr := fmt.Sprintf("VID:%d", f.VLANTag.VID)
		sb.WriteString("┌──────────────┬──────────────┬───────────────┬──────────────┬────────────────────────────┐\n")
		sb.WriteString(fmt.Sprintf("│ Dst: %-7s │ Src: %-7s │ \033[1;33m802.1Q %-6s\033[0m │ Type: %-6s │ Payload: %-18s │  \033[1;33m[Tagged / VLANタグ付き]\033[0m\n",
			f.DstMAC, f.SrcMAC, vlanStr, etherTypeStr, truncate(f.Payload, 18)))
		sb.WriteString("└──────────────┴──────────────┴───────────────┴──────────────┴────────────────────────────┘")
	}

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
