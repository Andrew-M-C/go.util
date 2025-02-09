
CREATE TABLE IF NOT EXISTS `%s` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
    `name` varchar(64) NOT NULL COMMENT '当前节点名称',
    `code` char(12) NOT NULL COMMENT '完整节点代码 (不包含后续的全 0)',
    `province` char(2) NOT NULL COMMENT '省级代码',
    `city` char(2) NOT NULL DEFAULT '00' COMMENT '市级代码',
    `county` char(2) NOT NULL DEFAULT '00' COMMENT '区县级代码',
    `town` char(3) NOT NULL DEFAULT '000' COMMENT '镇街道级代码',
    `village` char(3) NOT NULL DEFAULT '000' COMMENT '村居委会级代码',
    `create_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_at_ms` bigint(11) NOT NULL DEFAULT 0 COMMENT '删除时间戳, 毫秒',
    PRIMARY KEY (`id`),
    UNIQUE KEY `u_codes` (`province`, `city`, `county`, `town`, `village`),
    KEY `i_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='中国统计用行政区划';
