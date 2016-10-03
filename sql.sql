/*
Navicat MySQL Data Transfer

Source Server         : mysql
Source Server Version : 50714
Source Host           : 127.0.0.1:3306
Source Database       : baidu

Target Server Type    : MYSQL
Target Server Version : 50714
File Encoding         : 65001

Date: 2016-10-03 09:42:27
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for avaiuk
-- ----------------------------
DROP TABLE IF EXISTS `avaiuk`;
CREATE TABLE `avaiuk` (
`id`  bigint(20) NOT NULL ,
`uk`  bigint(20) NULL DEFAULT NULL ,
`flag`  int(1) NULL DEFAULT 0 ,
PRIMARY KEY (`id`)
)
ENGINE=InnoDB
DEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci

;

-- ----------------------------
-- Table structure for sharedata
-- ----------------------------
DROP TABLE IF EXISTS `sharedata`;
CREATE TABLE `sharedata` (
`id`  bigint(11) NOT NULL ,
`title`  varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL ,
`shareid`  varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL ,
`uinfo_id`  bigint(11) NULL DEFAULT NULL ,
PRIMARY KEY (`id`),
FOREIGN KEY (`uinfo_id`) REFERENCES `uinfo` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
)
ENGINE=InnoDB
DEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci

;

-- ----------------------------
-- Table structure for uinfo
-- ----------------------------
DROP TABLE IF EXISTS `uinfo`;
CREATE TABLE `uinfo` (
`id`  bigint(11) NOT NULL ,
`uname`  varchar(128) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL ,
`uk`  bigint(11) NULL DEFAULT NULL ,
`avatar_url`  varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL ,
`incache`  int(1) NULL DEFAULT 0 ,
PRIMARY KEY (`id`)
)
ENGINE=InnoDB
DEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci

;

-- ----------------------------
-- Indexes structure for table sharedata
-- ----------------------------
CREATE INDEX `uinfo_id` ON `sharedata`(`uinfo_id`) USING BTREE ;

-- ----------------------------
-- Indexes structure for table uinfo
-- ----------------------------
CREATE INDEX `uk` ON `uinfo`(`uk`) USING BTREE ;
