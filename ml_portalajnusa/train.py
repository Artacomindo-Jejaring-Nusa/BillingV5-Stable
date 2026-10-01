import json
import re
import time
import sys
import joblib
from Sastrawi.Stemmer.StemmerFactory import StemmerFactory
from Sastrawi.StopWordRemover.StopWordRemoverFactory import StopWordRemoverFactory
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.pipeline import Pipeline

sys.stdout.reconfigure(encoding='utf-8')

print("=" * 60)
print("  TRAINING CHATBOT ISP FTTH (MACHINE LEARNING LOCAL)")
print("=" * 60)

start_time = time.time()

# 1. Load dataset
print("1. Membaca dataset.json...")
with open('dataset.json', 'r', encoding='utf-8') as f:
    data = json.load(f)

metadata = data.get('metadata', {})
print(f"   Nama Dataset: {metadata.get('nama_dataset', 'ISP Bot')}")
print(f"   Versi: {metadata.get('versi', '2.0')}")

X_raw = []
y_labels = []
for intent in data['intents']:
    for pattern in intent.get('patterns', []):
        X_raw.append(pattern)
        y_labels.append(intent['tag'])

total_samples = len(X_raw)
total_intents = len(set(y_labels))
print(f"   Total data latih: {total_samples} kalimat dari {total_intents} intent aktif.")

# 2. Inisialisasi Sastrawi & Cache Stemming
print("\n2. Memulai Text Preprocessing & Stemming (Sastrawi)...")
factory = StemmerFactory()
stemmer = factory.create_stemmer()
stopword = StopWordRemoverFactory().create_stop_word_remover()

t_prep = time.time()
cleaned_tokens = []
unique_words = set()

for p in X_raw:
    text = p.lower()
    text = re.sub(r'[^a-z0-9\s]', '', text)
    text = stopword.remove(text)
    words = text.split()
    for w in words:
        unique_words.add(w)
    cleaned_tokens.append(words)

print(f"   Ditemukan {len(unique_words)} kata unik. Melakukan morphological root stemming...")
word_stem_cache = {}
for w in unique_words:
    word_stem_cache[w] = stemmer.stem(w)

X_clean = [" ".join(word_stem_cache.get(w, w) for w in words) for words in cleaned_tokens]
prep_duration = time.time() - t_prep
print(f"   Text preprocessing selesai dalam {prep_duration:.2f} detik!")

# 3. Simpan cache kamus kata dasar agar inferensi runtime instan
with open('word_stem_cache.json', 'w', encoding='utf-8') as f:
    json.dump(word_stem_cache, f, ensure_ascii=False)
print("   Cache kata dasar tersimpan di 'word_stem_cache.json'.")

# 4. Melatih Pipeline Model (TF-IDF + Logistic Regression)
print("\n3. Melatih Model Machine Learning (TF-IDF N-Gram + Logistic Regression)...")
pipeline = Pipeline([
    ('tfidf', TfidfVectorizer(ngram_range=(1, 2), sublinear_tf=True)),
    ('clf', LogisticRegression(max_iter=1000, C=10.0, random_state=42))
])

t_train = time.time()
pipeline.fit(X_clean, y_labels)
train_duration = time.time() - t_train
train_accuracy = pipeline.score(X_clean, y_labels) * 100

print(f"   Pelatihan model selesai dalam {train_duration:.2f} detik!")
print(f"   Akurasi Training: {train_accuracy:.2f}%")

# 5. Export Model Artifact
joblib.dump(pipeline, 'chatbot_model.pkl')
print(f"\n4. Model berhasil diekspor ke 'chatbot_model.pkl'")

total_elapsed = time.time() - start_time
print("=" * 60)
print(f"  SUKSES! Total waktu pelatihan: {total_elapsed:.2f} detik.")
print("=" * 60)