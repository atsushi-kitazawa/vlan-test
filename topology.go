package main

import "fmt"

// NetworkTopology ネットワーク全体の構成
type NetworkTopology struct {
	Switch1 *Switch
	Switch2 *Switch
	PCA     *Node // VLAN 10 (営業部)
	PCB     *Node // VLAN 10 (営業部)
	PCC     *Node // VLAN 20 (開発部)
	PCD     *Node // VLAN 20 (開発部)
}

// BuildTopology トポロジーを構築して結線する
func BuildTopology() *NetworkTopology {
	// スイッチ作成
	sw1 := NewSwitch("Switch-1 (1F)")
	sw2 := NewSwitch("Switch-2 (2F)")

	// ポート設定: Switch 1
	// Port 1: Access (VLAN 10)
	// Port 2: Access (VLAN 20)
	// Port 3: Trunk (VLAN 10, 20)
	sw1.AddPort(NewAccessPort(sw1.Name, 1, 10))
	sw1.AddPort(NewAccessPort(sw1.Name, 2, 20))
	sw1.AddPort(NewTrunkPort(sw1.Name, 3, []uint16{10, 20}))

	// ポート設定: Switch 2
	// Port 1: Access (VLAN 10)
	// Port 2: Access (VLAN 20)
	// Port 3: Trunk (VLAN 10, 20)
	sw2.AddPort(NewAccessPort(sw2.Name, 1, 10))
	sw2.AddPort(NewAccessPort(sw2.Name, 2, 20))
	sw2.AddPort(NewTrunkPort(sw2.Name, 3, []uint16{10, 20}))

	// スイッチ間トランクリンクの接続 (Switch 1 Port 3 <---> Switch 2 Port 3)
	sw1.Ports[3].Connect(sw2, 3)
	sw2.Ports[3].Connect(sw1, 3)

	// PC端末作成
	pcA := NewNode("PC-A", "192.168.10.1", "AA:AA:AA:AA:AA:01", "VLAN 10 (営業部)")
	pcB := NewNode("PC-B", "192.168.10.2", "BB:BB:BB:BB:BB:01", "VLAN 10 (営業部)")
	pcC := NewNode("PC-C", "192.168.20.1", "CC:CC:CC:CC:CC:01", "VLAN 20 (開発部)")
	pcD := NewNode("PC-D", "192.168.20.2", "DD:DD:DD:DD:DD:01", "VLAN 20 (開発部)")

	// 端末とスイッチの接続
	// PC-A -> Switch 1 Port 1 (VLAN 10)
	pcA.Connect(sw1, 1)
	// PC-C -> Switch 1 Port 2 (VLAN 20)
	pcC.Connect(sw1, 2)
	// PC-B -> Switch 2 Port 1 (VLAN 10)
	pcB.Connect(sw2, 1)
	// PC-D -> Switch 2 Port 2 (VLAN 20)
	pcD.Connect(sw2, 2)

	return &NetworkTopology{
		Switch1: sw1,
		Switch2: sw2,
		PCA:     pcA,
		PCB:     pcB,
		PCC:     pcC,
		PCD:     pcD,
	}
}

// ResetAllReceived 全端末の受信ログをクリア
func (t *NetworkTopology) ResetAllReceived() {
	t.PCA.ResetReceived()
	t.PCB.ResetReceived()
	t.PCC.ResetReceived()
	t.PCD.ResetReceived()
}

// PrintTopologyDiagram アスキーアートで構成図を表示
func (t *NetworkTopology) PrintTopologyDiagram() {
	fmt.Print(`
===================================================================================================
                                      【 ネットワーク構成図 】
===================================================================================================
      [ VLAN 10 (営業部) ]                                     [ VLAN 20 (開発部) ]
      ┌───────────────────────┐                               ┌───────────────────────┐
      │  PC-A (192.168.10.1)  │                               │  PC-C (192.168.20.1)  │
      │  AA:AA:AA:AA:AA:01    │                               │  CC:CC:CC:CC:CC:01    │
      └───────────┬───────────┘                               └───────────┬───────────┘
                  │ (タグなし / Untagged)                                 │ (タグなし / Untagged)
                  │                                                       │
          [Port 1 │ Access: VLAN 10]                              [Port 2 │ Access: VLAN 20]
      ┌───────────┴───────────────────────────────────────────────────────┴───────────┐
      │                           Switch 1 (1F スイッチ)                              │
      └───────────────────────────────────┬───────────────────────────────────────────┘
                                  [Port 3 │ Trunk: VLAN 10, 20]
                                          │
                                          │  ★ トランク回線 (Trunk Link)
                                          │    802.1Q タグ (VID=10 または 20) が付与されて通過
                                          │
                                  [Port 3 │ Trunk: VLAN 10, 20]
      ┌───────────────────────────────────┴───────────────────────────────────────────┐
      │                           Switch 2 (2F スイッチ)                              │
      └───────────┬───────────────────────────────────────────────────────┬───────────┘
          [Port 1 │ Access: VLAN 10]                              [Port 2 │ Access: VLAN 20]
                  │                                                       │
                  │ (タグなし / Untagged)                                 │ (タグなし / Untagged)
      ┌───────────┴───────────┐                               ┌───────────┴───────────┐
      │  PC-B (192.168.10.2)  │                               │  PC-D (192.168.20.2)  │
      │  BB:BB:BB:BB:BB:01    │                               │  DD:DD:DD:DD:DD:01    │
      └───────────────────────┘                               └───────────────────────┘
      [ VLAN 10 (営業部) ]                                     [ VLAN 20 (開発部) ]
===================================================================================================
`)
}
