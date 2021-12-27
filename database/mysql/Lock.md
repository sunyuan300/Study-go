根据加锁的范围，MySQL 里面的锁大致可以分成 `全局锁、表级锁和行级锁`三类。在实现上分为 `服务器层`和 `存储引擎层`。

**Server层**：实现了表级锁，分别是 `表锁`和 `元数据锁`（meta data lock,MDL）。
**存储引擎层**：不同的存储引擎实现了不同的锁机制。

- MyISAM 和 MEMORY 存储引擎采用的是`表级锁`。
- InnoDB 存储引擎支持`行级锁和表级锁`，默认情况下是采用*行级锁*。

# 1. 全局锁

全局锁就是对整个 `数据库实例`加锁，让整个库处于只读状态，DDL、DML等都会失效。典型应用场景是：`做全库逻辑备份`。

全局锁加锁命令：`Flush tables with read lock `

# 2. 表级锁

## 2.1 Server层的表级锁

Server层实现的锁机制，不会影响存储引擎层的锁机制。

### 2.1.1 表锁

加锁：`lock tables … read/write`。
解锁：`unlock tables...`或客户端断开连接 `自动解锁`。

线程A执行 `lock tables t1 read,t2 write;`此时其他线程写t1、读写t2会**阻塞**；同时，**线程A也只能读t1、读写t2**，直至释放线程A释放表锁。

### 2.2.2 元数据锁MDL

MDL是MySQL5.5引入的锁，在**访问一张表**时MDL锁会被**自动加上**，直到**事务提交**才释放。

读锁：对一张表做 `增删改查`操作时，加 MDL读锁。
写锁：对表做 `结构变更`操作时，加 MDL 写锁。

**思考题**：修改表结构，为什么可能导致整个库挂调？

![image.png](https://fynotefile.oss-cn-zhangjiakou.aliyuncs.com/fynote/3005/1640445804000/f11fa341bf5546c4878cc3345e5f82e6.png)

1. 事务A会先对表T添加MDL读锁。
2. 事务B也会对表T添加MDL读锁。
3. 事务C需要对表T添加MDL写锁，但由于事务A和事务B还未释放读锁，所以会阻塞。
4. **虽然事务C处于阻塞状态，但是依然会影响后续事务**。所以事务D和其它事务都会处于阻塞状态，无论是DQL、DDL、DML，**整个表T完全不可读写**。

**如果表T的查询非常频繁，该库的线程很快就会被打爆**。

解决方案：

- 先暂停 DDL，或者 kill 掉这个长事务。
- 针对热点表的变更，kill 未必管用，因为会不断生成新的请求，所以最好在`alter table`里面设置超时时间，避免阻塞业务SQL，之后开发人员或者 DBA 再通过重试命令重复这个过程。

## 2.2 存储引擎层InnoDB的表级锁

存储引擎的表级锁，会影响存储引擎层的行级锁。

- 给表加`S锁`：
  - 别的事务可以继续获得该表的`S锁`
  - 别的事务可以继续获得该表中的某些记录的`S锁`
  - 别的事务不可以继续获得该表的`X锁`
  - 别的事务不可以继续获得该表中的某些记录的`X锁`
- 给表加`X锁`：
  - 别的事务不可以继续获得该表的`S锁`
  - 别的事务不可以继续获得该表中的某些记录的`S锁`
  - 别的事务不可以继续获得该表的`X锁`
  - 别的事务不可以继续获得该表中的某些记录的`X锁`

### 2.2.1 表级别的S锁、X锁

InnoDB存储引擎提供的表级S锁和X锁，只会在系统变量 `autocommit=0，innodb_table_locks = 1`时，才可以使用，但是并不常用。

```sql
LOCK TABLES t READ    # InnoDB存储引擎会对表t加表级别的S锁。
LOCK TABLES t WRITE   # InnoDB存储引擎会对表t加表级别的X锁。
```

### 2.2.2 表级别的IS锁、IX锁

意义：在加表级别的S锁和X锁时，判断表中是否有已经被加锁的记录。

过程：在对使⽤InnoDB存储引擎表的 `某些记录加S锁`之前，需要先在表级别加⼀个 `IS锁`；在对使⽤InnoDB存储引擎表的 `某些记录加X锁`之前，需要先在表级别加⼀个 `IX锁`。

# 3. 行级锁

InnoDB实现了以下几种类型的行级锁：

- 记录锁(Record Locks)
- Gap锁(Gap Locks)

## 3.1 记录锁(Record Locks)

仅仅给 `⼀条记录`加锁

### 3.1.1 加锁方法

- 隐式锁定：
  - 对于`增删改`语句，InnoDB存储引擎会自动给对应的记录添`X锁`。
- 显示锁定：普通的`SELECT`语句，InnoDB 不会加任何锁。

```sql
select ... lock in share mode;  # 添加S锁
select ... for update;          # 添加X锁
```

### 3.1.2 显示锁定

**select lock in share mode**
应用场景：确保查到的数据是 `最新数据`，且 `不允许其他人修改`，但 `自己不一定能修改`。

**select for update**
应用场景：确保查到的数据是 `最新数据`，且查到的数据 `只允许自己修改`。

**例子1**：

```sql
create table x(`id` int, `num` int, index `idx_id` (`id`));
insert into x values(1, 1), (2, 2);

-- 事务A
START TRANSACTION;
update x set id = 1 where id = 1;

-- 事务B
-- 如果事务A没有commit，id=1的记录拿不到X锁，将出现等待
START TRANSACTION;
update x set id = 1 where id = 1;

-- 事务C
-- id=2的记录可以拿到X锁，不会出现等待
START TRANSACTION;
update x set id = 2 where id = 2;
```

## 3.2 间隙锁(Gap Locks)

MySQL在 `REPEATABLE READ`隔离级别下是可以解决 `幻读`问题的，解决⽅案有两种：

- MVCC
- Gap锁：**隔离级别**需要达到`RR及以上`。

虽然有 `共享gap锁`和 `独占gap锁`的说法，但它们起到的**作⽤是一样的**；

### 3.2.1 Gap锁产生条件

- 使用唯一索引进行等值、范围检索
- 使用普通索引进行等值、范围检索

### 3.2.2 Gap锁是如何锁区间？

![索引树结构](https://img-blog.csdnimg.cn/201902122353375.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzIwNTk3NzI3,size_16,color_FFFFFF,t_70)
索引结构分为主索引树和辅助索引树，辅助索引树的叶子节点中包含了主键数据，主键数据影响着叶子节点的排序，Gap锁的关键就是 `锁住索引树叶子节点之间的间隙`。

#### 3.2.2.1 非唯一索引等值、范围检索Gap锁原则分析

```sql
create table t(
  letter varchar(2)  primary key,
  num int
);

create index t_num_index on t(num);

insert into t values('d',3);
insert into t values('g',6);
insert into t values('j',8);
```

![](https://img-blog.csdnimg.cn/2019021223541372.png)

**情况1**：

```sql
# session 1
select * from t where num=6 for update;
```

```sql
# session 2
insert into t values('a',3); # 插入成功
```

![在这里插入图片描述](https://img-blog.csdnimg.cn/20190212235441833.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzIwNTk3NzI3,size_16,color_FFFFFF,t_70)

**情况2**：

```sql
# session 1
select * from t where num=6 for update;
```

```sql
# session 2
insert into t values('e',3); # 阻塞
```

![在这里插入图片描述](https://img-blog.csdnimg.cn/20190212235453304.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzIwNTk3NzI3,size_16,color_FFFFFF,t_70)

**情况3**：

```sql
# session 1
select * from t where num=6 for update;
```

```sql
# session 2
insert into t values('h',6); # 阻塞
```

![在这里插入图片描述](https://img-blog.csdnimg.cn/20190212235504455.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzIwNTk3NzI3,size_16,color_FFFFFF,t_70)

**情况4**：

```sql
# session 1
select * from t where num=6 for update;
```

```sql
# session 2
insert into t values('h',9); # 插入成功
```

![在这里插入图片描述](https://img-blog.csdnimg.cn/20190212235526380.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzIwNTk3NzI3,size_16,color_FFFFFF,t_70)

**情况5**：

```sql
# session 1
select * from t where num=5 for update;
```

```sql
# session 2
insert into t values('a',3); # 插入成功
insert into t values('d',3); # 主键冲突
insert into t values('e',3); # 阻塞
insert into t values('k',5); # 阻塞
insert into t values('f',6); # 阻塞
insert into t values('g',6); # 主键冲突
insert into t values('g',9); # 主键冲突
insert into t values('h',6); # 插入成功
```

**情况6**：

```sql
# session 1
select * from t where num>5 for update;
```

```sql
# session 2
insert into t values('a',3); # 插入成功
insert into t values('d',3); # 主键冲突
insert into t values('d',4); # 主键冲突
insert into t values('d',10); # 主键冲突
insert into t values('e',3); # 阻塞
insert into t values('k',5); # 阻塞
insert into t values('f',6); # 阻塞
insert into t values('h',6); # 阻塞
insert into t values('z',9); # 阻塞
```

**情况7**：

```sql
# session 1
select * from t where num>5 and num <7 for update;
```

```sql
# session 2
insert into t values('a',3); # 插入成功
insert into t values('d',3); # 主键冲突
insert into t values('d',4); # 主键冲突
insert into t values('e',3); # 阻塞
insert into t values('k',5); # 阻塞
insert into t values('f',6); # 阻塞
insert into t values('h',6); # 阻塞
insert into t values('j',8); # 阻塞，并未出现主键冲突，表示被加Gap锁
insert into t values('j',10); # 阻塞，并未出现主键冲突，表示被加Gap锁
insert into t values('z',9); # 插入成功
```

**结论**：

- 等值检索的Gap锁范围是`离检索条件最近的两条记录的主键左开右闭区间`
- 范围检索的Gap锁范围是`离检索条件最近记录的主键左开右闭区间`

#### 3.2.2.2 唯一索引等值、范围检索Gap锁原则分析

**等值检索**

```sql
CREATE TABLE `test` (
  `id` int(1) NOT NULL AUTO_INCREMENT,
  `name` varchar(8) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `test` VALUES ('1', '小罗');
INSERT INTO `test` VALUES ('5', '小黄');
INSERT INTO `test` VALUES ('7', '小明');
INSERT INTO `test` VALUES ('11', '小红');
```

```sql
# session1
select * from `test` where`id` = 3 for update;
# 间隙区间：(1,5)
```

```sql
# session2
INSERT INTO `test` (`id`, `name`) VALUES (1, '小张1'); # 主键冲突
INSERT INTO `test` (`id`, `name`) VALUES (2, '小张2'); # 阻塞
INSERT INTO `test` (`id`, `name`) VALUES (3, '小张3'); # 阻塞
INSERT INTO `test` (`id`, `name`) VALUES (4, '小白4'); # 阻塞
INSERT INTO `test` (`id`, `name`) VALUES (5, '小白5'); # 主键冲突
INSERT INTO `test` (`id`, `name`) VALUES (6, '小东'); # 正常执行
```

**范围检索**：

```sql
create table h(
  id int  primary key,
  name varchar(11)
);

insert into h values(1,'a');
insert into h values(5,'h');
insert into h values(8,'m');
insert into h values(11,'ds');
```

**情况1**：

```sql
# session 1
select * from h where id >5 for update;
# 间隙区间(5,+&)
```

```sql
# session 2
insert into h values(4,'cc'); # 插入成功
insert into h values(5,'bb'); # 主键冲突
insert into h values(6,'cc'); # 阻塞
```

![在这里插入图片描述](https://img-blog.csdnimg.cn/20190214234535936.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzIwNTk3NzI3,size_16,color_FFFFFF,t_70)

**情况2**：

```sql
# session 1
select * from h where id >5  and id <11 for update;
# 间隙区间(5,11)
```

```sql
# session 2
insert into h values(4,'cc'); # 插入成功
insert into h values(5,'bb'); # 主键冲突
insert into h values(10,'cc'); # 阻塞
insert into h values(11,'cc'); # 主键冲突
insert into h values(12,'cc'); # 插入成功
```

**结论**：

- 等值检索的Gap锁范围是`离检索条件最近记录的开区间`。
- 范围检索的Gap锁范围是`离检索条件最近记录的开区间`，当查询条件取等时，会给对应数据加记录锁。

# 3.3 Next-key Locks

Next-key Locks = Record Locks+Gap Locks
