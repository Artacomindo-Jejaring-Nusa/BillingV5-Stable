-- Migration: Add indexes for billing date performance optimization
-- In table 'langganan': tgl_jatuh_tempo
-- In table 'invoices': tgl_jatuh_tempo

CREATE INDEX idx_langganan_tgl_jatuh_tempo ON langganan(tgl_jatuh_tempo);
CREATE INDEX idx_invoices_tgl_jatuh_tempo ON invoices(tgl_jatuh_tempo);
