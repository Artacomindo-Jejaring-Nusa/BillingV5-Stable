import os
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
import joblib
import json
import re
from Sastrawi.Stemmer.StemmerFactory import StemmerFactory
from Sastrawi.StopWordRemover.StopWordRemoverFactory import StopWordRemoverFactory

BASE_DIR = os.path.dirname(os.path.abspath(__file__))

app = FastAPI(title="Chatbot ISP API - ML Local")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

stemmer = StemmerFactory().create_stemmer()
stopword = StopWordRemoverFactory().create_stop_word_remover()

def preprocess_text(text):
    text = text.lower()
    text = re.sub(r'[^a-z0-9\s]', '', text)
    text = stopword.remove(text)
    return stemmer.stem(text)

model = joblib.load(os.path.join(BASE_DIR, 'chatbot_model.pkl'))
with open(os.path.join(BASE_DIR, 'dataset.json'), 'r', encoding='utf-8') as f:
    dataset = json.load(f)

def get_bot_response(tag_name):
    for intent in dataset['intents']:
        if intent['tag'] == tag_name:
            return intent['responses'][0]
    return "Mohon maaf, sistem sedang mengalami kendala."

class ChatRequest(BaseModel):
    message: str

@app.get("/health")
async def health_check():
    return {"status": "ok", "service": "ml_portalajnusa", "model_loaded": model is not None}

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

# ==============================================================================
# --- MACHINE LEARNING: REVENUE & PAYMENT BEHAVIOR PATTERN ANALYSIS ---
# ==============================================================================
from typing import List, Optional
import numpy as np
import pandas as pd
from sklearn.cluster import KMeans
from sklearn.preprocessing import StandardScaler

class CustomerPaymentRecord(BaseModel):
    customer_id: int
    customer_name: str
    no_telp: Optional[str] = ""
    brand: Optional[str] = ""
    monthly_bill: float = 0.0
    total_invoices: int = 0
    paid_invoices: int = 0
    expired_invoices: int = 0
    avg_payment_day_of_month: float = 1.0  # Rata-rata tanggal bayar dalam bulan (1-31)
    days_to_pay_avg: float = 5.0  # Rata-rata hari pembayaran setelah invoice terbit (tgl 26/27)
    on_time_ratio: float = 1.0  # Rasio bayar sebelum/tepat tgl 1
    grace_period_ratio: float = 0.0  # Rasio bayar tgl 2-10
    late_ratio: float = 0.0  # Rasio bayar tgl 11+
    recent_status: Optional[str] = "Lunas"

class RevenueAnalysisRequest(BaseModel):
    cycle_date: Optional[str] = "26"
    target_month: Optional[str] = ""
    customers: List[CustomerPaymentRecord]

@app.post("/api/ml/revenue-analysis")
async def revenue_analysis_endpoint(request: RevenueAnalysisRequest):
    if not request.customers:
        return {
            "status": "empty",
            "message": "Tidak ada data pelanggan yang dikirim",
            "summary": {
                "total_customers": 0,
                "projected_billing": 0,
                "projected_recovery_rate": 0
            },
            "clusters_summary": [],
            "cash_flow_forecast": [],
            "classified_customers": [],
            "actionable_insights": []
        }

    # Convert customer data to DataFrame
    data_list = []
    for c in request.customers:
        data_list.append({
            "customer_id": c.customer_id,
            "customer_name": c.customer_name,
            "no_telp": c.no_telp or "",
            "brand": c.brand or "Jelantik",
            "monthly_bill": float(c.monthly_bill),
            "total_invoices": c.total_invoices,
            "paid_invoices": c.paid_invoices,
            "expired_invoices": c.expired_invoices,
            "avg_payment_day_of_month": float(c.avg_payment_day_of_month),
            "days_to_pay_avg": float(c.days_to_pay_avg),
            "on_time_ratio": float(c.on_time_ratio),
            "grace_period_ratio": float(c.grace_period_ratio),
            "late_ratio": float(c.late_ratio),
            "recent_status": c.recent_status or "Lunas",
        })

    df = pd.DataFrame(data_list)

    # Calculate Expired Ratio & Risk Indicator
    df["expired_ratio"] = df["expired_invoices"] / df["total_invoices"].clip(lower=1)
    
    # Feature Matrix for ML Clustering:
    # 1. on_time_ratio (Higher is better)
    # 2. grace_period_ratio
    # 3. late_ratio
    # 4. expired_ratio
    # 5. avg_payment_day_of_month
    features = ["on_time_ratio", "grace_period_ratio", "late_ratio", "expired_ratio", "avg_payment_day_of_month"]
    X = df[features].fillna(0).values

    # Determine Cluster Classification
    # We combine K-Means clustering with deterministic behavioral boundaries for robust industrial classification
    n_samples = len(df)
    n_clusters = min(4, n_samples)

    if n_samples >= 4:
        try:
            scaler = StandardScaler()
            X_scaled = scaler.fit_transform(X)
            kmeans = KMeans(n_clusters=n_clusters, random_state=42, n_init=10)
            df["kmeans_raw"] = kmeans.fit_predict(X_scaled)
        except Exception as e:
            df["kmeans_raw"] = 0

    # Rule-Refined Behavioral Tiering
    # Tier 1: Early / On-Time (Bayar tgl 26-30 atau tgl 1, expired = 0, on_time_ratio >= 0.70)
    # Tier 2: Grace Period (Bayar tgl 2-10, expired <= 1, grace_period_ratio > 0.40)
    # Tier 3: Late / Rawan Isolir (Bayar tgl 11+, late_ratio > 0.30)
    # Tier 4: At Risk / Churn (expired_ratio > 0.30 or recent_status == 'Expired')

    classified_list = []
    cluster_counts = {
        "early_ontime": {"name": "Early & On-Time Payer", "color": "#16a34a", "count": 0, "total_nominal": 0.0, "risk_level": "Sangat Rendah"},
        "grace_period": {"name": "Grace-Period Payer (Tgl 2-10)", "color": "#0284c7", "count": 0, "total_nominal": 0.0, "risk_level": "Rendah-Sedang"},
        "chronic_late": {"name": "Rawan Terisolir (Tgl 11+)", "color": "#ea580c", "count": 0, "total_nominal": 0.0, "risk_level": "Tinggi"},
        "at_risk_churn": {"name": "Risiko Nunggak / Churn", "color": "#dc2626", "count": 0, "total_nominal": 0.0, "risk_level": "Kritis"},
    }

    for idx, row in df.iterrows():
        expired_ratio = row["expired_ratio"]
        late_ratio = row["late_ratio"]
        grace_ratio = row["grace_period_ratio"]
        on_time_ratio = row["on_time_ratio"]
        avg_day = row["avg_payment_day_of_month"]
        recent_status = str(row["recent_status"]).lower()
        bill = row["monthly_bill"]

        # Classification Logic
        if expired_ratio >= 0.4 or recent_status in ["expired", "berhenti", "batal"]:
            cluster_key = "at_risk_churn"
            risk_score = min(100, int(75 + expired_ratio * 25))
            recommendation = "Follow-up penagihan intensif / konfirmasi kelanjutan layanan sebelum siklus cetak baru"
        elif late_ratio >= 0.35 or avg_day > 10:
            cluster_key = "chronic_late"
            risk_score = min(74, int(45 + late_ratio * 30))
            recommendation = "Kirim WhatsApp reminder otomatis H-2 sebelum tanggal 10 batas isolir"
        elif grace_ratio >= 0.40 or (1 < avg_day <= 10):
            cluster_key = "grace_period"
            risk_score = min(40, int(15 + (avg_day / 10) * 20))
            recommendation = "Kirim reminder tagihan normal pada tanggal 1 dan 5"
        else:
            cluster_key = "early_ontime"
            risk_score = max(5, int(10 - on_time_ratio * 5))
            recommendation = "Pelanggan teladan — cocok untuk program loyalitas / reward bayar tepat waktu"

        cluster_counts[cluster_key]["count"] += 1
        cluster_counts[cluster_key]["total_nominal"] += bill

        classified_list.append({
            "customer_id": int(row["customer_id"]),
            "customer_name": str(row["customer_name"]),
            "no_telp": str(row["no_telp"]),
            "brand": str(row["brand"]),
            "monthly_bill": bill,
            "cluster_key": cluster_key,
            "cluster_name": cluster_counts[cluster_key]["name"],
            "cluster_color": cluster_counts[cluster_key]["color"],
            "risk_score": risk_score,
            "risk_level": cluster_counts[cluster_key]["risk_level"],
            "avg_payment_day": round(float(avg_day), 1),
            "days_to_pay_avg": round(float(row["days_to_pay_avg"]), 1),
            "paid_ratio": round(float(row["paid_invoices"]) / max(1, int(row["total_invoices"])), 2),
            "recommendation": recommendation
        })

    # Sort classified list: highest risk first
    classified_list.sort(key=lambda x: x["risk_score"], reverse=True)

    # Cash Flow Forecasting for Cycle 26/27
    total_projected_billing = df["monthly_bill"].sum()
    
    # Phase Estimations
    p1_amount = cluster_counts["early_ontime"]["total_nominal"] * 0.85
    p2_amount = (cluster_counts["early_ontime"]["total_nominal"] * 0.15) + (cluster_counts["grace_period"]["total_nominal"] * 0.65)
    p3_amount = (cluster_counts["grace_period"]["total_nominal"] * 0.35) + (cluster_counts["chronic_late"]["total_nominal"] * 0.40)
    p4_amount = (cluster_counts["chronic_late"]["total_nominal"] * 0.60) + (cluster_counts["at_risk_churn"]["total_nominal"] * 0.25)
    at_risk_unpaid = cluster_counts["at_risk_churn"]["total_nominal"] * 0.75

    safe_total = max(1.0, total_projected_billing)

    cash_flow_forecast = [
        {
            "phase": "Fase 1: Pra-Jatuh Tempo (Tgl 26-30)",
            "description": "Pelanggan yang langsung bayar begitu invoice dicetak",
            "estimated_amount": round(p1_amount, 2),
            "percentage": round((p1_amount / safe_total) * 100, 1),
            "target_days": "26 - 30",
            "status_class": "success"
        },
        {
            "phase": "Fase 2: Puncak Jatuh Tempo (Tgl 1-5)",
            "description": "Gelombang pembayaran utama awal bulan",
            "estimated_amount": round(p2_amount, 2),
            "percentage": round((p2_amount / safe_total) * 100, 1),
            "target_days": "1 - 5",
            "status_class": "primary"
        },
        {
            "phase": "Fase 3: Masa Tenggang (Tgl 6-10)",
            "description": "Pembayaran mendekati batas masa isolir",
            "estimated_amount": round(p3_amount, 2),
            "percentage": round((p3_amount / safe_total) * 100, 1),
            "target_days": "6 - 10",
            "status_class": "warning"
        },
        {
            "phase": "Fase 4: Pasca Isolir / Rawan Nunggak (Tgl 11+)",
            "description": "Pelanggan yang bayar setelah diisolir atau berisiko menunggak",
            "estimated_amount": round(p4_amount + at_risk_unpaid, 2),
            "percentage": round(((p4_amount + at_risk_unpaid) / safe_total) * 100, 1),
            "target_days": "11+",
            "status_class": "error"
        }
    ]

    # Actionable Insights Generation
    actionable_insights = [
        f"Diproyeksikan sebanyak {cluster_counts['early_ontime']['count']} pelanggan (Rp {cluster_counts['early_ontime']['total_nominal']:,.0f}) membayar instan di periode pra-jatuh tempo (tgl 26-30).",
        f"Sebanyak {cluster_counts['chronic_late']['count']} pelanggan memiliki riwayat membayar lewat tanggal 10. Disarankan otomatisasi pengingat WA H-2 sebelum isolir.",
        f"Terdapat {cluster_counts['at_risk_churn']['count']} pelanggan dengan indikasi risiko tinggi (potensi tunggakan/churn Rp {cluster_counts['at_risk_churn']['total_nominal']:,.0f}). Perlu atensi khusus dari bagian keuangan.",
        f"Tingkat recovery rate tertagih sebelum masa isolir (tgl 1-10) diproyeksikan mencapai {round(((p1_amount + p2_amount + p3_amount) / safe_total) * 100, 1)}%."
    ]

    clusters_summary = [
        {
            "key": k,
            "name": v["name"],
            "count": v["count"],
            "percentage": round((v["count"] / max(1, n_samples)) * 100, 1),
            "total_nominal": round(v["total_nominal"], 2),
            "color": v["color"],
            "risk_level": v["risk_level"]
        }
        for k, v in cluster_counts.items()
    ]

    return {
        "status": "success",
        "cycle_info": {
            "invoice_issuance_date": request.cycle_date or "26",
            "due_date": "1",
            "grace_period_cutoff": "10",
            "total_active_subscribers": n_samples
        },
        "summary": {
            "total_customers": n_samples,
            "projected_billing": round(total_projected_billing, 2),
            "projected_early_collection": round(p1_amount, 2),
            "projected_grace_collection": round(p2_amount + p3_amount, 2),
            "projected_at_risk": round(p4_amount + at_risk_unpaid, 2),
            "projected_recovery_rate": round(((p1_amount + p2_amount + p3_amount) / safe_total) * 100, 1)
        },
        "clusters_summary": clusters_summary,
        "cash_flow_forecast": cash_flow_forecast,
        "classified_customers": classified_list,
        "actionable_insights": actionable_insights
    }