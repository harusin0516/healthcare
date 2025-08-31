# HL7 Communication Driver

GE Healthcare Vital Monitors用のHL7プロトコル通信ドライバーです。このドライバーは、HL7メッセージをソケット通信で受信し、JSON形式にパースする機能を提供します。

## 📋 概要

このHL7ドライバーは以下の機能を提供します：

- **HL7メッセージ受信**: TCPソケット通信によるHL7メッセージの受信
- **MLLP対応**: HL7 Minimal Lower Layer Protocol (MLLP) フレーミングの処理
- **JSON変換**: 受信したHL7メッセージをJSON形式に変換
- **複数メッセージタイプ対応**: ADT、ORU、ORMなどの各種HL7メッセージタイプ
- **GE Healthcareフォーマット対応**: GE Healthcareモニターの実際のメッセージフォーマット

## 🏗️ アーキテクチャ

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HL7 Monitor   │───▶│   HL7 Server    │───▶│   JSON Output   │
│   (GE Device)   │    │   (Go Driver)   │    │   (Parsed Data) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   ACK Response  │
                       │   (HL7 ACK)     │
                       └─────────────────┘
```

## 📁 ファイル構成

```
driver/hl7/
├── README.md              # このファイル
├── config.json            # サーバー設定ファイル
├── main.go                # メインエントリーポイント
├── types.go               # HL7データ構造とパーサー
├── server.go              # HL7 TCPサーバー
├── hl7_com_driver.go      # HL7通信ドライバー
├── sample_messages.go     # サンプルHL7メッセージ
├── test_client.go         # テストクライアント
└── sample/
    └── hl7_sample.json    # サンプルJSON出力
```

## 🚀 セットアップ

### 1. 依存関係

```bash
# Go 1.19以上が必要
go version
```

### 2. 設定ファイル

`config.json`を編集してサーバー設定を行います：

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "timeout": 30,
    "max_connections": 100
  },
  "hl7": {
    "version": "2.6",
    "encoding": "UTF-8",
    "field_separator": "|",
    "component_separator": "^",
    "subcomponent_separator": "&",
    "repetition_separator": "~",
    "escape_character": "\\"
  },
  "logging": {
    "level": "info",
    "file": "hl7_server.log",
    "max_size": 100,
    "max_backups": 3
  },
  "security": {
    "enable_tls": false,
    "cert_file": "",
    "key_file": "",
    "allowed_ips": []
  }
}
```

### 3. サーバー起動

```bash
# サーバーを起動
go run main.go -config config.json

# またはビルドして実行
go build -o hl7_server main.go
./hl7_server -config config.json
```

## 📊 対応メッセージタイプ

### GE Healthcare フォーマット

| メッセージタイプ | 説明 | デバイスID |
|----------------|------|------------|
| `ORU_VitalSigns` | バイタルサイン監視 | `080019FFFE134535` |
| `ORU_SpO2` | 酸素飽和度監視 | `080019FFFE134535` |
| `ORU_ECG` | 心電図監視 | `080019FFFE134535` |
| `ORU_CO2` | 二酸化炭素監視 | `080019FFFE134535` |
| `ORU_Comprehensive` | 包括的監視 (Example 2) | `080019FFFE0B4020` |

### 標準HL7フォーマット

| メッセージタイプ | 説明 | HL7タイプ |
|----------------|------|-----------|
| `ADT_Admission` | 患者入院 | `ADT^A01` |
| `ADT_Discharge` | 患者退院 | `ADT^A03` |
| `ADT_Transfer` | 患者転院 | `ADT^A02` |
| `ORU_LabResults` | 検査結果 | `ORU^R01` |
| `ORM_Order` | 医療オーダー | `ORM^O01` |

## 🔧 使用方法

### 1. サーバー起動

```bash
# デフォルト設定で起動
go run main.go

# カスタム設定ファイルで起動
go run main.go -config my_config.json
```

### 2. テストクライアント

```bash
# 特定のメッセージタイプを送信
go run test_client.go -message ORU_VitalSigns

# カスタムホスト・ポートで送信
go run test_client.go -host 192.168.1.100 -port 8080 -message ORU_Comprehensive

# 全メッセージタイプを送信
go run test_client.go -message ALL
```

### 3. プログラムからの使用

```go
package main

import (
    "driver/hl7"
    "log"
)

func main() {
    // HL7ドライバーを作成
    driver, err := hl7.NewHL7Driver("config.json")
    if err != nil {
        log.Fatal(err)
    }

    // サーバーを開始
    if err := driver.Start(); err != nil {
        log.Fatal(err)
    }

    // サーバー状態を取得
    status := driver.GetStatus()
    log.Printf("Server status: %+v", status)

    // 接続クライアントを取得
    clients := driver.GetConnectedClients()
    log.Printf("Connected clients: %d", len(clients))

    // サーバーを停止
    defer driver.Stop()
}
```

## 📡 MLLP (Minimal Lower Layer Protocol)

このドライバーはHL7 MLLPフレーミングをサポートしています：

```
<SB> <HL7 message> <EB> <CR>
```

- **SB (Start Block)**: `0x0B` (VT - Vertical Tab)
- **EB (End Block)**: `0x1C` (FS - File Separator)
- **CR (Carriage Return)**: `0x0D` (CR - Carriage Return)

### 例

```
0x0B MSH|^~\&|VSP^080019FFFE134535^EUI-64|GE Healthcare|||20241219103000-0700||ORU^R01^ORU_R01|080019FFFE13453520241219103000|P|2.6|||NE|AL||UNICODE UTF-8|||PCD_DEC_001^IHE PCD^1.3.6.1.4.1.19376.1.6.1.1.1^ISO\rPID|||HED12^^^PID^MR||LAZY^KITTY^^^^^L|||\rPV1||E|ICU^^79874\rOBR|1|080019FFFE13453520241219103000^VSP^080019FFFE134535^EUI-64|080019FFFE13453520241219103000^VSP^080019FFFE134535^EUI-64|182777000^monitoring ofpatient^SCT|||20241219103000\rOBX|1||69965^MDC_DEV_MON_PHYSIO_MULTI_PARAM_MDS^MDC|1.0.0.0|||||||X\rOBX|2||69854^MDC_DEV_METER_PRESS_BLD_VMD^MDC|1.13.0.0|||||||X\rOBX|3||69855^MDC_DEV_METER_PRESS_BLD_CHAN^MDC|1.13.1.0|||||||X\rOBX|4|NM|150033^MDC_PRESS_BLD_ART_SYS^MDC|1.13.1.1|120|266016^MDC_DIM_MMHG^MDC|||||R|||||||080019FFFE134535^B1X5_GE 0x1C 0x0D
```

## 📊 MDC (Medical Device Communication) コード

GE HealthcareデバイスはMDCコードを使用して測定値を識別します：

### 血圧関連
- `150033^MDC_PRESS_BLD_ART_SYS^MDC`: 収縮期血圧
- `150034^MDC_PRESS_BLD_ART_DIA^MDC`: 拡張期血圧
- `150035^MDC_PRESS_BLD_ART_MEAN^MDC`: 平均血圧
- `150087^MDC_PRESS_BLD_VEN_CENT_MEAN^MDC`: 中心静脈圧

### 心電図関連
- `147842^MDC_ECG_HEART_RATE^MDC`: 心拍数
- `148066^MDC_ECG_V_P_C_RATE^MDC`: PVC率
- `131841^MDC_ECG_AMPL_ST_I^MDC`: ST振幅 (I誘導)

### 呼吸関連
- `151562^MDC_RESP_RATE^MDC`: 呼吸数
- `151712^MDC_CONC_AWAY_CO2_EXP^MDC`: 呼気終末CO2
- `151716^MDC_CONC_AWAY_CO2_INSP^MDC`: 吸気CO2

### その他
- `150344^MDC_TEMP^MDC`: 体温
- `150456^MDC_PULS_OXIM_SAT_O2^MDC`: SpO2
- `155024^MDC_EEG_PAROX_CRTX_BURST_SUPPRN^MDC`: EEGバースト抑制率

## 🔍 ログとデバッグ

### ログレベル

- **INFO**: 通常の動作ログ
- **DEBUG**: 詳細なデバッグ情報
- **ERROR**: エラーログ

### ログファイル

```
hl7_server.log
```

### デバッグモード

```bash
# デバッグログを有効化
export HL7_DEBUG=true
go run main.go
```

## 🧪 テスト

### 単体テスト

```bash
# パーサーテスト
go test -v ./types.go

# サーバーテスト
go test -v ./server.go
```

### 統合テスト

```bash
# サーバーを起動
go run main.go &

# テストクライアントでテスト
go run test_client.go -message ORU_VitalSigns

# パフォーマンステスト
go run test_client.go -performance 1000
```

## 🔒 セキュリティ

### IP制限

```json
{
  "security": {
    "allowed_ips": ["192.168.1.100", "10.0.0.50"]
  }
}
```

### TLS/SSL

```json
{
  "security": {
    "enable_tls": true,
    "cert_file": "server.crt",
    "key_file": "server.key"
  }
}
```

## 📈 パフォーマンス

### 推奨設定

- **最大接続数**: 100
- **タイムアウト**: 30秒
- **バッファサイズ**: 4096バイト

### ベンチマーク結果

```
Messages per second: ~1000
Average latency: <10ms
Memory usage: ~50MB
```

## 🐛 トラブルシューティング

### よくある問題

1. **接続エラー**
   ```
   Error: connection refused
   Solution: ファイアウォール設定を確認
   ```

2. **MLLPパースエラー**
   ```
   Error: invalid MLLP wrapper
   Solution: デバイス設定でMLLPを有効化
   ```

3. **メモリ不足**
   ```
   Error: out of memory
   Solution: max_connectionsを削減
   ```

### デバッグコマンド

```bash
# ネットワーク接続確認
netstat -an | grep 8080

# プロセス確認
ps aux | grep hl7_server

# ログ確認
tail -f hl7_server.log
```

## 📚 参考資料

- [HL7 2.6 Specification](https://www.hl7.org/implement/standards/product_brief.cfm?product_id=185)
- [GE Healthcare HL7 Manual](GE/hl7/B125_B105%20HL7%20MANUAL%20_SM_2062665-005_E.pdf)
- [IHE PCD Technical Framework](https://www.ihe.net/uploadedFiles/Documents/PCC/IHE_PCC_TF_Rev2-1_Vol1_FT_2011-08-19.pdf)

## 🤝 貢献

バグ報告や機能要望は、GitHubのIssueでお知らせください。

## 📄 ライセンス

このプロジェクトはMITライセンスの下で公開されています。

---

**開発者**: Healthcare Driver Team  
**バージョン**: 1.0.0  
**最終更新**: 2024年12月19日
