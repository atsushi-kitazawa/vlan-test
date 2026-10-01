package main

import (
	"fmt"
	"strings"
	"time"
)

// MACEntry MACアドレステーブルの1エントリ
type MACEntry struct {
	MAC       string
	VLANID    uint16
	PortID    int
	LearnedAt time.Time
}

// Switch L2スイッチ
type Switch struct {
	Name      string
	Ports     map[int]*Port
	MACTable  map[string]*MACEntry // キー: "MAC_VLANID" (IVL: Independent VLAN Learning)
	LogIndent string
}

// NewSwitch 新規スイッチ作成
func NewSwitch(name string) *Switch {
	return &Switch{
		Name:      name,
		Ports:     make(map[int]*Port),
		MACTable:  make(map[string]*MACEntry),
		LogIndent: "  ",
	}
}

func (s *Switch) GetName() string {
	return s.Name
}

// AddPort スイッチにポートを追加
func (s *Switch) AddPort(port *Port) {
	s.Ports[port.ID] = port
}

// ReceiveFrame ポートからフレームを受信したときの処理
func (s *Switch) ReceiveFrame(frame *EthernetFrame, fromPortID int) {
	inPort, exists := s.Ports[fromPortID]
	if !exists {
		fmt.Printf("\033[1;31m[%s] エラー: 未知のポート %d から受信しました\033[0m\n", s.Name, fromPortID)
		return
	}

	fmt.Printf("\n\033[1;36m================================================================================\033[0m\n")
	fmt.Printf("\033[1;36m▶ [%s] ポート %d (%s) でフレームを受信\033[0m\n", s.Name, fromPortID, inPort.Mode)
	fmt.Printf("\033[1;36m================================================================================\033[0m\n")

	// 1. VLAN IDの判定・タグの検証
	var vlanID uint16
	switch inPort.Mode {
	case PortModeAccess:
		// アクセスポートの場合: Untaggedフレームが来るはず
		if frame.VLANTag != nil {
			fmt.Printf("  \033[1;31m[DROP] アクセスポートにタグ付きフレームが到着したため破棄します\033[0m\n")
			return
		}
		// アクセスポートのPVIDがこのフレームのVLAN IDとなる
		vlanID = inPort.PVID
		fmt.Printf("  \033[1;32m[ポート判定]\033[0m アクセスポート (PVID: %d) のため、内部的に \033[1;33m[VLAN %d]\033[0m として扱います\n",
			vlanID, vlanID)
	case PortModeTrunk:
		// トランクポートの場合: 802.1Qタグが付いているはず
		if frame.VLANTag == nil {
			fmt.Printf("  \033[1;31m[DROP] トランクポートにタグなしフレームが到着したため破棄します\033[0m\n")
			return
		}
		vlanID = frame.VLANTag.VID
		// 許可されたVLANかチェック
		if !inPort.AllowedVLANs[vlanID] {
			fmt.Printf("  \033[1;31m[DROP] トランクポートで許可されていない VLAN %d のタグのため破棄します\033[0m\n", vlanID)
			return
		}
		fmt.Printf("  \033[1;32m[ポート判定]\033[0m トランクポートから 802.1Qタグ (\033[1;33mVID: %d\033[0m) を読み取りました（許可リストに含まれるため受信OK）\n",
			vlanID)
	}

	// 2. MACアドレス学習 (送信元MAC + VLAN ID -> 受信ポート)
	s.learnMAC(frame.SrcMAC, vlanID, fromPortID)

	// 3. 宛先ポートの決定と転送
	if frame.IsBroadcast() {
		fmt.Printf("  \033[1;35m[宛先判定]\033[0m 宛先がブロードキャスト (FF:FF:FF:FF:FF:FF) です。\n")
		fmt.Printf("             \033[1;33m同一 VLAN %d に所属するポートにのみフラッディング (Flooding)\033[0m します！\n", vlanID)
		s.flood(frame, vlanID, fromPortID)
	} else {
		// ユニキャスト
		macKey := s.macKey(frame.DstMAC, vlanID)
		entry, found := s.MACTable[macKey]
		if found {
			fmt.Printf("  \033[1;32m[宛先判定]\033[0m MACテーブルにヒット! 宛先: %s (VLAN %d) は \033[1;32mPort %d\033[0m に接続されています\n",
				frame.DstMAC, vlanID, entry.PortID)
			if entry.PortID == fromPortID {
				fmt.Printf("  \033[1;30m[FILTER] 受信元と同じポートのためフレームを破棄(フィルタリング)します\033[0m\n")
				return
			}
			s.forward(frame, vlanID, entry.PortID)
		} else {
			fmt.Printf("  \033[1;33m[宛先判定]\033[0m MACテーブルに宛先 %s (VLAN %d) が見つかりません (Unknown Unicast)。\n",
				frame.DstMAC, vlanID)
			fmt.Printf("             \033[1;33m同一 VLAN %d に所属する全ポートへフラッディング\033[0m して学習を促します\n", vlanID)
			s.flood(frame, vlanID, fromPortID)
		}
	}
}

// learnMAC MACアドレス学習
func (s *Switch) learnMAC(mac string, vlanID uint16, portID int) {
	key := s.macKey(mac, vlanID)
	existing, found := s.MACTable[key]
	if !found || existing.PortID != portID {
		s.MACTable[key] = &MACEntry{
			MAC:       mac,
			VLANID:    vlanID,
			PortID:    portID,
			LearnedAt: time.Now(),
		}
		fmt.Printf("  \033[1;34m[MAC学習]\033[0m MACテーブルに登録: [MAC: %s | VLAN: %2d] ==> \033[1mPort %d\033[0m\n",
			mac, vlanID, portID)
	}
}

// flood 同一VLANに所属するポート（受信ポート以外）へフレームを複製して転送
func (s *Switch) flood(frame *EthernetFrame, vlanID uint16, fromPortID int) {
	for portID, port := range s.Ports {
		if portID == fromPortID {
			continue // 受信ポートには送らない
		}

		// そのポートが該当VLANのトラフィックを通すかチェック
		switch port.Mode {
		case PortModeAccess:
			if port.PVID == vlanID {
				fmt.Printf("    ├── 送出先: \033[1;32mPort %d (Access, VLAN %d)\033[0m - VLAN一致のため転送\n", portID, port.PVID)
				s.sendFrameToPort(frame, vlanID, port)
			} else {
				fmt.Printf("    ├── \033[1;31mスキップ: Port %d (Access, VLAN %d)\033[0m - \033[1;31mVLAN不一致のため遮断 (隔離成功!)\033[0m\n",
					portID, port.PVID)
			}
		case PortModeTrunk:
			if port.AllowedVLANs[vlanID] {
				fmt.Printf("    ├── 送出先: \033[1;33mPort %d (Trunk)\033[0m - 許可リストに含まれるためタグ付きで転送\n", portID)
				s.sendFrameToPort(frame, vlanID, port)
			} else {
				fmt.Printf("    ├── \033[1;31mスキップ: Port %d (Trunk)\033[0m - VLAN %d はTrunkで不許可のため遮断\n", portID, vlanID)
			}
		}
	}
}

// forward 特定のポートへ転送
func (s *Switch) forward(frame *EthernetFrame, vlanID uint16, outPortID int) {
	outPort, exists := s.Ports[outPortID]
	if !exists {
		return
	}

	// ポートのVLAN整合性チェック
	if outPort.Mode == PortModeAccess {
		if outPort.PVID != vlanID {
			fmt.Printf("  \033[1;31m[DROP] 送出先 Port %d は VLAN %d のため、VLAN %d のパケットは送出できません (VLAN隔離)\033[0m\n",
				outPortID, outPort.PVID, vlanID)
			return
		}
		s.sendFrameToPort(frame, vlanID, outPort)
	} else if outPort.Mode == PortModeTrunk {
		if !outPort.AllowedVLANs[vlanID] {
			fmt.Printf("  \033[1;31m[DROP] 送出先 Trunk Port %d では VLAN %d が許可されていません\033[0m\n",
				outPortID, vlanID)
			return
		}
		s.sendFrameToPort(frame, vlanID, outPort)
	}
}

// sendFrameToPort ポート種別に応じたタグ加工（Tagging / Untagging）を行って送出
func (s *Switch) sendFrameToPort(frame *EthernetFrame, vlanID uint16, outPort *Port) {
	outFrame := frame.Clone()

	if outPort.Mode == PortModeAccess {
		// アクセスポートへ送出する場合: タグを除去 (Untagging)
		if outFrame.VLANTag != nil {
			fmt.Printf("    │   ↳ \033[1;35m[Untagging]\033[0m 802.1Qタグを外してタグなしフレームに変換\n")
		}
		outFrame.VLANTag = nil
		if outPort.ConnectedTo != nil {
			fmt.Printf("    │   ↳ \033[1;32m[送信]\033[0m %s へフレームを届けます\n", outPort.ConnectedTo.GetName())
			outPort.ConnectedTo.ReceiveFrame(outFrame, outPort.TargetPortID)
		}
	} else if outPort.Mode == PortModeTrunk {
		// トランクポートへ送出する場合: 802.1Qタグを付与 (Tagging)
		outFrame.VLANTag = &VLANHeader{
			TPID: EtherType8021Q,
			VID:  vlanID,
		}
		fmt.Printf("    │   ↳ \033[1;33m[Tagging]\033[0m 802.1Q VLANタグ (\033[1;33mVID=%d\033[0m) をフレームに付与して幹線へ送出\n", vlanID)
		if outPort.ConnectedTo != nil {
			fmt.Printf("    │   ↳ \033[1;33m[Trunk転送]\033[0m 対向の %s (Port %d) へタグ付きフレームを送信\n",
				outPort.ConnectedTo.GetName(), outPort.TargetPortID)
			outPort.ConnectedTo.ReceiveFrame(outFrame, outPort.TargetPortID)
		}
	}
}

func (s *Switch) macKey(mac string, vlanID uint16) string {
	return fmt.Sprintf("%s_VLAN%d", strings.ToUpper(mac), vlanID)
}

// PrintMACTable MACアドレステーブルを表示
func (s *Switch) PrintMACTable() {
	fmt.Printf("\n┌────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│  %-50s    │\n", fmt.Sprintf("MAC Address Table (%s)", s.Name))
	fmt.Printf("├───────────────────┬─────────────┬──────────────────────┤\n")
	fmt.Printf("│ MAC Address       │ VLAN ID     │ Port                 │\n")
	fmt.Printf("├───────────────────┼─────────────┼──────────────────────┤\n")
	if len(s.MACTable) == 0 {
		fmt.Printf("│ (空 - まだ学習していません)                            │\n")
	} else {
		for _, entry := range s.MACTable {
			fmt.Printf("│ %-17s │ VLAN %-6d │ Port %-15d │\n", entry.MAC, entry.VLANID, entry.PortID)
		}
	}
	fmt.Printf("└───────────────────┴─────────────┴──────────────────────┘\n")
}
