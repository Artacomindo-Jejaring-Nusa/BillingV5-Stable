import joblib
import re
from Sastrawi.Stemmer.StemmerFactory import StemmerFactory

factory = StemmerFactory()
stemmer = factory.create_stemmer()

def preprocess_text(text):
    text = text.lower()
    text = re.sub(r'[^a-z0-9\s]', '', text)
    return stemmer.stem(text)

model = joblib.load('chatbot_model.pkl')

print("--- Uji Coba Chatbot ISP (Ketik 'exit' untuk keluar) ---")
while True:
    user_input = input("\nPelanggan: ")
    if user_input.lower() == 'exit':
        break
    
    clean_input = preprocess_text(user_input)
    
    probabilities = model.predict_proba([clean_input])[0]
    
    max_prob = max(probabilities)
    predicted_intent = model.classes_[list(probabilities).index(max_prob)]

    if max_prob < 0.45:
        print("Bot: [fallback] Mohon maaf, saya belum memahami maksud pesan Anda. Bisa dijelaskan lebih detail?")
    else:
        print(f"Bot: [Memicu intent '{predicted_intent}' dengan confidence {max_prob*100:.1f}%]")