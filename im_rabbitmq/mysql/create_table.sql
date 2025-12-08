-- 创建数据库 go_postery
create database go_chatery;
create user 'go_chatery_tester' identified by '123456';
-- 将数据库 go_postery 的全部权限授予用户 go_chatery_tester
grant all on go_chatery.* to go_chatery_tester;
-- 切到 go_postery 数据库
use go_chatery;

create table if not exists user
(
    id int auto_increment comment '用户 id, 自增',
    primary key (id)
) default charset = utf8mb4 comment '用户信息表';


create table if not exists message
(
    id     bigint,
    time   bigint,
    `from` int,
    `to`   int,
    content varchar(100),
    primary key (id)
) default charset = utf8mb4 comment '消息表';
