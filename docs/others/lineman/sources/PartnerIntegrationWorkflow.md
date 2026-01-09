# Partner Integration Workflow

**LMWN Engineering Team**

---

## Agenda

- Architecture Overview
- Partner Integration Process
- API Walkthrough
  - Authentication
  - Menu
  - Order
  - Restaurant
- Onboarding Store to Integration
- Contact Channel

---

## Architecture Overview

### Partner Integration Background

**Order Flow:**

- LINEMAN system places orders directly into the Partner POS system.
- At the same time, orders continue to appear in the Wongnai Merchant App (WMA) for restaurant staff visibility.

**Fallback Handling:**

- If the system fails to deliver an order to the Partner POS, the order will still be sent to WMA.
- This ensures restaurants never miss an order, even in the event of integration or connectivity issues.

---

## Partner Integration Process

### Sandbox Environment

```
Partner Discussion → Partner Implementation → Testing → Rollout (Pilot rollout / Phase rollout)
```

### API Summary (4 mandatory, 6 optional)

| Scope          | API                              | Required  | Definition                                                                                   |
| -------------- | -------------------------------- | --------- | -------------------------------------------------------------------------------------------- |
| Authentication | Authentication                   | Mandatory | Partner system can access to LINE MAN system                                                 |
| Menu           | Menu sync                        | Mandatory | Partner can sync menu to LINE MAN.                                                           |
| Menu           | Trigger sync menu                | Mandatory | LINE MAN can trigger sync menu to partner                                                    |
| Menu           | Menu sync notification           | Optional  | LINE MAN notifies partner about the menu sync result                                         |
| Menu           | Update menu item status          | Optional  | Partner can sync menu item status for some menus                                             |
| Menu           | Update menu propertyValue status | Optional  | Partner can sync menu option item status for some option                                     |
| Order          | Place order notification         | Mandatory | LINE MAN notifies new order contains all order details including price, items, order id etc. |
| Order          | Status update notification       | Optional  | LINE MAN notifies order status if the order status is finished or canceled.                  |
| Order          | Order update notification        | Optional  | LINE MAN notifies order detail change if the order is edited                                 |
| Restaurant     | Force close/open restaurant      | Optional  | Partner can force close/open their store on LINE MAN from their POS                          |

---

## API Walkthrough

### Authentication

- **Oauth 2.0 - Client Credential Type**
- Both sides need to authenticate with each other

---

### Menu

#### Sync Menu Workflow

| API                               | Definition                                                                                                                                                      |
| --------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Menu sync                         | Partner can sync menu to LINE MAN. Whenever the menu/menu status are changed on the partner system, partner needs to push the entire menu snapshot to LINE MAN. |
| Menu sync notification (Optional) | LINE MAN notifies partner about the menu sync result (SUCCESS / FAILED)                                                                                         |

**Key Points:**

- We currently support only push menu workflow
- Partner can update menu anytime
- Partner branch staffs don't need to update menu status on the tablet (the feature is still available on the tablet)

#### Trigger Sync Menu

| API                               | Definition                                                                                                                                                      |
| --------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Menu sync                         | Partner can sync menu to LINE MAN. Whenever the menu/menu status are changed on the partner system, partner needs to push the entire menu snapshot to LINE MAN. |
| Trigger sync menu                 | Allow LINE MAN to trigger sync menu to partner, if there is any error in the sync menu process and LINE MAN knows it.                                           |
| Menu sync notification (Optional) | LINE MAN notifies partner about the menu sync result (SUCCESS / FAILED)                                                                                         |

**Key Points:**

- We currently support only push menu workflow
- Partner branch staffs don't need to update menu status on the tablet (the feature is still available on the tablet)
- **Avoid peak hours (10:00 - 14:00 / 17:00 - 19:00)**

#### Sync Menu Status Workflow (Optional)

> Note: This API will only synced the menu status.

| API                                         | Definition                                                                                                 |
| ------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| Update menu item status (Optional)          | Partner can sync menu item status for some menus: AVAILABLE (default), SOLD_OUT_TODAY, SUSPENDED           |
| Update menu propertyValue status (Optional) | Partner can sync menu option item status for some option: AVAILABLE (1), SOLD_OUT_TODAY (2), SUSPENDED (3) |

#### Menu Promotion

**Process for Menu Synchronization (Menu Promotion = LINE MAN Promotion):**

1. **Promotion Agreement:** Menu promotions to be displayed are agreed upon between LINE MAN BD and the Partner's business team.

   - Agree result in brief file

2. **Promotion Setup:** The LINE MAN Operations team sets up the promotion on the LINE MAN side based on the brief file provided by the business team.

3. **Menu Sync:** The Partner syncs their menu into LINE MAN.

   - The menuName, menuProperty, menuPropertyValue, Price of the promotion must exactly match the promotion configured on the LINE MAN side.
   - If they do not match, the promotion will not be orderable, and one of the teams must correct it.
   - Menu pictures must be sent to LINE MAN BD.

4. **Testing:** Normally, the menu sync is tested first in the test environment to verify correctness and to test order placement.

**How the System Differentiates Normal Menu vs. LINE MAN Promotion:**

- The system detects promotion menus from items whose names start with `[Promotion]`.
- **Important Rule:** Do not use `[Promotion]` as a prefix in normal menu item names — it is reserved exclusively for promotions.

---

### Order

#### Sync Order Workflow

| API                                   | Definition                                                                                   |
| ------------------------------------- | -------------------------------------------------------------------------------------------- |
| Place order notification              | LINE MAN notifies new order contains all order details including price, items, order id etc. |
| Status update notification (optional) | LINE MAN notifies order status if the order status is finished or canceled.                  |

> **Note:** Tablet is still required for Backup

#### Cancel Order Workflow (Optional)

| API                                   | Definition                                                                  |
| ------------------------------------- | --------------------------------------------------------------------------- |
| Status update notification (optional) | LINE MAN notifies order status if the order status is finished or canceled. |

> **Note:** Tablet is still required for Backup

#### Order Update Notification Workflow (Optional)

| API                                  | Definition                                                   |
| ------------------------------------ | ------------------------------------------------------------ |
| Order update notification (optional) | LINE MAN notifies order detail change if the order is edited |

#### Perform Operation Actions

- **Ready to Pickup (optional):** Can be done on WMA app
- **Cancel Order - required WMA App:** Update order status to POS
- **Edit Order - required WMA App:** Update order detail to POS

---

### Restaurant

#### Force Close/Open Restaurant Workflow (Optional)

| API                                    | Definition                                                          |
| -------------------------------------- | ------------------------------------------------------------------- |
| Force close/open restaurant (optional) | Partner can force close/open their store on LINE MAN from their POS |

---

## Onboarding Store to Integration

1. **Store ID Submission**

   - Partner provides the store ID to LINE MAN BD to initiate integration.

2. **LINE MAN Setup (SLA: 4 working days)**

   - LINE MAN team binds the store by linking store ID, partner, and restaurant ID (RID).

3. **Menu Synchronization**

   - Partner syncs menu data into LINE MAN.
   - ⚠ **The synced menu will immediately affect the store's live (production) menu.**

4. **Enable Integration**

---

## Contact Point

- **Contact Channel:** Line Group
- **Menu/Promotion Setup, Sync Menu:** Business Development
- **Others (Technical, Features):** LINE MAN Technical Team (App Support, Product Engineering Team)

---

_Source: LMWN Engineering Team - Partner Integration Workflow_
