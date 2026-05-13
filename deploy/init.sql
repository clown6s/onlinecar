-- =============================================
-- OnlineCar 网约车平台数据库初始化脚本
-- =============================================

-- 司机库
CREATE DATABASE IF NOT EXISTS onlinecar_driver DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE onlinecar_driver;

DROP TABLE IF EXISTS `driver`;
CREATE TABLE `driver` (
    `id`        BIGINT       NOT NULL AUTO_INCREMENT COMMENT '司机ID',
    `name`      VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '姓名',
    `phone`     VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '手机号',
    `password`  VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码(加密)',
    `car_no`    VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '车牌号',
    `car_type`  VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '车型: 经济型/舒适型/豪华型',
    `lng`       DOUBLE       NOT NULL DEFAULT 0 COMMENT '经度',
    `lat`       DOUBLE       NOT NULL DEFAULT 0 COMMENT '纬度',
    `status`    TINYINT      NOT NULL DEFAULT 1 COMMENT '状态: 0=离线 1=空闲 2=接单中 3=行程中',
    `created_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_phone` (`phone`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='司机表';

-- 乘客库
CREATE DATABASE IF NOT EXISTS onlinecar_passenger DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE onlinecar_passenger;

DROP TABLE IF EXISTS `passenger_user`;
CREATE TABLE `passenger_user` (
  `id`               BIGINT(32) UNSIGNED NOT NULL AUTO_INCREMENT,
  `gmt_create`       DATETIME          NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  `gmt_modified`     DATETIME          NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  `passenger_phone`  VARCHAR(16)       CHARACTER SET utf8mb4 NULL DEFAULT NULL,
  `passenger_name`   VARCHAR(16)       CHARACTER SET utf8mb4 NULL DEFAULT NULL,
  `passenger_gender` TINYINT(1)        NULL DEFAULT NULL COMMENT '0：未知，1：男，2：女',
  `state`            TINYINT(1)        NULL DEFAULT NULL COMMENT '0：有效，1：失效',
  `profile_photo`    VARCHAR(128)      CHARACTER SET utf8mb4 NULL DEFAULT NULL COMMENT '头像图片地址的url',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COMMENT = '乘客用户表';

DROP TABLE IF EXISTS `trip`;
CREATE TABLE `trip` (
    `id`           BIGINT       NOT NULL AUTO_INCREMENT COMMENT '行程ID',
    `passenger_id` BIGINT       NOT NULL DEFAULT 0 COMMENT '乘客ID',
    `driver_id`    BIGINT       NOT NULL DEFAULT 0 COMMENT '司机ID',
    `origin`       VARCHAR(255) NOT NULL DEFAULT '' COMMENT '出发地',
    `destination`  VARCHAR(255) NOT NULL DEFAULT '' COMMENT '目的地',
    `origin_lng`   DOUBLE       NOT NULL DEFAULT 0 COMMENT '出发经度',
    `origin_lat`   DOUBLE       NOT NULL DEFAULT 0 COMMENT '出发纬度',
    `dest_lng`     DOUBLE       NOT NULL DEFAULT 0 COMMENT '目的经度',
    `dest_lat`     DOUBLE       NOT NULL DEFAULT 0 COMMENT '目的纬度',
    `status`       TINYINT      NOT NULL DEFAULT 1 COMMENT '状态: 1=待接单 2=已接单 3=行程中 4=已完成 5=已取消',
    `fare`         DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '费用',
    `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_passenger` (`passenger_id`),
    KEY `idx_driver` (`driver_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='行程表';
