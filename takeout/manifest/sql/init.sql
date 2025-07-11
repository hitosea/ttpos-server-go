CREATE DATABASE `takeout` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci */;

-- takeout.echo definition

CREATE TABLE `echo` (
                        `id` bigint(20) NOT NULL AUTO_INCREMENT,
                        `uuid` bigint(20) NOT NULL,
                        `msg` varchar(100) DEFAULT NULL,
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `echo_unique` (`uuid`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO echo
(uuid, msg)
VALUES( 999, 'test msg');