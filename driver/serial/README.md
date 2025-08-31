# GE Healthcare Vital Monitor Driver

GE HealthcareのS/5システム用のバイタルモニタードライバーです。DRI（Digital Radiography Interface）仕様に基づいて、バイナリデータをJSON形式に変換する機能を提供します。

## 概要

このドライバーは、GE HealthcareのS/5システムから送信されるバイナリデータを解析し、人間が読めるJSON形式に変換します。主に以下の機能を提供します：

- **波形データ解析**: ECG、血圧、SpO2などの波形データの解析
- **トレンドデータ解析**: 60秒トレンド値、10秒トレンド値、表示値の解析
- **生理学的データ解析**: 補助情報、NIBP、CO、PCWPなどの解析
- **バイナリからJSON変換**: すべてのデータをJSON形式で出力

## 実装内容

### 1. データ型定義 (`driver/serial/type.go`)

#### 基本定数
- **DRIメッセージタイプ**: `DRI_MT_PHDB`, `DRI_MT_WAVE`, `DRI_MT_ALARM`など
- **DRIインターフェースレベル**: `DRI_LEVEL_95`から`DRI_LEVEL_06`まで
- **サブレコードタイプ**: `DRI_PH_DISPL`, `DRI_PH_10S_TREND`, `DRI_PH_60S_TREND`など
- **波形サブレコードタイプ**: `DRI_WF_ECG12`, `DRI_WF_INVP5`など

#### 主要構造体
```go
// Datex-Ohmeda Record Header
type DatexHeader struct {
    RLen      int16      // レコードの総長
    RNbr      byte       // レコード番号
    DriLevel  byte       // DRIレベル
    PlugID    uint16     // プラグID
    RTime     uint32     // 送信時刻（Unix時間）
    RMainType int16      // メインタイプ
    SrDesc    [8]SrDesc  // サブレコード記述子
}

// 波形データ
type WaveformData struct {
    Header  WaveformHeader
    Samples []int16       // 16ビットサンプル
}

// 生理学的データベースレコード
type PhysiologicalDatabaseRecord struct {
    Time         uint32
    PhysData     PhysiologicalDataUnion
    Marker       byte
    Reserved     byte
    ClDriLvlSubt uint16
}
```

#### 生理学的データグループ
- **O2 Group**: 酸素濃度データ
- **N2O Group**: 亜酸化窒素データ
- **Anesthesia Agent Group**: 麻酔薬データ
- **Flow & Volume Group**: 流量・容量データ
- **Cardiac Output & Wedge Pressure Group**: 心拍出量・楔入圧データ
- **NMT Group**: 神経筋伝導データ
- **ECG Extra Group**: ECG追加データ
- **SvO2 Group**: 混合静脈血酸素飽和度データ

### 2. 波形データ解析 (`driver/serial/parse_wave.go`)

#### 主要機能
- **バイナリデータ解析**: `UnmarshalBinary()`メソッド
- **JSON変換**: `ToJSON()`メソッド
- **物理値変換**: `ConvertSampleToPhysicalValue()`関数
- **サンプリングレート取得**: `GetSamplingRate()`関数

#### 使用例
```go
// バイナリデータをJSONに変換
jsonString, err := ParseAndConvertToJSON(binaryData)
if err != nil {
    log.Fatal(err)
}

// 構造体として取得
waveform, err := ParseAndConvertToStruct(binaryData)
if err != nil {
    log.Fatal(err)
}
```

### 3. トレンドデータ解析 (`driver/serial/parse_trend.go`)

#### 主要機能
- **60秒トレンド値解析**: 要求された60秒トレンドデータの処理
- **複数レコード解析**: `ParseMultipleTrends()`関数
- **データ妥当性検証**: `ValidateTrendData()`関数
- **サマリー取得**: `GetTrendSummary()`関数

### 4. アラームデータ解析 (`driver/serial/parse_alarm.go`)

#### 主要機能
- **アラームステータス解析**: アラームの状態とメッセージの解析
- **アラーム優先度管理**: 赤、黄、白の優先度レベル対応
- **サイレンス情報解析**: アラームのサイレンス状態の解析
- **複数アラーム処理**: 最大5つのアラームメッセージ対応

#### JSON出力構造
```go
type TrendJSON struct {
    Timestamp     string                 `json:"timestamp"`
    UnixTimestamp uint32                 `json:"unix_timestamp"`
    RecordType    string                 `json:"record_type"`
    RecordNumber  int                    `json:"record_number"`
    DriLevel      int                    `json:"dri_level"`
    DriLevelDesc  string                 `json:"dri_level_description"`
    PlugID        int                    `json:"plug_id"`
    MainType      int                    `json:"main_type"`
    MainTypeName  string                 `json:"main_type_name"`
    Subrecords    []SubrecordJSON        `json:"subrecords"`
    Groups        map[string]interface{} `json:"groups"`
    IsValid       bool                   `json:"is_valid"`
    ParseErrors   []string               `json:"parse_errors,omitempty"`
}
```

## サポートするデータタイプ

### 1. 波形データ
- **ECG**: 心電図データ（300-500 Hz）
- **血圧**: 侵襲的血圧データ（100 Hz）
- **SpO2**: 酸素飽和度データ（100 Hz）
- **CO2**: 二酸化炭素濃度データ（25 Hz）
- **流量・容量**: 呼吸流量・容量データ

### 2. トレンドデータ
- **表示値**: 現在の表示値
- **10秒トレンド**: 10秒間隔のトレンド値
- **60秒トレンド**: 60秒間隔のトレンド値（重点対応）
- **補助情報**: NIBP、CO、PCWP測定時間、体表面積

### 3. アラームデータ
- **アラームステータス**: アラームの状態とメッセージ
- **アラーム優先度**: 赤（高）、黄（中）、白（低）の優先度
- **サイレンス情報**: アラームのサイレンス状態（無音、無呼吸、心停止など）
- **アラームメッセージ**: 最大5つのアラームテキスト（例: "HR LOW"）

### 4. 生理学的データグループ
- **Basic**: 基本生理学的データ（ECG、血圧、体温、SpO2など）
- **Extended 1**: 不整脈解析、ST解析、12誘導ECGなど
- **Extended 2**: NMT、EEG、エントロピーなど
- **Extended 3**: ガス測定、スパイロメトリー、トノメトリーなど

## 使用方法

### 1. 波形データの解析
```go
package main

import (
    "fmt"
    "log"
    "driver/serial"
)

func main() {
    // バイナリデータをJSONに変換
    jsonString, err := serial.ParseAndConvertToJSON(binaryData)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(jsonString)
    
    // 構造体として取得
    waveform, err := serial.ParseAndConvertToStruct(binaryData)
    if err != nil {
        log.Fatal(err)
    }
    
    // サマリー取得
    summary := serial.GetWaveformSummary(waveform)
    fmt.Printf("Summary: %+v\n", summary)
}
```

### 2. トレンドデータの解析
```go
package main

import (
    "fmt"
    "log"
    "driver/serial"
)

func main() {
    // 単一のトレンドデータを解析
    jsonString, err := serial.ParseAndConvertToJSON(trendData)
    if err != nil {
        log.Fatal(err)
    }
    
    // 複数のトレンドデータを解析
    jsonString, err = serial.ParseMultipleTrendsToJSON(trendData)
    if err != nil {
        log.Fatal(err)
    }
    
    // データの妥当性検証
    err = serial.ValidateTrendData(trendData)
    if err != nil {
        log.Printf("Validation error: %v", err)
    }
}
```

### 3. アラームデータの解析
```go
package main

import (
    "fmt"
    "log"
    "driver/serial"
)

func main() {
    // 単一のアラームデータを解析
    jsonString, err := serial.ParseAndConvertToJSON(alarmData)
    if err != nil {
        log.Fatal(err)
    }
    
    // 複数のアラームデータを解析
    jsonString, err = serial.ParseMultipleAlarmsToJSON(alarmData)
    if err != nil {
        log.Fatal(err)
    }
    
    // データの妥当性検証
    err = serial.ValidateAlarmData(alarmData)
    if err != nil {
        log.Printf("Validation error: %v", err)
    }
    
    // アラームサマリー取得
    alarm, err := serial.ParseAndConvertToStruct(alarmData)
    if err != nil {
        log.Fatal(err)
    }
    summary := serial.GetAlarmSummary(alarm)
    fmt.Printf("Summary: %+v\n", summary)
}
```

## JSON出力例

### 波形データ出力
```json
{
  "header": {
    "act_len": 1000,
    "status": 0,
    "label": 0,
    "has_gap": false,
    "has_pacer_detected": false,
    "has_lead_off": false
  },
  "samples": [
    {
      "index": 0,
      "raw_value": 1234,
      "physical_value": 12.34,
      "unit": "μV",
      "is_control_code": false
    }
  ]
}
```

### トレンドデータ出力
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "unix_timestamp": 1705312200,
  "record_type": "Trend Data",
  "dri_level": 6,
  "dri_level_description": "2019 '19",
  "main_type": 1,
  "main_type_name": "Physiological Data",
  "subrecords": [
    {
      "index": 0,
      "type": 3,
      "type_name": "60 Second Trended Values",
      "is_valid": true,
      "data": {
        "timestamp": "2024-01-15T10:30:00Z",
        "data_class": 0,
        "data_class_name": "Basic"
      }
    }
  ],
  "is_valid": true
}
```

### アラームデータ出力
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "unix_timestamp": 1705312200,
  "record_type": "Alarm Data",
  "dri_level": 6,
  "dri_level_description": "2019 '19",
  "main_type": 4,
  "main_type_name": "Alarm Data",
  "alarm_data": {
    "type": "alarm_status_message",
    "data": {
      "sound_on_off": {
        "value": true,
        "status": true
      },
      "silence_info": {
        "value": 0,
        "description": "Alarms are not silenced at bedside",
        "is_silenced": false
      },
      "alarms": [
        {
          "text": {
            "value": "HR LOW",
            "changed": true
          },
          "color": {
            "value": 2,
            "name": "Yellow",
            "changed": false
          },
          "priority": {
            "level": 2,
            "is_active": true
          }
        }
      ],
      "active_alarm_count": 1
    }
  },
  "is_valid": true
}
```

## エラーハンドリング

ドライバーは包括的なエラーハンドリングを提供します：

- **データ長検証**: バイナリデータの長さチェック
- **DRIレベル検証**: サポートされているDRIレベルの確認
- **パースエラー収集**: 解析エラーの詳細な記録
- **妥当性検証**: データの整合性チェック

## 技術仕様

### 対応DRIレベル
- DRI_LEVEL_95 (1995 '95)
- DRI_LEVEL_97 (1997 '97)
- DRI_LEVEL_98 (1998 '98)
- DRI_LEVEL_99 (1999 '99)
- DRI_LEVEL_00 (2001 '01)
- DRI_LEVEL_01 (2002 '02)
- DRI_LEVEL_02 (2003 '03)
- DRI_LEVEL_03 (2005 '05)
- DRI_LEVEL_04 (2009 '09)
- DRI_LEVEL_05 (2015 '15)
- DRI_LEVEL_06 (2019 '19)

### バイトオーダー
- **Little Endian**: すべての数値データはリトルエンディアン形式

### タイムスタンプ形式
- **Unix時間**: 1970年1月1日00:00:00からの秒数
- **RFC3339**: 人間が読める形式（例: "2024-01-15T10:30:00Z"）

## ファイル構成

```
driver/
├── serial/
│   ├── type.go           # データ型定義
│   ├── parse_wave.go     # 波形データ解析
│   ├── parse_trend.go    # トレンドデータ解析
│   ├── parse_alarm.go    # アラームデータ解析
│   └── sample/
│       ├── trend_sample.json  # トレンドデータJSON出力サンプル
│       └── alarm_sample.json  # アラームデータJSON出力サンプル
└── README.md
```

## ライセンス

このプロジェクトは医療機器データ処理用のドライバーです。医療用途での使用には適切な認証と検証が必要です。

## 注意事項

- このドライバーはGE HealthcareのS/5システム専用です
- 医療用途での使用前に十分なテストと検証を行ってください
- 実際の医療機器との接続には、適切な認証とライセンスが必要です
- 60秒トレンドデータの処理に重点を置いて実装されています
