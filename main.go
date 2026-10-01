package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

var autoMode = false

func main() {
	autoFlag := flag.Bool("auto", false, "Enterキーを押さずに全シナリオを自動実行します")
	scenarioFlag := flag.Int("scenario", 0, "特定のシナリオ番号 (1-4) を直接実行します")
	flag.Parse()

	if *autoFlag {
		autoMode = true
	}

	topo := BuildTopology()

	if *scenarioFlag > 0 {
		switch *scenarioFlag {
		case 1:
			RunScenario1(topo)
		case 2:
			RunScenario2(topo)
		case 3:
			RunScenario3(topo)
		case 4:
			RunScenario4(topo)
		default:
			fmt.Printf("無効なシナリオ番号: %d (1-4 を指定してください)\n", *scenarioFlag)
		}
		return
	}

	if *autoFlag {
		RunAllScenarios(topo)
		return
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		printMainMenu()
		fmt.Print("番号を入力してください (1-6): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			topo = BuildTopology()
			RunScenario1(topo)
		case "2":
			topo = BuildTopology()
			RunScenario2(topo)
		case "3":
			topo = BuildTopology()
			RunScenario3(topo)
		case "4":
			topo = BuildTopology()
			// シナリオ4は学習前の通信も必要なので、まずシナリオ2を実行してから
			RunScenario4(topo)
		case "5":
			topo = BuildTopology()
			RunAllScenarios(topo)
		case "6":
			topo.PrintTopologyDiagram()
			printPortDetails(topo)
		case "q", "exit", "7":
			fmt.Println("終了します。お疲れ様でした！")
			return
		default:
			fmt.Println("無効な選択です。1〜6の番号を入力してください。")
		}

		fmt.Println("\n--------------------------------------------------------------------------------")
		fmt.Print("(Enterキーを押してメニューに戻る)")
		reader.ReadString('\n')
	}
}

func printMainMenu() {
	fmt.Print(`
================================================================================
                    📡 VLAN (仮想LAN) 実装・学習シミュレータ
================================================================================
  [1] シナリオ 1: 異なるVLAN間の通信隔離 (なぜ同じスイッチなのに届かない？)
  [2] シナリオ 2: トランクポートによる同一VLAN通信 (Tagging / Untagging の仕組み)
  [3] シナリオ 3: ブロードキャスト (ARP) の抑制 (ブロードキャストドメインの分割)
  [4] シナリオ 4: 双方向通信と MACアドレステーブル の学習
  [5] 完全デモ   : シナリオ 1 〜 4 を順番にすべて実行
  [6] 構成確認   : ネットワーク構成図とスイッチポート詳細を表示
  [q] 終了
================================================================================
`)
}

func printPortDetails(topo *NetworkTopology) {
	fmt.Println("\n【各スイッチのポート設定一覧】")
	fmt.Printf("▶ %s:\n", topo.Switch1.Name)
	for i := 1; i <= len(topo.Switch1.Ports); i++ {
		fmt.Printf("   - %s\n", topo.Switch1.Ports[i])
	}
	fmt.Printf("▶ %s:\n", topo.Switch2.Name)
	for i := 1; i <= len(topo.Switch2.Ports); i++ {
		fmt.Printf("   - %s\n", topo.Switch2.Ports[i])
	}
}
