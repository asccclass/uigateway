from fastapi import FastAPI, HTTPException, Body
from fastapi.responses import Response
from kokoro import KPipeline
import soundfile as sf
import io
import torch
import uvicorn
import os

app = FastAPI()

# Initialize pipeline
# 'a' for American English, but Kokoro supports others. 
# We might want to make this configurable or detect language.
# For now, defaulting to 'a' as per basic usage, but since the user is likely Chinese (zh-TW), 
# we should check if 'z' (Mandarin) is better or if 'a' covers it via phonemes.
# Kokoro's 'z' is for Mandarin. Let's try to support both or default to a mix?
# The user's content is mixed Chinese/English. 
# Let's initialize with 'a' for now as the base, but we might need 'z' for Chinese.
# Actually, looking at the docs, 'z' is for Mandarin. 
# Let's initialize 'z' for Chinese support since the user interface is in Traditional Chinese.
# However, the user might want English too. 
# Let's stick to 'a' (American English) as the default lang_code in KPipeline often handles others via fallback or we need to switch.
# Wait, the docs say: 🇨🇳 'z' => Mandarin Chinese.
# Let's use 'z' as the primary since the user is speaking Chinese.
try:
    device = 'cuda' if torch.cuda.is_available() else 'cpu'
    print(f"Loading Kokoro model on {device}...")
    # Using 'z' for Mandarin Chinese support
    pipeline = KPipeline(lang_code='z', device=device) 
    print("Kokoro model loaded.")
except Exception as e:
    print(f"Error loading model: {e}")
    pipeline = None

@app.post("/v1/audio/speech")
async def generate_speech(
    input: str = Body(..., embed=True),
    voice: str = Body("af_heart", embed=True), # Default voice
    speed: float = Body(1.0, embed=True)
):
    if pipeline is None:
        raise HTTPException(status_code=500, detail="TTS Model not initialized")

    try:
        # Generate audio
        # pipeline returns a generator
        generator = pipeline(input, voice=voice, speed=speed)
        
        # Collect all audio segments
        all_audio = []
        for _, _, audio in generator:
            all_audio.extend(audio)
            
        if not all_audio:
             raise HTTPException(status_code=400, detail="No audio generated")

        # Convert to wav in memory
        buffer = io.BytesIO()
        sf.write(buffer, all_audio, 24000, format='WAV')
        buffer.seek(0)
        
        return Response(content=buffer.read(), media_type="audio/wav")

    except Exception as e:
        print(f"Error generating speech: {e}")
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8880)
