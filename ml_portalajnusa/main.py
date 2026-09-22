from fastapi import FastAPI
from pydantic import BaseModel
import joblib
import json
import re
from Sastrawi.Stemmer.StemmerFactory import StemmerFactory
from Sastrawi.StopWordRemover.StopWordRemoverFactory import StopWordRemoverFactory

app = FastAPI(title="Chatbot ISP API")

stemmer = StemmerFactory().create_stemmer()
stopword = StopWordRemoverFactory().create_stop_word_remover()

def preprocess_text(text):
    text = text.lower()
    text = re.sub(r'[^a-z0-9\s]', '', text)
    text = stopword.remove(text)
    return stemmer.stem(text)

model = joblib.load('chatbot_model.pkl')
with open('dataset.json', 'r', encoding='utf-8') as f:
    dataset = json.load(f)

def get_bot_response(tag_name):
    for intent in dataset['intents']:
        if intent['tag'] == tag_name:
            return intent['responses'][0]
    return "Mohon maaf, sistem sedang mengalami kendala."

class ChatRequest(BaseModel):
    message: str

@app.post("/api/chat")
async def chat_endpoint(request: ChatRequest):
    clean_input = preprocess_text(request.message)
    
    probabilities = model.predict_proba([clean_input])[0]
    max_prob = max(probabilities)
    
    if max_prob < 0.50:
        predicted_tag = "fallback"
    else:
        predicted_tag = model.classes_[list(probabilities).index(max_prob)]
        
    bot_reply = get_bot_response(predicted_tag)
    
    return {
        "user_message": request.message,
        "predicted_intent": predicted_tag,
        "confidence_score": round(float(max_prob), 3),
        "bot_reply": bot_reply
    }