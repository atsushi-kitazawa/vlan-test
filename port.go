package main

import "fmt"

// PortMode ポートの動作モード
type PortMode string

const (
	PortModeAccess PortMode = "Access" // アクセスポート: 端末（PC等）が接続される。タグなし通信。
	PortModeTrunk  PortMode = "Trunk"  // トランクポート: スイッチ間等を接続。VLANタグ(802.1Q)を付けて複数VLANを多重化。
)

// NetworkInterface パケットを受信できる実体（PCやスイッチのポート）
type NetworkReceiver interface {
	ReceiveFrame(frame *EthernetFrame, fromPortID int)
	GetName() string
}

// Port スイッチの物理ポートを表現
type Port struct {
	ID           int             // ポート番号 (1, 2, 3...)
	SwitchName   string          // 所属するスイッチ名
	Mode         PortMode        // Access または Trunk
	PVID         uint16          // Accessポート時の所属VLAN ID (Port VLAN ID)
	AllowedVLANs map[uint16]bool // Trunkポート時に通過を許可するVLAN ID群
	ConnectedTo  NetworkReceiver // このポートに接続されている相手（PCまたは別スイッチ）
	TargetPortID int             // 相手がスイッチの場合の相手先ポートID
}

// NewAccessPort アクセスポートを作成
func NewAccessPort(switchName string, id int, vlanID uint16) *Port {
	return &Port{
		ID:           id,
		SwitchName:   switchName,
		Mode:         PortModeAccess,
		PVID:         vlanID,
		AllowedVLANs: make(map[uint16]bool),
	}
}

// NewTrunkPort トランクポートを作成
func NewTrunkPort(switchName string, id int, allowedVLANs []uint16) *Port {
	allowedMap := make(map[uint16]bool)
	for _, vid := range allowedVLANs {
		allowedMap[vid] = true
	}
	return &Port{
		ID:           id,
		SwitchName:   switchName,
		Mode:         PortModeTrunk,
		AllowedVLANs: allowedMap,
	}
}

// Connect 相手の機器とケーブルを接続
func (p *Port) Connect(target NetworkReceiver, targetPortID int) {
	p.ConnectedTo = target
	p.TargetPortID = targetPortID
}

// String 説明用文字列
func (p *Port) String() string {
	if p.Mode == PortModeAccess {
		return fmt.Sprintf("Port %d [Access (VLAN %d)]", p.ID, p.PVID)
	}
	var vlans []uint16
	for v := range p.AllowedVLANs {
		vlans = append(vlans, v)
	}
	return fmt.Sprintf("Port %d [Trunk (Allowed VLANs: %v)]", p.ID, vlans)
}
