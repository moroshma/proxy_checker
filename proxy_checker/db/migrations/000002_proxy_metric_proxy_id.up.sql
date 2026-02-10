ALTER TABLE proxy_metric ADD COLUMN proxy_id UUID REFERENCES proxy(proxy_id);
