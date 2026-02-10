package postgres

const (
	createTaskInTableId = "insert into public.check_table(create_at) values (now()) RETURNING check_id;"
	createTaskInProxy   = "insert into public.proxy(check_id, ip, port) values ($1, $2, $3) RETURNING proxy_id;"

	createTaskInProxyMetric = "insert into public.proxy_metric(check_id, proxy_id, type, status) values ($1, $2, $3, 'pending') returning proxy_metric_id;"

	selectTaskInWork = `select px.proxy_id, ct.check_id, host(px.ip), px.port, pm.proxy_metric_id, pm.type
		from check_table ct
			 join proxy px on px.check_id = ct.check_id
			 join proxy_metric pm on pm.proxy_id = px.proxy_id
		where pm.status = 'pending'
		order by ct.check_id, px.ip, px.port
		for update skip locked;
		`

	updateProxyMetric = `update public.proxy_metric
	set type   = $1,
    is_work=$2,
    speed=$3,
    status='checked'
	where proxy_metric_id = $4;
	`

	updateProxy = `update public.proxy
	set city   = $1,
    real_ip=$2::inet
	where proxy_id = $3;
	`

	getHistory = `
	SELECT ct.check_id, ct.create_at, COUNT(px.proxy_id) as proxy_count
	FROM check_table ct
	LEFT JOIN proxy px ON px.check_id = ct.check_id
	GROUP BY ct.check_id, ct.create_at
	ORDER BY ct.create_at DESC;`

	getStatusProxy = `
	SELECT ct.check_id, host(px.ip), px.port, COALESCE(px.city, ''), COALESCE(host(px.real_ip), ''),
	       COALESCE(pm.type, ''), COALESCE(pm.is_work, false), COALESCE(pm.speed, 0), pm.status
	FROM check_table ct
         JOIN proxy px ON px.check_id = ct.check_id
         JOIN proxy_metric pm ON pm.proxy_id = px.proxy_id
	WHERE ct.check_id = $1;`
)
