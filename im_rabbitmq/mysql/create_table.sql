-- 创建数据库 go_postery
create database go_chatery;
create user 'go_chatery_tester' identified by '123456';
-- 将数据库 go_postery 的全部权限授予用户 go_chatery_tester
grant all on go_chatery.* to go_chatery_tester;
-- 切到 go_postery 数据库
use go_chatery;

CREATE TABLE IF NOT EXISTS users
(
    id BIGINT NOT NULL COMMENT '用户 ID (雪花算法)',
    PRIMARY KEY (id)
) DEFAULT CHARSET = utf8mb4 COMMENT '用户表';


CREATE TABLE IF NOT EXISTS messages
(
    id           BIGINT   NOT NULL COMMENT 'ID',

    message_from BIGINT   NOT NULL COMMENT '发送方',
    message_to   BIGINT   NOT NULL COMMENT '接收方',

    content      TEXT     NOT NULL COMMENT '消息内容',
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at   DATETIME          DEFAULT NULL COMMENT '逻辑删除时间',

    PRIMARY KEY (id)
) DEFAULT CHARSET = utf8mb4 COMMENT '消息记录表';