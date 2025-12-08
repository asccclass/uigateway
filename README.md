# LLM client


# SPEAK
### 前置準備 (Prerequisites)
Kokoro 模型依賴 espeak-ng 來處理音素 (Phonemes)，在安裝前請確保您的系統已安裝此套件。
* Windows: 下載並安裝 [eSpeak NG](https://github.com/espeak-ng/espeak-ng) (安裝時請勾選 "English" 相關選項)。
* macOS: brew install espeak-ng
* Linux (Ubuntu/Debian): sudo apt-get install espeak-ng

Python: 建議使用 Python 3.10 或以上版本。

* 安裝

```
# 使用 pip
pip install kokoro-tts

# 或者從 GitHub 原始碼安裝 (如果您需要最新功能)
pip install git+https://github.com/nazdridoy/kokoro-tts.git
```

* Python

```
from kokoro_tts import KokoroTTS

# 初始化 TTS，第一次執行會自動下載模型權重 (~300MB)
tts = KokoroTTS(lang="en-us") # 支援 en-us, en-gb, ja, zh (中文支援視模型版本而定，主要為英文)

text = "Hello, this is a test of the Kokoro text to speech system."

# 生成語音並儲存
# voice 參數可選: af_bella, af_sarah, am_adam, am_michael 等
output_path = "output.wav"
tts.create_audio(
    text=text,
    output_file=output_path,
    voice="af_bella", 
    speed=1.0
)

print(f"語音已生成：{output_path}")

* CLI

```
# 基本用法
kokoro-tts "Hello world, this is an AI voice." output.wav

# 指定語速與聲音
kokoro-tts "Hello world" output.wav --voice af_sarah --speed 1.2

# 讀取電子書並轉檔
kokoro-tts technical_manual.pdf output_audiobook --lang en-us
```


```
docker run -d -p 8880:8880 ghcr.io/remsky/kokoro-fastapi-gpu:latest
# 若無 NVIDIA GPU，請改用: ghcr.io/remsky/kokoro-fastapi-cpu:latest
```

* 測試

```
curl http://172.18.124.210:8080/v1/audio/speech -H "Content-Type: application/json" -d '{
    "model": "kokoro",
    "input": "System analysis complete. Deploying updates now.",
    "voice": "af_bella",
    "response_format": "mp3",
    "speed": 1.0
  }' --output output.mp3
  ```

* 語言包

```
Kokoro TTS 模型目前的中文語音（Mandarin）主要分為 女性 (zf) 和 男性 (zm) 兩大類。這些語音 ID 的命名規則通常是 z (中文) + f/m (性別) + 拼音名稱。
以下是目前 Kokoro 官方及主流版本中可用的中文語音清單：

1. 女性語音 (Prefix: zf_)：這些聲音通常音質較為清亮或溫柔，適用於助手、朗讀或對話場景。

zf_xiaobei (小貝) - 最推薦/最常用，聲音甜美自然，適合大多數場景。
zf_xiaoni (小妮)
zf_xiaoxiao (小小)
zf_xiaoyi (小一)

2. 男性語音 (Prefix: zm_)：這些聲音較為沉穩，適合新聞播報或敘事。

zm_yunjian (雲健)
zm_yunxi (雲希)
zm_yunxia (雲夏)
zm_yunyang (雲揚)
```

### message hub test

```
curl "https://ws.justdrink.com.tw/api/send?msg=HelloAll"
```