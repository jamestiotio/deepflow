START TRANSACTION;


ALTER TABLE ch_app_label ADD COLUMN updated_at DATETIME NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE ch_target_label ADD COLUMN updated_at DATETIME NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE ch_prometheus_label_name ADD COLUMN updated_at DATETIME NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE ch_prometheus_metric_name ADD COLUMN updated_at DATETIME NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE ch_prometheus_metric_app_label_layout ADD COLUMN updated_at DATETIME NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE ch_prometheus_target_label_layout ADD COLUMN updated_at DATETIME NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP;


-- update db_version to latest, remeber update DB_VERSION_EXPECT in migrate/init.go
UPDATE db_version SET version='6.3.1.13';
-- modify end

COMMIT;