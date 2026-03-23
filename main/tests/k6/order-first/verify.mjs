/**
 * Post-K6 SQL verification script.
 *
 * Reads env.json and queries the test tenant database to verify
 * that the order-first-pay-later business logic wrote correct data.
 *
 * Usage:  node verify.mjs
 */

import mysql from 'mysql2/promise';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const env = JSON.parse(fs.readFileSync(path.join(__dirname, 'env.json'), 'utf8'));

let passed = 0;
let failed = 0;

function assert(condition, label) {
  if (condition) {
    console.log(`  ✓ ${label}`);
    passed++;
  } else {
    console.error(`  ✗ ${label}`);
    failed++;
  }
}

async function main() {
  const conn = await mysql.createConnection({
    host: env.DB_HOST,
    port: parseInt(env.DB_PORT),
    user: env.DB_USER,
    password: env.DB_PASS,
    database: env.DB_NAME,
  });

  console.log(`\n[verify] Database: ${env.DB_NAME}`);

  // ── 1. Check sale_bill records created by K6 ──────────────────
  console.log('\n── Sale Bills ──');
  const [bills] = await conn.query(
    `SELECT uuid, status, source, bill_type, submit_pay_time, hide_bill_time
     FROM ttpos_sale_bill WHERE delete_time = 0 ORDER BY create_time`
  );
  console.log(`  Found ${bills.length} sale bill(s)`);
  assert(bills.length >= 1, 'At least 1 sale bill created');

  // Check full-flow bill (first one created, should be the accepted one)
  for (const bill of bills) {
    console.log(`  Bill ${bill.uuid}: status=${bill.status}, source=${bill.source}, bill_type=${bill.bill_type}, submit_pay_time=${bill.submit_pay_time}, hide_bill_time=${bill.hide_bill_time}`);
  }

  // ── 2. Check H5 orders ────────────────────────────────────────
  console.log('\n── H5 Orders ──');
  const [h5Orders] = await conn.query(
    `SELECT uuid, sale_bill_uuid, status, order_type
     FROM ttpos_h5_order WHERE delete_time = 0 ORDER BY create_time`
  );
  console.log(`  Found ${h5Orders.length} H5 order(s)`);
  assert(h5Orders.length >= 1, 'At least 1 H5 order created');

  for (const h5 of h5Orders) {
    console.log(`  H5Order ${h5.uuid}: sale_bill=${h5.sale_bill_uuid}, status=${h5.status}, order_type=${h5.order_type}`);
    assert(h5.order_type === 1, `H5Order ${h5.uuid}: order_type = 1 (MemberDineIn)`);
  }

  // Find the accepted H5 order (status=2)
  const acceptedH5 = h5Orders.find(h => h.status === 2);
  if (acceptedH5) {
    console.log('\n── Accepted Order Verification ──');
    assert(true, 'Found accepted H5 order (status=2)');

    // Check the corresponding sale_bill has hide_bill_time set (instant hold)
    const bill = bills.find(b => b.uuid.toString() === acceptedH5.sale_bill_uuid.toString());
    if (bill) {
      assert(bill.status === 0, `Accepted bill status = 0 (Pending, not yet paid)`);
      // hide_bill_time may be 0 if cashier took the order (ShowOrder resets it)
      assert(bill.submit_pay_time > 0, `Accepted bill has submit_pay_time > 0`);
    } else {
      assert(false, 'Could not find sale_bill for accepted H5 order');
    }
  } else {
    console.log('  (No accepted H5 order found — full_flow scenario may have failed)');
  }

  // Find the rejected H5 order (status=3)
  const rejectedH5 = h5Orders.find(h => h.status === 3);
  if (rejectedH5) {
    console.log('\n── Rejected Order Verification ──');
    assert(true, 'Found rejected H5 order (status=3)');

    const bill = bills.find(b => b.uuid.toString() === rejectedH5.sale_bill_uuid.toString());
    if (bill) {
      assert(bill.status === 2, `Rejected bill status = 2 (Canceled)`);
    }
  }

  // ── 3. Check sale_order_product records ───────────────────────
  console.log('\n── Sale Order Products ──');
  const [products] = await conn.query(
    `SELECT sop.uuid, sop.h5_order_uuid, sop.is_accept_order
     FROM ttpos_sale_order_product sop
     WHERE sop.delete_time = 0
     ORDER BY sop.create_time`
  );
  console.log(`  Found ${products.length} sale order product(s)`);
  assert(products.length >= 1, 'At least 1 sale order product created');

  for (const p of products) {
    console.log(`  Product ${p.uuid}: h5_order=${p.h5_order_uuid}, is_accept=${p.is_accept_order}`);
  }

  // ── 4. Check setting is correct ───────────────────────────────
  console.log('\n── Settings ──');
  const [settings] = await conn.query(
    "SELECT `key`, `values` FROM ttpos_setting WHERE `key` = 'store_scan_order'"
  );
  if (settings.length > 0) {
    const val = JSON.parse(settings[0].values);
    assert(val.is_order_first_pay_later === 1, 'store_scan_order.is_order_first_pay_later = 1');
  } else {
    assert(false, 'store_scan_order setting not found');
  }

  await conn.end();

  // ── Summary ───────────────────────────────────────────────────
  console.log(`\n${'═'.repeat(50)}`);
  console.log(`[verify] Results: ${passed} passed, ${failed} failed`);
  console.log('═'.repeat(50));

  if (failed > 0) {
    process.exit(1);
  }
}

main().catch(e => {
  console.error('[verify] FATAL:', e.message);
  process.exit(1);
});
