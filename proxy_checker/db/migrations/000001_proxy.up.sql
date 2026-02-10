CREATE TABLE IF NOT EXISTS check_table
(
    check_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    create_at timestamp        default NOW()
);

CREATE TABLE IF NOT EXISTS proxy
(
    proxy_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    check_id UUID REFERENCES check_table (check_id),
    ip       inet,
    port     int,
    city     varchar(255),
    real_ip  inet
);





CREATE TABLE IF NOT EXISTS proxy_metric
(
    proxy_metric_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    check_id        UUID REFERENCES check_table (check_id),
    type            varchar(255),
    is_work         boolean,
    speed           integer,
    status          varchar(255) default 'pending'
);