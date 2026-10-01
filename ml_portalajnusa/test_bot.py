import os
import json
import re
import sys
import joblib
from Sastrawi.Stemmer.StemmerFactory import StemmerFactory
from Sastrawi.StopWordRemover.StopWordRemoverFactory import StopWordRemoverFactory

sys.stdout.reconfigure(encoding='utf-8')

BASE_DIR = os.path.dirname(os.path.abspath(__file__))

factory = StemmerFactory()
stemmer = factory.create_stemmer()
stopword = StopWordRemoverFactory().create_stop_word_remover()

word_stem_cache = {}
cache_file = os.path.join(BASE_DIR, 'word_stem_cache.json')
if os.path.exists(cache_file):
    with open(cache_file, 'r', encoding='utf-8') as f:
        word_stem_cache = json.load(f)

def preprocess_text(text: str) -> str:
    text = text.lower()
    text = re.sub(r'[^a-z0-9\s]', '', text)
    text = stopword.remove(text)
    words = [word_stem_cache.get(w, stemmer.stem(w)) for w in text.split()]
    return " ".join(words)

model = joblib.load(os.path.join(BASE_DIR, 'chatbot_model.pkl'))
with open(os.path.join(BASE_DIR, 'dataset.json'), 'r', encoding='utf-8') as f:
    dataset = json.load(f)

def get_bot_response(tag_name):
    for intent in dataset['intents']:
        if intent['tag'] == tag_name:
            responses = intent.get('responses', [])
            if responses:
                return responses[0]
    return "Mohon maaf, sistem sedang mengalami kendala."

print("=" * 60)
print("  Uji Coba Chatbot ISP Local ML (Ketik 'exit' untuk keluar)")
print("=" * 60)

while True:
    try:
        user_input = input("\nPelanggan: ")
    except (EOFError, KeyboardInterrupt):
        break
    if user_input.lower().strip() in ['exit', 'quit', 'keluar']:
        break
    if not user_input.strip():
        continue

    clean_input = preprocess_text(user_input)
    probabilities = model.predict_proba([clean_input])[0]
    max_prob = max(probabilities)

    CONFIDENCE_THRESHOLD = 0.50
    if max_prob < CONFIDENCE_THRESHOLD:
        predicted_tag = "fallback"
    else:
        predicted_tag = model.classes_[list(probabilities).index(max_prob)]

    handover_tags = {"fallback", "0.5_minta_dihubungkan_ke_cs", "keluhan_eskalasi", "0.6_kata_kasar_emosi"}
    needs_handover = (predicted_tag in handover_tags) or (max_prob < CONFIDENCE_THRESHOLD)

    bot_reply = get_bot_response(predicted_tag)

    status_tag = "🤖 [SOLVED BY ML]" if not needs_handover else "👤 [HANDOVER TO HUMAN CS]"
    print(f"\nStatus: {status_tag}")
    print(f"Intent: {predicted_tag} | Confidence: {max_prob*100:.1f}%")
    print(f"Bot:\n{bot_reply}")