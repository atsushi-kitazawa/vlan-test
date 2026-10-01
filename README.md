# VLAN (仮想LAN) 学習シミュレータ (Go言語)

VLAN（Virtual Local Area Network）の仕組みやパケットの動きを、Go言語で実装された仮想L2スイッチとイーサネットフレームを使ってステップ・バイ・ステップで可視化・体験できる学習用プロジェクトです。

---

## 💡 VLANとは？ なぜ理解が難しいのか？

### 1. VLANが必要な理由
- **物理スイッチは通常1つの「ブロードキャストドメイン（ブロードキャストが届く範囲）」** です。
- 接続端末が増えると、ARPなどのブロードキャストパケットが全員に届いて帯域や端末リソースを圧迫します（ブロードキャストストーム）。
- また、営業部と開発部など、別々の部署のPCが同じスイッチに刺さっているとセキュリティ上の境界が作れません。
- **VLANを使うことで、1台の物理スイッチの中に「仮想的なスイッチ」を複数作り、ネットワークを論理的に分割** できます。

### 2. つまずきやすいポイント
| 疑問 | 答え |
| :--- | :--- |
| **PCはVLANを知っている？** | **基本的に知りません。** PCは通常のイーサネットフレーム（タグなし）を送受信します。 |
| **誰がタグを付けるの？** | **スイッチ**です。スイッチのポート設定に応じて、スイッチが出入り口でタグを付けたり外したりします。 |
| **アクセスポートとは？** | PCやサーバが繋がるポート。**タグなし（Untagged）**で送受信されます。ポートに設定されたVLAN ID（PVID）が内部で適用されます。 |
| **トランクポートとは？** | スイッチ同士を繋ぐポート。複数のVLANの通信を通すため、フレームに**IEEE 802.1Qタグ（VLAN ID）を付与（Tagging）**して1本のケーブルを共有します。 |

---

## 🏗️ ネットワークトポロジー

本サンプルでは以下の構成をシミュレートしています。

```text
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
```

---

## 🚀 実行方法 (Nix環境)

### 1. Nix環境の起動

#### `nix develop` (Flakesをお使いの場合)
```bash
nix develop
```

#### `nix-shell` (標準のNixをお使いの場合)
```bash
nix-shell
```

どちらのコマンドでも、Go言語（`go`）と開発ツールが即座にセットアップされます。

### 2. シミュレータの起動

```bash
# 対話型メニュー起動
go run .

# または全自動デモ（Enterキーなしで通し実行）
go run . -auto

# 特定のシナリオのみ実行 (1〜4)
go run . -scenario=1
go run . -scenario=2
go run . -scenario=3
go run . -scenario=4
```

---

## 🔍 体験できる4つのシナリオ

### シナリオ 1: 異なるVLAN間の通信隔離
- **送信**: PC-A (Switch 1, VLAN 10) → PC-C (Switch 1, VLAN 20)
- **結果**: 物理的に同じスイッチに直結されているのに **パケットは届きません**。
- **ポイント**: スイッチはポート1から入ったフレームを「VLAN 10」として扱います。ポート2は「VLAN 20」のため、宛先であってもパケットを流さず破棄します。

### シナリオ 2: トランクポートによる同一VLAN通信 (Tagging / Untagging)
- **送信**: PC-A (Switch 1, VLAN 10) → PC-B (Switch 2, VLAN 10)
- **結果**: スイッチを跨いで **PC-Bにのみ正常にパケットが届きます**。
- **ポイント**:
  1. PC-A は通常のフレーム（Untagged）を送信。
  2. Switch 1 はトランクポート（Port 3）へ送出する際、**IEEE 802.1Qタグ (VID: 10)** を付与 (`Tagging`)。
  3. Trunk回線にはタグ付きフレームが流れる。
  4. Switch 2 はタグを読み取り、VLAN 10宛てであることを確認。
  5. Switch 2 はアクセスポート（Port 1）へ送出する際、**タグを除去** (`Untagging`)。
  6. PC-B はタグを一切意識せず通常のフレームを受信。

### シナリオ 3: ブロードキャスト (ARP) の抑制
- **送信**: PC-A が宛先 `FF:FF:FF:FF:FF:FF` のARPリクエストを送信
- **結果**: **同一VLANのPC-Bにだけ届き、VLAN 20のPC-CやPC-Dには一切届きません**。
- **ポイント**: これが「ブロードキャストドメインの分割」です。不要なブロードキャストトラフィックが他部署に漏れないため、ネットワーク負荷軽減とセキュリティが両立されます。

### シナリオ 4: 双方向通信とMACアドレステーブルの学習
- **送信**: PC-B から PC-A へ返信
- **結果**: Switch 1 と Switch 2 の双方が `[MACアドレス + VLAN ID] ==> Port番号` を学習。
- **ポイント**: スイッチはMACアドレスだけでなく「どのVLANのMACアドレスか」をセットで管理（IVL: Independent VLAN Learning）します。

---

## 📁 ディレクトリ構成

- [`flake.nix`](file:///Users/atsushi-kitazawa/workspace/golang/vlan-sample/flake.nix): Nix Flakes 環境定義
- [`shell.nix`](file:///Users/atsushi-kitazawa/workspace/golang/vlan-sample/shell.nix): Nix Shell 環境定義
- [`go.mod`](file:///Users/atsushi-kitazawa/workspace/golang/vlan-sample/go.mod): Go モジュール定義
- [`packet.go`](file:///Users/atsushi-kitazawa/workspace/golang/vlan-sample/packet.go): イーサネットフレームおよび IEEE 802.1Q VLAN タグ構造の定義
- [`port.go`](file:///Users/atsushi-kitazawa/workspace/golang/vlan-sample/port.go): アクセスポート / トランクポートの動作モデル
- [`switch.go`](file:///Users/atsushi-kitazawa/workspace/golang/vlan-sample/switch.go): L2スイッチ（MAC学習、タギング、アンタギング、フラッディング、フィルタリング）
- [`node.go`](file:///Users/atsushi-kitazawa/workspace/golang/vlan-sample/node.go): 仮想PC端末（ユニキャスト・ARPブロードキャスト送信）
- [`topology.go`](file:///Users/atsushi-kitazawa/workspace/golang/vlan-sample/topology.go): スイッチ2台・PC4台のトポロジー構築と図示
- [`scenarios.go`](file:///Users/atsushi-kitazawa/workspace/golang/vlan-sample/scenarios.go): 4つの学習シナリオの実行と詳細解説
- [`main.go`](file:///Users/atsushi-kitazawa/workspace/golang/vlan-sample/main.go): メニューUIおよびCLIエントリーポイント
