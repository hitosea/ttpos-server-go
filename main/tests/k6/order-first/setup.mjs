/**
 * Setup script for order-first-pay-later K6 tests.
 *
 * Creates a test tenant database (cloned from shop_template), seeds all required
 * data, generates JWT tokens, and writes a JSON env file for K6 to consume.
 *
 * Usage:  node setup.mjs
 * Output: ./env.json
 */

import mysql from 'mysql2/promise';
import crypto from 'crypto';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

// ── Config ──────────────────────────────────────────────────────────
const DB_HOST     = process.env.DB_HOST     || '10.180.10.10';
const DB_PORT     = parseInt(process.env.DB_PORT || '13306');
const DB_USER     = process.env.DB_USERNAME || 'saas';
const DB_PASS     = process.env.DB_PASSWORD || '11ca2c16594c7878';
const DB_ROOT_PWD = process.env.DB_ROOT_PASSWORD || '8e5ed3c0a7368953';
const JWT_SECRET  = process.env.JWT_SECRET  || 'dkjhd00a08';
const SERVICE_URL = process.env.SERVICE_URL || 'http://localhost:8080';

// ── Snowflake-like ID generator ─────────────────────────────────────
// Use smaller IDs that fit in Number.MAX_SAFE_INTEGER (2^53) to avoid
// precision loss when converting to JSON numbers in JWT payloads.
let counter = 0;
function snowflakeId() {
  // Shift by 10 bits instead of 16 to stay within safe integer range
  const id = Math.floor(Date.now() / 1000) * 1024 + (++counter & 0x3FF);
  return id;
}

// ── JWT token generator (HS256) ─────────────────────────────────────
function base64url(buf) {
  return buf.toString('base64').replace(/=/g, '').replace(/\+/g, '-').replace(/\//g, '_');
}

function signJWT(payload, secret) {
  const header = { alg: 'HS256', typ: 'JWT' };
  const segments = [
    base64url(Buffer.from(JSON.stringify(header))),
    base64url(Buffer.from(JSON.stringify(payload))),
  ];
  const sig = crypto.createHmac('sha256', secret).update(segments.join('.')).digest();
  segments.push(base64url(sig));
  return segments.join('.');
}

function generateMemberToken(companyUUID, memberUUID) {
  const now = Math.floor(Date.now() / 1000);
  return signJWT({
    source: 'member',
    company_uuid: companyUUID,
    staff_uuid: 0,
    member_uuid: memberUUID,
    device_id: '',
    is_refresh_token: false,
    iss: 'ttpos-test',
    iat: now,
    exp: now + 86400,
  }, JWT_SECRET);
}

function generateCashierToken(companyUUID, staffUUID, deviceId) {
  const now = Math.floor(Date.now() / 1000);
  return signJWT({
    source: 'cashier',
    company_uuid: companyUUID,
    staff_uuid: staffUUID,
    member_uuid: 0,
    device_id: deviceId,
    is_refresh_token: false,
    iss: 'ttpos-test',
    iat: now,
    exp: now + 86400,
  }, JWT_SECRET);
}

// ── Main setup ──────────────────────────────────────────────────────
async function main() {
  const companyUUID = snowflakeId();
  const dbName = `shop${companyUUID}`;
  const now = Math.floor(Date.now() / 1000);

  console.log(`[setup] Creating tenant: ${dbName}`);

  // Connect as root to create database
  const rootConn = await mysql.createConnection({
    host: DB_HOST, port: DB_PORT, user: 'root', password: DB_ROOT_PWD,
    multipleStatements: true,
  });

  // Clone from shop_template
  await rootConn.execute(`CREATE DATABASE IF NOT EXISTS \`${dbName}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`);
  await rootConn.execute(`GRANT ALL PRIVILEGES ON \`${dbName}\`.* TO '${DB_USER}'@'%'`);

  const [tables] = await rootConn.query(
    "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA='shop_template' AND TABLE_TYPE='BASE TABLE'"
  );

  const ddl = tables.map(r =>
    `CREATE TABLE IF NOT EXISTS \`${dbName}\`.\`${r.TABLE_NAME}\` LIKE \`shop_template\`.\`${r.TABLE_NAME}\``
  ).join(';\n') + ';';

  await rootConn.query(ddl);
  await rootConn.end();
  console.log(`[setup] Cloned ${tables.length} tables from shop_template`);

  // Connect to tenant DB
  const conn = await mysql.createConnection({
    host: DB_HOST, port: DB_PORT, user: DB_USER, password: DB_PASS,
    database: dbName,
  });

  // ── Apply missing migrations ────────────────────────────────────
  // submit_pay_time may be missing from shop_template if migration hasn't been applied
  const [cols] = await conn.query('SHOW COLUMNS FROM ttpos_sale_bill');
  const colNames = cols.map(c => c.Field);
  if (!colNames.includes('submit_pay_time')) {
    await conn.execute(
      "ALTER TABLE ttpos_sale_bill ADD COLUMN `submit_pay_time` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '提交支付时间戳'"
    );
    console.log('[setup] Added missing column: ttpos_sale_bill.submit_pay_time');
  }

  // ── Seed data ───────────────────────────────────────────────────
  const memberUUID = snowflakeId();
  const staffUUID  = snowflakeId();
  const deviceUUID = snowflakeId();
  const deviceId   = `k6-test-device-${Date.now()}`;
  const pkgUUID    = snowflakeId();
  const flavorUUID = snowflakeId();
  const bomUUID    = snowflakeId();

  // Company
  await conn.execute(
    `INSERT INTO ttpos_company (uuid, name, status, expire_time, is_enable_erp, create_time, update_time, delete_time)
     VALUES (?, 'K6 Test Company', 1, 0, 0, ?, ?, 0)`,
    [companyUUID, now, now]
  );

  // Company setting (enable h5_order for cashier H5 order list)
  await conn.execute(
    `INSERT INTO ttpos_company_setting (uuid, company_uuid, is_open_h5_order, is_open_h5, erpnext_site_code, erpnext_pos_profile_name, erpnext_admin_email, erpnext_company_abbr, erpnext_headquarter_abbr, headquarter_uuid, erpnext_branch_name, enable_data_management, create_time, update_time, delete_time)
     VALUES (?, ?, 1, 1, '', '', '', '', '', 0, '', 0, ?, ?, 0)`,
    [snowflakeId(), companyUUID, now, now]
  );

  // Member
  await conn.execute(
    `INSERT INTO ttpos_member (uuid, nickname, phone, is_visitor, create_time, update_time, delete_time)
     VALUES (?, 'K6 Test Member', '0899999999', 0, ?, ?, 0)`,
    [memberUUID, now, now]
  );

  // Staff (super admin cashier)
  await conn.execute(
    `INSERT INTO ttpos_staff (uuid, company_uuid, real_name, username, password, phone, is_super, duty_no, bind_key, create_time, update_time)
     VALUES (?, ?, 'K6 Cashier', 'k6_cashier', '$2a$10$dummy', '0811111111', 1, 'shift-k6', '', ?, ?)`,
    [staffUUID, companyUUID, now, now]
  );

  // Device (cashier)
  await conn.execute(
    `INSERT INTO ttpos_device (uuid, source, device_id, create_time, update_time, delete_time)
     VALUES (?, 'cashier', ?, ?, ?, 0)`,
    [deviceUUID, deviceId, now, now]
  );

  // Product package
  await conn.execute(
    `INSERT INTO ttpos_product_package (uuid, name, multi_language_name_uuid, status, product_type, create_time, update_time, delete_time)
     VALUES (?, '{"zh_name":"K6测试鸡排","en_name":"K6 Test Chicken"}', 0, 1, 0, ?, ?, 0)`,
    [pkgUUID, now, now]
  );

  // Product flavor
  await conn.execute(
    `INSERT INTO ttpos_product_flavor (uuid, name, multi_language_name_uuid, create_time, update_time, delete_time)
     VALUES (?, '{"zh_name":"默认规格","en_name":"Default"}', 0, ?, ?, 0)`,
    [flavorUUID, now, now]
  );

  // Product BOM (links flavor + package with price)
  await conn.execute(
    `INSERT INTO ttpos_product_bom (uuid, product_flavor_uuid, product_package_uuid, price, status, create_time, update_time, delete_time)
     VALUES (?, ?, ?, 25.00, 1, ?, ?, 0)`,
    [bomUUID, flavorUUID, pkgUUID, now, now]
  );

  // Second product for cashier add-product test
  const pkg2UUID    = snowflakeId();
  const flavor2UUID = snowflakeId();
  const bom2UUID    = snowflakeId();

  await conn.execute(
    `INSERT INTO ttpos_product_package (uuid, name, multi_language_name_uuid, status, product_type, create_time, update_time, delete_time)
     VALUES (?, '{"zh_name":"K6可乐","en_name":"K6 Cola"}', 0, 1, 0, ?, ?, 0)`,
    [pkg2UUID, now, now]
  );
  await conn.execute(
    `INSERT INTO ttpos_product_flavor (uuid, name, multi_language_name_uuid, create_time, update_time, delete_time)
     VALUES (?, '{"zh_name":"默认规格","en_name":"Default"}', 0, ?, ?, 0)`,
    [flavor2UUID, now, now]
  );
  await conn.execute(
    `INSERT INTO ttpos_product_bom (uuid, product_flavor_uuid, product_package_uuid, price, status, create_time, update_time, delete_time)
     VALUES (?, ?, ?, 10.00, 1, ?, ?, 0)`,
    [bom2UUID, flavor2UUID, pkg2UUID, now, now]
  );

  // Settings: business hours (all day)
  await conn.execute(
    "INSERT INTO ttpos_setting (`key`, `describe`, `values`, create_time, update_time, delete_time) VALUES ('business', 'business', ?, ?, ?, 0)",
    ['{"opening_hours":"00:00-23:59"}', now, now]
  );

  // Settings: enable order-first-pay-later
  await conn.execute(
    "INSERT INTO ttpos_setting (`key`, `describe`, `values`, create_time, update_time, delete_time) VALUES ('store_scan_order', 'store_scan_order', ?, ?, ?, 0)",
    ['{"is_order_first_pay_later":1}', now, now]
  );

  await conn.end();
  console.log('[setup] Seeded all test data');

  // ── Generate tokens ─────────────────────────────────────────────
  const memberToken  = generateMemberToken(companyUUID, memberUUID);
  const cashierToken = generateCashierToken(companyUUID, staffUUID, deviceId);

  // ── Write env.json ──────────────────────────────────────────────
  const env = {
    SERVICE_URL,
    DB_NAME: dbName,
    COMPANY_UUID: companyUUID,
    MEMBER_UUID: memberUUID,
    STAFF_UUID: staffUUID,
    DEVICE_ID: deviceId,
    BOM_UUID: bomUUID,
    BOM2_UUID: bom2UUID,
    MEMBER_TOKEN: memberToken,
    CASHIER_TOKEN: cashierToken,
    // DB connection for verify script
    DB_HOST, DB_PORT, DB_USER, DB_PASS,
  };

  const envPath = path.join(__dirname, 'env.json');
  fs.writeFileSync(envPath, JSON.stringify(env, null, 2));
  console.log(`[setup] Wrote ${envPath}`);
  console.log('[setup] Done. Ready for K6.');
}

main().catch(e => {
  console.error('[setup] FATAL:', e.message);
  process.exit(1);
});
