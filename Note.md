# 🎒 Context package ใน Go คืออะไร?

context เป็นแพ็กเกจพื้นฐานของ Go ที่ใช้สำหรับ
ส่งข้อมูลที่เกี่ยวกับ request, ควบคุม lifecycle, timeout, cancel, และ metadata ระหว่าง function chain ทั้งระบบ

ให้คิดง่าย ๆ ว่า context = กระเป๋าเอกสารที่พกไปตลอดเส้นทางของ request
ทุกชั้น—router → middleware → service → repository → database
จะใช้ context ตัวเดียวกัน

## 🧩 ใช้ทำอะไรได้บ้าง?

✅ 1. ยกเลิกงาน (cancel) เมื่อ request ถูกยกเลิก

เช่น user ปิด browser → server ยกเลิกงานทันทีเพื่อไม่เสีย resource

✅ 2. ตั้ง timeout

เช่น query DB ต้องไม่เกิน 3 วินาที

✅ 3. ส่งข้อมูลเสริม (metadata)

เช่น user ID, request ID, permission, locale, trace ID, correlation ID

✅ 4. ใช้ใน database driver (pgx, sqlx)

เช่น

```go
rows, err := db.Query(ctx, "SELECT * FROM users")

```

ทุก DB operation ต้องมี ctx เพื่อสามารถ timeout หรือ cancel ได้

## 🏗 ทำไม clean architecture ใช้ context เยอะ?

เพราะ clean architecture ทำงานเป็น layer
และ context ช่วย “ส่งข้อมูล request เดิม” ผ่านทุกชั้น เช่น:

```api
HTTP Handler
   ↓
Controller / Delivery
   ↓
UseCase / Service
   ↓
Repository
   ↓
Database

```

ทุกฟังก์ชันต้องรับ ctx context.Context เสมอ
เพื่อให้สามารถ ยกเลิกงานพร้อมกัน ได้เมื่อ request ตายไปแล้ว

## 🧠 context มีอะไรข้างในได้บ้าง?

**แบบ built-in:**

- Deadline (เวลาหมดอายุ)

- Done channel (สั่งยกเลิก)

- Error (บอก status)

**แบบ custom:**

- คุณสามารถ context.WithValue() เพื่อเก็บค่า เช่น:

- userID

- transactionID

- jwt claims

_**แต่ ควรใช้เฉพาะ metadata ที่เกี่ยวข้องกับ request เท่านั้น**_

## 🛠 ตัวอย่างจริงใน project ของคุณ

1.Handler

```go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var input CreateUserRequest
    json.NewDecoder(r.Body).Decode(&input)

    err := h.service.CreateUser(ctx, input)
    if err != nil {
        http.Error(w, err.Error(), 400)
        return
    }
}
```

2.Service Layer

```go
func (s *userService) CreateUser(ctx context.Context, input CreateUserRequest) error {
    // can set timeout
    ctx, cancel := context.WithTimeout(ctx, time.Second*3)
    defer cancel()

    return s.repo.Create(ctx, &user)
}
```

3.Repository Layer (pgx)

```go
func (r *repo) Create(ctx context.Context, u *User) error {
    query := `INSERT INTO users(first_name, last_name, email, password)
              VALUES ($1, $2, $3, $4)`

    _, err := r.db.Exec(ctx, query,
        u.FirstName, u.LastName, u.Email, u.Password,
    )
    return err
}

```

**เพราะ pgx ใช้ ctx เพื่อ:**

- timeout DB

- cancel query

- pass metadata

## 📌 ทำไมทุก function ต้องมี ctx?

### เพราะถ้า HTTP request ถูกยกเลิก Go จะ broadcast signal ผ่าน ctx ลงไปถึง DB

ตัวอย่าง:

👉 user ปิดเบราว์เซอร์
👉 server รับรู้ผ่าน r.Context().Done()
👉 DB query ยกเลิกทันที
👉 goroutine ต่าง ๆ หยุดใช้งาน

ทำให้ ไม่เสีย resource ทิ้ง
และ ระบบรองรับ high load ได้ดีขึ้น

## ❌ Context ไม่ควรใช้เก็บอะไร?

❌ ไม่ควรเก็บ business data

เช่น:

- product list

- user struct ทั้งตัว

- DTO / big object

- config

- settings

- db connection

เพราะ:

1. context ถูกออกแบบมาให้ immutable หรือเปลี่ยนแปลงน้อยที่สุด

2. context เป็น per-request ไม่ได้อยู่ตลอดชีวิตโปรแกรม

3. ถ้าโยนข้อมูลใหญ่ ๆ จะทำให้ performance แย่ลง

4.ทำให้ code อ่านยากมาก

## 🧠 คีย์สำคัญคือ "metadata per request"

**context ควรใช้แค่:**

- ใส่ค่าที่จำเป็นสำหรับ request

- ใส่ค่าที่ไม่ใช่ business logic

- ใส่ค่าที่ทุก layer ต้องรับรู้

- ประมาณนี้ถึงจะถูกต้องตาม Go convention

## 📦 ตัวอย่างที่ถูกต้อง

✔ Request ID

ใช้ทำ log trace

✔ User ID จาก JWT

ใช้ทำ RBAC ได้ทั้งระบบ

✔ Deadline

ใช้ควบคุม timeout

✔ Correlation ID

ใช้สำหรับ distributed tracing (microservices)

## 📦 ตัวอย่างที่ไม่ควรทำ (ผิด)

❌ โยน struct 10KB ลง context

❌ โยน DB config / JWT secret ลง context

❌ ใช้ context เป็นเหมือน global variable

❌ เอา business logic มาใช้ผ่าน context

## 🔥 ตัวอย่างดีที่สุดใน project คุณ (Go + chi + clean architecture)

### Middleware: extract userId → put in context

```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := extractFromJWT(r)
        ctx := context.WithValue(r.Context(), "userID", userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Handler: คว้า userID ไปใช้

```go
func (h *RoomHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    userID := ctx.Value("userID").(int64)
    rooms, err := h.service.GetUserRooms(ctx, userID)
```
