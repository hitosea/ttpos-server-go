# BPS Link POS Specification - HYPERCOM

**Version Number:** 1.9

**Provided By:** BPS

> Copyright © 2019 Bangkok Payment Solutions Company Limited. All rights reserved.

---

## Document Information

| Item           | Details                                    |
| -------------- | ------------------------------------------ |
| Document Title | BPS Link POS Specification – HYPERCOM      |
| Prepared By    | Kittiphong Ngoenthavon                     |
| Reviewed By    | Patchara Kalampavanich                     |
| Approved By    | Ching Hung Hui                             |
| File Name      | BPS Link POS Specification - HYPERCOM.docx |

---

## Revision History

| Rev. | Section        | Description                                                                                                                                                                                                                  | Date       | Author        |
| ---- | -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- | ------------- |
| 1.0  | All            | Initial draft                                                                                                                                                                                                                | 2019-06-17 | Kittiphong N. |
| 1.1  | 1              | **Update:** 1.3 Error Handle; **Add:** 1.6 Serial Port Configuration, Appendix j EDC normal workflow                                                                                                                         | 2019-10-30 | Kittiphong N. |
| 1.2  | All            | Update Specification: Add link POS parameter for both request and response. Add link POS specification for Redemption.                                                                                                       | 2020-08-17 | Kittiphong N. |
| 1.3  | All            | Update Specification: Revise Sale request POS to EDC, Revise Sale response EDC to POS, Revise Sale QR request POS to EDC, Revise Sale QR response EDC to POS                                                                 | 2020-09-17 | Kittiphong N. |
| 1.4  | 1              | Update 3.3 response code.                                                                                                                                                                                                    | 2021-01-15 | Kittiphong N. |
| 1.5  | -              | Edit Command: Sale command 20 to 56, Void, Sale welfare card 56+TAX; Add command settlement                                                                                                                                  | 2021-01-18 | Kittiphong N. |
| 1.6  | 7-10           | Add Command: Summary Report Command 90, Detail Report Command 91, Settlement Command 50                                                                                                                                      | 2021-04-01 | Anatthiwat L. |
| 1.7  | 3,4,6,7,8,9,10 | Remove unused fields; Add VAT Rebate support to Sale (All Payment Types) Command 20, Sale (All Card Scheme Types) Command 56; Add HO field the response of Settlement Command, Summary Report Command, Detail Report Command | 2021-04-02 | Patchara K.   |
| 1.8  | 3,6,8,9,10,11  | Add HA field request of Settlement Command, Summary Report Command, Detail Report Command; Update 30 field response of QR code.                                                                                              | 2021-05-28 | Thongchai B.  |
| 1.9  | 10,11,12,13    | Add New command: detail report - all host, summary report with all host; Edit format data field type HO – Batch total; Add Transaction DATA 3.4 Mapping issuer card index, 3.5 Host Name                                     | 2021-06-14 | Thongchai B.  |

---

## Table of Contents

1. [Protocol](#1-protocol)
   - 1.1 Message from POS to Terminal
   - 1.2 Message from EDC to POS
   - 1.3 Error Handling
   - 1.4 Error Handling when Link POS Message Format Error
   - 1.5 Error Handling when Host Reject
   - 1.6 Serial Port Configuration
2. [Link POS Message](#2-link-pos-message)
   - 2.1 Message Structure
   - 2.2 Data
3. [Transaction Data](#3-transaction-data)
   - 3.1 Transaction Code Definition
   - 3.2 Field Type Definition
   - 3.3 Response Code
   - 3.4 Mapping Issuer Card Index
   - 3.5 Host Name
4. [Sale (All Payment Types - Transaction Code 20) Command](#4-sale-all-payment-types---transaction-code-20-command)
5. [Void (All Payment Types - Transaction Code 26) Command](#5-void-all-payment-types---transaction-code-26-command)
6. [Sale QR Payment (Transaction Code 70)](#6-sale-qr-payment-transaction-code-70)
7. [Sale (All Card Scheme Payment Types – Transaction Code 56) Command](#7-sale-all-card-scheme-payment-types--transaction-code-56-command)
8. [Settlement (Transaction Code 50) Command Each Host](#8-settlement-transaction-code-50-command-each-host)
9. [Settlement (Transaction Code 50) Command All Host](#9-settlement-transaction-code-50-command-all-host)
10. [Summary Report (Transaction Code 90) Command Each Host](#10-summary-report-transaction-code-90-command-each-host)
11. [Detail Report (Transaction Code 91) Command](#11-detail-report-transaction-code-91-command)
12. [Summary Report (Transaction Code 90) All Host](#12-summary-report-transaction-code-90-all-host)
13. [Detail Report (Transaction Code 91) All Host](#13-detail-report-transaction-code-91-all-host)

---

## 1 Protocol

In order to prevent a single error on the line from causing a message to be lost, a message will be acknowledged when the receiver returns an ACK character.

Only a single unacknowledged message can be outstanding at one time in one direction. The receipt of a message does not imply the previous message sent has been received.

### 1.1 Message from POS to Terminal

The POS transmits a message. The Terminal acknowledges receipt of the message by transmitting a single **ACK (06h)** character.

```
POS ──[Message]──> Terminal
POS <──[ACK]────── Terminal
```

### 1.2 Message from EDC to POS

The Terminal transmits a message. The POS acknowledges receipt of the message by transmitting a single **ACK (06h)** character.

```
Terminal ──[Message]──> POS
Terminal <──[ACK]────── POS
```

### 1.3 Error Handling

- If no ACK is received within timeout, retransmit the message
- Maximum retries: 3 times
- If all retries fail, report communication error

### 1.4 Error Handling when Link POS Message Format Error

When the EDC receives a message with format errors, it should respond with an error response.

### 1.5 Error Handling when Host Reject

When the host rejects a transaction, the EDC should forward the rejection response to the POS.

### 1.6 Serial Port Configuration

| Parameter    | Value                                 |
| ------------ | ------------------------------------- |
| Baud Rate    | 9600 / 19200 / 38400 / 57600 / 115200 |
| Data Bits    | 8                                     |
| Parity       | None                                  |
| Stop Bits    | 1                                     |
| Flow Control | None                                  |

---

## 2 Link POS Message

### 2.1 Message Structure

```
┌─────────┬──────────────────────┬─────────────────────┬──────────┬─────┐
│   STX   │   Transport Header   │ Presentation Header │   Data   │ ETX │
│  (02h)  │      (2 bytes)       │     (Variable)      │(Variable)│(03h)│
└─────────┴──────────────────────┴─────────────────────┴──────────┴─────┘
```

**Message Format:**

- **STX (02h):** Start of Text
- **Transport Header:** 2 bytes
- **Presentation Header:** Variable length
- **Data:** Field data
- **ETX (03h):** End of Text

### 2.2 Data

#### 2.2.1 Transport Header

| Byte | Description      |
| ---- | ---------------- |
| 1    | Length High Byte |
| 2    | Length Low Byte  |

#### 2.2.2 Presentation Header

| Field            | Length | Description                     |
| ---------------- | ------ | ------------------------------- |
| Format Version   | 1      | Version of message format       |
| Request/Response | 1      | '0' = Request, '1' = Response   |
| Transaction Code | 2      | Transaction type identifier     |
| Response Code    | 2      | Response status (Response only) |

#### 2.2.3 Field Data

Field data follows the presentation header and contains transaction-specific information.

#### 2.2.4 Field Element

Each field element has the following structure:

```
┌────────────┬────────────┬────────────┐
│ Field Type │   Length   │   Value    │
│  (2 bytes) │ (2 bytes)  │ (Variable) │
└────────────┴────────────┴────────────┘
```

#### 2.2.5 Extra Field Separator

Field separator: **FS (1Ch)**

---

## 3 Transaction Data

### 3.1 Transaction Code Definition

| Code | Transaction Type             |
| ---- | ---------------------------- |
| 20   | Sale (All Payment Types)     |
| 26   | Void (All Payment Types)     |
| 50   | Settlement                   |
| 56   | Sale (All Card Scheme Types) |
| 70   | Sale QR Payment              |
| 90   | Summary Report               |
| 91   | Detail Report                |

### 3.2 Field Type Definition

| Field | Length | Field Name            | Description                           |
| ----- | ------ | --------------------- | ------------------------------------- |
| 00    | 2      | Transaction Code      | Transaction type identifier           |
| 01    | 2      | Response Code         | Response status                       |
| 02    | 40     | Response Text         | Text message from host/EDC            |
| 03    | 6      | Transaction Date      | YYMMDD                                |
| 04    | 6      | Transaction Time      | HHMMSS                                |
| 16    | 8      | Terminal ID           | Terminal identification               |
| 30    | 12     | Amount                | Transaction amount (no decimal)       |
| 40    | 19     | Card Number           | Masked card number                    |
| 50    | 6      | Batch Number          | Batch number                          |
| 51    | 6      | Trace Number          | Transaction trace number              |
| 52    | 6      | Approval Code         | Authorization code                    |
| 53    | 12     | Reference Number      | RRN                                   |
| 61    | 62     | VAT Information       | VAT rebate data                       |
| D0    | 69     | Merchant Name/Address | 3 lines x 23 characters               |
| D1    | 15     | Merchant ID           | MID                                   |
| D2    | 40     | Card Holder Name      | Cardholder name                       |
| HA    | ...20  | Host Name             | Acquirer host name (variable, max 20) |
| HN    | 3      | NII                   | Network International Identifier      |
| HO    | 96     | Batch Totals          | Settlement totals                     |

### 3.3 Response Code

| Code | Description               |
| ---- | ------------------------- |
| 00   | Approved                  |
| 01   | Refer to card issuer      |
| 03   | Invalid merchant          |
| 04   | Pick up card              |
| 05   | Do not honor              |
| 12   | Invalid transaction       |
| 13   | Invalid amount            |
| 14   | Invalid card number       |
| 30   | Format error              |
| 41   | Lost card                 |
| 43   | Stolen card               |
| 51   | Insufficient funds        |
| 54   | Expired card              |
| 55   | Incorrect PIN             |
| 57   | Transaction not permitted |
| 58   | Transaction not allowed   |
| 61   | Exceeds withdrawal limit  |
| 91   | Issuer unavailable        |
| 94   | Duplicate transaction     |
| 96   | System malfunction        |
| XX   | Other errors              |

### 3.4 Mapping Issuer Card Index

| Index | Card Type  |
| ----- | ---------- |
| 1     | VISA       |
| 2     | MasterCard |
| 3     | JCB        |
| 4     | AMEX       |
| 5     | UnionPay   |
| 6     | Other      |

### 3.5 Host Name

Available host names for settlement and reports:

1. KBANK
2. TPN
3. KPLUS
4. SMARTPAY
5. REDEEM
6. DCC
7. QR_CREDIT
8. ALIPAY
9. WECHAT
10. AMEX
11. AMEX_EPP
12. BAY_INSTALLMENT
13. AYCAP_T1C
14. DOLFIN
15. SCB_IPP
16. SCB_OLS
17. BAY_REDEEM

---

## 4 Sale (All Payment Types - Transaction Code 20) Command

### 4.1 Request from POS to EDC

| M/O | Field | Length | Field Name      | Comments                                        |
| --- | ----- | ------ | --------------- | ----------------------------------------------- |
| M   | 30    | 12     | Amount          | Transaction amount (right-aligned, zero-padded) |
| O   | 61    | 62     | VAT Information | VAT rebate data (if applicable)                 |

### 4.2 Response from EDC to POS

| M/O | Field | Length | Field Name            | Comments                |
| --- | ----- | ------ | --------------------- | ----------------------- |
| M   | 01    | 2      | Response Code         | Host response code      |
| M   | 02    | 40     | Response Text         | Text message            |
| M   | 03    | 6      | Transaction Date      | YYMMDD                  |
| M   | 04    | 6      | Transaction Time      | HHMMSS                  |
| M   | 16    | 8      | Terminal ID           | Terminal identification |
| M   | 30    | 12     | Amount                | Transaction amount      |
| M   | 40    | 19     | Card Number           | Masked card number      |
| M   | 50    | 6      | Batch Number          | Batch number            |
| M   | 51    | 6      | Trace Number          | Transaction trace       |
| M   | 52    | 6      | Approval Code         | Authorization code      |
| M   | 53    | 12     | Reference Number      | RRN                     |
| M   | D0    | 69     | Merchant Name/Address | 3 lines x 23 chars      |
| M   | D1    | 15     | Merchant ID           | MID                     |
| O   | D2    | 40     | Card Holder Name      | Cardholder name         |
| O   | 61    | 62     | VAT Information       | VAT rebate response     |

---

## 5 Void (All Payment Types - Transaction Code 26) Command

### 5.1 Request from POS to EDC

| M/O | Field | Length | Field Name   | Comments                                  |
| --- | ----- | ------ | ------------ | ----------------------------------------- |
| M   | 51    | 6      | Trace Number | Original transaction trace number to void |

### 5.2 Response from EDC to POS

| M/O | Field | Length | Field Name            | Comments                |
| --- | ----- | ------ | --------------------- | ----------------------- |
| M   | 01    | 2      | Response Code         | Host response code      |
| M   | 02    | 40     | Response Text         | Text message            |
| M   | 03    | 6      | Transaction Date      | YYMMDD                  |
| M   | 04    | 6      | Transaction Time      | HHMMSS                  |
| M   | 16    | 8      | Terminal ID           | Terminal identification |
| M   | 30    | 12     | Amount                | Voided amount           |
| M   | 40    | 19     | Card Number           | Masked card number      |
| M   | 50    | 6      | Batch Number          | Batch number            |
| M   | 51    | 6      | Trace Number          | Transaction trace       |
| M   | 52    | 6      | Approval Code         | Authorization code      |
| M   | 53    | 12     | Reference Number      | RRN                     |
| M   | D0    | 69     | Merchant Name/Address | 3 lines x 23 chars      |
| M   | D1    | 15     | Merchant ID           | MID                     |

---

## 6 Sale QR Payment (Transaction Code 70)

### 6.1 Request from POS to EDC

| M/O | Field | Length | Field Name | Comments           |
| --- | ----- | ------ | ---------- | ------------------ |
| M   | 30    | 12     | Amount     | Transaction amount |

### 6.2 Response from EDC to POS

| M/O | Field | Length | Field Name            | Comments                |
| --- | ----- | ------ | --------------------- | ----------------------- |
| M   | 01    | 2      | Response Code         | Host response code      |
| M   | 02    | 40     | Response Text         | Text message            |
| M   | 03    | 6      | Transaction Date      | YYMMDD                  |
| M   | 04    | 6      | Transaction Time      | HHMMSS                  |
| M   | 16    | 8      | Terminal ID           | Terminal identification |
| M   | 30    | 12     | Amount                | Transaction amount      |
| M   | 50    | 6      | Batch Number          | Batch number            |
| M   | 51    | 6      | Trace Number          | Transaction trace       |
| M   | 52    | 6      | Approval Code         | Authorization code      |
| M   | 53    | 12     | Reference Number      | RRN                     |
| M   | D0    | 69     | Merchant Name/Address | 3 lines x 23 chars      |
| M   | D1    | 15     | Merchant ID           | MID                     |
| O   | HN    | 3      | NII                   | Network identifier      |

---

## 7 Sale (All Card Scheme Payment Types – Transaction Code 56) Command

### 7.1 Request from POS to EDC

| M/O | Field | Length | Field Name      | Comments                        |
| --- | ----- | ------ | --------------- | ------------------------------- |
| M   | 30    | 12     | Amount          | Transaction amount              |
| O   | 61    | 62     | VAT Information | VAT rebate data (if applicable) |

### 7.2 Response from EDC to POS

| M/O | Field | Length | Field Name            | Comments                |
| --- | ----- | ------ | --------------------- | ----------------------- |
| M   | 01    | 2      | Response Code         | Host response code      |
| M   | 02    | 40     | Response Text         | Text message            |
| M   | 03    | 6      | Transaction Date      | YYMMDD                  |
| M   | 04    | 6      | Transaction Time      | HHMMSS                  |
| M   | 16    | 8      | Terminal ID           | Terminal identification |
| M   | 30    | 12     | Amount                | Transaction amount      |
| M   | 40    | 19     | Card Number           | Masked card number      |
| M   | 50    | 6      | Batch Number          | Batch number            |
| M   | 51    | 6      | Trace Number          | Transaction trace       |
| M   | 52    | 6      | Approval Code         | Authorization code      |
| M   | 53    | 12     | Reference Number      | RRN                     |
| M   | D0    | 69     | Merchant Name/Address | 3 lines x 23 chars      |
| M   | D1    | 15     | Merchant ID           | MID                     |
| O   | D2    | 40     | Card Holder Name      | Cardholder name         |
| O   | 61    | 62     | VAT Information       | VAT rebate response     |

---

## 8 Settlement (Transaction Code 50) Command Each Host

### 8.1 Request from POS to EDC

| M/O | Field | Length | Field Name | Comments                                                                      |
| --- | ----- | ------ | ---------- | ----------------------------------------------------------------------------- |
| M   | HA    | ...20  | Host Name  | Specific host to settle. Example: `KBANK` (Hex: `48 41 00 05 4B 42 41 4E 4B`) |

**Available Host Names:**

1. KBANK
2. TPN
3. KPLUS
4. SMARTPAY
5. REDEEM
6. DCC
7. QR_CREDIT
8. ALIPAY
9. WECHAT
10. AMEX
11. AMEX_EPP
12. BAY_INSTALLMENT
13. AYCAP_T1C
14. DOLFIN
15. SCB_IPP
16. SCB_OLS
17. BAY_REDEEM

### 8.2 Response from EDC to POS

| M/O | Field | Length | Field Name            | Comments                         |
| --- | ----- | ------ | --------------------- | -------------------------------- |
| M   | 01    | 2      | Response Code         | Host response code               |
| M   | 02    | 40     | Response Text         | Text message                     |
| M   | 03    | 6      | Transaction Date      | YYMMDD                           |
| M   | 04    | 6      | Transaction Time      | HHMMSS                           |
| M   | 16    | 8      | Terminal ID           | Terminal identification          |
| M   | 50    | 6      | Batch Number          | Batch number                     |
| M   | D0    | 69     | Merchant Name/Address | 3 lines x 23 chars               |
| M   | D1    | 15     | Merchant ID           | MID                              |
| M   | HN    | 3      | NII                   | Network identifier               |
| M   | HO    | 96     | Batch Totals          | See Batch Totals structure below |

**Batch Totals (HO) Structure:**

| Position | Length | Field                   |
| -------- | ------ | ----------------------- |
| 1-3      | 3      | Capture Sale Count      |
| 4-15     | 12     | Capture Sale Amount     |
| 16-18    | 3      | Capture Refund Count    |
| 19-30    | 12     | Capture Refund Amount   |
| 31-33    | 3      | Debit Sale Count        |
| 34-45    | 12     | Debit Sale Amount       |
| 46-48    | 3      | Debit Refund Count      |
| 49-60    | 12     | Debit Refund Amount     |
| 61-63    | 3      | Authorize Sale Count    |
| 64-75    | 12     | Authorize Sale Amount   |
| 76-78    | 3      | Authorize Refund Count  |
| 79-90    | 12     | Authorize Refund Amount |

---

## 9 Settlement (Transaction Code 50) Command All Host

### 9.1 Request from POS to EDC

| M/O | Field | Length | Field Name | Comments                       |
| --- | ----- | ------ | ---------- | ------------------------------ |
| M   | HN    | 3      | NII        | Send `999` to settle all hosts |

### 9.2 Response from EDC to POS

Same as Section 8.2

---

## 10 Summary Report (Transaction Code 90) Command Each Host

### 10.1 Request from POS to EDC

| M/O | Field | Length | Field Name | Comments                         |
| --- | ----- | ------ | ---------- | -------------------------------- |
| M   | HA    | ...20  | Host Name  | Specific host for summary report |

### 10.2 Response from EDC to POS

| M/O | Field | Length | Field Name            | Comments                   |
| --- | ----- | ------ | --------------------- | -------------------------- |
| M   | 01    | 2      | Response Code         | Response code              |
| M   | 02    | 40     | Response Text         | Text message               |
| M   | 03    | 6      | Transaction Date      | YYMMDD                     |
| M   | 04    | 6      | Transaction Time      | HHMMSS                     |
| M   | 16    | 8      | Terminal ID           | Terminal identification    |
| M   | 50    | 6      | Batch Number          | Batch number               |
| M   | D0    | 69     | Merchant Name/Address | 3 lines x 23 chars         |
| M   | D1    | 15     | Merchant ID           | MID                        |
| M   | HN    | 3      | NII                   | Network identifier         |
| M   | HO    | 96     | Batch Totals          | See Batch Totals structure |

---

## 11 Detail Report (Transaction Code 91) Command

### 11.1 Request from POS to EDC

| M/O | Field | Length | Field Name | Comments                        |
| --- | ----- | ------ | ---------- | ------------------------------- |
| M   | HA    | ...20  | Host Name  | Specific host for detail report |

### 11.2 Response from EDC to POS

| M/O | Field | Length | Field Name            | Comments                   |
| --- | ----- | ------ | --------------------- | -------------------------- |
| M   | 01    | 2      | Response Code         | Response code              |
| M   | 02    | 40     | Response Text         | Text message               |
| M   | 03    | 6      | Transaction Date      | YYMMDD                     |
| M   | 04    | 6      | Transaction Time      | HHMMSS                     |
| M   | 16    | 8      | Terminal ID           | Terminal identification    |
| M   | 50    | 6      | Batch Number          | Batch number               |
| M   | D0    | 69     | Merchant Name/Address | 3 lines x 23 chars         |
| M   | D1    | 15     | Merchant ID           | MID                        |
| M   | HN    | 3      | NII                   | Network identifier         |
| M   | HO    | 96     | Batch Totals          | See Batch Totals structure |

---

## 12 Summary Report (Transaction Code 90) All Host

### 12.1 Request from POS to EDC

| M/O | Field | Length | Field Name | Comments                                |
| --- | ----- | ------ | ---------- | --------------------------------------- |
| M   | HN    | 3      | NII        | Send `999` to get summary for all hosts |

### 12.2 Response from EDC to POS

| M/O | Field | Length | Field Name            | Comments                        |
| --- | ----- | ------ | --------------------- | ------------------------------- |
| M   | 02    | 40     | Response Text         | Text message                    |
| M   | D0    | 69     | Merchant Name/Address | 3 lines x 23 chars              |
| M   | 16    | 8      | Terminal ID           | Fixed: `00000000`               |
| M   | D1    | 15     | Merchant ID           | Fixed: `000000000000000`        |
| M   | 50    | 6      | Batch Number          | Fixed: `000001`                 |
| M   | 03    | 6      | Transaction Date      | YYMMDD                          |
| M   | 04    | 6      | Transaction Time      | HHMMSS                          |
| M   | HN    | 3      | NII                   | Fixed: `999`                    |
| M   | HO    | 96     | Batch Totals          | Aggregated totals for all hosts |

---

## 13 Detail Report (Transaction Code 91) All Host

### 13.1 Request from POS to EDC

| M/O | Field | Length | Field Name | Comments                               |
| --- | ----- | ------ | ---------- | -------------------------------------- |
| M   | HN    | 3      | NII        | Send `999` to get detail for all hosts |

### 13.2 Response from EDC to POS

| M/O | Field | Length | Field Name            | Comments                        |
| --- | ----- | ------ | --------------------- | ------------------------------- |
| M   | 02    | 40     | Response Text         | Text message                    |
| M   | D0    | 69     | Merchant Name/Address | 3 lines x 23 chars              |
| M   | 16    | 8      | Terminal ID           | Fixed: `00000000`               |
| M   | D1    | 15     | Merchant ID           | Fixed: `000000000000000`        |
| M   | 50    | 6      | Batch Number          | Fixed: `000001`                 |
| M   | 03    | 6      | Transaction Date      | YYMMDD                          |
| M   | 04    | 6      | Transaction Time      | HHMMSS                          |
| M   | HN    | 3      | NII                   | Fixed: `999`                    |
| M   | HO    | 96     | Batch Totals          | Aggregated totals for all hosts |

---

## Appendix A: VAT Rebate

### VAT Rebate Example

| Product         | Amount  | Amount for TAX                 | VAT             |
| --------------- | ------- | ------------------------------ | --------------- |
| Pen             | 100     | -                              | -               |
| Durian Snack    | 100     | 100 (OTOP)                     | 7               |
| Book            | 100     | 100 (Book)                     | 7               |
| Alcohol         | 100     | -                              | -               |
| **Total**       | **400** |                                |                 |
| VAT 7%          | 28      | 200 (Amount for TAX Allowance) | 14 (VAT Rebate) |
| **Amount Paid** | **428** |                                |                 |

### VAT Information Field (61) Structure

**Total Length:** 62 bytes

| Subfield | Position | Format | Description                                                                                                                                                                               |
| -------- | -------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1        | 1-18     | n-18   | **VAT Rebate Amount** - Cashback amount for welfare card and debit card. Used in 0180 authorization requests, 0400 and 0420 reversal requests. Example: 14.00 Baht = `000000000000001400` |
| 2        | 19-36    | n-18   | **Amount for TAX Allowance** - Amount used to calculate VAT in subfield 1. Example: 200.00 Baht = `000000000000020000`                                                                    |
| 3        | 37-56    | an-20  | **Merchant Unique Value** - Unique reference from POS for reconciliation. Example: `00010293000000000257`                                                                                 |
| 4        | 57-62    | an-6   | **Campaign Type** - Campaign number in format YYNNNN. Example: `190001` (Government stimulus), `190002` (OTOP promotion)                                                                  |

**Example VAT Information:**

```
000000000000001400000000000002000000010293000000000257190001
```

---

## Appendix B: Hex Examples

### ACK Character

```
Hex: 06
```

### STX Character

```
Hex: 02
```

### ETX Character

```
Hex: 03
```

### Field Separator (FS)

```
Hex: 1C
```

### Host Name Example (KBANK)

```
Field: HA
Hex: 48 41 00 05 4B 42 41 4E 4B
     H  A  len=5 K  B  A  N  K
```

---

_BPS Confidential - Copyright © 2019 Bangkok Payment Solutions Company Limited. All rights reserved._
