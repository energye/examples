-- 电商系统 · 示例模型 · wg/dbmodeler 生成的模拟 DDL
-- 方言: MySQL | 表: 11 | 字段: 72 | 关系: 14
-- 说明: 纯内存模型生成，未连接任何数据库

CREATE TABLE `users` (
  `id` BIGINT AUTO_INCREMENT NOT NULL COMMENT '用户 ID',
  `username` VARCHAR(64) NOT NULL COMMENT '登录名',
  `email` VARCHAR(128) COMMENT '邮箱',
  `password_hash` VARCHAR(128) NOT NULL COMMENT '密码哈希',
  `phone` VARCHAR(20) COMMENT '手机号',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1 正常 0 禁用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
  `updated_at` DATETIME COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE (`username`),
  UNIQUE (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户主表';

CREATE TABLE `addresses` (
  `id` BIGINT AUTO_INCREMENT NOT NULL COMMENT '地址 ID',
  `user_id` BIGINT NOT NULL COMMENT '所属用户',
  `receiver` VARCHAR(64) NOT NULL COMMENT '收货人',
  `phone` VARCHAR(20) NOT NULL COMMENT '联系电话',
  `province` VARCHAR(32) COMMENT '省',
  `city` VARCHAR(32) COMMENT '市',
  `detail` VARCHAR(255) COMMENT '详细地址',
  `is_default` BOOLEAN NOT NULL DEFAULT 0 COMMENT '默认地址',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='收货地址';

CREATE TABLE `categories` (
  `id` BIGINT AUTO_INCREMENT NOT NULL COMMENT '分类 ID',
  `parent_id` BIGINT COMMENT '父分类',
  `name` VARCHAR(64) NOT NULL COMMENT '分类名',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品分类（自引用）';

CREATE TABLE `products` (
  `id` BIGINT AUTO_INCREMENT NOT NULL COMMENT '商品 ID',
  `category_id` BIGINT NOT NULL COMMENT '所属分类',
  `name` VARCHAR(128) NOT NULL COMMENT '商品名',
  `sku` VARCHAR(64) NOT NULL COMMENT 'SKU 编码',
  `price` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '售价',
  `stock` INT NOT NULL DEFAULT 0 COMMENT '库存',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1 上架 0 下架',
  `description` TEXT COMMENT '详情',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上架时间',
  PRIMARY KEY (`id`),
  UNIQUE (`sku`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品 SPU';

CREATE TABLE `product_images` (
  `id` BIGINT AUTO_INCREMENT NOT NULL COMMENT '图片 ID',
  `product_id` BIGINT NOT NULL COMMENT '所属商品',
  `url` VARCHAR(255) NOT NULL COMMENT '图片地址',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品图片';

CREATE TABLE `cart_items` (
  `id` BIGINT AUTO_INCREMENT NOT NULL COMMENT '明细 ID',
  `user_id` BIGINT NOT NULL COMMENT '所属用户',
  `product_id` BIGINT NOT NULL COMMENT '商品',
  `quantity` INT NOT NULL DEFAULT 1 COMMENT '数量',
  `added_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='购物车明细';

CREATE TABLE `orders` (
  `id` BIGINT AUTO_INCREMENT NOT NULL COMMENT '订单 ID',
  `order_no` VARCHAR(32) NOT NULL COMMENT '订单号',
  `user_id` BIGINT NOT NULL COMMENT '下单用户',
  `address_id` BIGINT COMMENT '收货地址',
  `total_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '订单总额',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0 待付 1 已付 2 发货 3 完成',
  `paid_at` DATETIME COMMENT '支付时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '下单时间',
  PRIMARY KEY (`id`),
  UNIQUE (`order_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单主表';

CREATE TABLE `order_items` (
  `id` BIGINT AUTO_INCREMENT NOT NULL COMMENT '明细 ID',
  `order_id` BIGINT NOT NULL COMMENT '所属订单',
  `product_id` BIGINT NOT NULL COMMENT '商品快照',
  `product_name` VARCHAR(128) NOT NULL COMMENT '商品名快照',
  `unit_price` DECIMAL(10,2) NOT NULL COMMENT '成交单价',
  `quantity` INT NOT NULL DEFAULT 1 COMMENT '数量',
  `subtotal` DECIMAL(10,2) NOT NULL COMMENT '小计',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单明细';

CREATE TABLE `payments` (
  `id` BIGINT AUTO_INCREMENT NOT NULL COMMENT '流水 ID',
  `order_id` BIGINT NOT NULL COMMENT '关联订单',
  `channel` VARCHAR(32) NOT NULL COMMENT '支付渠道',
  `trade_no` VARCHAR(64) COMMENT '第三方交易号',
  `amount` DECIMAL(10,2) NOT NULL COMMENT '支付金额',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0 待支付 1 成功 2 失败',
  `paid_at` DATETIME COMMENT '回调时间',
  PRIMARY KEY (`id`),
  UNIQUE (`trade_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付流水';

CREATE TABLE `coupons` (
  `id` BIGINT AUTO_INCREMENT NOT NULL COMMENT '券 ID',
  `code` VARCHAR(32) NOT NULL COMMENT '券码',
  `discount` DECIMAL(10,2) NOT NULL COMMENT '面额',
  `min_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '使用门槛',
  `starts_at` DATETIME COMMENT '生效时间',
  `ends_at` DATETIME COMMENT '失效时间',
  PRIMARY KEY (`id`),
  UNIQUE (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券模板';

CREATE TABLE `user_coupons` (
  `id` BIGINT AUTO_INCREMENT NOT NULL COMMENT '记录 ID',
  `user_id` BIGINT NOT NULL COMMENT '用户',
  `coupon_id` BIGINT NOT NULL COMMENT '优惠券',
  `order_id` BIGINT COMMENT '使用的订单',
  `used_at` DATETIME COMMENT '使用时间',
  PRIMARY KEY (`id`),
  UNIQUE (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户领券记录';

-- 外键约束
ALTER TABLE `addresses` ADD CONSTRAINT `fk_addresses_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT;
ALTER TABLE `categories` ADD CONSTRAINT `fk_categories_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `categories` (`id`) ON DELETE RESTRICT;
ALTER TABLE `products` ADD CONSTRAINT `fk_products_category_id` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE RESTRICT;
ALTER TABLE `product_images` ADD CONSTRAINT `fk_product_images_product_id` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE RESTRICT;
ALTER TABLE `cart_items` ADD CONSTRAINT `fk_cart_items_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT;
ALTER TABLE `cart_items` ADD CONSTRAINT `fk_cart_items_product_id` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE RESTRICT;
ALTER TABLE `orders` ADD CONSTRAINT `fk_orders_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT;
ALTER TABLE `orders` ADD CONSTRAINT `fk_orders_address_id` FOREIGN KEY (`address_id`) REFERENCES `addresses` (`id`) ON DELETE RESTRICT;
ALTER TABLE `order_items` ADD CONSTRAINT `fk_order_items_order_id` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE RESTRICT;
ALTER TABLE `order_items` ADD CONSTRAINT `fk_order_items_product_id` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE RESTRICT;
ALTER TABLE `payments` ADD CONSTRAINT `fk_payments_order_id` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE RESTRICT;
ALTER TABLE `user_coupons` ADD CONSTRAINT `fk_user_coupons_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT;
ALTER TABLE `user_coupons` ADD CONSTRAINT `fk_user_coupons_coupon_id` FOREIGN KEY (`coupon_id`) REFERENCES `coupons` (`id`) ON DELETE RESTRICT;
ALTER TABLE `user_coupons` ADD CONSTRAINT `fk_user_coupons_order_id` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE RESTRICT;
