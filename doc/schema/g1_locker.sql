-- locker G1 · 快递柜格口 + 包裹（模型看到的最终表结构，照此写代码，不新建/不改表）
CREATE TABLE IF NOT EXISTS t_locker_cell (
  id          BIGINT       NOT NULL AUTO_INCREMENT,
  cell_no     VARCHAR(32)  NOT NULL,
  size        VARCHAR(16)  NOT NULL,
  status      VARCHAR(16)  NOT NULL DEFAULT 'FREE',
  create_by   VARCHAR(64)  NULL,
  create_time DATETIME     NULL,
  update_by   VARCHAR(64)  NULL,
  update_time DATETIME     NULL,
  deleted_at  DATETIME(3)  NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_cell_no (cell_no),
  KEY idx_cell_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS t_parcel (
  id          BIGINT       NOT NULL AUTO_INCREMENT,
  parcel_no   VARCHAR(32)  NOT NULL,
  cell_id     BIGINT       NOT NULL,
  phone       VARCHAR(32)  NOT NULL,
  size        VARCHAR(16)  NOT NULL,
  status      VARCHAR(16)  NOT NULL DEFAULT 'STORED',
  stored_at   DATETIME     NULL,
  picked_at   DATETIME     NULL,
  create_by   VARCHAR(64)  NULL,
  create_time DATETIME     NULL,
  update_by   VARCHAR(64)  NULL,
  update_time DATETIME     NULL,
  deleted_at  DATETIME(3)  NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_parcel_no (parcel_no),
  KEY idx_parcel_cell (cell_id),
  KEY idx_parcel_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
