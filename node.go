package main

import (
	"fmt"
)

// Node ネットワーク端末（PC）
type Node struct {
	Name            string
	IP              string
	MAC             string
	VLANDescription string // 所属組織の説明 (例: "開発部 VLAN 10")
	ConnectedSwitch *Switch
	ConnectedPortID int
	ReceivedFrames  []*EthernetFrame
}

// NewNode 新規端末作成
func NewNode(name, ip, mac, vlanDesc string) *Node {
	return &Node{
		Name:            name,
		IP:              ip,
		MAC:             mac,
		VLANDescription: vlanDesc,
		ReceivedFrames:  make([]*EthernetFrame, 0),
	}
}

func (n *Node) GetName() string {
	return fmt.Sprintf("%s (%s / %s)", n.Name, n.IP, n.VLANDescription)
}

// Connect スイッチのポートに接続
func (n *Node) Connect(sw *Switch, portID int) {
	n.ConnectedSwitch = sw
	n.ConnectedPortID = portID

	port := sw.Ports[portID]
	port.Connect(n, 0)
}

// SendUnicast 他の端末宛てにユニキャスト（通常のパケット）を送信
func (n *Node) SendUnicast(dstMAC, message string) {
	fmt.Printf("\n\033[1;32m┌────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ 💻 [%s] がパケットを送信します\n", n.Name)
	fmt.Printf("│    送信元: %s (%s)\n", n.Name, n.MAC)
	fmt.Printf("│    宛先  : %s\n", dstMAC)
	fmt.Printf("│    内容  : \"%s\"\n", message)
	fmt.Printf("└────────────────────────────────────────────────────────┘\033[0m\n")

	// PCはVLANタグを知らないので、Untaggedフレームとして送出
	frame := &EthernetFrame{
		DstMAC:    dstMAC,
		SrcMAC:    n.MAC,
		VLANTag:   nil, // PC端末はタグを付けない（普通のEthernetフレーム）
		EtherType: EtherTypeIPv4,
		Payload:   message,
	}

	fmt.Println("【送信されるフレーム構造 (PCからはタグなしで送出)】:")
	fmt.Println(frame.VisualString())

	// スイッチへ渡す
	n.ConnectedSwitch.ReceiveFrame(frame, n.ConnectedPortID)
}

// SendBroadcast ARPなどのブロードキャストを送信
func (n *Node) SendBroadcast(queryMessage string) {
	fmt.Printf("\n\033[1;33m┌────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ 📢 [%s] がブロードキャスト (ARP等) を送信します\n", n.Name)
	fmt.Printf("│    送信元: %s (%s)\n", n.Name, n.MAC)
	fmt.Printf("│    宛先  : FF:FF:FF:FF:FF:FF (全端末宛て)\n")
	fmt.Printf("│    内容  : \"%s\"\n", queryMessage)
	fmt.Printf("└────────────────────────────────────────────────────────┘\033[0m\n")

	frame := &EthernetFrame{
		DstMAC:    "FF:FF:FF:FF:FF:FF",
		SrcMAC:    n.MAC,
		VLANTag:   nil,
		EtherType: EtherTypeARP,
		Payload:   queryMessage,
	}

	fmt.Println("【送信されるフレーム構造 (ブロードキャスト)】:")
	fmt.Println(frame.VisualString())

	// スイッチへ渡す
	n.ConnectedSwitch.ReceiveFrame(frame, n.ConnectedPortID)
}

// ReceiveFrame フレームを受信したときの処理
func (n *Node) ReceiveFrame(frame *EthernetFrame, fromPortID int) {
	n.ReceivedFrames = append(n.ReceivedFrames, frame)

	// 受信確認
	fmt.Printf("      \033[1;42;30m SUCCESS \033[0m \033[1m[%s] がフレームを受信しました！\033[0m\n", n.Name)
	if frame.VLANTag != nil {
		fmt.Printf("      ⚠️ 警告: エンド端末なのにタグ付きフレームが届きました (不正)\n")
	} else {
		fmt.Printf("      ↳ タグなしフレームとして正常に着信。ペイロード: \"%s\"\n", frame.Payload)
	}
}

// ResetReceived 受信履歴クリア
func (n *Node) ResetReceived() {
	n.ReceivedFrames = make([]*EthernetFrame, 0)
}

func (n *Node) HasReceived() bool {
	return len(n.ReceivedFrames) > 0
}
