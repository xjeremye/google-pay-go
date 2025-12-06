/*
 Navicat Premium Data Transfer

 Source Server         : zzxxzxzx
 Source Server Type    : MySQL
 Source Server Version : 80036 (8.0.36)
 Source Host           : rm-1ud2pn3f73r0d15t2oo.mysql.rds.aliyuncs.com:3306
 Source Schema         : django-vue3-admin

 Target Server Type    : MySQL
 Target Server Version : 80036 (8.0.36)
 File Encoding         : 65001

 Date: 01/12/2025 20:45:41
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for auth_group
-- ----------------------------
DROP TABLE IF EXISTS `auth_group`;
CREATE TABLE `auth_group` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(150) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for auth_group_permissions
-- ----------------------------
DROP TABLE IF EXISTS `auth_group_permissions`;
CREATE TABLE `auth_group_permissions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `group_id` int NOT NULL,
  `permission_id` int NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `auth_group_permissions_group_id_permission_id_0cd325b0_uniq` (`group_id`,`permission_id`),
  KEY `auth_group_permissio_permission_id_84c5c92e_fk_auth_perm` (`permission_id`),
  CONSTRAINT `auth_group_permissio_permission_id_84c5c92e_fk_auth_perm` FOREIGN KEY (`permission_id`) REFERENCES `auth_permission` (`id`),
  CONSTRAINT `auth_group_permissions_group_id_b120cbf9_fk_auth_group_id` FOREIGN KEY (`group_id`) REFERENCES `auth_group` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for auth_permission
-- ----------------------------
DROP TABLE IF EXISTS `auth_permission`;
CREATE TABLE `auth_permission` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `content_type_id` int NOT NULL,
  `codename` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `auth_permission_content_type_id_codename_01ab375a_uniq` (`content_type_id`,`codename`),
  CONSTRAINT `auth_permission_content_type_id_2f476e4b_fk_django_co` FOREIGN KEY (`content_type_id`) REFERENCES `django_content_type` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=801 DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for captcha_captchastore
-- ----------------------------
DROP TABLE IF EXISTS `captcha_captchastore`;
CREATE TABLE `captcha_captchastore` (
  `id` int NOT NULL AUTO_INCREMENT,
  `challenge` varchar(32) NOT NULL,
  `response` varchar(32) NOT NULL,
  `hashkey` varchar(40) NOT NULL,
  `expiration` datetime(6) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `hashkey` (`hashkey`)
) ENGINE=InnoDB AUTO_INCREMENT=682 DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for django_apscheduler_djangojob
-- ----------------------------
DROP TABLE IF EXISTS `django_apscheduler_djangojob`;
CREATE TABLE `django_apscheduler_djangojob` (
  `id` varchar(255) NOT NULL,
  `next_run_time` datetime(6) DEFAULT NULL,
  `job_state` longblob NOT NULL,
  PRIMARY KEY (`id`),
  KEY `django_apscheduler_djangojob_next_run_time_2f022619` (`next_run_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for django_apscheduler_djangojobexecution
-- ----------------------------
DROP TABLE IF EXISTS `django_apscheduler_djangojobexecution`;
CREATE TABLE `django_apscheduler_djangojobexecution` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `status` varchar(50) NOT NULL,
  `run_time` datetime(6) NOT NULL,
  `duration` decimal(15,2) DEFAULT NULL,
  `finished` decimal(15,2) DEFAULT NULL,
  `exception` varchar(1000) DEFAULT NULL,
  `traceback` longtext,
  `job_id` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_job_executions` (`job_id`,`run_time`),
  KEY `django_apscheduler_djangojobexecution_run_time_16edd96b` (`run_time`),
  CONSTRAINT `django_apscheduler_djangojobexecution_job_id_daf5090a_fk` FOREIGN KEY (`job_id`) REFERENCES `django_apscheduler_djangojob` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for django_celery_beat_clockedschedule
-- ----------------------------
DROP TABLE IF EXISTS `django_celery_beat_clockedschedule`;
CREATE TABLE `django_celery_beat_clockedschedule` (
  `id` int NOT NULL AUTO_INCREMENT,
  `clocked_time` datetime(6) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for django_celery_beat_crontabschedule
-- ----------------------------
DROP TABLE IF EXISTS `django_celery_beat_crontabschedule`;
CREATE TABLE `django_celery_beat_crontabschedule` (
  `id` int NOT NULL AUTO_INCREMENT,
  `minute` varchar(240) NOT NULL,
  `hour` varchar(96) NOT NULL,
  `day_of_week` varchar(64) NOT NULL,
  `day_of_month` varchar(124) NOT NULL,
  `month_of_year` varchar(64) NOT NULL,
  `timezone` varchar(63) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for django_celery_beat_intervalschedule
-- ----------------------------
DROP TABLE IF EXISTS `django_celery_beat_intervalschedule`;
CREATE TABLE `django_celery_beat_intervalschedule` (
  `id` int NOT NULL AUTO_INCREMENT,
  `every` int NOT NULL,
  `period` varchar(24) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for django_celery_beat_periodictask
-- ----------------------------
DROP TABLE IF EXISTS `django_celery_beat_periodictask`;
CREATE TABLE `django_celery_beat_periodictask` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(200) NOT NULL,
  `task` varchar(200) NOT NULL,
  `args` longtext NOT NULL,
  `kwargs` longtext NOT NULL,
  `queue` varchar(200) DEFAULT NULL,
  `exchange` varchar(200) DEFAULT NULL,
  `routing_key` varchar(200) DEFAULT NULL,
  `expires` datetime(6) DEFAULT NULL,
  `enabled` tinyint(1) NOT NULL,
  `last_run_at` datetime(6) DEFAULT NULL,
  `total_run_count` int unsigned NOT NULL,
  `date_changed` datetime(6) NOT NULL,
  `description` longtext NOT NULL,
  `crontab_id` int DEFAULT NULL,
  `interval_id` int DEFAULT NULL,
  `solar_id` int DEFAULT NULL,
  `one_off` tinyint(1) NOT NULL,
  `start_time` datetime(6) DEFAULT NULL,
  `priority` int unsigned DEFAULT NULL,
  `headers` longtext NOT NULL DEFAULT (_utf8mb4'{}'),
  `clocked_id` int DEFAULT NULL,
  `expire_seconds` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  KEY `django_celery_beat_p_crontab_id_d3cba168_fk_django_ce` (`crontab_id`),
  KEY `django_celery_beat_p_interval_id_a8ca27da_fk_django_ce` (`interval_id`),
  KEY `django_celery_beat_p_solar_id_a87ce72c_fk_django_ce` (`solar_id`),
  KEY `django_celery_beat_p_clocked_id_47a69f82_fk_django_ce` (`clocked_id`),
  CONSTRAINT `django_celery_beat_p_clocked_id_47a69f82_fk_django_ce` FOREIGN KEY (`clocked_id`) REFERENCES `django_celery_beat_clockedschedule` (`id`),
  CONSTRAINT `django_celery_beat_p_crontab_id_d3cba168_fk_django_ce` FOREIGN KEY (`crontab_id`) REFERENCES `django_celery_beat_crontabschedule` (`id`),
  CONSTRAINT `django_celery_beat_p_interval_id_a8ca27da_fk_django_ce` FOREIGN KEY (`interval_id`) REFERENCES `django_celery_beat_intervalschedule` (`id`),
  CONSTRAINT `django_celery_beat_p_solar_id_a87ce72c_fk_django_ce` FOREIGN KEY (`solar_id`) REFERENCES `django_celery_beat_solarschedule` (`id`),
  CONSTRAINT `django_celery_beat_periodictask_chk_1` CHECK ((`total_run_count` >= 0)),
  CONSTRAINT `django_celery_beat_periodictask_chk_2` CHECK ((`priority` >= 0)),
  CONSTRAINT `django_celery_beat_periodictask_chk_3` CHECK ((`expire_seconds` >= 0))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for django_celery_beat_periodictasks
-- ----------------------------
DROP TABLE IF EXISTS `django_celery_beat_periodictasks`;
CREATE TABLE `django_celery_beat_periodictasks` (
  `ident` smallint NOT NULL,
  `last_update` datetime(6) NOT NULL,
  PRIMARY KEY (`ident`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for django_celery_beat_solarschedule
-- ----------------------------
DROP TABLE IF EXISTS `django_celery_beat_solarschedule`;
CREATE TABLE `django_celery_beat_solarschedule` (
  `id` int NOT NULL AUTO_INCREMENT,
  `event` varchar(24) NOT NULL,
  `latitude` decimal(9,6) NOT NULL,
  `longitude` decimal(9,6) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `django_celery_beat_solar_event_latitude_longitude_ba64999a_uniq` (`event`,`latitude`,`longitude`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for django_celery_results_chordcounter
-- ----------------------------
DROP TABLE IF EXISTS `django_celery_results_chordcounter`;
CREATE TABLE `django_celery_results_chordcounter` (
  `id` int NOT NULL AUTO_INCREMENT,
  `group_id` varchar(255) NOT NULL,
  `sub_tasks` longtext NOT NULL,
  `count` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `group_id` (`group_id`),
  CONSTRAINT `django_celery_results_chordcounter_chk_1` CHECK ((`count` >= 0))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for django_celery_results_groupresult
-- ----------------------------
DROP TABLE IF EXISTS `django_celery_results_groupresult`;
CREATE TABLE `django_celery_results_groupresult` (
  `id` int NOT NULL AUTO_INCREMENT,
  `group_id` varchar(255) NOT NULL,
  `date_created` datetime(6) NOT NULL,
  `date_done` datetime(6) NOT NULL,
  `content_type` varchar(128) NOT NULL,
  `content_encoding` varchar(64) NOT NULL,
  `result` longtext,
  PRIMARY KEY (`id`),
  UNIQUE KEY `group_id` (`group_id`),
  KEY `django_cele_date_cr_bd6c1d_idx` (`date_created`),
  KEY `django_cele_date_do_caae0e_idx` (`date_done`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for django_celery_results_taskresult
-- ----------------------------
DROP TABLE IF EXISTS `django_celery_results_taskresult`;
CREATE TABLE `django_celery_results_taskresult` (
  `id` int NOT NULL AUTO_INCREMENT,
  `task_id` varchar(255) NOT NULL,
  `status` varchar(50) NOT NULL,
  `content_type` varchar(128) NOT NULL,
  `content_encoding` varchar(64) NOT NULL,
  `result` longtext,
  `date_done` datetime(6) NOT NULL,
  `traceback` longtext,
  `meta` longtext,
  `task_args` longtext,
  `task_kwargs` longtext,
  `task_name` varchar(255) DEFAULT NULL,
  `worker` varchar(100) DEFAULT NULL,
  `date_created` datetime(6) NOT NULL,
  `periodic_task_name` varchar(255) DEFAULT NULL,
  `date_started` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `task_id` (`task_id`),
  KEY `django_cele_task_na_08aec9_idx` (`task_name`),
  KEY `django_cele_status_9b6201_idx` (`status`),
  KEY `django_cele_worker_d54dd8_idx` (`worker`),
  KEY `django_cele_date_cr_f04a50_idx` (`date_created`),
  KEY `django_cele_date_do_f59aad_idx` (`date_done`),
  KEY `django_cele_periodi_1993cf_idx` (`periodic_task_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for django_content_type
-- ----------------------------
DROP TABLE IF EXISTS `django_content_type`;
CREATE TABLE `django_content_type` (
  `id` int NOT NULL AUTO_INCREMENT,
  `app_label` varchar(100) NOT NULL,
  `model` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `django_content_type_app_label_model_76bd3d3b_uniq` (`app_label`,`model`)
) ENGINE=InnoDB AUTO_INCREMENT=201 DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for django_migrations
-- ----------------------------
DROP TABLE IF EXISTS `django_migrations`;
CREATE TABLE `django_migrations` (
  `id` int NOT NULL AUTO_INCREMENT,
  `app` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `applied` datetime(6) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=76 DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for django_session
-- ----------------------------
DROP TABLE IF EXISTS `django_session`;
CREATE TABLE `django_session` (
  `session_key` varchar(40) NOT NULL,
  `session_data` longtext NOT NULL,
  `expire_date` datetime(6) NOT NULL,
  PRIMARY KEY (`session_key`),
  KEY `django_session_expire_date_a5c62663` (`expire_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_agent_payment
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_agent_payment`;
CREATE TABLE `dvadmin_agent_payment` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `url` longtext NOT NULL COMMENT '代付链接',
  `money` bigint NOT NULL DEFAULT '0' COMMENT '金额',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态',
  `ticket_no` varchar(64) DEFAULT NULL COMMENT '订单号',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  `order_id` varchar(30) DEFAULT NULL COMMENT '关联订单',
  PRIMARY KEY (`id`),
  KEY `dvadmin_agent_payment_creator_id_a924fa0b` (`creator_id`),
  KEY `dvadmin_agent_payment_writeoff_id_8fb1674c` (`writeoff_id`),
  KEY `dvadmin_agent_payment_order_id_927ccd0d` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='代付';

-- ----------------------------
-- Table structure for dvadmin_alipay_complain
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_complain`;
CREATE TABLE `dvadmin_alipay_complain` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `status` varchar(128) NOT NULL COMMENT '投诉单状态',
  `process_code` varchar(128) NOT NULL COMMENT '商家处理结果码',
  `opposite_name` varchar(128) NOT NULL COMMENT '被投诉人名称',
  `gmt_process` datetime(6) DEFAULT NULL COMMENT '投诉单处理时间',
  `gmt_complain` datetime(6) DEFAULT NULL COMMENT '投诉单创建时间',
  `gmt_overdue` datetime(6) DEFAULT NULL COMMENT '投诉单过期时间',
  `gmt_risk_finish_time` datetime(6) DEFAULT NULL COMMENT '推送时间',
  `complain_content` longtext COMMENT '投诉内容',
  `process_remark` longtext COMMENT '商家处理备注',
  `process_img_url_list` json DEFAULT NULL COMMENT '商家处理备注图片url列表',
  `trade_no` varchar(1024) NOT NULL COMMENT '支付宝单号',
  `task_id` varchar(32) NOT NULL COMMENT '投诉单号id',
  `contact` varchar(128) DEFAULT NULL COMMENT '联系方式',
  `complain_amount` bigint NOT NULL COMMENT '投诉金额',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `alipay_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dvadmin_alipay_complain_creator_id_8dbf2040` (`creator_id`),
  KEY `dvadmin_alipay_complain_alipay_id_7a7a0d6e` (`alipay_id`)
) ENGINE=InnoDB AUTO_INCREMENT=282789822 DEFAULT CHARSET=utf8mb3 COMMENT='支付宝投诉';

-- ----------------------------
-- Table structure for dvadmin_alipay_complain_info
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_complain_info`;
CREATE TABLE `dvadmin_alipay_complain_info` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(32) NOT NULL COMMENT '投诉单号id',
  `trade_no` varchar(1024) NOT NULL COMMENT '支付宝单号',
  `out_no` varchar(128) NOT NULL COMMENT '商家订单号',
  `gmt_trade` datetime(6) DEFAULT NULL COMMENT '交易时间',
  `gmt_refund` datetime(6) DEFAULT NULL COMMENT '退款时间',
  `status` varchar(128) NOT NULL COMMENT '状态',
  `amount` bigint NOT NULL COMMENT '投诉金额',
  `complaint_record_id` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_alipay_complain_info_complaint_record_id_87ceb25a` (`complaint_record_id`),
  KEY `dvadmin_alipay_complain_info_creator_id_93e82a88` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝投诉涉及';

-- ----------------------------
-- Table structure for dvadmin_alipay_product
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_product`;
CREATE TABLE `dvadmin_alipay_product` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `name` varchar(255) NOT NULL COMMENT '项目名称',
  `account_type` int NOT NULL COMMENT '账户类型',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `uid` varchar(255) DEFAULT NULL COMMENT '支付宝UID',
  `app_id` varchar(255) DEFAULT NULL COMMENT '应用ID',
  `limit_money` int NOT NULL COMMENT '限额',
  `max_money` int NOT NULL COMMENT '最大金额',
  `min_money` int NOT NULL COMMENT '最小金额',
  `float_max_money` int NOT NULL COMMENT '浮动最大金额',
  `float_min_money` int NOT NULL COMMENT '浮动最小金额',
  `collection_type` int NOT NULL DEFAULT '2' COMMENT '收账类型',
  `sign_type` int NOT NULL DEFAULT '0' COMMENT '签名类型',
  `public_key` longtext COMMENT '支付宝公钥',
  `private_key` longtext COMMENT '应用私钥',
  `app_public_crt` longtext COMMENT '应用公钥证书',
  `alipay_public_crt` longtext COMMENT '支付宝公钥证书',
  `alipay_root_crt` longtext COMMENT '支付宝根证书',
  `can_pay` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否允许进单',
  `ver` bigint NOT NULL,
  `email` varchar(254) DEFAULT NULL COMMENT '邮箱',
  `pwd` varchar(100) DEFAULT NULL COMMENT '邮箱密码',
  `proxy_ip` varchar(255) DEFAULT NULL COMMENT '代理ip',
  `proxy_port` int NOT NULL DEFAULT '0' COMMENT '代理端口',
  `proxy_user` varchar(255) DEFAULT NULL COMMENT '代理用户名',
  `proxy_pwd` varchar(255) DEFAULT NULL COMMENT '代理密码',
  `settled_moneys` json NOT NULL COMMENT '固定金额列表',
  `max_fail_count` int NOT NULL DEFAULT '0' COMMENT '最多失败次数',
  `ip_day_limit` int NOT NULL DEFAULT '0' COMMENT '同IP一天内支付次数',
  `user_id_day_limit` int NOT NULL DEFAULT '0' COMMENT '同UserId一天内支付次数',
  `day_count_limit` int NOT NULL DEFAULT '0' COMMENT '日笔数限制',
  `subject` varchar(100) DEFAULT NULL,
  `proceeds_qr` longtext COMMENT '收款二维码',
  `app_auth_token` varchar(48) DEFAULT NULL COMMENT '服务商授权',
  `is_delete` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否删除',
  `split_async` tinyint(1) NOT NULL DEFAULT '1' COMMENT '分账是否异步',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `parent_id` bigint DEFAULT NULL COMMENT '父级',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  KEY `dvadmin_alipay_product_creator_id_50f2d526` (`creator_id`),
  KEY `dvadmin_alipay_product_parent_id_78943e04` (`parent_id`),
  KEY `dvadmin_alipay_product_writeoff_id_d88dfd3e` (`writeoff_id`)
) ENGINE=InnoDB AUTO_INCREMENT=45 DEFAULT CHARSET=utf8mb3 COMMENT='支付宝项目';

-- ----------------------------
-- Table structure for dvadmin_alipay_product_allow_pay_channels
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_product_allow_pay_channels`;
CREATE TABLE `dvadmin_alipay_product_allow_pay_channels` (
  `id` int NOT NULL AUTO_INCREMENT,
  `alipayproduct_id` bigint NOT NULL,
  `paychannel_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_alipay_product_a_alipayproduct_id_paychan_d59143e8_uniq` (`alipayproduct_id`,`paychannel_id`),
  KEY `dvadmin_alipay_product_allo_alipayproduct_id_1db10670` (`alipayproduct_id`),
  KEY `dvadmin_alipay_product_allow_pay_channels_paychannel_id_e9d21fe0` (`paychannel_id`)
) ENGINE=InnoDB AUTO_INCREMENT=56 DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_alipay_product_c2c_groups
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_product_c2c_groups`;
CREATE TABLE `dvadmin_alipay_product_c2c_groups` (
  `id` int NOT NULL AUTO_INCREMENT,
  `alipayproduct_id` bigint NOT NULL,
  `alipaysplitusergroup_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_alipay_product_c_alipayproduct_id_alipays_b5c7d4f6_uniq` (`alipayproduct_id`,`alipaysplitusergroup_id`),
  KEY `dvadmin_alipay_product_c2c_groups_alipayproduct_id_397c77ff` (`alipayproduct_id`),
  KEY `dvadmin_alipay_product_c2c__alipaysplitusergroup_id_cea98bb9` (`alipaysplitusergroup_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_alipay_product_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_product_day`;
CREATE TABLE `dvadmin_alipay_product_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `product_id` bigint DEFAULT NULL COMMENT '关联项目',
  `pay_channel_id` bigint DEFAULT NULL COMMENT '关联通道',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_alipay_product_d_date_product_id_pay_chan_fbfb3b3c_uniq` (`date`,`product_id`,`pay_channel_id`),
  KEY `dvadmin_alipay_product_day_product_id_143b1e8f` (`product_id`),
  KEY `dvadmin_alipay_product_day_pay_channel_id_0ca39bf0` (`pay_channel_id`)
) ENGINE=InnoDB AUTO_INCREMENT=75 DEFAULT CHARSET=utf8mb3 COMMENT='支付宝项目每日统计';

-- ----------------------------
-- Table structure for dvadmin_alipay_product_groups
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_product_groups`;
CREATE TABLE `dvadmin_alipay_product_groups` (
  `id` int NOT NULL AUTO_INCREMENT,
  `alipayproduct_id` bigint NOT NULL,
  `alipaysplitusergroup_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_alipay_product_g_alipayproduct_id_alipays_22ea9b44_uniq` (`alipayproduct_id`,`alipaysplitusergroup_id`),
  KEY `dvadmin_alipay_product_groups_alipayproduct_id_5c7a1ae2` (`alipayproduct_id`),
  KEY `dvadmin_alipay_product_groups_alipaysplitusergroup_id_66a10387` (`alipaysplitusergroup_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_alipay_product_tag
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_product_tag`;
CREATE TABLE `dvadmin_alipay_product_tag` (
  `name` varchar(32) NOT NULL COMMENT '标签名称',
  `sort` int NOT NULL DEFAULT '6' COMMENT '排序',
  `system_user_id` bigint DEFAULT NULL COMMENT '关联用户',
  PRIMARY KEY (`name`),
  KEY `dvadmin_alipay_product_tag_system_user_id_14967382` (`system_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝项目标签';

-- ----------------------------
-- Table structure for dvadmin_alipay_product_tags
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_product_tags`;
CREATE TABLE `dvadmin_alipay_product_tags` (
  `id` int NOT NULL AUTO_INCREMENT,
  `alipayproduct_id` bigint NOT NULL,
  `alipayproducttag_id` varchar(32) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_alipay_product_t_alipayproduct_id_alipayp_6f2b6d15_uniq` (`alipayproduct_id`,`alipayproducttag_id`),
  KEY `dvadmin_alipay_product_tags_alipayproduct_id_f44734a6` (`alipayproduct_id`),
  KEY `dvadmin_alipay_product_tags_alipayproducttag_id_667bd071` (`alipayproducttag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_alipay_product_transfer_groups
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_product_transfer_groups`;
CREATE TABLE `dvadmin_alipay_product_transfer_groups` (
  `id` int NOT NULL AUTO_INCREMENT,
  `alipayproduct_id` bigint NOT NULL,
  `alipaysplitusergroup_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_alipay_product_t_alipayproduct_id_alipays_14e7b206_uniq` (`alipayproduct_id`,`alipaysplitusergroup_id`),
  KEY `dvadmin_alipay_product_transfer_groups_alipayproduct_id_9a5cfec1` (`alipayproduct_id`),
  KEY `dvadmin_alipay_product_tran_alipaysplitusergroup_id_0b1b36ce` (`alipaysplitusergroup_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_alipay_public_pool
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_public_pool`;
CREATE TABLE `dvadmin_alipay_public_pool` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `is_delete` tinyint(1) NOT NULL DEFAULT '0' COMMENT '软删除',
  `alipay_id` bigint NOT NULL COMMENT '支付宝主体',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_alipay_public_pool_is_delete_03386749` (`is_delete`),
  KEY `dvadmin_alipay_public_pool_alipay_id_cf707c49` (`alipay_id`),
  KEY `dvadmin_alipay_public_pool_creator_id_9606270b` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝公池';

-- ----------------------------
-- Table structure for dvadmin_alipay_public_pool_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_public_pool_day`;
CREATE TABLE `dvadmin_alipay_public_pool_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `pool_id` bigint DEFAULT NULL COMMENT '关联项目',
  `pay_channel_id` bigint DEFAULT NULL COMMENT '关联通道',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_alipay_public_po_date_pool_id_pay_channel_f7236ac4_uniq` (`date`,`pool_id`,`pay_channel_id`),
  KEY `dvadmin_alipay_public_pool_day_pool_id_a26f4050` (`pool_id`),
  KEY `dvadmin_alipay_public_pool_day_pay_channel_id_a7245d28` (`pay_channel_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝公池每日统计';

-- ----------------------------
-- Table structure for dvadmin_alipay_quick_transfer
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_quick_transfer`;
CREATE TABLE `dvadmin_alipay_quick_transfer` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `auto` tinyint(1) NOT NULL DEFAULT '0' COMMENT '自动划转',
  `lower_limit` bigint NOT NULL DEFAULT '100000' COMMENT '金额下限',
  `run_interval` int NOT NULL DEFAULT '800' COMMENT '执行间隔毫秒',
  `money` bigint NOT NULL DEFAULT '4900' COMMENT '转账金额',
  `alipay_id` bigint DEFAULT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  UNIQUE KEY `alipay_id` (`alipay_id`),
  KEY `dvadmin_alipay_quick_transfer_creator_id_07332e66` (`creator_id`)
) ENGINE=InnoDB AUTO_INCREMENT=45 DEFAULT CHARSET=utf8mb3 COMMENT='支付宝快速转账设置';

-- ----------------------------
-- Table structure for dvadmin_alipay_safe_transfer
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_safe_transfer`;
CREATE TABLE `dvadmin_alipay_safe_transfer` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `external_agreement_no` varchar(32) NOT NULL COMMENT '本系统中的协议号',
  `agreement_no` varchar(32) DEFAULT NULL COMMENT '支付宝协议号',
  `account_book_id` varchar(16) DEFAULT NULL COMMENT '资金记账本id',
  `alipay_id` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `parent_alipay_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `external_agreement_no` (`external_agreement_no`),
  UNIQUE KEY `alipay_id` (`alipay_id`),
  KEY `dvadmin_alipay_safe_transfer_creator_id_8c9f5856` (`creator_id`),
  KEY `dvadmin_alipay_safe_transfer_parent_alipay_id_0b3fe8f2` (`parent_alipay_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝安全发设置';

-- ----------------------------
-- Table structure for dvadmin_alipay_settle_confirm_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_settle_confirm_history`;
CREATE TABLE `dvadmin_alipay_settle_confirm_history` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(25) NOT NULL COMMENT '请求订单号',
  `trade_no` varchar(100) DEFAULT NULL COMMENT '支付宝交易号',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否完成',
  `alipay_id` bigint NOT NULL COMMENT '关联支付宝主体',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `order_id` varchar(30) NOT NULL COMMENT '关联订单号',
  PRIMARY KEY (`id`),
  KEY `dvadmin_alipay_settle_confirm_history_alipay_id_71b9b900` (`alipay_id`),
  KEY `dvadmin_alipay_settle_confirm_history_creator_id_bf78ac75` (`creator_id`),
  KEY `dvadmin_alipay_settle_confirm_history_order_id_193414c4` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝结算记录';

-- ----------------------------
-- Table structure for dvadmin_alipay_shenma
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_shenma`;
CREATE TABLE `dvadmin_alipay_shenma` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `limit_money` int NOT NULL DEFAULT '0' COMMENT '日限额',
  `ver` bigint NOT NULL,
  `alipay_id` bigint NOT NULL COMMENT '父级',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `tenant_id` bigint NOT NULL COMMENT '共享租户',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_alipay_shenma_tenant_id_alipay_id_76fbf419_uniq` (`tenant_id`,`alipay_id`),
  KEY `dvadmin_alipay_shenma_alipay_id_b79457e7` (`alipay_id`),
  KEY `dvadmin_alipay_shenma_creator_id_74d5b866` (`creator_id`),
  KEY `dvadmin_alipay_shenma_tenant_id_556d41fa` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝神码';

-- ----------------------------
-- Table structure for dvadmin_alipay_shenma_allow_pay_channels
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_shenma_allow_pay_channels`;
CREATE TABLE `dvadmin_alipay_shenma_allow_pay_channels` (
  `id` int NOT NULL AUTO_INCREMENT,
  `alipayshenma_id` bigint NOT NULL,
  `paychannel_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_alipay_shenma_al_alipayshenma_id_paychann_b2b8024e_uniq` (`alipayshenma_id`,`paychannel_id`),
  KEY `dvadmin_alipay_shenma_allow_alipayshenma_id_c90a8e6d` (`alipayshenma_id`),
  KEY `dvadmin_alipay_shenma_allow_pay_channels_paychannel_id_71b8daa3` (`paychannel_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_alipay_shenma_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_shenma_day`;
CREATE TABLE `dvadmin_alipay_shenma_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `shenma_id` bigint DEFAULT NULL COMMENT '关联项目',
  `pay_channel_id` bigint DEFAULT NULL COMMENT '关联通道',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_alipay_shenma_da_date_shenma_id_pay_chann_bda1ee91_uniq` (`date`,`shenma_id`,`pay_channel_id`),
  KEY `dvadmin_alipay_shenma_day_shenma_id_16ab3514` (`shenma_id`),
  KEY `dvadmin_alipay_shenma_day_pay_channel_id_e839ee65` (`pay_channel_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝项目每日统计';

-- ----------------------------
-- Table structure for dvadmin_alipay_split_user
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_split_user`;
CREATE TABLE `dvadmin_alipay_split_user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `username_type` int NOT NULL COMMENT '用户名类型',
  `username` varchar(255) NOT NULL COMMENT '账号',
  `name` varchar(255) DEFAULT NULL COMMENT '姓名',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `limit_money` bigint NOT NULL DEFAULT '0' COMMENT '总限额',
  `percentage` decimal(5,2) NOT NULL DEFAULT '100.00' COMMENT '分账百分比',
  `risk` int NOT NULL DEFAULT '0' COMMENT '风控,0无,1潜在,2高风险',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `group_id` bigint NOT NULL COMMENT '关联用户组',
  PRIMARY KEY (`id`),
  KEY `dvadmin_alipay_split_user_creator_id_bc4a8b65` (`creator_id`),
  KEY `dvadmin_alipay_split_user_group_id_2fca0658` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝分账用户';

-- ----------------------------
-- Table structure for dvadmin_alipay_split_user_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_split_user_flow`;
CREATE TABLE `dvadmin_alipay_split_user_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `flow` bigint NOT NULL COMMENT '流水',
  `alipay_product_id` bigint DEFAULT NULL COMMENT '关联项目',
  `alipay_user_id` bigint DEFAULT NULL COMMENT '关联接收人',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `tenant_id` bigint DEFAULT NULL COMMENT '租户',
  PRIMARY KEY (`id`),
  KEY `dvadmin_alipay_split_user_flow_alipay_product_id_d6322dab` (`alipay_product_id`),
  KEY `dvadmin_alipay_split_user_flow_alipay_user_id_796a5042` (`alipay_user_id`),
  KEY `dvadmin_alipay_split_user_flow_creator_id_cbe4ebf1` (`creator_id`),
  KEY `dvadmin_alipay_split_user_flow_tenant_id_9d4db788` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝分账用户流水';

-- ----------------------------
-- Table structure for dvadmin_alipay_split_user_group
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_split_user_group`;
CREATE TABLE `dvadmin_alipay_split_user_group` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `name` varchar(128) NOT NULL COMMENT '组名',
  `telegram` varchar(256) DEFAULT NULL COMMENT 'tg',
  `pre_status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '预付模式',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `weight` int NOT NULL DEFAULT '1' COMMENT '权重',
  `tax` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '费率',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `tenant_id` bigint NOT NULL COMMENT '租户',
  `writeoff_id` bigint DEFAULT NULL COMMENT '核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_alipay_split_user_group_creator_id_ff3b465b` (`creator_id`),
  KEY `dvadmin_alipay_split_user_group_tenant_id_a7082157` (`tenant_id`),
  KEY `dvadmin_alipay_split_user_group_writeoff_id_e7a2acb5` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝用户组';

-- ----------------------------
-- Table structure for dvadmin_alipay_split_user_group_add_money
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_split_user_group_add_money`;
CREATE TABLE `dvadmin_alipay_split_user_group_add_money` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `date` date NOT NULL COMMENT '日期',
  `add_money` bigint NOT NULL DEFAULT '0' COMMENT '打款',
  `ver` bigint NOT NULL,
  `group_id` bigint NOT NULL COMMENT '组',
  PRIMARY KEY (`id`),
  KEY `dvadmin_alipay_split_user_group_add_money_group_id_9a479063` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝用户组预付打款';

-- ----------------------------
-- Table structure for dvadmin_alipay_split_user_group_pre
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_split_user_group_pre`;
CREATE TABLE `dvadmin_alipay_split_user_group_pre` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `pre_pay` bigint NOT NULL DEFAULT '0' COMMENT '预付金额',
  `ver` bigint NOT NULL,
  `group_id` bigint NOT NULL COMMENT '关联用户组',
  PRIMARY KEY (`id`),
  UNIQUE KEY `group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝用户组预付';

-- ----------------------------
-- Table structure for dvadmin_alipay_split_user_group_prehistory
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_split_user_group_prehistory`;
CREATE TABLE `dvadmin_alipay_split_user_group_prehistory` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `pre_pay` bigint NOT NULL DEFAULT '0' COMMENT '预付金额',
  `before` bigint NOT NULL DEFAULT '0' COMMENT '改动前金额',
  `after` bigint NOT NULL DEFAULT '0' COMMENT '改动后金额',
  `user` varchar(255) DEFAULT NULL,
  `ver` bigint NOT NULL,
  `rate` varchar(32) DEFAULT '0' COMMENT 'usdt汇率',
  `usdt` varchar(32) DEFAULT '0' COMMENT 'usdt',
  `cert` longtext COMMENT '转账凭证',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `group_id` bigint NOT NULL COMMENT '关联用户组',
  PRIMARY KEY (`id`),
  KEY `dvadmin_alipay_split_user_group_prehistory_creator_id_043b74e8` (`creator_id`),
  KEY `dvadmin_alipay_split_user_group_prehistory_group_id_4f26c164` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝用户组预付历史记录';

-- ----------------------------
-- Table structure for dvadmin_alipay_sub_product
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_sub_product`;
CREATE TABLE `dvadmin_alipay_sub_product` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `external_id` varchar(16) NOT NULL,
  `name` varchar(128) DEFAULT NULL COMMENT '进件的二级商户名称',
  `alias_name` varchar(512) DEFAULT NULL COMMENT '商户别名',
  `merchant_type` varchar(2) DEFAULT NULL COMMENT '商家类型',
  `mcc` varchar(10) DEFAULT NULL COMMENT '商户类别码mcc',
  `cert_no` varchar(20) DEFAULT NULL COMMENT '商户证件编号',
  `cert_type` varchar(20) DEFAULT NULL COMMENT '商户证件类型',
  `cert_image` varchar(256) DEFAULT NULL COMMENT '商户证件图片url',
  `cert_image_back` varchar(256) DEFAULT NULL COMMENT '商户证件图片url',
  `legal_name` varchar(64) DEFAULT NULL COMMENT '法人名称',
  `legal_cert_no` varchar(18) DEFAULT NULL COMMENT '法人身份证号',
  `contact_infos` json NOT NULL DEFAULT (_utf8mb4'[]') COMMENT '联系人信息',
  `biz_cards` json NOT NULL DEFAULT (_utf8mb4'[]') COMMENT '结算银行卡',
  `alipay_logon_id` varchar(64) DEFAULT NULL COMMENT '结算支付宝账号',
  `binding_alipay_logon_id` varchar(64) DEFAULT NULL COMMENT '签约支付宝账户',
  `service` json NOT NULL DEFAULT (_utf8mb4'[]') COMMENT '商户使用服务',
  `default_settle_rule` json NOT NULL DEFAULT (_utf8mb4'{}') COMMENT '默认结算规则',
  `business_address` json NOT NULL DEFAULT (_utf8mb4'{}') COMMENT '经营地址',
  `invoice_info` json NOT NULL DEFAULT (_utf8mb4'{}') COMMENT '开票资料信息',
  `out_door_images` json NOT NULL DEFAULT (_utf8mb4'[]') COMMENT '门头照',
  `in_door_images` json NOT NULL DEFAULT (_utf8mb4'[]') COMMENT '内景照',
  `legal_cert_back_image` varchar(256) DEFAULT NULL COMMENT '法人身份证反面图',
  `legal_cert_front_image` varchar(256) DEFAULT NULL COMMENT '法人身份证正面图',
  `license_auth_letter_image` varchar(256) DEFAULT NULL COMMENT '授权函',
  `sites` json NOT NULL DEFAULT (_utf8mb4'[]') COMMENT '商户站点信息',
  `qualifications` json NOT NULL DEFAULT (_utf8mb4'[]') COMMENT '商户行业资质',
  `smid` varchar(32) DEFAULT NULL COMMENT '二级商户id',
  `service_phone` varchar(20) DEFAULT NULL COMMENT '客服电话',
  `sign_time_with_isv` varchar(20) DEFAULT NULL COMMENT '二级商户与服务商的签约时间',
  `cert_name` varchar(64) DEFAULT NULL COMMENT '个体工商户营业执照',
  `legal_cert_type` varchar(8) DEFAULT NULL COMMENT '身份证类型',
  `merchant_nature` varchar(32) DEFAULT NULL COMMENT '商家性质',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `indirect_id` bigint NOT NULL COMMENT '直付通父级',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`external_id`),
  KEY `dvadmin_alipay_sub_product_creator_id_13ae50ac` (`creator_id`),
  KEY `dvadmin_alipay_sub_product_indirect_id_2550dfc0` (`indirect_id`),
  KEY `dvadmin_alipay_sub_product_writeoff_id_a5a6a0e6` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝直付通二级商户';

-- ----------------------------
-- Table structure for dvadmin_alipay_sub_product_request_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_sub_product_request_history`;
CREATE TABLE `dvadmin_alipay_sub_product_request_history` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `order_id` varchar(64) NOT NULL COMMENT '申请单id',
  `status` varchar(3) NOT NULL DEFAULT '031' COMMENT '状态',
  `sub_confirm` varchar(10) DEFAULT NULL COMMENT '二级商户确认状态',
  `fk_audit` varchar(10) DEFAULT NULL COMMENT '风控审核状态',
  `fk_audit_memo` varchar(64) DEFAULT NULL COMMENT '风控审批备注',
  `kz_audit` varchar(10) DEFAULT NULL COMMENT '客资审核状态',
  `kz_audit_memo` varchar(64) DEFAULT NULL COMMENT '客资审批备注',
  `card_alias_no` varchar(32) DEFAULT NULL COMMENT '卡编号',
  `smid` varchar(32) DEFAULT NULL COMMENT '二级商户id',
  `app_pre_auth` varchar(5) DEFAULT NULL COMMENT '线上预授权',
  `face_pre_auth` varchar(5) DEFAULT NULL COMMENT '线下预授权',
  `is_face_limit` varchar(5) DEFAULT NULL COMMENT '权限版本',
  `reason` varchar(256) DEFAULT NULL COMMENT '失败理由',
  `sub_sign_qr_code_url` varchar(256) DEFAULT NULL COMMENT '二维码链接',
  `sub_sign_short_chain_url` varchar(256) DEFAULT NULL COMMENT '短链接',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `sub_merchant_id` varchar(16) NOT NULL COMMENT '直付通二级商户',
  PRIMARY KEY (`id`),
  KEY `dvadmin_alipay_sub_product_request_history_creator_id_bfccc6b9` (`creator_id`),
  KEY `dvadmin_alipay_sub_product__sub_merchant_id_0a5be2cf` (`sub_merchant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝直付通二级商户进件请求';

-- ----------------------------
-- Table structure for dvadmin_alipay_transfer_user
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_transfer_user`;
CREATE TABLE `dvadmin_alipay_transfer_user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `username_type` int NOT NULL COMMENT '用户名类型',
  `username` varchar(255) NOT NULL COMMENT '账号',
  `name` varchar(255) DEFAULT NULL COMMENT '姓名',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `limit_money` bigint NOT NULL DEFAULT '0' COMMENT '总限额',
  `alipay_product_id` bigint DEFAULT NULL COMMENT '关联项目',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_alipay_transfer_user_alipay_product_id_2ef4546b` (`alipay_product_id`),
  KEY `dvadmin_alipay_transfer_user_creator_id_33cc11e4` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝转账用户';

-- ----------------------------
-- Table structure for dvadmin_alipay_transfer_user_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_transfer_user_flow`;
CREATE TABLE `dvadmin_alipay_transfer_user_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `flow` bigint NOT NULL COMMENT '流水',
  `alipay_product_id` bigint DEFAULT NULL COMMENT '关联项目',
  `alipay_user_id` bigint DEFAULT NULL COMMENT '关联接收人',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_alipay_transfer_user_flow_alipay_product_id_ecdae317` (`alipay_product_id`),
  KEY `dvadmin_alipay_transfer_user_flow_alipay_user_id_4e15ebf9` (`alipay_user_id`),
  KEY `dvadmin_alipay_transfer_user_flow_creator_id_f27c6970` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝转账用户流水';

-- ----------------------------
-- Table structure for dvadmin_alipay_weight
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_alipay_weight`;
CREATE TABLE `dvadmin_alipay_weight` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `weight` int NOT NULL DEFAULT '1' COMMENT '权重',
  `alipay_id` bigint NOT NULL COMMENT '父级',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `pay_channel_id` bigint NOT NULL COMMENT '支付通道',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_alipay_weight_alipay_id_pay_channel_id_8a499484_uniq` (`alipay_id`,`pay_channel_id`),
  KEY `dvadmin_alipay_weight_alipay_id_a5c7d4b0` (`alipay_id`),
  KEY `dvadmin_alipay_weight_creator_id_a048d9d3` (`creator_id`),
  KEY `dvadmin_alipay_weight_pay_channel_id_64238ad9` (`pay_channel_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='支付宝权重';

-- ----------------------------
-- Table structure for dvadmin_api_white_list
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_api_white_list`;
CREATE TABLE `dvadmin_api_white_list` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `url` varchar(200) NOT NULL COMMENT 'url地址',
  `method` int DEFAULT NULL COMMENT '接口请求方法',
  `enable_datasource` tinyint(1) NOT NULL COMMENT '激活数据权限',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_api_white_list_creator_id_fd335789` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='接口白名单';

-- ----------------------------
-- Table structure for dvadmin_ban_ip
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_ban_ip`;
CREATE TABLE `dvadmin_ban_ip` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `ip_address` varchar(255) NOT NULL COMMENT 'Ip地址',
  `from_complain` tinyint(1) NOT NULL DEFAULT '0',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `tenant_id` bigint DEFAULT NULL COMMENT '租户',
  PRIMARY KEY (`id`),
  KEY `dvadmin_ban_ip_creator_id_475e4be5` (`creator_id`),
  KEY `dvadmin_ban_ip_tenant_id_0a4ddc9d` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='封禁Ip';

-- ----------------------------
-- Table structure for dvadmin_ban_user_id
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_ban_user_id`;
CREATE TABLE `dvadmin_ban_user_id` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `user_id` varchar(32) NOT NULL COMMENT '用户id',
  `from_complain` tinyint(1) NOT NULL DEFAULT '0',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `tenant_id` bigint NOT NULL COMMENT '租户',
  PRIMARY KEY (`id`),
  KEY `dvadmin_ban_user_id_creator_id_110c96b1` (`creator_id`),
  KEY `dvadmin_ban_user_id_tenant_id_65ac1899` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='封禁UserId';

-- ----------------------------
-- Table structure for dvadmin_card_key
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_card_key`;
CREATE TABLE `dvadmin_card_key` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `money` bigint NOT NULL DEFAULT '0' COMMENT '计划金额',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `recharge_type` int NOT NULL COMMENT '充值类型',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_card_key_creator_id_4498ce15` (`creator_id`),
  KEY `dvadmin_card_key_writeoff_id_80ae2ff7` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='卡密';

-- ----------------------------
-- Table structure for dvadmin_card_key_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_card_key_flow`;
CREATE TABLE `dvadmin_card_key_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `flow` bigint NOT NULL DEFAULT '0' COMMENT '已充值金额',
  `card_id` bigint DEFAULT NULL COMMENT '关联卡密',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_card_key_flow_card_id_682248f6` (`card_id`),
  KEY `dvadmin_card_key_flow_creator_id_1a7208a8` (`creator_id`),
  KEY `dvadmin_card_key_flow_writeoff_id_c5a3efff` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='卡密流水';

-- ----------------------------
-- Table structure for dvadmin_card_key_info
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_card_key_info`;
CREATE TABLE `dvadmin_card_key_info` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `money` bigint NOT NULL DEFAULT '0' COMMENT '金额',
  `card_no` longtext NOT NULL COMMENT '卡号',
  `card_pwd` longtext NOT NULL COMMENT '卡密码',
  `order_no` varchar(255) NOT NULL COMMENT '订单号',
  `card_id` bigint NOT NULL COMMENT '关联卡密',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `order_id` varchar(30) DEFAULT NULL COMMENT '订单',
  PRIMARY KEY (`id`),
  KEY `dvadmin_card_key_info_card_id_9170f68e` (`card_id`),
  KEY `dvadmin_card_key_info_creator_id_3d92d767` (`creator_id`),
  KEY `dvadmin_card_key_info_order_id_a77be97c` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='卡密详细';

-- ----------------------------
-- Table structure for dvadmin_collection_day_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_collection_day_flow`;
CREATE TABLE `dvadmin_collection_day_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `flow` bigint NOT NULL DEFAULT '0' COMMENT '已充值金额',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `user_id` bigint NOT NULL COMMENT '关联归集用户',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_collection_day_flow_date_user_id_86aae3d5_uniq` (`date`,`user_id`),
  KEY `dvadmin_collection_day_flow_creator_id_1adf9c5c` (`creator_id`),
  KEY `dvadmin_collection_day_flow_user_id_e792931e` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='归集日统计';

-- ----------------------------
-- Table structure for dvadmin_collection_user
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_collection_user`;
CREATE TABLE `dvadmin_collection_user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL COMMENT '账号',
  `name` varchar(255) DEFAULT NULL COMMENT '姓名',
  `remarks` longtext COMMENT '备注',
  `tenant_id` bigint DEFAULT NULL COMMENT '关联租户',
  PRIMARY KEY (`id`),
  KEY `dvadmin_collection_user_tenant_id_e9188825` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='归集用户';

-- ----------------------------
-- Table structure for dvadmin_common_recharge_account
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_common_recharge_account`;
CREATE TABLE `dvadmin_common_recharge_account` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `shop_type` int NOT NULL DEFAULT '0' COMMENT '类型',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `account` varchar(255) NOT NULL COMMENT '账号',
  `limit_money` bigint NOT NULL DEFAULT '0' COMMENT '日限额',
  `cookie` longtext COMMENT 'ck',
  `warning_msg` longtext COMMENT '风控说明',
  `password` varchar(32) DEFAULT NULL COMMENT '密码',
  `extra` json NOT NULL COMMENT '额外信息',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_common_recharge_account_creator_id_476c753d` (`creator_id`),
  KEY `dvadmin_common_recharge_account_writeoff_id_2b88cfdb` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='综合充值账号';

-- ----------------------------
-- Table structure for dvadmin_common_recharge_account_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_common_recharge_account_day`;
CREATE TABLE `dvadmin_common_recharge_account_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `account_id` bigint DEFAULT NULL COMMENT '关联综合充值',
  PRIMARY KEY (`id`),
  KEY `dvadmin_common_recharge_account_day_account_id_a47da9d7` (`account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='综合充值日统计';

-- ----------------------------
-- Table structure for dvadmin_day_statistics
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_day_statistics`;
CREATE TABLE `dvadmin_day_statistics` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `ver` bigint NOT NULL,
  `unknown_count` int NOT NULL DEFAULT '0' COMMENT '未知设备订单数',
  `android_count` int NOT NULL DEFAULT '0' COMMENT '安卓订单数',
  `ios_count` int NOT NULL DEFAULT '0' COMMENT '苹果订单数',
  `pc_count` int NOT NULL DEFAULT '0' COMMENT '电脑(web)订单数',
  `total_tax` bigint NOT NULL DEFAULT '0' COMMENT '总利润',
  `submit_money` bigint NOT NULL COMMENT '总提交收入',
  `date` date NOT NULL COMMENT '日期',
  PRIMARY KEY (`id`),
  UNIQUE KEY `date` (`date`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb3 COMMENT='日统计';

-- ----------------------------
-- Table structure for dvadmin_day_statistics_channel_writeoff
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_day_statistics_channel_writeoff`;
CREATE TABLE `dvadmin_day_statistics_channel_writeoff` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `unknown_count` int NOT NULL DEFAULT '0' COMMENT '未知设备订单数',
  `android_count` int NOT NULL DEFAULT '0' COMMENT '安卓订单数',
  `ios_count` int NOT NULL DEFAULT '0' COMMENT '苹果订单数',
  `pc_count` int NOT NULL DEFAULT '0' COMMENT '电脑(web)订单数',
  `total_tax` bigint NOT NULL DEFAULT '0' COMMENT '总利润',
  `pay_channel_id` bigint DEFAULT NULL COMMENT '关联支付通道',
  `writeoff_id` bigint DEFAULT NULL COMMENT '核销',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_day_statistics_c_date_writeoff_id_pay_cha_ad24f854_uniq` (`date`,`writeoff_id`,`pay_channel_id`),
  KEY `dvadmin_day_statistics_channel_writeoff_pay_channel_id_f1a35e3d` (`pay_channel_id`),
  KEY `dvadmin_day_statistics_channel_writeoff_writeoff_id_34479be6` (`writeoff_id`)
) ENGINE=InnoDB AUTO_INCREMENT=27 DEFAULT CHARSET=utf8mb3 COMMENT='核销通道日统计';

-- ----------------------------
-- Table structure for dvadmin_day_statistics_merchant
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_day_statistics_merchant`;
CREATE TABLE `dvadmin_day_statistics_merchant` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `unknown_count` int NOT NULL DEFAULT '0' COMMENT '未知设备订单数',
  `android_count` int NOT NULL DEFAULT '0' COMMENT '安卓订单数',
  `ios_count` int NOT NULL DEFAULT '0' COMMENT '苹果订单数',
  `pc_count` int NOT NULL DEFAULT '0' COMMENT '电脑(web)订单数',
  `total_tax` bigint NOT NULL DEFAULT '0' COMMENT '总利润',
  `real_money` bigint NOT NULL DEFAULT '0' COMMENT '实际收入',
  `merchant_id` bigint DEFAULT NULL COMMENT '商户',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_day_statistics_merchant_date_merchant_id_1421cd92_uniq` (`date`,`merchant_id`),
  KEY `dvadmin_day_statistics_merchant_merchant_id_3706e55d` (`merchant_id`)
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb3 COMMENT='商户日统计';

-- ----------------------------
-- Table structure for dvadmin_day_statistics_pay_channel
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_day_statistics_pay_channel`;
CREATE TABLE `dvadmin_day_statistics_pay_channel` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `unknown_count` int NOT NULL DEFAULT '0' COMMENT '未知设备订单数',
  `android_count` int NOT NULL DEFAULT '0' COMMENT '安卓订单数',
  `ios_count` int NOT NULL DEFAULT '0' COMMENT '苹果订单数',
  `pc_count` int NOT NULL DEFAULT '0' COMMENT '电脑(web)订单数',
  `total_tax` bigint NOT NULL DEFAULT '0' COMMENT '总利润',
  `real_money` bigint NOT NULL DEFAULT '0' COMMENT '实际收入',
  `merchant_id` bigint DEFAULT NULL COMMENT '商户',
  `pay_channel_id` bigint DEFAULT NULL COMMENT '关联支付通道',
  `tenant_id` bigint DEFAULT NULL COMMENT '租户',
  `writeoff_id` bigint DEFAULT NULL COMMENT '核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_day_statistics_pay_channel_merchant_id_e55f6742` (`merchant_id`),
  KEY `dvadmin_day_statistics_pay_channel_pay_channel_id_1c30c852` (`pay_channel_id`),
  KEY `dvadmin_day_statistics_pay_channel_tenant_id_e5ff8f7e` (`tenant_id`),
  KEY `dvadmin_day_statistics_pay_channel_writeoff_id_9d314a11` (`writeoff_id`)
) ENGINE=InnoDB AUTO_INCREMENT=27 DEFAULT CHARSET=utf8mb3 COMMENT='支付通道日统计';

-- ----------------------------
-- Table structure for dvadmin_day_statistics_tenant
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_day_statistics_tenant`;
CREATE TABLE `dvadmin_day_statistics_tenant` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `unknown_count` int NOT NULL DEFAULT '0' COMMENT '未知设备订单数',
  `android_count` int NOT NULL DEFAULT '0' COMMENT '安卓订单数',
  `ios_count` int NOT NULL DEFAULT '0' COMMENT '苹果订单数',
  `pc_count` int NOT NULL DEFAULT '0' COMMENT '电脑(web)订单数',
  `total_tax` bigint NOT NULL DEFAULT '0' COMMENT '总利润',
  `tenant_id` bigint DEFAULT NULL COMMENT '租户',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_day_statistics_tenant_date_tenant_id_8b7e059d_uniq` (`date`,`tenant_id`),
  KEY `dvadmin_day_statistics_tenant_tenant_id_aa8794af` (`tenant_id`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb3 COMMENT='租户日统计';

-- ----------------------------
-- Table structure for dvadmin_day_statistics_writeoff
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_day_statistics_writeoff`;
CREATE TABLE `dvadmin_day_statistics_writeoff` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `unknown_count` int NOT NULL DEFAULT '0' COMMENT '未知设备订单数',
  `android_count` int NOT NULL DEFAULT '0' COMMENT '安卓订单数',
  `ios_count` int NOT NULL DEFAULT '0' COMMENT '苹果订单数',
  `pc_count` int NOT NULL DEFAULT '0' COMMENT '电脑(web)订单数',
  `total_tax` bigint NOT NULL DEFAULT '0' COMMENT '总利润',
  `submit_money` bigint NOT NULL COMMENT '总提交收入',
  `writeoff_id` bigint DEFAULT NULL COMMENT '核销',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_day_statistics_writeoff_date_writeoff_id_fe774b84_uniq` (`date`,`writeoff_id`),
  KEY `dvadmin_day_statistics_writeoff_writeoff_id_fbd314ff` (`writeoff_id`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb3 COMMENT='核销日统计';

-- ----------------------------
-- Table structure for dvadmin_ding_user
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_ding_user`;
CREATE TABLE `dvadmin_ding_user` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(11) NOT NULL COMMENT '手机号',
  `corp_id` varchar(255) NOT NULL COMMENT 'corpId',
  `process_code` varchar(255) DEFAULT NULL COMMENT '审批code',
  `process_name` varchar(255) DEFAULT NULL COMMENT '审批名称',
  `url` longtext NOT NULL COMMENT 'url',
  `online` tinyint(1) NOT NULL DEFAULT '0' COMMENT '在线',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `limit_money` bigint NOT NULL DEFAULT '0' COMMENT '限制金额',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_ding_user_creator_id_acba0c33` (`creator_id`),
  KEY `dvadmin_ding_user_writeoff_id_5227870c` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='钉钉用户';

-- ----------------------------
-- Table structure for dvadmin_ding_user_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_ding_user_day`;
CREATE TABLE `dvadmin_ding_user_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `user_id` varchar(11) DEFAULT NULL COMMENT '关联',
  PRIMARY KEY (`id`),
  KEY `dvadmin_ding_user_day_user_id_6ffc96cd` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='钉钉用户日统计';

-- ----------------------------
-- Table structure for dvadmin_ding_user_groups
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_ding_user_groups`;
CREATE TABLE `dvadmin_ding_user_groups` (
  `id` int NOT NULL AUTO_INCREMENT,
  `dinguser_id` varchar(11) NOT NULL,
  `alipaysplitusergroup_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_ding_user_groups_dinguser_id_alipaysplitu_e2f57510_uniq` (`dinguser_id`,`alipaysplitusergroup_id`),
  KEY `dvadmin_ding_user_groups_dinguser_id_96eb280e` (`dinguser_id`),
  KEY `dvadmin_ding_user_groups_alipaysplitusergroup_id_3bf6d85d` (`alipaysplitusergroup_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_ding_user_split_mapping
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_ding_user_split_mapping`;
CREATE TABLE `dvadmin_ding_user_split_mapping` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(32) NOT NULL COMMENT '绑定id',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `split_user_id` bigint NOT NULL COMMENT '关联用户',
  `user_id` varchar(11) NOT NULL COMMENT '关联用户',
  PRIMARY KEY (`id`),
  KEY `dvadmin_ding_user_split_mapping_creator_id_68b7e945` (`creator_id`),
  KEY `dvadmin_ding_user_split_mapping_split_user_id_27cc8455` (`split_user_id`),
  KEY `dvadmin_ding_user_split_mapping_user_id_953a6405` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='钉钉绑定中间表';

-- ----------------------------
-- Table structure for dvadmin_douyin_hongbao
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_douyin_hongbao`;
CREATE TABLE `dvadmin_douyin_hongbao` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `id` varchar(255) NOT NULL COMMENT '红包ID',
  `money` int NOT NULL COMMENT '金额',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `account_id` varchar(16) DEFAULT NULL COMMENT '关联dy账号',
  PRIMARY KEY (`id`),
  KEY `dvadmin_douyin_hongbao_create_datetime_f1354f6c` (`create_datetime`),
  KEY `dvadmin_douyin_hongbao_creator_id_c46302ce` (`creator_id`),
  KEY `dvadmin_douyin_hongbao_account_id_cf97be34` (`account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='国内抖音红包';

-- ----------------------------
-- Table structure for dvadmin_douyin_hongbao_account
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_douyin_hongbao_account`;
CREATE TABLE `dvadmin_douyin_hongbao_account` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `limit_money` bigint NOT NULL DEFAULT '0' COMMENT '日限额',
  `max_money` int NOT NULL DEFAULT '0' COMMENT '单笔最大',
  `min_money` int NOT NULL DEFAULT '0' COMMENT '单笔最小',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `online` tinyint(1) NOT NULL DEFAULT '0' COMMENT '登录状态',
  `uid` varchar(16) NOT NULL COMMENT 'uid',
  `id` varchar(16) NOT NULL COMMENT '用户ID',
  `nickname` varchar(255) NOT NULL COMMENT '昵称',
  `phone` varchar(11) NOT NULL COMMENT '手机号',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_douyin_hongbao_account_creator_id_229347ff` (`creator_id`),
  KEY `dvadmin_douyin_hongbao_account_writeoff_id_ac8e12bd` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='国内抖音红包账号';

-- ----------------------------
-- Table structure for dvadmin_douyin_hongbao_account_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_douyin_hongbao_account_day`;
CREATE TABLE `dvadmin_douyin_hongbao_account_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `account_id` varchar(16) DEFAULT NULL COMMENT '关联dy账号',
  PRIMARY KEY (`id`),
  KEY `dvadmin_douyin_hongbao_account_day_account_id_6db76c3f` (`account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='国内抖音账号红包日统计';

-- ----------------------------
-- Table structure for dvadmin_douyin_hongbao_message
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_douyin_hongbao_message`;
CREATE TABLE `dvadmin_douyin_hongbao_message` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(255) NOT NULL COMMENT '消息id',
  `content` varchar(255) DEFAULT NULL COMMENT '内容',
  `sender` varchar(255) DEFAULT NULL COMMENT '发送人名称',
  `is_hongbao` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否红包',
  `money` int NOT NULL COMMENT '金额',
  `ver` bigint NOT NULL,
  `account_id` varchar(16) DEFAULT NULL COMMENT '关联dy账号',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `order_id` varchar(30) DEFAULT NULL COMMENT '关联订单',
  PRIMARY KEY (`id`),
  KEY `dvadmin_douyin_hongbao_message_sender_92ef4689` (`sender`),
  KEY `dvadmin_douyin_hongbao_message_account_id_32be8fb7` (`account_id`),
  KEY `dvadmin_douyin_hongbao_message_creator_id_7dfe9e2e` (`creator_id`),
  KEY `dvadmin_douyin_hongbao_message_order_id_10961cbc` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='国内抖音账号消息';

-- ----------------------------
-- Table structure for dvadmin_etc_account
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_etc_account`;
CREATE TABLE `dvadmin_etc_account` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `username` varchar(255) NOT NULL COMMENT '账号',
  `password` varchar(255) NOT NULL COMMENT '密码',
  `cookie` json NOT NULL COMMENT 'cookie',
  `online` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否在线',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  KEY `dvadmin_etc_account_creator_id_abf6d698` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='Etc用户';

-- ----------------------------
-- Table structure for dvadmin_etc_card
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_etc_card`;
CREATE TABLE `dvadmin_etc_card` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `card_no` varchar(255) NOT NULL COMMENT '卡号',
  `card_id` varchar(255) NOT NULL COMMENT '卡ID',
  `car_no` varchar(255) NOT NULL COMMENT '车牌号',
  `limit` bigint NOT NULL DEFAULT '0' COMMENT '限制金额',
  `balance` bigint NOT NULL DEFAULT '0' COMMENT '初始余额',
  `current_balance` bigint NOT NULL DEFAULT '0' COMMENT '实时余额',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `recharge_status` varchar(255) DEFAULT NULL COMMENT '充值状态',
  `ver` bigint NOT NULL,
  `account_id` bigint NOT NULL COMMENT '关联账号',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_etc_card_account_id_d105a778` (`account_id`),
  KEY `dvadmin_etc_card_creator_id_dcb94109` (`creator_id`),
  KEY `dvadmin_etc_card_writeoff_id_7892ce92` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='EtcCard';

-- ----------------------------
-- Table structure for dvadmin_etc_card_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_etc_card_flow`;
CREATE TABLE `dvadmin_etc_card_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `flow` bigint NOT NULL DEFAULT '0' COMMENT '已充值金额',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `etc_card_id` bigint DEFAULT NULL COMMENT '关联etc卡号',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_etc_card_flow_creator_id_f02b8a88` (`creator_id`),
  KEY `dvadmin_etc_card_flow_etc_card_id_9884d61f` (`etc_card_id`),
  KEY `dvadmin_etc_card_flow_writeoff_id_4284b379` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='etc流水';

-- ----------------------------
-- Table structure for dvadmin_google_auth
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_google_auth`;
CREATE TABLE `dvadmin_google_auth` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `status` tinyint(1) NOT NULL COMMENT '状态',
  `token` varchar(255) NOT NULL COMMENT '秘钥',
  `user_id` bigint DEFAULT NULL COMMENT '绑定的系统用户',
  PRIMARY KEY (`id`),
  UNIQUE KEY `token` (`token`),
  UNIQUE KEY `user_id` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb3 COMMENT='谷歌验证';

-- ----------------------------
-- Table structure for dvadmin_house_notification_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_house_notification_history`;
CREATE TABLE `dvadmin_house_notification_history` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `url` longtext NOT NULL COMMENT '通知地址',
  `request_body` longtext COMMENT '请求参数',
  `request_method` varchar(8) DEFAULT NULL COMMENT '请求方式',
  `response_code` int NOT NULL COMMENT '响应状态码',
  `json_result` longtext COMMENT '返回信息',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `house_id` varchar(21) NOT NULL COMMENT '关联通知',
  PRIMARY KEY (`id`),
  KEY `dvadmin_house_notification_history_creator_id_5b6f04ae` (`creator_id`),
  KEY `dvadmin_house_notification_history_house_id_47db9744` (`house_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='户号通知记录';

-- ----------------------------
-- Table structure for dvadmin_house_product
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_house_product`;
CREATE TABLE `dvadmin_house_product` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(21) NOT NULL COMMENT '户号订单号',
  `house_card` varchar(255) NOT NULL COMMENT '户号',
  `province` varchar(100) DEFAULT NULL COMMENT '省份',
  `city` varchar(100) DEFAULT NULL COMMENT '城市',
  `house_type` int NOT NULL COMMENT '类型',
  `init_balance` bigint NOT NULL DEFAULT '0' COMMENT '初始余额',
  `house_order_no` varchar(255) NOT NULL COMMENT '供货商订单号',
  `min_pay` int DEFAULT NULL COMMENT '最低充值金额',
  `money` int NOT NULL COMMENT '计划金额',
  `notify_url` varchar(255) DEFAULT NULL COMMENT '回调地址',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态',
  `is_active` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否存在',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  UNIQUE KEY `house_order_no` (`house_order_no`),
  KEY `dvadmin_house_product_creator_id_28a3ba2e` (`creator_id`),
  KEY `dvadmin_house_product_writeoff_id_167b574e` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='户号库存';

-- ----------------------------
-- Table structure for dvadmin_house_product_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_house_product_flow`;
CREATE TABLE `dvadmin_house_product_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `flow` bigint NOT NULL DEFAULT '0' COMMENT '已充值金额',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `house_id` varchar(21) DEFAULT NULL COMMENT '关联户号',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_house_product_flow_creator_id_2e5fb3d3` (`creator_id`),
  KEY `dvadmin_house_product_flow_house_id_bb5bc0c4` (`house_id`),
  KEY `dvadmin_house_product_flow_writeoff_id_53b02d3f` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='户号流水';

-- ----------------------------
-- Table structure for dvadmin_jch_user
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_jch_user`;
CREATE TABLE `dvadmin_jch_user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `username` varchar(64) NOT NULL,
  `password` varchar(64) NOT NULL,
  `token` longtext NOT NULL,
  `client_id` varchar(32) NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  KEY `dvadmin_jch_user_creator_id_a15f6deb` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='鲸车惠用户';

-- ----------------------------
-- Table structure for dvadmin_jd_address
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_jd_address`;
CREATE TABLE `dvadmin_jd_address` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `province` varchar(16) NOT NULL COMMENT '省',
  `city` varchar(32) NOT NULL COMMENT '城市',
  `county` varchar(32) NOT NULL COMMENT '区',
  `town` varchar(32) NOT NULL COMMENT '街道',
  `detail` longtext NOT NULL COMMENT '详细地址',
  `name` varchar(16) NOT NULL COMMENT '姓名',
  `phone` varchar(16) NOT NULL COMMENT '手机号',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `tenant_id` bigint NOT NULL COMMENT '关联租户',
  PRIMARY KEY (`id`),
  KEY `dvadmin_jd_address_creator_id_0fa708a3` (`creator_id`),
  KEY `dvadmin_jd_address_tenant_id_681e9712` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='京东地址';

-- ----------------------------
-- Table structure for dvadmin_jd_game
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_jd_game`;
CREATE TABLE `dvadmin_jd_game` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `shop_type` int NOT NULL DEFAULT '0' COMMENT '类型',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `mall_name` varchar(255) NOT NULL COMMENT '店铺标题',
  `brand_id` varchar(255) NOT NULL COMMENT 'brand_id',
  `limit_money` bigint NOT NULL DEFAULT '0' COMMENT '日限额',
  `money` int NOT NULL DEFAULT '0' COMMENT '拉单金额',
  `real_money` int NOT NULL DEFAULT '0' COMMENT '商品金额',
  `sku_id` varchar(255) NOT NULL COMMENT 'sku_id',
  `item_pic` longtext COMMENT '商品图片',
  `cookie` longtext COMMENT '店铺cookie',
  `really` tinyint(1) NOT NULL DEFAULT '0' COMMENT '真充',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_jd_game_creator_id_90ad2d53` (`creator_id`),
  KEY `dvadmin_jd_game_writeoff_id_8a1ddfc5` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='京东游戏';

-- ----------------------------
-- Table structure for dvadmin_jd_game_account
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_jd_game_account`;
CREATE TABLE `dvadmin_jd_game_account` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `account` varchar(255) NOT NULL COMMENT '账号',
  `limit_money` varchar(255) NOT NULL COMMENT '日限额',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否启用',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `game_id` bigint DEFAULT NULL COMMENT '关联京东游戏',
  PRIMARY KEY (`id`),
  KEY `dvadmin_jd_game_account_creator_id_29d85aee` (`creator_id`),
  KEY `dvadmin_jd_game_account_game_id_923a554c` (`game_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='京东游戏充值账号';

-- ----------------------------
-- Table structure for dvadmin_jd_game_account_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_jd_game_account_day`;
CREATE TABLE `dvadmin_jd_game_account_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `account_id` bigint DEFAULT NULL COMMENT '关联充值账号',
  PRIMARY KEY (`id`),
  KEY `dvadmin_jd_game_account_day_account_id_4d41ed0a` (`account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='京东游戏充值账号日统计';

-- ----------------------------
-- Table structure for dvadmin_jd_game_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_jd_game_day`;
CREATE TABLE `dvadmin_jd_game_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `game_id` bigint DEFAULT NULL COMMENT '关联京东游戏',
  PRIMARY KEY (`id`),
  KEY `dvadmin_jd_game_day_game_id_ac3303c4` (`game_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='京东游戏日统计';

-- ----------------------------
-- Table structure for dvadmin_jd_game_pre_order
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_jd_game_pre_order`;
CREATE TABLE `dvadmin_jd_game_pre_order` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态,-1过期,0等待,1支付中,2支付成功',
  `money` int NOT NULL DEFAULT '0' COMMENT '拉单金额',
  `order_id` varchar(128) DEFAULT NULL COMMENT '系统订单号',
  `out_order_no` varchar(128) DEFAULT NULL COMMENT '商户订单号',
  `pay_url` longtext COMMENT '支付链接',
  `jd_order_id` varchar(128) NOT NULL COMMENT '京东订单号',
  `expire_datetime` datetime(6) DEFAULT NULL COMMENT '过期时间',
  `extra` json NOT NULL COMMENT '额外数据',
  `sku_id` varchar(255) NOT NULL COMMENT 'sku_id',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  `task_id` bigint NOT NULL COMMENT '关联任务',
  PRIMARY KEY (`id`),
  KEY `dvadmin_jd_game_pre_order_creator_id_39eaaec3` (`creator_id`),
  KEY `dvadmin_jd_game_pre_order_writeoff_id_7ad0e296` (`writeoff_id`),
  KEY `dvadmin_jd_game_pre_order_task_id_0d7cf7d4` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='京东预拉单';

-- ----------------------------
-- Table structure for dvadmin_jd_game_pre_order_task
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_jd_game_pre_order_task`;
CREATE TABLE `dvadmin_jd_game_pre_order_task` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `product_id` bigint NOT NULL COMMENT '关联京东游戏',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态,0进行中,1暂停,2完成',
  `expire` int NOT NULL DEFAULT '600' COMMENT '过期秒数',
  `max_count` int NOT NULL DEFAULT '1' COMMENT '最大数量',
  `forever` tinyint(1) NOT NULL DEFAULT '0' COMMENT '一直运行',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`product_id`),
  KEY `dvadmin_jd_game_pre_order_task_creator_id_9a65446d` (`creator_id`),
  KEY `dvadmin_jd_game_pre_order_task_writeoff_id_c5b21a52` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='京东预拉单任务';

-- ----------------------------
-- Table structure for dvadmin_jiweipay_user
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_jiweipay_user`;
CREATE TABLE `dvadmin_jiweipay_user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `mch` varchar(255) NOT NULL COMMENT '商户号',
  `app_id` varchar(255) NOT NULL COMMENT 'app_id',
  `key` varchar(255) NOT NULL COMMENT '密钥',
  `alipay_status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '支付宝状态',
  `wechat_status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '微信状态',
  `limit_money` bigint NOT NULL DEFAULT '0' COMMENT '日限额',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_jiweipay_user_creator_id_d9a7808f` (`creator_id`),
  KEY `dvadmin_jiweipay_user_writeoff_id_9afe2d87` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='计为支付';

-- ----------------------------
-- Table structure for dvadmin_jiweipay_user_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_jiweipay_user_day`;
CREATE TABLE `dvadmin_jiweipay_user_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `user_id` bigint DEFAULT NULL COMMENT '关联计为用户',
  PRIMARY KEY (`id`),
  KEY `dvadmin_jiweipay_user_day_user_id_d5f8597f` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='计为支付日统计';

-- ----------------------------
-- Table structure for dvadmin_jt_backend_account
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_jt_backend_account`;
CREATE TABLE `dvadmin_jt_backend_account` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `token` varchar(100) DEFAULT NULL,
  `admin_token` varchar(100) DEFAULT NULL,
  `id` varchar(32) NOT NULL COMMENT '后台管理账号',
  `password` varchar(32) NOT NULL COMMENT '后台管理密码',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '账户启用状态',
  `is_admin` tinyint(1) NOT NULL DEFAULT '0',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_jt_backend_account_creator_id_6cb2a3b4` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='JT后台账户管理';

-- ----------------------------
-- Table structure for dvadmin_jt_backend_account_products
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_jt_backend_account_products`;
CREATE TABLE `dvadmin_jt_backend_account_products` (
  `id` int NOT NULL AUTO_INCREMENT,
  `jtbackendaccount_id` varchar(32) NOT NULL,
  `jtproduct_id` varchar(48) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_jt_backend_accou_jtbackendaccount_id_jtpr_b62c2e40_uniq` (`jtbackendaccount_id`,`jtproduct_id`),
  KEY `dvadmin_jt_backend_account_products_jtbackendaccount_id_887730e0` (`jtbackendaccount_id`),
  KEY `dvadmin_jt_backend_account_products_jtproduct_id_addf015e` (`jtproduct_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_jt_product
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_jt_product`;
CREATE TABLE `dvadmin_jt_product` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(48) NOT NULL,
  `name` varchar(255) NOT NULL COMMENT '商品名称',
  `url` varchar(500) NOT NULL COMMENT '商品访问链接',
  `product_amount` bigint NOT NULL COMMENT '实际商品金额，单位：分',
  `amount` bigint NOT NULL COMMENT '商品金额，单位：分',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '商品启用状态',
  `ver` bigint NOT NULL,
  `admin_id` varchar(32) NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dvadmin_jt_product_admin_id_61790caf` (`admin_id`),
  KEY `dvadmin_jt_product_creator_id_ab004521` (`creator_id`),
  KEY `dvadmin_jt_product_writeoff_id_34c48a38` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='JT商品管理';

-- ----------------------------
-- Table structure for dvadmin_jt_product_day_statistics
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_jt_product_day_statistics`;
CREATE TABLE `dvadmin_jt_product_day_statistics` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `product_id` varchar(48) NOT NULL COMMENT '关联的JT商品',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_jt_product_day_statistics_date_product_id_9fb93cca_uniq` (`date`,`product_id`),
  KEY `dvadmin_jt_product_day_statistics_product_id_d3727a4b` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='JT商品每日统计';

-- ----------------------------
-- Table structure for dvadmin_merchant
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_merchant`;
CREATE TABLE `dvadmin_merchant` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `telegram` varchar(255) DEFAULT NULL COMMENT 'Telegram群的id',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `system_user_id` bigint DEFAULT NULL COMMENT '绑定的系统用户',
  `parent_id` bigint NOT NULL COMMENT '上级租户',
  PRIMARY KEY (`id`),
  UNIQUE KEY `system_user_id` (`system_user_id`),
  KEY `dvadmin_merchant_parent_id_f0b2817b_fk_dvadmin_tenant_id` (`parent_id`),
  KEY `dvadmin_merchant_creator_id_db35e8ba` (`creator_id`),
  CONSTRAINT `dvadmin_merchant_parent_id_f0b2817b_fk_dvadmin_tenant_id` FOREIGN KEY (`parent_id`) REFERENCES `dvadmin_tenant` (`id`),
  CONSTRAINT `dvadmin_merchant_system_user_id_6ea24d74_fk_dvadmin_s` FOREIGN KEY (`system_user_id`) REFERENCES `dvadmin_system_users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=20006 DEFAULT CHARSET=utf8mb3 COMMENT='商户';

-- ----------------------------
-- Table structure for dvadmin_merchant_notification
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_merchant_notification`;
CREATE TABLE `dvadmin_merchant_notification` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `status` int NOT NULL COMMENT '通知状态',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `order_id` varchar(30) NOT NULL COMMENT '关联订单',
  PRIMARY KEY (`id`),
  UNIQUE KEY `order_id` (`order_id`),
  KEY `dvadmin_merchant_notification_creator_id_80c42fad` (`creator_id`)
) ENGINE=InnoDB AUTO_INCREMENT=184619 DEFAULT CHARSET=utf8mb3 COMMENT='商户通知';

-- ----------------------------
-- Table structure for dvadmin_merchant_notification_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_merchant_notification_history`;
CREATE TABLE `dvadmin_merchant_notification_history` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `url` longtext NOT NULL COMMENT '通知地址',
  `request_body` longtext COMMENT '请求参数',
  `request_method` varchar(8) DEFAULT NULL COMMENT '请求方式',
  `response_code` int NOT NULL COMMENT '响应状态码',
  `json_result` longtext COMMENT '返回信息',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `notification_id` bigint NOT NULL COMMENT '关联通知',
  PRIMARY KEY (`id`),
  KEY `dvadmin_merchant_notification_history_creator_id_e3138ab6` (`creator_id`),
  KEY `dvadmin_merchant_notification_history_notification_id_1d76b34c` (`notification_id`)
) ENGINE=InnoDB AUTO_INCREMENT=196790 DEFAULT CHARSET=utf8mb3 COMMENT='商户通知记录';

-- ----------------------------
-- Table structure for dvadmin_merchant_pay_channel
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_merchant_pay_channel`;
CREATE TABLE `dvadmin_merchant_pay_channel` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `tax` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '费率',
  `limit` int NOT NULL DEFAULT '0' COMMENT '并发限制',
  `status` tinyint(1) NOT NULL COMMENT '通道状态',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `merchant_id` bigint NOT NULL COMMENT '绑定的商户',
  `pay_channel_id` bigint NOT NULL COMMENT '绑定的支付通道',
  PRIMARY KEY (`id`),
  KEY `dvadmin_merchant_pay_channel_creator_id_006d65fb` (`creator_id`),
  KEY `dvadmin_merchant_pay_channel_merchant_id_ed7e642d` (`merchant_id`),
  KEY `dvadmin_merchant_pay_channel_pay_channel_id_832cd758` (`pay_channel_id`)
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb3 COMMENT='商户支付通道';

-- ----------------------------
-- Table structure for dvadmin_merchant_pre
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_merchant_pre`;
CREATE TABLE `dvadmin_merchant_pre` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `pre_pay` bigint NOT NULL DEFAULT '0' COMMENT '预付金额',
  `ver` bigint NOT NULL,
  `merchant_id` bigint NOT NULL COMMENT '关联商户',
  PRIMARY KEY (`id`),
  UNIQUE KEY `merchant_id` (`merchant_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb3 COMMENT='商户预付';

-- ----------------------------
-- Table structure for dvadmin_merchant_pre_add_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_merchant_pre_add_history`;
CREATE TABLE `dvadmin_merchant_pre_add_history` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `merchant_id` bigint NOT NULL DEFAULT '0' COMMENT '商户id',
  `rate` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT 'rate',
  `cny` bigint NOT NULL DEFAULT '0' COMMENT 'cny',
  `usdt` bigint NOT NULL DEFAULT '0' COMMENT 'usdt',
  `to_address` varchar(64) NOT NULL,
  `telegram` varchar(64) DEFAULT NULL,
  `user` varchar(64) DEFAULT NULL,
  `other_user` varchar(64) DEFAULT NULL,
  `tx_id` varchar(64) DEFAULT NULL,
  `ver` bigint NOT NULL,
  `status` tinyint(1) NOT NULL DEFAULT '0',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_merchant_pre_add_history_creator_id_4577a482` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='商户预付打款历史记录';

-- ----------------------------
-- Table structure for dvadmin_merchant_pre_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_merchant_pre_history`;
CREATE TABLE `dvadmin_merchant_pre_history` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `pre_pay` bigint NOT NULL DEFAULT '0' COMMENT '预付金额',
  `before` bigint NOT NULL DEFAULT '0' COMMENT '改动前金额',
  `after` bigint NOT NULL DEFAULT '0' COMMENT '改动后金额',
  `user` varchar(255) DEFAULT NULL,
  `ver` bigint NOT NULL,
  `rate` varchar(32) DEFAULT '0' COMMENT 'usdt汇率',
  `usdt` varchar(32) DEFAULT '0' COMMENT 'usdt',
  `cert` longtext COMMENT '转账凭证',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `merchant_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dvadmin_merchant_pre_history_creator_id_dfe60d61` (`creator_id`),
  KEY `dvadmin_merchant_pre_history_merchant_id_da901c0e` (`merchant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='商户预付历史记录';

-- ----------------------------
-- Table structure for dvadmin_message_center
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_message_center`;
CREATE TABLE `dvadmin_message_center` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `title` varchar(100) NOT NULL COMMENT '标题',
  `content` longtext NOT NULL COMMENT '内容',
  `target_type` int NOT NULL COMMENT '目标类型',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_message_center_creator_id_60e2080e` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='消息中心';

-- ----------------------------
-- Table structure for dvadmin_message_center_target_role
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_message_center_target_role`;
CREATE TABLE `dvadmin_message_center_target_role` (
  `id` int NOT NULL AUTO_INCREMENT,
  `messagecenter_id` bigint NOT NULL,
  `role_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_message_center_t_messagecenter_id_role_id_f5a77970_uniq` (`messagecenter_id`,`role_id`),
  KEY `dvadmin_message_center_target_role_messagecenter_id_41a7bd9d` (`messagecenter_id`),
  KEY `dvadmin_message_center_target_role_role_id_661a61bb` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_message_center_target_user
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_message_center_target_user`;
CREATE TABLE `dvadmin_message_center_target_user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `is_read` tinyint(1) DEFAULT NULL COMMENT '是否已读',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `messagecenter_id` bigint NOT NULL COMMENT '关联消息中心表',
  `users_id` bigint NOT NULL COMMENT '关联用户表',
  PRIMARY KEY (`id`),
  KEY `dvadmin_message_center_target_user_creator_id_0a27a561` (`creator_id`),
  KEY `dvadmin_message_center_target_user_messagecenter_id_54f35bf8` (`messagecenter_id`),
  KEY `dvadmin_message_center_target_user_users_id_9ff81ff5` (`users_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='消息中心目标用户表';

-- ----------------------------
-- Table structure for dvadmin_oil_gun
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_oil_gun`;
CREATE TABLE `dvadmin_oil_gun` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `station_name` varchar(255) NOT NULL COMMENT '加油站名称',
  `oil_type` varchar(4) NOT NULL COMMENT '油类型',
  `oil_gun` int NOT NULL COMMENT '油枪号',
  `plugin_key` varchar(100) DEFAULT NULL COMMENT '插件key',
  `latitude` varchar(20) DEFAULT NULL COMMENT '纬度',
  `longitude` varchar(20) DEFAULT NULL COMMENT '经度',
  `day_limit` int NOT NULL DEFAULT '0' COMMENT '日限额',
  `limit` bigint NOT NULL DEFAULT '0' COMMENT '总限额',
  `min_money` bigint NOT NULL DEFAULT '0' COMMENT '最小金额',
  `max_money` bigint NOT NULL DEFAULT '0' COMMENT '最大金额',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `support_pay` int NOT NULL DEFAULT '0' COMMENT '支持的支付方式',
  `extra_data` json NOT NULL COMMENT '额外数据',
  `ver` bigint NOT NULL,
  `platform_id` varchar(64) DEFAULT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `platform_user_id` bigint DEFAULT NULL,
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  `plugin_id` bigint NOT NULL COMMENT '渠道类型',
  PRIMARY KEY (`id`),
  KEY `dvadmin_oil_gun_creator_id_e9540270` (`creator_id`),
  KEY `dvadmin_oil_gun_platform_user_id_6aed192d` (`platform_user_id`),
  KEY `dvadmin_oil_gun_writeoff_id_6236c39f` (`writeoff_id`),
  KEY `dvadmin_oil_gun_plugin_id_dc467e9a` (`plugin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='油枪';

-- ----------------------------
-- Table structure for dvadmin_oil_gun_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_oil_gun_flow`;
CREATE TABLE `dvadmin_oil_gun_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `flow` bigint NOT NULL DEFAULT '0' COMMENT '已充值金额',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `gun_id` bigint DEFAULT NULL COMMENT '关联油枪',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_oil_gun_flow_creator_id_d0ef6510` (`creator_id`),
  KEY `dvadmin_oil_gun_flow_gun_id_96660afb` (`gun_id`),
  KEY `dvadmin_oil_gun_flow_writeoff_id_b474fb62` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='油枪流水';

-- ----------------------------
-- Table structure for dvadmin_order
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_order`;
CREATE TABLE `dvadmin_order` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `id` varchar(30) NOT NULL COMMENT '订单Id',
  `order_no` varchar(32) NOT NULL COMMENT '本系统订单号',
  `out_order_no` varchar(32) NOT NULL COMMENT '商户订单号',
  `order_status` int NOT NULL COMMENT '订单状态',
  `money` int NOT NULL COMMENT '金额',
  `tax` int NOT NULL COMMENT '手续费',
  `pay_datetime` datetime(6) DEFAULT NULL COMMENT '支付时间',
  `product_name` varchar(255) DEFAULT NULL COMMENT '通道名称',
  `req_extra` longtext COMMENT '额外请求参数',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `compatible` int NOT NULL DEFAULT '0' COMMENT '系统兼容',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `merchant_id` bigint DEFAULT NULL COMMENT '关联商户',
  `writeoff_id` bigint DEFAULT NULL COMMENT '核销',
  `pay_channel_id` bigint DEFAULT NULL COMMENT '关联支付通道',
  PRIMARY KEY (`id`),
  UNIQUE KEY `order_no` (`order_no`),
  UNIQUE KEY `out_order_no` (`out_order_no`),
  KEY `dvadmin_order_order_status_8284af99` (`order_status`),
  KEY `dvadmin_order_create_datetime_eafae3f1` (`create_datetime`),
  KEY `dvadmin_order_creator_id_e909860e` (`creator_id`),
  KEY `dvadmin_order_merchant_id_12a68200` (`merchant_id`),
  KEY `dvadmin_order_writeoff_id_dbb09f4c` (`writeoff_id`),
  KEY `dvadmin_order_pay_channel_id_06b81603` (`pay_channel_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='订单';

-- ----------------------------
-- Table structure for dvadmin_order_detail
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_order_detail`;
CREATE TABLE `dvadmin_order_detail` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `notify_url` longtext NOT NULL COMMENT '通知地址',
  `jump_url` longtext NOT NULL COMMENT '跳转地址',
  `product_id` varchar(255) DEFAULT NULL COMMENT '商品',
  `cookie_id` varchar(255) DEFAULT NULL COMMENT '小号',
  `notify_money` int NOT NULL COMMENT '通知金额',
  `ticket_no` varchar(255) DEFAULT NULL COMMENT '官方流水号',
  `query_no` varchar(255) DEFAULT NULL COMMENT '查询订单号',
  `plugin_type` varchar(255) DEFAULT NULL COMMENT '插件支付类型',
  `plugin_upstream` int NOT NULL DEFAULT '-1' COMMENT '插件大类',
  `merchant_tax` int NOT NULL DEFAULT '0' COMMENT '商户手续费',
  `extra` json NOT NULL COMMENT '额外数据',
  `remarks` longtext COMMENT '备注',
  `buyer_id` varchar(255) DEFAULT NULL COMMENT '买家ID',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `order_id` varchar(30) NOT NULL COMMENT '关联订单',
  `writeoff_id` bigint DEFAULT NULL COMMENT '核销',
  `domain_id` bigint DEFAULT NULL COMMENT '域名',
  `plugin_id` bigint DEFAULT NULL COMMENT '插件',
  PRIMARY KEY (`id`),
  UNIQUE KEY `order_id` (`order_id`),
  KEY `dvadmin_order_detail_product_id_1bf5ab80` (`product_id`),
  KEY `dvadmin_order_detail_plugin_upstream_7a9a2812` (`plugin_upstream`),
  KEY `dvadmin_order_detail_creator_id_b1f389c0` (`creator_id`),
  KEY `dvadmin_order_detail_writeoff_id_b1d3391e` (`writeoff_id`),
  KEY `dvadmin_order_detail_domain_id_6abaf959` (`domain_id`),
  KEY `dvadmin_order_detail_plugin_id_e2689eff` (`plugin_id`)
) ENGINE=InnoDB AUTO_INCREMENT=369169 DEFAULT CHARSET=utf8mb3 COMMENT='订单详细';

-- ----------------------------
-- Table structure for dvadmin_order_device_detail
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_order_device_detail`;
CREATE TABLE `dvadmin_order_device_detail` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `ip_address` varchar(255) NOT NULL COMMENT 'Ip地址',
  `address` varchar(32) DEFAULT NULL COMMENT '归属地',
  `device_type` int NOT NULL COMMENT '设备类型',
  `device_fingerprint` varchar(255) DEFAULT NULL COMMENT '设备指纹',
  `pid` int NOT NULL DEFAULT '-1' COMMENT '代理省ip',
  `cid` int NOT NULL DEFAULT '-1' COMMENT '代理城市ip',
  `user_id` varchar(32) DEFAULT NULL COMMENT '用户id',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `order_id` varchar(30) NOT NULL COMMENT '系统订单',
  PRIMARY KEY (`id`),
  UNIQUE KEY `order_id` (`order_id`),
  KEY `dvadmin_order_device_detail_ip_address_a33a7a3d` (`ip_address`),
  KEY `dvadmin_order_device_detail_device_type_c159f7ce` (`device_type`),
  KEY `dvadmin_order_device_detail_user_id_9dffbe8d` (`user_id`),
  KEY `dvadmin_order_device_detail_creator_id_e24dd674` (`creator_id`)
) ENGINE=InnoDB AUTO_INCREMENT=334221 DEFAULT CHARSET=utf8mb3 COMMENT='订单用户信息';

-- ----------------------------
-- Table structure for dvadmin_order_log
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_order_log`;
CREATE TABLE `dvadmin_order_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `out_order_no` varchar(32) NOT NULL COMMENT '外部订单号',
  `sign_raw` longtext COMMENT '签名原始数据',
  `sign` varchar(32) DEFAULT NULL COMMENT '签名数据',
  `request_body` longtext COMMENT '请求参数',
  `request_method` varchar(8) DEFAULT NULL COMMENT '请求方式',
  `response_code` varchar(32) DEFAULT NULL COMMENT '响应状态码',
  `json_result` longtext COMMENT '返回信息',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  UNIQUE KEY `out_order_no` (`out_order_no`),
  KEY `dvadmin_order_log_creator_id_4fee3955` (`creator_id`)
) ENGINE=InnoDB AUTO_INCREMENT=374520 DEFAULT CHARSET=utf8mb3 COMMENT='订单日志';

-- ----------------------------
-- Table structure for dvadmin_pay_channel
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_pay_channel`;
CREATE TABLE `dvadmin_pay_channel` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `name` varchar(64) NOT NULL COMMENT '通道名称',
  `status` tinyint(1) NOT NULL COMMENT '通道状态',
  `max_money` int NOT NULL COMMENT '单笔最大金额',
  `min_money` int NOT NULL COMMENT '单笔最小金额',
  `float_max_money` int NOT NULL COMMENT '浮动单笔最大金额',
  `float_min_money` int NOT NULL COMMENT '浮动单笔最小金额',
  `settled` tinyint(1) NOT NULL COMMENT '固定金额模式',
  `moneys` json NOT NULL COMMENT '固定金额列表',
  `start_time` varchar(8) NOT NULL COMMENT '启用时间',
  `end_time` varchar(8) NOT NULL COMMENT '结束时间',
  `extra_arg` int DEFAULT NULL COMMENT '额外参数',
  `ban_ip` json NOT NULL COMMENT '封禁IP列表',
  `logo` longtext COMMENT '图标',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `plugin_id` bigint NOT NULL COMMENT '支付插件',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  KEY `dvadmin_pay_channel_creator_id_4b904277` (`creator_id`),
  KEY `dvadmin_pay_channel_plugin_id_231d48e0` (`plugin_id`)
) ENGINE=InnoDB AUTO_INCREMENT=8008 DEFAULT CHARSET=utf8mb3 COMMENT='支付通道';

-- ----------------------------
-- Table structure for dvadmin_pay_channel_tax
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_pay_channel_tax`;
CREATE TABLE `dvadmin_pay_channel_tax` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `tax` decimal(5,2) NOT NULL COMMENT '费率',
  `mark` varchar(100) NOT NULL COMMENT '标志(通道id-租户id)',
  `status` tinyint(1) NOT NULL COMMENT '通道状态',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `pay_channel_id` bigint NOT NULL COMMENT '绑定的支付通道',
  `tenant_id` bigint NOT NULL COMMENT '绑定的租户',
  PRIMARY KEY (`id`),
  UNIQUE KEY `mark` (`mark`),
  KEY `dvadmin_pay_channel_tax_creator_id_9df07417` (`creator_id`),
  KEY `dvadmin_pay_channel_tax_pay_channel_id_aa4f3331` (`pay_channel_id`),
  KEY `dvadmin_pay_channel_tax_tenant_id_a2a7383d` (`tenant_id`)
) ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=utf8mb3 COMMENT='支付通道费率';

-- ----------------------------
-- Table structure for dvadmin_pay_domain
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_pay_domain`;
CREATE TABLE `dvadmin_pay_domain` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `url` varchar(255) NOT NULL COMMENT '域名',
  `app_id` varchar(255) DEFAULT NULL COMMENT 'app_id',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `pay_status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '支付状态',
  `wechat_status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '微信状态',
  `sign_type` int NOT NULL DEFAULT '0' COMMENT '签名类型',
  `public_key` longtext COMMENT '支付宝公钥',
  `private_key` longtext COMMENT '应用私钥',
  `app_public_crt` longtext COMMENT '应用公钥证书',
  `alipay_public_crt` longtext COMMENT '支付宝公钥证书',
  `alipay_root_crt` longtext COMMENT '支付宝根证书',
  `auth_status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '鉴权状态',
  `auth_timeout` int NOT NULL DEFAULT '0' COMMENT '鉴权时间',
  `auth_key` varchar(255) DEFAULT NULL COMMENT '鉴权密钥',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  UNIQUE KEY `url` (`url`),
  KEY `dvadmin_pay_domain_creator_id_edd771f6` (`creator_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb3 COMMENT='支付域名';

-- ----------------------------
-- Table structure for dvadmin_pay_plugin
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_pay_plugin`;
CREATE TABLE `dvadmin_pay_plugin` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `name` varchar(64) NOT NULL COMMENT '支付插件名称',
  `description` longtext NOT NULL COMMENT '支付插件描述',
  `status` tinyint(1) NOT NULL COMMENT '支付插件状态',
  `can_divide` tinyint(1) NOT NULL COMMENT '是否可以分账',
  `can_transfer` tinyint(1) NOT NULL COMMENT '是否可以转账',
  `support_device` int NOT NULL COMMENT '支持设备',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  KEY `dvadmin_pay_plugin_creator_id_27676731` (`creator_id`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb3 COMMENT='支付插件';

-- ----------------------------
-- Table structure for dvadmin_pay_plugin_config
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_pay_plugin_config`;
CREATE TABLE `dvadmin_pay_plugin_config` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `title` varchar(50) NOT NULL COMMENT '标题',
  `key` varchar(20) NOT NULL COMMENT '关键字',
  `value` json DEFAULT NULL COMMENT '值',
  `sort` int NOT NULL COMMENT '排序',
  `status` tinyint(1) NOT NULL COMMENT '启用状态',
  `data_options` json DEFAULT NULL COMMENT '数据options',
  `form_item_type` int NOT NULL COMMENT '表单类型',
  `rule` json DEFAULT NULL COMMENT '校验规则',
  `placeholder` varchar(50) DEFAULT NULL COMMENT '提示信息',
  `setting` json DEFAULT NULL COMMENT '配置',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `parent_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_pay_plugin_config_key_parent_id_5ccc9259_uniq` (`key`,`parent_id`),
  KEY `dvadmin_pay_plugin_config_key_2b902fff` (`key`),
  KEY `dvadmin_pay_plugin_config_creator_id_ddbc7464` (`creator_id`),
  KEY `dvadmin_pay_plugin_config_parent_id_a40b2164` (`parent_id`)
) ENGINE=InnoDB AUTO_INCREMENT=253 DEFAULT CHARSET=utf8mb3 COMMENT='插件配置表';

-- ----------------------------
-- Table structure for dvadmin_pay_plugin_menus
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_pay_plugin_menus`;
CREATE TABLE `dvadmin_pay_plugin_menus` (
  `id` int NOT NULL AUTO_INCREMENT,
  `payplugin_id` bigint NOT NULL,
  `menu_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_pay_plugin_menus_payplugin_id_menu_id_d3613e49_uniq` (`payplugin_id`,`menu_id`),
  KEY `dvadmin_pay_plugin_menus_payplugin_id_2171e1e4` (`payplugin_id`),
  KEY `dvadmin_pay_plugin_menus_menu_id_5fc3489d` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_pay_plugin_pay_types
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_pay_plugin_pay_types`;
CREATE TABLE `dvadmin_pay_plugin_pay_types` (
  `id` int NOT NULL AUTO_INCREMENT,
  `payplugin_id` bigint NOT NULL,
  `paytype_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_pay_plugin_pay_t_payplugin_id_paytype_id_48bb7a12_uniq` (`payplugin_id`,`paytype_id`),
  KEY `dvadmin_pay_plugin_pay_types_payplugin_id_fcec6ffb` (`payplugin_id`),
  KEY `dvadmin_pay_plugin_pay_types_paytype_id_302299d0` (`paytype_id`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_pay_type
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_pay_type`;
CREATE TABLE `dvadmin_pay_type` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `name` varchar(64) NOT NULL COMMENT '支付方式名称',
  `key` varchar(64) NOT NULL COMMENT '支付方式关键字',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '支付方式状态',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  UNIQUE KEY `key` (`key`),
  KEY `dvadmin_pay_type_creator_id_a079f154` (`creator_id`)
) ENGINE=InnoDB AUTO_INCREMENT=61 DEFAULT CHARSET=utf8mb3 COMMENT='支付方式';

-- ----------------------------
-- Table structure for dvadmin_person_code
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_person_code`;
CREATE TABLE `dvadmin_person_code` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `name` varchar(255) NOT NULL COMMENT '名称',
  `code_type` int NOT NULL COMMENT '类型',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态',
  `limit_money` int NOT NULL DEFAULT '0' COMMENT '限制金额',
  `image` longtext COMMENT '图片',
  `session` longtext COMMENT 'session',
  `extra` json NOT NULL COMMENT '额外信息',
  `ver` bigint NOT NULL,
  `alipay_id` bigint DEFAULT NULL COMMENT '关联支付宝主体',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_person_code_alipay_id_e2e15f47` (`alipay_id`),
  KEY `dvadmin_person_code_creator_id_2f47f09e` (`creator_id`),
  KEY `dvadmin_person_code_writeoff_id_aff5d861` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='个码';

-- ----------------------------
-- Table structure for dvadmin_person_code_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_person_code_day`;
CREATE TABLE `dvadmin_person_code_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `code_id` bigint DEFAULT NULL COMMENT '关联个码',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_person_code_day_code_id_e8c9dda2` (`code_id`),
  KEY `dvadmin_person_code_day_writeoff_id_96708239` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='个码每日统计';

-- ----------------------------
-- Table structure for dvadmin_phone_order_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_phone_order_flow`;
CREATE TABLE `dvadmin_phone_order_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `flow` bigint NOT NULL DEFAULT '0' COMMENT '已充值金额',
  `refund` bigint NOT NULL DEFAULT '0' COMMENT '退款金额',
  `charge_type` int NOT NULL COMMENT '订单类型',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_phone_order_flow_date_writeoff_id_charge__a7b6c552_uniq` (`date`,`writeoff_id`,`charge_type`),
  KEY `dvadmin_phone_order_flow_creator_id_657185a7` (`creator_id`),
  KEY `dvadmin_phone_order_flow_writeoff_id_bfb8aeb4` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='话单流水';

-- ----------------------------
-- Table structure for dvadmin_phone_product
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_phone_product`;
CREATE TABLE `dvadmin_phone_product` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(21) NOT NULL COMMENT '话单订单号',
  `phone` varchar(11) NOT NULL COMMENT '手机号',
  `province` varchar(100) DEFAULT NULL COMMENT '省份',
  `city` varchar(100) DEFAULT NULL COMMENT '城市',
  `province_code` varchar(20) DEFAULT NULL COMMENT '省份码',
  `city_code` varchar(20) DEFAULT NULL COMMENT '城市码',
  `company` int NOT NULL DEFAULT '0' COMMENT '运营商',
  `phone_order_no` varchar(255) NOT NULL COMMENT '供货商订单号',
  `money` int NOT NULL COMMENT '金额',
  `notify_url` varchar(255) NOT NULL COMMENT '回调地址',
  `charge_type` int NOT NULL COMMENT '订单类型',
  `order_status` int NOT NULL DEFAULT '0' COMMENT '订单状态',
  `finish_datetime` datetime(6) DEFAULT NULL COMMENT '完成时间',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `order_id` varchar(30) DEFAULT NULL COMMENT '系统订单',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  UNIQUE KEY `phone_order_no` (`phone_order_no`),
  UNIQUE KEY `order_id` (`order_id`),
  KEY `dvadmin_phone_product_creator_id_390b5d43` (`creator_id`),
  KEY `dvadmin_phone_product_writeoff_id_1e34daed` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='话单库存';

-- ----------------------------
-- Table structure for dvadmin_phone_product_notification_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_phone_product_notification_history`;
CREATE TABLE `dvadmin_phone_product_notification_history` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `url` longtext NOT NULL COMMENT '通知地址',
  `request_body` longtext COMMENT '请求参数',
  `request_method` varchar(8) DEFAULT NULL COMMENT '请求方式',
  `response_code` int NOT NULL COMMENT '响应状态码',
  `json_result` longtext COMMENT '返回信息',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `product_id` varchar(21) NOT NULL COMMENT '关联话单',
  PRIMARY KEY (`id`),
  KEY `dvadmin_phone_product_notification_history_creator_id_aac34dae` (`creator_id`),
  KEY `dvadmin_phone_product_notification_history_product_id_c81a7b76` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='话单通知记录';

-- ----------------------------
-- Table structure for dvadmin_product_tax
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_product_tax`;
CREATE TABLE `dvadmin_product_tax` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `product_id` varchar(255) NOT NULL COMMENT '产品',
  `plugin_key` varchar(255) NOT NULL COMMENT '插件Key',
  `extra_id` int NOT NULL DEFAULT '0' COMMENT '特殊标识符',
  `ver` bigint NOT NULL,
  `pre_tax` bigint NOT NULL DEFAULT '0' COMMENT '占用金额',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='插件项目占用金额';

-- ----------------------------
-- Table structure for dvadmin_qiandao_order
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_qiandao_order`;
CREATE TABLE `dvadmin_qiandao_order` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(64) NOT NULL COMMENT 'id',
  `send` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否发货',
  `confirm` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否收货',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `order_id` varchar(30) NOT NULL,
  `product_id` varchar(64) NOT NULL COMMENT '关联',
  PRIMARY KEY (`id`),
  UNIQUE KEY `order_id` (`order_id`),
  KEY `dvadmin_qiandao_order_creator_id_f68e0db4` (`creator_id`),
  KEY `dvadmin_qiandao_order_product_id_b59e4b67` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='千岛商品订单';

-- ----------------------------
-- Table structure for dvadmin_qiandao_product
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_qiandao_product`;
CREATE TABLE `dvadmin_qiandao_product` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(64) NOT NULL COMMENT 'productId',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `limit_money` bigint NOT NULL DEFAULT '0' COMMENT '限制金额',
  `money` int NOT NULL DEFAULT '0' COMMENT '金额',
  `extra` json NOT NULL DEFAULT (_utf8mb4'{}') COMMENT '额外参数',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `user_id` varchar(11) NOT NULL COMMENT '关联用户',
  PRIMARY KEY (`id`),
  KEY `dvadmin_qiandao_product_creator_id_9c0b88d1` (`creator_id`),
  KEY `dvadmin_qiandao_product_user_id_da4b2ca1` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='千岛商品';

-- ----------------------------
-- Table structure for dvadmin_qiandao_product_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_qiandao_product_day`;
CREATE TABLE `dvadmin_qiandao_product_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `product_id` varchar(64) DEFAULT NULL COMMENT '关联',
  PRIMARY KEY (`id`),
  KEY `dvadmin_qiandao_product_day_product_id_cc591f2d` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='千岛商品日统计';

-- ----------------------------
-- Table structure for dvadmin_qiandao_user
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_qiandao_user`;
CREATE TABLE `dvadmin_qiandao_user` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(11) NOT NULL COMMENT '手机号',
  `nick_name` varchar(255) DEFAULT NULL COMMENT '昵称',
  `user_id` varchar(255) NOT NULL COMMENT 'user_id',
  `token` longtext COMMENT 'ck',
  `online` tinyint(1) NOT NULL DEFAULT '0' COMMENT '在线',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `limit_money` bigint NOT NULL DEFAULT '0' COMMENT '限制金额',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_qiandao_user_creator_id_590ffd2a` (`creator_id`),
  KEY `dvadmin_qiandao_user_writeoff_id_d1e6b825` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='千岛用户';

-- ----------------------------
-- Table structure for dvadmin_qianniu_account
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_qianniu_account`;
CREATE TABLE `dvadmin_qianniu_account` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(255) NOT NULL COMMENT '千牛用户ID',
  `nickname` varchar(255) NOT NULL COMMENT '名称',
  `session` longtext NOT NULL COMMENT 'session',
  `extra` json NOT NULL COMMENT '额外信息',
  `online` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `expires` datetime(6) DEFAULT NULL COMMENT '过期时间',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_qianniu_account_creator_id_6b249a80` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='千牛账户';

-- ----------------------------
-- Table structure for dvadmin_qianniu_charger
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_qianniu_charger`;
CREATE TABLE `dvadmin_qianniu_charger` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `warning` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否风控',
  `ver` bigint NOT NULL,
  `id` varchar(28) NOT NULL,
  `background_id` varchar(255) NOT NULL COMMENT '关联账号',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_qianniu_charger_background_id_1237c5ed` (`background_id`),
  KEY `dvadmin_qianniu_charger_creator_id_96ae20b7` (`creator_id`),
  KEY `dvadmin_qianniu_charger_writeoff_id_8ffae2c9` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='千牛零钱包';

-- ----------------------------
-- Table structure for dvadmin_qianniu_charger_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_qianniu_charger_flow`;
CREATE TABLE `dvadmin_qianniu_charger_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `flow` bigint NOT NULL DEFAULT '0' COMMENT '已充值金额',
  `charger_id` varchar(28) DEFAULT NULL COMMENT '关联千牛',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_qianniu_charger_flow_charger_id_15097c76` (`charger_id`),
  KEY `dvadmin_qianniu_charger_flow_creator_id_a068e696` (`creator_id`),
  KEY `dvadmin_qianniu_charger_flow_writeoff_id_1de2dbbc` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='千牛账户零钱';

-- ----------------------------
-- Table structure for dvadmin_qianniu_receiver
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_qianniu_receiver`;
CREATE TABLE `dvadmin_qianniu_receiver` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(28) NOT NULL,
  `limit_money` bigint NOT NULL DEFAULT '0' COMMENT '限制金额',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `auto_get` tinyint(1) NOT NULL DEFAULT '0' COMMENT '自动领取',
  `ver` bigint NOT NULL,
  `background_id` varchar(255) NOT NULL COMMENT '关联账号',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `account_id` varchar(28) NOT NULL COMMENT '关联千牛',
  PRIMARY KEY (`id`),
  KEY `dvadmin_qianniu_receiver_background_id_a60f5ad8` (`background_id`),
  KEY `dvadmin_qianniu_receiver_creator_id_8b463d5f` (`creator_id`),
  KEY `dvadmin_qianniu_receiver_account_id_ccfd01ec` (`account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='千牛接收人';

-- ----------------------------
-- Table structure for dvadmin_qianniu_receiver_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_qianniu_receiver_flow`;
CREATE TABLE `dvadmin_qianniu_receiver_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `flow` bigint NOT NULL DEFAULT '0' COMMENT '已充值金额',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `receiver_id` varchar(28) DEFAULT NULL COMMENT '关联千牛接收人',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_qianniu_receiver_flow_creator_id_8e085a8f` (`creator_id`),
  KEY `dvadmin_qianniu_receiver_flow_receiver_id_20a4fc5b` (`receiver_id`),
  KEY `dvadmin_qianniu_receiver_flow_writeoff_id_80ea9408` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='千牛接收人流水';

-- ----------------------------
-- Table structure for dvadmin_qianniu_sender
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_qianniu_sender`;
CREATE TABLE `dvadmin_qianniu_sender` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `warning` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否风控',
  `ver` bigint NOT NULL,
  `id` varchar(28) NOT NULL,
  `withdrawals` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否提现回调',
  `background_id` varchar(255) NOT NULL COMMENT '关联账号',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_qianniu_sender_background_id_97a9044a` (`background_id`),
  KEY `dvadmin_qianniu_sender_creator_id_0340833b` (`creator_id`),
  KEY `dvadmin_qianniu_sender_writeoff_id_f76188f8` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='千牛发送者';

-- ----------------------------
-- Table structure for dvadmin_qianniu_sender_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_qianniu_sender_flow`;
CREATE TABLE `dvadmin_qianniu_sender_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `flow` bigint NOT NULL DEFAULT '0' COMMENT '已充值金额',
  `account_id` varchar(28) DEFAULT NULL COMMENT '关联千牛发送人',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_qianniu_sender_flow_account_id_0130e555` (`account_id`),
  KEY `dvadmin_qianniu_sender_flow_creator_id_8857f670` (`creator_id`),
  KEY `dvadmin_qianniu_sender_flow_writeoff_id_4a27650c` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='千牛账户流水';

-- ----------------------------
-- Table structure for dvadmin_qianniu_small_account
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_qianniu_small_account`;
CREATE TABLE `dvadmin_qianniu_small_account` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `warning` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否风控',
  `ver` bigint NOT NULL,
  `id` varchar(28) NOT NULL,
  `background_id` varchar(255) NOT NULL COMMENT '关联账号',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_qianniu_small_account_background_id_244e9354` (`background_id`),
  KEY `dvadmin_qianniu_small_account_creator_id_bfd470a7` (`creator_id`),
  KEY `dvadmin_qianniu_small_account_writeoff_id_acb73a2f` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='千牛小号';

-- ----------------------------
-- Table structure for dvadmin_qianniu_small_account_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_qianniu_small_account_day`;
CREATE TABLE `dvadmin_qianniu_small_account_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `small_id` varchar(28) DEFAULT NULL COMMENT '关联千牛小号',
  PRIMARY KEY (`id`),
  KEY `dvadmin_qianniu_small_account_day_small_id_2b45e053` (`small_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='千牛小号记录';

-- ----------------------------
-- Table structure for dvadmin_query_log
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_query_log`;
CREATE TABLE `dvadmin_query_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `out_order_no` varchar(32) DEFAULT NULL COMMENT '外部订单号',
  `order_no` varchar(32) DEFAULT NULL COMMENT '系统订单号',
  `url` longtext COMMENT '地址',
  `request_body` longtext COMMENT '请求参数',
  `request_method` varchar(8) DEFAULT NULL COMMENT '请求方式',
  `response_code` varchar(32) DEFAULT NULL COMMENT '响应状态码',
  `json_result` longtext COMMENT '返回信息',
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_query_log_out_order_no_6b05ed6c` (`out_order_no`),
  KEY `dvadmin_query_log_order_no_b3406480` (`order_no`),
  KEY `dvadmin_query_log_remarks_1db21aee` (`remarks`),
  KEY `dvadmin_query_log_creator_id_e7248e05` (`creator_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3606702 DEFAULT CHARSET=utf8mb3 COMMENT='查询日志';

-- ----------------------------
-- Table structure for dvadmin_quick_transfer_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_quick_transfer_history`;
CREATE TABLE `dvadmin_quick_transfer_history` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `money` int NOT NULL COMMENT '金额',
  `error` longtext COMMENT '错误原因',
  `ticket_order_no` varchar(255) DEFAULT NULL COMMENT '支付宝订单号',
  `uid` varchar(255) DEFAULT NULL COMMENT '支付宝UID',
  `product_name` varchar(255) DEFAULT NULL COMMENT '支付宝产品名称',
  `user_username` varchar(255) DEFAULT NULL COMMENT '支付宝用户账号',
  `user_username_type` int NOT NULL DEFAULT '0' COMMENT '支付宝用户账号类型',
  `user_name` varchar(255) DEFAULT NULL COMMENT '支付宝用户姓名',
  `writeoff_name` varchar(255) DEFAULT NULL COMMENT '核销名称',
  `writeoff` bigint NOT NULL DEFAULT '0' COMMENT '核销',
  `tenant_id` bigint NOT NULL DEFAULT '0' COMMENT '租户',
  `product_type` int NOT NULL DEFAULT '0' COMMENT '分账产品类型',
  `split_type` int NOT NULL DEFAULT '0' COMMENT '分账类型',
  `settle_no` varchar(64) DEFAULT NULL COMMENT '分账返回号',
  `ver` bigint NOT NULL,
  `id` varchar(30) NOT NULL COMMENT '转账订单号',
  `transfer_status` int NOT NULL DEFAULT '0' COMMENT '转账状态',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `alipay_product_id` bigint DEFAULT NULL COMMENT '关联项目',
  `alipay_user_id` bigint DEFAULT NULL COMMENT '关联支付宝用户',
  `alipay_user_group_id` bigint DEFAULT NULL COMMENT '关联用户组',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `order_id` varchar(30) DEFAULT NULL COMMENT '关联订单',
  PRIMARY KEY (`id`),
  KEY `dvadmin_quick_transfer_history_tenant_id_6a6e055b` (`tenant_id`),
  KEY `dvadmin_quick_transfer_history_transfer_status_47824870` (`transfer_status`),
  KEY `dvadmin_quick_transfer_history_create_datetime_bc97913a` (`create_datetime`),
  KEY `dvadmin_quick_transfer_history_alipay_product_id_906c821d` (`alipay_product_id`),
  KEY `dvadmin_quick_transfer_history_alipay_user_id_208488c1` (`alipay_user_id`),
  KEY `dvadmin_quick_transfer_history_alipay_user_group_id_b3367b21` (`alipay_user_group_id`),
  KEY `dvadmin_quick_transfer_history_creator_id_5fd8a62a` (`creator_id`),
  KEY `dvadmin_quick_transfer_history_order_id_60febe28` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='快速转账记录';

-- ----------------------------
-- Table structure for dvadmin_re_order
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_re_order`;
CREATE TABLE `dvadmin_re_order` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `old_order_status` int NOT NULL COMMENT '补单前订单状态',
  `new_order_status` int NOT NULL COMMENT '补单后订单状态',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `order_id` varchar(30) NOT NULL COMMENT '补单订单',
  PRIMARY KEY (`id`),
  KEY `dvadmin_re_order_creator_id_8c51450f` (`creator_id`),
  KEY `dvadmin_re_order_order_id_56ca1952` (`order_id`)
) ENGINE=InnoDB AUTO_INCREMENT=173 DEFAULT CHARSET=utf8mb3 COMMENT='补单';

-- ----------------------------
-- Table structure for dvadmin_recharge_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_recharge_history`;
CREATE TABLE `dvadmin_recharge_history` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(30) NOT NULL,
  `exchange_rates` int NOT NULL COMMENT '汇率',
  `cny_amount` bigint NOT NULL COMMENT '充值金额(CNY)',
  `usdt_amount` bigint NOT NULL COMMENT '充值金额(USDT)',
  `pay_hash` varchar(100) DEFAULT NULL COMMENT '支付哈希',
  `payment_address` varchar(100) DEFAULT NULL COMMENT '支付地址',
  `payee_address` varchar(100) NOT NULL COMMENT '收款地址',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `user_id` bigint DEFAULT NULL COMMENT '充值用户',
  PRIMARY KEY (`id`),
  KEY `dvadmin_recharge_history_creator_id_d2918565` (`creator_id`),
  KEY `dvadmin_recharge_history_user_id_c9e75c09` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='自助充值记录';

-- ----------------------------
-- Table structure for dvadmin_safe_book
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_safe_book`;
CREATE TABLE `dvadmin_safe_book` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(64) NOT NULL COMMENT 'ID',
  `agreement_no` varchar(64) DEFAULT NULL COMMENT '通银签约号',
  `name` varchar(64) NOT NULL COMMENT '名称',
  `mch_id` varchar(64) NOT NULL COMMENT 'mchId',
  `status` varchar(64) NOT NULL COMMENT '状态',
  `extra` json NOT NULL COMMENT '额外信息',
  `amount` bigint NOT NULL DEFAULT '0' COMMENT '可用余额',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_safe_book_creator_id_af3fa1ea` (`creator_id`),
  KEY `dvadmin_safe_book_writeoff_id_4dd992ee` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='安全发账簿';

-- ----------------------------
-- Table structure for dvadmin_safe_book_recharge_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_safe_book_recharge_history`;
CREATE TABLE `dvadmin_safe_book_recharge_history` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(22) NOT NULL COMMENT 'ID',
  `order_id` varchar(64) NOT NULL COMMENT 'order_id',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态',
  `pay_url` longtext,
  `amount` bigint NOT NULL DEFAULT '0' COMMENT '金额',
  `ver` bigint NOT NULL,
  `third_order_id` varchar(64) DEFAULT NULL COMMENT '支付宝转账订单号',
  `third_fund_order_id` varchar(64) DEFAULT NULL COMMENT '支付宝资金流水号',
  `book_id` varchar(64) NOT NULL COMMENT '关联账簿',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_safe_book_recharge_history_book_id_f4fe5a49` (`book_id`),
  KEY `dvadmin_safe_book_recharge_history_creator_id_938d86a8` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='安全发账簿充值记录';

-- ----------------------------
-- Table structure for dvadmin_safe_book_sign
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_safe_book_sign`;
CREATE TABLE `dvadmin_safe_book_sign` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(22) NOT NULL COMMENT 'ID',
  `agreement_no` varchar(64) NOT NULL COMMENT '通银签约号',
  `third_user_id` varchar(64) DEFAULT NULL COMMENT '支付宝用户ID',
  `third_logon_id` varchar(64) DEFAULT NULL COMMENT '支付宝签约账号',
  `third_agreement_no` varchar(64) DEFAULT NULL COMMENT '支付宝协议号',
  `url` longtext,
  `mch_id` varchar(64) NOT NULL COMMENT 'mchId',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_safe_book_sign_creator_id_94d4d6d7` (`creator_id`),
  KEY `dvadmin_safe_book_sign_writeoff_id_78b10540` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='安全发签约';

-- ----------------------------
-- Table structure for dvadmin_safe_book_transfer_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_safe_book_transfer_history`;
CREATE TABLE `dvadmin_safe_book_transfer_history` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(22) NOT NULL COMMENT 'ID',
  `order_id` varchar(64) NOT NULL COMMENT 'order_id',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态',
  `result` int NOT NULL DEFAULT '0' COMMENT '状态',
  `count` int NOT NULL DEFAULT '0' COMMENT '批次转账笔数',
  `success_count` int NOT NULL DEFAULT '0' COMMENT '批次转账成功笔数',
  `fail_count` int NOT NULL DEFAULT '0' COMMENT '批次转账失败笔数',
  `amount` bigint NOT NULL DEFAULT '0' COMMENT '金额',
  `fail_amount` bigint NOT NULL DEFAULT '0' COMMENT '失败金额',
  `success_amount` bigint NOT NULL DEFAULT '0' COMMENT '成功金额',
  `ver` bigint NOT NULL,
  `book_id` varchar(64) NOT NULL COMMENT '关联账簿',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_safe_book_transfer_history_book_id_62c2d2fc` (`book_id`),
  KEY `dvadmin_safe_book_transfer_history_creator_id_1aa8550e` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='安全发账簿转出记录';

-- ----------------------------
-- Table structure for dvadmin_safe_book_transfer_user_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_safe_book_transfer_user_history`;
CREATE TABLE `dvadmin_safe_book_transfer_user_history` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(22) NOT NULL COMMENT 'ID',
  `order_id` varchar(64) DEFAULT NULL COMMENT 'order_id',
  `third_order_id` varchar(64) DEFAULT NULL COMMENT '支付宝转账订单号',
  `third_fund_order_id` varchar(64) DEFAULT NULL COMMENT '支付宝资金流水号',
  `fail_msg` varchar(64) DEFAULT NULL COMMENT 'failMsg',
  `type` varchar(16) NOT NULL COMMENT 'type,A：到户,B：到卡',
  `title` varchar(64) NOT NULL COMMENT 'title',
  `identity` varchar(64) NOT NULL COMMENT '收款人账号（支付宝账号/银行卡号）',
  `name` varchar(32) NOT NULL COMMENT '收款人姓名',
  `account_type` varchar(32) DEFAULT NULL COMMENT '收款账户类型。1:对公2:对私',
  `inst_name` varchar(64) DEFAULT NULL COMMENT '对公户银行机构名称必选',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态',
  `amount` bigint NOT NULL DEFAULT '0' COMMENT '金额',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `parent_id` varchar(22) NOT NULL COMMENT '关联记录',
  PRIMARY KEY (`id`),
  KEY `dvadmin_safe_book_transfer_user_history_creator_id_3aea7b59` (`creator_id`),
  KEY `dvadmin_safe_book_transfer_user_history_parent_id_a148aef8` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='安全发账簿转出用户记录';

-- ----------------------------
-- Table structure for dvadmin_safe_book_transfer_user_history_yf
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_safe_book_transfer_user_history_yf`;
CREATE TABLE `dvadmin_safe_book_transfer_user_history_yf` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(22) NOT NULL COMMENT 'ID',
  `order_id` varchar(64) DEFAULT NULL COMMENT 'order_id',
  `fail_msg` longtext COMMENT 'failMsg',
  `type` int NOT NULL DEFAULT '1' COMMENT 'type,1：到户,2：到卡',
  `title` varchar(64) NOT NULL COMMENT 'title',
  `identity` varchar(64) NOT NULL COMMENT '收款人账号（支付宝账号/银行卡号）',
  `name` varchar(32) NOT NULL COMMENT '收款人姓名',
  `account_type` varchar(32) NOT NULL COMMENT '收款账户类型。1:对公2:对私',
  `inst_name` varchar(64) DEFAULT NULL COMMENT '对公户银行机构名称必选',
  `inst_branch_name` varchar(128) DEFAULT NULL COMMENT '收款银行所属支行 对公必填 对私可不填',
  `inst_city` varchar(64) DEFAULT NULL COMMENT '收款银行所在市 对公必填 对私可不填',
  `inst_province` varchar(64) DEFAULT NULL COMMENT '银行所在省份 对公必填 对私可不填',
  `bank_code` varchar(64) DEFAULT NULL COMMENT '银行支行联行号 对公必填 对私可不填',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态,-1失败,0等待,1成功',
  `money` bigint NOT NULL DEFAULT '0' COMMENT '金额',
  `ver` bigint NOT NULL,
  `domain` int NOT NULL DEFAULT '1' COMMENT '类型',
  `alipay_user_id` bigint DEFAULT NULL COMMENT '关联支付宝用户',
  `alipay_user_group_id` bigint DEFAULT NULL COMMENT '关联用户组',
  `book_id` bigint NOT NULL COMMENT '关联账簿',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `merchant_id` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `dvadmin_safe_book_transfer__alipay_user_id_c91645bb` (`alipay_user_id`),
  KEY `dvadmin_safe_book_transfer__alipay_user_group_id_12751742` (`alipay_user_group_id`),
  KEY `dvadmin_safe_book_transfer_user_history_yf_book_id_43f90c0e` (`book_id`),
  KEY `dvadmin_safe_book_transfer_user_history_yf_creator_id_efaf7350` (`creator_id`),
  KEY `dvadmin_safe_book_transfer_user_history_yf_merchant_id_52fb5ed4` (`merchant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='安全发账簿转出用户记录';

-- ----------------------------
-- Table structure for dvadmin_safe_book_xh
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_safe_book_xh`;
CREATE TABLE `dvadmin_safe_book_xh` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `account_book_id` varchar(64) NOT NULL,
  `name` varchar(100) NOT NULL,
  `phone` varchar(100) NOT NULL,
  `app_id` varchar(64) NOT NULL,
  `proxy_id` varchar(64) NOT NULL,
  `agreement_id` varchar(64) NOT NULL,
  `key` varchar(128) NOT NULL,
  `balance` bigint NOT NULL DEFAULT '0' COMMENT '余额',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`account_book_id`),
  KEY `dvadmin_safe_book_xh_creator_id_c888d151` (`creator_id`),
  KEY `dvadmin_safe_book_xh_writeoff_id_4bbacd37` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_safe_book_xh_withdraw_order
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_safe_book_xh_withdraw_order`;
CREATE TABLE `dvadmin_safe_book_xh_withdraw_order` (
  `id` varchar(22) NOT NULL COMMENT 'ID',
  `order_id` varchar(64) DEFAULT NULL,
  `name` varchar(64) NOT NULL,
  `identity` varchar(128) NOT NULL,
  `money` bigint NOT NULL DEFAULT '0',
  `cash_money` bigint NOT NULL DEFAULT '0',
  `tax` bigint NOT NULL DEFAULT '0',
  `type` varchar(1) NOT NULL,
  `account_type` varchar(1) NOT NULL,
  `inst_name` varchar(64) DEFAULT NULL,
  `bank_code` varchar(64) DEFAULT NULL,
  `status` int NOT NULL DEFAULT '0' COMMENT '状态,-1失败,0等待,1成功',
  `processed_time` datetime(6) DEFAULT NULL,
  `platform_timestamp` bigint DEFAULT NULL,
  `fail_msg` longtext,
  `ver` bigint NOT NULL,
  `book_id` varchar(64) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dvadmin_safe_book_xh_withdraw_order_book_id_f6f3b465` (`book_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_safe_book_yf
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_safe_book_yf`;
CREATE TABLE `dvadmin_safe_book_yf` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `app_id` varchar(64) NOT NULL COMMENT 'app_id',
  `app_key` varchar(64) NOT NULL COMMENT 'app_key',
  `api_key` varchar(64) NOT NULL COMMENT 'api_key',
  `name` varchar(64) NOT NULL COMMENT '名称',
  `balance` bigint NOT NULL DEFAULT '0' COMMENT '可用余额',
  `fee` bigint NOT NULL DEFAULT '0' COMMENT '可用系统费',
  `ver` bigint NOT NULL,
  `domain` int NOT NULL DEFAULT '1' COMMENT '类型',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  UNIQUE KEY `api_key` (`api_key`),
  UNIQUE KEY `dvadmin_safe_book_yf_app_id_domain_be4f9bff_uniq` (`app_id`,`domain`),
  KEY `dvadmin_safe_book_yf_creator_id_834b394e` (`creator_id`),
  KEY `dvadmin_safe_book_yf_writeoff_id_9bec785e` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='安全发账簿';

-- ----------------------------
-- Table structure for dvadmin_safe_book_yf_quick_transfer
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_safe_book_yf_quick_transfer`;
CREATE TABLE `dvadmin_safe_book_yf_quick_transfer` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `auto` tinyint(1) NOT NULL DEFAULT '0' COMMENT '自动划转',
  `lower_limit` bigint NOT NULL DEFAULT '100000' COMMENT '金额下限',
  `run_interval` int NOT NULL DEFAULT '800' COMMENT '执行间隔毫秒',
  `money` bigint NOT NULL DEFAULT '0' COMMENT '转账金额',
  `title` varchar(64) NOT NULL DEFAULT '提成',
  `book_id` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  UNIQUE KEY `book_id` (`book_id`),
  KEY `dvadmin_safe_book_yf_quick_transfer_creator_id_5ad27fd7` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='安全发转账设置';

-- ----------------------------
-- Table structure for dvadmin_safe_book_yf_quick_transfer_groups
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_safe_book_yf_quick_transfer_groups`;
CREATE TABLE `dvadmin_safe_book_yf_quick_transfer_groups` (
  `id` int NOT NULL AUTO_INCREMENT,
  `yfsafequicktransfer_id` bigint NOT NULL,
  `alipaysplitusergroup_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_safe_book_yf_qui_yfsafequicktransfer_id_a_37f30e89_uniq` (`yfsafequicktransfer_id`,`alipaysplitusergroup_id`),
  KEY `dvadmin_safe_book_yf_quick__yfsafequicktransfer_id_f1173155` (`yfsafequicktransfer_id`),
  KEY `dvadmin_safe_book_yf_quick__alipaysplitusergroup_id_3b76adbe` (`alipaysplitusergroup_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_safe_book_yf_user_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_safe_book_yf_user_flow`;
CREATE TABLE `dvadmin_safe_book_yf_user_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `identity` varchar(64) NOT NULL COMMENT '收款人账号（支付宝账号/银行卡号）',
  `name` varchar(32) NOT NULL COMMENT '收款人姓名',
  `flow` bigint NOT NULL COMMENT '流水',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_safe_book_yf_user_flow_identity_757ebcb0` (`identity`),
  KEY `dvadmin_safe_book_yf_user_flow_name_ca80f40e` (`name`),
  KEY `dvadmin_safe_book_yf_user_flow_date_d3d28b34` (`date`),
  KEY `dvadmin_safe_book_yf_user_flow_creator_id_a433ae65` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='转账账用户流水';

-- ----------------------------
-- Table structure for dvadmin_safe_yf_merchant_detail
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_safe_yf_merchant_detail`;
CREATE TABLE `dvadmin_safe_yf_merchant_detail` (
  `merchant_id` bigint NOT NULL,
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `balance` bigint NOT NULL DEFAULT '0' COMMENT '金额',
  `ver` bigint NOT NULL,
  PRIMARY KEY (`merchant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='安全发商户详情';

-- ----------------------------
-- Table structure for dvadmin_safe_yf_merchant_detail_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_safe_yf_merchant_detail_flow`;
CREATE TABLE `dvadmin_safe_yf_merchant_detail_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `money` bigint NOT NULL DEFAULT '0' COMMENT '金额',
  `before_balance` bigint NOT NULL DEFAULT '0' COMMENT '消耗前金额',
  `after_balance` bigint NOT NULL DEFAULT '0' COMMENT '消耗后金额',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `history_id` varchar(22) DEFAULT NULL,
  `merchant_id` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `history_id` (`history_id`),
  KEY `dvadmin_safe_yf_merchant_detail_flow_creator_id_9930b918` (`creator_id`),
  KEY `dvadmin_safe_yf_merchant_detail_flow_merchant_id_e2497118` (`merchant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='安全发商户详情记录';

-- ----------------------------
-- Table structure for dvadmin_split_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_split_history`;
CREATE TABLE `dvadmin_split_history` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `money` int NOT NULL COMMENT '金额',
  `error` longtext COMMENT '错误原因',
  `ticket_order_no` varchar(255) DEFAULT NULL COMMENT '支付宝订单号',
  `uid` varchar(255) DEFAULT NULL COMMENT '支付宝UID',
  `product_name` varchar(255) DEFAULT NULL COMMENT '支付宝产品名称',
  `user_username` varchar(255) DEFAULT NULL COMMENT '支付宝用户账号',
  `user_username_type` int NOT NULL DEFAULT '0' COMMENT '支付宝用户账号类型',
  `user_name` varchar(255) DEFAULT NULL COMMENT '支付宝用户姓名',
  `writeoff_name` varchar(255) DEFAULT NULL COMMENT '核销名称',
  `writeoff` bigint NOT NULL DEFAULT '0' COMMENT '核销',
  `tenant_id` bigint NOT NULL DEFAULT '0' COMMENT '租户',
  `product_type` int NOT NULL DEFAULT '0' COMMENT '分账产品类型',
  `split_type` int NOT NULL DEFAULT '0' COMMENT '分账类型',
  `settle_no` varchar(64) DEFAULT NULL COMMENT '分账返回号',
  `ver` bigint NOT NULL,
  `id` varchar(25) NOT NULL COMMENT '分账订单号',
  `split_status` int NOT NULL DEFAULT '0' COMMENT '分账状态',
  `percentage` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '分账比例',
  `hide` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否隐藏',
  `is_async` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否异步',
  `alipay_product_id` bigint DEFAULT NULL COMMENT '关联项目',
  `alipay_user_id` bigint DEFAULT NULL COMMENT '关联支付宝用户',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `order_id` varchar(30) DEFAULT NULL COMMENT '关联订单',
  `parent_id` varchar(25) DEFAULT NULL COMMENT '父级',
  PRIMARY KEY (`id`),
  KEY `dvadmin_split_history_tenant_id_f8f6c215` (`tenant_id`),
  KEY `dvadmin_split_history_hide_9e4f3784` (`hide`),
  KEY `dvadmin_split_history_alipay_product_id_a9bbad29` (`alipay_product_id`),
  KEY `dvadmin_split_history_alipay_user_id_44d69642` (`alipay_user_id`),
  KEY `dvadmin_split_history_creator_id_c246f607` (`creator_id`),
  KEY `dvadmin_split_history_order_id_353e857d` (`order_id`),
  KEY `dvadmin_split_history_parent_id_12c122a7` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='分账记录';

-- ----------------------------
-- Table structure for dvadmin_strategic_good
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_strategic_good`;
CREATE TABLE `dvadmin_strategic_good` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `card_type` int NOT NULL COMMENT '类型',
  `place` varchar(100) DEFAULT NULL COMMENT '地区',
  `account` varchar(255) NOT NULL COMMENT '账号',
  `password` varchar(255) NOT NULL COMMENT '密码',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_strategic_good_creator_id_4139dbe8` (`creator_id`),
  KEY `dvadmin_strategic_good_writeoff_id_33220e11` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='战略物资';

-- ----------------------------
-- Table structure for dvadmin_strategic_goods_card
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_strategic_goods_card`;
CREATE TABLE `dvadmin_strategic_goods_card` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `card` varchar(255) NOT NULL COMMENT '卡号',
  `balance` int NOT NULL DEFAULT '0' COMMENT '初始余额',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `cookie` json NOT NULL COMMENT 'cookie',
  `money` int NOT NULL DEFAULT '0' COMMENT '计划金额',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `parent_id` bigint DEFAULT NULL COMMENT '关联战略物资',
  PRIMARY KEY (`id`),
  KEY `dvadmin_strategic_goods_card_creator_id_c90abfe1` (`creator_id`),
  KEY `dvadmin_strategic_goods_card_parent_id_d763cd6c` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='战略物资卡';

-- ----------------------------
-- Table structure for dvadmin_strategic_goods_card_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_strategic_goods_card_flow`;
CREATE TABLE `dvadmin_strategic_goods_card_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `flow` bigint NOT NULL DEFAULT '0' COMMENT '已充值金额',
  `card_id` bigint DEFAULT NULL COMMENT '关联战略物资卡',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_strategic_goods_card_flow_card_id_23701d88` (`card_id`),
  KEY `dvadmin_strategic_goods_card_flow_creator_id_671a079a` (`creator_id`),
  KEY `dvadmin_strategic_goods_card_flow_writeoff_id_86d46fa8` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='战略物资卡流水';

-- ----------------------------
-- Table structure for dvadmin_system_area
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_area`;
CREATE TABLE `dvadmin_system_area` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `name` varchar(100) NOT NULL COMMENT '名称',
  `code` varchar(20) NOT NULL COMMENT '地区编码',
  `level` bigint NOT NULL COMMENT '地区层级(1省份 2城市 3区县 4乡级)',
  `pinyin` varchar(255) NOT NULL COMMENT '拼音',
  `initials` varchar(20) NOT NULL COMMENT '首字母',
  `enable` tinyint(1) NOT NULL COMMENT '是否启用',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `pcode_id` varchar(20) DEFAULT NULL COMMENT '父地区编码',
  PRIMARY KEY (`id`),
  UNIQUE KEY `code` (`code`),
  KEY `dvadmin_system_area_creator_id_a5046ac0` (`creator_id`),
  KEY `dvadmin_system_area_pcode_id_f9b21462` (`pcode_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='地区表';

-- ----------------------------
-- Table structure for dvadmin_system_config
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_config`;
CREATE TABLE `dvadmin_system_config` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `title` varchar(50) NOT NULL COMMENT '标题',
  `key` varchar(20) NOT NULL COMMENT '键',
  `value` json DEFAULT NULL COMMENT '值',
  `sort` int NOT NULL COMMENT '排序',
  `status` tinyint(1) NOT NULL COMMENT '启用状态',
  `data_options` json DEFAULT NULL COMMENT '数据options',
  `form_item_type` int NOT NULL COMMENT '表单类型',
  `rule` json DEFAULT NULL COMMENT '校验规则',
  `placeholder` varchar(50) DEFAULT NULL COMMENT '提示信息',
  `setting` json DEFAULT NULL COMMENT '配置',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `parent_id` bigint DEFAULT NULL COMMENT '父级',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_system_config_key_parent_id_f8627867_uniq` (`key`,`parent_id`),
  KEY `dvadmin_system_config_key_473a4f8d` (`key`),
  KEY `dvadmin_system_config_creator_id_ba7fd60a` (`creator_id`),
  KEY `dvadmin_system_config_parent_id_1ff841b5` (`parent_id`)
) ENGINE=InnoDB AUTO_INCREMENT=95 DEFAULT CHARSET=utf8mb3 COMMENT='系统配置表';

-- ----------------------------
-- Table structure for dvadmin_system_dictionary
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_dictionary`;
CREATE TABLE `dvadmin_system_dictionary` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `label` varchar(100) DEFAULT NULL COMMENT '字典名称',
  `value` varchar(200) DEFAULT NULL COMMENT '字典编号',
  `type` int NOT NULL COMMENT '数据值类型',
  `color` varchar(20) DEFAULT NULL COMMENT '颜色',
  `is_value` tinyint(1) NOT NULL COMMENT '是否为value值',
  `status` tinyint(1) NOT NULL COMMENT '状态',
  `sort` int DEFAULT NULL COMMENT '显示排序',
  `remark` varchar(2000) DEFAULT NULL COMMENT '备注',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `parent_id` bigint DEFAULT NULL COMMENT '父级',
  PRIMARY KEY (`id`),
  KEY `dvadmin_system_dictionary_creator_id_d1b44b9d` (`creator_id`),
  KEY `dvadmin_system_dictionary_parent_id_4cceb110` (`parent_id`)
) ENGINE=InnoDB AUTO_INCREMENT=72 DEFAULT CHARSET=utf8mb3 COMMENT='字典表';

-- ----------------------------
-- Table structure for dvadmin_system_field_permission
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_field_permission`;
CREATE TABLE `dvadmin_system_field_permission` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `is_query` tinyint(1) NOT NULL COMMENT '是否可查询',
  `is_create` tinyint(1) NOT NULL COMMENT '是否可创建',
  `is_update` tinyint(1) NOT NULL COMMENT '是否可更新',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `field_id` bigint NOT NULL COMMENT '字段',
  `role_id` bigint NOT NULL COMMENT '角色',
  PRIMARY KEY (`id`),
  KEY `dvadmin_system_field_permission_creator_id_44eb775e` (`creator_id`),
  KEY `dvadmin_system_field_permission_field_id_73711ad8` (`field_id`),
  KEY `dvadmin_system_field_permission_role_id_ef32fd10` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='字段权限表';

-- ----------------------------
-- Table structure for dvadmin_system_file_list
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_file_list`;
CREATE TABLE `dvadmin_system_file_list` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `name` varchar(200) DEFAULT NULL COMMENT '名称',
  `url` varchar(100) DEFAULT NULL COMMENT '路径',
  `file_url` varchar(255) NOT NULL COMMENT '文件地址',
  `engine` varchar(100) NOT NULL COMMENT '引擎',
  `mime_type` varchar(100) NOT NULL COMMENT 'Mime类型',
  `size` bigint NOT NULL COMMENT '文件大小',
  `md5sum` varchar(36) NOT NULL COMMENT '文件md5',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_system_file_list_creator_id_dec6acb5` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='文件管理';

-- ----------------------------
-- Table structure for dvadmin_system_login_log
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_login_log`;
CREATE TABLE `dvadmin_system_login_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `username` varchar(150) DEFAULT NULL COMMENT '登录用户名',
  `ip` varchar(32) DEFAULT NULL COMMENT '登录ip',
  `agent` longtext COMMENT 'agent信息',
  `browser` varchar(200) DEFAULT NULL COMMENT '浏览器名',
  `os` varchar(200) DEFAULT NULL COMMENT '操作系统',
  `continent` varchar(50) DEFAULT NULL COMMENT '州',
  `country` varchar(50) DEFAULT NULL COMMENT '国家',
  `province` varchar(50) DEFAULT NULL COMMENT '省份',
  `city` varchar(50) DEFAULT NULL COMMENT '城市',
  `district` varchar(50) DEFAULT NULL COMMENT '县区',
  `isp` varchar(50) DEFAULT NULL COMMENT '运营商',
  `area_code` varchar(50) DEFAULT NULL COMMENT '区域代码',
  `country_english` varchar(50) DEFAULT NULL COMMENT '英文全称',
  `country_code` varchar(50) DEFAULT NULL COMMENT '简称',
  `longitude` varchar(50) DEFAULT NULL COMMENT '经度',
  `latitude` varchar(50) DEFAULT NULL COMMENT '纬度',
  `login_type` int NOT NULL COMMENT '登录类型',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_system_login_log_creator_id_5f6dc165` (`creator_id`)
) ENGINE=InnoDB AUTO_INCREMENT=61 DEFAULT CHARSET=utf8mb3 COMMENT='登录日志';

-- ----------------------------
-- Table structure for dvadmin_system_menu
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_menu`;
CREATE TABLE `dvadmin_system_menu` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `icon` varchar(64) DEFAULT NULL COMMENT '菜单图标',
  `name` varchar(64) NOT NULL COMMENT '菜单名称',
  `sort` int DEFAULT NULL COMMENT '显示排序',
  `is_link` tinyint(1) NOT NULL COMMENT '是否外链',
  `is_catalog` tinyint(1) NOT NULL COMMENT '是否目录',
  `web_path` varchar(128) DEFAULT NULL COMMENT '路由地址',
  `component` varchar(128) DEFAULT NULL COMMENT '组件地址',
  `component_name` varchar(50) DEFAULT NULL COMMENT '组件名称',
  `status` tinyint(1) NOT NULL COMMENT '菜单状态',
  `frame_out` tinyint(1) NOT NULL COMMENT '是否主框架外',
  `cache` tinyint(1) NOT NULL COMMENT '是否页面缓存',
  `visible` tinyint(1) NOT NULL COMMENT '侧边栏中是否显示',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `parent_id` bigint DEFAULT NULL COMMENT '上级菜单',
  PRIMARY KEY (`id`),
  KEY `dvadmin_system_menu_creator_id_430cdc1c` (`creator_id`),
  KEY `dvadmin_system_menu_parent_id_bc6f21bc` (`parent_id`)
) ENGINE=InnoDB AUTO_INCREMENT=90 DEFAULT CHARSET=utf8mb3 COMMENT='菜单表';

-- ----------------------------
-- Table structure for dvadmin_system_menu_button
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_menu_button`;
CREATE TABLE `dvadmin_system_menu_button` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `name` varchar(64) NOT NULL COMMENT '名称',
  `value` varchar(64) NOT NULL COMMENT '权限值',
  `api` varchar(200) NOT NULL COMMENT '接口地址',
  `method` int DEFAULT NULL COMMENT '接口请求方法',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `menu_id` bigint DEFAULT NULL COMMENT '关联菜单',
  PRIMARY KEY (`id`),
  KEY `dvadmin_system_menu_button_creator_id_3df058f7` (`creator_id`),
  KEY `dvadmin_system_menu_button_menu_id_f6aafcd8` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='菜单权限表';

-- ----------------------------
-- Table structure for dvadmin_system_menu_field
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_menu_field`;
CREATE TABLE `dvadmin_system_menu_field` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `model` varchar(64) NOT NULL COMMENT '表名',
  `field_name` varchar(64) NOT NULL COMMENT '模型表字段名',
  `title` varchar(64) NOT NULL COMMENT '字段显示名',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `menu_id` bigint NOT NULL COMMENT '菜单',
  PRIMARY KEY (`id`),
  KEY `dvadmin_system_menu_field_creator_id_084838f6` (`creator_id`),
  KEY `dvadmin_system_menu_field_menu_id_ebf37091` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='菜单字段表';

-- ----------------------------
-- Table structure for dvadmin_system_operation_log
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_operation_log`;
CREATE TABLE `dvadmin_system_operation_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `request_modular` varchar(64) DEFAULT NULL COMMENT '请求模块',
  `request_path` varchar(400) DEFAULT NULL COMMENT '请求地址',
  `request_body` longtext COMMENT '请求参数',
  `request_method` varchar(8) DEFAULT NULL COMMENT '请求方式',
  `request_msg` longtext COMMENT '操作说明',
  `request_ip` varchar(255) DEFAULT NULL COMMENT '请求ip地址',
  `request_browser` varchar(64) DEFAULT NULL COMMENT '请求浏览器',
  `request_os` varchar(64) DEFAULT NULL COMMENT '操作系统',
  `response_code` varchar(32) DEFAULT NULL COMMENT '响应状态码',
  `json_result` longtext COMMENT '返回信息',
  `status` tinyint(1) NOT NULL COMMENT '响应状态',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_system_operation_log_creator_id_0914479c` (`creator_id`)
) ENGINE=InnoDB AUTO_INCREMENT=979713 DEFAULT CHARSET=utf8mb3 COMMENT='操作日志';

-- ----------------------------
-- Table structure for dvadmin_system_role
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_role`;
CREATE TABLE `dvadmin_system_role` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `name` varchar(64) NOT NULL COMMENT '角色名称',
  `key` varchar(64) NOT NULL COMMENT '权限字符',
  `sort` int NOT NULL COMMENT '角色顺序',
  `status` tinyint(1) NOT NULL COMMENT '角色状态',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  UNIQUE KEY `key` (`key`),
  KEY `dvadmin_system_role_creator_id_a89a9bc7` (`creator_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb3 COMMENT='角色表';

-- ----------------------------
-- Table structure for dvadmin_system_role_menu
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_role_menu`;
CREATE TABLE `dvadmin_system_role_menu` (
  `id` int NOT NULL AUTO_INCREMENT,
  `role_id` bigint NOT NULL,
  `menu_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_system_role_menu_role_id_menu_id_06192289_uniq` (`role_id`,`menu_id`),
  KEY `dvadmin_system_role_menu_role_id_dcc80258` (`role_id`),
  KEY `dvadmin_system_role_menu_menu_id_7bbf1cb9` (`menu_id`)
) ENGINE=InnoDB AUTO_INCREMENT=56 DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_system_role_permission
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_role_permission`;
CREATE TABLE `dvadmin_system_role_permission` (
  `id` int NOT NULL AUTO_INCREMENT,
  `role_id` bigint NOT NULL,
  `menubutton_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_system_role_perm_role_id_menubutton_id_46c1e3ca_uniq` (`role_id`,`menubutton_id`),
  KEY `dvadmin_system_role_permission_role_id_bf988ad5` (`role_id`),
  KEY `dvadmin_system_role_permission_menubutton_id_7ba32ee0` (`menubutton_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_system_users
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_users`;
CREATE TABLE `dvadmin_system_users` (
  `password` varchar(128) NOT NULL,
  `last_login` datetime(6) DEFAULT NULL,
  `is_superuser` tinyint(1) NOT NULL,
  `first_name` varchar(150) NOT NULL,
  `last_name` varchar(150) NOT NULL,
  `is_staff` tinyint(1) NOT NULL,
  `is_active` tinyint(1) NOT NULL,
  `date_joined` datetime(6) NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `username` varchar(150) NOT NULL COMMENT '账号',
  `email` varchar(255) DEFAULT NULL COMMENT '邮箱',
  `mobile` varchar(255) DEFAULT NULL COMMENT '电话',
  `avatar` varchar(255) DEFAULT NULL COMMENT '头像',
  `name` varchar(40) NOT NULL COMMENT '昵称',
  `gender` int DEFAULT NULL COMMENT '性别',
  `status` tinyint(1) NOT NULL COMMENT '状态',
  `last_token` varchar(255) DEFAULT NULL COMMENT '最后一次登录Token',
  `key` varchar(32) DEFAULT NULL COMMENT 'key',
  `op_pwd` varchar(32) DEFAULT NULL COMMENT '操作密码',
  `telegram_user` varchar(256) DEFAULT NULL COMMENT 'tg',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `role_id` bigint DEFAULT NULL COMMENT '关联角色',
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `key` (`key`),
  KEY `dvadmin_system_users_creator_id_28556713` (`creator_id`),
  KEY `dvadmin_system_users_role_id_af3c60a6` (`role_id`)
) ENGINE=InnoDB AUTO_INCREMENT=42 DEFAULT CHARSET=utf8mb3 COMMENT='用户表';

-- ----------------------------
-- Table structure for dvadmin_system_users_groups
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_users_groups`;
CREATE TABLE `dvadmin_system_users_groups` (
  `id` int NOT NULL AUTO_INCREMENT,
  `users_id` bigint NOT NULL,
  `group_id` int NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_system_users_groups_users_id_group_id_7460f482_uniq` (`users_id`,`group_id`),
  KEY `dvadmin_system_users_groups_group_id_42e8a6dc_fk_auth_group_id` (`group_id`),
  CONSTRAINT `dvadmin_system_users_groups_group_id_42e8a6dc_fk_auth_group_id` FOREIGN KEY (`group_id`) REFERENCES `auth_group` (`id`),
  CONSTRAINT `dvadmin_system_users_users_id_f20fa5bc_fk_dvadmin_s` FOREIGN KEY (`users_id`) REFERENCES `dvadmin_system_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_system_users_user_permissions
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_system_users_user_permissions`;
CREATE TABLE `dvadmin_system_users_user_permissions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `users_id` bigint NOT NULL,
  `permission_id` int NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_system_users_use_users_id_permission_id_24cd72ef_uniq` (`users_id`,`permission_id`),
  KEY `dvadmin_system_users_permission_id_c8ec58dc_fk_auth_perm` (`permission_id`),
  CONSTRAINT `dvadmin_system_users_permission_id_c8ec58dc_fk_auth_perm` FOREIGN KEY (`permission_id`) REFERENCES `auth_permission` (`id`),
  CONSTRAINT `dvadmin_system_users_users_id_fd3b0217_fk_dvadmin_s` FOREIGN KEY (`users_id`) REFERENCES `dvadmin_system_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_taote_card_key
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_taote_card_key`;
CREATE TABLE `dvadmin_taote_card_key` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `card_no` varchar(255) NOT NULL COMMENT '卡号',
  `card_pwd` varchar(255) NOT NULL COMMENT '卡密码',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `item_id` bigint DEFAULT NULL COMMENT '关联商品',
  `order_id` varchar(30) NOT NULL COMMENT '关联订单',
  PRIMARY KEY (`id`),
  KEY `dvadmin_taote_card_key_creator_id_558b3658` (`creator_id`),
  KEY `dvadmin_taote_card_key_item_id_76458e64` (`item_id`),
  KEY `dvadmin_taote_card_key_order_id_5c5e56c5` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='淘特卡密';

-- ----------------------------
-- Table structure for dvadmin_taote_item
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_taote_item`;
CREATE TABLE `dvadmin_taote_item` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `shop_type` int NOT NULL DEFAULT '0' COMMENT '类型',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `mall_name` varchar(255) NOT NULL COMMENT '店铺名称',
  `item_name` varchar(255) NOT NULL COMMENT '商品标题',
  `limit_money` bigint NOT NULL DEFAULT '0' COMMENT '日限额',
  `money` int NOT NULL DEFAULT '0' COMMENT '商品金额',
  `item_id` varchar(16) NOT NULL COMMENT '商品id',
  `sku_id` varchar(16) NOT NULL DEFAULT '' COMMENT 'sku_id',
  `item_pic` longtext NOT NULL COMMENT '商品图片',
  `extra` longtext NOT NULL COMMENT 'extra',
  `get_card` tinyint(1) NOT NULL DEFAULT '0' COMMENT '提取卡密',
  `delay` int NOT NULL DEFAULT '10' COMMENT '延迟收货',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_taote_item_creator_id_ce4da9af` (`creator_id`),
  KEY `dvadmin_taote_item_writeoff_id_73a32762` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='淘特商品';

-- ----------------------------
-- Table structure for dvadmin_taote_item_account
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_taote_item_account`;
CREATE TABLE `dvadmin_taote_item_account` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `account` varchar(255) NOT NULL COMMENT '账号',
  `limit_money` varchar(255) NOT NULL COMMENT '日限额',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否开启',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `item_id` bigint DEFAULT NULL COMMENT '关联',
  PRIMARY KEY (`id`),
  KEY `dvadmin_taote_item_account_creator_id_d8ecbf13` (`creator_id`),
  KEY `dvadmin_taote_item_account_item_id_35450efb` (`item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='淘特充值账号';

-- ----------------------------
-- Table structure for dvadmin_taote_item_account_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_taote_item_account_day`;
CREATE TABLE `dvadmin_taote_item_account_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `account_id` bigint DEFAULT NULL COMMENT '关联账号',
  PRIMARY KEY (`id`),
  KEY `dvadmin_taote_item_account_day_account_id_88a3a211` (`account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='淘特充值账号日统计';

-- ----------------------------
-- Table structure for dvadmin_taote_item_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_taote_item_day`;
CREATE TABLE `dvadmin_taote_item_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `item_id` bigint DEFAULT NULL COMMENT '关联',
  PRIMARY KEY (`id`),
  KEY `dvadmin_taote_item_day_item_id_b5e7ff3d` (`item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='淘特商品日统计';

-- ----------------------------
-- Table structure for dvadmin_taote_item_order_detail
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_taote_item_order_detail`;
CREATE TABLE `dvadmin_taote_item_order_detail` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `status` int NOT NULL DEFAULT '0' COMMENT '订单状态',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `item_id` bigint DEFAULT NULL COMMENT '关联商品',
  `order_id` varchar(30) NOT NULL COMMENT '关联订单',
  PRIMARY KEY (`id`),
  KEY `dvadmin_taote_item_order_detail_creator_id_d6c0efa4` (`creator_id`),
  KEY `dvadmin_taote_item_order_detail_item_id_2157e741` (`item_id`),
  KEY `dvadmin_taote_item_order_detail_order_id_dc9cf3f6` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='淘特订单详细';

-- ----------------------------
-- Table structure for dvadmin_tenant
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_tenant`;
CREATE TABLE `dvadmin_tenant` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `balance` bigint NOT NULL DEFAULT '0' COMMENT '金额',
  `telegram` varchar(255) DEFAULT NULL COMMENT 'Telegram群的id',
  `trust` tinyint(1) NOT NULL DEFAULT '0' COMMENT '允许负数拉单',
  `order` tinyint(1) NOT NULL DEFAULT '0' COMMENT '允许拉取收银台',
  `pre_tax` int NOT NULL DEFAULT '0' COMMENT '占用金额',
  `polling` tinyint(1) NOT NULL DEFAULT '0' COMMENT '轮训归集',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `system_user_id` bigint DEFAULT NULL COMMENT '绑定的系统用户',
  PRIMARY KEY (`id`),
  UNIQUE KEY `system_user_id` (`system_user_id`),
  KEY `dvadmin_tenant_creator_id_73f2a1fc` (`creator_id`),
  CONSTRAINT `dvadmin_tenant_system_user_id_7cdd5ae3_fk_dvadmin_s` FOREIGN KEY (`system_user_id`) REFERENCES `dvadmin_system_users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10007 DEFAULT CHARSET=utf8mb3 COMMENT='租户';

-- ----------------------------
-- Table structure for dvadmin_tenant_cashflow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_tenant_cashflow`;
CREATE TABLE `dvadmin_tenant_cashflow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `old_money` bigint NOT NULL COMMENT '变更前余额',
  `new_money` bigint NOT NULL COMMENT '变更后余额',
  `change_money` bigint NOT NULL COMMENT '变更余额',
  `flow_type` int NOT NULL DEFAULT '0' COMMENT '流水类型',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `order_id` varchar(30) DEFAULT NULL COMMENT '系统订单',
  `pay_channel_id` bigint DEFAULT NULL COMMENT '支付通道',
  `tenant_id` bigint NOT NULL COMMENT '关联租户',
  PRIMARY KEY (`id`),
  KEY `dvadmin_tenant_cashflow_creator_id_aef2f5b8` (`creator_id`),
  KEY `dvadmin_tenant_cashflow_order_id_4d48a320` (`order_id`),
  KEY `dvadmin_tenant_cashflow_pay_channel_id_283e540e` (`pay_channel_id`),
  KEY `dvadmin_tenant_cashflow_tenant_id_b2ec1564` (`tenant_id`)
) ENGINE=InnoDB AUTO_INCREMENT=32976 DEFAULT CHARSET=utf8mb3 COMMENT='租户资金流水';

-- ----------------------------
-- Table structure for dvadmin_tenant_cookie
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_tenant_cookie`;
CREATE TABLE `dvadmin_tenant_cookie` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `content` json NOT NULL COMMENT '小号信息',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '小号状态',
  `real_name` tinyint(1) NOT NULL DEFAULT '0' COMMENT '实名状态',
  `address` tinyint(1) NOT NULL DEFAULT '0' COMMENT '地址',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `plugin_id` bigint DEFAULT NULL COMMENT '关联插件',
  `tenant_id` bigint DEFAULT NULL COMMENT '关联租户',
  `file_id` bigint NOT NULL COMMENT '小号文件',
  PRIMARY KEY (`id`),
  KEY `dvadmin_tenant_cookie_creator_id_6c3d65e4` (`creator_id`),
  KEY `dvadmin_tenant_cookie_plugin_id_e8247340` (`plugin_id`),
  KEY `dvadmin_tenant_cookie_tenant_id_d3a08ac2` (`tenant_id`),
  KEY `dvadmin_tenant_cookie_file_id_66758fe8` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='租户小号';

-- ----------------------------
-- Table structure for dvadmin_tenant_cookie_day_statistics
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_tenant_cookie_day_statistics`;
CREATE TABLE `dvadmin_tenant_cookie_day_statistics` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `cookie_id` bigint NOT NULL COMMENT '关联小号',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_tenant_cookie_da_date_cookie_id_a0a5753f_uniq` (`date`,`cookie_id`),
  KEY `dvadmin_tenant_cookie_day_statistics_cookie_id_ed91b294` (`cookie_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='租户小号日统计';

-- ----------------------------
-- Table structure for dvadmin_tenant_cookie_file
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_tenant_cookie_file`;
CREATE TABLE `dvadmin_tenant_cookie_file` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `filename` varchar(255) NOT NULL COMMENT '文件名',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `plugin_id` bigint DEFAULT NULL COMMENT '关联插件',
  `tenant_id` bigint DEFAULT NULL COMMENT '关联租户',
  PRIMARY KEY (`id`),
  KEY `dvadmin_tenant_cookie_file_creator_id_4f8e29b7` (`creator_id`),
  KEY `dvadmin_tenant_cookie_file_plugin_id_0264bae3` (`plugin_id`),
  KEY `dvadmin_tenant_cookie_file_tenant_id_915080d9` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='租户小号文件';

-- ----------------------------
-- Table structure for dvadmin_tenant_tax
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_tenant_tax`;
CREATE TABLE `dvadmin_tenant_tax` (
  `tenant_id` bigint NOT NULL COMMENT '关联租户',
  `pre_tax` bigint NOT NULL DEFAULT '0' COMMENT '占用金额',
  `ver` bigint NOT NULL,
  PRIMARY KEY (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='租户预占用金额';

-- ----------------------------
-- Table structure for dvadmin_tenant_yufu_user
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_tenant_yufu_user`;
CREATE TABLE `dvadmin_tenant_yufu_user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `telegram` varchar(255) DEFAULT NULL COMMENT 'Telegram用户id',
  `tenant_id` bigint DEFAULT NULL COMMENT '租户',
  PRIMARY KEY (`id`),
  KEY `dvadmin_tenant_yufu_user_tenant_id_8e49bb97_fk_dvadmin_tenant_id` (`tenant_id`),
  CONSTRAINT `dvadmin_tenant_yufu_user_tenant_id_8e49bb97_fk_dvadmin_tenant_id` FOREIGN KEY (`tenant_id`) REFERENCES `dvadmin_tenant` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='租户预付';

-- ----------------------------
-- Table structure for dvadmin_tiktok
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_tiktok`;
CREATE TABLE `dvadmin_tiktok` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `account` varchar(255) NOT NULL COMMENT '抖音号',
  `mstoken` longtext NOT NULL COMMENT 'mstoken',
  `place` varchar(255) NOT NULL COMMENT '地区',
  `limit` bigint NOT NULL DEFAULT '0' COMMENT '限制金额',
  `balance` bigint NOT NULL DEFAULT '0' COMMENT '余额',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_tiktok_creator_id_0857dc70` (`creator_id`),
  KEY `dvadmin_tiktok_writeoff_id_80d05e2e` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='Tiktok';

-- ----------------------------
-- Table structure for dvadmin_tiktok_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_tiktok_flow`;
CREATE TABLE `dvadmin_tiktok_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `flow` bigint NOT NULL DEFAULT '0' COMMENT '已充值金额',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `tiktok_id` bigint DEFAULT NULL COMMENT '关联Tiktok',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_tiktok_flow_creator_id_af1db3c2` (`creator_id`),
  KEY `dvadmin_tiktok_flow_tiktok_id_1d28594e` (`tiktok_id`),
  KEY `dvadmin_tiktok_flow_writeoff_id_d98fdd85` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='Tiktok流水';

-- ----------------------------
-- Table structure for dvadmin_transfer_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_transfer_history`;
CREATE TABLE `dvadmin_transfer_history` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `money` int NOT NULL COMMENT '金额',
  `error` longtext COMMENT '错误原因',
  `ticket_order_no` varchar(255) DEFAULT NULL COMMENT '支付宝订单号',
  `uid` varchar(255) DEFAULT NULL COMMENT '支付宝UID',
  `product_name` varchar(255) DEFAULT NULL COMMENT '支付宝产品名称',
  `user_username` varchar(255) DEFAULT NULL COMMENT '支付宝用户账号',
  `user_username_type` int NOT NULL DEFAULT '0' COMMENT '支付宝用户账号类型',
  `user_name` varchar(255) DEFAULT NULL COMMENT '支付宝用户姓名',
  `writeoff_name` varchar(255) DEFAULT NULL COMMENT '核销名称',
  `writeoff` bigint NOT NULL DEFAULT '0' COMMENT '核销',
  `tenant_id` bigint NOT NULL DEFAULT '0' COMMENT '租户',
  `product_type` int NOT NULL DEFAULT '0' COMMENT '分账产品类型',
  `split_type` int NOT NULL DEFAULT '0' COMMENT '分账类型',
  `settle_no` varchar(64) DEFAULT NULL COMMENT '分账返回号',
  `ver` bigint NOT NULL,
  `id` varchar(25) NOT NULL COMMENT '转账订单号',
  `transfer_status` int NOT NULL DEFAULT '0' COMMENT '转账状态',
  `alipay_product_id` bigint DEFAULT NULL COMMENT '关联项目',
  `alipay_user_id` bigint DEFAULT NULL COMMENT '关联支付宝用户',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `order_id` varchar(30) DEFAULT NULL COMMENT '关联订单',
  PRIMARY KEY (`id`),
  KEY `dvadmin_transfer_history_tenant_id_a92aecb4` (`tenant_id`),
  KEY `dvadmin_transfer_history_alipay_product_id_6b561c47` (`alipay_product_id`),
  KEY `dvadmin_transfer_history_alipay_user_id_adbfd8ea` (`alipay_user_id`),
  KEY `dvadmin_transfer_history_creator_id_cb31e573` (`creator_id`),
  KEY `dvadmin_transfer_history_order_id_74d4c7af` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='转账记录';

-- ----------------------------
-- Table structure for dvadmin_wechat_account
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_wechat_account`;
CREATE TABLE `dvadmin_wechat_account` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(30) NOT NULL,
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `content` json NOT NULL COMMENT '微信账号信息',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `background_id` varchar(255) NOT NULL COMMENT '关联账号',
  PRIMARY KEY (`id`),
  KEY `dvadmin_wechat_account_creator_id_93073954` (`creator_id`),
  KEY `dvadmin_wechat_account_background_id_97bd7192` (`background_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='微信账号';

-- ----------------------------
-- Table structure for dvadmin_wechat_account_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_wechat_account_day`;
CREATE TABLE `dvadmin_wechat_account_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `account_id` varchar(30) DEFAULT NULL COMMENT '关联微信账号',
  PRIMARY KEY (`id`),
  KEY `dvadmin_wechat_account_day_account_id_6d9e89e4` (`account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='微信账号每日统计';

-- ----------------------------
-- Table structure for dvadmin_wechat_background
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_wechat_background`;
CREATE TABLE `dvadmin_wechat_background` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(255) NOT NULL COMMENT '微信openid',
  `name` varchar(255) NOT NULL COMMENT '微信昵称',
  `device_id` varchar(16) DEFAULT NULL COMMENT '设备ID',
  `phone` varchar(20) NOT NULL COMMENT '手机号',
  `online` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `relogin` json NOT NULL DEFAULT (_utf8mb4'true') COMMENT '重登数据',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_wechat_background_creator_id_e3c7eedd` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='微信账号';

-- ----------------------------
-- Table structure for dvadmin_wechat_inbound
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_wechat_inbound`;
CREATE TABLE `dvadmin_wechat_inbound` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `apply_id` varchar(255) NOT NULL,
  `merchant_name` varchar(255) NOT NULL,
  `type` int NOT NULL COMMENT '商户类型',
  `status` int NOT NULL COMMENT '进件状态',
  `submit_date` datetime(6) DEFAULT NULL COMMENT '最后提交',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_wechat_inbound_creator_id_438a3afd` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='微信进件管理';

-- ----------------------------
-- Table structure for dvadmin_wechat_personcode_account
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_wechat_personcode_account`;
CREATE TABLE `dvadmin_wechat_personcode_account` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(255) NOT NULL COMMENT '微信openid',
  `name` varchar(255) NOT NULL COMMENT '微信昵称',
  `device_id` varchar(36) DEFAULT NULL COMMENT '设备ID',
  `online` int NOT NULL DEFAULT '1' COMMENT '是否在线(0-关闭，1-开启，-1-风控)',
  `relogin` json NOT NULL DEFAULT (_utf8mb4'true') COMMENT '重登数据',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否启用',
  `content` json NOT NULL COMMENT '微信账号信息',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_wechat_personcode_account_creator_id_fbc91630` (`creator_id`),
  KEY `dvadmin_wechat_personcode_account_writeoff_id_4b161cdd` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='微信个码账号';

-- ----------------------------
-- Table structure for dvadmin_wechat_personcode_account_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_wechat_personcode_account_day`;
CREATE TABLE `dvadmin_wechat_personcode_account_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `account_id` varchar(255) DEFAULT NULL COMMENT '关联项目',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_wechat_personcod_date_account_id_6ee237d5_uniq` (`date`,`account_id`),
  KEY `dvadmin_wechat_personcode_account_day_account_id_d0f9ecad` (`account_id`),
  KEY `dvadmin_wechat_personcode_account_day_writeoff_id_442e7097` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='微信个码账号每日统计';

-- ----------------------------
-- Table structure for dvadmin_wechat_product
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_wechat_product`;
CREATE TABLE `dvadmin_wechat_product` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `name` varchar(255) NOT NULL COMMENT '项目名称',
  `account_type` int NOT NULL DEFAULT '0' COMMENT '账户类型',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `mch_id` varchar(255) NOT NULL COMMENT '商户ID',
  `app_id` varchar(255) NOT NULL COMMENT '应用ID',
  `apiv3_key` varchar(255) NOT NULL COMMENT 'apiv3_key',
  `cert_serial_no` varchar(255) NOT NULL COMMENT '证书序列号',
  `limit_money` int NOT NULL COMMENT '限额',
  `max_money` int NOT NULL COMMENT '最大金额',
  `min_money` int NOT NULL COMMENT '最小金额',
  `float_max_money` int NOT NULL COMMENT '浮动最大金额',
  `float_min_money` int NOT NULL COMMENT '浮动最小金额',
  `collection_type` int NOT NULL COMMENT '收账类型',
  `apiclient_key` longtext NOT NULL COMMENT '证书',
  `max_fail_count` int NOT NULL DEFAULT '0' COMMENT '最多失败次数',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  KEY `dvadmin_wechat_product_creator_id_e3783d5d` (`creator_id`),
  KEY `dvadmin_wechat_product_writeoff_id_f313fb11` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='微信主体';

-- ----------------------------
-- Table structure for dvadmin_wechat_product_allow_pay_channels
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_wechat_product_allow_pay_channels`;
CREATE TABLE `dvadmin_wechat_product_allow_pay_channels` (
  `id` int NOT NULL AUTO_INCREMENT,
  `wechatproduct_id` bigint NOT NULL,
  `paychannel_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_wechat_product_a_wechatproduct_id_paychan_0ba3d036_uniq` (`wechatproduct_id`,`paychannel_id`),
  KEY `dvadmin_wechat_product_allo_wechatproduct_id_598d79b0` (`wechatproduct_id`),
  KEY `dvadmin_wechat_product_allow_pay_channels_paychannel_id_584db3f9` (`paychannel_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_wechat_product_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_wechat_product_day`;
CREATE TABLE `dvadmin_wechat_product_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `product_id` bigint DEFAULT NULL COMMENT '关联项目',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_wechat_product_day_date_product_id_fae9609e_uniq` (`date`,`product_id`),
  KEY `dvadmin_wechat_product_day_product_id_8fa365da` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='微信项目每日统计';

-- ----------------------------
-- Table structure for dvadmin_wechat_recharge
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_wechat_recharge`;
CREATE TABLE `dvadmin_wechat_recharge` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(30) NOT NULL,
  `recharge_type` int NOT NULL DEFAULT '0' COMMENT '充值类型',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态',
  `recharge_account` varchar(255) DEFAULT NULL COMMENT '充值账号',
  `account` varchar(255) DEFAULT NULL COMMENT '账号',
  `limit_money` bigint NOT NULL DEFAULT '0' COMMENT '限额',
  `content` json NOT NULL COMMENT '微信账号信息',
  `ver` bigint NOT NULL,
  `background_id` varchar(255) NOT NULL COMMENT '关联账号',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_wechat_recharge_background_id_202b88d1` (`background_id`),
  KEY `dvadmin_wechat_recharge_creator_id_4043a830` (`creator_id`),
  KEY `dvadmin_wechat_recharge_writeoff_id_08c37906` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='微信充值';

-- ----------------------------
-- Table structure for dvadmin_wechat_recharge_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_wechat_recharge_day`;
CREATE TABLE `dvadmin_wechat_recharge_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `account_id` varchar(30) DEFAULT NULL COMMENT '关联微信账号',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_wechat_recharge_day_account_id_1d9d6a65` (`account_id`),
  KEY `dvadmin_wechat_recharge_day_writeoff_id_fd86f1e9` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='微信充值每日统计';

-- ----------------------------
-- Table structure for dvadmin_weibo_account
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_weibo_account`;
CREATE TABLE `dvadmin_weibo_account` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `id` varchar(255) NOT NULL COMMENT '微博用户ID',
  `phone` varchar(13) NOT NULL COMMENT '手机',
  `nickname` varchar(255) NOT NULL COMMENT '名称',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态',
  `warning` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否风控',
  `balance` bigint NOT NULL DEFAULT '0' COMMENT '余额',
  `credit` int NOT NULL DEFAULT '0' COMMENT '信用分',
  `limit_money` int NOT NULL DEFAULT '0' COMMENT '日限额',
  `aid` varchar(60) NOT NULL COMMENT 'aid',
  `gsid` varchar(100) NOT NULL COMMENT 'gsid',
  `extra` json NOT NULL COMMENT '额外信息',
  `recv` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否是接收人',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint DEFAULT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_weibo_account_creator_id_872977c3` (`creator_id`),
  KEY `dvadmin_weibo_account_writeoff_id_2fbc569c` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='微博账号';

-- ----------------------------
-- Table structure for dvadmin_weibo_account_groups
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_weibo_account_groups`;
CREATE TABLE `dvadmin_weibo_account_groups` (
  `id` int NOT NULL AUTO_INCREMENT,
  `weiboaccount_id` varchar(255) NOT NULL,
  `weibogroup_id` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_weibo_account_gr_weiboaccount_id_weibogro_c8af2a08_uniq` (`weiboaccount_id`,`weibogroup_id`),
  KEY `dvadmin_weibo_account_groups_weiboaccount_id_607a3d26` (`weiboaccount_id`),
  KEY `dvadmin_weibo_account_groups_weibogroup_id_72c5a6f2` (`weibogroup_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for dvadmin_weibo_group
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_weibo_group`;
CREATE TABLE `dvadmin_weibo_group` (
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `name` varchar(255) NOT NULL COMMENT '群昵称',
  `group_id` varchar(255) NOT NULL COMMENT '群id',
  `group_no` varchar(255) NOT NULL COMMENT '群号',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`group_no`),
  KEY `dvadmin_weibo_group_creator_id_96159dd4` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='微博群组';

-- ----------------------------
-- Table structure for dvadmin_weibo_group_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_weibo_group_day`;
CREATE TABLE `dvadmin_weibo_group_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功领取数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '成功领取金额',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `account_id` varchar(255) DEFAULT NULL COMMENT '关联微博用户',
  `group_id` varchar(255) DEFAULT NULL COMMENT '关联群组',
  PRIMARY KEY (`id`),
  KEY `dvadmin_weibo_group_day_account_id_c0f806b8` (`account_id`),
  KEY `dvadmin_weibo_group_day_group_id_8f6c315a` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='微博群组每日统计';

-- ----------------------------
-- Table structure for dvadmin_weibo_hongbao
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_weibo_hongbao`;
CREATE TABLE `dvadmin_weibo_hongbao` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '启用状态',
  `account_id` varchar(255) DEFAULT NULL COMMENT '关联接收用户',
  `cookie_id` bigint DEFAULT NULL COMMENT '关联发送用户',
  `group_id` varchar(255) DEFAULT NULL COMMENT '关联群组',
  `order_id` varchar(30) DEFAULT NULL COMMENT '关联订单',
  PRIMARY KEY (`id`),
  KEY `dvadmin_weibo_hongbao_status_342a5d42` (`status`),
  KEY `dvadmin_weibo_hongbao_account_id_e65b0761` (`account_id`),
  KEY `dvadmin_weibo_hongbao_cookie_id_1368425c` (`cookie_id`),
  KEY `dvadmin_weibo_hongbao_group_id_a579ad3e` (`group_id`),
  KEY `dvadmin_weibo_hongbao_order_id_440b7de1` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='微博群组每日统计';

-- ----------------------------
-- Table structure for dvadmin_writeoff
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_writeoff`;
CREATE TABLE `dvadmin_writeoff` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `balance` bigint DEFAULT NULL COMMENT '金额',
  `white` json NOT NULL DEFAULT (_utf8mb4'[]') COMMENT '白名单',
  `telegram` varchar(255) DEFAULT NULL COMMENT 'Telegram群的id',
  `ver` bigint NOT NULL,
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `parent_id` bigint NOT NULL COMMENT '上级租户',
  `parent_writeoff_id` bigint DEFAULT NULL COMMENT '上级核销',
  `system_user_id` bigint DEFAULT NULL COMMENT '绑定的系统用户',
  PRIMARY KEY (`id`),
  UNIQUE KEY `system_user_id` (`system_user_id`),
  KEY `dvadmin_writeoff_parent_id_46b5f051_fk_dvadmin_tenant_id` (`parent_id`),
  KEY `dvadmin_writeoff_creator_id_2675e6b0` (`creator_id`),
  KEY `dvadmin_writeoff_parent_writeoff_id_7b900ef1` (`parent_writeoff_id`),
  CONSTRAINT `dvadmin_writeoff_parent_id_46b5f051_fk_dvadmin_tenant_id` FOREIGN KEY (`parent_id`) REFERENCES `dvadmin_tenant` (`id`),
  CONSTRAINT `dvadmin_writeoff_system_user_id_3e2c5b9d_fk_dvadmin_s` FOREIGN KEY (`system_user_id`) REFERENCES `dvadmin_system_users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=30006 DEFAULT CHARSET=utf8mb3 COMMENT='核销';

-- ----------------------------
-- Table structure for dvadmin_writeoff_brokerage
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_writeoff_brokerage`;
CREATE TABLE `dvadmin_writeoff_brokerage` (
  `writeoff_id` bigint NOT NULL COMMENT '核销',
  `brokerage` bigint NOT NULL DEFAULT '0' COMMENT '佣金',
  `ver` bigint NOT NULL,
  PRIMARY KEY (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='核销佣金';

-- ----------------------------
-- Table structure for dvadmin_writeoff_brokerage_flow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_writeoff_brokerage_flow`;
CREATE TABLE `dvadmin_writeoff_brokerage_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `old_money` bigint NOT NULL COMMENT '变更前余额',
  `new_money` bigint NOT NULL COMMENT '变更后余额',
  `change_money` bigint NOT NULL COMMENT '变更余额',
  `tax` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '费率',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `from_writeoff_id` bigint DEFAULT NULL COMMENT '来源核销',
  `order_id` varchar(30) DEFAULT NULL COMMENT '系统订单',
  `pay_channel_id` bigint DEFAULT NULL COMMENT '支付通道',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_writeoff_brokerage_flow_creator_id_e5a1d2c0` (`creator_id`),
  KEY `dvadmin_writeoff_brokerage_flow_from_writeoff_id_5209399c` (`from_writeoff_id`),
  KEY `dvadmin_writeoff_brokerage_flow_order_id_d6e20f08` (`order_id`),
  KEY `dvadmin_writeoff_brokerage_flow_pay_channel_id_b36dbe49` (`pay_channel_id`),
  KEY `dvadmin_writeoff_brokerage_flow_writeoff_id_208baef9` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='核销佣金流水';

-- ----------------------------
-- Table structure for dvadmin_writeoff_cashflow
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_writeoff_cashflow`;
CREATE TABLE `dvadmin_writeoff_cashflow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `old_money` bigint NOT NULL COMMENT '变更前余额',
  `new_money` bigint NOT NULL COMMENT '变更后余额',
  `change_money` bigint NOT NULL COMMENT '变更余额',
  `flow_type` int NOT NULL DEFAULT '0' COMMENT '流水类型',
  `tax` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '费率',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `order_id` varchar(30) DEFAULT NULL COMMENT '系统订单',
  `pay_channel_id` bigint DEFAULT NULL COMMENT '支付通道',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  PRIMARY KEY (`id`),
  KEY `dvadmin_writeoff_cashflow_creator_id_67c04f81` (`creator_id`),
  KEY `dvadmin_writeoff_cashflow_order_id_1f9f5855` (`order_id`),
  KEY `dvadmin_writeoff_cashflow_pay_channel_id_4a43e400` (`pay_channel_id`),
  KEY `dvadmin_writeoff_cashflow_writeoff_id_8defa60b` (`writeoff_id`)
) ENGINE=InnoDB AUTO_INCREMENT=184604 DEFAULT CHARSET=utf8mb3 COMMENT='核销资金流水';

-- ----------------------------
-- Table structure for dvadmin_writeoff_pay_channel
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_writeoff_pay_channel`;
CREATE TABLE `dvadmin_writeoff_pay_channel` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `tax` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '费率',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '通道状态',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `pay_channel_id` bigint NOT NULL COMMENT '绑定的支付通道',
  `writeoff_id` bigint NOT NULL COMMENT '绑定的核销',
  PRIMARY KEY (`id`),
  UNIQUE KEY `dvadmin_writeoff_pay_cha_pay_channel_id_writeoff__3f90ef15_uniq` (`pay_channel_id`,`writeoff_id`),
  KEY `dvadmin_writeoff_pay_channel_creator_id_f8ffb76a` (`creator_id`),
  KEY `dvadmin_writeoff_pay_channel_pay_channel_id_d2db6f56` (`pay_channel_id`),
  KEY `dvadmin_writeoff_pay_channel_writeoff_id_395cd7ae` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='核销支付通道';

-- ----------------------------
-- Table structure for dvadmin_writeoff_pre
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_writeoff_pre`;
CREATE TABLE `dvadmin_writeoff_pre` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `pre_pay` bigint NOT NULL DEFAULT '0' COMMENT '预付金额',
  `ver` bigint NOT NULL,
  `writeoff_id` bigint NOT NULL COMMENT '关联商户',
  PRIMARY KEY (`id`),
  UNIQUE KEY `writeoff_id` (`writeoff_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb3 COMMENT='核销预付';

-- ----------------------------
-- Table structure for dvadmin_writeoff_pre_history
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_writeoff_pre_history`;
CREATE TABLE `dvadmin_writeoff_pre_history` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `pre_pay` bigint NOT NULL DEFAULT '0' COMMENT '预付金额',
  `before` bigint NOT NULL DEFAULT '0' COMMENT '改动前金额',
  `after` bigint NOT NULL DEFAULT '0' COMMENT '改动后金额',
  `user` varchar(255) DEFAULT NULL,
  `ver` bigint NOT NULL,
  `rate` varchar(32) DEFAULT '0' COMMENT 'usdt汇率',
  `usdt` varchar(32) DEFAULT '0' COMMENT 'usdt',
  `cert` longtext COMMENT '转账凭证',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dvadmin_writeoff_pre_history_creator_id_58dbd833` (`creator_id`),
  KEY `dvadmin_writeoff_pre_history_writeoff_id_91257972` (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='核销预付历史记录';

-- ----------------------------
-- Table structure for dvadmin_writeoff_tax
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_writeoff_tax`;
CREATE TABLE `dvadmin_writeoff_tax` (
  `writeoff_id` bigint NOT NULL COMMENT '核销',
  `pre_tax` bigint NOT NULL DEFAULT '0' COMMENT '占用金额',
  `ver` bigint NOT NULL,
  PRIMARY KEY (`writeoff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='核销预占用金额';

-- ----------------------------
-- Table structure for dvadmin_yunshu_code
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_yunshu_code`;
CREATE TABLE `dvadmin_yunshu_code` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `url` longtext COMMENT 'url',
  `limit_money` bigint NOT NULL DEFAULT '0' COMMENT '限制金额',
  `money` int NOT NULL DEFAULT '0' COMMENT '金额',
  `extra` json NOT NULL DEFAULT (_utf8mb4'{}') COMMENT '额外参数',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `machine_id` bigint NOT NULL COMMENT '关联设备',
  PRIMARY KEY (`id`),
  KEY `dvadmin_yunshu_code_creator_id_58b7c2ec` (`creator_id`),
  KEY `dvadmin_yunshu_code_machine_id_fd4197eb` (`machine_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='云数码';

-- ----------------------------
-- Table structure for dvadmin_yunshu_code_day
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_yunshu_code_day`;
CREATE TABLE `dvadmin_yunshu_code_day` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功订单数',
  `submit_count` int NOT NULL DEFAULT '0' COMMENT '总提交订单数',
  `success_money` bigint NOT NULL DEFAULT '0' COMMENT '总收入',
  `date` date NOT NULL COMMENT '日期',
  `ver` bigint NOT NULL,
  `code_id` bigint DEFAULT NULL COMMENT '关联码',
  PRIMARY KEY (`id`),
  KEY `dvadmin_yunshu_code_day_code_id_b0242aab` (`code_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='云数码日统计';

-- ----------------------------
-- Table structure for dvadmin_yunshu_machine
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_yunshu_machine`;
CREATE TABLE `dvadmin_yunshu_machine` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
  `online` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否存活',
  `group_id` varchar(255) DEFAULT NULL COMMENT '群组ID',
  `group_name` varchar(255) DEFAULT NULL COMMENT '群组名称',
  `machine_id` varchar(255) NOT NULL COMMENT '机器ID',
  `machine_type` int NOT NULL DEFAULT '0' COMMENT '设备类型',
  `alipay_id` bigint DEFAULT NULL COMMENT '授权支付宝',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  `writeoff_id` bigint NOT NULL COMMENT '关联核销',
  `user_id` bigint DEFAULT NULL COMMENT '关联设备用户',
  PRIMARY KEY (`id`),
  KEY `dvadmin_yunshu_machine_alipay_id_ef51d413` (`alipay_id`),
  KEY `dvadmin_yunshu_machine_creator_id_43f74de9` (`creator_id`),
  KEY `dvadmin_yunshu_machine_writeoff_id_a0394d01` (`writeoff_id`),
  KEY `dvadmin_yunshu_machine_user_id_5e28f8b7` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='云数机器';

-- ----------------------------
-- Table structure for dvadmin_yunshu_user
-- ----------------------------
DROP TABLE IF EXISTS `dvadmin_yunshu_user`;
CREATE TABLE `dvadmin_yunshu_user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `remarks` varchar(255) DEFAULT NULL COMMENT '备注',
  `modifier` varchar(255) DEFAULT NULL COMMENT '修改人',
  `update_datetime` datetime(6) DEFAULT NULL COMMENT '修改时间',
  `create_datetime` datetime(6) DEFAULT NULL COMMENT '创建时间',
  `username` varchar(255) NOT NULL COMMENT '账号',
  `pwd` varchar(255) NOT NULL COMMENT '密码',
  `cookie` longtext COMMENT 'ck',
  `online` tinyint(1) NOT NULL DEFAULT '0' COMMENT '在线',
  `machine_type` int NOT NULL DEFAULT '0' COMMENT '设备类型',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  KEY `dvadmin_yunshu_user_creator_id_27700f39` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='云数用户';

-- ----------------------------
-- Table structure for token_blacklist_blacklistedtoken
-- ----------------------------
DROP TABLE IF EXISTS `token_blacklist_blacklistedtoken`;
CREATE TABLE `token_blacklist_blacklistedtoken` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `blacklisted_at` datetime(6) NOT NULL,
  `token_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `token_id` (`token_id`),
  CONSTRAINT `token_blacklist_blacklistedtoken_token_id_3cc7fe56_fk` FOREIGN KEY (`token_id`) REFERENCES `token_blacklist_outstandingtoken` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for token_blacklist_outstandingtoken
-- ----------------------------
DROP TABLE IF EXISTS `token_blacklist_outstandingtoken`;
CREATE TABLE `token_blacklist_outstandingtoken` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `token` longtext NOT NULL,
  `created_at` datetime(6) DEFAULT NULL,
  `expires_at` datetime(6) NOT NULL,
  `user_id` bigint DEFAULT NULL,
  `jti` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `token_blacklist_outstandingtoken_jti_hex_d9bdf6f7_uniq` (`jti`),
  KEY `token_blacklist_outs_user_id_83bc629a_fk_dvadmin_s` (`user_id`),
  CONSTRAINT `token_blacklist_outs_user_id_83bc629a_fk_dvadmin_s` FOREIGN KEY (`user_id`) REFERENCES `dvadmin_system_users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=132 DEFAULT CHARSET=utf8mb3;

SET FOREIGN_KEY_CHECKS = 1;
