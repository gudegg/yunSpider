/*
Navicat MySQL Data Transfer

Source Server         : mysql
Source Server Version : 50517
Source Host           : localhost:3306
Source Database       : baidu

Target Server Type    : MYSQL
Target Server Version : 50517
File Encoding         : 65001

Date: 2016-12-11 16:54:00
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for avaiuk
-- ----------------------------
DROP TABLE IF EXISTS `avaiuk`;
CREATE TABLE `avaiuk` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uk` bigint(20) DEFAULT NULL,
  `flag` int(1) DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for sharedata
-- ----------------------------
DROP TABLE IF EXISTS `sharedata`;
CREATE TABLE `sharedata` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `title` varchar(255) DEFAULT NULL,
  `shareid` varchar(64) DEFAULT NULL,
  `uinfo_id` bigint(20) DEFAULT NULL,
  `category` char(2) DEFAULT NULL,
  `album_id` varchar(32) DEFAULT NULL,
  `feed_time` datetime DEFAULT NULL COMMENT '分享时间',
  `filecount` int(11) DEFAULT NULL COMMENT '文件总数目',
  `filesize` bigint(20) DEFAULT NULL COMMENT '文件总大小',
  `dir_cnt` smallint(5) DEFAULT NULL COMMENT '文件夹数目',
  PRIMARY KEY (`id`),
  KEY `uinfoid` (`uinfo_id`),
  CONSTRAINT `uinfoid` FOREIGN KEY (`uinfo_id`) REFERENCES `uinfo` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for uinfo
-- ----------------------------
DROP TABLE IF EXISTS `uinfo`;
CREATE TABLE `uinfo` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uk` bigint(20) DEFAULT NULL,
  `uname` varchar(255) DEFAULT NULL,
  `avatar_url` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `uk` (`uk`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
