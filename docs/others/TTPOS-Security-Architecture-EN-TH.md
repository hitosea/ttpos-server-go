# TTPOS – Security Architecture
# สถาปัตยกรรมความปลอดภัย TTPOS

> **Version**: 1.0  
> **Last Updated**: 2025-11-27  
> **Document Type**: Security Documentation

---

## Table of Contents / สารบัญ

1. [Overview / ภาพรวม](#1-overview--ภาพรวม)
2. [Data Transmission Security / ความปลอดภัยในการส่งข้อมูล](#2-data-transmission-security--ความปลอดภัยในการส่งข้อมูล)
3. [User Permissions and Roles / สิทธิ์ผู้ใช้และบทบาท](#3-user-permissions-and-roles--สิทธิ์ผู้ใช้และบทบาท)
4. [Backend API Security / ความปลอดภัย API ฝั่งเซิร์ฟเวอร์](#4-backend-api-security--ความปลอดภัย-api-ฝั่งเซิร์ฟเวอร์)
5. [Data Storage Security / ความปลอดภัยในการจัดเก็บข้อมูล](#5-data-storage-security--ความปลอดภัยในการจัดเก็บข้อมูล)
6. [Audit Logging and Monitoring / การบันทึกและการตรวจสอบ](#6-audit-logging-and-monitoring--การบันทึกและการตรวจสอบ)
7. [Security Best Practices / แนวปฏิบัติที่ดีด้านความปลอดภัย](#7-security-best-practices--แนวปฏิบัติที่ดีด้านความปลอดภัย)

---

## 1. Overview / ภาพรวม

### English

TTPOS (TableTop Point of Sale) is a modern restaurant POS system built with a microservices architecture. Security is implemented at multiple layers to ensure data confidentiality, integrity, and availability across all system components.

**Key Security Principles:**
- Defense in depth (multiple security layers)
- Principle of least privilege
- Data encryption at rest and in transit
- Comprehensive audit logging
- Role-based access control (RBAC)

### ภาษาไทย

TTPOS (TableTop Point of Sale) เป็นระบบ POS สำหรับร้านอาหารที่ทันสมัย ออกแบบด้วยสถาปัตยกรรมไมโครเซอร์วิส การรักษาความปลอดภัยถูกนำมาใช้ในหลายระดับเพื่อให้มั่นใจว่าข้อมูลมีความลับ ความสมบูรณ์ และพร้อมใช้งานในทุกส่วนประกอบของระบบ

**หลักการความปลอดภัยหลัก:**
- การป้องกันแบบหลายชั้น (Defense in Depth)
- หลักการให้สิทธิ์น้อยที่สุด (Least Privilege)
- การเข้ารหัสข้อมูลทั้งขณะจัดเก็บและขณะส่ง
- การบันทึกการตรวจสอบอย่างครอบคลุม
- การควบคุมการเข้าถึงตามบทบาท (RBAC)

---

## 2. Data Transmission Security / ความปลอดภัยในการส่งข้อมูล

### 2.1 HTTPS/TLS Encryption | การเข้ารหัส HTTPS/TLS

#### English

All external communications between clients and the TTPOS server are encrypted using TLS (Transport Layer Security).

**Architecture:**
```
[Client App] <--HTTPS/TLS--> [Nginx Reverse Proxy] <--HTTP--> [Backend Services]
                                                              (Docker Internal Network)
```

**Security Features:**
- TLS 1.2+ enforced for all external connections
- SSL certificates managed via Let's Encrypt or commercial CA
- Nginx acts as TLS termination point
- Internal Docker network traffic is isolated from external access

**Nginx Configuration Highlights:**
- Real IP forwarding from proxy headers
- Secure header configurations
- GZIP compression for optimized data transfer
- CORS (Cross-Origin Resource Sharing) controls

#### ภาษาไทย

การสื่อสารภายนอกทั้งหมดระหว่างแอปพลิเคชันไคลเอนต์และเซิร์ฟเวอร์ TTPOS จะถูกเข้ารหัสด้วย TLS (Transport Layer Security)

**สถาปัตยกรรม:**
```
[แอปไคลเอนต์] <--HTTPS/TLS--> [Nginx Reverse Proxy] <--HTTP--> [บริการหลังบ้าน]
                                                                 (เครือข่ายภายใน Docker)
```

**คุณสมบัติด้านความปลอดภัย:**
- บังคับใช้ TLS 1.2 ขึ้นไปสำหรับการเชื่อมต่อภายนอกทั้งหมด
- จัดการใบรับรอง SSL ผ่าน Let's Encrypt หรือ CA เชิงพาณิชย์
- Nginx ทำหน้าที่เป็นจุดสิ้นสุด TLS
- การรับส่งข้อมูลเครือข่าย Docker ภายในถูกแยกจากการเข้าถึงภายนอก

**การกำหนดค่า Nginx ที่สำคัญ:**
- การส่งต่อ Real IP จาก proxy headers
- การกำหนดค่า secure headers
- การบีบอัด GZIP เพื่อการถ่ายโอนข้อมูลที่เหมาะสม
- การควบคุม CORS (Cross-Origin Resource Sharing)

---

### 2.2 End-to-End API Encryption (Optional) | การเข้ารหัส API แบบ End-to-End (ตัวเลือก)

#### English

TTPOS supports an additional layer of application-level encryption for sensitive API communications using a hybrid RSA + AES encryption scheme.

**Encryption Flow:**
1. Client generates a random AES symmetric key
2. Client encrypts request body with AES-CBC
3. Client encrypts the AES key with server's RSA public key
4. Server decrypts AES key with RSA private key
5. Server decrypts request body with AES key

**Implementation:**
- **RSA**: 2048-bit key pairs for key exchange
- **AES**: 256-bit keys with CBC mode for data encryption
- **IV**: Random initialization vector for each request

**Request Header:**
```
X-Encrypt: client_id=xxxx;client_key=xxxxx;type=jsencrypt
```

#### ภาษาไทย

TTPOS รองรับการเข้ารหัสระดับแอปพลิเคชันเพิ่มเติมสำหรับการสื่อสาร API ที่มีความละเอียดอ่อน โดยใช้รูปแบบการเข้ารหัสแบบไฮบริด RSA + AES

**ขั้นตอนการเข้ารหัส:**
1. ไคลเอนต์สร้างคีย์ AES สมมาตรแบบสุ่ม
2. ไคลเอนต์เข้ารหัส request body ด้วย AES-CBC
3. ไคลเอนต์เข้ารหัสคีย์ AES ด้วยคีย์สาธารณะ RSA ของเซิร์ฟเวอร์
4. เซิร์ฟเวอร์ถอดรหัสคีย์ AES ด้วยคีย์ส่วนตัว RSA
5. เซิร์ฟเวอร์ถอดรหัส request body ด้วยคีย์ AES

**การนำไปใช้:**
- **RSA**: คู่คีย์ 2048 บิตสำหรับการแลกเปลี่ยนคีย์
- **AES**: คีย์ 256 บิตพร้อมโหมด CBC สำหรับการเข้ารหัสข้อมูล
- **IV**: Initialization vector แบบสุ่มสำหรับแต่ละคำขอ

**Request Header:**
```
X-Encrypt: client_id=xxxx;client_key=xxxxx;type=jsencrypt
```

---

## 3. User Permissions and Roles / สิทธิ์ผู้ใช้และบทบาท

### 3.1 Role-Based Access Control (RBAC) | การควบคุมการเข้าถึงตามบทบาท

#### English

TTPOS implements a comprehensive role-based access control system to manage user permissions across different terminals and functionalities.

**User Types:**

| User Type | Description | Access Level |
|-----------|-------------|--------------|
| Super Admin | Full system access | All modules |
| Store Admin | Store management | Store-level operations |
| Cashier | POS operations | Transaction processing |
| Kitchen Staff | Order display | KDS viewing |
| Assistant | Order assistance | Limited POS functions |
| Member | Customer account | Member portal only |

**Database Schema:**
```
Staff (employee) ←→ StaffRole (junction) ←→ Role
                                              ↓
                                          RoleAccess → Access (permissions)
```

**Key Permission Fields:**
- `is_super`: Super administrator flag (full access)
- `has_data_permission`: Data management privileges
- `user_type`: Account type (0=headquarters, 1=store)
- `is_disable`: Account disable flag

#### ภาษาไทย

TTPOS ใช้ระบบควบคุมการเข้าถึงตามบทบาทอย่างครอบคลุมเพื่อจัดการสิทธิ์ผู้ใช้ในเทอร์มินัลและฟังก์ชันต่างๆ

**ประเภทผู้ใช้:**

| ประเภทผู้ใช้ | คำอธิบาย | ระดับการเข้าถึง |
|-------------|----------|----------------|
| Super Admin | การเข้าถึงระบบเต็มรูปแบบ | ทุกโมดูล |
| Store Admin | การจัดการร้านค้า | การดำเนินงานระดับร้าน |
| Cashier | การดำเนินงาน POS | การประมวลผลธุรกรรม |
| Kitchen Staff | การแสดงคำสั่ง | การดู KDS |
| Assistant | ผู้ช่วยสั่งอาหาร | ฟังก์ชัน POS จำกัด |
| Member | บัญชีลูกค้า | พอร์ทัลสมาชิกเท่านั้น |

**โครงสร้างฐานข้อมูล:**
```
Staff (พนักงาน) ←→ StaffRole (ตารางเชื่อม) ←→ Role
                                               ↓
                                           RoleAccess → Access (สิทธิ์)
```

**ฟิลด์สิทธิ์หลัก:**
- `is_super`: แฟล็กผู้ดูแลระบบระดับสูง (เข้าถึงได้ทั้งหมด)
- `has_data_permission`: สิทธิ์การจัดการข้อมูล
- `user_type`: ประเภทบัญชี (0=สำนักงานใหญ่, 1=ร้านค้า)
- `is_disable`: แฟล็กปิดการใช้งานบัญชี

---

### 3.2 Multi-Terminal Authentication | การยืนยันตัวตนหลายเทอร์มินัล

#### English

TTPOS supports multiple client terminals, each with specific authentication requirements and access scopes:

| Terminal | Code | Description |
|----------|------|-------------|
| POS | `cashier` | Front counter cashier |
| Shop Backend | `shop` | Store management portal |
| Kitchen Display | `kitchen` | Kitchen order display |
| Tablet | `tablet` | Table-side ordering |
| Assistant | `assistant` | Order assistant device |
| Member App | `member` | Customer member portal |
| H5 (Mobile) | `h5` | Mobile web ordering |

**Authentication Flow:**
1. User submits credentials (username/password)
2. Server validates credentials against database
3. Server generates JWT token with user claims
4. Token includes: source, company_uuid, staff_uuid, device_id
5. Client includes token in Authorization header for subsequent requests

#### ภาษาไทย

TTPOS รองรับเทอร์มินัลไคลเอนต์หลายตัว แต่ละตัวมีข้อกำหนดการยืนยันตัวตนและขอบเขตการเข้าถึงเฉพาะ:

| เทอร์มินัล | รหัส | คำอธิบาย |
|-----------|------|---------|
| POS | `cashier` | แคชเชียร์หน้าเคาน์เตอร์ |
| Shop Backend | `shop` | พอร์ทัลการจัดการร้านค้า |
| Kitchen Display | `kitchen` | การแสดงคำสั่งครัว |
| Tablet | `tablet` | การสั่งอาหารข้างโต๊ะ |
| Assistant | `assistant` | อุปกรณ์ผู้ช่วยสั่งอาหาร |
| Member App | `member` | พอร์ทัลสมาชิกลูกค้า |
| H5 (Mobile) | `h5` | การสั่งอาหารบนเว็บมือถือ |

**ขั้นตอนการยืนยันตัวตน:**
1. ผู้ใช้ส่งข้อมูลรับรอง (ชื่อผู้ใช้/รหัสผ่าน)
2. เซิร์ฟเวอร์ตรวจสอบข้อมูลรับรองกับฐานข้อมูล
3. เซิร์ฟเวอร์สร้าง JWT token พร้อม user claims
4. Token ประกอบด้วย: source, company_uuid, staff_uuid, device_id
5. ไคลเอนต์รวม token ใน Authorization header สำหรับคำขอถัดไป

---

## 4. Backend API Security / ความปลอดภัย API ฝั่งเซิร์ฟเวอร์

### 4.1 JWT Authentication | การยืนยันตัวตนด้วย JWT

#### English

TTPOS uses JSON Web Tokens (JWT) for stateless authentication across all API endpoints.

**JWT Token Structure:**
```json
{
  "source": "cashier",
  "company_uuid": 12345,
  "staff_uuid": 67890,
  "device_uuid": 11111,
  "device_id": "POS-001",
  "is_refresh_token": false,
  "exp": 1699999999,
  "iat": 1699900000
}
```

**Token Types:**
- **Access Token**: Short-lived token for API access
- **Refresh Token**: Long-lived token for obtaining new access tokens

**Security Measures:**
- HS256 (HMAC-SHA256) signing algorithm
- Token expiration enforcement
- Token validation on every request
- Separate refresh token mechanism

#### ภาษาไทย

TTPOS ใช้ JSON Web Tokens (JWT) สำหรับการยืนยันตัวตนแบบ stateless ทั่วทุก API endpoints

**โครงสร้าง JWT Token:**
```json
{
  "source": "cashier",
  "company_uuid": 12345,
  "staff_uuid": 67890,
  "device_uuid": 11111,
  "device_id": "POS-001",
  "is_refresh_token": false,
  "exp": 1699999999,
  "iat": 1699900000
}
```

**ประเภท Token:**
- **Access Token**: Token อายุสั้นสำหรับการเข้าถึง API
- **Refresh Token**: Token อายุยาวสำหรับการขอ access token ใหม่

**มาตรการความปลอดภัย:**
- อัลกอริทึมการเซ็น HS256 (HMAC-SHA256)
- การบังคับใช้วันหมดอายุ token
- การตรวจสอบ token ทุกคำขอ
- กลไก refresh token แยกต่างหาก

---

### 4.2 API Access Control | การควบคุมการเข้าถึง API

#### English

**Path-Based Access Control:**
- API paths are validated against user's source (terminal type)
- Pattern: `/api/v{version}/{source}/...`
- Cross-terminal access is blocked

**Request Validation:**
```go
// Path validation example
if !regexp.MustCompile(`^/api/v\d+/`+claims.Source).Match([]byte(urlPath)) {
    // Access denied
}
```

**Security Headers:**
- Authorization: Bearer {token}
- X-Real-IP: Client IP address
- X-Forwarded-For: Proxy chain

**Input Validation:**
- Strict parameter binding with validation tags
- Required field enforcement
- Type checking and range validation

#### ภาษาไทย

**การควบคุมการเข้าถึงตามเส้นทาง:**
- เส้นทาง API ถูกตรวจสอบกับ source ของผู้ใช้ (ประเภทเทอร์มินัล)
- รูปแบบ: `/api/v{version}/{source}/...`
- การเข้าถึงข้ามเทอร์มินัลถูกบล็อก

**การตรวจสอบคำขอ:**
```go
// ตัวอย่างการตรวจสอบเส้นทาง
if !regexp.MustCompile(`^/api/v\d+/`+claims.Source).Match([]byte(urlPath)) {
    // ปฏิเสธการเข้าถึง
}
```

**Security Headers:**
- Authorization: Bearer {token}
- X-Real-IP: ที่อยู่ IP ไคลเอนต์
- X-Forwarded-For: Proxy chain

**การตรวจสอบอินพุต:**
- การ binding พารามิเตอร์อย่างเข้มงวดพร้อมแท็กการตรวจสอบ
- การบังคับใช้ฟิลด์ที่จำเป็น
- การตรวจสอบประเภทและช่วง

---

### 4.3 CORS Configuration | การกำหนดค่า CORS

#### English

Cross-Origin Resource Sharing (CORS) is configured to control which domains can access the API:

**Configuration:**
```go
c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
c.Writer.Header().Set("Access-Control-Allow-Headers", "...")
```

**Allowed Headers Include:**
- Authorization, Content-Type
- X-CSRF-TOKEN, X-Requested-With
- Device-ID, Platform, Version-Name
- Custom headers for encryption (X-Sign, Encrypt)

#### ภาษาไทย

Cross-Origin Resource Sharing (CORS) ถูกกำหนดค่าเพื่อควบคุมว่าโดเมนใดสามารถเข้าถึง API ได้:

**การกำหนดค่า:**
```go
c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
c.Writer.Header().Set("Access-Control-Allow-Headers", "...")
```

**Headers ที่อนุญาตรวมถึง:**
- Authorization, Content-Type
- X-CSRF-TOKEN, X-Requested-With
- Device-ID, Platform, Version-Name
- Custom headers สำหรับการเข้ารหัส (X-Sign, Encrypt)

---

## 5. Data Storage Security / ความปลอดภัยในการจัดเก็บข้อมูล

### 5.1 Password Security | ความปลอดภัยรหัสผ่าน

#### English

All user passwords are securely hashed before storage using industry-standard algorithms.

**Password Hashing:**
- Algorithm: bcrypt with default cost factor
- Salt: Automatically generated per password
- Storage: Only hash stored, never plaintext

**Implementation:**
```go
import "golang.org/x/crypto/bcrypt"
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

**Password Policies:**
- `password_change_count`: Tracks password change history
- `password_change_time`: Timestamp of last password change
- `permission_password`: Separate encrypted password for sensitive operations

#### ภาษาไทย

รหัสผ่านผู้ใช้ทั้งหมดถูกแฮชอย่างปลอดภัยก่อนจัดเก็บโดยใช้อัลกอริทึมมาตรฐานอุตสาหกรรม

**การแฮชรหัสผ่าน:**
- อัลกอริทึม: bcrypt พร้อม default cost factor
- Salt: สร้างอัตโนมัติต่อรหัสผ่าน
- การจัดเก็บ: เก็บเฉพาะ hash ไม่เก็บ plaintext

**การนำไปใช้:**
```go
import "golang.org/x/crypto/bcrypt"
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

**นโยบายรหัสผ่าน:**
- `password_change_count`: ติดตามประวัติการเปลี่ยนรหัสผ่าน
- `password_change_time`: เวลาประทับของการเปลี่ยนรหัสผ่านครั้งล่าสุด
- `permission_password`: รหัสผ่านเข้ารหัสแยกต่างหากสำหรับการดำเนินงานที่ละเอียดอ่อน

---

### 5.2 Sensitive Data Encryption | การเข้ารหัสข้อมูลที่ละเอียดอ่อน

#### English

Sensitive data fields are encrypted at rest using AES-GCM encryption.

**AES-GCM Configuration:**
- Key Size: 256-bit
- Mode: Galois/Counter Mode (GCM) for authenticated encryption
- Encoding: Base32 (lowercase) for URL-safe storage

**Use Cases:**
- QR code tokens for table ordering
- Menu access tokens
- Device binding keys

**Implementation:**
```go
// AES-GCM encryption
block, _ := aes.NewCipher(aesSecretKey)  // 32-byte key
aesGCM, _ := cipher.NewGCM(block)
nonce := make([]byte, aesGCM.NonceSize())
ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
```

#### ภาษาไทย

ฟิลด์ข้อมูลที่ละเอียดอ่อนถูกเข้ารหัสขณะจัดเก็บโดยใช้การเข้ารหัส AES-GCM

**การกำหนดค่า AES-GCM:**
- ขนาดคีย์: 256 บิต
- โหมด: Galois/Counter Mode (GCM) สำหรับการเข้ารหัสที่ยืนยันตัวตน
- การเข้ารหัส: Base32 (ตัวพิมพ์เล็ก) สำหรับการจัดเก็บที่ปลอดภัยสำหรับ URL

**กรณีการใช้งาน:**
- QR code tokens สำหรับการสั่งอาหารที่โต๊ะ
- Menu access tokens
- Device binding keys

**การนำไปใช้:**
```go
// การเข้ารหัส AES-GCM
block, _ := aes.NewCipher(aesSecretKey)  // คีย์ 32 ไบต์
aesGCM, _ := cipher.NewGCM(block)
nonce := make([]byte, aesGCM.NonceSize())
ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
```

---

### 5.3 Database Security | ความปลอดภัยฐานข้อมูล

#### English

**MySQL Configuration:**
- `strict_trans_tables`: Enforces strict SQL mode
- `innodb_flush_log_at_trx_commit = 1`: Full ACID compliance
- `slow_query_log`: Enabled for performance monitoring
- `max_connections = 65535`: Connection limit protection

**Multi-Tenant Isolation:**
- Separate database per company (company_uuid)
- Dynamic database connection based on authenticated user
- No cross-company data access possible

**Query Security:**
- Parameterized queries (prevents SQL injection)
- ORM-based data access (GORM/ThinkPHP)
- Input sanitization at API layer

#### ภาษาไทย

**การกำหนดค่า MySQL:**
- `strict_trans_tables`: บังคับใช้โหมด SQL เข้มงวด
- `innodb_flush_log_at_trx_commit = 1`: การปฏิบัติตาม ACID เต็มรูปแบบ
- `slow_query_log`: เปิดใช้งานสำหรับการตรวจสอบประสิทธิภาพ
- `max_connections = 65535`: การป้องกันขีดจำกัดการเชื่อมต่อ

**การแยก Multi-Tenant:**
- ฐานข้อมูลแยกต่างหากต่อบริษัท (company_uuid)
- การเชื่อมต่อฐานข้อมูลแบบไดนามิกตามผู้ใช้ที่ยืนยันตัวตน
- ไม่สามารถเข้าถึงข้อมูลข้ามบริษัทได้

**ความปลอดภัยของคิวรี:**
- คิวรีแบบพารามิเตอร์ (ป้องกัน SQL injection)
- การเข้าถึงข้อมูลผ่าน ORM (GORM/ThinkPHP)
- การฆ่าเชื้ออินพุตที่ชั้น API

---

### 5.4 Redis Security | ความปลอดภัย Redis

#### English

**Configuration:**
- Memory limit: 256MB with LRU eviction
- Docker network isolation
- No external port exposure in production

**Data Stored:**
- Session data
- Encryption key pairs (temporary)
- Cache data
- Rate limiting counters

#### ภาษาไทย

**การกำหนดค่า:**
- ขีดจำกัดหน่วยความจำ: 256MB พร้อมการขับ LRU
- การแยกเครือข่าย Docker
- ไม่เปิดเผยพอร์ตภายนอกในการใช้งานจริง

**ข้อมูลที่จัดเก็บ:**
- ข้อมูลเซสชัน
- คู่คีย์เข้ารหัส (ชั่วคราว)
- ข้อมูลแคช
- ตัวนับการจำกัดอัตรา

---

## 6. Audit Logging and Monitoring / การบันทึกและการตรวจสอบ

### 6.1 Authentication Logs | บันทึกการยืนยันตัวตน

#### English

All authentication events are logged for security audit purposes.

**Staff Login Log (ttpos_staff_login_log):**
| Field | Description |
|-------|-------------|
| staff_uuid | Employee identifier |
| username | Login username |
| ip | Client IP address |
| result | Login result (success/failure) |
| create_time | Timestamp |

**Logged Events:**
- Successful logins
- Failed login attempts
- Token refresh operations
- Session terminations

#### ภาษาไทย

เหตุการณ์การยืนยันตัวตนทั้งหมดถูกบันทึกเพื่อวัตถุประสงค์ในการตรวจสอบความปลอดภัย

**บันทึกการเข้าสู่ระบบพนักงาน (ttpos_staff_login_log):**
| ฟิลด์ | คำอธิบาย |
|-------|---------|
| staff_uuid | ตัวระบุพนักงาน |
| username | ชื่อผู้ใช้เข้าสู่ระบบ |
| ip | ที่อยู่ IP ไคลเอนต์ |
| result | ผลการเข้าสู่ระบบ (สำเร็จ/ล้มเหลว) |
| create_time | เวลาประทับ |

**เหตุการณ์ที่บันทึก:**
- การเข้าสู่ระบบที่สำเร็จ
- ความพยายามเข้าสู่ระบบที่ล้มเหลว
- การดำเนินการ token refresh
- การสิ้นสุดเซสชัน

---

### 6.2 Operation Logs | บันทึกการดำเนินงาน

#### English

Critical business operations are logged for accountability and troubleshooting.

**Staff Operation Log (ttpos_staff_operation_log):**
| Field | Description |
|-------|-------------|
| staff_uuid | Employee identifier |
| title | Operation title |
| url | API endpoint called |
| request_data | Request payload (sanitized) |
| response_data | Response summary |
| type | Operation type |
| ip | Client IP |
| source | Terminal source |
| agent | User agent string |

**Tracked Operations:**
- Order modifications (discount, cancel, refund)
- Price changes
- Configuration updates
- User management actions

#### ภาษาไทย

การดำเนินงานธุรกิจที่สำคัญถูกบันทึกเพื่อความรับผิดชอบและการแก้ไขปัญหา

**บันทึกการดำเนินงานพนักงาน (ttpos_staff_operation_log):**
| ฟิลด์ | คำอธิบาย |
|-------|---------|
| staff_uuid | ตัวระบุพนักงาน |
| title | ชื่อการดำเนินงาน |
| url | API endpoint ที่เรียก |
| request_data | ข้อมูลคำขอ (ฆ่าเชื้อแล้ว) |
| response_data | สรุปการตอบกลับ |
| type | ประเภทการดำเนินงาน |
| ip | IP ไคลเอนต์ |
| source | แหล่งเทอร์มินัล |
| agent | สตริง user agent |

**การดำเนินงานที่ติดตาม:**
- การแก้ไขคำสั่ง (ส่วนลด, ยกเลิก, คืนเงิน)
- การเปลี่ยนแปลงราคา
- การอัปเดตการกำหนดค่า
- การดำเนินการจัดการผู้ใช้

---

### 6.3 Request Logging | การบันทึกคำขอ

#### English

Comprehensive request logging middleware captures all API interactions for debugging and monitoring.

**Logged Information:**
- Timestamp
- HTTP method and URI
- Client IP and User-Agent
- Request headers
- Query parameters and path parameters
- Request body (JSON formatted)
- Response status

**Privacy Considerations:**
- Sensitive headers are not logged in production
- Password fields are excluded from logs
- Request bodies are truncated for large payloads

#### ภาษาไทย

มิดเดิลแวร์การบันทึกคำขออย่างครอบคลุมจับการโต้ตอบ API ทั้งหมดสำหรับการดีบักและการตรวจสอบ

**ข้อมูลที่บันทึก:**
- เวลาประทับ
- วิธี HTTP และ URI
- IP ไคลเอนต์และ User-Agent
- Request headers
- Query parameters และ path parameters
- Request body (รูปแบบ JSON)
- สถานะการตอบกลับ

**ข้อควรพิจารณาเรื่องความเป็นส่วนตัว:**
- Headers ที่ละเอียดอ่อนไม่ถูกบันทึกในการใช้งานจริง
- ฟิลด์รหัสผ่านถูกแยกออกจากบันทึก
- Request bodies ถูกตัดสำหรับ payloads ขนาดใหญ่

---

### 6.4 Database Query Logging | การบันทึกคิวรีฐานข้อมูล

#### English

**MySQL Slow Query Log:**
- Location: `/var/lib/mysql/mysql-slow.log`
- Threshold: 3 seconds
- Used for: Performance optimization, identifying problematic queries

**Binary Logging:**
- Enabled for replication and point-in-time recovery
- Log expiration: 5 days

#### ภาษาไทย

**บันทึกคิวรีช้า MySQL:**
- ตำแหน่ง: `/var/lib/mysql/mysql-slow.log`
- เกณฑ์: 3 วินาที
- ใช้สำหรับ: การปรับปรุงประสิทธิภาพ, การระบุคิวรีที่มีปัญหา

**การบันทึก Binary:**
- เปิดใช้งานสำหรับการจำลองและการกู้คืน point-in-time
- การหมดอายุบันทึก: 5 วัน

---

## 7. Security Best Practices / แนวปฏิบัติที่ดีด้านความปลอดภัย

### 7.1 Development Guidelines | แนวทางการพัฒนา

#### English

**Input Validation:**
```go
// ✅ Strict validation with binding tags
type CreateUserReq struct {
    Username string `binding:"required,min=2,max=20,alphanum"`
    Email    string `binding:"required,email"`
    Age      int    `binding:"required,min=1,max=150"`
}
```

**SQL Injection Prevention:**
```go
// ✅ Parameterized queries
db.Where("username = ?", username).First(&user)

// ❌ Never use string concatenation
query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", username)
```

**XSS Prevention:**
```go
// ✅ Escape output
import "html"
safeContent := html.EscapeString(userInput)
```

#### ภาษาไทย

**การตรวจสอบอินพุต:**
```go
// ✅ การตรวจสอบอย่างเข้มงวดด้วย binding tags
type CreateUserReq struct {
    Username string `binding:"required,min=2,max=20,alphanum"`
    Email    string `binding:"required,email"`
    Age      int    `binding:"required,min=1,max=150"`
}
```

**การป้องกัน SQL Injection:**
```go
// ✅ คิวรีแบบพารามิเตอร์
db.Where("username = ?", username).First(&user)

// ❌ อย่าใช้การต่อสตริง
query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", username)
```

**การป้องกัน XSS:**
```go
// ✅ หลีกเลี่ยงเอาต์พุต
import "html"
safeContent := html.EscapeString(userInput)
```

---

### 7.2 Deployment Checklist | รายการตรวจสอบการปรับใช้

#### English

**Production Security Checklist:**

- [ ] Disable debug mode
- [ ] Enable HTTPS with valid SSL certificate
- [ ] Configure proper CORS policies
- [ ] Implement rate limiting
- [ ] Configure firewall rules
- [ ] Remove default credentials
- [ ] Enable all security headers
- [ ] Review and minimize exposed ports
- [ ] Enable audit logging
- [ ] Set up monitoring alerts

#### ภาษาไทย

**รายการตรวจสอบความปลอดภัยการใช้งานจริง:**

- [ ] ปิดโหมดดีบัก
- [ ] เปิดใช้งาน HTTPS พร้อมใบรับรอง SSL ที่ถูกต้อง
- [ ] กำหนดค่านโยบาย CORS ที่เหมาะสม
- [ ] ใช้การจำกัดอัตรา
- [ ] กำหนดค่ากฎไฟร์วอลล์
- [ ] ลบข้อมูลรับรองเริ่มต้น
- [ ] เปิดใช้งาน security headers ทั้งหมด
- [ ] ตรวจสอบและลดพอร์ตที่เปิดเผย
- [ ] เปิดใช้งานการบันทึกการตรวจสอบ
- [ ] ตั้งค่าการแจ้งเตือนการตรวจสอบ

---

### 7.3 Prohibited Practices | แนวปฏิบัติที่ห้าม

#### English

| Prohibited | Reason |
|------------|--------|
| ❌ SQL string concatenation | SQL injection risk |
| ❌ Storing sensitive data in JWT | Security exposure |
| ❌ Plaintext password storage | Must be hashed |
| ❌ Skipping input validation | Security vulnerabilities |
| ❌ Outputting raw user input | XSS risk |
| ❌ Logging sensitive information | Data leakage |
| ❌ Using default secrets in production | Easy to compromise |

#### ภาษาไทย

| ห้าม | เหตุผล |
|------|--------|
| ❌ การต่อสตริง SQL | ความเสี่ยง SQL injection |
| ❌ จัดเก็บข้อมูลที่ละเอียดอ่อนใน JWT | การเปิดเผยด้านความปลอดภัย |
| ❌ จัดเก็บรหัสผ่าน plaintext | ต้องถูกแฮช |
| ❌ ข้ามการตรวจสอบอินพุต | ช่องโหว่ด้านความปลอดภัย |
| ❌ แสดงอินพุตผู้ใช้ดิบ | ความเสี่ยง XSS |
| ❌ บันทึกข้อมูลที่ละเอียดอ่อน | การรั่วไหลของข้อมูล |
| ❌ ใช้ secrets เริ่มต้นในการใช้งานจริง | ง่ายต่อการถูกโจมตี |

---

## Summary / สรุป

### English

TTPOS implements a comprehensive, multi-layered security architecture designed to protect sensitive business and customer data. Key security measures include:

1. **Transport Security**: TLS encryption for all external communications
2. **Authentication**: JWT-based stateless authentication with refresh token support
3. **Authorization**: Role-based access control with multi-tenant isolation
4. **Data Protection**: Bcrypt password hashing and AES-GCM encryption for sensitive data
5. **Audit Trail**: Comprehensive logging of authentication, operations, and requests

The system is designed to meet the security requirements of modern POS systems while maintaining flexibility and scalability for restaurant operations.

### ภาษาไทย

TTPOS ใช้สถาปัตยกรรมความปลอดภัยแบบหลายชั้นที่ครอบคลุม ออกแบบมาเพื่อปกป้องข้อมูลธุรกิจและลูกค้าที่ละเอียดอ่อน มาตรการความปลอดภัยหลักประกอบด้วย:

1. **ความปลอดภัยในการส่ง**: การเข้ารหัส TLS สำหรับการสื่อสารภายนอกทั้งหมด
2. **การยืนยันตัวตน**: การยืนยันตัวตนแบบ stateless ด้วย JWT พร้อมการสนับสนุน refresh token
3. **การอนุญาต**: การควบคุมการเข้าถึงตามบทบาทพร้อมการแยก multi-tenant
4. **การปกป้องข้อมูล**: การแฮชรหัสผ่าน Bcrypt และการเข้ารหัส AES-GCM สำหรับข้อมูลที่ละเอียดอ่อน
5. **บันทึกการตรวจสอบ**: การบันทึกอย่างครอบคลุมของการยืนยันตัวตน, การดำเนินงาน และคำขอ

ระบบได้รับการออกแบบมาเพื่อตอบสนองข้อกำหนดด้านความปลอดภัยของระบบ POS สมัยใหม่ ในขณะที่ยังคงความยืดหยุ่นและความสามารถในการขยายสำหรับการดำเนินงานร้านอาหาร

---

**Document Maintainer**: TTPOS Development Team  
**Contact**: security@ttpos.com  
**Last Review Date**: 2025-11-27

