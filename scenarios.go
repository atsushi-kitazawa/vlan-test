package main

import (
	"fmt"
	"time"
)

// Scenario1 異なるVLAN間の通信隔離デモ
func RunScenario1(topo *NetworkTopology) {
	topo.ResetAllReceived()
	printHeader("シナリオ 1: 異なるVLAN間の通信（VLAN 10 営業部 → VLAN 20 開発部）")
	fmt.Print(`
【このシナリオで学ぶこと】
物理的に同じスイッチ (Switch 1) に接続されていても、
VLANが異なるとパケットが完全に隔離され、直接通信できないことを確認します。

  送信元: PC-A (Port 1, Access VLAN 10)
  宛先  : PC-C (Port 2, Access VLAN 20)
`)
	pressEnterToContinue()

	topo.PCA.SendUnicast(topo.PCC.MAC, "こんにちは、PC-Cさん！(営業部から開発部へ)")

	fmt.Printf("\n--------------------------------------------------------------------------------\n")
	fmt.Printf("【受信結果の確認】:\n")
	printDeliveryStatus("PC-A (送信元)", topo.PCA.HasReceived())
	printDeliveryStatus("PC-B (VLAN 10)", topo.PCB.HasReceived())
	printDeliveryStatus("PC-C (VLAN 20 宛先端末)", topo.PCC.HasReceived())
	printDeliveryStatus("PC-D (VLAN 20)", topo.PCD.HasReceived())

	if !topo.PCC.HasReceived() {
		fmt.Printf("\n\033[1;32m✔ [成功] PC-Cにはパケットが届きませんでした（正しく隔離されています）！\033[0m\n")
		fmt.Printf("💡 【解説】\n")
		fmt.Printf("   Switch 1 の内部で、Port 1 から入ったパケットは「VLAN 10」として扱われます。\n")
		fmt.Printf("   Port 2 は「VLAN 20」に所属しているため、スイッチはフレームを破棄・遮断します。\n")
		fmt.Printf("   このように、1台の物理スイッチを論理的に複数のスイッチへ分割できるのがVLANの基本です。\n")
	}
}

// Scenario2 トランクポートを経由した同一VLAN間の通信
func RunScenario2(topo *NetworkTopology) {
	topo.ResetAllReceived()
	printHeader("シナリオ 2: スイッチ間トランクを跨いだ同一VLAN通信（Tagging / Untagging）")
	fmt.Print(`
【このシナリオで学ぶこと】
別々のスイッチに接続された同一VLAN同士 (PC-A と PC-B) が通信する際、
スイッチ間幹線（トランクポート）をパケットがどう通過するかを確認します。
  1. 送信元PCは「タグなし」で送信
  2. スイッチ1がトランクポートへ送出する際に「802.1Q VLANタグ (VID=10)」を付与 (Tagging)
  3. スイッチ2がトランクポートでタグを受信し、VID=10 を識別
  4. スイッチ2がPC-Bへ送出する際に「タグを除去」 (Untagging)
  5. 宛先PCは「通常のタグなしパケット」として受信

  送信元: PC-A (Switch 1, VLAN 10)
  宛先  : PC-B (Switch 2, VLAN 10)
`)
	pressEnterToContinue()

	topo.PCA.SendUnicast(topo.PCB.MAC, "社内見積書データ.pdf (営業部内でのファイル共有)")

	fmt.Printf("\n--------------------------------------------------------------------------------\n")
	fmt.Printf("【受信結果の確認】:\n")
	printDeliveryStatus("PC-A (送信元)", topo.PCA.HasReceived())
	printDeliveryStatus("PC-B (VLAN 10 宛先端末)", topo.PCB.HasReceived())
	printDeliveryStatus("PC-C (VLAN 20)", topo.PCC.HasReceived())
	printDeliveryStatus("PC-D (VLAN 20)", topo.PCD.HasReceived())

	if topo.PCB.HasReceived() && !topo.PCC.HasReceived() && !topo.PCD.HasReceived() {
		fmt.Printf("\n\033[1;32m✔ [成功] PC-Bにのみ正常にフレームが届きました！\033[0m\n")
		fmt.Printf("💡 【解説】\n")
		fmt.Printf("   スイッチ間を繋ぐケーブルは1本ですが、フレームに「VLANタグ (IEEE 802.1Q)」を\n")
		fmt.Printf("   付けることで、複数のVLANのトラフィックが混ざらずに1本の幹線を共有（多重化）できます。\n")
		fmt.Printf("   また、PC-AやPC-Bなどのエンド端末はタグの存在を知る必要がありません。\n")
		fmt.Printf("   （スイッチが自動的にタグの付け外しを行ってくれるため）\n")
	}
}

// Scenario3 ブロードキャスト（ARP）の伝搬とVLANによる抑制
func RunScenario3(topo *NetworkTopology) {
	topo.ResetAllReceived()
	printHeader("シナリオ 3: ブロードキャスト（ARPリクエスト）のVLAN内限定フラッディング")
	fmt.Print(`
【このシナリオで学ぶこと】
ブロードキャストパケット (宛先 FF:FF:FF:FF:FF:FF) を送信した際、
通常のスイッチなら全ポートに拡散してしまいますが、
VLANが設定されている場合、同一VLAN内のポートにしか届かないことを確認します。

  送信元: PC-A (VLAN 10 営業部)
  宛先  : FF:FF:FF:FF:FF:FF (ブロードキャスト)
  目的  : 「192.168.10.2 を持っている人、MACアドレスを教えて！」(ARPリクエスト)
`)
	pressEnterToContinue()

	topo.PCA.SendBroadcast("ARP: Who has 192.168.10.2? Tell 192.168.10.1")

	fmt.Printf("\n--------------------------------------------------------------------------------\n")
	fmt.Printf("【受信結果の確認】:\n")
	printDeliveryStatus("PC-A (送信元)", topo.PCA.HasReceived())
	printDeliveryStatus("PC-B (VLAN 10 営業部)", topo.PCB.HasReceived())
	printDeliveryStatus("PC-C (VLAN 20 開発部)", topo.PCC.HasReceived())
	printDeliveryStatus("PC-D (VLAN 20 開発部)", topo.PCD.HasReceived())

	if topo.PCB.HasReceived() && !topo.PCC.HasReceived() && !topo.PCD.HasReceived() {
		fmt.Printf("\n\033[1;32m✔ [成功] 同一VLAN (営業部) のPC-Bにのみ届き、開発部のPC-C/Dには届きませんでした！\033[0m\n")
		fmt.Printf("💡 【解説】\n")
		fmt.Printf("   LAN内の端末が増えると、ARPなどのブロードキャストがネットワーク全体を圧迫します\n")
		fmt.Printf("   （ブロードキャストストームの原因）。\n")
		fmt.Printf("   VLANを設定することで「ブロードキャストドメイン（届く範囲）」を分割し、\n")
		fmt.Printf("   無関係な端末のCPU負荷やネットワーク帯域の消費を防ぎます。\n")
	}
}

// Scenario4 双方向通信とMACアドレステーブル学習
func RunScenario4(topo *NetworkTopology) {
	topo.ResetAllReceived()
	printHeader("シナリオ 4: 双方向通信によるMACアドレステーブルの学習とピンポイント転送")
	fmt.Print(`
【このシナリオで学ぶこと】
PC-B から PC-A へ返信を行うことで、双方向のスイッチがMACアドレスを学習し、
次回以降のパケット転送が無駄なフラッディングなしに「ピンポイント（ユニキャスト）」で
転送される様子を確認します。
`)
	pressEnterToContinue()

	fmt.Printf("▶ ステップ 1: PC-B (Switch 2) から PC-A (Switch 1) へ返信パケットを送信\n")
	topo.PCB.SendUnicast(topo.PCA.MAC, "受信了解しました！(PC-Bからの返信)")

	fmt.Printf("\n--------------------------------------------------------------------------------\n")
	fmt.Printf("▶ ステップ 2: 両スイッチの MACアドレステーブル の状態を確認\n")
	topo.Switch1.PrintMACTable()
	topo.Switch2.PrintMACTable()

	fmt.Print(`
💡 【解説】
   MACアドレステーブルに [MACアドレス] と [VLAN ID] がペアで記録されていることに注目してください。
   同じMACアドレスであっても、VLANが異なれば別のエントリとして厳密に管理されます (IVL)。
   これ以降、PC-A と PC-B 間の通信はフラッディング不要となり、学習したポートへ直行します！
`)
}

// RunAllScenarios すべてのシナリオを順番に実行
func RunAllScenarios(topo *NetworkTopology) {
	fmt.Print("\n\033[1;36m================================================================================\033[0m\n")
	fmt.Print("          🚀 VLAN 理解のためのステップ・バイ・ステップ完全デモ開始\n")
	fmt.Print("\033[1;36m================================================================================\033[0m\n")

	topo.PrintTopologyDiagram()
	time.Sleep(1 * time.Second)

	RunScenario1(topo)
	time.Sleep(1 * time.Second)

	RunScenario2(topo)
	time.Sleep(1 * time.Second)

	RunScenario3(topo)
	time.Sleep(1 * time.Second)

	RunScenario4(topo)
}

func printHeader(title string) {
	fmt.Printf("\n\033[1;34m┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓\033[0m\n")
	fmt.Printf("\033[1;34m┃\033[0m \033[1m%-75s\033[0m \033[1;34m┃\033[0m\n", title)
	fmt.Printf("\033[1;34m┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛\033[0m\n")
}

func printDeliveryStatus(nodeDesc string, received bool) {
	if received {
		fmt.Printf("  • %-30s : \033[1;32m[ 着信あり ✔ ]\033[0m\n", nodeDesc)
	} else {
		fmt.Printf("  • %-30s : \033[1;90m[ 着信なし (遮断) ✘ ]\033[0m\n", nodeDesc)
	}
}

func pressEnterToContinue() {
	if autoMode {
		time.Sleep(600 * time.Millisecond)
		return
	}
	fmt.Printf("\n\033[2m(Enterキーを押すとシミュレーションを実行します...)\033[0m")
	var dummy string
	fmt.Scanln(&dummy)
}
