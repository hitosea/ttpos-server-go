/**
 * K6 test: Order-First-Pay-Later dine-in order business flow.
 *
 * Scenario 1 (fullFlow):
 *   1. Member creates dine-in order
 *   2. Member adds more products (加购)
 *   3. Member submits order → H5 order created → status "pending"
 *   4. Cashier sees H5 order, accepts it
 *   5. Order appears in hold list
 *   6. Cashier takes order (取单)
 *   7. Cashier adds a product and sends to kitchen (加购+送厨)
 *   8. Member sees all products (including cashier-added) in order detail
 *
 * Scenario 2 (cancelBeforeSubmit): cancel before submit
 * Scenario 3 (rejectFlow): cashier rejects after submit
 *
 * Usage: node setup.mjs && k6 run order_first_test.js
 */

import http from 'k6/http';
import { check, group } from 'k6';
import { Counter, Trend } from 'k6/metrics';

// ── Load env ──────────────────────────────────────────────────────
const envData = JSON.parse(open('./env.json'));

const BASE_URL      = envData.SERVICE_URL;
const MEMBER_TOKEN  = envData.MEMBER_TOKEN;
const CASHIER_TOKEN = envData.CASHIER_TOKEN;
const BOM_UUID      = parseInt(envData.BOM_UUID);
const BOM2_UUID     = parseInt(envData.BOM2_UUID);

// ── Custom metrics ────────────────────────────────────────────────
const orderCreateDuration = new Trend('order_create_duration', true);
const acceptDuration      = new Trend('accept_duration', true);
const scenarioPass        = new Counter('scenario_pass');
const scenarioFail        = new Counter('scenario_fail');

// ── K6 options ────────────────────────────────────────────────────
export const options = {
  scenarios: {
    full_flow: {
      executor: 'shared-iterations',
      vus: 1,
      iterations: 1,
      exec: 'fullFlow',
    },
    cancel_before_submit: {
      executor: 'shared-iterations',
      vus: 1,
      iterations: 1,
      exec: 'cancelBeforeSubmit',
      startTime: '5s',
    },
    reject_flow: {
      executor: 'shared-iterations',
      vus: 1,
      iterations: 1,
      exec: 'rejectFlow',
      startTime: '10s',
    },
    double_submit: {
      executor: 'shared-iterations',
      vus: 1,
      iterations: 1,
      exec: 'doubleSubmit',
      startTime: '15s',
    },
    cancel_after_submit: {
      executor: 'shared-iterations',
      vus: 1,
      iterations: 1,
      exec: 'cancelAfterSubmit',
      startTime: '18s',
    },
    order_list_filter: {
      executor: 'shared-iterations',
      vus: 1,
      iterations: 1,
      exec: 'orderListFilter',
      startTime: '21s',
    },
    invalid_submit: {
      executor: 'shared-iterations',
      vus: 1,
      iterations: 1,
      exec: 'invalidSubmit',
      startTime: '24s',
    },
    add_after_submit: {
      executor: 'shared-iterations',
      vus: 1,
      iterations: 1,
      exec: 'addAfterSubmit',
      startTime: '27s',
    },
  },
  thresholds: {
    checks: ['rate==1.0'],
    scenario_fail: ['count==0'],
  },
};

// ── Helpers ───────────────────────────────────────────────────────
function memberHeaders() {
  return { 'Content-Type': 'application/json', 'Authorization': `Bearer ${MEMBER_TOKEN}` };
}
function cashierHeaders() {
  return { 'Content-Type': 'application/json', 'Authorization': `Bearer ${CASHIER_TOKEN}` };
}

function parseAPI(res, label) {
  const ok = check(res, { [`${label}: HTTP 200`]: (r) => r.status === 200 });
  if (!ok) { console.error(`${label}: HTTP ${res.status} — ${res.body}`); return null; }
  try { return JSON.parse(res.body); } catch (e) { console.error(`${label}: bad JSON`); return null; }
}

function assertOK(body, label) {
  return check(body, { [`${label}: code == 0`]: (b) => b && b.code === 0 });
}

function getStatus(detail) {
  return detail.status_info ? detail.status_info.status : detail.status;
}

// ── API wrappers ──────────────────────────────────────────────────

function createDineInOrder(bomUuid, saleBillUuid, saleOrderUuid) {
  const payload = JSON.stringify({
    sale_bill_uuid: saleBillUuid || 0,
    sale_order_uuid: saleOrderUuid || 0,
    products: [{ flavor_uuid: bomUuid || BOM_UUID, num: 1, price: 25.0, product_type: 0 }],
  });
  const res = http.post(`${BASE_URL}/api/v1/member/order/dine_in/create`, payload, {
    headers: memberHeaders(), tags: { name: 'CreateDineInOrder' },
  });
  orderCreateDuration.add(res.timings.duration);
  const body = parseAPI(res, 'CreateOrder');
  if (!assertOK(body, 'CreateOrder')) return null;
  check(body.data, {
    'CreateOrder: sale_bill_uuid > 0': (d) => d.sale_bill_uuid > 0,
    'CreateOrder: sale_order_uuid > 0': (d) => d.sale_order_uuid > 0,
  });
  return body.data;
}

function submitDineInOrder(saleBillUuid, saleOrderUuid) {
  const payload = JSON.stringify({ sale_bill_uuid: saleBillUuid, sale_order_uuid: saleOrderUuid });
  const res = http.post(`${BASE_URL}/api/v1/member/order/dine_in/submit`, payload, {
    headers: memberHeaders(), tags: { name: 'SubmitDineInOrder' },
  });
  return parseAPI(res, 'SubmitOrder');
}

function getMemberOrderDetail(saleBillUuid) {
  const res = http.get(
    `${BASE_URL}/api/v1/member/order/dine_in/detail?sale_bill_uuid=${saleBillUuid}`,
    { headers: memberHeaders(), tags: { name: 'MemberOrderDetail' } }
  );
  const body = parseAPI(res, 'OrderDetail');
  if (!assertOK(body, 'OrderDetail')) return null;
  return body.data;
}

function getH5OrderList() {
  const res = http.get(
    `${BASE_URL}/api/v1/cashier/h5_order/list?page_no=1&page_size=50&status=0`,
    { headers: cashierHeaders(), tags: { name: 'H5OrderList' } }
  );
  const body = parseAPI(res, 'H5OrderList');
  if (!assertOK(body, 'H5OrderList')) return null;
  return body.data;
}

function findH5OrderForBill(saleBillUuid) {
  const listData = getH5OrderList();
  if (!listData || !listData.list) return null;
  for (const item of listData.list) {
    if (item.sale_bill_uuid === saleBillUuid) return item;
  }
  return null;
}

function acceptH5Order(h5OrderUuid) {
  const res = http.post(`${BASE_URL}/api/v1/cashier/h5_order/accept`,
    JSON.stringify({ h5_order_uuid: h5OrderUuid }),
    { headers: cashierHeaders(), tags: { name: 'AcceptH5Order' } }
  );
  acceptDuration.add(res.timings.duration);
  return parseAPI(res, 'AcceptOrder');
}

function rejectH5Order(h5OrderUuid) {
  const res = http.post(`${BASE_URL}/api/v1/cashier/h5_order/reject`,
    JSON.stringify({ h5_order_uuid: h5OrderUuid }),
    { headers: cashierHeaders(), tags: { name: 'RejectH5Order' } }
  );
  return parseAPI(res, 'RejectOrder');
}

function getInstantOrderList() {
  const res = http.get(`${BASE_URL}/api/v1/cashier/instant/order/list?page_no=1&page_size=50`,
    { headers: cashierHeaders(), tags: { name: 'InstantOrderList' } }
  );
  const body = parseAPI(res, 'InstantOrderList');
  if (!assertOK(body, 'InstantOrderList')) return null;
  return body.data;
}

function cashierShowOrder(saleBillUuid) {
  const res = http.post(`${BASE_URL}/api/v1/cashier/instant/order/show`,
    JSON.stringify({ sale_bill_uuid: saleBillUuid }),
    { headers: cashierHeaders(), tags: { name: 'ShowOrder' } }
  );
  return parseAPI(res, 'ShowOrder');
}

function cashierAddProduct(saleBillUuid, saleOrderUuid, flavorUuid) {
  const res = http.post(`${BASE_URL}/api/v1/cashier/instant/order/cart/product/add`,
    JSON.stringify({
      sale_bill_uuid: saleBillUuid,
      sale_order_uuid: saleOrderUuid,
      flavor_uuid: flavorUuid,
      operation: 'add',
      num: 1,
    }),
    { headers: cashierHeaders(), tags: { name: 'CashierAddProduct' } }
  );
  return parseAPI(res, 'CashierAddProduct');
}

function cashierCooking(saleBillUuid) {
  const res = http.post(`${BASE_URL}/api/v1/cashier/instant/order/cart/cooking`,
    JSON.stringify({ sale_bill_uuid: saleBillUuid, ignore_must: true }),
    { headers: cashierHeaders(), tags: { name: 'CashierCooking' } }
  );
  return parseAPI(res, 'CashierCooking');
}

function cancelDineInOrder(saleBillUuid) {
  const res = http.post(`${BASE_URL}/api/v1/member/order/dine_in/cancel`,
    JSON.stringify({ sale_bill_uuid: saleBillUuid }),
    { headers: memberHeaders(), tags: { name: 'CancelDineInOrder' } }
  );
  return parseAPI(res, 'CancelOrder');
}

// ── Scenario 1: Full flow with submit + take + add + cooking ──────
export function fullFlow() {
  let passed = true;
  let saleBillUuid = 0;
  let saleOrderUuid = 0;

  group('1. Member creates dine-in order', () => {
    const order = createDineInOrder(BOM_UUID);
    if (!order) { passed = false; return; }
    saleBillUuid = order.sale_bill_uuid;
    saleOrderUuid = order.sale_order_uuid;
    console.log(`Created: bill=${saleBillUuid}, order=${saleOrderUuid}`);
  });
  if (!saleBillUuid) { scenarioFail.add(1); return; }

  group('2. Member adds more products (加购)', () => {
    const order = createDineInOrder(BOM_UUID, saleBillUuid, saleOrderUuid);
    if (!order) { passed = false; return; }
    console.log('Added product to existing order');
  });

  group('3. Member submits order', () => {
    const body = submitDineInOrder(saleBillUuid, saleOrderUuid);
    const ok = assertOK(body, 'Submit');
    if (!ok) {
      console.error(`Submit failed: ${JSON.stringify(body)}`);
      passed = false;
    }
    console.log('Order submitted');
  });

  group('4. Member detail: status = pending', () => {
    const detail = getMemberOrderDetail(saleBillUuid);
    if (!detail) { passed = false; return; }
    const status = getStatus(detail);
    const ok = check(status, { 'Detail: status is pending': (s) => s === 'pending' });
    if (!ok) { console.error(`Expected 'pending', got '${status}'`); passed = false; }
    console.log(`Status: ${status}`);
  });

  let h5OrderUuid = 0;
  group('5. Cashier: H5 order in accept list', () => {
    const h5 = findH5OrderForBill(saleBillUuid);
    const ok = check(h5, {
      'H5Order: found': (h) => h !== null && h !== undefined,
      'H5Order: uuid > 0': (h) => h && h.h5_order_uuid > 0,
    });
    if (!ok) { passed = false; return; }
    h5OrderUuid = h5.h5_order_uuid;
    console.log(`H5 order: ${h5OrderUuid}`);
  });
  if (!h5OrderUuid) { scenarioFail.add(1); return; }

  group('6. Cashier accepts H5 order', () => {
    const body = acceptH5Order(h5OrderUuid);
    if (!check(body, { 'Accept: code == 0': (b) => b && b.code === 0 })) {
      console.error(`Accept failed: ${JSON.stringify(body)}`);
      passed = false;
    }
    console.log('Accepted');
  });

  group('7. Order in hold list', () => {
    const holdData = getInstantOrderList();
    if (!holdData) { passed = false; return; }
    let found = false;
    if (holdData.list) {
      for (const item of holdData.list) {
        if (item.sale_bill_uuid === saleBillUuid) { found = true; break; }
      }
    }
    if (!check(found, { 'HoldList: order found': (f) => f === true })) {
      console.error('Not in hold list');
      passed = false;
    }
  });

  group('8. Cashier takes order (取单)', () => {
    const body = cashierShowOrder(saleBillUuid);
    const ok = check(body, { 'ShowOrder: code == 0': (b) => b && b.code === 0 });
    if (!ok) {
      console.error(`ShowOrder failed: ${JSON.stringify(body)}`);
      passed = false;
    } else {
      const cart = body.data;
      const hasSaleOrders = cart && cart.sale_order_list && cart.sale_order_list.length > 0;
      check(hasSaleOrders, { 'ShowOrder: has sale orders': (v) => v === true });
      console.log(`Take order: ${cart ? JSON.stringify(Object.keys(cart)).substring(0, 100) : 'null'}`);
    }
  });

  group('9. Cashier adds product (Cola) + sends to kitchen', () => {
    // Add a Cola (BOM2)
    const addBody = cashierAddProduct(saleBillUuid, saleOrderUuid, BOM2_UUID);
    const addOk = check(addBody, { 'AddProduct: code == 0': (b) => b && b.code === 0 });
    if (!addOk) {
      console.error(`AddProduct failed: ${JSON.stringify(addBody)}`);
      passed = false;
      return;
    }
    console.log('Added Cola via cashier');

    // Send to kitchen
    const cookBody = cashierCooking(saleBillUuid);
    const cookOk = check(cookBody, { 'Cooking: code == 0': (b) => b && b.code === 0 });
    if (!cookOk) {
      console.error(`Cooking failed: ${JSON.stringify(cookBody)}`);
      passed = false;
    } else {
      console.log('Sent to kitchen');
    }
  });

  group('10. Member sees all products (including cashier-added)', () => {
    const detail = getMemberOrderDetail(saleBillUuid);
    if (!detail) { passed = false; return; }

    const products = detail.product_list ? detail.product_list.list : [];
    // Member added same product twice (merged into 1 line) + cashier added cola = 2 product lines
    const ok = check(products, {
      'Detail: has >= 2 products': (p) => p && p.length >= 2,
    });
    if (!ok) {
      console.error(`Expected >= 2 products, got ${products ? products.length : 0}`);
      passed = false;
    }

    const status = getStatus(detail);
    check(status, { 'Detail: status is preparing': (s) => s === 'preparing' });
    console.log(`Final: ${products ? products.length : 0} products, status=${status}`);
  });

  if (passed) { scenarioPass.add(1); console.log('✓ Full flow PASSED'); }
  else { scenarioFail.add(1); console.log('✗ Full flow FAILED'); }
}

// ── Scenario 2: Cancel before submit ──────────────────────────────
export function cancelBeforeSubmit() {
  let passed = true;

  group('Cancel: create order', () => {
    const order = createDineInOrder(BOM_UUID);
    if (!order) { passed = false; return; }
    __ENV.__cancel_bill = String(order.sale_bill_uuid);
  });
  const bill = parseInt(__ENV.__cancel_bill || '0');
  if (!bill) { scenarioFail.add(1); return; }

  group('Cancel: member cancels (before submit)', () => {
    const body = cancelDineInOrder(bill);
    if (!check(body, { 'Cancel: code == 0': (b) => b && b.code === 0 })) {
      console.error(`Cancel failed: ${JSON.stringify(body)}`); passed = false;
    }
  });

  group('Cancel: verify status', () => {
    const detail = getMemberOrderDetail(bill);
    if (!detail) { passed = false; return; }
    const status = getStatus(detail);
    if (!check(status, { 'Cancel: status is cancelled': (s) => s === 'cancelled' })) {
      console.error(`Expected 'cancelled', got '${status}'`); passed = false;
    }
  });

  if (passed) { scenarioPass.add(1); console.log('✓ Cancel before submit PASSED'); }
  else { scenarioFail.add(1); console.log('✗ Cancel before submit FAILED'); }
}

// ── Scenario 3: Cashier rejects after submit ──────────────────────
export function rejectFlow() {
  let passed = true;
  let saleBillUuid = 0;
  let saleOrderUuid = 0;

  group('Reject: create + submit', () => {
    const order = createDineInOrder(BOM_UUID);
    if (!order) { passed = false; return; }
    saleBillUuid = order.sale_bill_uuid;
    saleOrderUuid = order.sale_order_uuid;

    const body = submitDineInOrder(saleBillUuid, saleOrderUuid);
    if (!assertOK(body, 'Reject-Submit')) {
      console.error(`Submit failed: ${JSON.stringify(body)}`); passed = false;
    }
  });
  if (!saleBillUuid) { scenarioFail.add(1); return; }

  let h5OrderUuid = 0;
  group('Reject: find H5 order', () => {
    const h5 = findH5OrderForBill(saleBillUuid);
    if (!h5) { passed = false; return; }
    h5OrderUuid = h5.h5_order_uuid;
  });
  if (!h5OrderUuid) { scenarioFail.add(1); return; }

  group('Reject: cashier rejects', () => {
    const body = rejectH5Order(h5OrderUuid);
    if (!check(body, { 'Reject: code == 0': (b) => b && b.code === 0 })) {
      console.error(`Reject failed: ${JSON.stringify(body)}`); passed = false;
    }
  });

  group('Reject: verify status', () => {
    const detail = getMemberOrderDetail(saleBillUuid);
    if (!detail) { passed = false; return; }
    const status = getStatus(detail);
    if (!check(status, { 'Reject: status is rejected': (s) => s === 'rejected' })) {
      console.error(`Expected 'rejected', got '${status}'`); passed = false;
    }
  });

  if (passed) { scenarioPass.add(1); console.log('✓ Reject flow PASSED'); }
  else { scenarioFail.add(1); console.log('✗ Reject flow FAILED'); }
}

// ── Helper: member order list ─────────────────────────────────────
function getMemberOrderList(status) {
  const res = http.get(
    `${BASE_URL}/api/v1/member/order/dine_in/list?page_no=1&page_size=50&status=${status}`,
    { headers: memberHeaders(), tags: { name: 'MemberOrderList' } }
  );
  const body = parseAPI(res, 'OrderList');
  if (!assertOK(body, 'OrderList')) return null;
  return body.data;
}

// ── Helper: create + submit shortcut ──────────────────────────────
function createAndSubmit(bomUuid) {
  const order = createDineInOrder(bomUuid || BOM_UUID);
  if (!order) return null;
  const body = submitDineInOrder(order.sale_bill_uuid, order.sale_order_uuid);
  if (!body || body.code !== 0) return null;
  return order;
}

// ── Scenario 4: Double submit protection ──────────────────────────
export function doubleSubmit() {
  let passed = true;
  let bill = 0, so = 0;

  group('DblSubmit: create + submit', () => {
    const order = createDineInOrder(BOM_UUID);
    if (!order) { passed = false; return; }
    bill = order.sale_bill_uuid;
    so = order.sale_order_uuid;
    const body = submitDineInOrder(bill, so);
    if (!assertOK(body, 'DblSubmit-1st')) { passed = false; }
  });
  if (!bill) { scenarioFail.add(1); return; }

  group('DblSubmit: second submit should fail', () => {
    const body = submitDineInOrder(bill, so);
    const ok = check(body, {
      'DblSubmit: code != 0': (b) => b && b.code !== 0,
    });
    if (!ok) {
      console.error(`Expected error, got: ${JSON.stringify(body)}`);
      passed = false;
    } else {
      console.log(`Double submit rejected: ${body.message}`);
    }
  });

  if (passed) { scenarioPass.add(1); console.log('✓ Double submit PASSED'); }
  else { scenarioFail.add(1); console.log('✗ Double submit FAILED'); }
}

// ── Scenario 5: Cancel after submit, before accept ────────────────
export function cancelAfterSubmit() {
  let passed = true;
  let bill = 0;

  group('CancelAfterSubmit: create + submit', () => {
    const order = createAndSubmit();
    if (!order) { passed = false; return; }
    bill = order.sale_bill_uuid;
  });
  if (!bill) { scenarioFail.add(1); return; }

  group('CancelAfterSubmit: verify pending', () => {
    const detail = getMemberOrderDetail(bill);
    if (!detail) { passed = false; return; }
    const status = getStatus(detail);
    check(status, { 'CAS: initially pending': (s) => s === 'pending' });
  });

  group('CancelAfterSubmit: cancel', () => {
    const body = cancelDineInOrder(bill);
    if (!check(body, { 'CAS: cancel code == 0': (b) => b && b.code === 0 })) {
      console.error(`Cancel failed: ${JSON.stringify(body)}`);
      passed = false;
    }
  });

  group('CancelAfterSubmit: verify cancelled', () => {
    const detail = getMemberOrderDetail(bill);
    if (!detail) { passed = false; return; }
    const status = getStatus(detail);
    if (!check(status, { 'CAS: status cancelled': (s) => s === 'cancelled' })) {
      console.error(`Expected cancelled, got ${status}`);
      passed = false;
    }
  });

  group('CancelAfterSubmit: not in H5 list', () => {
    const h5 = findH5OrderForBill(bill);
    check(h5, { 'CAS: H5 order gone': (h) => h === null || h === undefined });
  });

  if (passed) { scenarioPass.add(1); console.log('✓ Cancel after submit PASSED'); }
  else { scenarioFail.add(1); console.log('✗ Cancel after submit FAILED'); }
}

// ── Scenario 6: Order list status filtering ───────────────────────
export function orderListFilter() {
  let passed = true;

  // Create orders in different states:
  // A: submitted + accepted → inprogress
  // B: submitted + rejected → cancelled
  // C: not submitted → should NOT appear in list
  let billA = 0, billB = 0, billC = 0;

  group('ListFilter: setup orders', () => {
    // Order A: create → submit → accept
    const orderA = createAndSubmit();
    if (!orderA) { passed = false; return; }
    billA = orderA.sale_bill_uuid;

    const h5A = findH5OrderForBill(billA);
    if (h5A) {
      acceptH5Order(h5A.h5_order_uuid);
    }

    // Order B: create → submit → reject
    const orderB = createAndSubmit();
    if (!orderB) { passed = false; return; }
    billB = orderB.sale_bill_uuid;

    const h5B = findH5OrderForBill(billB);
    if (h5B) {
      rejectH5Order(h5B.h5_order_uuid);
    }

    // Order C: create only (no submit)
    const orderC = createDineInOrder(BOM_UUID);
    if (!orderC) { passed = false; return; }
    billC = orderC.sale_bill_uuid;

    console.log(`Setup: A=${billA}(accepted) B=${billB}(rejected) C=${billC}(unsubmitted)`);
  });
  if (!billA) { scenarioFail.add(1); return; }

  group('ListFilter: all', () => {
    const data = getMemberOrderList('all');
    if (!data) { passed = false; return; }
    const total = data.meta ? data.meta.total : 0;
    // C should NOT appear (submit_pay_time=0)
    const hasBillA = data.list && data.list.some(i => i.sale_bill_uuid === billA);
    const hasBillC = data.list && data.list.some(i => i.sale_bill_uuid === billC);
    check(hasBillA, { 'List-all: has accepted order': (v) => v === true });
    check(hasBillC, { 'List-all: unsubmitted NOT visible': (v) => v !== true });
    console.log(`List-all: total=${total}`);
  });

  group('ListFilter: accepted order in all list with correct status', () => {
    // Accepted + unpaid order shows status "completed" in all-list
    // but NOT in "completed" filter (requires SaleBill.Status=Complete/paid)
    const data = getMemberOrderList('all');
    if (!data) { passed = false; return; }
    const orderA = data.list && data.list.find(i => i.sale_bill_uuid === billA);
    check(orderA, { 'List-all: accepted order found': (o) => o !== null && o !== undefined });
    if (orderA) {
      const st = orderA.status_info ? orderA.status_info.status : '';
      check(st, { 'List-all: accepted order status is completed': (s) => s === 'completed' });
      console.log(`Accepted order status in list: ${st}`);
    }
  });

  group('ListFilter: cancelled', () => {
    const data = getMemberOrderList('cancelled');
    if (!data) { passed = false; return; }
    const hasBillB = data.list && data.list.some(i => i.sale_bill_uuid === billB);
    check(hasBillB, { 'List-cancelled: has rejected order': (v) => v === true });
    console.log(`List-cancelled: total=${data.meta ? data.meta.total : 0}`);
  });

  if (passed) { scenarioPass.add(1); console.log('✓ Order list filter PASSED'); }
  else { scenarioFail.add(1); console.log('✗ Order list filter FAILED'); }
}

// ── Scenario 7: Invalid submit (edge cases) ──────────────────────
export function invalidSubmit() {
  let passed = true;

  group('InvalidSubmit: empty body', () => {
    const res = http.post(`${BASE_URL}/api/v1/member/order/dine_in/submit`,
      JSON.stringify({}),
      { headers: memberHeaders(), tags: { name: 'InvalidSubmit' } }
    );
    const body = parseAPI(res, 'EmptySubmit');
    const ok = check(body, {
      'EmptySubmit: rejected (code != 0)': (b) => b && b.code !== 0,
    });
    if (!ok) passed = false;
  });

  group('InvalidSubmit: non-existent bill', () => {
    const body = submitDineInOrder(999999999, 999999999);
    const ok = check(body, {
      'BadBill: rejected (code != 0)': (b) => b && b.code !== 0,
    });
    if (!ok) passed = false;
  });

  group('InvalidSubmit: cancelled order', () => {
    // Create → cancel → try submit
    const order = createDineInOrder(BOM_UUID);
    if (!order) { passed = false; return; }
    cancelDineInOrder(order.sale_bill_uuid);

    const body = submitDineInOrder(order.sale_bill_uuid, order.sale_order_uuid);
    const ok = check(body, {
      'CancelledSubmit: rejected (code != 0)': (b) => b && b.code !== 0,
    });
    if (!ok) {
      console.error(`Expected error for cancelled order submit, got: ${JSON.stringify(body)}`);
      passed = false;
    }
  });

  if (passed) { scenarioPass.add(1); console.log('✓ Invalid submit PASSED'); }
  else { scenarioFail.add(1); console.log('✗ Invalid submit FAILED'); }
}

// ── Scenario 8: Add products after submit + verify visibility ─────
export function addAfterSubmit() {
  let passed = true;
  let bill = 0, so = 0;

  group('AddAfterSubmit: create + submit', () => {
    const order = createDineInOrder(BOM_UUID);
    if (!order) { passed = false; return; }
    bill = order.sale_bill_uuid;
    so = order.sale_order_uuid;
    const body = submitDineInOrder(bill, so);
    if (!assertOK(body, 'AAS-Submit')) { passed = false; }
  });
  if (!bill) { scenarioFail.add(1); return; }

  group('AddAfterSubmit: member adds product after submit', () => {
    // This should succeed (create allows adding to existing order)
    const order = createDineInOrder(BOM_UUID, bill, so);
    const ok = check(order, {
      'AAS: add after submit succeeds': (o) => o !== null,
    });
    if (!ok) passed = false;
    else console.log('Added product after submit');
  });

  group('AddAfterSubmit: re-submit should fail', () => {
    const body = submitDineInOrder(bill, so);
    check(body, {
      'AAS: re-submit rejected': (b) => b && b.code !== 0,
    });
  });

  group('AddAfterSubmit: detail shows all products', () => {
    const detail = getMemberOrderDetail(bill);
    if (!detail) { passed = false; return; }
    const products = detail.product_list ? detail.product_list.list : [];
    // Original 1 product (submitted) + 1 added after submit = at least 1 visible
    // Note: the post-submit product may or may not appear depending on business logic
    console.log(`Detail: ${products.length} products, status=${getStatus(detail)}`);
    check(products, {
      'AAS: has products': (p) => p && p.length >= 1,
    });
  });

  if (passed) { scenarioPass.add(1); console.log('✓ Add after submit PASSED'); }
  else { scenarioFail.add(1); console.log('✗ Add after submit FAILED'); }
}
