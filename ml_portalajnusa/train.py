import json
import re
import joblib
from Sastrawi.Stemmer.StemmerFactory import StemmerFactory
from Sastrawi.StopWordRemover.StopWordRemoverFactory import StopWordRemoverFactory
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.svm import SVC
from sklearn.pipeline import Pipeline

factory = StemmerFactory()
stemmer = factory.create_stemmer()
stopword = StopWordRemoverFactory().create_stop_word_remover()

def preprocess_text(text):
    text = text.lower()
    text = re.sub(r'[^a-z0-9\s]', '', text) 
    text = stopword.remove(text)
    return stemmer.stem(text) 

with open('dataset.json', 'r', encoding='utf-8') as f:
    data = json.load(f)

X_raw = []
y_labels = []

for intent in data['intents']:
    for pattern in intent['patterns']:
        X_raw.append(pattern)
        y_labels.append(intent['tag'])

print("Memulai text preprocessing (proses ini memakan waktu beberapa detik)...")
X_clean = [preprocess_text(text) for text in X_raw]

pipeline = Pipeline([
    ('tfidf', TfidfVectorizer()),
    ('clf', SVC(kernel='linear', probability=True))
])

print("Melatih model SVM...")
pipeline.fit(X_clean, y_labels)
print(f"Pelatihan selesai! Akurasi training: {pipeline.score(X_clean, y_labels):.2f}")

joblib.dump(pipeline, 'chatbot_model.pkl')
print("Model berhasil diekspor ke 'chatbot_model.pkl'")